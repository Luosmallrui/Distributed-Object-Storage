package minIo

import (
	"bytes"
	"context"
	"distributed-object-storage/pkg/log"
	"distributed-object-storage/types"
	"fmt"
	"github.com/minio/minio-go/v7"
	client "go.etcd.io/etcd/client/v3"
	"io"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/minio/minio-go/v7/pkg/credentials"
)

var StorageClient *MinioHelper
var etcdClient *client.Client

const (
	ChunkPartSize = 1024 * 1024 * 5
	FilePermMode  = os.FileMode(0664)
)

func GetStorageNodeList() ([]types.KvStorage, error) {
	storageNodeList := make([]types.KvStorage, 0)
	var err error
	etcdClient, err = client.New(client.Config{
		Endpoints: []string{"http://0.0.0.0:2379"},
	})
	if err != nil {
		return storageNodeList, err
	}
	resp, err := etcdClient.Get(context.Background(), "minio/", client.WithPrefix())
	if err != nil {
		return storageNodeList, err
	}
	for _, node := range resp.Kvs {
		suffix := strings.TrimPrefix(string(node.Key), "minio/")
		storageNodeList = append(storageNodeList, types.KvStorage{Key: suffix, Value: string(node.Value)})
	}
	return storageNodeList, nil
}

func GetMinioClient(endpoint string) *MinioHelper {
	core, err := minio.NewCore(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4("root", "rootroot", ""),
		Secure: false,
	})
	if err != nil {
		panic(err.Error())
	}
	StorageClient = &MinioHelper{
		MinioCore: core,
	}
	return StorageClient
}

type FileInfo struct {
	Name    string
	Size    int64
	ModTime time.Time
}

func (f *FileInfo) GetName() string {
	return f.Name
}

func (f *FileInfo) GetSize() int64 {
	return f.Size
}

func (f *FileInfo) GetModeTime() time.Time {
	return f.ModTime
}

type MultipartFile struct {
	UploadId       string               `json:"uploadId"`
	BucketName     string               `json:"bucketName"`
	FileInfo       FileInfo             `json:"fileInfo"`
	CompletedParts []minio.CompletePart `json:"-"`
}

type fileChunk struct {
	PartNumber int
	Start      int
	End        int
}

type MinioHelper struct {
	MinioCore *minio.Core
}

func (helper *MinioHelper) Upload(ctx context.Context, bucketName, objectName string, reader io.Reader, size int64, UploadID string) (*minio.UploadInfo, error) {
	// If the size is small enough, upload directly
	if size <= ChunkPartSize {
		return uploadFile(ctx, bucketName, objectName, reader, size)
	}

	// For larger files, use multipart upload
	return uploadFileWithCP(ctx, reader, bucketName, objectName, size, UploadID)
}

// uploadFileWithCP 通过checkPoint文件上传
func uploadFileWithCP(ctx context.Context, reader io.Reader, bucketName, objectName string, size int64, uploadID string) (*minio.UploadInfo, error) {
	chunkCount := int(size / ChunkPartSize)
	if size%ChunkPartSize != 0 {
		chunkCount++
	}

	// Create a new MultipartFile
	cpFile := &MultipartFile{
		BucketName: bucketName,
		FileInfo: FileInfo{
			Name:    objectName,
			Size:    size,
			ModTime: time.Now(),
		},
	}

	// Initialize multipart upload
	minioUploadID, err := StorageClient.MinioCore.NewMultipartUpload(ctx, bucketName, objectName, minio.PutObjectOptions{})
	if err != nil {
		log.Errorf("NewMultipartUpload error: %v", err)
		return nil, err
	}
	cpFile.UploadId = minioUploadID

	// 检查上传任务是否存在
	types.UploadTasks.RLock()
	status, exists := types.UploadTasks.Tasks[uploadID]
	types.UploadTasks.RUnlock()
	if !exists {
		return nil, fmt.Errorf("upload task not found")
	}

	// Prepare chunks
	chunks := make([]fileChunk, chunkCount)
	for i := 0; i < chunkCount; i++ {
		chunks[i] = fileChunk{
			PartNumber: i + 1,
			Start:      i * ChunkPartSize,
			End:        min((i+1)*ChunkPartSize, int(size)),
		}
	}

	// Set up channels for parallel processing
	jobs := make(chan fileChunk, len(chunks))
	results := make(chan minio.CompletePart, len(chunks))
	errs := make(chan error, 1)

	// Start worker goroutines
	var wg sync.WaitGroup
	workerCount := 5
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for chunk := range jobs {
				// 检查取消状态
				if status.IsCanceled {
					errs <- fmt.Errorf("upload canceled")
					return
				}

				// 检查暂停状态
				for {
					status.Mutex.Lock()
					if !status.IsPaused {
						status.Mutex.Unlock()
						break
					}
					status.Mutex.Unlock()
					time.Sleep(time.Second)
				}

				buffer := make([]byte, chunk.End-chunk.Start)
				_, err := io.ReadFull(reader, buffer)
				if err != nil {
					errs <- fmt.Errorf("read chunk error: %v", err)
					return
				}

				part, err := chunkUpload(ctx, buffer, bucketName, objectName, minioUploadID, chunk.PartNumber)
				if err != nil {
					errs <- fmt.Errorf("upload chunk error: %v", err)
					return
				}

				// 更新上传状态
				status.Mutex.Lock()
				status.CurrentPart = chunk.PartNumber
				status.CompletedParts = append(status.CompletedParts, part)
				status.Mutex.Unlock()

				select {
				case results <- part:
				case <-ctx.Done():
					errs <- ctx.Err()
					return
				}
			}
		}()
	}

	// Send jobs to workers
	go func() {
		for _, chunk := range chunks {
			select {
			case jobs <- chunk:
			case <-ctx.Done():
				return
			}
		}
		close(jobs)
	}()

	// Collect results and handle errors
	cpFile.CompletedParts = make([]minio.CompletePart, chunkCount)
	resultCount := 0

	// Start a goroutine to wait for all workers to finish
	go func() {
		wg.Wait()
		close(results)
	}()

	for {
		select {
		case part, ok := <-results:
			if !ok {
				// All workers have finished
				if resultCount == chunkCount {
					// 所有分片都上传成功
					return complete(ctx, bucketName, objectName, cpFile, "")
				}
				// Some parts failed
				_ = StorageClient.MinioCore.AbortMultipartUpload(ctx, bucketName, objectName, minioUploadID)
				return nil, fmt.Errorf("incomplete upload: got %d/%d parts", resultCount, chunkCount)
			}
			cpFile.CompletedParts[part.PartNumber-1] = part
			resultCount++

		case err := <-errs:
			// 如果是取消操作，清理已上传的部分
			if status.IsCanceled {
				_ = StorageClient.MinioCore.AbortMultipartUpload(ctx, bucketName, objectName, minioUploadID)
			}
			return nil, err

		case <-ctx.Done():
			_ = StorageClient.MinioCore.AbortMultipartUpload(ctx, bucketName, objectName, minioUploadID)
			return nil, ctx.Err()
		}
	}
}

// uploadFile 不分片直接上传
func uploadFile(ctx context.Context, bucketName, objectName string, reader io.Reader, size int64) (*minio.UploadInfo, error) {
	uploadInfo, err := StorageClient.MinioCore.PutObject(ctx, bucketName, objectName, reader, size, "", "", minio.PutObjectOptions{})
	if err != nil {
		log.Errorf("put object error: %v", err)
		return nil, err
	}
	return &uploadInfo, nil
}

// chunkUpload 上传分片
func chunkUpload(ctx context.Context, buf []byte, bucketName string, fileName, uploadId string, partNumber int) (minio.CompletePart, error) {
	buffer := bytes.NewBuffer(buf)
	objectPart, err := StorageClient.MinioCore.PutObjectPart(ctx, bucketName, fileName, uploadId, partNumber, buffer, int64(buffer.Len()), minio.PutObjectPartOptions{})
	if err != nil {
		log.Errorf("Upload part error: %s", err)
		return minio.CompletePart{}, err
	}
	log.Info("Upload chunk success, objectPart PartNumber:", objectPart.PartNumber)
	return minio.CompletePart{
		ETag:       objectPart.ETag,
		PartNumber: objectPart.PartNumber,
	}, nil
}

// complete 合并分片
func complete(ctx context.Context, bucketName, objectName string, cpFile *MultipartFile, cpFilePath string) (*minio.UploadInfo, error) {
	sort.Slice(cpFile.CompletedParts, func(i, j int) bool {
		return cpFile.CompletedParts[i].PartNumber < cpFile.CompletedParts[j].PartNumber
	})

	uploadInfo, err := StorageClient.MinioCore.CompleteMultipartUpload(ctx, bucketName, objectName, cpFile.UploadId, cpFile.CompletedParts, minio.PutObjectOptions{})
	if err != nil {
		log.Errorf("CompleteMultipartUpload err: %s", err)
		return nil, err
	}

	return &uploadInfo, nil
}

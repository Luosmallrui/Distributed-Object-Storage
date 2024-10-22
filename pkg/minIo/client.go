package minIo

import (
	"bytes"
	"context"
	"distributed-object-storage/pkg/log"
	"distributed-object-storage/types"
	"encoding/json"
	"fmt"
	"github.com/minio/minio-go/v7"
	client "go.etcd.io/etcd/client/v3"
	"io"
	"io/fs"
	"os"
	"sort"
	"strings"
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

func (cp *MultipartFile) getChunks(ctx context.Context, chunkCount int, bucketName, objectName string) ([]fileChunk, error) {
	// 初始化长度为分片数，并且为零值
	cp.CompletedParts = make([]minio.CompletePart, chunkCount, chunkCount)

	partInfos, err := listObjectParts(ctx, bucketName, objectName, cp.UploadId)
	if err != nil {
		uploadId, err := StorageClient.MinioCore.NewMultipartUpload(ctx, cp.BucketName, objectName, minio.PutObjectOptions{})
		if err != nil {
			log.Errorf("NewMultipartUpload error: %v", err)
			return nil, err
		}
		cp.UploadId = uploadId
		return []fileChunk{}, nil
	}

	var completedPartMap = make(map[int]minio.ObjectPart)
	var chunks = make([]fileChunk, 0, chunkCount-len(partInfos))
	for _, partInfo := range partInfos {
		cp.CompletedParts[partInfo.PartNumber-1] = minio.CompletePart{PartNumber: partInfo.PartNumber, ETag: partInfo.ETag}
		completedPartMap[partInfo.PartNumber] = partInfo
	}

	for i := 0; i < chunkCount; i++ {
		if _, ok := completedPartMap[i+1]; ok {
			continue
		}
		var chunk fileChunk
		chunk.PartNumber = i + 1
		chunk.Start = i * ChunkPartSize
		end := chunk.Start + ChunkPartSize
		if i == chunkCount-1 {
			end = int(cp.FileInfo.GetSize())
		}
		chunk.End = end
		chunks = append(chunks, chunk)
	}
	return chunks, nil
}

func (cp *MultipartFile) load(ctx context.Context, bucketName, objectName string, fileInfo os.FileInfo) (string, error) {
	cpFilePath := objectName + ".cp"
	cpFile, err := os.Open(cpFilePath)
	defer cpFile.Close()
	if err != nil {
		log.Errorf("open cpFile failed: %v, newMultipart", err)
		cp.FileInfo = FileInfo{Name: fileInfo.Name(), Size: fileInfo.Size(), ModTime: fileInfo.ModTime()}
		err = cp.newMultipart(ctx, bucketName, objectName)
		if err != nil {
			log.Errorf("newMultipart failed: %v", err)
			return "", err
		}
	} else {
		cpFileBytes, err := io.ReadAll(cpFile)
		if err != nil {
			log.Errorf("read cpFile failed: %v", err)
			return "", err
		}
		err = json.Unmarshal(cpFileBytes, &cp)
		if err != nil {
			log.Errorf("unmarshal cpFile failed: %v", err)
			return "", err
		}

		// 判断文是否相同，否则重新创建分片上传
		if !cp.isSameFile(fileInfo, bucketName) {
			cp.FileInfo = FileInfo{Name: fileInfo.Name(), Size: fileInfo.Size(), ModTime: fileInfo.ModTime()}
			err = cp.newMultipart(ctx, bucketName, objectName)
			if err != nil {
				log.Errorf("newMultipart failed: %v", err)
				return "", err
			}
		}
	}
	return cpFilePath, nil
}

func (cp *MultipartFile) newMultipart(ctx context.Context, bucketName, objectName string) error {
	uploadId, err := StorageClient.MinioCore.NewMultipartUpload(ctx, bucketName, objectName, minio.PutObjectOptions{})
	if err != nil {
		log.Errorf("NewMultipartUpload error: %v", err)
		return err
	}
	cp.UploadId = uploadId
	cp.BucketName = bucketName
	return nil
}

func (cp *MultipartFile) dump(cpFilePath string) error {
	cpFileBytes, err := json.Marshal(cp)
	if err != nil {
		return err
	}
	err = os.WriteFile(cpFilePath, cpFileBytes, FilePermMode)
	if err != nil {
		return err
	}
	return nil
}

func (cp *MultipartFile) isSameFile(fileInfo fs.FileInfo, bucketName string) bool {
	return cp.FileInfo.GetName() == fileInfo.Name() &&
		cp.FileInfo.GetSize() == fileInfo.Size() &&
		cp.FileInfo.GetModeTime().Equal(fileInfo.ModTime()) &&
		cp.BucketName != bucketName
}

type MinioHelper struct {
	MinioCore *minio.Core
}

func (helper *MinioHelper) Upload(ctx context.Context, bucketName, objectName string, reader io.Reader, size int64) (*minio.UploadInfo, error) {
	// If the size is small enough, upload directly
	if size <= ChunkPartSize {
		return uploadFile(ctx, bucketName, objectName, reader, size)
	}

	// For larger files, use multipart upload
	return uploadFileWithCP(ctx, reader, bucketName, objectName, size)
}

// listObjectParts 列出已上传的分片
func listObjectParts(ctx context.Context, bucketName, objectName, uploadID string) (map[int]minio.ObjectPart, error) {
	// Part number marker for the next batch of request.
	var nextPartNumberMarker int
	var partsInfo = make(map[int]minio.ObjectPart)
	for {
		// Get list of uploaded parts a maximum of 1000 per request.
		listObjPartsResult, err := StorageClient.MinioCore.ListObjectParts(ctx, bucketName, objectName, uploadID, nextPartNumberMarker, 1000)
		if err != nil {
			return nil, err
		}
		for _, part := range listObjPartsResult.ObjectParts {
			// Trim off the odd double quotes from ETag in the beginning and end.
			part.ETag = strings.TrimPrefix(part.ETag, "")
			part.ETag = strings.TrimSuffix(part.ETag, "")
			partsInfo[part.PartNumber] = part
		}
		// Keep part number marker, for the next iteration.
		nextPartNumberMarker = listObjPartsResult.NextPartNumberMarker
		// Listing ends result is not truncated, return right here.
		if !listObjPartsResult.IsTruncated {
			break
		}
	}

	// Return all the parts.
	return partsInfo, nil
}

// uploadFileWithCP 通过checkPoint文件上传
func uploadFileWithCP(ctx context.Context, reader io.Reader, bucketName, objectName string, size int64) (*minio.UploadInfo, error) {
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
	uploadID, err := StorageClient.MinioCore.NewMultipartUpload(ctx, bucketName, objectName, minio.PutObjectOptions{})
	if err != nil {
		log.Errorf("NewMultipartUpload error: %v", err)
		return nil, err
	}
	cpFile.UploadId = uploadID

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
	workerCount := 5 // You can adjust this number
	for i := 0; i < workerCount; i++ {
		go func() {
			for chunk := range jobs {
				buffer := make([]byte, chunk.End-chunk.Start)
				_, err := io.ReadFull(reader, buffer)
				if err != nil {
					errs <- fmt.Errorf("read chunk error: %v", err)
					return
				}

				part, err := chunkUpload(ctx, buffer, bucketName, objectName, uploadID, chunk.PartNumber)
				if err != nil {
					errs <- fmt.Errorf("upload chunk error: %v", err)
					return
				}

				results <- part
			}
		}()
	}

	// Send jobs to workers
	go func() {
		for _, chunk := range chunks {
			jobs <- chunk
		}
		close(jobs)
	}()

	// Collect results
	cpFile.CompletedParts = make([]minio.CompletePart, chunkCount)
	for i := 0; i < chunkCount; i++ {
		select {
		case part := <-results:
			cpFile.CompletedParts[part.PartNumber-1] = part
		case err := <-errs:
			return nil, err
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	// Complete multipart upload
	return complete(ctx, bucketName, objectName, cpFile, "")
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

// work 分片上传worker
func work(ctx context.Context, parts <-chan fileChunk, fileBytes []byte, bucketName, fileName string, uploadId string, failed chan error, results chan minio.CompletePart, die chan bool) {
	for part := range parts {
		log.Info("upload chunk chunkNumber:", part.PartNumber)
		completePart, err := chunkUpload(ctx, fileBytes[part.Start:part.End], bucketName, fileName, uploadId, part.PartNumber)
		if err != nil {
			log.Errorf("upload chunk error: %s, chunkNumber: %d", err, part.PartNumber)
			failed <- err
			break
		}
		select {
		case <-die:
			return
		default:
		}
		results <- completePart
	}

}

// scheduler function
func scheduler(jobs chan fileChunk, chunks []fileChunk) {
	for _, chunk := range chunks {
		jobs <- chunk
	}
	close(jobs)
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

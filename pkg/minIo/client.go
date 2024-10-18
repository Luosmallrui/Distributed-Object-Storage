package minIo

import (
	"bytes"
	"context"
	"distributed-object-storage/pkg/log"
	"distributed-object-storage/types"
	"encoding/json"
	"github.com/minio/minio-go/v7"
	client "go.etcd.io/etcd/client/v3"
	"io"
	"io/fs"
	"io/ioutil"
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

func (helper *MinioHelper) Upload(ctx context.Context, bucketName, objectName string, filePath string) (*minio.UploadInfo, error) {
	openFile, err := os.Open(filePath)
	defer openFile.Close()
	if err != nil {
		log.Errorf("openfile failed: %v", err)
		return nil, err
	}

	fileInfo, err := openFile.Stat()
	if err != nil {
		log.Errorf("get openfile statInfo error: %v", err)
		return nil, err
	}
	chunkCount := int(fileInfo.Size()) / ChunkPartSize
	if chunkCount <= 1 {
		return uploadFile(ctx, bucketName, objectName, filePath, fileInfo)
	}
	return uploadFileWithCP(ctx, filePath, bucketName, objectName, chunkCount)
}

func GetMinioHelper() *MinioHelper {
	return StorageClient
}

func isObjectExist(ctx context.Context, bucketName string, objectName string, fileSize int64) bool {
	objInfo, err := StorageClient.MinioCore.StatObject(ctx, bucketName, objectName, minio.StatObjectOptions{})
	if err != nil {
		log.Errorf("get object error, bucket: %s, obj: %s, error: %s", bucketName, objectName, err)
		return false
	}
	log.Info("exist file info: ", objInfo)
	return fileSize == objInfo.Size
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
func uploadFileWithCP(ctx context.Context, filePath, bucketName, objectName string, chunkCount int) (*minio.UploadInfo, error) {
	openFile, err := os.Open(filePath)
	defer openFile.Close()
	if err != nil {
		return nil, err
	}
	fileInfo, err := openFile.Stat()
	if err != nil {
		return nil, err
	}

	// 获取分片上传信息
	var cpFile = new(MultipartFile)

	cpFilePath, err := cpFile.load(ctx, bucketName, objectName, fileInfo)
	if err != nil {
		return nil, err
	}
	// 查询已上传分片
	chunks, err := cpFile.getChunks(ctx, chunkCount, bucketName, objectName)
	if err != nil {
		return nil, err
	}

	fileBytes, err := ioutil.ReadAll(openFile)
	if err != nil {
		return nil, err
	}

	// 继续上传未上传分片
	var jobs = make(chan fileChunk, len(chunks))
	var results = make(chan minio.CompletePart, len(chunks))
	var failed = make(chan error)
	var die = make(chan bool)

	// 分片上传
	routines := 5
	if len(chunks) < 5 {
		routines = len(chunks)
	}
	for i := 0; i < routines; i++ {
		go work(ctx, jobs, fileBytes, bucketName, objectName, cpFile.UploadId, failed, results, die)
	}

	go scheduler(jobs, chunks)

	completed := 0
	for completed < len(chunks) {
		select {
		case part := <-results:
			completed++
			cpFile.CompletedParts[part.PartNumber-1] = part
			cpFile.dump(cpFilePath)
		case err := <-failed:
			close(die) // 停止worker
			return nil, err
		}

		if completed >= len(chunks) {
			break
		}
	}

	return complete(ctx, bucketName, objectName, cpFile, cpFilePath)
}

// uploadFile 不分片直接上传
func uploadFile(ctx context.Context, bucketName, objectName, filePath string, fileInfo os.FileInfo) (*minio.UploadInfo, error) {
	// 只有一片直接上传
	openFile, err := os.Open(filePath)
	defer openFile.Close()
	if err != nil {
		return nil, err
	}
	fileName := fileInfo.Name()
	if objectName != "" {
		fileName = objectName
	}
	uploadInfo, err := StorageClient.MinioCore.PutObject(ctx, bucketName, fileName, openFile, fileInfo.Size(), "", "", minio.PutObjectOptions{})
	if err != nil {
		log.Errorf("put object error: %v", err)
		return nil, err
	}
	return &uploadInfo, nil
}

// chunkUpload 上传分片
func chunkUpload(ctx context.Context, buf []byte, bucketName string, fileName, uploadId string, partNumber int) (minio.CompletePart, error) {
	buffer := bytes.Buffer{}
	buffer.Write(buf)
	objectPart, err := StorageClient.MinioCore.PutObjectPart(ctx, bucketName, fileName, uploadId, partNumber, &buffer, int64(buffer.Len()), minio.PutObjectPartOptions{})
	if err != nil {
		log.Errorf("read buff err: %s", err)
		return minio.CompletePart{}, err
	}
	log.Info("upload chunk success, objectPart PartNumber:", objectPart.PartNumber)
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
	res := &minio.UploadInfo{}
	UploadInfo, err := StorageClient.MinioCore.CompleteMultipartUpload(ctx, bucketName, objectName, cpFile.UploadId, cpFile.CompletedParts, minio.PutObjectOptions{})
	if err != nil {
		log.Errorf("CompleteMultipartUpload err: %s", err)
		return res, err
	}
	err = os.Remove(cpFilePath)
	if err != nil {
		log.Errorf("Failed to remove checkpoint path err: %s", err)
	}
	res = &UploadInfo
	return res, nil
}

package svc

import (
	"context"
	"distributed-object-storage/config"
	"distributed-object-storage/pkg/db/dao"
	"distributed-object-storage/types"
	"errors"
	"fmt"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"io"
	"log"
	"strconv"
	"strings"
	"time"
)

type StorageNode interface {
	PutObject(ctx context.Context, bucketName, objectName string, data io.Reader, size int64, metadata map[string]string) (string, error)
	GetObject(ctx context.Context, bucketName, objectName string) (io.ReadCloser, types.ObjectInfo, error)
}

type StorageNodeSvc struct {
}

func NewStorageNodeSvc(s *dao.S) *StorageNodeSvc {
	return &StorageNodeSvc{}
}

/*
PutObject 存储⼀个完整的对象。
输⼊:
  - ctx: 上下⽂，⽤于处理超时和取消
  - bucketName: 桶名称
  - objectName: 对象名称
  - data: 对象数据的读取器
  - size: 对象的⼤⼩
  - metadata: 对象的元数据
    输出:
  - string: 存储成功后的ETag
  - error: 如果存储成功返回nil，否则返回错误
    实现建议:
  - 使⽤⾼效的I/O操作来写⼊数据
  - 考虑数据的完整性校验（如计算MD5）
  - 实现数据的冗余存储或纠删码
  - 考虑磁盘空间管理和数据均衡
*/
func (s *StorageNodeSvc) PutObject(ctx context.Context, bucketName, objectName string, data io.Reader, metadata map[string]string) (string, error) {
	client, err := config.ConfigDetail.OssConfig.NewOssClient()
	if err != nil {
		return "", err
	}
	bucket, err := client.Bucket(bucketName)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	// 设置分片大小（单位：字节），指定5MB为分片大小。
	partSize := int64(5 * 1024 * 1024)

	// 调用分片上传函数。
	if err := uploadMultipart(bucket, objectName, data, partSize, metadata); err != nil {
		log.Fatalf("Failed to upload multipart: %v", err)
	}
	props, err := bucket.GetObjectDetailedMeta(objectName)
	return strings.Trim(props.Get("ETag"), "\""), nil
}

// 分片上传函数
func uploadMultipart(bucket *oss.Bucket, objectName string, file io.Reader, partSize int64, metadata map[string]string) error {
	size, err := strconv.ParseInt(metadata["size"], 10, 64)
	if err != nil {
		return err
	}
	// 将本地文件分片
	chunks, err := SplitFileByPartSize(size, partSize)
	if err != nil {
		return fmt.Errorf("failed to split file into chunks: %w", err)
	}

	// 步骤1：初始化一个分片上传事件。
	imur, err := bucket.InitiateMultipartUpload(objectName)
	if err != nil {
		return fmt.Errorf("failed to initiate multipart upload: %w", err)
	}

	// 步骤2：上传分片。
	var parts []oss.UploadPart
	for _, chunk := range chunks {
		part, err := bucket.UploadPart(imur, file, chunk.Size, chunk.Number)
		if err != nil {
			// 如果上传某个部分失败，尝试取消整个上传任务。
			if abortErr := bucket.AbortMultipartUpload(imur); abortErr != nil {
				log.Printf("Failed to abort multipart upload: %v", abortErr)
			}
			return fmt.Errorf("failed to upload part: %w", err)
		}
		parts = append(parts, part)
	}

	// 指定Object的读写权限为私有，默认为继承Bucket的读写权限。
	objectAcl := oss.ObjectACL(oss.ACLPrivate)

	// 步骤3：完成分片上传。
	_, err = bucket.CompleteMultipartUpload(imur, parts, objectAcl)
	if err != nil {
		// 如果完成上传失败，尝试取消上传。
		if abortErr := bucket.AbortMultipartUpload(imur); abortErr != nil {
			log.Printf("Failed to abort multipart upload: %v", abortErr)
		}
		return fmt.Errorf("failed to complete multipart upload: %w", err)
	}

	log.Printf("Multipart upload completed successfully.")
	return nil
}

// SplitFileByPartSize splits big file into parts by the size of parts.
// Splits the file by the part size. Returns the FileChunk when error is nil.
func SplitFileByPartSize(fileSize int64, chunkSize int64) ([]oss.FileChunk, error) {
	if chunkSize <= 0 {
		return nil, errors.New("chunkSize invalid")
	}
	var chunkN = fileSize / chunkSize
	if chunkN >= 10000 {
		return nil, errors.New("Too many parts, please increase part size")
	}

	var chunks []oss.FileChunk
	var chunk = oss.FileChunk{}
	for i := int64(0); i < chunkN; i++ {
		chunk.Number = int(i + 1)
		chunk.Offset = i * chunkSize
		chunk.Size = chunkSize
		chunks = append(chunks, chunk)
	}

	if fileSize%chunkSize > 0 {
		chunk.Number = len(chunks) + 1
		chunk.Offset = int64(len(chunks)) * chunkSize
		chunk.Size = fileSize % chunkSize
		chunks = append(chunks, chunk)
	}

	return chunks, nil
}

func (s *StorageNodeSvc) GetObject(ctx context.Context, bucketName, objectName string) (io.ReadCloser, types.ObjectInfo, error) {
	client, err := config.ConfigDetail.OssConfig.NewOssClient()
	ObjectInfo := types.ObjectInfo{}
	if err != nil {
		return nil, types.ObjectInfo{}, err
	}
	// 获取 Bucket
	bucket, err := client.Bucket(bucketName)
	if err != nil {
		return nil, ObjectInfo, err
	}

	// 获取对象
	object, err := bucket.GetObject(objectName)
	if err != nil {
		return nil, ObjectInfo, err
	}

	// 获取对象元数据
	var size int64
	var LastModified time.Time
	headers, _ := bucket.GetObjectMeta(objectName)
	if contentLengthStr := headers.Get("Content-Length"); contentLengthStr != "" {
		size, err = strconv.ParseInt(contentLengthStr, 10, 64)
	}
	if lastModifiedStr := headers.Get("Last-Modified"); lastModifiedStr != "" {
		LastModified, err = time.Parse(time.RFC1123, lastModifiedStr)
	}
	objectInfo := types.ObjectInfo{
		Name:         objectName,
		Size:         size,
		ETag:         strings.Trim(headers.Get("ETag"), "\""),
		LastModified: LastModified,
	}

	return object, objectInfo, nil
}
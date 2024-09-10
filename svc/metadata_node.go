package svc

import (
	"context"
	"distributed-object-storage/config"
	"distributed-object-storage/pkg/db/dao"
	"distributed-object-storage/types"
	"fmt"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"strconv"
	"strings"
	"time"
)

type MetadataNode interface {
	CreateObjectMetadata(ctx context.Context, meta types.ObjectMetadata) error
	GetObjectMetadata(ctx context.Context, bucketName, objectName string) (types.ObjectMetadata, error)
	UpdateObjectMetadata(ctx context.Context, meta types.ObjectMetadata) error
	DeleteObjectMetadata(ctx context.Context, bucketName, objectName string) error
	CreateBucket(ctx context.Context, bucketName string) error
	DeleteBucket(ctx context.Context, bucketName string) error
	ListBuckets(ctx context.Context, prefix string, maxKeys int) ([]types.BucketInfo, error)
	ListObjects(ctx context.Context, bucketName string, prefix string, maxKeys int) ([]types.ObjectInfo, error)
	PutObjectVersion(ctx context.Context, meta types.ObjectMetadata) error
	GetObjectVersions(ctx context.Context, bucketName, objectName string) ([]types.ObjectMetadata, error)
	InitiateMultipartUpload(ctx context.Context, bucketName, objectName string) (string, error)
	CompleteMultipartUpload(ctx context.Context, bucketName, objectName, uploadID string, parts []types.CompletedPart) error
	RecordObjectCopy(ctx context.Context, sourceBucket, sourceObject, destBucket, destObject string) error
	RecordObjectMigration(ctx context.Context, bucketName, objectName string, fromNode, toNode string) error
}

type MetadataSvc struct {
	MetaDataDao *dao.MetadataNode
}

func NewMetadataSvc(s *dao.S) *MetadataSvc {
	return &MetadataSvc{
		MetaDataDao: &s.MetadataNode,
	}
}

func (m *MetadataSvc) CreateObjectMetadata(ctx context.Context, meta types.ObjectMetadata) error {
	//TODO implement me
	panic("implement me")
}

func (m *MetadataSvc) GetObjectMetadata(ctx context.Context, bucketName, objectName string) (types.ObjectMetadata, error) {
	res := types.ObjectMetadata{}
	ossConfig := config.ConfigDetail.OssConfig

	// 创建OSS客户端
	client, err := oss.New(ossConfig.Endpoint, ossConfig.AK, ossConfig.SK)
	if err != nil {
		return res, fmt.Errorf("failed to create OSS client: %v", err)
	}

	// 获取bucket实例
	bucket, err := client.Bucket(bucketName)
	if err != nil {
		return res, fmt.Errorf("failed to get bucket: %v", err)
	}

	// 获取对象详细元数据
	props, err := bucket.GetObjectDetailedMeta(objectName)
	if err != nil {
		return res, fmt.Errorf("failed to get object metadata: %v", err)
	}

	// 提取并处理元数据
	res.ETag = strings.Trim(props.Get("ETag"), "\"")

	// 解析LastModified时间
	if lastModifiedStr := props.Get("Last-Modified"); lastModifiedStr != "" {
		lastModified, err := time.Parse(time.RFC1123, lastModifiedStr)
		if err != nil {
			return res, fmt.Errorf("failed to parse Last-Modified date: %v", err)
		}
		res.LastModified = lastModified.Local()
	}

	res.ContentType = props.Get("Content-Type")
	if contentLengthStr := props.Get("Content-Length"); contentLengthStr != "" {
		if size, err := strconv.ParseInt(contentLengthStr, 10, 64); err == nil {
			res.Size = size
		} else {
			return res, fmt.Errorf("failed to parse Content-Length: %v", err)
		}
	}

	res.BucketName = bucketName
	res.ObjectName = objectName
	res.VersionID = props.Get("x-oss-version-id")

	return res, nil
}

func (m *MetadataSvc) UpdateObjectMetadata(ctx context.Context, meta types.ObjectMetadata) error {
	//TODO implement me
	panic("implement me")
}

func (m *MetadataSvc) DeleteObjectMetadata(ctx context.Context, bucketName, objectName string) error {
	//TODO implement me
	panic("implement me")
}

func (m *MetadataSvc) CreateBucket(ctx context.Context, bucketName string) error {
	client, err := config.ConfigDetail.OssConfig.NewOssClient()
	if err != nil {
		return fmt.Errorf("failed to create OSS client: %v", err)
	}
	return client.CreateBucket(bucketName, oss.StorageClass(oss.StorageIA), oss.ACL(oss.ACLPublicRead), oss.RedundancyType(oss.RedundancyZRS))
}

func (m *MetadataSvc) DeleteBucket(ctx context.Context, bucketName string) error {
	client, err := config.ConfigDetail.OssConfig.NewOssClient()
	if err != nil {
		return fmt.Errorf("failed to create OSS client: %v", err)
	}
	return client.DeleteBucket(bucketName)
}

func (m *MetadataSvc) ListBuckets(ctx context.Context, prefix string, maxKeys int) ([]types.BucketInfo, error) {
	res := make([]types.BucketInfo, 0)
	ossConfig := config.ConfigDetail.OssConfig

	// 创建OSS客户端
	client, err := oss.New(ossConfig.Endpoint, ossConfig.AK, ossConfig.SK)
	if err != nil {
		return nil, fmt.Errorf("failed to create OSS client: %v", err)
	}
	var options []oss.Option
	if maxKeys > 0 {
		options = append(options, oss.MaxKeys(maxKeys))
	}
	if prefix != "" {
		options = append(options, oss.Prefix(prefix))
	}
	lsRes, err := client.ListBuckets(options...)
	if err != nil {
		return res, err
	}
	for _, bucket := range lsRes.Buckets {
		res = append(res, types.BucketInfo{
			Name:         bucket.Name,
			CreationDate: bucket.CreationDate,
			Location:     bucket.Location,
			StorageClass: bucket.StorageClass,
			Region:       bucket.Region,
		})
	}
	return res, nil
}

func (m *MetadataSvc) ListObjects(ctx context.Context, bucketName string, prefix string, maxKeys int) ([]types.ObjectInfo, error) {
	res := make([]types.ObjectInfo, 0)
	ossConfig := config.ConfigDetail.OssConfig

	// 创建OSS客户端
	client, err := oss.New(ossConfig.Endpoint, ossConfig.AK, ossConfig.SK)
	if err != nil {
		return nil, fmt.Errorf("failed to create OSS client: %v", err)
	}

	// 获取bucket实例
	bucket, err := client.Bucket(bucketName)
	if err != nil {
		return nil, fmt.Errorf("failed to get bucket: %v", err)
	}

	// 构建可选参数
	options := []oss.Option{}
	if maxKeys > 0 {
		options = append(options, oss.MaxKeys(maxKeys))
	}
	if prefix != "" {
		options = append(options, oss.Prefix(prefix))
	}

	// 列出对象
	lsRes, err := bucket.ListObjects(options...)
	if err != nil {
		return nil, fmt.Errorf("failed to list objects: %v", err)
	}

	// 处理对象信息
	for _, object := range lsRes.Objects {
		res = append(res, types.ObjectInfo{
			Name:         object.Key,
			Size:         object.Size,
			ETag:         strings.Trim(object.ETag, "\""),
			LastModified: object.LastModified,
		})
	}

	return res, nil
}

func (m *MetadataSvc) PutObjectVersion(ctx context.Context, meta types.ObjectMetadata) error {
	//TODO implement me
	panic("implement me")
}

func (m *MetadataSvc) GetObjectVersions(ctx context.Context, bucketName, objectName string) ([]types.ObjectMetadata, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MetadataSvc) InitiateMultipartUpload(ctx context.Context, bucketName, objectName string) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MetadataSvc) CompleteMultipartUpload(ctx context.Context, bucketName, objectName, uploadID string, parts []types.CompletedPart) error {
	//TODO implement me
	panic("implement me")
}

func (m *MetadataSvc) RecordObjectCopy(ctx context.Context, sourceBucket, sourceObject, destBucket, destObject string) error {
	//TODO implement me
	panic("implement me")
}

func (m *MetadataSvc) RecordObjectMigration(ctx context.Context, bucketName, objectName string, fromNode, toNode string) error {
	//TODO implement me
	panic("implement me")
}

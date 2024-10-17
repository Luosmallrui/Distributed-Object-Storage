package svc

import (
	"context"
	"distributed-object-storage/config"
	"distributed-object-storage/pkg/db/dao"
	"distributed-object-storage/pkg/minIo"
	"distributed-object-storage/types"
	"fmt"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/minio/minio-go/v7"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

var regions = []string{
	"us-east-1",
	"us-east-2",
	"us-west-1",
	"us-west-2",
	"ca-central-1",
	"eu-west-1",
	"eu-west-2",
	"eu-west-3",
	"eu-central-1",
	"eu-north-1",
	"ap-east-1",
	"ap-south-1",
	"ap-southeast-1",
	"ap-southeast-2",
	"ap-northeast-1",
	"ap-northeast-2",
	"ap-northeast-3",
	"me-south-1",
	"sa-east-1",
	"us-gov-west-1",
	"us-gov-east-1",
	"cn-north-1",
	"cn-northwest-1",
}

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
	client := minIo.GetMinioClient("127.0.0.1:9000")
	if client == nil {
		return fmt.Errorf("minio client is nil")
	}
	region := regions[rand.Intn(len(regions))]
	return client.MinioCore.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{
		Region:        region,
		ObjectLocking: false,
	})
}

func (m *MetadataSvc) DeleteBucket(ctx context.Context, bucketName string) error {
	client := minIo.GetMinioClient("127.0.0.1:9000")
	if client == nil {
		return fmt.Errorf("minio client is nil")
	}
	return client.MinioCore.RemoveBucket(ctx, bucketName)
}

func (m *MetadataSvc) ListBuckets(ctx context.Context, prefix string, maxKeys int) ([]types.BucketInfo, error) {
	client := minIo.GetMinioClient("127.0.0.1:9000")
	res := make([]types.BucketInfo, 0)
	buckets, err := client.MinioCore.ListBuckets(ctx)
	if err != nil {
		return res, err
	}
	for _, bucket := range buckets {
		region, _ := client.MinioCore.GetBucketLocation(ctx, bucket.Name)
		bucketInfo := types.BucketInfo{
			Name:         bucket.Name,
			CreationDate: bucket.CreationDate,
			Owner:        "",
			Location:     "127.0.0.1:9000",
			StorageClass: "",
			Region:       region,
		}
		res = append(res, bucketInfo)
	}
	return res, nil
}

func (m *MetadataSvc) ListObjects(ctx context.Context, bucketName string, prefix string, maxKeys int) ([]types.ObjectInfo, error) {
	//res := make([]types.ObjectInfo, 0)
	client := minIo.GetMinioClient("127.0.0.1:9000")

	if maxKeys <= 0 {
		maxKeys = 100 // 设置默认值
	}
	var allObjects []minio.ObjectInfo
	continuationToken := ""
	for {
		result, err := client.MinioCore.ListObjectsV2(bucketName, prefix, "", continuationToken, "", maxKeys)
		if err != nil {
			return nil, err
		}

		allObjects = append(allObjects, result.Contents...)
		// 检查是否还有更多对象
		if len(result.Contents) < maxKeys {
			break // 没有更多对象了
		}
		continuationToken = result.NextContinuationToken // 更新 continuationToken
	}
	res := make([]types.ObjectInfo, 0)
	for _, object := range allObjects {
		objectInfo := types.ObjectInfo{
			Name:         object.Key,
			ETag:         strings.Trim(object.ETag, "\""),
			Size:         object.Size,
			LastModified: object.LastModified,
			StorageClass: object.StorageClass,
		}
		if len(object.Metadata) == 0 {
			objectInfo.Header = make(map[string][]string)
		}
		res = append(res, objectInfo)
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

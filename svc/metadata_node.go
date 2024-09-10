package svc

import (
	"context"
	"distributed-object-storage/pkg/db/dao"
	"distributed-object-storage/types"
	"fmt"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

type MetadataNode interface {
	CreateObjectMetadata(ctx context.Context, meta types.ObjectMetadata) error
	GetObjectMetadata(ctx context.Context, bucketName, objectName string) (types.ObjectMetadata, error)
	UpdateObjectMetadata(ctx context.Context, meta types.ObjectMetadata) error
	DeleteObjectMetadata(ctx context.Context, bucketName, objectName string) error
	CreateBucket(ctx context.Context, bucketName string, owner string) error
	DeleteBucket(ctx context.Context, bucketName string) error
	ListBuckets(ctx context.Context, owner string) ([]types.BucketInfo, error)
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
	client, err := oss.New("",
		"cfg.AK",
		"cfg.SK")
	if err != nil {
		return res, err
	}
	bucket, err := client.Bucket(bucketName)
	if err != nil {
		return res, err
	}
	props, err := bucket.GetObjectDetailedMeta("exampledir/exampleobject.txt")
	if err != nil {
		return res, err
	}
	fmt.Println(props)
	res.ObjectName = ""
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

func (m *MetadataSvc) CreateBucket(ctx context.Context, bucketName string, owner string) error {
	//TODO implement me
	panic("implement me")
}

func (m *MetadataSvc) DeleteBucket(ctx context.Context, bucketName string) error {
	//TODO implement me
	panic("implement me")
}

func (m *MetadataSvc) ListBuckets(ctx context.Context, owner string) ([]types.BucketInfo, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MetadataSvc) ListObjects(ctx context.Context, bucketName string, prefix string, maxKeys int) ([]types.ObjectInfo, error) {
	//TODO implement me
	panic("implement me")
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

package svc

import (
	"context"
	"distributed-object-storage/types"
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

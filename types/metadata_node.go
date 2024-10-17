package types

import (
	"io"
	"time"
)

// ObjectMetadata 定义了对象的元数据结构。
type ObjectMetadata struct {
	BucketName   string    `json:"bucket_name"`   //对象所属的桶名称
	ObjectName   string    `json:"object_name"`   //对象的名称
	Size         int64     `json:"size"`          //对象的⼤⼩（字节）
	ContentType  string    `json:"content_type"`  // 对象的内容类型
	ETag         string    `json:"e_tag"`         // 对象的 ETag （通常是内容的 MD5 哈希）
	LastModified time.Time `json:"last_modified"` //对象最后修改时间
	StorageNodes []string  `json:"storage_nodes"` // 存储该对象的节点列表
	VersionID    string    `json:"version_id"`    // 对象的版本 ID （如果启⽤了版本控制）
	IsLatest     bool      `json:"is_latest"`     // 是否是最新版本
}

// BucketInfo 定义了桶的基本信息
type BucketInfo struct {
	Name         string    `json:"name"`          //桶的名称
	CreationDate time.Time `json:"creation_date"` // 桶的创建时间
	Owner        string    `json:"owner"`         // 桶的所有者
	Location     string    `json:"location"`      // Bucket datacenter
	StorageClass string    `json:"storage_class"` // Bucket storage class
	Region       string    `json:"region"`        // Bucket region
}

// ObjectInfo 定义了对象的基本信息，通常⽤于列出对象时。
type ObjectInfo struct {
	Name         string              `json:"name"`          //对象的名称
	Size         int64               `json:"size"`          //对象的⼤⼩（字节）
	ETag         string              `json:"etag"`          // 对象的 ETag
	LastModified time.Time           `json:"last_modified"` //对象最后修改时间
	Header       map[string][]string `json:"hear"`
	StorageClass string              `json:"storage_class"`
}

// CompletedPart 定义了已完成上传的分⽚信息。
type CompletedPart struct {
	PartNumber int    //分⽚的编号
	ETag       string //分片的Etag
}

type GetObjectMetadataReq struct {
	BucketName string `json:"bucket_name" form:"bucket_name" `
	ObjectName string `json:"object_name" form:"object_name" `
	FilePath   string `json:"file_path" form:"file_path" `
}

type ListObjectMetadataReq struct {
	BucketName string `json:"bucket_name" form:"bucket_name" `
	Prefix     string `json:"prefix" form:"prefix" `
	MaxKeys    int    `json:"max_keys" form:"max_keys" `
}

type ListBucketReq struct {
	Prefix  string `json:"prefix" form:"prefix" `
	MaxKeys int    `json:"max_keys" form:"max_keys" `
}

type CreateBucketReq struct {
	BucketName string `json:"bucket_name" form:"bucket_name" `
}

type GetObjectReq struct {
	BucketName string        `json:"bucket_name" form:"bucket_name" `
	FileReader io.ReadCloser `json:"file_reader" form:"file_reader" `
	ObjectInfo ObjectInfo    `json:"object_info" form:"object_info" `
}

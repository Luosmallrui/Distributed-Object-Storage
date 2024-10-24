package types

import (
	"github.com/minio/minio-go/v7"
	"sync"
	"time"
)

var UploadTasks = struct {
	sync.RWMutex
	Tasks map[string]*UploadStatus
	err   error
}{
	Tasks: make(map[string]*UploadStatus),
}

type PartInfo struct {
	PartNumber   int       //分片的编号
	Size         int64     //分片的大小
	ETag         string    // 对象的 ETag （通常是内容的 MD5 哈希）
	LastModified time.Time //对象最后修改时间
}

// DiskUsage 定义了存储节点的磁盘使⽤情况。
type DiskUsage struct {
	TotalSpace      int64   // 总存储空间（字节）
	UsedSpace       int64   // 已使⽤的存储空间（字节）
	AvailableSpace  int64   //可⽤存储空间（字节）
	UsagePercentage float64 // 使⽤率（百分⽐）
}

type UploadReq struct {
	BucketName string `json:"bucket_name" form:"bucket_name" `
	ObjectName string `json:"object_name" form:"object_name" `
	FilePath   string `json:"file_path" form:"file_path" `
}

type UploadStatus struct {
	IsPaused       bool
	IsCanceled     bool
	UploadID       string
	Mutex          sync.Mutex
	CompletedParts []minio.CompletePart
	CurrentPart    int
	Error          error
}

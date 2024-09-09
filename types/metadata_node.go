package types

import "time"

// ObjectMetadata 定义了对象的元数据结构。
type ObjectMetadata struct {
	BucketName   string    //对象所属的桶名称
	ObjectName   string    //对象的名称
	Size         int64     //对象的⼤⼩（字节）
	ContentType  string    // 对象的内容类型
	ETag         string    // 对象的 ETag （通常是内容的 MD5 哈希）
	LastModified time.Time //对象最后修改时间
	StorageNodes []string  // 存储该对象的节点列表
	VersionID    string    // 对象的版本 ID （如果启⽤了版本控制）
	IsLatest     bool      // 是否是最新版本
}

// BucketInfo 定义了桶的基本信息
type BucketInfo struct {
	Name         string    //桶的名称
	CreationDate time.Time // 桶的创建时间
	Owner        string    // 桶的所有者
}

// ObjectInfo 定义了对象的基本信息，通常⽤于列出对象时。
type ObjectInfo struct {
	Name         string    //对象的名称
	Size         int64     //对象的⼤⼩（字节）
	ETag         string    // 对象的 ETag
	LastModified time.Time //对象最后修改时间
}

// CompletedPart 定义了已完成上传的分⽚信息。
type CompletedPart struct {
	PartNumber int    //分⽚的编号
	ETag       string //分片的Etag
}

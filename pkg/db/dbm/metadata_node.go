package dbm

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

func (obj *ObjectMetadata) TableName() string {
	return "object_metadata"
}

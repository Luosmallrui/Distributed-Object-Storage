package types

import (
	"crypto/tls"
	"io"
	"time"
)

// Request 表⽰客⼾端的请求。
type Request struct {
	Method      string            //Http方法
	Path        string            //请求路径
	Headers     map[string]string //请求头
	QueryParams map[string]string //查询参数
	Body        io.Reader         //请求体
}

// Response 表⽰对客⼾端请求的响应
type Response struct {
	StatusCode int               //Http状态码
	Headers    map[string]string //响应头
	Body       io.Reader         //响应体
}

// AuthInfo 包含请求的⾝份认证信息
type AuthInfo struct {
	UserID    string    //用户Id
	Roles     []string  //用户角色
	ExpiresAt time.Time //认证过期时间
}

// NodeInfo 包含节点的基本信息。
type NodeInfo struct {
	ID       string   // 节点唯⼀标识符
	Type     NodeType //节点类型（存储节点、元数据节点等）
	Address  string   // 节点地址
	Capacity int64    //节点容量（对于存储节点）
}

// NodeType 表⽰节点的类型。
type NodeType int

const (
	StorageNode NodeType = iota
	MetadataNode
)

// PerformanceMetrics  包含性能监控的指标
type PerformanceMetrics struct {
	RequestRate       float64   //每秒请求数
	ErrorRate         float64   //平均延迟（毫秒）
	AverageLatency    float64   //错误率
	ActiveConnections int       // 活跃连接数
	CPUUsage          float64   //CPU 使⽤率
	MemoryUsage       float64   //内存使⽤率
	Timestamp         time.Time // 指标收集时间
}

// GatewayConfig 网关的配置信息
type GatewayConfig struct {
	ListenAddress     string        // 监听地址
	TLSConfig         *tls.Config   // TLS配置
	MaxConcurrency    int           // 最⼤并发请求数
	IdleTimeout       time.Duration // 空闲连接超时时间
	ReadTimeout       time.Duration // 读取超时时间
	WriteTimeout      time.Duration // 写⼊超时时间
	CacheSize         int64         // 缓存⼤⼩（字节）
	AuthenticationURL string        // ⾝份认证服务 URL
}

// SystemStatus 表⽰整个系统的状态信息
type SystemStatus struct {
	TotalNodes     int       //总节点数
	HealthyNodes   int       //健康节点数
	TotalCapacity  int64     //总存储容量（字节）
	UsedCapacity   int64     //已使⽤存储容量（字节）
	SystemLoad     float64   //系统负载
	OverallHealth  string    //整体健康状态（如 "Good", "Warning", "Critical"）
	LastUpdateTime time.Time //最后更新时间
}

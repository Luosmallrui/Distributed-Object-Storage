package syncer

import (
	"context"
	"distributed-object-storage/etcd"
	"distributed-object-storage/pkg/log"
	"distributed-object-storage/pkg/minIo"
	"net/http"
	"time"
)

type NodeHealthCheck struct {
}

func (c *NodeHealthCheck) Interval() time.Duration {
	return time.Second * 2
}

func (c *NodeHealthCheck) BeforeStart(ctx context.Context) {
	time.Sleep(time.Second * 1)
	return
}

func (c *NodeHealthCheck) RunOnce() bool {
	return false
}

func (c *NodeHealthCheck) EnvIsolation() bool {
	return false
}

type NodeHealthCheckSyncer struct {
	NodeHealthCheck
}

func NewNodeHealthCheckSyncer() *NodeHealthCheckSyncer {
	return &NodeHealthCheckSyncer{}
}

func (g NodeHealthCheckSyncer) Sync(ctx context.Context) error {
	servicesList, _ := minIo.GetStorageNodeList()
	for _, service := range servicesList {
		resp, err := http.Get(service.Value + "/minio/health/live")
		if err != nil {
			log.Errorf("%s is unhealthy: %v\n", service, err)
			continue
		}
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			log.Infof("%s is healthy, reinjecting etcd node.\n", service)
			client := etcd.GetEtcdClient()
			if client != nil {
				return err
			} else {
				client.Put(ctx, "minio/"+service.Key, service.Value)
			}
		} else {
			log.Errorf("%s is unhealthy: %d\n", service, resp.StatusCode)
		}
	}
	return nil
}

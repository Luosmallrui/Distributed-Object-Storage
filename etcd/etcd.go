package etcd

import clientv3 "go.etcd.io/etcd/client/v3"

func GetEtcdClient() *clientv3.Client {
	etcdClient, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"http://0.0.0.0:2379"},
	})
	if err != nil {
		return nil
	}
	return etcdClient
}

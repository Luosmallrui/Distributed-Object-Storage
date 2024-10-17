package minIo

import (
	"context"
	"distributed-object-storage/types"
	"fmt"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	client "go.etcd.io/etcd/client/v3"
	"strings"
)

func GetStorageNodeList() ([]types.KvStorage, error) {
	storageNodeList := make([]types.KvStorage, 0)
	etcdClient, err := client.New(client.Config{
		Endpoints: []string{"http://0.0.0.0:2379"},
	})
	if err != nil {
		return storageNodeList, err
	}
	resp, err := etcdClient.Get(context.Background(), "minio/", client.WithPrefix())
	if err != nil {
		return storageNodeList, err
	}
	for _, node := range resp.Kvs {
		suffix := strings.TrimPrefix(string(node.Key), "minio/")
		storageNodeList = append(storageNodeList, types.KvStorage{Key: suffix, Value: string(node.Value)})
	}
	return storageNodeList, nil
}

func GetSeverInfo() int {
	endpoint := "0.0.0.0:9001"
	accessKeyID := "root"
	secretAccessKey := "rootroot"

	// Initialize minio client object.
	_, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: true,
	})
	if err != nil {
		fmt.Println(err)
	}
	return 1
}

package etcdserver

import (
	"context"
	"fmt"

	clientv3 "go.etcd.io/etcd/client/v3"
)

func EtcdGetServerListAddrs(serviceName string) ([]string, error) {
	serverListAddrs := make([]string, 0)

	// 1、创建一个etcd client
	cli, err := clientv3.New(defaultEtcdConfig)
	if err != nil {
		errs := fmt.Errorf("etcd client connect failed: %v", err)
		return serverListAddrs, errs
	}
	defer cli.Close()

	// 获取服务名下的所有 key-value 对

	resp, err := cli.Get(context.Background(), serviceName, clientv3.WithPrefix())
	if err != nil {
		errs := fmt.Errorf("etcd Get serverlistaddr failed: %v", err)
		return serverListAddrs, errs
	}

	for _, kv := range resp.Kvs {
		serverListAddrs = append(serverListAddrs, string(kv.Value))
	}

	//
	return serverListAddrs, nil
}

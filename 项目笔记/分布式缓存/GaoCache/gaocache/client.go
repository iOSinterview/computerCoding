package gaocache

import (
	"context"
	"fmt"
	"gaocache/etcdserver"
	pb "gaocache/gaocachepb"
	"log"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

// client 模块实现gaocache访问其他远程节点 从而获取缓存的能力

type client struct {
	serverName string // 服务名称 gaocache
	name       string // 服务地址 ip:port
}

func NewClient(serviceName string, name string) *client {
	if serviceName == "" {
		serviceName = ServiceName
	}

	return &client{
		serverName: serviceName,
		name:       "",
	}
}

// Fetch 从remote peer获取对应缓存值
func (c *client) Fetch(group string, key string) ([]byte, error) {
	// 1、创建一个etcd client
	cli, err := clientv3.New(defaultEtcdConfig)
	if err != nil {
		log.Fatalf("etcd client connect failed: %v", err)
		return nil, err
	}
	defer cli.Close()
	// 2、发现服务，获取与grpc服务的连接
	conn, serverAddr, err := etcdserver.EtcdDialDiscover(cli, c.serverName)
	if err != nil {
		return nil, err
	}
	c.name = serverAddr
	defer conn.Close()
	// 创建一个 gRPC 客户端的服务端点 c，用于调用 gRPC 服务端的函数。
	grpcClient := pb.NewGaocacheClient(conn)
	// 执行RPC调用并打印收到的响应数据
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	// 调用远程服务
	resp, err := grpcClient.Get(ctx, &pb.GetRequest{
		Group: group,
		Key:   key,
	})
	if err != nil {
		return nil, fmt.Errorf("could not get %s/%s from peer %s", group, key, c.name)
	}
	// 返回获得的值
	return resp.GetValue(), nil
}

// 测试Client是否实现了Fetcher接口
var _ Fetcher = (*client)(nil)

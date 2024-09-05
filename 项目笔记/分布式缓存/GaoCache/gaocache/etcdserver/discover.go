package etcdserver

import (
	"fmt"
	"math/rand"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/resolver"
	"google.golang.org/grpc"
)

// EtcdDiscover 向grpc请求一个服务
// 通过提供一个etcd client和service name即可获得Connection
func EtcdDialDiscover(cli *clientv3.Client, serviceName string) (*grpc.ClientConn, string, error) {
	// etcd解析器
	etcdResolver, err := resolver.NewBuilder(cli)
	if err != nil {
		return nil, "", err
	}
	// 先区etcd中获取服务地址，如果有就建立连接，没有就返回
	resp, _ := cli.Get(cli.Ctx(), serviceName, clientv3.WithPrefix())
	if resp.Count == 0 {
		err := fmt.Errorf("client etcd get serverAddr faild,server not exit in etcd")
		return nil, "", err
	}

	// 自己实现负载均衡
	serverListAddrs := make([]string, 0)
	for _, kv := range resp.Kvs {
		serverListAddrs = append(serverListAddrs, string(kv.Value))
	}

	// 随机选取一个，其实采用轮询会好一点（自己实现）
	// 创建一个新的随机数生成器，使用当前时间的纳秒数作为种子
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	// 生成一个随机整数（0到99之间）
	randIdx := rng.Intn(len(serverListAddrs))
	serverAddr := serverListAddrs[randIdx]
	// 建立到grpc服务的连接
	grpcConn, err := grpc.Dial(
		serverAddr,                       // 服务地址
		grpc.WithResolvers(etcdResolver), // etcd解析器解析服务地址
		grpc.WithInsecure(),              // 连接不安全，不适用TLS加密
		grpc.WithBlock(),                 // 连接调用阻塞
	)
	return grpcConn, serverAddr, err
}

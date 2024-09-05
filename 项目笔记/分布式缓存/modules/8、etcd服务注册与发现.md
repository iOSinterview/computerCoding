## 8、etcd服务注册与发现

---

为什么需要etcd服务注册与发现呢？现在我们是一个分布式缓存系统，存在着多个服务节点，那么客户端再获取连接时，怎么直到有哪些节点再提供服务呢？

比如说有`8001、8002、8003、8004`四个服务节点，本来四个都在运行，但是8004节点突然宕机停止了服务了？而此时恰好选择了8004进行服务，这就会造成请求不到服务，即客户端感知不到哪些节点正在实时运行服务。

![gaocache_serverport](https://s2.loli.net/2024/09/02/uQKzb5V1pqh4BnU.jpg)

因此需要一个中心，每次一个节点启动服务，就将自己的服务信息（服务名：ip：port）注册到中心，并且中心对其进行健康检测（心跳：KeepAlive），而客户端请求服务的时候，向中心获取服务，然后更新自己的一致性哈希环，负载均衡选择服务节点。

etcd 提供了而服务注册与发现的功能，如下图所示。

![/images/grpc  基于 etcd 服务发现/完整整体逻辑.png](https://s2.loli.net/2024/08/31/Nn8QwxzKARSsTCJ.png)

使用 etcd 进行服务注册与服务发现的流程通常涉及以下步骤：

1. **启动 etcd 服务**：
   - 在分布式环境中，首先需要启动 etcd 服务来提供服务注册与发现的功能。etcd 是一个分布式键值存储系统，常用于服务发现和配置共享。
2. **在 gRPC Server 中集成 etcd 客户端**：
   - 在 gRPC Server 的代码中，需要引入 etcd 的 Go 客户端库，例如 `go.etcd.io/etcd/client/v3`，用于与 etcd 服务进行通信。
3. **服务注册**：
   - 当 gRPC Server 启动时，可以将其地址和其他相关信息注册到 etcd 中，以便客户端能够发现该服务。这通常涉及向 etcd 写入键值对，其中键是服务名称，值是 gRPC Server 的地址。
4. **服务发现**：
   - 客户端在需要调用 gRPC 服务时，首先通过 etcd 客户端查询 etcd 中注册的服务信息，获取可用的 gRPC Server 地址。
5. **与 gRPC Server 通信**：
   - 客户端使用获取到的 gRPC Server 地址与 gRPC Server 建立连接，并发送 gRPC 请求。
6. **处理服务下线和变更**：
   - gRPC Server 在关闭时应该从 etcd 中注销服务，以避免客户端继续尝试连接到已下线的服务。同时，在服务实例发生变更（例如新增或移除）时，客户端需要及时更新服务列表。

这两篇文章不错

[快速学习安装使用etcd-CSDN博客](https://blog.csdn.net/qq_55272229/article/details/141607072)

[基于 etcd 实现 grpc 服务注册与发现 - 知乎 (zhihu.com)](https://zhuanlan.zhihu.com/p/623998314)

## 服务注册

首先你需要下载安装好 etcd，并再终端启动etcd服务，这里就不赘述了。

使用的 `etcd`版本是 `v3.5.8`，并且为了方便，也可以同时安装`etcdkeeper`，Web UI版本，可以实时看到信息。

![image-20240902211936577](https://s2.loli.net/2024/09/02/wbTPORosCXlMBea.png)

### 1.1、启动服务

在`server`结构体中增加两个字段

```go
// server 和 Group 是解耦合的 所以server要自己实现并发控制
type server struct {
	pb.UnimplementedGaocacheServer
	addr       string // ip:port
	mu         sync.Mutex
	consHash   *consistenthash.Consistency // 一致性哈希
	clients    map[string]*client          // 每个remotePeer 对应一个client
	status     bool                        // 服务运行状态 true:runing false:stop
	stopSignal chan os.Signal              // 监听服务运行状态，如果宕机则通知etcd撤销服务
}
```

- `status`表示服务的运行状态
- stopSignal 监控系统结束信号并通知`etcd`撤销服务

接着来看启动服务主要方法`Start()`

`gaocache/server.go`

```go
var (
	ttl               = 5 * time.Second // 租约过期时间
	defaultEtcdConfig = clientv3.Config{
		Endpoints:   []string{"localhost:2379"},
		DialTimeout: ttl,
	}
	ServiceName = "gaocache"
)


```





服务端

1. 监听服务，lis, err := net.Listen("tcp", ":"+port)

2. // 注册grpcserver

     grpcServer := grpc.NewServer()

     pb.RegisterGaocacheServer(grpcServer, s)

3. 将服务注册至etcd  ：etcdserver.EtcdRegister

4. 停止服务信号的处理，收到停止信号就要取消注册，删除etcd中服务

   cli.Delete(context.Background(), key)

5. 监听etcd中前缀gaocache的节点变化，如果有变化就更新哈希，setpeers

6. 监听并处理请求：grpcServer.Serve(lis)，svr.Get()



注册端：

1. 1、创建一个etcd client

     cli, err := clientv3.New(defaultEtcdConfig)

2. 向etcd使用get服务地址，存在表示已经注册，返回

3. 不存在则进行注册，创建租约，并将服务地址注册etcd中

4. 保持心跳，keetalive，并清空返回的ch

   

客户端：

1. 创建一个etcd client

2. 发现服务，获取服务列表，先执行Get流程

3. 通过一致性哈希选择要服务的节点

   

4. 采用获取的列表进行一致性哈希表选择。

5. 返回grpc服务连接conn，根据grpc服务连接创建一个客户端的服务端点，用来获取数据

6. 执行调用



## 服务发现

修改`gaocache/client.go`的`Fetch()`

```go
// Fetch 从remote peer获取对应缓存值
func (c *client) Fetch(group string, key string) ([]byte, error) {
	// 1、创建一个etcd client
	cli, err := clientv3.New(defaultEtcdConfig)
	if err != nil {
		log.Fatalf("etcd client connect failed: %v", err)
	}
	defer cli.Close()
	// 2、发现服务，获取与grpc服务的连接
	conn, err := etcdserver.EtcdDialDiscover(cli, c.name)
	if err != nil {
		return nil, err
	}
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
```

其中，`etcdserver.EtcdDialDiscover(cli, c.name)`在`gaocache/etcdserver/discover.go`

```go
package etcdserver

import (
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/resolver"
	"google.golang.org/grpc"
)

// EtcdDiscover 向grpc请求一个服务
// 通过提供一个etcd client和service name即可获得Connection
func EtcdDialDiscover(c *clientv3.Client, service string) (*grpc.ClientConn, error) {
	// etcd解析器
	etcdResolver, err := resolver.NewBuilder(c)
	if err != nil {
		return nil, err
	}
	// 建立到grpc服务的连接
	return grpc.Dial(
		"etcd:///"+service,               // 服务地址
		grpc.WithResolvers(etcdResolver), // etcd解析器解析服务地址
		grpc.WithInsecure(),              // 连接不安全，不适用TLS加密
		grpc.WithBlock(),                 // 连接调用阻塞
	)
}
```



## 问题记录：

1、还是版本依赖问题

首先，下面这两个是有区别的，大多参考的代码给的是`client/v3`的

```go
go get go.etcd.io/etcd/clientv3			// 不要装这个，虽然我还没弄明白，但是好多依赖问题
go get go.etcd.io/etcd/client/v3		// 装这个
```

`clientv3中就没有New()方法。`

其次，grpc版本与etcd版本冲突的问题。原因 etcd3.5.8 的 release 版本要求 grpc 的版本是 v1.26.0 之前的。而此时 go.mod 里面的 google.golang.org/grpc 是 v1.59.1

```go
go: gaocache imports         go.etcd.io/etcd/clientv3 tested by         go.etcd.io/etcd/clientv3.test imports         github.com/coreos/etcd/auth imports         github.com/coreos/etcd/mvcc/backend imports         github.com/coreos/bbolt: github.com/
```

解决：在go.mod里面这样添加，然后`go mod tidy`

```go
replace github.com/coreos/bbolt v1.3.4 => go.etcd.io/bbolt v1.3.4

replace google.golang.org/grpc v1.59.1 => google.golang.org/grpc v1.26.0
```

但是，在`client`端用`grpc.Dial()`建立与grpc服务连接的时候。v1.26.0版本并没有`grpc.WithResolvers(etcdResolver)`这个函数，于是我又将grpc版本换成了v1.38.0

```go
return grpc.Dial(
		"etcd:///"+service,
		grpc.WithResolvers(etcdResolver),
		grpc.WithInsecure(),
		grpc.WithBlock(),
	)
```

```go
replace github.com/coreos/bbolt v1.3.4 => go.etcd.io/bbolt v1.3.4

replace google.golang.org/grpc v1.59.1 => google.golang.org/grpc v1.38.0
```

各种问题，弄到奔溃。。。

参考下面这位大哥的记录，我遇到的也差不多，没有及时记录。

[使用 etcd 和 grpc 遇到的版本冲突的那些事儿 | Go 技术论坛 (learnku.com)](https://learnku.com/articles/43758)
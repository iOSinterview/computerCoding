## 4、gRPC服务通信

---

分布式缓存需要节点间的通信，这里，我们建立基于`gRPC`的通信，这样，客户端可以调用远程服务端的方法，且采用`protoBuf`通信。

![gRPC_communication](https://s2.loli.net/2024/09/01/dU9ibAKnWek4yao.jpg)

## （1）定义`protoc`

`gaocache/gaocachepb/gaocache.protoc`

```go
syntax = "proto3"; 	// 版本声明，使用Protocol Buffers v3版本

package gaocachepb;		// 包名

// 指定生成的Go代码在你项目中的导入路径
option go_package = "gaocache/gaocachepb";

// 请求信息
//message MessageName {
//  FieldType fieldName = FieldNumber;
//}
message GetRequest {
  string group = 1;	// 字段唯一标识号
  string key = 2;
}

// 响应信息
message GetResponse {
  bytes value = 1;
}

// 定义服务
service PeanutCache {
  // Get方法，这里采用普通rpc
  rpc Get(GetRequest) returns (GetResponse);
}
```

运行下面命令生成`Go`语言代码

```go
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative gaocachepb/gaocache.proto
```

此时目录：

```go
│  byteview.go
│  cache.go
│  gaocache.go
│  gaocache_test.go
│  go.mod
│
├─gaocachepb
│      gaocache.pb.go
│      gaocache.proto
│      gaocache_grpc.pb.go
│
└─models
    ├─lfu
    └─lru
            lru.go
```

## （2）`server`端

`server`端，我只需要实现上面定义的`Get()`服务方法就行，并开启服务。

`gaocache/server.go`

```go
package gaocache

import (
	"context"
	"fmt"
	pb "gaocache/gaocachepb"
	"log"
)

const (
	defaultAddr = "127.0.0.1:8080"
)

// server 和 Group 是解耦合的 所以server要自己实现并发控制
type Server struct {
	pb.UnimplementedGaocacheServer
	addr string
}

// NewServer 创建cache的svr 若addr为空 则使用defaultAddr
func NewServer(addr string) (*Server, error) {
	if addr == "" {
		addr = defaultAddr
	}
	// 判断是否满足 x.x.x.x:port 的格式
	if !validPeerAddr(addr) {
		return nil, fmt.Errorf("invalid addr %s, it should be x.x.x.x:port", addr)
	}
	return &Server{addr: addr}, nil
}

// Get 实现gaocache service的Get接口
func (s *Server) Get(ctx context.Context, in *pb.GetRequest) (*pb.GetResponse, error) {
	group, key := in.GetGroup(), in.GetKey()
	resp := &pb.GetResponse{}

	log.Printf("[gaocache_svr %s] Recv RPC Request - (%s)/(%s)", s.addr, group, key)
	if key == "" {
		return resp, fmt.Errorf("key required")
	}
	g := GetGroup(group)
	if g == nil {
		return resp, fmt.Errorf("group not found")
	}
	view, err := g.Get(key)
	if err != nil {
		return resp, err
	}
	resp.Value = view.ByteSlice()
	return resp, nil
}
```

**这里，服务`Server`其实最好采用小写，只允许内部包访问，但是为了，后面这个模块的更好测试，我先采用大写，后面再改回来。**

其中，`validPeerAddr`放在`utils.go`里

`gaocache/utils.go`

```go
package gaocache

import "strings"

// 判断是否满足 x.x.x.x:port 的格式
func validPeerAddr(addr string) bool {
	token1 := strings.Split(addr, ":")
	if len(token1) != 2 {
		return false
	}
	token2 := strings.Split(token1[0], ".")
	if token1[0] != "localhost" && len(token2) != 4 {
		return false
	}
	return true
}
```

## （3）`client` 端

在`client`端，我们只需要，与gRPC 服务器建立连接，然后发送请求并接收响应即可，不需要了解`server`端是怎么实现的。

```go
package gaocache

import (
	"context"
	"fmt"
	pb "gaocache/gaocachepb"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// client 模块实现gaocache访问其他远程节点 从而获取缓存的能力

type client struct {
	name string // 服务名称 ip:port
}

func NewClient(service string) *client {
	return &client{name: service}
}

// Fetch 从remote peer获取对应缓存值
func (c *client) Fetch(group string, key string) ([]byte, error) {

	// 连接到server端
	conn, err := grpc.NewClient(c.name, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("grpc client connect failed: %v", err)
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
	fmt.Println(resp)
	if err != nil {
		return nil, fmt.Errorf("could not get %s/%s from peer %s", group, key, c.name)
	}
	// 返回获得的值
	return resp.GetValue(), nil
}
```

## （4）测试

这个模块不太好测试，用`testing`出现了许多问题，为了验证`grpc`通信的有效性，我才去了最笨的办法测试。

首先，我将`gaocache`包的文件，全部放入了一个新的目录，然后外面新建了一个`main.go`用作测试入口。

文件结构如下：

```go
│  go.mod
│  go.sum
│  main.go
│
├─gaocache
│      byteview.go
│      cache.go
│      client.go
│      gaocache.go
│      gaocache_test.go
│      server.go
│      server_test.go
│      utils.go
│
├─gaocachepb
│      gaocache.pb.go
│      gaocache.proto
│      gaocache_grpc.pb.go
│
└─models
    ├─lfu
    └─lru
            lru.go
```

其中，`main.go`代码如下：

```go
package main

import (
	"fmt"
	"gaocache/gaocache"
	"gaocache/gaocachepb"
	"log"
	"net"

	"google.golang.org/grpc"
)

func createTestSvr() (*gaocache.Group, *gaocache.Server) {
	mysql := map[string]string{
		"Tom":   "630",
		"Jack":  "589",
		"Sam":   "567",
		"Alice": "788",
	}

	g := gaocache.NewGroup("scores", 2<<10, gaocache.RetrieverFunc(
		func(key string) ([]byte, error) {
			log.Println("[Mysql] search key", key)
			if v, ok := mysql[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))

	// 随机一个端口 避免冲突
	port := 8080
	addr := fmt.Sprintf("localhost:%d", port)

	svr, err := gaocache.NewServer(addr)
	if err != nil {
		fmt.Printf("err:%v", err)
	}
	return g, svr

}

func main() {
	// 监听本地 8080 端口
	lis, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Printf("failed to listen: %v", err)
		return
	}
	s := grpc.NewServer() // 创建一个新的 gRPC 服务器实例
	_, svr := createTestSvr()
	gaocachepb.RegisterGaocacheServer(s, svr) // 注册服务至grpc
	err = s.Serve(lis)                        // 启动服务
}
```

接着，我将`client.go`和`gaocachepb`目录下的文件复制出来，重新弄了个项目。然后编写如下`main.go`

```go
package main

import (
	"context"
	"fmt"
	pb "main/gaocachepb"

	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// client 模块实现gaocache访问其他远程节点 从而获取缓存的能力

type client struct {
	name string // 服务名称 gaocache/ip:addr
}

func NewClient(service string) *client {
	return &client{name: service}
}

// Fetch 从remote peer获取对应缓存值
func (c *client) Fetch(group string, key string) ([]byte, error) {

	// 连接到server端
	conn, err := grpc.NewClient(c.name, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("grpc client connect failed: %v", err)
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
		// return nil, fmt.Errorf("could not get %s/%s from peer %s", group, key, c.name)
		return nil, err
	}
	// 返回获得的值
	return resp.GetValue(), nil
}

func main() {
	svrname := "127.0.0.1:8080"
	cli := NewClient(svrname)
	group := "scores"
	keys := []string{"Tom", "Alice", "Sam", "NotExist"}
    // 第一次查询不在缓存中，回调函数会从数据库中查
	for i := range keys {
		value, err := cli.Fetch(group, keys[i])
		if err != nil {
			fmt.Printf("err:%v\n", err)
			return
		}
		fmt.Printf("key:%v,value:%v\n", keys[i], string(value))
	}
    // 第二次查询，看看是否在缓存中
    for i := range keys {
		value, err := cli.Fetch(group, keys[i])
		if err != nil {
			fmt.Printf("err:%v\n", err)
			return
		}
		fmt.Printf("key:%v,value:%v\n", keys[i], string(value))
	}
}
```

最后，分别运行两个项目的`main.go`，结果如下：

`client`端

```go
key:Tom,value:630
key:Alice,value:788
key:Sam,value:567
err:rpc error: code = Unknown desc = NotExist not exist
key:NotExist,value:
key:Tom,value:630
key:Alice,value:788
key:Sam,value:567
err:rpc error: code = Unknown desc = NotExist not exist
```

`server`端

```go
2024/09/01 05:06:45 [gaocache_svr localhost:8080] Recv RPC Request - (scores)/(Tom)
2024/09/01 05:06:45 [Mysql] search key Tom
2024/09/01 05:06:45 [gaocache_svr localhost:8080] Recv RPC Request - (scores)/(Alice)
2024/09/01 05:06:45 [Mysql] search key Alice
2024/09/01 05:06:45 [gaocache_svr localhost:8080] Recv RPC Request - (scores)/(Sam)  
2024/09/01 05:06:45 [Mysql] search key Sam
2024/09/01 05:06:45 [gaocache_svr localhost:8080] Recv RPC Request - (scores)/(NotExist)
2024/09/01 05:06:45 [Mysql] search key NotExist
//-------以下为第二次查询结果----------------//
2024/09/01 05:06:45 [gaocache_svr localhost:8080] Recv RPC Request - (scores)/(Tom)  
2024/09/01 05:06:45 cache hit
2024/09/01 05:06:45 [gaocache_svr localhost:8080] Recv RPC Request - (scores)/(Alice)
2024/09/01 05:06:45 cache hit
2024/09/01 05:06:45 [gaocache_svr localhost:8080] Recv RPC Request - (scores)/(Sam)  
2024/09/01 05:06:45 cache hit
2024/09/01 05:06:45 [gaocache_svr localhost:8080] Recv RPC Request - (scores)/(NotExist)
2024/09/01 05:06:45 [Mysql] search key NotExist
```

可以看出，第一次由于缓存中没有值，全部会从数据库中查询返回；但是，数据库查询后会将`key-value`存入缓存，因此，第二次就会缓存命中了。






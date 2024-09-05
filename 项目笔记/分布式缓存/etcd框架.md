## `etcd`框架——高可用的分布式`key-value`存储系统

---

## `etcd`介绍

![头图](https://s2.loli.net/2024/08/30/dJZvrQYDNnkKUjR.png)

[`etcd`](https://etcd.io/)是使用Go语言开发的一个开源的、高可用的分布式`key-value`存储系统，可以用于配置共享和服务的注册和发现。

类似项目有zookeeper和consul。

etcd具有以下特点：

- 完全复制：集群中的每个节点都可以使用完整的存档
- 高可用性：Etcd可用于避免硬件的单点故障或网络问题
- 一致性：每次读取都会返回跨多主机的最新写入
- 简单：包括一个定义良好、面向用户的API（gRPC）
- 安全：实现了带有可选的客户端证书身份验证的自动化TLS
- 快速：每秒10000次写入的基准速度
- 可靠：使用Raft算法实现了强一致、高可用的服务存储目录

## `etcd` 应用场景

### 服务注册与发现

**服务发现**指服务实例向服务注册与发现中心获取的其他服务实例信息，用于进行后续的远程调用。

**服务发现**也是分布式系统中最常见的问题之一，即在同一个分布式集群中的进程或服务，要如何才能找到对方并建立连接。本质上来说，**服务发现就是想要了解集群中是否有进程在监听 udp 或 tcp 端口，并且通过名字就可以查找和连接。**要解决服务发现的问题，需要有下面三大支柱：

- 一个强一致性、高可用的服务存储目录。基于 Raft 算法的 etcd 天生就是这样一个强一致性高可用的服务存储目录
- 一种注册服务和监控服务健康状态的机制。用户可以在 etcd 中注册服务，并且对注册的服务设置 key TTL 值，定时保持服务的心跳以达到监控服务健康状态的效果。
- 一种查找和连接服务的机制。为了确保连接，我们可以在每个服务机器上都部署一个 Proxy 模式的 etcd，这样就可以确保能访问 etcd 集群的服务都能互相连接。

**服务注册**指服务实例启动的时候将自身的信息注册到服务注册与发现中心，并在运行的时候通过心跳的方式向服务注册发现中心汇报自身服务状态

![img](https://s2.loli.net/2024/08/30/6ULpAWoGmuVgS4P.png)

日常开发集群管理功能中，如果要设计可以动态调整集群大小，那么首先就要支持服务发现，就是说当一个新的节点启动时，可以将自己的信息注册给 `master`，然后让 `master` 把它加入到集群里，关闭之后也可以把自己从集群中删除。etcd 提供了很好的服务注册与发现的基础功，我们采用 etcd 来做服务发现时，可以把精力用于服务本身的业务处理上。

服务注册与发现的作用：

- **管理实例信息**。管理当前注册到**服务注册与发现中心**的服务实例元数据信息，这些信息包括服务实例的服务名，IP地址，端口号服务状态和服务描述等等信息。
- **健康检查**。服务注册与发现中心会与已经注册 ok 的微服务实例维持心跳，定期**检查注册表**中的服务是否正常在线，并且会在过程中剔除掉无效的服务实例信息。
- **提供服务发现的作用**。如一个服务需要调用服务注册与发现中心中的微服务实例，可以通过**服务注册与发现中心**获取到其具体的服务实例信息



![img](https://s2.loli.net/2024/08/31/iwxGlkb1hD4rHMV.webp)







### 配置中心

将一些配置信息放到 etcd 上进行集中管理。

这类场景的使用方式通常是这样：应用在启动的时候主动从 etcd 获取一次配置信息，同时，在 etcd 节点上注册一个 Watcher 并等待，以后每次配置有更新的时候，etcd 都会实时通知订阅者，以此达到获取最新配置信息的目的。

### 分布式锁

因为 etcd 使用 Raft 算法保持了数据的强一致性，某次操作存储到集群中的值必然是全局一致的，所以很容易实现分布式锁。锁服务有两种使用方式，一是保持独占，二是控制时序。

- **保持独占即所有获取锁的用户最终只有一个可以得到**。etcd 为此提供了一套实现分布式锁原子操作 CAS（`CompareAndSwap`）的 API。通过设置`prevExist`值，可以保证在多个节点同时去创建某个目录时，只有一个成功。而创建成功的用户就可以认为是获得了锁。
- 控制时序，即所有想要获得锁的用户都会被安排执行，但是**获得锁的顺序也是全局唯一的，同时决定了执行顺序**。etcd 为此也提供了一套 API（自动创建有序键），对一个目录建值时指定为`POST`动作，这样 etcd 会自动在目录下生成一个当前最大的值为键，存储这个新的值（客户端编号）。同时还可以使用 API 按顺序列出所有当前目录下的键值。此时这些键的值就是客户端的时序，而这些键中存储的值可以是代表客户端的编号。

![img](https://s2.loli.net/2024/08/30/hvV9QNJpi8c7SDE.png)

## 为什么选择ETCD?

1. 简单。使用 Go 语言编写部署简单；支持HTTP/JSON API，使用简单；使用 Raft 算法保证强一致性让用户易于理解。
2. etcd 默认数据一更新就进行持久化。
3. etcd 支持 SSL 客户端安全认证。

最后，etcd 作为一个年轻的项目，正在高速迭代和开发中，这既是一个优点，也是一个缺点。优点是它的未来具有无限的可能性，缺点是无法得到大项目长时间使用的检验。然而，目前 `CoreOS`、`Kubernetes`和`CloudFoundry`等知名项目均在生产环境中使用了`etcd`。

## Go语言操作 etcd

这里使用官方的[`etcd/clientv3`](https://github.com/etcd-io/etcd/tree/master/client/v3)包来连接etcd并进行相关操作。

### 安装

#### 在本地安装 etcd

首先你要本地机器上安装[Etcd](https://github.com/etcd-io/etcd/releases/tag/v3.5.15)，下载对应的版本。

![image-20240830233138559](https://s2.loli.net/2024/08/30/KDz3Yplig71PFCv.png)

解压后：

![image-20240830233546917](https://s2.loli.net/2024/08/30/fUua1zYbXtA27mr.png)

将 `etcd` 和 `etcdctl` 二进制文件路径添加到系统的环境变量中。安装完成后，可以通过命令行启动 etcd：

![image-20240830234110904](https://s2.loli.net/2024/08/30/NFgOCd6kVqB8hnY.png)

etcd 默认监听 `127.0.0.1:2379` 端口。

你可以通过 `etcdctl` 命令行工具与 etcd 交互。它可以执行各种操作，如读写键值、查看集群状态等。这里就先不展开。

#### 在Go中安装使用

```go
go get go.etcd.io/etcd/clientv3			// 不要装这个，虽然我还没弄明白，但是好多依赖问题
go get go.etcd.io/etcd/client/v3		// 装这个
```

安装后使用的过程会有一些以来问题，因为`clientv3`不支持`grpcv1.26.0`以上版本的，可以进行如下解决：

```go
go clean -modcache
go mod init 
go mod edit -replace github.com/coreos/bbolt@v1.3.4=go.etcd.io/bbolt@v1.3.4
go mod edit -require google.golang.org/grpc@v1.26.0
go mod tidy
go run main.go
```

### `put`和`get`操作

`put`命令用来设置键值对数据，`get`命令用来根据`key`获取值。

```go
package main

import (
	"context"
	"fmt"
	"time"

	"go.etcd.io/etcd/clientv3"
)

func main() {
	// 创建一个 etcd 客户端连接到本地的 etcd 服务器
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"127.0.0.1:2379"}, // etcd 服务器的端点地址。
		DialTimeout: 5 * time.Second,            // 连接超时时间
	})
	if err != nil {
		// handle error!
		fmt.Printf("connect to etcd failed, err:%v\n", err)
		return
	}
	defer cli.Close()
	fmt.Println("connect to etcd success")

	// put
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second) // 带超时的上下文
	_, err = cli.Put(ctx, "game", "nice")                                   // key-value=game-nice
	cancel()
	if err != nil {
		fmt.Printf("put to etcd failed, err:%v\n", err)
		return
	}

	// get
	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	resp, err := cli.Get(ctx, "game")
	cancel()
	if err != nil {
		fmt.Printf("get from etcd failed, err:%v\n", err)
		return
	}
	for _, ev := range resp.Kvs {
		fmt.Printf("%s:%s\n", ev.Key, ev.Value)
	}
}
```

结果：

```go
connect to etcd success
game:nice
```

### `watch`操作

`watch`用来获取未来更改的通知。

```go
package main

import (
	"context"
	"fmt"
	"time"

	"go.etcd.io/etcd/clientv3"
)

// watch demo

func main() {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"127.0.0.1:2379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		fmt.Printf("connect to etcd failed, err:%v\n", err)
		return
	}
	fmt.Println("connect to etcd success")
	defer cli.Close()
	// watch key:q1mi change
	rch := cli.Watch(context.Background(), "game") // <-chan WatchResponse
	for wresp := range rch {
		for _, ev := range wresp.Events {
			fmt.Printf("Type: %s Key:%s Value:%s\n", ev.Type, ev.Kv.Key, ev.Kv.Value)
		}
	}
}
```

将上面的代码保存编译执行，此时程序就会等待etcd中`game`这个key的变化。

例如：我们打开终端执行以下命令修改、删除、设置`game`这个key。

```go
etcdctl.exe --endpoints=http://127.0.0.1:2379 put game "wangzhe"
OK
etcdctl.exe --endpoints=http://127.0.0.1:2379 del game
1
etcdctl.exe --endpoints=http://127.0.0.1:2379 put game "chiji"
OK
```

上面的程序都能收到如下通知。

```go
connect to etcd success
Type: PUT Key:game Value:wangzhe
Type: DELETE Key:game Value:
Type: PUT Key:game Value:chiji
```

### `lease`租约

```go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.etcd.io/etcd/clientv3"
)

// etcd lease

func main() {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"127.0.0.1:2379"},
		DialTimeout: time.Second * 5,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("connect to etcd success.")
	defer cli.Close()

	// 创建一个5秒的租约
	resp, err := cli.Grant(context.TODO(), 5)
	if err != nil {
		log.Fatal(err)
	}

	// 5秒钟之后, /group:8080/ 这个key就会被移除
	_, err = cli.Put(context.TODO(), "/group:8080/", "Score", clientv3.WithLease(resp.ID))
	if err != nil {
		log.Fatal(err)
	}

	// 先get看一下
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	resp2, err := cli.Get(ctx, "/group:8080/")
	cancel()
	if err != nil {
		fmt.Printf("get from etcd failed, err:%v\n", err)
		return
	}
	for _, ev := range resp2.Kvs {
		fmt.Printf("%s:%s\n", ev.Key, ev.Value)
	}
	// 休眠10s再get下
	time.Sleep(10 * time.Second)
	ctx, cancel = context.WithTimeout(context.Background(), time.Second)
	resp2, err = cli.Get(ctx, "/group:8080/")
	cancel()
	if err != nil {
		fmt.Printf("get from etcd failed, err:%v\n", err)
		return
	}
	fmt.Println(resp2)
	for _, ev := range resp2.Kvs {
		fmt.Printf("%s:%s\n", ev.Key, ev.Value)
	}
}
```

过了10s后再去get，返回里面是空的

```go
connect to etcd success.
/group:8080/:Score
&{cluster_id:14841639068965178418 member_id:10276657743932975437 revision:17 raft_term:2  [] false 0 {} [] 0}
```

### `keepAlive`

设置心跳检测。

```go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.etcd.io/etcd/clientv3"
)

// etcd keepAlive

func main() {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"127.0.0.1:2379"},
		DialTimeout: time.Second * 5,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("connect to etcd success.")
	defer cli.Close()
	// 请求租约
	resp, err := cli.Grant(context.TODO(), 5)
	if err != nil {
		log.Fatal(err)
	}
	// put key-value
	_, err = cli.Put(context.TODO(), "/group:8080/", "Score", clientv3.WithLease(resp.ID))
	if err != nil {
		log.Fatal(err)
	}

	// the key 'foo' will be kept forever
	// 维持这个租约的存活状态，该方法返回一个通道，不断地接收租约的更新信息。
	ch, err := cli.KeepAlive(context.TODO(), resp.ID)
	if err != nil {
		log.Fatal(err)
	}
	// Create a timer to stop printing TTL after lease expires
	// 计时器，5s后触发，这里即停止打印ttl
	timer := time.NewTimer(5 * time.Second)
	// 在一个无限循环中，从通道中读取租约的更新信息，打印出租约的 TTL（生存时间）。
	for {
		select {
		case ka := <-ch:
			fmt.Println("ttl:", ka.TTL)
		case <-timer.C:
			fmt.Println("Lease expired. Stopping TTL updates.")
			return
		}
	}
}
```

结果如下：

```go
connect to etcd success.
ttl: 5
ttl: 5
ttl: 5
Lease expired. Stopping TTL updates.
```


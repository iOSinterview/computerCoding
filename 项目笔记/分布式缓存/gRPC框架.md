## gRPC框架——一款高性能通用的RPC框架

---

## 引言

​		在当今快速发展的互联网时代，**高效、可靠且跨语言的通信协议对于构建分布式系统至关重要**。gRPC，作为Google开源的一款高性能、通用的RPC（Remote Procedure Call）框架，正逐渐成为现代微服务架构中的明星技术。

## RPC框架

### 什么是RPC呢？

**RPC**，即（`Remote Procedure Call`，远程过程调用），简单来说，**客户端在不知道调用细节的情况下，调用存在于远程服务器上的摸一个对象，就像调用本地程序中的对象一样。**为了跨越这个障碍，**一种通过网络从远程服务器程序上请求服务，而不需要了解底层网络技术的协议。**

因此，它拥有以下特点：

- **RPC是协议**。既然是协议规范，就需要有人遵循规范并实现，典型的RPC框架有gRPC、Thrift、Dubbo、Hetty.
- **网络协议和网络I/O模型对其透明。**RPC的客户端认为自己是在调用本地对象，那么传输层使用的是TCP/UDP还是HTTP协议，又或者是一些其他的网络协议它就不需要关心了。
- **信息格式对其透明。**对于远程调用来说，参数会以某种信息格式传递给网络上的另外一台计算机，这个信息格式是怎样构成的，调用方是不需要关心的。
- **应该有跨语言能力。**调用方实际上也不清楚**远程服务器**的应用程序是使用什么语言运行的。那么对于调用方来说，无论服务器方使用的是什么语言，本次调用都应该成功，并且返回值也应该按照调用方程序语言所能理解的形式进行描述。

调用模型图如下：

![img](https://s2.loli.net/2024/08/29/SNud4X8OZGgUyQV.png)

### 为什么要用RPC?

当我们的系统访问量增大、业务增多时，我们会发现一台单机运行此系统已经无法承受。此时，我们可以将业务拆分成几个互不关联的应用，分别部署在各自机器上，以划清逻辑并减小压力。

但是当业务量到达一定程度上时，我们会发现有些功能很难划分，此时，**可以将公共业务逻辑抽离出来，将之组成独立的服务Service应用 。而原有的、新增的应用都可以与那些独立的Service应用 交互，以此来完成完整的业务功能。**

所以我们急需一种高效的应用程序之间的通讯手段来完成这种需求，即RPC框架。其实这也是服务化 、微服务和分布式系统架构的基础场景

## RPC原理

### RPC调用流程

要让网络通信细节对使用者透明，我们需要对通信细节进行封装，我们先看下一个RPC调用的流程涉及到哪些通信细节：

![img](https://s2.loli.net/2024/08/29/6MWzJFfovujApYI.png)

RPC的目标就是要2~8这些步骤都封装起来，让用户对这些细节透明。

1. 服务消费方（client）调用以本地调用方式调用服务；
2. client stub接收到调用后负责将方法、参数等组装成能够进行网络传输的消息体；
3. client stub找到服务地址，并将消息发送到服务端；
4. server stub收到消息后进行解码；
5. server stub根据解码结果调用本地的服务；
6. 本地服务执行并将结果返回给server stub；
7. server stub将返回结果打包成消息并发送至消费方；
8. client stub接收到消息，并进行解码；
9. 服务消费方得到最终结果。

### 如何透明化远程服务调用？

Go 标准库中自带了 `net/rpc` 包，可以用于实现基本的 RPC 功能，但在实际项目中为了更方便地进行远程服务调用，通常会选择使用更强大和灵活的 RPC 框架，比如 gRPC。

### 如何对消息进行编码和解码？

#### （1）确定消息数据结构

客户端的请求消息结构一般需要包括以下内容：

- 接口名称：在我们的例子里接口名是“HelloWorldService”，如果不传，服务端就不知道调用哪个接口了；
- 方法名：一个接口内可能有很多方法，如果不传方法名服务端也就不知道调用哪个方法；
- 参数类型&参数值：参数类型有很多，比如有bool、int、long、double、string、map、list，甚至如struct等，以及相应的参数值；
- 超时时间 + requestID（标识唯一请求id）

服务端返回的消息结构一般包括以下内容：

- 状态code + 返回值
- requestID

#### （2）序列化

**序列化**就是将数据结构或对象转换成二进制串的过程，**也就是编码的过程。**

**反序列化**是将在序列化过程中所生成的二进制串转换成数据结构或者对象的过程，**也就是解码的过程。**

比如`Protobuf`就是好的序列化方案。

### 如何发布服务？

Go常用 ETCD，服务端进行服务注册和心跳，客户端发现服务，获取机器列表。

![ETCD](https://s2.loli.net/2024/08/30/ZnkewOfVGpo1aJz.jpg)

## gRPC框架理论

### gRPC简介

gRPC是一个高性能、通用的开源RPC框架，其由Google 2015年主要面向移动应用开发并基于HTTP/2协议标准而设计，基于ProtoBuf序列化协议开发，且支持众多开发语言。

由于是开源框架，通信的双方可以进行二次开发，所以客户端和服务器端之间的通信会更加专注于业务层面的内容，减少了对由gRPC框架实现的底层通信的关注。

如下图，DATA部分即业务层面内容，下面所有的信息都由gRPC进行封装。

![img](https://s2.loli.net/2024/08/29/CKSXsR6EOtfa4Uu.png)

### gRPC特点

- 语言中立，支持多种语言；
- 基于 IDL（Interface Definition Language，接口描述语言） 文件（如 Protocol Buffers）定义服务，通过 proto3 工具生成指定语言的数据结构、服务端接口以及客户端 Stub；
- 通信协议基于标准的 HTTP/2 设计，支持双向流、消息头压缩、单 TCP 的多路复用、服务端推送等特性，这些特性使得 gRPC 在移动端设备上更加省电和节省网络流量；
- 序列化支持 ProtoBuf 和 JSON，ProtoBuf 是一种语言无关的高性能序列化框架，基于 HTTP/2 + ProtoBuf, 保障了 RPC 调用的高性能。

### gRPC基本通信流程

![gRPC_communication](https://s2.loli.net/2024/08/29/XMjkNUvag7JQCz5.jpg)

1.gRPC通信的第一步是定义IDL，即我们的接口文档（后缀为`.proto`）

2.第二步是编译`proto`文件，得到存根（`stub`）文件。

3.第三步是服务端（`gRPC Server`）实现第一步定义的接口并启动，这些接口的定义在`stub`文件里面。

4.最后一步是客户端借助`stub`文件调用服务端的函数，虽然客户端调用的函数是由服务端实现的，但是调用起来就像是**本地函数**一样。

### Protocol Buffer

ProtoBuf是一种更加灵活、高效的数据格式，与XML、JSON类似，在一些高性能且对响应速度有要求的数据传输场景非常适用。ProtoBuf在gRPC的框架中主要有三个作用：

- 定义数据结构
- 定义服务接口
- 通过序列化和反序列化，二进制传输提升传输效率。

Protocol Buffers对比JSON、XML的优点：

- 简单，体积小，数据描述文件大小只有1/10至1/3；
- 传输和解析的速率快，相比XML等，解析速度提升20倍甚至更高；
- 可编译性强。

### 基于HTTP 2.0标准设计

由于gRPC基于HTTP 2.0标准设计，带来了更多强大功能，如多路复用、二进制帧、头部压缩、推送机制。这些功能给设备带来重大益处，如节省带宽、降低TCP连接次数、节省CPU使用等。gRPC既能够在客户端应用，也能够在服务器端应用，从而以透明的方式实现两端的通信和简化通信系统的构建。

HTTP 版本分为HTTP 1.X、 HTTP 2.0，其中HTTP 1.X是当前使用最广泛的HTTP协议，**HTTP 2.0称为超文本传输协议第二代**。HTTP 1.X定义了四种与服务器交互的方式，分别为：`GET、POST、PUT、DELETE`，这些在HTTP 2.0中均保留。HTTP 2.0的新特性：

- **双向流、多路复用**
- **二进制帧**
- **头部压缩**

## gRPC示例

gRPC开发分三步

### 1、编写`.proto`文件定义服务

像许多 RPC 系统一样，**gRPC 基于定义服务的思想**，指定可以通过参数和返回类型远程调用的方法。默认情况下，gRPC 使用 [protocol buffers](https://developers.google.com/protocol-buffers)作为接口定义语言(IDL)来描述服务接口和有效负载消息的结构。可以根据需要使用其他的IDL代替。

`Protocol Buffers`是一种与语言无关，平台无关的可扩展机制，用于序列化结构化数据。使用`Protocol Buffers`可以一次定义结构化的数据，然后可以使用特殊生成的源代码轻松地在各种数据流中使用各种语言编写和读取结构化数据。

关于`Protocol Buffers`的教程可以查看[Protocol Buffers V3中文指南](https://www.liwenzhou.com/posts/Go/Protobuf3-language-guide-zh/)，本文后续内容默认读者熟悉`Protocol Buffers`。

```protobuf
syntax = "proto3"; 	// 版本声明，使用Protocol Buffers v3版本

package grpcDemopb;		// 包名

// 指定生成的Go代码在你项目中的导入路径
option go_package = "grpcDemo/server/grpcDemopb";

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

### 2、生成指定语言的代码

在 `.proto` 文件中的定义好服务之后，gRPC 提供了生成客户端和服务器端代码的 protocol buffers 编译器插件。

我们使用这些插件可以根据需要生成`Java`、`Go`、`C++`、`Python`等语言的代码。我们通常会在客户端调用这些 API，并在服务器端实现相应的 API。

- 在服务器端，服务器实现服务声明的方法，并运行一个 gRPC 服务器来处理客户端发来的调用请求。gRPC 底层会对传入的请求进行解码，执行被调用的服务方法，并对服务响应进行编码。
- 在客户端，客户端有一个称为存根（stub）的本地对象，它实现了与服务相同的方法。然后，客户端可以在本地对象上调用这些方法，将调用的参数包装在适当的 protocol buffers 消息类型中—— gRPC 在向服务器发送请求并返回服务器的 protocol buffers 响应之后进行处理。

我在`grpcDemo`目录下分别建了`server`和`client`两个项目

在`server`项目根目录下执行以下命令，根据`getcache.proto`生成 go 源码文件。

```bash
protoc --go_out=. --go_opt=paths=source_relative \
--go-grpc_out=. --go-grpc_opt=paths=source_relative \
grpcDemopb/getcache.proto
```

**注意** 如果你的终端不支持`\`符（例如某些同学的Windows），那么你就复制粘贴下面不带`\`的命令执行。

```bash
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative grpcDemopb/getcache.proto
```

生成后的go源码文件会保存在pb文件夹下。

在我的这个目录结构（见下）中，在`client`端， 只需要修改以下代码保存路径：

```go
// 指定生成的Go代码在你项目中的导入路径
option go_package = "grpcDemo/client/grpcDemopb";
```

我`grpcDemo`下的目录结构为（`>tree /f` 获取）：

```go
D:.
├─client
│  │  client.go
│  │  go.mod
│  │  go.sum
│  │
│  └─grpcDemopb
│          getcache.pb.go
│          getcache.proto
│          getcache_grpc.pb.go
│
└─server
    │  go.mod
    │  go.mod
    │  go.sum
    │  server.go
    │
    └─grpcDemopb
            getcache.pb.go
            getcache.proto
            getcache_grpc.pb.go
```

可以看到，分别在server和client项目里生成了Go语言的代码。

### 3、编写业务逻辑代码

gRPC 帮我们解决了 RPC 中的服务调用、数据传输以及消息编解码，我们剩下的工作就是要编写业务逻辑代码。

在服务端编写业务代码实现具体的服务方法，在客户端按需调用这些方法。

### （1）编写Server端代码

`server`端，我只需要实现上面定义的`Get()`服务方法就行，并开启服务。

```go
package main

import (
	"context"
	"encoding/binary"
	"fmt"
	"log"
	pb "main/grpcDemopb"
	"net"

	"google.golang.org/grpc"
)

type Server struct {
	pb.UnimplementedPeanutCacheServer
	addr string // format= ip:port
}

// 缓存命名空间
type Group struct {
	name  string		 // 缓存空间名
	cache map[string]int  // 底层缓存
}

func NewGroup(name string) *Group {
	return &Group{
		name: name,
		cache: map[string]int{
			"Tom":   88,
			"Jack":  98,
			"Alice": 100,
		},
	}
}

func (g *Group) Get(key string) (int, error) {
	if key == "" {
		return -1, fmt.Errorf("key required")
	}
	// fmt.Printf("key:%s\n", key)
	// fmt.Printf("cache:%v\n", g.cache)
	if value, ok := g.cache[key]; ok {
		log.Println("cache hit")
		return value, nil
	} else {
		return -1, fmt.Errorf("key not in cache")
	}
}

// Get 实现PeanutCache service的Get接口
func (s *Server) Get(ctx context.Context, in *pb.GetRequest) (*pb.GetResponse, error) {
	group, key := in.GetGroup(), in.GetKey()
	resp := &pb.GetResponse{}

	log.Printf("[peanutcache_svr %s] Recv RPC Request - (%s)/(%s)", s.addr, group, key)
	if key == "" {
		return resp, fmt.Errorf("key required")
	}
	g := NewGroup(group)
	view, err := g.Get(key)
	if err != nil {
		return resp, err
	}
	byteSlice := make([]byte, 10)
	// 将整数转换为字节切片
	binary.LittleEndian.PutUint32(byteSlice, uint32(view))
	resp.Value = byteSlice
	return resp, nil
}

func main() {
	// 监听本地 8000 端口
	lis, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Printf("failed to listen: %v", err)
		return
	}
	s := grpc.NewServer()                  // 创建一个新的 gRPC 服务器实例
	svr := &Server{addr: "localhost:8080"} // 创建自定义gRPC服务，用来处理gRPC请求
	pb.RegisterPeanutCacheServer(s, svr)   // 在gRPC服务端注册自定义服务
	// 启动服务
	err = s.Serve(lis)
	if err != nil {
		fmt.Printf("failed to serve: %v", err)
		return
	}
}
```

然后，我们在server端启动服务

```go
go run ./server.go
```

### （2）编写Client端代码

在`client`端，我们只需要，与gRPC 服务器建立连接，然后发送请求并接收响应即可，不需要了解`server`端是怎么实现的。

```go
package main

import (
	"context"
	"flag"
	"log"
	pb "main/grpcDemopb"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	addr  = flag.String("addr", "127.0.0.1:8080", "the address to connect to")
	group = "Score"
	key   = "Tom"
)

func main() {
    // flag.Parse() 被调用以解析命令行参数，
    // 用于设置 gRPC 客户端连接的地址或其他相关参数。
	// flag.Parse()	
	// 连接到server端，此处禁用安全传输
	conn, err := grpc.NewClient(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	// 创建一个 gRPC 客户端的服务端点 c，用于调用 gRPC 服务端的函数。
	c := pb.NewPeanutCacheClient(conn)

	// 执行RPC调用并打印收到的响应数据
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, err := c.Get(ctx, &pb.GetRequest{Group: group, Key: key})
	if err != nil {
		log.Fatalf("cache found error: %v", err)
	}
	log.Printf("get cache value: %v", r.Value)
}
```

同样的，接着我们运行`client.go`即可。

**结果：**

`client`

```go
2024/08/30 01:49:36 get cache value: [88 0 0 0 0 0 0 0 0 0]
```

`server`

```go
2024/08/30 01:49:36 [peanutcache_svr localhost:8080] Recv RPC Request - (Score)/(Tom)
2024/08/30 01:49:36 cache hit
```

## gRPC服务类型方法

在gRPC中你可以定义四种类型的服务方法

- **普通 rpc，**客户端向服务器发送一个请求，然后得到一个响应，就像普通的函数调用一样。

  ```protobuf
  // 普通rpc
  rpc Get(GetRequest) returns (GetResponse);
  ```

- **服务器流式 rpc，**其中客户端向服务器发送请求，并获得一个流来读取一系列消息。客户端从返回的流中读取，直到没有更多的消息。gRPC 保证在单个 RPC 调用中的消息是有序的。

  ```protobuf
  // 服务端返回流式数据
  rpc LotsOfReplies(HelloRequest) returns (stream HelloResponse);
  ```

- **客户端流式 rpc，**其中客户端写入一系列消息并将其发送到服务器，同样使用提供的流。一旦客户端完成了消息的写入，它就等待服务器读取消息并返回响应。同样，gRPC 保证在单个 RPC 调用中对消息进行排序。

  ```protobuf
  // 客户端发送流式数据
  rpc LotsOfGreetings(stream HelloRequest) returns (HelloResponse);
  ```

- **双向流式 rpc，**其中双方使用读写流发送一系列消息。这两个流独立运行，因此客户端和服务器可以按照自己喜欢的顺序读写: 例如，服务器可以等待接收所有客户端消息后再写响应，或者可以交替读取消息然后写入消息，或者其他读写组合。每个流中的消息是有序的。

  `````protobuf
  // 双向流式数据
  rpc BidiHello(stream HelloRequest) returns (stream HelloResponse);
  `````

## gRPC安装教程

### 1. 安装gRPC

在你的项目目录下执行以下命令，获取 gRPC 作为项目依赖。

```go
go get google.golang.org/grpc@latest
```

在上述命令执行前，要确定你的项目是否有`go.mod`和`go.sum`这两个文件。没有，记得初始化。

```go
// 第一步
go mod init main
// 第二步
go mod tidy
```

### 2. 安装 Protocol Buffers v3

安装用于生成gRPC服务代码的协议编译器，最简单的方法是从下面的链接：https://github.com/google/protobuf/releases下载适合你平台的预编译好的二进制文件（`protoc-<version>-<platform>.zip`）。记得找v3版本的，然后解压，放到那个目录都行，我习惯放D盘。

- bin 目录下的 protoc 是可执行文件。
- include 目录下的是 google 定义的`.proto`文件，我们`import "google/protobuf/timestamp.proto"`就是从此处导入。
- 我们需要将下载得到的可执行文件`protoc`所在的 bin 目录加到我们电脑的环境变量中。

![image-20240829231252959](https://s2.loli.net/2024/08/29/2rmJuzBs59fGk1C.png)

### 3. 安装插件

如果是使用Go语言做开发，接下来执行下面的命令安装`protoc`的Go插件：

**（1）安装go语言插件：**

```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28
```

该插件会根据`.proto`文件生成一个后缀为`.pb.go`的文件，包含所有`.proto`文件中定义的类型及其序列化方法。

**（2）安装grpc插件：**

```bash
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2
```

该插件会生成一个后缀为`_grpc.pb.go`的文件，其中包含：

- 一种接口类型(或存根`stub`) ，供客户端调用的服务方法。
- 服务器要实现的接口类型。

上述命令会默认将插件安装到`$GOPATH/bin`，为了`protoc`编译器能找到这些插件，请确保你的`$GOPATH/bin`在环境变量中。

**[protocol-buffers 官方Go教程](https://developers.google.com/protocol-buffers/docs/gotutorial)**

### 4. 检查

依次执行以下命令检查一下是否开发环境都准备完毕。

1. 确认 protoc 安装完成。

   ```bash
   ❯ protoc --version
   libprotoc 3.20.1
   ```

2. 确认 protoc-gen-go 安装完成。

   ```bash
   ❯ protoc-gen-go --version
   protoc-gen-go v1.28.0
   ```

   如果这里提示`protoc-gen-go`不是可执行的程序，请确保你的 GOPATH 下的 bin 目录在你电脑的环境变量中。

3. 确认 protoc-gen-go-grpc 安装完成。

   ```bash
   ❯ protoc-gen-go-grpc --version
   protoc-gen-go-grpc 1.2.0
   ```

   如果这里提示`protoc-gen-go-grpc`不是可执行的程序，请确保你的 GOPATH 下的 bin 目录在你电脑的环境变量中。


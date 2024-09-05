## 7、防止缓存击穿

---

当并发了 N 个请求 `key=Tom`，8003 节点向 8001 同时发起了 N 次请求。假设对数据库的访问没有做任何限制的，很可能向数据库也发起 N 次请求，容易导致缓存击穿和穿透。即使对数据库做了防护，针对相同的 key，8003 节点向 8001 发起三次请求也是没有必要的。那这种情况下，我们如何做到只向远端节点发起一次请求呢？

在`groupcache`中，采用了一个`singleFlight`机制来解决这个问题。

## （1）singleflight实现

`/gaocache/singleflight/singleflight.go`

```go
package singlefilght

import (
	"sync"
)

// singlefilght 为gaocache提供缓存击穿的保护
// 当cache并发访问 peer 获取缓存时 如果peer未缓存该值
// 则会向db发送大量的请求获取 造成db的压力骤增
// 因此 将所有由key产生的请求抽象成flight
// 这个flight只会起飞一次(single) 这样就可以缓解击穿的可能性
// flight载有key对应的请求数据 称为packet

// packet表示正在进⾏或者已经结束的请求：
type packet struct {
	wg  sync.WaitGroup // 避免重入
	val interface{}
	err error
}

// ⽤⼀个map来管理不同key对应的请求，因为该操作并发访问需要加锁所以封装成类
type Flight struct {
	mu     sync.Mutex
	flight map[string]*packet // 一个key只对应一个packet
}

// Fly 负责key航班的飞行 fn是获取packet的方法
func (f *Flight) Fly(key string, fn func() (interface{}, error)) (interface{}, error) {
	f.mu.Lock()
	if f.flight == nil {
		f.flight = make(map[string]*packet)
	}
	if p, ok := f.flight[key]; ok {
		f.mu.Unlock()
		p.wg.Wait()         // 获取value正在进行，等待
		return p.val, p.err // 得到value，返回
	}
	p := new(packet)
	p.wg.Add(1)       // 发请求前加锁，
	f.flight[key] = p // 请求正在进行
	f.mu.Unlock()

	p.val, p.err = fn() // 调用fn()执行请求
	p.wg.Done()         // 请求结束

	f.mu.Lock()
	delete(f.flight, key) // 删除完成的请求
	f.mu.Unlock()

	return p.val, p.err
}
```

调⽤Fly⽅法后⾸先检查map中该key对应的请求是否存在，

- 如果存在就通过`waitGroup`的`Wait()`⽅法进⾏阻塞等待， 等待请求完成后直接取⾛最新值. 
- 如果不存在就先将`key`写入`map`，再调⽤请求函数，同时要⽤`Add()`⽅法加锁通知其他协程，执⾏完成后更新map中的值，然后进⾏`Done()`通知其他协程

## （2）singleflight的使用

- 修改 `geecache.go` 中的 `Group`，添加成员变量 `flight`，并更新构建函数 `NewGroup`。
- 修改 `load` 函数，将原来的 load 的逻辑，使用 `g.loader.Fly` 包裹起来即可，这样确保了并发场景下针对相同的 key，`load` 过程只会调用一次。

```go
// Group 提供命名管理缓存/填充缓存的能力
type Group struct {
	name      string               // 缓存空间名
	mainCache *cache               // 主缓存
	server    Picker               // 获取分布式节点能力
	flight    *singlefilght.Flight // 防止缓存击穿
	retriever Retriever            // 回调函数
}

// NewGroup 创建一个新的缓存空间
func NewGroup(name string, capacity int64, retriever Retriever) *Group {
	if retriever == nil {
		panic("Group retriever must be existed!")
	}
	g := &Group{
		name:      name,
		mainCache: newCache(capacity),
		flight:    &singlefilght.Flight{},
		retriever: retriever,
	}
	mu.Lock()
	groups[name] = g
	mu.Unlock()
	return g
}

// 流程2（待实现）本地缓存获取失败，从远程节点获取
func (g *Group) load(key string) (ByteView, error) {
	view , err := g.flight.Fly(key func() (interface{}, error) {
		if g.server != nil {
			// 从远程节点获取
			if fetcher, ok := g.server.PickPeer(key); ok {
				bytes, err := fetcher.Fetch(g.name, key)
				if err == nil {
					log.Printf("[remote peer]:cache hit \n")
					log.Printf("key=%v , value=%v\n", key, string(bytes))
					return ByteView{b: cloneBytes(bytes)}, nil
				}
				log.Printf("fail to get *%s* from peer, %s.\n", key, err.Error())
			}
		}
		// 远程节点获取失败，启用回调函数
		return g.getLocally(key)
	})
	if err == nil{
		return view.(ByteView),err
	}
	// 没有获取到，返回空
	return ByteView{},err
}
```

## （3）测试

只需要更改`client`测试端代码，看看是否只选择了一次。

`Client/main.go`

```go
package main

import (
	"flag"
	"fmt"
	"sync"

	"gaocache"
)

// 模仿数据库
var (
	db = map[string]string{
		"Tom":  "630",
		"Jack": "589",
		"Sam":  "567",
	}
	keys = []string{"Tom", "Tom", "Tom"}

	wg = sync.WaitGroup{}
)

func startClient(apiAddr string, groupName string) {
	defer wg.Done()
	cli := gaocache.NewClient(apiAddr)
	// 第一次查询不在缓存中，回调函数会从数据库中查
	for i := range keys {
		value, err := cli.Fetch(groupName, keys[i])
		if err != nil {
			fmt.Printf("err:%v\n", err)
		}
		fmt.Printf("key:%v,value:%v\n", keys[i], string(value))
	}
	// 第二次查询，此时应该已经加入到了上个节点缓存中
	for i := range keys {
		value, err := cli.Fetch(groupName, keys[i])
		if err != nil {
			fmt.Printf("err:%v\n", err)
		}
		fmt.Printf("key:%v,value:%v\n", keys[i], string(value))
	}
}

func main() {
	// 解析终端命令
	var port int
	// var api bool
	var name string
	flag.StringVar(&name, "name", "scores", "gaocache group name")
	flag.IntVar(&port, "port", 8001, "gaocache server port")
	// flag.BoolVar(&api, "api", false, "Start a api server?")
	flag.Parse()

	// 节点服务端接口 => x.x.x.x:port
	addrMap := map[int]string{
		8001: "localhost:8001",
		8002: "localhost:8002",
		8003: "localhost:8003",
		8004: "localhost:8004",
		9999: "localhost:9999",
	}

	wg.Add(1)
	go startClient(addrMap[port], name)
	wg.Wait()
}

```

结果

![image-20240902025921323](https://s2.loli.net/2024/09/02/L6nuJlheEDwXUPT.png)
package gaocache

import (
	"fmt"
	"gaocache/singleflight"
	"log"
	"sync"
)

var (
	mu     sync.RWMutex              // 管理读写groups并发控制
	groups = make(map[string]*Group) // 多个缓存空间组
)

// Retriever 要求对象实现从数据源获取数据的能力
type Retriever interface {
	retrieve(string) ([]byte, error)
}

// 接口型函数
type RetrieverFunc func(key string) ([]byte, error)

// RetrieverFunc 通过实现retrieve方法，使得任意匿名函数func
// 通过被RetrieverFunc(func)类型强制转换后，实现了 Retriever 接口的能力
func (f RetrieverFunc) retrieve(key string) ([]byte, error) {
	return f(key)
}

// Group 提供命名管理缓存/填充缓存的能力
type Group struct {
	name      string               // 缓存空间名
	mainCache *cache               // 主缓存
	server    Picker               // 获取分布式节点能力
	flight    *singleflight.Flight // 防止缓存击穿
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
		flight:    &singleflight.Flight{},
		retriever: retriever,
	}
	mu.Lock()
	groups[name] = g
	mu.Unlock()
	return g
}

// GetGroup 获取对应命名空间的缓存
func GetGroup(name string) *Group {
	mu.RLock()
	g := groups[name]
	mu.RUnlock()
	return g
}

// RegisterSvr 为 Group 注册 Server
func (g *Group) RegisterSvr(p Picker) {
	if g.server != nil {
		panic("group had been registered server")
	}
	g.server = p
}

// 获取value
func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key required")
	}
	view , err := g.flight.Fly(key,func() (interface{}, error) {
		if value, ok := g.mainCache.get(key); ok {
			log.Println("cache hit")
			return value, nil
		}
		// cache missing, get it another way，本地获取失败
		return g.load(key)
	})
	if err == nil {
		return view.(ByteView), err
	}
	// 没有获取到，返回空
	return ByteView{}, err
}

// 流程2（待实现）本地缓存获取失败，从远程节点获取
func (g *Group) load(key string) (ByteView, error) {
	if g.server != nil {
		// 从远程节点获取，这里相当于实在服务端做了节点选择，而不是客户端
		// 但是严格来讲时grpc的客户端进行了选择，因为是grpc client调用了Get方法
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
}

// getLocally 本地向Retriever取回数据并填充缓存
func (g *Group) getLocally(key string) (ByteView, error) {
	bytes, err := g.retriever.retrieve(key)
	if err != nil {
		return ByteView{}, err
	}
	value := ByteView{b: cloneBytes(bytes)}
	// 将从DB等处获取的key-value填充至缓存
	g.populateCache(key, value)
	return value, nil
}

// 本地回调函数获取key-value后，也要将key-value填充至缓存
func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value)
}

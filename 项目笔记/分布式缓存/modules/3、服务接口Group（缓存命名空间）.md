## 3、服务接口Group（缓存命名空间）

---

**Group** 是 gaocache 最核心的数据结构**，负责与用户的交互，提供命名管理缓存/填充缓存的能力,并且控制缓存值存储和获取的流程。**

缓存的获取和存储流程如下：

<img src="https://s2.loli.net/2024/08/31/OaqViYEcok6KpZQ.jpg" alt="key获取流程图jpg" style="zoom:50%;" />

接下来，我们先将实现流程中的（1）和（3），**以提供本地单机并发缓存的能力**，远程交互的将在后面实现。

## （1）回调函数-RetrieverFunc

```go
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
```

- 定义接口 `Retriever` 和 回调函数 `retrieve(key string)([]byte, error)`，参数是 key，返回值是 []byte。
- 定义函数类型 `RetrieverFunc`，并实现 Getter 接口的 `Get` 方法。
- 函数类型实现某一个接口，称之为接口型函数，方便使用者在调用时既能够传入函数作为参数，也能够传入实现了该接口的结构体作为参数。

>定义一个函数类型 F，并且实现接口 A 的方法，然后在这个方法中调用自己。这是 Go 语言中将其他函数（参数返回值定义与 F 一致）转换为接口 A 的常用技巧。

## （2）Group 定义



```go
var (
	mu     sync.RWMutex // 管理读写groups并发控制
	groups = make(map[string]*Group)  // 不同缓存空间组
)

// Group 提供命名管理缓存/填充缓存的能力
type Group struct {
	name      string    // 缓存空间名
	mainCache *cache    // 主缓存
	retriever Retriever // 回调函数
}
```

创建一个新的缓存空间函数：

```go
// NewGroup 创建一个新的缓存空间
func NewGroup(name string, capacity int64, retriever Retriever) *Group {
	if retriever == nil {
		panic("Group retriever must be existed!")
	}
	g := &Group{
		name:      name,
		mainCache: newCache(capacity),
		retriever: retriever,
	}
	mu.Lock()
	groups[name] = g
	mu.Unlock()
	return g
}
```

获取对应的命名空间Group的存在

```go
// GetGroup 获取对应命名空间的缓存
func GetGroup(name string) *Group {
	mu.RLock()
	g := groups[name]
	mu.RUnlock()
	return g
}
```

- 一个 Group 可以认为是一个缓存的命名空间，每个 Group 拥有一个唯一的名称 `name`。比如可以创建三个 Group，缓存学生的成绩命名为 scores，缓存学生信息的命名为 info，缓存学生课程的命名为 courses。
- 第二个属性是 ` retriever Retriever`，即缓存未命中时获取源数据的回调(callback)。
- 第三个属性是 `mainCache cache`，即一开始实现的并发缓存，也是我们的主要缓存。
- 构建函数 `NewGroup` 用来实例化 Group，并且将 group 存储在全局变量 `groups` 中。
- `GetGroup` 用来特定名称的 Group，这里使用了只读锁 `RLock()`，因为不涉及任何冲突变量的写操作。

## （3）Group 的 Get() 方法

```go
// 获取value
func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key required")
	}
	if value, ok := g.cache.get(key); ok {
		log.Println("cache hit")
		return value, nil
	}
	// cache missing, get it another way
	return g.load(key)
}

//  流程2（待实现）本地缓存获取失败，从远程节点获取
func (g *Group) load(key string) (ByteView, error) {
    // 启用回调函数从其他处获取
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
```

- Get 方法实现了上述所说的流程 ⑴ 和 ⑶。
- 流程 ⑴ ：从 mainCache 中查找缓存，如果存在则返回缓存值。
- 流程 ⑶ ：缓存不存在，则调用 load 方法，load 调用 getLocally（分布式场景下会调用 getFromPeer 从其他节点获取），getLocally 调用用户回调函数 `g.retriever.retrieve(key)` 获取源数据，并且将源数据添加到缓存 `mainCache` 中（通过 `populateCache` 方法）

## 测试：

```go
package gaocache

import (
	"fmt"
	"log"
	"testing"
)

func TestGet(t *testing.T) {
	mysql := map[string]string{
		"Tom":  "630",
		"Jack": "589",
		"Sam":  "567",
	}
	loadCounts := make(map[string]int, len(mysql))

	g := NewGroup("scores", 2<<10, RetrieverFunc(
		func(key string) ([]byte, error) {
			log.Println("[Mysql] search key", key)
			if v, ok := mysql[key]; ok {
				if _, ok := loadCounts[key]; !ok {
					loadCounts[key] = 0
				}
				loadCounts[key]++
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))

	for k, v := range mysql {
		if view, err := g.Get(k); err != nil || view.String() != v {
			t.Fatalf("failed to get value of %s", k)
		}
		if _, err := g.Get(k); err != nil || loadCounts[k] > 1 {
			t.Fatalf("cache %s miss", k)
		}
	}

	if view, err := g.Get("unknown"); err == nil {
		t.Fatalf("the value of unknow should be empty, but %s got", view)
	} else {
		log.Println(err)
	}
}
```

- 在这个测试用例中，我们主要测试了 2 种情况
- 1）在缓存为空的情况下，能够通过回调函数获取到源数据。
- 2）在缓存已经存在的情况下，是否直接从缓存中获取，为了实现这一点，使用 `loadCounts` 统计某个键调用回调函数的次数，如果次数大于1，则表示调用了多次回调函数，没有缓存。

测试结果如下：

```go
=== RUN   TestGet
2024/09/01 00:25:44 [Mysql] search key Sam
2024/09/01 00:25:44 cache hit
2024/09/01 00:25:44 [Mysql] search key Tom
2024/09/01 00:25:44 cache hit
2024/09/01 00:25:44 [Mysql] search key Jack
2024/09/01 00:25:44 cache hit
2024/09/01 00:25:44 [Mysql] search key unknown
2024/09/01 00:25:44 unknown not exist
--- PASS: TestGet (0.01s)
PASS
ok      gaocache        0.486s
```


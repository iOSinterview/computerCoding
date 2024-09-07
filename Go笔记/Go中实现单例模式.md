## Go中实现单例模式

---

## 简介

**单例模式**（Singleton Pattern）是一种创建型设计模式，旨在确保一个类在整个应用程序生命周期中只能创建一个实例，并且为该实例提供全局访问点。这个模式的主要特点是：

1. **唯一性**：单例模式保证某个类只有一个实例。
2. **全局访问**：提供一个全局访问点，任何地方都可以获取这个实例。
3. **延迟实例化**：实例通常是在第一次访问时创建的，即所谓的 "懒汉式"（lazy initialization）。当然，也可以在程序启动时直接创建。

### 单例模式的应用场景

单例模式在以下场景非常有用：

- **日志管理器**：确保日志系统在应用程序中只有一个实例，用于统一记录日志。
- **数据库连接池**：通常应用程序只需要一个数据库连接池实例来管理连接。
- **线程池**：在多线程环境下，线程池的实例需要唯一。
- **配置管理器**：确保全局配置在系统中只有一个对象负责管理。

### 单例模式的实现要素

1. **构造函数私有化**：防止外部创建实例。
2. **静态/全局访问方法**：提供一个全局的访问点来获取该类的唯一实例。
3. **线程安全**：在多线程环境下，确保单例的创建不会出现竞争条件。

### 单例模式的实现方式

- **饿汉式（Eager Initialization）**：在程序启动时直接创建实例，无论是否使用。
- **懒汉式（Lazy Initialization）**：实例在第一次调用时创建，通常使用 `sync.Once` 或双重检查锁来确保线程安全。

### 单例模式的优缺点

#### 优点：

- **控制实例数量**：确保类只有一个实例，节省内存和资源。
- **全局访问**：提供全局唯一的访问点，方便共享状态。

#### 缺点：

- **难以扩展**：单例模式会导致类的依赖性增加，难以进行拓展

## 代码实现

```go
package main

import (
	"fmt"
	"sync"
)

// 单例结构体
type Singleton struct {
	name string
}

// 定义一个私有的单例实例
var singleton *Singleton

// 用sync.Once 确保单例只被初始化一次
var once sync.Once

func GetSingleton() *Singleton {
	once.Do(func() {
		singleton = &Singleton{name: "singleName"}
	})
	return singleton
}

func main() {
	// 获取单例实例
	s1 := GetSingleton()
	fmt.Printf("first time:%s\n", s1.name)
	s2 := GetSingleton()
	fmt.Printf("second time:%s\n", s2.name)

	// 看看是否为同一个实例
	fmt.Printf("s1 = s2？：%v\n", s1 == s2)
}
```


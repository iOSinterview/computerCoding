package singleflight

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

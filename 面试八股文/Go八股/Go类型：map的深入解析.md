# Go类型：map的深入解析

---

# Map的底层实现？

map 是由 `key-value` 对组成的；`key` 只会出现一次。

Go中的map基于`Hash表`实现，并且使用的是`链表法`解决冲突。

主要由两个结构体，一个hmap（A header for a Go map），另一个是bmap（a bucket for a Go map，通常叫其 bucket，桶）。

每个 bucket 设计成最多只能放 8 个 key-value 对，如果有第 9 个 key-value 落入当前的 bucket，那就需要再构建一个 bucket ，通过 `overflow` 指针连接起来。

![img](assets/v2-e0b9187cfc0dd066239b44df1f4594ba_r.jpg)

# Map的扩容过程？

​	首先是map的扩容时机，在向map插入新的key的时候，会进行以下两个条件检测

- 转载因子（6.5）超过阈值。$loadFachtor=mapCount / 2^B$，B决定了bucket的数量（$2^B$)
- overflow的bucket （溢出桶）数量过多。

在 Go 的 map 实现中，当 map 需要扩容时，为了避免一次性复制整个哈希表的数据导致的大量计算，**Go 采用了分批（incrementally）渐进式的复制方式。原有的 key 并不会一 次性搬迁完毕，每次最多只会搬迁 2 个 bucket。**

**如果Map正在扩容，那么在删除、插入或更新元素时都会执行一次迁移操作，仅针对要变更的bucket进行扩容。这样可以确保扩容过程的平滑进行，而不会因为其他操作的干扰而中断。**

# Map的查找？

- 当key经过hash函数计算后得到hash值，共64bit，先用到最后B个bit位（这个B是hmap结构体里面的，B决定了bucket的数量（$2^B$)），找到对应的bucket；
- 再利用hash值的高8位，找到key在bucket中的位置；
- 当前bmap的bucket未找到，则查询对应的overflow bucket，有数据再对比完整的hash值；
- 如果当前map处于数据迁移状态，则优先从oldbuckets中找。

# map是并发安全的吗？为什么？怎么解决？

​		**`map` 不是并发安全的。当多个 goroutines 并发地对一个 `map` 进行读写操作时，会产生数据竞争（data race），导致程序出现意料之外的行为。**

**主要原因**：

- **非原子性操作：** `map` 的操作不是原子的，即使是读取操作也可能由于数据的并发修改而导致不一致性。
- **无锁机制：** `map` 在实现上并没有提供内置的锁机制来保护并发访问，这意味着多个 goroutines 同时对一个 `map` 进行读写可能导致数据错乱。

**解决方案：**

- **使用互斥锁（Mutex）：** 可以通过在对 `map` 进行读写操作前使用互斥锁来保护 `map`，确保同一时间只有一个 goroutine 可以对 `map` 进行操作。

- **使用读写锁（RWMutex）：** 如果对 `map` 的读操作比写操作频繁，可以考虑使用读写锁来提高并发性能，允许多个 goroutines 同时读取 `map`，但只有一个 goroutine 可以进行写操作。

- **使用并发安全的数据结构：** 也可以考虑使用 Go 标准库中提供的并发安全的数据结构，如 `sync.Map`，它内部实现了并发安全的访问机制。

  ```go
  sync.Map主要方法
  Store 方法：存储键值对，
  Load 方法：加载值，
  Delete 方法：删除键，
  Range 方法：遍历整个 sync.Map。
  ```

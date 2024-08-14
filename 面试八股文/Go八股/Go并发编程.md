

[TOC]



# Go并发编程

# 1、mutex几种状态？

-  mutexLocked — 表示互斥锁的锁定状态； 
- mutexWoken — 表示从正常模式被唤醒； 
- mutexStarving — 当前的互斥锁进入饥饿状态； 
-  waitersCount — 当前互斥锁上等待的 Goroutine 个数；

# 2、Mutex 正常模式和饥饿模式

## 正常模式（非公平锁）

 正常模式下，所有等待锁的 goroutine 按照 FIFO（先进先出）顺序等待。唤醒 的 goroutine 不会直接拥有锁，而是会和新请求 goroutine 竞争锁。新请求的 goroutine 更容易抢占：因为它正在 CPU 上执行，所以刚刚唤醒的 goroutine 有很大可能在锁竞争中失败。在这种情况下，这个被唤醒的 goroutine 会加入 到等待队列的前面。

##  饥饿模式（公平锁）

 为了解决了等待 goroutine 队列的长尾问题。

饥饿模式下，mutex 的所有权直接从解锁的 goroutine 递交到等待队列中排在最前方的 goroutine。新进来的 goroutine 不会参与抢锁也不会进入自旋状态，会直接进入等待队列的尾部。这样很好的解决了老的 goroutine 一直抢不到锁的场景。 

饥饿模式的触发条件：当一个 goroutine 等待锁时间超过 1 毫秒时，或者当前队列只剩下一个 goroutine 的时候，mutex 切换到饥饿模式。 

**总结**： 对于两种模式，正常模式下的性能是最好的，goroutine 可以连续多次获取锁，饥饿模式解决了取锁公平的问题，避免了病态情况下的尾部延迟，但是性能会下降，这其实是性能和公平的一个平衡模式。

# 3、Mutex 允许自旋的条件

如果 Goroutine 占用锁资源的时间比较短，那么每次都调用信号量来阻塞唤起 goroutine，将会很**浪费**资源。

因此在符合一定条件后，mutex 会让当前的 Goroutine 去**空转** CPU，在空转完后再次调用 CAS 方法去尝试性的占有锁资源，直到不满足自旋条件，则最终会加入到等待队列里。

-   锁已被占用，并且锁不处于饥饿模式。 
- 积累的自旋次数小于最大自旋次数（active_spin=4）。 
-  CPU 核数大于 1。 
- 有空闲的 P。 
-  当前 Goroutine 所挂载的 P 下，本地待运行队列为空。

# 4、RWMutex实现

```go
type RWMutex struct {
  w           Mutex   // 互斥锁解决多个writer的竞争
  writerSem   uint32  // writer信号量
  readerSem   uint32  // reader信号量
  readerCount int32   // reader的数量
  readerWait  int32   // writer等待完成的reader的数量
}

const rwmutexMaxReaders = 1 << 30
```

通过记录 readerCount 读锁的数量来进行控制，这个readerCount有两个含义：

- 没有writer竞争或者持有锁的时候，readerCount就是当前reader的计数。
- 有writer竞争或者持有锁的时候，此时readerCount为负数，用来标识。

过程：

- 当一个reader释放（RUnlock）的时候，将readerCount减1。
- 当有一个writer获得锁之后，会将readerCount字段设为负数（相反数），此时readerCount既保存了reader的数量，又表示当前有writer持有锁。
- 当有writer进行抢占锁时，如果readerCount不为0，writer进入阻塞状态，直到所有活跃的reader都释放锁完毕，才会唤醒这个writer。
- 当一个writer释放锁（Unlock）的时候，会再次反转readerCount字段，唤醒之后新来阻塞的reader。

# 5、RWMutex 注意事项

- RWMutex 是单写多读锁，该锁可以加多个读锁或者一个写锁 
- **读锁占用的情况下会阻止写，不会阻止读，多个 Goroutine 可以同时获取读锁** 
- **写锁会阻止其他 Goroutine（无论读和写）进来，整个锁由该 Goroutine  独占** 
- **适用于读多写少的场景** 
- RWMutex 类型变量的零值是一个未锁定状态的互斥锁
- RWMutex 在首次被使用之后就不能再被拷贝 
- RWMutex 的读锁或写锁在未锁定状态，解锁操作都会引发 panic 
-  RWMutex 的一个写锁去锁定临界区的共享资源，如果临界区的共享资源已被（读锁或写锁）锁定，这个写锁操作的 goroutine 将被阻塞直到解锁 
- RWMutex 的读锁不要用于递归调用，比较容易产生死锁。
-  RWMutex 的锁定状态与特定的 goroutine 没有关联。一个 goroutine 可以 RLock（Lock），另一个 goroutine 可以 RUnlock（Unlock） 
- 写锁被解锁后，所有因操作锁定读锁而被阻塞的 goroutine 会被唤醒，并都可以成功锁定读锁 
- 读锁被解锁后，在没有被其他读锁锁定的前提下，所有因操作锁定写锁而被阻塞的 Goroutine，其中等待时间最长的一个 Goroutine 会被唤醒

# 6、WaitGroup 

## （1）用法

Go中的WaitGroup是一种常见的并发控制方式，一个 WaitGroup 对象可以等待一组goroutine结束。

一共有三个方法：

- `Add` 方法用于设置 WaitGroup 的计数值，可以理解为goroutine的数量
- `Done` 方法用于将 WaitGroup 的计数值减1，一个goroutine执行完后调用。
- `Wait` 方法用于阻塞调用者，直到 WaitGroup 的计数值为0，即所有goroutine都完成

## （2） 实现原理

```go
type WaitGroup struct {
    noCopy noCopy // noCopy 字段标识，由于 WaitGroup 不能复制，方便工具检测

    state1 [3]uint32  // 12个字节，8个字节标识 计数值和等待数量，4个字节用于标识信号量
}
```

-  WaitGroup 主要维护了一个复合字段state1，共12字节，分成了两个部分。第一部分前面8个字节标识 2 个计数器，一个是请求计数器 ，一个是等待计数器 ，二者组成一个 8字节 的值，请求计数器占高 4字节，等待计数器占低 4字节。第二部分 剩余 `32位`（4个字节）`semap` 用于标识信号量。
- 每次 Add 执行，请求计数器 加 1，Done 方法执行，等待计数器 1，请求计数器 为 0 时通过信号量唤醒 Wait()。

## （3）注意事项

- 保证 Add 在 Wait 前调用
- Add 中不传递负数
- 任务完成后不要忘记调用 Done 方法，建议使用 defer wg.Done()
- 不要复制使用 WaitGroup，函数传递时使用指针传递
- 尽量不复用 WaigGroup，减少出问题的风险

# 8、什么是 sync.Once？

特性：`sync.Once` 可以确保在并发环境下某个操作只被执行一次，无论有多少 Goroutine 尝试执行它，是一个并发安全的操作。

使用：`sync.Once` 主要通过 `Do` 方法来执行操作，该方法接收一个无参数无返回值函数作为参数，确保该函数只被执行一次。

```go
var once sync.Once
// 无参数无返回值的函数
func initialize() {
    // 初始化操作
}

func main() {
    once.Do(initialize) // 确保 initialize 函数只被执行一次
}
```

使用场景：

- **单例模式**：`sync.Once` 可以用于实现单例模式，确保某个对象只被初始化一次。
- **延迟初始化**：当需要在程序运行时执行某个初始化操作，并且只需要执行一次时，`sync.Once` 是一个很好的选择。
- **初始化资源**：在需要确保某个资源只被初始化一次的情况下，例如全局配置信息的加载等。

# 9、什么操作叫做原子操作？

原子操作是指在多线程或并发编程中**不可被中断的操作，要么完全执行，要么完全不执行，不会被其他线程的操作所干扰。**原子操作通常是 CPU 提供的特性，可以保证在多线程并发执行的情况下，**对共享数据的操作是线程安全的。**

主要有以下特性：

- **不可分割性**：即要么完全执行，要么完全不执行，不可被中断。
- **并发安全**：原子操作能够保证在多线程并发的情况下，对共享数据的操作是线程安全的，不需要额外的同步机制（如锁）。
- **原子性**：原子操作是原子性的，不会出现竞争条件（race condition）。

Go中的 `sync/atomic` 包提供了原子操作。

# 10、原子操作和锁的区别？

- 原子操作由底层硬件支持，而锁是基于原子操作+信号量完成的，由操作系统的调度器实现。若实现相同的功能，原子操作通常会更有效率，能利用计算机多核的优势。而锁的使用会引入额外的开销，可能会降低程序的性能。
- **原子操作**适用于简单的操作，如对计数器的增减、标记位的设置等。**锁**适用于需要保证一系列操作的原子性，或者需要对共享资源进行复杂的操作和控制的场景。
- 原子操作不需要额外的同步机制（如锁）来保护共享资源，可以提高并发性能。
- 原子操作是单个指令的互斥操作﹔互斥锁/读写锁是一种数据结构，可以完成临界区（多个指令)的互斥操作，扩大原子操作的范围，锁保护的是一段逻辑。
- 原子操作是无锁操作，属于乐观锁（假设操作值未曾被改编）；说起锁的时候，一般属于悲观锁（假设会有并发的操作想要修改被操作的值）。

# 11、sync.Pool 有什么用

频繁地分配、回收内存会给 GC 带来一定的负担，`sync.Pool` 用于缓存临时对象，以便在需要时重用这些临时对象，从而减少内存分配GC的压力。`sync.Pool` 在高性能的并发编程中非常有用，特别是在需要频繁创建和销毁临时对象的场景下，比如说`对象池`的使用。

主要有两个方法：

```go
Get()    // 从池中获取对象    obj := pool.Get().(string)
Put()    // 将对象放回池中    pool.Put(obj)
```

# 12、什么是 CAS ?

CAS（Compare and Swap）是一种原子操作，用于在无锁情况下保证数据一致性。CAS操作包含三个操作数：内存位置、预期原值和新值。在执行CAS操作时，会将内存位置的值与预期原值进行比较。如果两者相等，则处理器会自动将该位置的值更新为新值；如果不相等，则处理器不做任何操作。这个过程是原子的，即在整个操作期间，不会被其他线程或进程中断。

CAS操作的原理基于乐观锁机制，通过硬件指令实现原子操作。在多线程并发编程中，CAS操作可以避免传统的锁机制引起的线程阻塞和上下文切换等问题，提高程序的并发性能。

在实际应用中，CAS机制可以用于实现一些高性能的并发算法和数据结构，如非阻塞队列、并发计数器、自旋锁等。通过CAS机制，可以避免传统锁机制中的线程阻塞和唤醒操作，提高了并发性能和吞吐量。另外，CAS机制还可以用于实现一些乐观锁的算法，如乐观锁的[数据库](https://cloud.baidu.com/solution/database.html)操作和分布式锁的实现。

然而，CAS机制也存在一些问题和限制。首先，CAS机制需要硬件的支持，不是所有的平台和处理器都能够完全支持CAS指令，因此在一些旧的或特定的硬件平台上可能无法使用CAS机制。其次，CAS机制在高并发情况下可能会出现ABA问题，即在执行CAS操作时，共享变量的值可能已经被其他线程修改过了，导致CAS操作成功但实际上并没有达到预期的效果。针对这个问题，可以使用版本号或标记位来解决。

Go中示例：

```go
package main

import (
    "fmt"
    "sync/atomic"
)

func main() {
    var value int32 = 0

    // 进行 CAS 操作
    expected := int32(0)
    new := int32(1)

    // 如果 value 的值等于 expected，则将 value 替换为 new
    success := atomic.CompareAndSwapInt32(&value, expected, new)

    if success {
        fmt.Println("CAS succeeded. New value:", value)
    } else {
        fmt.Println("CAS failed. Current value:", value)
    }
}
```


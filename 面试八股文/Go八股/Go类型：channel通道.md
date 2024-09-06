## Go类型：channel通道

---

​       Go 语言中，不要**通过共享内存来通信，而要通过通信来实现内存共享**。Go 的 CSP(Communicating Sequential Process)并发模型，中文可以叫做**通信顺序进程**，是通过 goroutine 和 channel 来实现的。 channel 收发遵循先进先出 FIFO 的原则，它是并发安全的。

## **channel可以分为两类：**

- 无缓冲channel：可以看作**同步模式**，收发放两者都ready的情况下，数据才能传输，否则将会阻塞。
- 有缓冲channel：可以分为**异步模式**。

也可以分为单向通道可双向通道，其中 单向通道使其只能发送或只能接收数据。这种单向通道可以增加程序的安全性。

## **channel的基本用法**：

- 读取 <- chan
- 写入 chan <-
- 关闭 close(chan)：接收方可以通过第二个返回值来判断通道是否被关闭。
- 获取channel长度 len(chan)
- 获取channel容量 cap(chan)
- **select非阻塞访问方式，从所有的case中挑选一个不会阻塞的channel进行读写操作，或是default执行。**

但是在使用时有一些**要注意的点：**

- **如果向一个`nil`（未初始化）的channel发送或者接收数据会造成永久阻塞。**
- **给一个已经关闭的channel发送数据，会引起panic。**
- **从一个已经关闭的channel接收数据，如果缓冲区为空，会返回一个 0 值。**

## **底层结构**：

## channel的底层结构是hchan，维护底层的一个循环队列（ring buffer）

```go
buf      unsafe.Pointer       // 指向循环队列的指针
sendx    uint                 // 已发送元素在循环队列中的位置
recvx    uint                 // 已接收元素在循环队列中的位置
recvq    waitq                // 等待接收的goroutine的等待队列
sendq    waitq                // 等待发送的goroutine的等待队列
lock mutex                    // 控制chan并发访问的互斥锁
```

**当 channel 因为缓冲区不足而阻塞了队列，则使用双向链表存储。**
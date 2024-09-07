## Go并发模型——生产消费者模型

---

思路：利用一个`channel`，生产者发送数据，消费者接收数据。

## 单生产者-单消费者

```go
package main

import (
	"fmt"
	"time"
)

func Producer(ch chan<- int, cnt int) {
	for i := 0; i < cnt; i++ {
		fmt.Printf("produce data : %d\n", i)
		ch <- i
		time.Sleep(50 * time.Millisecond)
	}
	defer close(ch)
}

func Consumer(ch <-chan int) {
	for data := range ch {
		fmt.Printf("consume data : %d\n", data)
		time.Sleep(10 * time.Millisecond)
	}
}

func main() {
	ch := make(chan int)
	go Producer(ch, 10)
	go Consumer(ch)

	time.Sleep(time.Second * 5)
	fmt.Println("main is over")
}
```

消费者的速度小于生产者，可能产生阻塞。当生产的数据速度大于`channel`容量，容易阻塞。

通过channel实现了经典的生产者和消费者模型，利用了channel的特性。但要注意，当消费者的速度小于生产者时，channel就有可能产生拥塞，导致占用内存增加，所以，**在实际场景中需要考虑channel的缓冲区的大小**。设置了channel的大小，当生产的数据大于channel的容量时，生产者将会阻塞，这些问题都是要在实际场景中需要考虑的。
**一个解决办法就是使用一个固定的数组或切片作为环形缓冲区，而非channel，通过Sync包的机制来进行同步，实现生产者消费者模型**，这样可以避免由于channel满而导致消费者端阻塞。**但，对于环形缓冲区而言，可能会覆盖老的数据**，同样需要考虑具体的使用场景。关于环形缓冲区的原理和实现，在分析Sync包的使用时再进一步分析。


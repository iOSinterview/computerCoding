

# Go中使用heap实现最小/最大堆

堆的性质如下。堆的逻辑结构为完全二叉树、底层数据结构通常为数组。对于最大堆，该二叉树父节点值皆比子节点值大，最小堆则反之。这种大小关系可以被称为堆序性。因此，对于最大堆，根为堆中最大的数。因此，最大/小堆也被称为大/小根堆。

对于堆，主要有两种操作，`Push()`插入新的元素，`Pop()`弹出堆顶（即二叉树根）元素，如果是最大堆，则弹出最大的元素，最小堆则相反。堆在进行插入和弹出操作后，都会自动调整元素位置，保证堆序性质。Go中并没有内建可以直接调用的堆容器，需要实现一些接口才可使用。

参考`"container/heap"`内的定义，你需要实现`sort.Interface`内的方法（即`Less()`、`Len()`、`Swap()`），和`Push()`、`Pop()`方法。才可以创建一个堆。

- ## 最小堆

```go
package main

import "container/heap"

// 定义一个整数切片，实现heap.Interface 接口方法
type IntHeap []int // 这里int是堆的数据类型，也可以是结构体

func (h IntHeap) Len() int           { return len(h) }
func (h IntHeap) Less(i, j int) bool { return h[i] < h[j] }
// 最大堆 的区别就是Less()，改为 > 即可
// func (h IntHeap) Less(i, j int) bool { return h[i] > h[j] }
func (h IntHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *IntHeap) Push(x interface{}) {
	*h = append(*h, x.(int))
}

func (h *IntHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[:n-1]
	return x
}

func main() {
	// 初始化堆
	h := &IntHeap{}
	heap.Init(h)
	x := 0
	// 入堆
	heap.Push(x)
	// 出堆
	heap.Pop(x)
}
```




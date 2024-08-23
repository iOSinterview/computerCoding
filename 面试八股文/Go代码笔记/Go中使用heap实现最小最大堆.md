

# Go中使用heap实现最小/最大堆

堆的性质如下。堆的逻辑结构为完全二叉树、底层数据结构通常为数组。对于最大堆，该二叉树父节点值皆比子节点值大，最小堆则反之。这种大小关系可以被称为堆序性。因此，对于最大堆，根为堆中最大的数。因此，最大/小堆也被称为大/小根堆。

对于堆，主要有两种操作，`Push()`插入新的元素，`Pop()`弹出堆顶（即二叉树根）元素，如果是最大堆，则弹出最大的元素，最小堆则相反。堆在进行插入和弹出操作后，都会自动调整元素位置，保证堆序性质。Go中并没有内建可以直接调用的堆容器，需要实现一些接口才可使用。

参考`"container/heap"`内的定义，你需要实现`sort.Interface`内的方法（即`Less()`、`Len()`、`Swap()`），和`Push()`、`Pop()`方法。才可以创建一个堆。

```go
type Interface interface {
    sort.Interface
    Push(x interface{}) // 向末尾添加元素
    Pop() interface{}   // 从末尾删除元素
}

type Interface interface {
    // Len方法返回集合中的元素个数
    Len() int
    // Less方法报告索引i的元素是否比索引j的元素小
    Less(i, j int) bool
    // Swap方法交换索引i和j的两个元素
    Swap(i, j int)
}
```

# 最小堆/最大堆

当我们每次取剩余数组中最大/小的一个或者几个值的时候，可以想到采用最大/小堆。

```go
// This example demonstrates an integer heap built using the heap interface.
package heap_test
import (
    "container/heap"
    "fmt"
)
// An IntHeap is a min-heap of ints.
type IntHeap []int
func (h IntHeap) Len() int           { return len(h) }
func (h IntHeap) Less(i, j int) bool { return h[i] < h[j] }
// MaxHeap
// func (h IntHeap) Less(i, j int) bool { return h[i] > h[j] }
func (h IntHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h *IntHeap) Push(x interface{}) {
    // Push and Pop use pointer receivers because they modify the slice's length,
    // not just its contents.
    *h = append(*h, x.(int))
}
func (h *IntHeap) Pop() interface{} {
    old := *h
    n := len(old)
    x := old[n-1]
    *h = old[0 : n-1]
    return x
}
// This example inserts several ints into an IntHeap, checks the minimum,
// and removes them in order of priority.
func Example_intHeap() {
    h := &IntHeap{2, 1, 5}
    heap.Init(h)
    heap.Push(h, 3)
    fmt.Printf("minimum: %d\n", (*h)[0])
    for h.Len() > 0 {
        fmt.Printf("%d ", heap.Pop(h))
    }
    // Output:
    // minimum: 1
    // 1 2 3 5
}
```

# 优先队列

```go
// This example demonstrates a priority queue built using the heap interface.
package heap_test
import (
    "container/heap"
    "fmt"
)
// An Item is something we manage in a priority queue.
type Item struct {
    value    string // The value of the item; arbitrary.
    priority int    // The priority of the item in the queue.
    // The index is needed by update and is maintained by the heap.Interface methods.
    index int // The index of the item in the heap.
}
// A PriorityQueue implements heap.Interface and holds Items.
type PriorityQueue []*Item
func (pq PriorityQueue) Len() int { return len(pq) }
func (pq PriorityQueue) Less(i, j int) bool {
    // We want Pop to give us the highest, not lowest, priority so we use greater than here.
    return pq[i].priority > pq[j].priority
}
func (pq PriorityQueue) Swap(i, j int) {
    pq[i], pq[j] = pq[j], pq[i]
    pq[i].index = i
    pq[j].index = j
}
func (pq *PriorityQueue) Push(x interface{}) {
    n := len(*pq)
    item := x.(*Item)
    item.index = n
    *pq = append(*pq, item)
}
func (pq *PriorityQueue) Pop() interface{} {
    old := *pq
    n := len(old)
    item := old[n-1]
    item.index = -1 // for safety
    *pq = old[0 : n-1]
    return item
}
// update modifies the priority and value of an Item in the queue.
func (pq *PriorityQueue) update(item *Item, value string, priority int) {
    item.value = value
    item.priority = priority
    heap.Fix(pq, item.index)
}
// This example creates a PriorityQueue with some items, adds and manipulates an item,
// and then removes the items in priority order.
func Example_priorityQueue() {
    // Some items and their priorities.
    items := map[string]int{
        "banana": 3, "apple": 2, "pear": 4,
    }
    // Create a priority queue, put the items in it, and
    // establish the priority queue (heap) invariants.
    pq := make(PriorityQueue, len(items))
    i := 0
    for value, priority := range items {
        pq[i] = &Item{
            value:    value,
            priority: priority,
            index:    i,
        }
        i++
    }
    heap.Init(&pq)
    // Insert a new item and then modify its priority.
    item := &Item{
        value:    "orange",
        priority: 1,
    }
    heap.Push(&pq, item)
    pq.update(item, item.value, 5)
    // Take the items out; they arrive in decreasing priority order.
    for pq.Len() > 0 {
        item := heap.Pop(&pq).(*Item)
        fmt.Printf("%.2d:%s ", item.priority, item.value)
    }
    // Output:
    // 05:orange 04:pear 03:banana 02:apple
}
```


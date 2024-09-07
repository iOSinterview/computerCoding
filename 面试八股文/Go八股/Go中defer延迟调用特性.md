## Go中defer延迟调用特性

---

由于`defer`语句延迟调用的特性，所以`defer`语句能非常方便的处理资源释放问题。比如：资源清理、文件关闭、解锁及记录时间等。

## （1）多个defer的执行顺序

多个defer出现的时候，**它是一个“栈”的关系，也就是先进后出**。一个函数中，写在前面的defer会比写在后面的defer调用的晚。

## （2）defer与return执行顺序

**在Go语言的函数中`return`语句在底层并不是原子操作，它分为给返回值赋值和RET指令两步。而`defer`语句执行为：return 赋值—>defer—>RET指令，**所以使用defer可以达到修改返回值的目的。

**关键点在于`具名返回值`（也就是返回的参数列表中的变量）是否参与`defer` 运算，参与运算则返回值会受到`defer`的处理影响，不参与则与之无关。**

### 经典案例1：

```go
func f1() int {
	x := 5
	defer func() {
		x++
	}()
    return x  // 1、返回值=x=5；2、defer：x=6；3、RET：返回值=5
}

func f2() (x int) {
	defer func() {
		x++
	}()
	return 5 // 1、返回值x=5；2、defer：x=6；3、RET：返回值x=6
}

func f3() (y int) {
	x := 5
	defer func() {
		x++
	}()
	return x // 1、返回值y=x=5；2、defer：x=6；3、RET：返回值y=5
}
func f4() (x int) {
	defer func(x int) {
		x++
	}(x)
	return 5 // 1、返回值x=5；2、defer：x当做形参传入，不会改变返回值，x=6；3、RET：返回值x=5
}
func main() {
	fmt.Println(f1()) // 5
	fmt.Println(f2()) // 6
	fmt.Println(f3()) // 5
	fmt.Println(f4()) // 5
}
```

## （3）defer 后的函数形参在声明时确认

**defer注册要延迟执行的函数时该函数所有的参数都需要确定其值（预计算参数）**

```go
func calc(index string, a, b int) int {
	ret := a + b
	fmt.Println(index, a, b, ret)
	return ret
}

func main() {
	x := 1
	y := 2
	defer calc("AA", x, calc("A", x, y))
	x = 10
	defer calc("BB", x, calc("B", x, y))
	y = 20
}
```

结果：

```
A 1 2 3
B 10 2 12
BB 10 12 22
AA 1 3 4
```

分析：

```
1、第一个defer注册时，先要都确定参数，所以参数calc("A", x, y)要执行
  输出：'A', 1, 2, 3 且 defer1 calc("AA", 1, 3)
2、x=10
3、注册第二个defer，
  输出：'B', 10, 2, 12 且 defer2 calc("BB", 10, 12)
4、y=20
5、执行defer2：输出：'BB', 10, 12, 22
6、执行defer1：输出：'AA', 1, 3, 4
```

## （4）defer作用域

**defer 作用域仅为当前函数，在当前函数最后执行，所以不同函数下拥有不同的 defer 栈。**

## （5）发生 panic 时，已声明的 defer 会出栈执行

**当出现 panic 时，会触发已经声明的 defer 出栈执行，随后在再 panic，而在 panic 之后声明的 defer 将得不到执行。**

```go
func demo5_1() {
 defer fmt.Println(1)
 defer fmt.Println(2)
 defer fmt.Println(3)

 panic("没点赞异常") // 触发defer出栈执行

 defer fmt.Println(4) // 得不到执行
}
```

**正是利用这个特性，在 defer 中可以通过 recover 捕获 panic，防止程序崩溃。**

```go
func demo5_2() {
 defer func() {
     if err := recover(); err != nil {
         fmt.Println(err, "问题不大")
     }
 }()

 panic("没点赞异常") // 触发defer出栈执行

 // ...
}
```

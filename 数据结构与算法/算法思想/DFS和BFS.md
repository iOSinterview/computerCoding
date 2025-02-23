

# 一、深度优先搜索（DFS）

## 1、什么是DFS？

`DFS`即`Depth First Search`，深度优先搜索。简单地理解为一条路走到黑。那么什么叫一条路走到黑呢？假设我们想在如下的地图中走出一条最长的路，那么最粗暴的方式就是枚举出每一种情况。

![在这里插入图片描述](https://s2.loli.net/2024/08/25/p8R7PMH53JhfvaD.png)

因此，按照DFS一条路走到黑的思想，我们将会出现如下路线：

![在这里插入图片描述](https://s2.loli.net/2024/08/25/rCfs8jvzildLJuP.png)

简而言之，就是我们一头扎进去，撞了南墙，我就退一步，但是决不放弃，在原基础上做出局部的改变去尝试第二条路，直到所有的情况我都试了，实在没有其他情况了，那我就回到A，从头出发，再做选择，再一头扎进去，直到成功。

## 2、DFS 代码模板

### （1）全排列问题

![在这里插入图片描述](https://s2.loli.net/2024/08/25/tiI7f1OdDA3MLqe.png)

![在这里插入图片描述](https://s2.loli.net/2024/08/25/mLPvoKbf1iZRdax.png)![在这里插入图片描述](https://s2.loli.net/2024/08/25/B7v5YTxWr9KIyhm.png)

### （2）模板

`````go
// 全排列
func permute(N int) {

	ans := [][]int{}         // 记录答案
	path := []int{}          // 记录路径
	visit := make([]bool, N) // 标记是否访问过
	nums := make([]int, N)
	for i := 0; i < N; i++ {
		nums[i] = i + 1
	}

	var dfs func() // 定义函数
	dfs = func() {
		// 走到底，终止条件
		if len(path) == N {
			tmp := make([]int, N)
			copy(tmp, path)
			ans = append(ans, tmp)
		}
		// 循环
		for i, val := range nums {
			// 找没有访问过的节点
			if !visit[i] {
				// 做选择
				visit[i] = true
				path = append(path, val)
				dfs() // 一条路走到黑
				// 开始回溯，撤销选择
				path = path[:len(path)-1]
				visit[i] = false
			}
		}
	}
	dfs()
	fmt.Println(ans)
}
`````

# 二、广度优先搜索（BFS）

## 1、什么是BFS？

BFS即`Breadth First Search`，即广度优先搜索。如果说DFS是一条路走到黑的话，BFS就完全相反了。BFS会在每个岔路口都各向前走一步。因此其遍历顺序如下图所示：

![在这里插入图片描述](https://s2.loli.net/2024/08/25/1NZf6sQm9hSr7Gx.png)


我们发现每次搜索的位置都是距离当前节点最近的点。因此，**BFS是具有最短路的性质的**。为什么呢？这就类似于我们后面要学习的贪心策略。这里简单地介绍一下贪心，假设我们可以做出12次选择。我们想得到一个最好的方案。**那么我们可以在第一次选择的时候，做出当前最好的选择，在第二次选择的时候，再做出那时候最好的选择，由此积累。当我们在每次的选择面前，都做到了当前最好的选择，那么我们就可以由局部最优推出整体最优。**

这里也是类似的，我们可以在每次出发的时候，走到离自己最近的点，由此我们每次都保证走最近的，那从局部最近推整体最近，必有一条路是整体最近的。**所以我们可以利用BFS做最短路问题。**

## 2、BFS代码模板

## （1）走迷宫

![在这里插入图片描述](https://s2.loli.net/2024/08/25/HFunU26hYRAcGeW.png)

## （2）代码模板

`````go
package main

import "fmt"

// 走迷宫
func f(matrix [][]int, start, end []int) int {
	row, col := len(matrix), len(matrix[0])
	visit := make([][]int, row) // 标记是否走过，以及移动到当前位置的步数
	for i := range visit {
		visit[i] = make([]int, col)
		for j := range visit[i] {
			visit[i][j] = -1 // -1标识没被走过
		}
	}
	visit[start[0]][start[1]] = 0                    // 起点步数置为0
	dxy := [][]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}} // 方向向量：左、右、上、下
	queue := [][]int{{start[0], start[1]}}           // 初始位置入队
	var bfs func() int
	bfs = func() int {
		// 队列不为空
		for len(queue) > 0 {
			cur := queue[0] // 以队首为起点，展开搜索
			// 遍历四个方向走
			for _, dir := range dxy {
				x, y := cur[0]+dir[0], cur[1]+dir[1]
				newxy := []int{x, y}
				// 不越界 && 能走 && 没有访问过
				if x >= 0 && x < row && y >= 0 && y < col && matrix[x][y] == 0 && visit[x][y] == -1 {
					queue = append(queue, newxy)            // 将当前位置入队
					visit[x][y] = visit[cur[0]][cur[1]] + 1 // 记录走过，在当前位置的步数上+1
				}
			}
			queue = queue[1:] // 出队，所有四个方向已经走完
		}
		return visit[row-1][col-1] // 返回此时最右上角的步数
	}
	return bfs()
}

func main() {
	mp := [][]int{
		{0, 1, 0, 0, 0},
		{0, 1, 0, 1, 0},
		{0, 0, 0, 0, 0},
		{0, 1, 1, 1, 0},
		{0, 0, 0, 1, 0},
	}
	star := []int{0, 0}
	end := []int{4, 4}
	fmt.Println(f(mp, star, end))
}
`````

## 3、BFS注意的点

### （1）使用队列

BFS要保证的第一件事就是我们需要先走最近的，因此，队列的作用就是基于此的。

![在这里插入图片描述](https://s2.loli.net/2024/08/25/r8BDjOl9YGUZWvK.png)

![在这里插入图片描述](https://s2.loli.net/2024/08/25/TrVX24kZUCbQYhl.png)

![在这里插入图片描述](https://s2.loli.net/2024/08/25/FLH9cgMa2QUYkGT.png)

### （2）矩阵中方向向量的使用

如果是四个方向走，那么分别

|      | 左   | 右   | 上   | 下   |
| ---- | ---- | ---- | ---- | ---- |
| x    | -1   | 1    | 0    | 0    |
| y    | 0    | 0    | -1   | -1   |

如果是八个方向走，同样的当立也可以推导

### （3）为什么最后的输出是最短路径？

我们每个点都是同时向外拓展一步，并且只拓展一次。那么我们将其速度看作1步/次。每个点都向外探索一次。那么此时我们的次数可以类比为时间，由此每条路的速度和时间都是一样的，因此每条路的路程都是一样的。

而各个点都是从起点开始扩散的。我们看下面的例子：

![在这里插入图片描述](https://s2.loli.net/2024/08/25/Tr4kCxYplutfgMs.png)

**某时刻，绿色线到达了B点，此时各个路线的长度都是`L`，那么接下来再走的话，蓝色线的路程和黄色线的路程只会更长，因此其再到达B点的时候，必不如绿色线近。因此，第一次到达某个点的路线，就是最短的路线**

**由于`visit`数组中的点，踩过一次后，就不许再经过了。于是，我们惊奇地发现，每个点记录的路程都是从起点到该点的最短路！！！**

## 题目

### [102. 二叉树的层序遍历](https://leetcode.cn/problems/binary-tree-level-order-traversal/)

```go
func levelOrder(root *TreeNode) [][]int {
    // 借助一个队列，先压入根节点，出队头节点，将出队节点的左右孩子分别压入
    queue := []*TreeNode{}
    res := make([][]int, 0)
    if root == nil{
        return res
    }
    queue = append(queue, root)
    var tmpArr []int //存储当前层节点值
    for len(queue) > 0{
        levelLen := len(queue)  // 计算本层长度
        for levelLen > 0{
            node := queue[0]
            queue = queue[1:]   // 出队
            if node.Left != nil{
                queue = append(queue, node.Left)
            }
            if node.Right != nil{
                queue = append(queue, node.Right)
            }
            tmpArr = append(tmpArr, node.Val)   //存储本层节点值
            levelLen -= 1
        }
        res = append(res, tmpArr)
        tmpArr = []int{}    // 清空当前层
    }
    return res
}
```

### 跳马【hard】【BFS】

马是象棋(包括中国象棋只和国际象棋）中的棋子，走法是每步直一格再斜一格，即先横着或直着走一格，然后再斜着走一个对角线，可进可退，可越过河界，俗称马走 “日“ 字。

给项m行n列的棋盘(网格图)，棋盘上只有象棋中的棋子“马”，并目每个棋子有等级之分，等级为K的马可以跳1~k步(走的方式与象棋中“马”的规则一样，不可以超出棋盘位置)，**问是否能将所有马跳到同一位置**，如果存在，输出最少需要的总步数(每匹马的步数相加) ，不存在则输出-1。

**注:**允许不同的马在跳的过程中跳到同一位置，坐标为(x,y)的马跳一次可以跳到到坐标为(x+1,y+2),(x+1,y-2),(x+2,y+1),(x+2,y-1). (x-1,y+2),(x-1,y-2),(x-2,y+1),(x-2,y-1),的格点上，但是不可以超出棋盘范围。

#### 输入描述

第一行输入m,n代表m行n列的网格图棋盘(1 <= m,n <= 25);

接下来输入m行n列的网格图棋盘，如果第i行,第j列的元素为 “.” 代表此格点没有棋子，如果为数字k (1<= k <=9)，代表此格点存在等级为的“马”。

#### 输出描述

输出最少需要的总步数 (每匹马的步数相加)，不存在则输出-1。

#### 示例1

```
输入：
3 2
..
2.
..

输出：
0
```

#### 示例2

```
输入：
3 5
47.48
4744.
7....

输出：
17
```

#### 思路

- 采用BFS的思想，从当前节点出发，向外走k层
- 采用一个二维数组arrive表示当前格点走过的棋子个数
- 采用一个二维数组step表示当前马走到当前格点的最少步数
- 查找arrive中值为棋子个数的的格点，以及对应step中的步数，输出最小的。
- 如果arrive中没有值为棋子个数的格点，输出-1。

#### 代码

```go
package main

import (
	"fmt"
	"math"
	"strconv"
	"unicode"
)

type Index struct {
	x int
	y int
}

func main() {
	var m, n int
	fmt.Scan(&m, &n)
	tu := make([][]int, m)
	qiNum := 0
	for i := 0; i < m; i++ {
		var s string
		fmt.Scan(&s)
		tu[i] = make([]int, n)
		for j, v := range s {
			if unicode.IsNumber(rune(v)) {
				tu[i][j], _ = strconv.Atoi(string(s[j]))
				qiNum++
			}
		}
	}
	arriveCnt := make([][]int, m) // 记录到此格点马的个数
	stepCnt := make([][]int, m)   // 记录到此格点马的总步数
	for i := 0; i < m; i++ {
		arriveCnt[i] = make([]int, n)
		stepCnt[i] = make([]int, n)
	}

	var bfs func(queue []Index, k int, visited [][]bool)
	bfs = func(queue []Index, k int, visited [][]bool) {
		// 方向数组
		dicretion := [][]int{
			{1, 2}, {1, -2}, {2, 1}, {2, -1},
			{-1, 2}, {-1, -2}, {-2, 1}, {-2, -1},
		}
		step := 0
		for len(queue) > 0 && step <= k {
			cur := queue[0]
			// 遍历方向
			for _, dxy := range dicretion {
				var newIdx Index
				newIdx.x = cur.x + dxy[0]
				newIdx.y = cur.y + dxy[1]

				if newIdx.x >= 0 && newIdx.x < m && newIdx.y >= 0 && newIdx.y < n && visited[newIdx.x][newIdx.y] == false {
					arriveCnt[newIdx.x][newIdx.y]++
					queue = append(queue, newIdx)
					stepCnt[newIdx.x][newIdx.y] += step
					visited[newIdx.x][newIdx.y] = true
				}
			}
			step++
			// 方向遍历完，出队
			queue = queue[1:]
		}
	}

	// 遍历地图，对每个棋子进行BFS
	fmt.Println(tu)
	for i := 0; i < m; i++ {
		queue := make([]Index, 0)
		for j := 0; j < n; j++ {
			if tu[i][j] != 0 {
				// 初始化是否访问
				visited := make([][]bool, m)
				for i := 0; i < m; i++ {
					visited[i] = make([]bool, n)
				}
				// 将当前棋子加入队列，对当前棋子进行BFS，标记当前位置访问
				arriveCnt[i][j]++
				stepCnt[i][j] += 0
				visited[i][j] = true
				queue = append(queue, Index{i, j})
				bfs(queue, tu[i][j], visited)
			}
		}
	}
	//fmt.Println(arriveCnt)
	//fmt.Println(stepCnt)
	// 遍历arriveCnt,找出为qiNum的点，并输出最小的步数
	stepMin := math.MaxInt
	for i := 0; i < m; i++ {
		for j := 0; j < n; j++ {
			if arriveCnt[i][j] == qiNum {
				if stepCnt[i][j] < stepMin {
					stepMin = stepCnt[i][j]
				}
			}
		}
	}
	if stepMin == math.MaxInt {
		fmt.Println(-1)
	} else {
		fmt.Print(stepMin)
	}
}
```


package main

import "fmt"

// 城市聚集度

type unionFind struct {
	count  int
	parent []int
}

func New(n int) unionFind {
	parent := make([]int, n+1)
	for i := 1; i <= n; i++ {
		parent[i] = i
	}
	return unionFind{n, parent}
}

func (u *unionFind) Union(p, q int) {
	rp, rq := u.Find(p), u.Find(q)
	if rp == rq {
		return
	}
	u.parent[rp] = rq
	u.count--
}

func (u *unionFind) Connected(p, q int) bool {
	return u.Find(p) == u.Find(q)
}

func (u *unionFind) Find(p int) int {
	if p != u.parent[p] {
		u.parent[p] = u.Find(u.parent[p])
	}
	return u.parent[p]
}

func main() {
	var N int
	fmt.Scan(&N)
	// input := [][2]int{}
	u := New(N)
	for i := 1; i < N; i++ {
		var x, y int
		fmt.Scan(&x, &y)
		// input = append(input, [2]int{x, y})
		u.Union(x, y)
	}

	nums := make([]int, N+1)
	for i := 1; i <= N; i++ {
		for j := 1; j <= N; j++ {
			if i == j {
				continue
			}
			if u.Find(j) == i {
				nums[i]++
			}
		}
	}

	dp := make([]int, N+1)
	mx := N + 1
	for i := 1; i <= N; i++ {
		x := u.parent[i]
		dp[i] = max(nums[x]-nums[i], nums[i]+1)
		if mx > dp[i] {
			mx = dp[i]
		}
	}

	for i := 1; i <= N; i++ {
		if mx == dp[i] {
			fmt.Println(i)
		}
	}
}

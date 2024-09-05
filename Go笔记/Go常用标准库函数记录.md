



# Sort库

## sort.Slice()

`func Slice(x any, less func(i, j int) bool)`

给定less 函数的情况下对x进行排序。x不为切片时panic。不保证稳定性，保证稳定性使用SliceStable()函数。

```go
package main

import (
	"fmt"
	"sort"
)

func main() {
	people := []struct {
		Name string
		Age  int
	}{
		{"Gopher", 7},
		{"Alice", 55},
		{"Vera", 24},
		{"Bob", 75},
	}
	// 按name字母表的递增排序，返回字母ACII码更小的
	sort.Slice(people, func(i, j int) bool { return people[i].Name < people[j].Name })
	fmt.Println("By name:", people) // By name: [{Alice 55} {Bob 75} {Gopher 7} {Vera 24}]
	// 按Age实现年龄递增排序
	sort.Slice(people, func(i, j int) bool { return people[i].Age < people[j].Age })
	fmt.Println("By age:", people) // By age: [{Gopher 7} {Vera 24} {Alice 55} {Bob 75}]
}
```

# strings库

# strconv库

# Unicode库


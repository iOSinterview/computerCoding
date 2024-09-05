# Go中初始化一个二维数组

---

`````go
mark := make([][]bool, x)
for i := range mark {
mark[i] = make([]bool, y)
}
`````

`````
package main

import (
"bufio"
"fmt"
"os"
"strconv"
"strings"
)

func main() {
reader := bufio.NewReader(os.Stdin)
str, _ := reader.ReadString('\n')
strs := strings.Split(strings.TrimSpace(str), " ")
mp := map[int]bool{}
trans := map[string]int{
"A": 14,
"K": 13,
"Q": 12,
"J": 11,
}
for _, s := range strs {
if q, ok := trans[s]; ok {
mp[q] = true
} else {
t, _ := strconv.Atoi(s)
mp[t] = true
}
}

var flag bool
for i := 3; i < 15; i++ {
if !mp[i] {
continue
}
tmp := []int{}
for j := 0; j < 15; j++ {
if mp[i+j] {
tmp = append(tmp, i+j)
} else {
break
}
}

if len(tmp) >= 5 {
flag = true
for _, v := range tmp {
if v <= 10 {
fmt.Printf("%d ", v)
} else if v == 11 {
fmt.Printf("%c ", 'J')
} else if v == 12 {
fmt.Printf("%c ", 'Q')
} else if v == 13 {
fmt.Printf("%c ", 'K')
} else {
fmt.Printf("%c ", 'A')
}
}
fmt.Println()
i += len(tmp) - 1
}
}

if !flag {
fmt.Println("No")
}
}

`````


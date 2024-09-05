# Go语言中的进制转换

---

## 代码实现

### 十进制转其他进制

十进制转二进制：除 2 取余法，反复除2得到的余数即为二进制

```go
func DecimalToBinary(a int) string {
	b := ""
	for a > 0 {
		t := a % 2
		b = strconv.Itoa(t) + b
		a /= 2
	}
	return b
}
```

同样的，十进制转八进制、十六进制可以用除8、16代替.

但是16进制特殊点，换算的时候需要根据余数对应的字母，因此需要map。

```go
func DecimalToHex(a int) string {
	mp := map[int]string{
		10: "A", 11: "B", 12: "C",
		13: "D", 14: "E", 15: "F",
	}
	b := ""
	for a > 0 {
		t := a % 16
		if t > 9 {
			b = mp[t] + b
		} else {
			b = strconv.Itoa(t) + b
		}
		a /= 16
	}
	return b
}
```

### 其他进制转十进制

二进制转十进制：就跟我们计算一样，将二进制从低位到高位，每一位乘以2的相应幂次，累加。

```go
func BinaryToDecimal(b string) int {
	sum := 0
	for i := len(b) - 1; i >= 0; i-- {
		x, _ := strconv.Atoi(string(b[i]))
		sum += x * int(math.Pow(2, float64(len(b)-1-i)))
	}
	return sum
}
```

八进制、十六进制转十进制也是同样的思路

## Go标准库中的进制转换函数

### 十进制转其他进制

```go
strconv.FormatInt(int64(x),2)	// 将x转为二进制
strconv.FormatInt(int64(x),8)	// 将x转为八进制
strconv.FormatInt(int64(x),10)	// 将x转为十六进制
```

### 其他进制转十进制

```go
strconv.ParseInt(str, 2, 64)  // 将二进制str转为int64
strconv.ParseInt(str, 8, 64)  // 将八进制str转为int64
strconv.ParseInt(str, 16, 64) // 将十六进制str转为int64
```


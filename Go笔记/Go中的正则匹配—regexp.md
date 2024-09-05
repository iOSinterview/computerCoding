## Go中的正则匹配包——regexp

---

## 为什么要学**[正则表达式]**？

因为利用**正则表达式**可以非常方便的匹配我们想要的任何字符串。比如，在一大堆字符串中，我们想找包含“Go语言”并且以“架构师”结尾的所有字符串，利用正则表达式就能非常方便快速的查找出来：

```go
re, _ := regexp.Compile(`Go语言.*架构师`)
strs := "kfhewGo语言jiohjnfew.fewujj架构师"
fmt.Println(re.FindString(strs))  // 输出：Go语言jiohjnfew.fewujj架构师
fmt.Println(re.MatchString(strs))  // 输出：true
```

Go语言的正则表达式在regexp包中。一般，在使用正则表达式之前，我们会把模式字符串编译成正则表达式实例。

## 语法规则

字符

![image-20240831163547011](https://s2.loli.net/2024/08/31/q6PTkvgcK2I8zjV.png)

![image-20240831163826118](https://s2.loli.net/2024/08/31/nJc6BgyPFlNSoDm.png)

![image-20240831163856954](https://s2.loli.net/2024/08/31/EwLZlB6GARqrYFQ.png)

![image-20240831163917632](https://s2.loli.net/2024/08/31/J26o7qUO3aGtHSZ.png)

![image-20240831163944760](https://s2.loli.net/2024/08/31/8ujDOYFemspdxIv.png)

了解了模式字符串规则，我们来看一下`regexp`的编译函数。

## 编译函数

### `Compile`

```go
regexp.Compile(expr string) (*Regexp, error)
```


Compile函数会尝试把模式字符串`expr`编译成正则表达式。什么是模式字符串？也就是本文最开始例子中的“Go语言.*架构师”。这个模式字符串表示先匹配“Go语言”这个字符串，紧接着匹配任意个字符串，最后匹配“架构师”这个字符串。

`*Regexp`就是返回的编译后的正则表达式实例指针，然后我们可以利用这个正则表达式做很多字符串操作。那么，我们可以写出哪些模式字符串？哪些模式字符串才是合法的？先来看一下模式字符串的语法规则。

### `MustCompile(expr string)`

```go
regexp.MustCompile(expr string) *Regexp
```

Compile函数编译了正则表达式之后，如果正则表达式不合法，会返回error。MustCompile在编译正则表达式过程中，如果正则表达式不合法，会抛出panic异常。MustCompile在一些正则表达式全局变量初始化的场景下很有用。

```go
re := regexp.MustCompile(`Go语言.*架构师`)
strs := "kfhewGo语言jiohjnfew.fewujj架构师"
fmt.Println(re.FindString(strs))  // 输出：Go语言jiohjnfew.fewujj架构师
fmt.Println(re.MatchString(strs))  // 输出：true
```

## 匹配

编译得到正则表达式实例之后，就可以用来任意匹配字符串了，最典型的就是Match函数。

`regexp`包中有两个`Match`函数：

1. `Regexp的Match方法；`
2. `regexp包的Match函数。`

如果正则表达式能匹配到传入的字符串（或字节切片），则会返回true。

```go
func (re *Regexp) MatchString(s string) bool {
  return re.doMatch(nil, nil, s)
}
func (re *Regexp) Match(b []byte) bool {
  return re.doMatch(nil, b, "")
}
 
func MatchString(pattern string, s string) (matched bool, err error) {
  re, err := Compile(pattern)
  if err != nil {
    return false, err
  }
  return re.MatchString(s), nil
}
func Match(pattern string, b []byte) (matched bool, err error) {
  re, err := Compile(pattern)
  if err != nil {
    return false, err
  }
  return re.Match(b), nil
}
```

可以看到，regexp包的Match函数内部先把模式字符串pattern编译成了正则表达式实例，然后在调用Regexp的Match方法来实现的。模式字符串编译成正则表达式实例的这个操作，如果是在大规模字符串处理中，是非常耗时的。因此我们可以先把编译好的正则表达式实例缓存在全局变量中，要用的时候直接匹配就好了。

```go
re := regexp.MustCompile(`^Go语言.*架构师`)
strs := "kfhewGo语言jiohjnfew.fewujj架构师"
fmt.Println(re.MatchString(strs))  // 输出：false
```

## 查找

正则表达式的查找是非常有用的功能之一。在我们写爬虫的时候，拿到了网页的源代码，如果想从html中提取出下一页的连接，我们就可以写正则表达式查找出来。查找最典型的函数就`FindString`和`FindStringSubmatch`。我们来看一下，主要有以下几类：

```go
// 查找能匹配的字符串，查找所匹配的字符串的起止位置
func (re *Regexp) FindString(s string) string
func (re *Regexp) FindStringIndex(s string) (loc []int)
func (re *Regexp) FindReaderIndex(r io.RuneReader) (loc []int)
 
// 查找能匹配的字符串和所有的匹配组
func (re *Regexp) FindStringSubmatch(s string) []string
func (re *Regexp) FindStringSubmatchIndex(s string) []int
func (re *Regexp) FindReaderSubmatchIndex(r io.RuneReader) []int
 
// 查找所有能匹配的字符串（最多查找n次。如果n为负数，则查找所有能匹配到的字符串，以切片形式返回）
func (re *Regexp) FindAllString(s string, n int) []string
func (re *Regexp) FindAllStringIndex(s string, n int) [][]int
 
// 查找所有能匹配的字符串（最多查找n次。如果n为负数，则查找所有能匹配到的字符串，以切片形式返回）和所有的匹配组
func (re *Regexp) FindAllStringSubmatch(s string, n int) [][]string
func (re *Regexp) FindAllStringSubmatchIndex(s string, n int) [][]int
```

### 第一类：`FindString`和`FindStringIndex`

`FindString`函数会查找第一个能被正则表达式匹配到的字符串，并返回匹配到的字符串。`FindStringIndex`会返回匹配到的字符串的起止位置。

来看一下例子

```go
re := regexp.MustCompile(`Go语言(.*?)架构师`)  // 此处小括号中的问号表示勉强型匹配，见下文第五点
str := "前缀____Go语言_中间字符串111_架构师____中间字符串222____Go语言_中间字符串333_架构师"
fmt.Println(re.FindString(str))  // 输出：Go语言_中间字符串111_架构师
fmt.Println(re.FindStringIndex(str))  // 输出：[7 44]
```

可以看到，`FindString`找到了字符串中第一个被正则表达式匹配到的字符串，`FindStringIndex`返回了它的起止位置（由此可见，`FindStringIndex`结果切片只会包含2个元素）。如果`FindString`找不到（结果返回空字符串，不是nil），则F`indStringIndex`会返回`nil`。


### 第二类：FindStringSubmatch和FindStringSubmatchIndex

FindStringSubmatch函数不仅会查找第一个能被正则表达式匹配到的字符串，还会找出其中匹配组所匹配到的字符串（即正则表达式中小括号里的内容），会放在切片中一起返回。`FindStringSubmatchIndex`不仅会返回匹配到的字符串的起止位置，还会返回匹配组所匹配到的字符串起止位置。

这段文字比较绕，来看一下例子就明白了。

```go
re := regexp.MustCompile(`Go语言(.*?)架构师`)
str := "前缀____Go语言_中间字符串111_架构师____中间字符串222____Go语言_中间字符串333_架构师"
fmt.Println(re.FindStringSubmatch(str))  // 输出：[Go语言_中间字符串111_架构师 _中间字符串111_]
fmt.Println(re.FindStringSubmatchIndex(str))  // 输出：[7 44 15 35]
```

FindStringSubmatch结果中的第一个元素“Go语言_中间字符串111_架构师”就是整个正则表达式所匹配到的第一个字符串；结果中的第二个元素就是匹配组（即正则表达式的小括号内的内容）所匹配到的第一个字符串。匹配组是干啥用的？匹配组就是在整个正则表达式的匹配结果上，再进行的一次匹配。

理解了FindStringSubmatch，那么FindStringSubmatchIndex就自然而然理解了。

### 第三类：`FindAllString`和`FindAllStringIndex`

和FindString不一样，FindString会查找第一个能被正则表达式匹配到的字符串。FindAllString函数会查找n个（n是FindAllString的第二个参数）能被正则表达式匹配到的字符串，并返回所有匹配到的字符串所组成的切片。FindAllStringIndex会返回所有匹配到的字符串的起止位置，结果是个二维切片。如果n为整数，则最多匹配n次；如果n为负数，则会返回所有匹配结果。

来看一下例子：

```go
re := regexp.MustCompile(`Go语言(.*?)架构师`)
str := "前缀____Go语言_中间字符串111_架构师____中间字符串222____Go语言_中间字符串333_架构师"
fmt.Println(re.FindAllString(str, -1))  // 输出：[Go语言_中间字符串111_架构师 Go语言_中间字符串333_架构师]
fmt.Println(re.FindAllStringIndex(str, -1))  // 输出：[[7 44] [64 101]]
```

第二个参数传入-1，表示要返回所有匹配结果。可以看到，有匹配结果时，FindAllStringIndex的结果是个二维切片。

### 第四类：FindAllStringSubmatch和FindAllStringSubmatchIndex

和第二类一样，都是在FindAllString的结果上，再返回匹配组所匹配到的内容。看一下例子：

```go
re := regexp.MustCompile(`Go语言(.*?)架构师`)
str := "前缀____Go语言_中间字符串111_架构师____中间字符串222____Go语言_中间字符串333_架构师"
fmt.Println(re.FindAllStringSubmatch(str, -1))  // 输出：[[Go语言_中间字符串111_架构师 _中间字符串111_] [Go语言_中间字符串333_架构师 _中间字符串333_]]
fmt.Println(re.FindAllStringSubmatchIndex(str, -1))  // 输出：[[7 44 15 35] [64 101 72 92]]
```

对于字符串str，FindAllString能查找到两个匹配结果，在这两个匹配结果上，FindAllStringSubmatch再返回匹配组所匹配到的内容，那么结果就是下面这个二维切片了：

```go
[[Go语言_中间字符串111_架构师 _中间字符串111_] [Go语言_中间字符串333_架构师 _中间字符串333_]]
```

这四类函数在解析爬虫网页的时候特别有用。

## 替换

替换是正则表达式中另一个非常有用的功能。替换主要有三个函数：

```go
func (re *Regexp) ReplaceAllString(src, repl string) string
func (re *Regexp) ReplaceAllLiteralString(src, repl string) string
func (re *Regexp) ReplaceAllStringFunc(src string, repl func(string) string) string
```

### ReplaceAllString

ReplaceAllString会把第一个参数所表示的字符串中所有匹配到的内容用第二个参数代替。在第二个参数中，可以使用$符号来引用匹配组所匹配到的内容。$0表示第0个匹配组所匹配到的内容，即整个正则表达式所匹配到的内容。$1表示第一个匹配组所匹配到的内容，即正则表达式中第一个小括号内的正则表达式所匹配到的内容。

来看一下例子：

```go
re := regexp.MustCompile(`Go语言(.*?)架构师`)
str := "前缀____Go语言_中间字符串111_架构师____中间字符串222____Go语言_中间字符串333_架构师"
fmt.Println(re.ReplaceAllString(str, "$1"))  // 输出：前缀_____中间字符串111_____中间字符串222_____中间字符串333_
```

这个例子把str字符串中所有被正则表达式所匹配到的内容用第一个匹配组的内容进行替换。“Go语言(.*?)架构师”这个正则表达式能匹配到字符串“Go语言_中间字符串111_架构师”，而第一个匹配组（即.*?）所匹配到的内容是“_中间字符串111_”，所以最终结果就是“前缀_____中间字符串111_____中间字符串222_____中间字符串333_”。

### ReplaceAllLiteralString

跟ReplaceAllString不一样，ReplaceAllLiteralString中的第二个参数不能引用匹配组的内容，会把第二个参数当做字符串字面量去做替换。

看一下例子：

```go
re := regexp.MustCompile(`Go语言(.*?)架构师`)
str := "前缀____Go语言_中间字符串111_架构师____中间字符串222____Go语言_中间字符串333_架构师"
fmt.Println(re.ReplaceAllLiteralString(str, "$1"))  // 输出：前缀____$1____中间字符串222____$1
```

可以看到跟ReplaceAllString的结果不一样。

### ReplaceAllStringFunc

ReplaceAllStringFunc函数的第二个参数是个函数，表示我们可以自己编写函数来决定如何替换掉匹配到的字符串。

一起来看一下三个函数的使用案例：

```go
re := regexp.MustCompile(`Go语言(.*?)架构师`)
str := "前缀____Go语言_中间字符串111_架构师____中间字符串222____Go语言_中间字符串333_架构师"
fmt.Println(re.ReplaceAllStringFunc(str, func(s string) string {
  return "Q" + s + "Q"
}))
```

我们在匹配到的字符串前后都加了一个字母Q，可以看到，输出符合预期：

```go
前缀____QGo语言_中间字符串111_架构师Q____中间字符串222____QGo语言_中间字符串333_架构师Q
```

## 三种模式

任何语言的正则表达式匹配都避免不了正则表达式的三种匹配模式：贪婪型、勉强型、占有型。如果你是语言深耕者，一定要对这三种模式了如指掌。

贪婪型属于正常的表示（平时写的那些），勉强型则在后面加个“问号”，占有型加个“加号”，都只作用于前面的问号、星号、加号、大括号，因为前面如果没有这些，就变成普通的问号和加号了（也就是变成贪婪型了）。

- 贪婪型匹配模式表示尽可能多的去匹配字符。


- 勉强型匹配模式表示尽可能少的去匹配字符。


- 占有型匹配模式表示尽可能做完全匹配。


### 贪婪型

贪婪型匹配模式的正则表达式形式为星号或者加号。我们知道，星号表示匹配0个或任意多个字符，加号表示匹配1个或者任意多个字符。

贪婪型匹配，先一直匹配到最后，发现最后的字符不匹配时，往前退一格再尝试匹配，不匹配时再退一格。

看一下例子就很明白了：

```go
re := regexp.MustCompile(`我是.*字符串`)
str := "我是第1个字符串_我是第2个字符串"
fmt.Println(re.FindString(str))  // 输出：我是第1个字符串_我是第2个字符串
```

这个例子的输出是“我是第1个字符串_我是第2个字符串”，而不是“我是第1个字符串”。这个结果和勉强型匹配形成了强烈的对比。

### 勉强型

勉强型匹配模式的正则表达式形式为星号（`*`）或者加号（`+`），后面再加个问号（注意与贪婪型的区别）。我们知道，`*`表示匹配0个或任意多个字符，`+`表示匹配1个或者任意多个字符。后面加个问号表示尽可能少的去匹配字符。

看一下例子就很明白了，还是上面那个例子，在正则表达式的星号后面加个问号：

```go
re := regexp.MustCompile(`我是.*?字符串`)
str := "我是第1个字符串_我是第2个字符串"
fmt.Println(re.FindString(str))  // 输出：我是第1个字符串
```

这个例子的输出是“我是第1个字符串”。和贪婪型匹配结果形成了强烈的对比。

### 占有型

占有型匹配模式的正则表达式形式为星号或者加号，后面再加个加号（注意与贪婪型、勉强型的区别）。我们知道，星号表示匹配0个或任意多个字符，加号表示匹配1个或者任意多个字符。后面加个加号表示正则表达式必须完全匹配整个字符串。

Go语言中正则表达式没有“占有型”。

我们可以尝试一下，还是上面那个例子，在正则表达式的星号后面加个加号：

```go
re := regexp.MustCompile(`我是.*+字符串`)
str := "我是第1个字符串_我是第2个字符串"
fmt.Println(re.FindString(str))
```

发现编译时报错。Go语言中如果想实现完全匹配，在正则表达式中使用“^”和“$”表示首尾就好了。


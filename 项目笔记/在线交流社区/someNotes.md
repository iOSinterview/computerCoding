# 记录项目中的一些小细节点

# `r := gin.New()`与 `r := gin.Default()`的区别

项目`./routers/routers.go`中的 `注册路由：SetupRouter`方法。没有使用`r := gin.Default()`进行路由初始化，而是使用`r := gin.New()`，两者区别如下：

`r := gin.New()` 和 `r := gin.Default()` 是 Gin 框架中常见的两种方式来创建一个 Gin 的路由引擎实例。

- **gin.New()**:
  - `gin.New()` 创建一个不包含任何中间件的新的路由引擎实例。
  - 这意味着你需要手动添加需要的中间件，如日志记录、恢复、认证等。
  - 适用于需要完全控制中间件添加的场景，能够根据自身需求精确地选择所需的中间件。
- **gin.Default()**:
  - `gin.Default()` 创建一个默认的 Gin 路由引擎实例，已经预先添加了 Logger 和 Recovery 中间件。
  - Logger 中间件用于记录请求信息，Recovery 中间件用于在发生恐慌时恢复。
  - 这样的默认设置适用于大多数情况下，尤其是在快速搭建原型或者小型项目时，可以减少一些基本的配置工作。

# splx中的`db.Get()`、`db.Select()`、`db.Exec()`

-  `db.Get` 用于执行查询并将单个结果映射到指定的目标对象中。查询结果有多个会报错。
- `db.Select` 用于执行查询并将多行结果映射到指定的切片或集合中。
- `db.Exec()` 用于执行 `INSERT、UPDATE、DELETE` 等不返回行数据的操作。


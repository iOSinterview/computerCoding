[TOC]



# 接口定义

路由组：`v1 := r.Group("/api/v1")`

**非权限接口：**

| 接口             | 类型 | 接口调用说明                                 |
| :--------------- | ---- | -------------------------------------------- |
| 用户业务         |      |                                              |
| /login           | POST | 用户登录                                     |
| /signup          | POST | 用户注册                                     |
|                  |      |                                              |
| 帖子业务         |      |                                              |
| /posts           | GET  | 分页展示帖子列表                             |
| /posts2          | GET  | 根据社区id及时间或者分数排序分页展示帖子列表 |
| /post/:id        | GET  | 根据ID查询帖子详情                           |
| /search          | GET  | 搜索业务-搜索帖子                            |
|                  |      |                                              |
| 社区业务         |      |                                              |
| /community       | GET  | 获取分类社区列表                             |
| /community/:id   | GET  | 根据ID查找社区详情                           |
|                  |      |                                              |
| /github_trending | GET  | Github热榜                                   |

**JWT权限接口**：根据登录用户的token以及user_id才能执行的功能

`v1.Use(middlewares.JWTAuthMiddleware()) // 应用JWT认证中间件`

| 接口       | 类型 | 接口调用说明     |
| ---------- | ---- | ---------------- |
| /post      | POST | 创建帖子         |
| /vote      | POST | 点赞             |
| /comment   | POST | 评论             |
| /recomment | POST | 回复评论（待写） |
| /comment   | GET  | 展开评论列表     |

# MySQL表设计

## `user表`

| 字段          | 类型          | 说明                                                 |
| :------------ | ------------- | ---------------------------------------------------- |
| `id`          | `bigint(20)`  | `NOT NULL AUTO_INCREMENT（非空自增）`                |
| `user_id`     | `bigint(20)`  | `NOT NULL`                                           |
| `user_name`   | `varchar(64)` | `COLLATE utf8mb4_general_ci（字符集和校对规则）非空` |
| `password`    | `varchar(64)` | `COLLATE utf8mb4_general_ci（字符集和校对规则）非空` |
| `email`       | `varchar(64)` | `COLLATE utf8mb4_general_ci（字符集和校对规则）`     |
| `gender`      | `tinyint(4)`  | `非空，默认为 0（未知）`                             |
| `create_time` | `timestamp`   | `默认为当前时间戳`                                   |
| `update_time` | `timestamp`   | `更新时自动更新时间`                                 |

**`user`索引：**

```go
PRIMARY KEY (`id`),			// 主键
UNIQUE KEY `idx_username` (`username`) USING BTREE, // 建立在username列上的唯一索引
UNIQUE KEY `idx_user_id` (`user_id`) USING BTREE // 建立在user_id列上的唯一索引
```

```sql
CREATE TABLE `user` (
    `id` bigint(20) NOT NULL AUTO_INCREMENT,
    `user_id` bigint(20) NOT NULL,
    `username` varchar(64) COLLATE utf8mb4_general_ci NOT NULL,
    `password` varchar(64) COLLATE utf8mb4_general_ci NOT NULL,
    `email` varchar(64) COLLATE utf8mb4_general_ci,
    `gender` tinyint(4) NOT NULL DEFAULT '0',
    `create_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
    `update_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_username` (`username`) USING BTREE,
    UNIQUE KEY `idx_user_id` (`user_id`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
```

## `community表`

```sql
CREATE TABLE `community` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `community_id` int(10) unsigned NOT NULL,
  `community_name` varchar(128) COLLATE utf8mb4_general_ci NOT NULL,
  `introduction` varchar(256) COLLATE utf8mb4_general_ci NOT NULL,
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_community_id` (`community_id`),
  UNIQUE KEY `idx_community_name` (`community_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
```

社区表要管理员先创建好。

## `post表`

```sql
CREATE TABLE `post` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `post_id` bigint(20) NOT NULL COMMENT '帖子id',
  `title` varchar(128) COLLATE utf8mb4_general_ci NOT NULL COMMENT '标题',
  `content` varchar(8192) COLLATE utf8mb4_general_ci NOT NULL COMMENT '内容',
  `author_id` bigint(20) NOT NULL COMMENT '作者的user_id',
  `community_id` bigint(20) NOT NULL COMMENT '所属社区',
  `status` tinyint(4) NOT NULL DEFAULT '1' COMMENT '帖子状态',
  `create_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_post_id` (`post_id`),		// 唯一索引
  KEY `idx_author_id` (`author_id`),		// 普通索引
  KEY `idx_community_id` (`community_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
```

## `comment表`

```sql
CREATE TABLE `comment` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `comment_id` bigint(20) unsigned NOT NULL,
  `content` text COLLATE utf8mb4_general_ci NOT NULL,
  `post_id` bigint(20) NOT NULL,
  `author_id` bigint(20) NOT NULL,
  `parent_id` bigint(20) NOT NULL DEFAULT '0',
  `status` tinyint(3) unsigned NOT NULL DEFAULT '1',  // 评论是否审核
  `create_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_comment_id` (`comment_id`),
  KEY `idx_author_Id` (`author_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
```

# Redis的key-value设计

```go
const (
	KeyPostInfoHashPrefix = "bluebell-plus:post:"		// string;value为post_id
	KeyPostTimeZSet       = "bluebell-plus:post:time"  // zset;帖子及发帖时间定义
	KeyPostScoreZSet      = "bluebell-plus:post:score" // zset;帖子及投票分数定义
	//KeyPostVotedUpSetPrefix   = "bluebell-plus:post:voted:down:"
	//KeyPostVotedDownSetPrefix = "bluebell-plus:post:voted:up:"
	KeyPostVotedZSetPrefix    = "bluebell-plus:post:voted:" // zSet;记录用户及投票类型;参数是post_id
	KeyCommunityPostSetPrefix = "bluebell-plus:community:"  // set保存每个分区下帖子的id
)
```

![postkey](https://s2.loli.net/2024/08/21/BtHTw1zphcR6YPU.jpg)

![votekey](https://s2.loli.net/2024/08/21/MdHWfVY5O8SBqx2.jpg)

![community](https://s2.loli.net/2024/08/21/OX6VCnZwc4Ie5qR.jpg)

![img](https://s2.loli.net/2024/08/21/wBVemWLFfgbRpSH.png)

# 结构体设计

## user类

```go
// User 定义请求参数结构体
type User struct {
	UserID       uint64 `json:"user_id,string" db:"user_id"` // 指定json序列化/反序列化时使用小写user_id
	UserName     string `json:"username" db:"username"`
	Password     string `json:"password" db:"password"`
	Email        string `json:"email" db:"gender"`  // 邮箱
	Gender       int    `json:"gender" db:"gender"` // 性别
	AccessToken  string
	RefreshToken string
}

// RegisterForm 注册请求参数
type RegisterForm struct {
	UserName        string `json:"username" binding:"required"`  // 用户名
	Email           string `json:"email" binding:"required"`     // 邮箱
	Gender          int    `json:"gender" binding:"oneof=0 1 2"` // 性别 0:未知 1:男 2:女
	Password        string `json:"password" binding:"required"`  // 密码
	ConfirmPassword string `json:"confirm_password" binding:"required,eqfield=Password"`
}

// LoginForm 登录请求参数
type LoginForm struct {
	UserName string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}
```

这里可能需要为这些类型实现自定义的UnmarshalJSON方法。

## post 结构体

```go
// Post 帖子Post结构体 内存对齐概念 字段类型相同的对齐 缩小变量所占内存大小
type Post struct {
	PostID      uint64    `json:"post_id,string" db:"post_id"`
	AuthorId    uint64    `json:"author_id" db:"author_id"`
	CommunityID uint64    `json:"community_id" db:"community_id" binding:"required"`
	Status      int32     `json:"status" db:"status"`
	Title       string    `json:"title" db:"title" binding:"required"`
	Content     string    `json:"content" db:"content" binding:"required"`
	CreateTime  time.Time `json:"-" db:"create_time"`
	UpdateTime  time.Time `json:"-" db:"update_time"`
}
```

## community 结构体

```go
// Community Community结构体
type Community struct {
	CommunityID   uint64 `json:"community_id" db:"community_id"`
	CommunityName string `json:"community_name" db:"community_name"`
}

// CommunityDetail 社区详情model
type CommunityDetail struct {
	CommunityID   uint64    `json:"community_id" db:"community_id"`
	CommunityName string    `json:"community_name" db:"community_name"`
	Introduction  string    `json:"introduction,omitempty" db:"introduction"` // omitempty 当Introduction为空时不展示
	CreateTime    time.Time `json:"create_time" db:"create_time"`
}
```

## comment 结构体

```go
type Comment struct {
	PostID     uint64    `db:"post_id" json:"post_id"`
	ParentID   uint64    `db:"parent_id" json:"parent_id"`
	CommentID  uint64    `db:"comment_id" json:"comment_id"`
	AuthorID   uint64    `db:"author_id" json:"author_id"`
	Content    string    `db:"content" json:"content"`
	CreateTime time.Time `db:"create_time" json:"create_time"`
	UpdateTime time.Time `db:"update_time" json:"update_time"`
}

// 包括子评论的评论
type CommentPlus struct {
	Comment
	SubComment []Comment `db:"sub_comment" json:"sub_comment"`
}
```

# 接口实现说明

# 一、用户业务

用户结构：

```go
// User 定义请求参数结构体
type User struct {
	UserID       uint64 `json:"user_id,string" db:"user_id"` // 指定json序列化/反序列化时使用小写user_id
	UserName     string `json:"username" db:"username"`
	Password     string `json:"password" db:"password"`
	Email        string `json:"email" db:"gender"`  // 邮箱
	Gender       int    `json:"gender" db:"gender"` // 性别
	AccessToken  string
	RefreshToken string
}
```

## 1、注册

`v1.POST("/signup", controller.SignUpHandler)`  

```go
c *gin.Context  // 上下文

// RegisterForm 注册请求参数
type RegisterForm struct {
	UserName        string `json:"username" binding:"required"`  // 用户名
	Email           string `json:"email" binding:"required"`     // 邮箱
	Gender          int    `json:"gender" binding:"oneof=0 1 2"` // 性别 0:未知 1:男 2:女
	Password        string `json:"password" binding:"required"`  // 密码
	ConfirmPassword string `json:"confirm_password" binding:"required,eqfield=Password"`
}
```

- `controller.SignUpHandler`层。通过`c.ShouldBindJSON(&fo)`验证数据有效性，解析到`fo`，有错误则返回。没有则进入业务逻辑处理层`logic.SignUp(fo)`。
- `logic.SignUp(fo)`层。
  1. 通过`UserName`查询数据库判断用户是否存在：`mysql.CheckUserExist(p.UserName)`
  2. 雪花算法生成`userID`：`userId, err := snowflake.GetID()`
  3. 构造`User`实例，将其保存入数据库。`mysql.InsertUser(u)`
- dao层。
  - `mysql.CheckUserExist(p.UserName)`
    1. 查询语句：`select count(user_id) from user where username = ?`，统计user表中等于给定用户名的记录数量。
    2. `db.Get(&count, sqlStr, username)`。使用数据库连接 `db` 执行查询，并将结果存储到 `count` 变量中。`count>0`则表明已经有用户存在。
  - `mysql.InsertUser(u)`
    1. 使用标准库`crypto/md5`算法对密码加密。
    2. 执行SQL语句：`sqlstr := `insert into user(user_id,username,password,email,gender) values(?,?,?,?,?)`
    3. `db.Exec(sqlstr, user.UserID, user.UserName, user.Password, user.Email, user.Gender)`

## 2、登录

`v1.POST("/login", controller.LoginHandler)`。

- `controller`层先校验参数。
- 业务处理。`user, err := logic.Login(u)`
  - `mysql.Login(user)`。
  - 生成JWT。`accessToken, refreshToken, err := jwt.GenToken(user.UserID, user.UserName)`
- dao层：`mysql.Login(user)`==> 验证用户是否存在以及密码是否正确。
  1. 先对密码加密处理，然后用户名查询。`sqlStr := "select user_id, username, password from user where username = ?"`，`db.Get(user, sqlStr, user.UserName)`
  2. 查询数据库出错，返回。用户不存在，返回。
  3. 将生成加密密码与查询到的密码比较，验证密码是否正确。

# 二、帖子业务

帖子结构：

```go
// Post 帖子Post结构体 内存对齐概念 字段类型相同的对齐 缩小变量所占内存大小
type Post struct {
	PostID      uint64    `json:"post_id,string" db:"post_id"`
	AuthorId    uint64    `json:"author_id" db:"author_id"`
	CommunityID uint64    `json:"community_id" db:"community_id" binding:"required"`
	Status      int32     `json:"status" db:"status"`
	Title       string    `json:"title" db:"title" binding:"required"`
	Content     string    `json:"content" db:"content" binding:"required"`
	CreateTime  time.Time `json:"-" db:"create_time"`
	UpdateTime  time.Time `json:"-" db:"update_time"`
}
```



## 1、分页展示帖子详情

`v1.GET("/posts", controller.PostListHandler)    // 分页展示帖子列表`

- 这里主要有两个参数：page 和 size。page表示第几页，size表示煤业有几个帖子。默认page=1，size=10；用户可以自己输入：`/posts?page=2&size=5`

- 在业务层获取数据并返回。`data, err := logic.GetPostList(page, size)`。

  - 首先会获取帖子列表：`postList, err := mysql.GetPostList(page, size)`
  - 接着还要根据列表中的帖子中作者ID查询到帖子对应的作者信息`user, err := mysql.GetUserByID(post.AuthorId)`以及根据帖子中社区ID查询到对应的社区信息`community, err := mysql.GetCommunityByID(post.CommunityID)`。然后将帖子信息、作者信息、社区信息进行拼接。

- dao层查询帖子列表如下。时间降序排列，即新创建的放前面。`limit 偏移量,限制数量`

  ```go
  sqlStr := `select post_id, title, content, author_id, community_id, create_time
  	from post
  	ORDER BY create_time
  	DESC 
  	limit ?,?
  	`
  	posts = make([]*models.Post, 0, 2) // 0：长度  2：容量
  	err = db.Select(&posts, sqlStr, (page-1)*size, size)
  ```

## 2、根据社区id、时间、或者点赞数排序分页展示帖子列表

`v1.GET("/posts2", controller.PostList2Handler)`

`// GET请求参数(query string)： /api/v1/posts2?page=1&size=10&order=time`

logic层`logic.GetPostListNew(p)`这里有两种情况，如果没有community_id，则查询全部帖子。如果有community_id，则根据community_id查询帖子。

**（1）查询全部帖子：GetPostList2**

```go
// 在redis中分别对应的key为
// redis key 注意使用命名空间的方式，方便查询和拆分
const (
	KeyPostInfoHashPrefix = "bluebell-plus:post:"
	KeyPostTimeZSet       = "bluebell-plus:post:time"  // zset;帖子及发帖时间定义
	KeyPostScoreZSet      = "bluebell-plus:post:score" // zset;帖子及投票分数定义
	//KeyPostVotedUpSetPrefix   = "bluebell-plus:post:voted:down:"
	//KeyPostVotedDownSetPrefix = "bluebell-plus:post:voted:up:"
	KeyPostVotedZSetPrefix    = "bluebell-plus:post:voted:" // zSet;记录用户及投票类型;参数是post_id
	KeyCommunityPostSetPrefix = "bluebell-plus:community:"  // set保存每个分区下帖子的id
)
```

1. 首先从mysql获取帖子列表总数：`total, err := mysql.GetPostTotalCount()`

2. 然后根据参数中的排序规则（order=time或者score）去resdis中查询id列表。`ids, err := redis.GetPostIDsInOrder(p)`。查询语句：`client.ZRevRange(key, start, end).Result()`.

   `key=KeyPostTimeZSet`。

3. 接着根据ids去redis中查询好每篇帖子的投票数：

   ```go
   for _, id := range ids {
       key := KeyPostVotedZSetPrefix + id
       // 查找key中分数是1的元素数量 -> 统计每篇帖子的赞成票的数量
       v := client.ZCount(key, "1", "1").Val()
       data = append(data, v)
   }
   ```

   client.ZCount(key, "1", "1")用于计算指定有序集合 `key` 中 score 值在 "1" 到 "1" 范围内的成员数量。

4. 根据id去数据库查询帖子详细信息。// 返回的数据还要按照我给定的id的顺序返回  order by FIND_IN_SET(post_id, ?)

   ```go
   // GetPostListByIDs 根据给定的id列表查询帖子数据
   func GetPostListByIDs(ids []string) (postList []*models.Post, err error) {
   	sqlStr := `select post_id, title, content, author_id, community_id, create_time
   	from post
   	where post_id in (?)
   	order by FIND_IN_SET(post_id, ?)`
   	// 动态填充id
   	query, args, err := sqlx.In(sqlStr, ids, strings.Join(ids, ","))
   	if err != nil {
   		return
   	}
   	// sqlx.In 返回带 `?` bindvar的查询语句, 我们使用Rebind()重新绑定它
   	query = db.Rebind(query)
   	err = db.Select(&postList, query, args...)
   	return
   }
   ```

5. 组合数据， 将帖子的作者及分区信息查询出来填充到帖子中。

**（2）根据`communityID`分页排序展示帖子列表：`GetCommunityPostList(p)`**

1. 从mysql获取该社区下帖子列表总数`total`。

2. 根据参数 p 中的排序规则去`redis`查询 id列表 `ids`。

3. 按照`ids`去`redis`查询帖子的投票数。

4. 根据 `ids`去`mysql`查询帖子详细信息，返回的数据按照id的顺序返回。`order by FIND_IN_SET(post_id,?)`

   ```sql
   select post_id, title, content, author_id, community_id, create_time
   	from post
   	where post_id in (?)    // 将查询的post_id按照给定的post_id排序
   	order by FIND_IN_SET(post_id, ?) // 返回帖子在列表中的索引
   ```

5. 根据社区id查询社区详细信息，并过滤掉不属于该社区的帖子。`post.communityID != p.communityID`

6. 根据作者ID查询作者信息。

## 3、根据帖子id查询帖子详情

`v1.GET("/post/:id", controller.PostDetailHandler)`

- 首先将 id 解析为10进制的64位帧数。
- `post, err := logic.GetPostById(postId)`，查询帖子信息并组合程我们想要的接口。
  1. 根据 `postID` 查询帖子信息。
  2. 根据 `post.AuthorId` 查询作者信息。
  3. 根据 `post.CommunityID` 查询社区详细信息。
  4. 根据 `postID` 去 `redis` 查询帖子的投票数。
- 将上面信息拼接。

## 4、搜索帖子

`v1.GET("/search", controller.PostSearchHandler) `

```go
// ParamPostList 获取帖子列表query 参数
type ParamPostList struct {
	Search      string `json:"search" form:"search"`               // 关键字搜索
	CommunityID uint64 `json:"community_id" form:"community_id"`   // 可以为空
	Page        int64  `json:"page" form:"page"`                   // 页码
	Size        int64  `json:"size" form:"size"`                   // 每页数量
	Order       string `json:"order" form:"order" example:"score"` // 排序依据
}
```

1. 根据帖子标题或者帖子内容模糊查询帖子列表总数。

   ```go
   func GetPostListTotalCount(p *models.ParamPostList) (count int64, err error) {
   	// 根据帖子标题或者帖子内容模糊查询帖子列表总数
   	sqlStr := `select count(post_id)
   	from post
   	where title like ?
   	or content like ?   // 这里使用like模糊查询
   	`
   	// %keyword%
   	p.Search = "%" + p.Search + "%"  // %%为通配符匹配
   	err = db.Get(&count, sqlStr, p.Search, p.Search)
   	return
   }
   ```

2. 根据关键字再去mysql分页查询帖子列表。

   ```go
   // GetPostListByKeywords 根据关键词查询帖子列表
   func GetPostListByKeywords(p *models.ParamPostList) (posts []*models.Post, err error) {
   	// 根据帖子标题或者帖子内容模糊查询帖子列表
   	sqlStr := `select post_id, title, content, author_id, community_id, create_time
   	from post
   	where title like ?
   	or content like ?
   	ORDER BY create_time
   	DESC
   	limit ?,?
   	`
   	// %keyword%
   	p.Search = "%" + p.Search + "%"
   	posts = make([]*models.Post, 0, 2) // 0：长度  2：容量
   	err = db.Select(&posts, sqlStr, p.Search, p.Search, (p.Page-1)*p.Size, p.Size)
   	return
   }
   ```

3. 去redis查询投票数。

4. 查询作者信息及社区信息。

5. 数据拼接。

# 三、社区业务

## 1、获取分类社区列表

`v1.GET("/community", controller.CommunityHandler)`

- 查询到所有的社区(community_id,community_name)以列表的形式返回

  ```go
  sqlStr := "select community_id, community_name from community"
  err = db.Select(&communityList, sqlStr)
  ```

## 2、根据communityID查找社区详情

`v1.GET("/community/:id", controller.CommunityDetailHandler)`

```go
community := new(models.CommunityDetail)
sqlStr := `select community_id, community_name, introduction, create_time
	from community
	where community_id = ?`
err := db.Get(community, sqlStr, id)
```

# 四、权限业务

## 1、创建帖子

`v1.POST("/post", controller.CreatePostHandler) // 创建帖子`

- `c.ShouldBindJSON(&post)`，并根据当前JWT里面的userID，将post.AuthorID=userID。

- 生成postID，并将post保存进数据库。

  ```sql
  `insert into post(
  	post_id, title, content, author_id, community_id)
  	values(?,?,?,?,?)`
  ```

- 并根据`communityID`在`mysql`中获取`community`详细信息。

- 去`redis`使用`pipeline`存储 post 信息。`pipeline := client.TxPipeline() // 事务操作`

  1. `投票（zSet）。成员-分值对存储（pipeline.ZAdd()）。`

     ```go
     voteKey = bluebell-plus:post:voted:postID
     value = redis.Z{ // 作者默认投赞成票
     		Score:  1,
     		Member: userID,
     	}
     pipeline.Expire(votedKey, time.Second*OneMonthInSeconds*6) // 过期时间：6个月
     ```

  2. `post用Hash存储。pipeline.HMSet()`

     ```go
     key = bluebell-plus:post:postID
     value = postInfo
     postInfo := map[string]interface{}{
     		"title":    title,
     		"summary":  summary,
     		"post:id":  postID,
     		"user:id":  userID,
     		"time":     now,
     		"votes":    1,
     		"comments": 0,
     	}
     ```

  3. 添加score（ZSet）。pipeline.ZAdd()

     ```go
     key = bluebell-plus:post:score
     value = redis.Z{
     		Score:  now + VoteScore,  // 每投一票分数+432
     		Member: postID,
     	}
     ```

  4. 添加时间（ZSet）。

     ```go
     key = bluebell-plus:post:time
     value = redis.Z{
     		Score:  now,
     		Member: postID,
     	}
     ```

  5. `添加对应community（set）。pipeline.SAdd(communityKey, postID)`

     ```go
     communityKey = bluebell-plus:community:CommunityID
     value = postID
     ```

  6. `执行管道中的事务。pipeline.Exec()`

  这里创建帖子用的一致性策略是，先写MySQL再写Redis。

## 2、投票vote

`v1.POST("/vote", controller.VoteHandler) // 投票`

<img src="https://s2.loli.net/2024/08/20/l28Qc5tGbyxsJzK.png" alt="img" style="zoom: 50%;" />

- 获取当前的投票参数：post_id 和 direction，并得到当前的 userID

投票算法：

```go
// 投一票+432分 
// 因为 一天 = 24*60*60=86400，86400/200 = 432，代表投200票就可以给帖子再首页续一天

/* PostVote 为帖子投票
投票分为四种情况：1.投赞成票(1) 2.投反对票(-1) 3.取消投票(0) 4.反转投票

记录文章参与投票的人
更新文章分数：赞成票要加分；反对票减分

v=1时，有两种情况
	1.之前没投过票，现在要投赞成票		--> 更新分数和投票记录		差值的绝对值：1  +432
	2.之前投过反对票，现在要改为赞成票	--> 更新分数和投票记录		差值的绝对值：2  +432*2
v=0时，有两种情况
	1.之前投过反对票，现在要取消			--> 更新分数和投票记录		差值的绝对值：1  +432
	2.之前投过赞成票，现在要取消			--> 更新分数和投票记录		差值的绝对值：1  -432
v=-1时，有两种情况
	1.之前没投过票，现在要投反对票		--> 更新分数和投票记录		差值的绝对值：1  -432
	2.之前投过赞成票，现在要改为反对票	--> 更新分数和投票记录		差值的绝对值：2  -432*2

投票的限制：
每个帖子子发表之日起一个星期之内允许用户投票，超过一个星期就不允许投票了
	1、到期之后将redis中保存的赞成票数及反对票数存储到mysql表中，这里没有实现。
	2、到期之后删除那个 KeyPostVotedZSetPrefix，这里没有实现。
*/
```

1. 判断投票限制，如果超过一星期，则不能投票。

2. 计算要更新的分数。

   ```go
   // ZIncrBy 用于将有序集合中的成员分数增加指定数量
   _, err = pipeline.ZIncrBy(KeyPostScoreZSet, incrementScore, postID).Result() // 更新分数
   ```

3. 如果取消投票（操作为0），将当前的用户投票信息移除。

   ```go
   _, err = client.ZRem(key, userID).Result() // 从有序集合中移除指定的成员userID
    // bluebell-plus:post:voted:postID
   ```

4. 否则就记录投票。

5. 更新帖子投票数。

   ```go
   pipeline.HIncrBy(KeyPostInfoHashPrefix+postID, "votes", int64(op))
   // 对哈希（Hash）中指定字段("votes")的值进行递增或递减操作
   ```

## 3、评论及回复评论`comment`（亮点）

`v1.POST("/comment", controller.CommentHandler)`

- 获取回复评论内容content、user_id、post_id、parent_id，生成comment_id。

- 去mysql创建评论。

  ```go
  sqlStr := `insert into comment(
  	comment_id, content, post_id, author_id, parent_id)
  	values(?,?,?,?,?)`
  	_, err = db.Exec(sqlStr, comment.CommentID, comment.Content, comment.PostID,comment.AuthorID, comment.ParentID)
  ```

## 4、查看评论列表（亮点）

`v1.GET("/comment", controller.CommentListHandler) // 评论列表`

- 获取请求参数中的 post_id。

- 去数据库中查询当前帖子下的所有comment。

  ```go
  sqlStr := `select comment_id, content, post_id, author_id, parent_id, update_time
  	from comment
  	where post_id = ?`
  ```

- 创建 结构体如下，增加`SubComment`字段，存储子评论。

  ```go
  type CommentPlus struct {
  	Comment
  	SubComment []Comment `db:"sub_comment" json:"sub_comment"`
  }
  ```

- 在返回的帖子评论里面，首先查找parent_id==0 的评论，表示这是一个父评论；再遍历所有的父评论，并在所有的评论里面查找parent_id与等于父评论的comment_id的评论，将其加入到父评论的子评论。代码见`logic.GetCommentListByPostID(postID)`

  


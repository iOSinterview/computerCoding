# Go安装ETCD-clientv3运行时报错问题的解释

---

再go中使用etcd时，使用`go get go.etcd.io/etcd/clientv3`进行了安装，但是再项目运行时出现了如下错误：

```go
 D:\GoPath\src\etcdDemo> go run .\main.go
# github.com/coreos/etcd/clientv3/balancer/picker
..\..\pkg\mod\github.com\coreos\etcd@v3.3.27+incompatible\clientv3\balancer\picker\err.go:37:53: undefined: balancer.PickOptions
..\..\pkg\mod\github.com\coreos\etcd@v3.3.27+incompatible\clientv3\balancer\picker\roundrobin_balanced.go:55:63: undefined: balancer.PickOptions
# github.com/coreos/etcd/clientv3/balancer/resolver/endpoint
..\..\pkg\mod\github.com\coreos\etcd@v3.3.27+incompatible\clientv3\balancer\resolver\endpoint\endpoint.go:114:87: undefined: resolver.BuildOption
..\..\pkg\mod\github.com\coreos\etcd@v3.3.27+incompatible\clientv3\balancer\resolver\endpoint\endpoint.go:115:16: target.Authority undefined (type resolver.Target has no field or method Authority)
..\..\pkg\mod\github.com\coreos\etcd@v3.3.27+incompatible\clientv3\balancer\resolver\endpoint\endpoint.go:118:15: target.Authority undefined (type resolver.Target has no field or method Authority)
..\..\pkg\mod\github.com\coreos\etcd@v3.3.27+incompatible\clientv3\balancer\resolver\endpoint\endpoint.go:182:40: undefined: resolver.ResolveNowOption   
```

大概是说原因是 google.golang.org/grpc 1.26 后的版本是不支持 clientv3 的。

也就是说要把这个改成 1.26 版本的就可以了。

第一种方式：
具体操作方法是在 go.mod 里加上：

```go
replace google.golang.org/grpc => google.golang.org/grpc v1.26.0
```

![etcd套路（七）安装clientv3报错问题的解释](https://s2.loli.net/2024/08/30/jCJbmWPLMAxuNhG.png)

但是我试了，没用！

```go
 go run .\main.go
# github.com/coreos/etcd/clientv3/balancer/resolver/endpoint
..\..\pkg\mod\github.com\coreos\etcd@v3.3.27+incompatible\clientv3\balancer\resolver\endpoint\endpoint.go:114:87: undefined: resolver.BuildOption
..\..\pkg\mod\github.com\coreos\etcd@v3.3.27+incompatible\clientv3\balancer\resolver\endpoint\endpoint.go:115:16: target.Authority undefined (type resolver.Target has no field or method Authority)
..\..\pkg\mod\github.com\coreos\etcd@v3.3.27+incompatible\clientv3\balancer\resolver\endpoint\endpoint.go:118:15: target.Authority undefined (type resolver.Target has no field or method Authority)
..\..\pkg\mod\github.com\coreos\etcd@v3.3.27+incompatible\clientv3\balancer\resolver\endpoint\endpoint.go:182:40: undefined: resolver.ResolveNowOption       
# github.com/coreos/etcd/clientv3/balancer/picker
..\..\pkg\mod\github.com\coreos\etcd@v3.3.27+incompatible\clientv3\balancer\picker\err.go:37:53: undefined: balancer.PickOptions
..\..\pkg\mod\github.com\coreos\etcd@v3.3.27+incompatible\clientv3\balancer\picker\roundrobin_balanced.go:55:63: undefined: balancer.PickOptions
```

其实主要还是不兼容的问题，我用的grpc时v1.60.0，clientv3只支持1.26.0，因此，先删除掉已经生成的go.mod和go.sum，依次执行如下

```go
go mod init 
go mod edit -replace github.com/coreos/bbolt@v1.3.4=go.etcd.io/bbolt@v1.3.4
go mod edit -replace google.golang.org/grpc@v1.66.0=google.golang.org/grpc@v1.26.0
go mod tidy
go run main.go
```

这里连接成功，但是依然有错误，应该是我业务代码的问题：

```go
go run .\main.go
connect to etcd success
{"level":"warn","ts":"2024-08-30T23:14:21.520+0800","caller":"clientv3/retry_interceptor.go:62","msg":"retrying of unary invoker failed","target":"endpoint://client-1fcae15f-6ac7-45d4-a9b2-77042e04452e/127.0.0.1:2379","attempt":0,"error":"rpc error: code = DeadlineExceeded desc = latest balancer error: all SubConns are in TransientFailure, latest connection error: connection error: desc = \"transport: Error while dialing dial tcp 127.0.0.1:2379: connectex: No connection could be made because the target machine actively refused it.\""}
put to etcd failed, err:context deadline exceeded
```


package main

import (
	"flag"
	"fmt"
	"log"

	"gaocache"
)

// 模仿数据库
var (
	db = map[string]string{
		"Tom":   "630",
		"Jack":  "589",
		"Sam":   "567",
		"Alice": "666",
	}
)

func createGroup(groupName string) *gaocache.Group {
	return gaocache.NewGroup(groupName, 2<<10, gaocache.RetrieverFunc(
		func(key string) ([]byte, error) {
			log.Println("[MySQL DB] search key", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))
}

func startCacheServer(addr string, gao *gaocache.Group) {
	svr, err := gaocache.NewServer(addr)
	if err != nil {
		log.Println("NewServer failed at", addr)
		return
	}

	gao.RegisterSvr(svr)
	log.Println("gaocache is running at", addr)

	// 启动服务(注册服务至etcd/计算一致性哈希...)

	// Start将不会return 除非服务stop或者抛出error
	err = svr.Start()
	if err != nil {
		log.Fatal(err)
	}
}

// func startClient(apiAddr string, groupName string) {
// 	cli := gaocache.NewClient(apiAddr)
// 	keys := []string{"Tom", "Alice", "Sam", "NotExist"}
// 	// 第一次查询不在缓存中，回调函数会从数据库中查
// 	for i := range keys {
// 		value, err := cli.Fetch(groupName, keys[i])
// 		if err != nil {
// 			fmt.Printf("err:%v\n", err)

// 		}
// 		fmt.Printf("key:%v,value:%v\n", keys[i], string(value))
// 	}
// 	// 第二次查询不在缓存中，回调函数会从数据库中查
// 	for i := range keys {
// 		value, err := cli.Fetch(groupName, keys[i])
// 		if err != nil {
// 			fmt.Printf("err:%v\n", err)

// 		}
// 		fmt.Printf("key:%v,value:%v\n", keys[i], string(value))
// 	}
// }

func main() {
	// 解析终端命令
	var port string
	var api bool
	var name string
	var host string
	flag.StringVar(&name, "name", "scores", "gaocache group name")
	flag.StringVar(&port, "port", "8001", "gaocache server port")
	flag.BoolVar(&api, "api", false, "Start a api server?")
	flag.StringVar(&host, "host", "localhost", "gaocache node server host's IP?")
	flag.Parse()
	// 节点服务端接口 => x.x.x.x:port

	// 创建一个 Group为name的缓存空间
	gao := createGroup(name)
	// if api {
	// 	go startClient(addrMap[port], name)
	// }
	// 启动服务
	addr := host + ":" + port
	startCacheServer(addr, gao)
}

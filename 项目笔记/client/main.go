package main

import (
	"flag"
	"fmt"

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
	keys  = []string{"Tom", "Sam", "Alice", "NotExist"}
	keys2 = []string{"Tom"}

	// wg = sync.WaitGroup{}
)

func StartClient(serverName string, groupName string) {
	// defer wg.Done()
	cli := gaocache.NewClient(serverName)
	// 第一次查询不在缓存中，回调函数会从数据库中查
	for _, v := range keys {
		value, err := cli.Fetch(groupName, v)
		if err != nil {
			fmt.Printf("search error:%v", err)
		}
		fmt.Printf("key:%v,value:%v\n", v, string(value))
	}
}

func main() {

	// 解析终端命令
	// var port int
	// var api bool
	var groupName string
	var serverName string
	flag.StringVar(&groupName, "groupName", "scores", "gaocache group name")
	flag.StringVar(&serverName, "serverName", "gaocache", "gaocache server name")

	flag.Parse()

	// 节点服务端接口 => x.x.x.x:port
	// addrMap := map[int]string{
	// 	8001: "localhost:8001",
	// 	8002: "localhost:8002",
	// 	8003: "localhost:8003",
	// 	8004: "localhost:8004",
	// 	9999: "localhost:9999",
	// }
	// cnt := 4
	// wg.Add(cnt)
	// for i := 0; i < cnt; i++ {
	// 	go StartClient(serverName, groupName)
	// }
	// wg.Wait()
	StartClient(serverName, groupName)
	// time.Sleep(2 * time.Second)
}

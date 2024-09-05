package gaocache

import "fmt"

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
	cli := NewClient(serverName)
	// 第一次查询不在缓存中，回调函数会从数据库中查
	for _, v := range keys {
		value, err := cli.Fetch(groupName, v)
		if err != nil {
			fmt.Printf("search error:%v", err)
		}
		fmt.Printf("key:%v,value:%v\n", v, string(value))
	}
}

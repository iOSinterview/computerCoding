package gaocache

import (
	"fmt"
	"gaocache/gaocachepb"
	"log"
	"net"
	"testing"

	"google.golang.org/grpc"
)

func createTestSvr() (*Group, *server) {
	mysql := map[string]string{
		"Tom":  "630",
		"Jack": "589",
		"Sam":  "567",
	}

	g := NewGroup("scores", 2<<10, RetrieverFunc(
		func(key string) ([]byte, error) {
			log.Println("[Mysql] search key", key)
			if v, ok := mysql[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))

	v, err := g.Get("Tom")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("val:%v", v)
	// 随机一个端口 避免冲突
	port := 8080
	addr := fmt.Sprintf("localhost:%d", port)

	svr, err := NewServer(addr)
	if err != nil {
		fmt.Printf("err:%v", err)
	}
	return g, svr

}

func TestServer_GetExistsKey(t *testing.T) {
	// 监听本地 8080 端口
	lis, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Printf("failed to listen: %v", err)
		return
	}
	s := grpc.NewServer() // 创建一个新的 gRPC 服务器实例
	_, svr := createTestSvr()
	gaocachepb.RegisterGaocacheServer(s, svr) // 注册服务至grpc
	err = s.Serve(lis)                        // 启动服务
	if err != nil {
		t.Fatal(err)
	}

}

func TestClient_GetExistsKey(t *testing.T) {

	svrname := "127.0.0.0:8080"
	cli := NewClient("", svrname)
	group := "scores"
	keys := []string{"Tom", "Alice"}
	val, err := cli.Fetch(group, keys[0])
	if err != nil {
		fmt.Printf("err:%v\n", err)
	}
	fmt.Println(val)
}

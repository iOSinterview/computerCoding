package gaocache

import (
	"context"
	"fmt"
	"gaocache/consistenthash"
	"gaocache/etcdserver"
	pb "gaocache/gaocachepb"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
)

const (
	defaultAddr     = "127.0.0.1:8001"
	defaultReplicas = 50
)

var (
	ttl               = 5 * time.Second // 租约过期时间
	defaultEtcdConfig = clientv3.Config{
		Endpoints:   []string{"localhost:2379"},
		DialTimeout: ttl,
	}
	ServiceName = "gaocache"
)

// server 和 Group 是解耦合的 所以server要自己实现并发控制
type server struct {
	pb.UnimplementedGaocacheServer
	addr       string // ip:port
	mu         sync.Mutex
	consHash   *consistenthash.Consistency // 一致性哈希
	clients    map[string]*client          // 每个remotePeer 对应一个client
	status     bool                        // 服务运行状态 true:runing false:stop
	stopSignal chan os.Signal              // 监听服务运行状态，如果宕机则通知etcd撤销服务
}

// NewServer 创建cache的svr 若addr为空 则使用defaultAddr
func NewServer(addr string) (*server, error) {
	if addr == "" {
		addr = defaultAddr
	}
	// 判断是否满足 x.x.x.x:port 的格式
	if !validPeerAddr(addr) {
		return nil, fmt.Errorf("invalid addr %s, it should be x.x.x.x:port", addr)
	}
	return &server{addr: addr}, nil
}

// Get 实现gaocache service的Get接口
func (s *server) Get(ctx context.Context, in *pb.GetRequest) (*pb.GetResponse, error) {
	group, key := in.GetGroup(), in.GetKey()
	resp := &pb.GetResponse{}

	log.Printf("[gaocache_svr %s] Recv RPC Request - (%s)/(%s)", s.addr, group, key)
	if key == "" {
		return resp, fmt.Errorf("key required")
	}
	g := GetGroup(group)
	if g == nil {
		return resp, fmt.Errorf("group not found")
	}
	view, err := g.Get(key)
	if err != nil {
		return resp, err
	}
	resp.Value = view.ByteSlice()
	return resp, nil
}

// SetPeers 实例化了一致性哈希算法，将各个远端主机IP配置到server里
// 这样Server就可以Pick他们了
// 注意: 此操作是*覆写*操作！
// 注意: peersIP必须满足 x.x.x.x:port的格式
func (s *server) SetPeers(peersAddr ...string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.consHash = consistenthash.New(defaultReplicas, nil)
	s.consHash.Register(peersAddr...)
	s.clients = make(map[string]*client)
	for _, peerAddr := range peersAddr {
		if !validPeerAddr(peerAddr) {
			panic(fmt.Sprintf("[peer %s] invalid address format, it should be x.x.x.x:port", peerAddr))
		}
		// 服务名：ip:port
		serviceAddr := peerAddr
		// peerAddr : client
		s.clients[peerAddr] = NewClient("", serviceAddr)
		// fmt.Printf("peerAddr:%v  client:%v \n", peerAddr, service)
	}
}

// Pick 根据一致性哈希选举出key应存放在的cache
// return false 代表从本地获取cache
func (s *server) PickPeer(key string) (Fetcher, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	peerAddr := s.consHash.GetPeer(key)
	// Pick itself
	if peerAddr == s.addr {
		log.Printf("ooh! pick myself, I am %s\n", s.addr)
		return nil, false
	}
	log.Printf("[node cache %s] pick remote peer: %s key:[%s]\n", s.addr, peerAddr, key)
	return s.clients[peerAddr], true
}

// Start 启动cache服务
func (s *server) Start() error {
	s.mu.Lock() // 当前server（ip:port）只允许启动一个服务
	if s.status {
		s.mu.Unlock() // 启动完成
		return fmt.Errorf("[server %s] already started", s.addr)
	}

	// -----------------启动服务----------------------
	// 1. 初始化tcp socket并监听服务
	// 2. 注册rpc服务至grpc，这样grpc收到request可以分发给server处理
	// 5. 将自己的服务名/Host地址注册至etcd 这样client可以通过etcd
	//    获取服务Host地址 从而进行通信。这样的好处是client只需知道服务名
	//    以及etcd的Host即可获取对应服务IP 无需写死至client代码中
	// ----------------------------------------------

	// 1、监听服务
	port := strings.Split(s.addr, ":")[1]
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return fmt.Errorf("server failed to listen:%v\n err=%v", port, err)
	}
	// 2、注册grpcserver
	grpcServer := grpc.NewServer()
	pb.RegisterGaocacheServer(grpcServer, s)
	s.status = true

	// 3、将服务注册至etcd
	// 服务名：gaocache/ip:port
	err = etcdserver.EtcdRegister(ServiceName, s.addr, s.stopSignal)
	if err != nil {
		return fmt.Errorf("etcd server rigist failed:%v", err)
	}
	log.Printf("[%s] register service ok\n", s.addr)
	s.mu.Unlock()

	// 4、关闭信号处理
	s.stopSignal = make(chan os.Signal, 1)
	signal.Notify(s.stopSignal, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP, syscall.SIGQUIT)
	go func() {
		sig := <-s.stopSignal
		etcdserver.UnRegister(ServiceName, s.addr)
		if i, ok := sig.(syscall.Signal); ok {
			os.Exit(int(i))
		} else {
			os.Exit(0)
		}
	}()

	// 5、监听etcd中前缀为ServiceName的键值变化，更新哈希表
	// 服务端必须更新自己的哈希表，因为如果客户端获取到了最新的服务列表，
	// 而服务端的哈希表还是原来的旧的，很可能pick到已经停掉的服务
	// 而且哈希表的更新应该是个监控的过程，有新服务或者服务下线都应该更新

	// 初始化 etcd 客户端
	cli, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"localhost:2379"},
	})
	if err != nil {
		log.Fatal(err)
	}
	defer cli.Close()

	// 创建 Watcher
	watchChan := cli.Watch(context.Background(), ServiceName, clientv3.WithPrefix())

	// 获取服务名下的所有 key-value 对
	ServerListAddrs := make([]string, 0)
	getAllKeys := func() {
		resp, err := cli.Get(context.Background(), ServiceName, clientv3.WithPrefix())
		if err != nil {
			log.Printf("Failed to get all keys: %v", err)
			return
		}

		for _, kv := range resp.Kvs {
			ServerListAddrs = append(ServerListAddrs, string(kv.Value))
		}
	}
	// ticker := time.NewTicker(2 * time.Second)
	// 监听服务名的变化
	go func() {
		for _ = range watchChan {
			// 重新获取服务名下的所有 key-value 对
			getAllKeys()
			// 处理事件
			// for _, event := range resp.Events {
			// 	fmt.Printf("Event received! Type: %v\n", event.Type)
			// }
			s.SetPeers(ServerListAddrs...)
			// <-ticker.C
		}
	}()

	// ticker := time.NewTicker(2 * time.Second) // 2s获取一次
	// go func() {
	// 	for {
	// 		ServerListAddrs, err = etcdserver.EtcdGetServerListAddrs(ServiceName)
	// 		if err != nil {
	// 			log.Fatalf("[server %s] etcdGetServerListAddrs failed...", s.addr)
	// 			return
	// 		}
	// 		// fmt.Printf("get serverlistaddrs:%v\n", ServerListAddrs)
	// 		s.SetPeers(ServerListAddrs...)
	// 		<-ticker.C
	// 	}
	// }()

	// 额外需求，10s在服务端输出一次serverlistaddrs，打印太频繁了显得臃肿
	// ticker2 := time.NewTicker(10 * time.Second)
	// go func() {
	// 	for {
	// 		fmt.Printf("get serverlistaddrs:%v\n", ServerListAddrs)
	// 		<-ticker2.C
	// 	}
	// }()

	// 6、监听并处理请求
	if err := grpcServer.Serve(lis); s.status && err != nil {
		return fmt.Errorf("failed to serve: %v", err)
	}
	return nil
}

// Stop 停止server运行 如果server没有运行 这将是一个no-op
func (s *server) Stop() {
	s.mu.Lock()
	if !s.status {
		s.mu.Unlock()
		return
	}
	s.stopSignal <- nil // 发送停止keepalive信号
	s.status = false    // 设置server运行状态为stop
	s.clients = nil     // 清空一致性哈希信息 有助于垃圾回收
	s.consHash = nil
	s.mu.Unlock()
}

// 测试Server是否实现了Picker接口
var _ Picker = (*server)(nil)

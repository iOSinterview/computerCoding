package etcdserver

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

// 提供服务Service注册至etcd的能力

var (
	ttl               = 5 * time.Second
	defaultEtcdConfig = clientv3.Config{
		Endpoints:   []string{"localhost:2379"},
		DialTimeout: ttl,
	}
)

func EtcdRegister(serviceName string, serverAddr string, stopSignal chan os.Signal) error {

	// 1、创建一个etcd client
	cli, err := clientv3.New(defaultEtcdConfig)
	if err != nil {
		return fmt.Errorf("create etcd client failed: %v", err)
	}
	defer cli.Close()

	// 2、向etcd使用get服务地址，存在表示已经注册，返回
	// 与etcd建立长连接，每5s计时执行一次，并保证连接不断(心跳检测)
	ticker := time.NewTicker(time.Second * time.Duration(ttl))

	serverkey := serviceName + "/" + serverAddr
	for {
		resp, err := cli.Get(context.Background(), serverkey)
		//fmt.Printf("resp:%+v\n", resp)
		if err != nil {
			return fmt.Errorf("serverAddr get faild:%v", err)
		} else if resp.Count == 0 { //尚未注册
			err = keepAlive(cli, serviceName, serverAddr)
			if err != nil {
				return fmt.Errorf("keep long connect failed:%s", err)
			}
			return nil
		}
		<-ticker.C
	}
}

func keepAlive(cli *clientv3.Client, serviceName string, serverAddr string) error {
	// 1、创建租约
	lease, err := cli.Grant(cli.Ctx(), int64(ttl))
	if err != nil {
		return fmt.Errorf("create lease failed: %v", err)
	}

	// 2、将服务地址注册etcd
	serverkey := serviceName + "/" + serverAddr
	_, err = cli.Put(cli.Ctx(), serverkey, serverAddr, clientv3.WithLease(lease.ID))
	if err != nil {
		return fmt.Errorf("register etcd record failed: %v", err)
	}
	log.Printf("[%s] register service ok\n", serverAddr)

	// 3、心跳检测
	ch, err := cli.KeepAlive(context.Background(), lease.ID)
	if err != nil {
		return fmt.Errorf("keepAlive failed:%v", err)
	}

	//清空keepAlive返回的channel
	go func() {
		for {
			<-ch
		}
	}()
	return nil
}

func UnRegister(serviceName string, serverAddr string) error {
	// 1、创建一个etcd client
	cli, err := clientv3.New(defaultEtcdConfig)
	if err != nil {
		return fmt.Errorf("unregister create etcd client failed: %v", err)
	}
	defer cli.Close()

	if cli != nil {
		serverkey := serviceName + "/" + serverAddr
		cli.Delete(context.Background(), serverkey)
	}
	return nil
}

package consistenthash

// consistenthash 模块负责实现一致性哈希
// 用于确定key与peer之间的映射,key选择哪个节点 peer

import (
	"hash/crc32"
	"sort"
	"strconv"
)

// HashFunc 定义哈希函数输入输出
type HashFunc func(data []byte) uint32

// Consistency 维护peer与其hash值的关联
type Consistency struct {
	hashFunc HashFunc       // 哈希函数依赖
	replicas int            // 虚拟节点个数(防止数据倾斜)
	hashRing []int          // uint32哈希环
	peerMap  map[int]string // hashValue -> peerName，虚拟节点：真实节点
}

func New(replicas int, fn HashFunc) *Consistency {
	c := &Consistency{
		replicas: replicas,
		hashFunc: fn,
		peerMap:  make(map[int]string),
	}
	if c.hashFunc == nil {
		c.hashFunc = crc32.ChecksumIEEE
	}
	return c
}

// Register 将各个peer注册到哈希环上
func (c *Consistency) Register(peersName ...string) {
	for _, peerName := range peersName {
		for i := 0; i < c.replicas; i++ {
			hashValue := int(c.hashFunc([]byte(strconv.Itoa(i) + peerName)))
			c.hashRing = append(c.hashRing, hashValue)
			c.peerMap[hashValue] = peerName
		}
	}
	sort.Ints(c.hashRing)
}

// GetPeer 计算key应缓存到的peer
func (c *Consistency) GetPeer(key string) string {
	// if len(c.hashRing) == 0 {
	// 	return ""
	// }
	hashValue := int(c.hashFunc([]byte(key)))
	idx := sort.Search(len(c.hashRing), func(i int) bool {
		return c.hashRing[i] >= hashValue
	})
	return c.peerMap[c.hashRing[idx%len(c.hashRing)]]
}

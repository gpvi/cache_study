package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

// 哈希映射 到 2^32 -1
type Hash func(data []byte) uint32

type Map struct {
	hash Hash
	// 一个节点对应几个虚拟节点
	replicas int
	// 存储 Hash映射后的值，哈希环
	keys []int
	//虚拟节点 与实际节点的映射
	hashMap map[int]string
}

// 返回 Map ,传入 虚拟节点与实际节点的比值，传入自定义Hash函数
func New(replicas int, fn Hash) *Map {
	m := &Map{
		replicas: replicas,
		hash:     fn,
		hashMap:  make(map[int]string),
	}

	if m.hash == nil {
		// 映射规则初始化
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

func (m *Map) Add(keys ...string) {
	for _, key := range keys {
		for i := 0; i < m.replicas; i++ {
			// 计算虚拟节点的哈希值，数字加key 转字节码后哈希映射
			hash := int(m.hash([]byte(strconv.Itoa(i) + key)))
			m.keys = append(m.keys, hash)
			m.hashMap[hash] = key
		}
	}
	sort.Ints(m.keys)
}

// 寻找最近的节点(实际节点)
func (m Map) Get(key string) string {
	if len(m.keys) == 0 {
		return ""
	}
	hash := int(m.hash([]byte(key)))

	index := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= hash
	})

	return m.hashMap[m.keys[index%len(m.keys)]]

}

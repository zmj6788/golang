package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

type Hash func(data []byte) uint32

type Map struct {
	//hash算法
	hash Hash
	// 虚拟节点倍数
	replicas int
	// 匹配环
	keys []int
	// 虚拟节点hash值与真实节点的映射
	hashMap map[int]string
}

func New(replicas int, fn Hash) *Map {
	m := &Map{
		replicas: replicas,
		hash:     fn,
		hashMap:  make(map[int]string),
	}
	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

// 添加虚拟节点进入匹配环，增加虚拟节点hash值与真实节点映射
func (m *Map) Add(keys ...string) {
	for _, key := range keys {
		for i := 0; i < m.replicas; i++ {
			// 这个函数是通过将key和一个从0开始递增的整数i拼接成一个新的字符串，
			// 然后对该字符串进行哈希计算，最终得到一个整数hash的值。
			hash := int(m.hash([]byte(strconv.Itoa(i) + key)))
			m.keys = append(m.keys, hash)
			m.hashMap[hash] = key
		}
		// 升序排序
		sort.Ints(m.keys)
	}
}
func (m *Map) Get(key string) string {
	if len(m.keys) == 0 {
		return ""
	}
	hash := int(m.hash([]byte(key)))
	// 在匹配环中寻找第一个大于等于key的hash值的节点
	// 返回序号
	// 特殊情况：环上没有一个大于等于key的hash值的节点，
	// 返回idx=len(m.keys)
	idx := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= hash
	})
	// 根据序号获取虚拟节点hash值,再根据此值获取真实节点
	//最后返回真实节点
	//特殊情况：环上没有一个大于等于key的hash值的节点，
	//则取第一个节点
	return m.hashMap[m.keys[idx%len(m.keys)]]
}

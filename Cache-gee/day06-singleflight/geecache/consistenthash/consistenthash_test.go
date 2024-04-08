package consistenthash

import (
	"strconv"
	"testing"
)

func TestHashing(t *testing.T) {
	// 自定义Hash算法，将key转换为int类型返回
	hash := New(3, func(key []byte) uint32 {
		i, _ := strconv.Atoi(string(key))
		return uint32(i)
	})
	// hash := int(m.hash([]byte(strconv.Itoa(i) + key)))
	// 虚拟节点分布
	// 2, 4, 6, 12, 14, 16, 22, 24, 26
	hash.Add("6", "4", "2")
	// key期望匹配的节点
	testCases := map[string]string{
		"2":  "2",
		"11": "2",
		"23": "4",
		"27": "2",
	}
	// 期望值与真实值比较
	for key, value := range testCases {
		if hash.Get(key) != value {
			t.Errorf("Asking for %s, expected %s, got %s", key, value, hash.Get(key))
		}
	}
	// 增加节点后
	hash.Add("8")
	// 27期望匹配的节点变化
	testCases["27"] = "8"
	// 期望值与真实值再比较
	for key, value := range testCases {
		if hash.Get(key) != value {
			t.Errorf("Asking for %s, expected %s, got %s", key, value, hash.Get(key))
		}
	}
}

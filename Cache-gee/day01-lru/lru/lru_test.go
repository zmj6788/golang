package lru

import (
	"reflect"
	"testing"
)

type String string

func (d String) Len() int {
	return len(d)
}

// 测试命令 go test
// 测试命令 go test -v
func TestGet(t *testing.T) {
	//测试逻辑
	//目的：测试Get方法
	//测试步骤：
	//1. 添加key1，value为1234
	//2. 获取key1，判断key1存在且value为1234
	//3. 获取key2，判断key2不存在
	//若测试通过，不返回值
	lru := New(int64(10), nil)
	lru.Add("key1", String("1234"))
	if v, ok := lru.Get("key1"); !ok || string(v.(String)) != "1234" {
		t.Fatalf("cache hit key1=1234 failed")
	}
	if _, ok := lru.Get("key2"); !ok {
		t.Fatalf("cache miss key2 failed")
	}
}

func TestRemoveoldest(t *testing.T) {
	// 测试逻辑
	// 测试步骤：
	// 1. 声明3个key1，key2，key3，key1的value为value1，key2的value为value2，key3的value为value3
	// 2. 设置缓存容量只能缓存key1和key2
	// 3. 添加key1和key2，此时key1和key2都存在
	// 3. 添加key3，key3的value为value3，此时key1由于缓存空间不足，key1被删除
	// 4. 获取key1，判断key1不存在
	// 若测试通过，不返回值
	k1, k2, k3 := "key1", "key2", "key3"
	v1, v2, v3 := "value1", "value2", "value3"
	cap := len(k1 + k2 + v1 + v2)
	lru := New(int64(cap), nil)
	lru.Add(k1, String(v1))
	lru.Add(k2, String(v2))
	lru.Add(k3, String(v3))

	if _, ok := lru.Get("key1"); ok || lru.Len() != 2 {
		t.Fatalf("Removeoldest key1 failed")
	}
}

func TestOnEvicted(t *testing.T) {
	// 测试逻辑
	// 测试步骤：
	// 1. 声明一个回调函数，用于存储被删除的key
	// 2. 设置缓存容量只能缓存key1和k2
	// 3. 添加key1，key1的value为123456
	// 4. 添加k2，k2的value为k2
	// 5. 添加k3，k3的value为k3
	// 6. 添加k4，k4的value为k4
	// 7. 由于缓存空间不足，key1和k2应被删除
	// 8. keys 应存储被删除的key1和k2
	// 9. 设置keys的期望值expect = ["key1", "k2"]
	// 10. 比较keys和expect，若不相等，测试失败，返回语句
	// 若测试通过，不返回值
	keys := make([]string, 0)
	callback := func(key string, value Value) {
		// 存储被删除的key
		keys = append(keys, key)
	}
	lru := New(int64(10), callback)
	lru.Add("key1", String("123456"))
	lru.Add("k2", String("k2"))
	lru.Add("k3", String("k3"))
	lru.Add("k4", String("k4"))

	expect := []string{"key1", "k2"}

	if !reflect.DeepEqual(expect, keys) {
		t.Fatalf("Call OnEvicted failed, expect keys equals to %s", expect)
	}
}

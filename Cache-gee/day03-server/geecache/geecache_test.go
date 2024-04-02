package geecache

import (
	"fmt"
	"log"
	"reflect"
	"testing"
)

var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

func TestGetter(t *testing.T) {
	var f Getter = GetterFunc(func(key string) ([]byte, error) {
		return []byte(key), nil
	})
	expect := []byte("key")
	if v, _ := f.Get("key"); !reflect.DeepEqual(v, expect) {
		t.Errorf("callback failed")
	}
}

// 测试Get函数
// 测试步骤

//  1. 创建了一个loadCounts的map，用于记录每个key的加载次数。
//  2. 使用NewGroup函数创建了一个名为scores的组，
//     并传入了一个自定义的GetterFunc，该函数用于从db中获取数据。
//     如果db中存在该key，则将对应的value转换为[]byte类型并返回；
//     如果不存在，则返回错误信息。
//  3. 缓存命中，返回缓存值，并将缓存值刷新到缓存中
func TestGet(t *testing.T) {
	loadCounts := make(map[string]int, len(db))
	gee := NewGroup("scores", 2<<10, GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := db[key]; ok {
				if _, ok := loadCounts[key]; !ok {
					// 第一次从db中获取key对应数据
					// 初始化key的获取次数
					loadCounts[key] = 0
				}
				// 获取成功次数加一
				// 但是最多也就是一了
				// 因为获取一次后就会被添加到缓存中了
				loadCounts[key] += 1
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))

	for k, v := range db {
		// 测试缓存值是否正确
		if view, err := gee.Get(k); err != nil || view.String() != v {
			t.Fatal("failed to get value of ", k)
		}
		// Get方法会将缓存值从db中重新添加到缓存中，所以key的加载次数应该为1
		if _, err := gee.Get(k); err != nil || loadCounts[k] > 1 {
			t.Fatalf("cache %s miss", k)
		}
	}

	if view, err := gee.Get("unknown"); err == nil {
		t.Fatalf("the value of unknow should be empty, but %s got", view)
	}

}

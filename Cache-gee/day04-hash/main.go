package main

import (
	"fmt"
	"log"
	"net/http"
	"project/Cache-gee/day04-hash/geecache"
)

var db = map[string]string{
	"Tom":  "666",
	"Jack": "888",
	"Sam":  "6788",
}

func main() {
	// 创建缓存组
	geecache.NewGroup("scores", 2<<10, geecache.GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))
	// 创建HTTPPool节点
	addr := "localhost:9999"
	peers := geecache.NewHTTPPool(addr)
	log.Println("geecache is running at", addr)
	// 启动HTTPPool节点服务
	log.Fatal(http.ListenAndServe(addr, peers))
}

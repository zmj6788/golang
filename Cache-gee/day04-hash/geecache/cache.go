package geecache

import (
	"project/Cache-gee/day04-hash/geecache/lru"
	"sync"
)

// 嵌套Cache结构体,对其方法进行并发支持
type cache struct {
	mu  sync.Mutex
	lru *lru.Cache
	// 记录缓存的最大字节数
	cacheBytes int64
}

// 对底层Add方法进行并发支持
func (c *cache) add(key string, value ByteView) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil {
		c.lru = lru.New(c.cacheBytes, nil)
	}
	c.lru.Add(key, value)
}

// 对底层Get方法进行并发支持
func (c *cache) get(key string) (value ByteView, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil {
		return
	}
	if v, ok := c.lru.Get(key); ok {
		return v.(ByteView), true
	}
	return
}

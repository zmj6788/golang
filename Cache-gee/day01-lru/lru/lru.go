package lru

import (
	"container/list"
)

type Cache struct {
	maxBytes int64
	nbytes   int64
	//双向链表内部是存储了键值对的结构体entry
	ll *list.List
	//缓存存储了键和指向双向链表中存储了键值对的结构体entry的指针
	cache map[string]*list.Element
	//被用来作为缓存项被移除时的处理操作
	//用户可以根据需要在该函数中实现自定义的处理逻辑，
	//比如将被移除的缓存项记录到日志中、从数据库中删除等操作。
	OnEvicted func(key string, value Value)
}

type entry struct {
	key   string
	value Value
}

type Value interface {
	Len() int
}

func New(maxBytes int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes: maxBytes,
		//返回一个空的双向链表
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

func (c *Cache) Get(key string) (value Value, ok bool) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		//断言用于将一个接口类型的值转换为特定类型
		//将ele.Value断言为*entry类型，并将其赋值给kv变量
		// 这意味着ele.Value必须是*entry类型，
		// 否则断言将会失败并引发panic。
		kv := ele.Value.(*entry)
		return kv.value, true
	}
	return
}

func (c *Cache) RemoveOldest() {
	ele := c.ll.Back()
	if ele != nil {
		c.ll.Remove(ele)
		kv := ele.Value.(*entry)
		delete(c.cache, kv.key)
		//更新当前缓存
		c.nbytes -= int64(len(kv.key)) + int64(kv.value.Len())
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)
		}
	}
}

func (c *Cache) Add(key string, value Value) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		c.nbytes += int64(value.Len()) - int64(kv.value.Len())
		kv.value = value
	} else {
		ele := c.ll.PushFront(&entry{key, value})
		c.cache[key] = ele
		c.nbytes += int64(len(key) + value.Len())
	}
	for c.maxBytes != 0 && c.maxBytes < c.nbytes {
		c.RemoveOldest()
	}
}

func (c *Cache) Len() int {
	return c.ll.Len()
}

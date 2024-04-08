# 动手写分布式缓存 - GeeCache第六天 防止缓存击穿

## 缓存雪崩、缓存击穿与缓存穿透的概念简介。

## 使用 singleflight 防止缓存击穿，实现与测试。

## 1.缓存雪崩、缓存击穿与缓存穿透

> **缓存雪崩**：缓存在同一时刻全部失效，造成瞬时DB请求量大、压力骤增，引起雪崩。缓存雪崩通常因为缓存服务器宕机、缓存的 key 设置了相同的过期时间等引起。

> **缓存击穿**：一个存在的key，在缓存过期的一刻，同时有大量的请求，这些请求都会击穿到 DB ，造成瞬时DB请求量大、压力骤增。

> **缓存穿透**：查询一个不存在的数据，因为不存在则不会写到缓存中，所以每次都会去请求 DB，如果瞬间流量过大，穿透到 DB，导致宕机。

## 2.为什么要防止缓存击穿？

day05测试的时候，我们并发了 3 个请求 `?key=Tom`，从日志中可以看到，三次均选择了节点 `8001`，这是一致性哈希算法的功劳。但是有一个问题在于，同时向 `8001` 发起了 3 次请求。试想，假如有 10 万个在并发请求该数据呢？那就会向 `8001` 同时发起 10 万次请求，如果 `8001` 又同时向数据库发起 10 万次查询请求，很容易导致缓存被击穿。

那这种情况下，我们如何做到只向远端节点发起一次请求呢？

解决方案：在准备从远程节点获取数据时，提前判断如果当前存在并发的请求同一个数据时，我们只让它向远程节点发起一个请求最后返回给当前并发请求的所有节点请求的数据即可。

## 3.singleflight.go

首先定义结构体call和Group

call结构体，用来记录key的请求状态，以及key的返回值

c.mu是用来保证任务执行完成

Group结构体，用来并发记录call组

`g.mu` 是保护 Group 的成员变量 `m` 不被并发读写而加上的锁。保证同一个数据不被多个协程多次更改。

```
type call struct {
    wg  sync.WaitGroup
    val interface{}
    err error
}
type Group struct {
    mu  sync.Mutex
    m   map[string]*Call
}
```

Do方法的实现

Do方法的作用，当有大量并发请求key时，只调用一次请求处理函数，返回返回值给所有请求

Do 的作用就是，针对相同的 key，无论 Do 被调用多少次，函数 `fn` 都只会被调用一次，等待 fn 调用结束了，返回返回值或错误。

```
func (g *Group) Do(key string, fn func()(interface{}, error)) (interface{}, error) {
    g.mu.Lock()
    if g.m == nil {
        g.m = make(map[string]*call)
    }
    if v, ok := g.m[key]; ok {
        g.mu.Unlock()
        v.wg.Wait()
        return v.val, v.err
    }
    c := new(Call)
    c.wg.Add(1)
    g.m[key] = c
    g.mu.Unlock()
    
    c.val, c.err = fn()
    c.wg.Done()
    
    g.mu.Lock()
    delete(g.m, key)
    g.mu.Unlock()
    
    return c.val, c.err
}
```

## 4.geecache.go

Group结构体新增属性loader，用来控制请求响应次数

```
type Group struct {
	name      string
	getter    Getter
	mainCache cache
	peers     PeerPicker
	// use singleflight.Group to make sure that
	// each key is only fetched once
	loader *singleflight.Group
}
```

NewGroup新增属性初始化

```
func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
    // ...
	g := &Group{
        // ...
		loader:    &singleflight.Group{},
	}
	return g
}
```

修改 `load` 函数，将原来的 load 的逻辑，使用 `g.loader.Do` 包裹起来即可，这样确保了并发场景下针对相同的 key，`load` 过程只会调用一次。

load方法之前作用，从远程节点获取数据，

现在改变，当大量节点请求并发同一个数据时，能够只从缓存服务器获取一次数据，返回给所有请求

```
func (g *Group) load(key string) (value ByteView, err error) {
	// each key is only fetched once (either locally or remotely)
	// regardless of the number of concurrent callers.
	viewi, err := g.loader.Do(key, func() (interface{}, error) {
		if g.peers != nil {
			if peer, ok := g.peers.PickPeer(key); ok {
				if value, err = g.getFromPeer(peer, key); err == nil {
					return value, nil
				}
				log.Println("[GeeCache] Failed to get from peer", err)
			}
		}

		return g.getLocally(key)
	})

	if err == nil {
		return viewi.(ByteView), nil
	}
	return
}
```


# 动手写分布式缓存 - GeeCache第一天 LRU 缓存淘汰策略

## LRU缓存机制

LRU 认为，如果数据最近被访问过，那么将来被访问的概率也会更高。LRU 算法的实现非常简单，维护一个队列，如果某条记录被访问了，则移动到队尾，那么队首则是最近最少访问的数据，淘汰该条记录即可。

## 为什么要采取优秀的缓存机制？

缓存全部存储在内存中，内存是有限的，因此不可能无限制地添加数据。假定我们设置缓存能够使用的内存大小为 N，那么在某一个时间点，添加了某一条缓存记录之后，占用内存超过了 N，这个时候就需要从缓存中移除一条或多条数据了。那移除谁呢？我们肯定希望尽可能移除“没用”的数据，那如何判定数据“有用”还是“没用”呢？

就要用到我们的缓存机制了

## 什么是缓存？为什么要用缓存？

缓存就是数据交换的缓冲区（称作Cache），是存贮数据（使用频繁的数据）的临时地方。当用户查询数据，首先在缓存中寻找，如果找到了则直接执行。如果找不到，则去数据库中查找。

缓存的本质就是用空间换时间，牺牲数据的实时性，以服务器内存中的数据暂时代替从数据库读取最新的数据。

缓存帮助我们减少数据库IO，减轻服务器压力，减少网络延迟，加快页面打开速度。

## 1.lru.go

 LRU 算法最核心的 2 个数据结构

- 绿色的是字典(map)，存储键和值的映射关系。这样根据某个键(key)查找对应的值(value)的复杂是`O(1)`，在字典中插入一条记录的复杂度也是`O(1)`。
- 红色的是双向链表(double linked list)实现的队列。将所有的值放到双向链表中，这样，当访问到某个值时，将其移动到队尾的复杂度是`O(1)`，在队尾新增一条记录以及删除一条记录的复杂度均为`O(1)`。

根据数据结构设计我们的缓存结构体

maxBytes 规定我们缓存的最大容量

nbytes    显示我们当前的缓存容量

ll 双向链表内部存储我们的键值对

cache 中存储了键和指向双向链表中键值对的指针

OnEvicted 编写在我们删除某个键值对后的处理操作

```
type Cache struct {
 maxBytes int64
 nbytes   int64
 ll       *list.List
 cache    map[string]*list.Element
 // optional and executed when an entry is purged.
 OnEvicted func(key string, value Value)
}
```

entry结构体用于存储键值对，

双向链表中可以直接存储一个含有键值对信息的entry结构体

便于我们对键与值的直接引用，淘汰队首节点时，需要用 key 从字典中删除对应的映射。

```
type entry struct {
 key   string
 value Value
}
```

设置Value接口

由于entry结构体内部存在Value这个属性，entry就实现了这个Value接口，

value就可以直接使用接口中的方法

```
type Value interface {
 Len() int
}
```

Len方法 返回值所占用内存的大小

```
func (c *Cache) Len() int {
 return c.ll.Len()
}
```

New函数实例化 `Cache`

```
func New(maxBytes int64, onEvicted func(string, Value)) *Cache {
 return &Cache{
  maxBytes:  maxBytes,
  ll:        list.New(),
  cache:     make(map[string]*list.Element),
  OnEvicted: onEvicted,
 }
}
```

Get函数查找指定键对应的值

查找主要有 3 个步骤，第一步是从字典中找到对应的双向链表的节点，第二步，将该节点移动到队尾。

第三步，返回查询结果

```
func (c *Cache) Get(key string) (value Value, ok bool) {
 if ele, ok := c.cache[key]; ok {
  c.ll.MoveToFront(ele)
  kv := ele.Value.(*entry)
  return kv.value, true
 }
 return
}
```

RemoveOldest函数是缓存淘汰

步骤：移除最近最少访问的节点（队首），删除map中对应映射关系，更新当前缓存；

执行回调函数处理已删除的键值对信息

```
func (c *Cache) RemoveOldest() {
 ele := c.ll.Back()
 if ele != nil {
  c.ll.Remove(ele)
  kv := ele.Value.(*entry)
  delete(c.cache, kv.key)
  c.nbytes -= int64(len(kv.key)) + int64(kv.value.Len())
  if c.OnEvicted != nil {
   c.OnEvicted(kv.key, kv.value)
  }
 }
}
```

Add函数机制：

- 如果键存在，则更新对应节点的值，并将该节点移到队尾。
- 不存在则是新增场景，首先队尾添加新节点 `&entry{key, value}`, 并字典中添加 key 和节点的映射关系。
- 更新 `c.nbytes`，如果超过了设定的最大值 `c.maxBytes`，则移除最少访问的节点。

```
func (c *Cache) Add(key string, value Value) {
 if ele, ok := c.cache[key]; ok {
  c.ll.MoveToFront(ele)
  kv := ele.Value.(*entry)
  c.nbytes += int64(value.Len()) - int64(kv.value.Len())
  kv.value = value
 } else {
  ele := c.ll.PushFront(&entry{key, value})
  c.cache[key] = ele
  c.nbytes += int64(len(key)) + int64(value.Len())
 }
 for c.maxBytes != 0 && c.maxBytes < c.nbytes {
  c.RemoveOldest()
 }
}
```

2.lru_test.go

测试lru中的函数能否根据我们的输入返回我们想要的结果

TestGet

测试逻辑

  目的：测试Get方法

  测试步骤：

1. 添加key1，value为1234
2. 获取key1，判断key1存在且value为1234
3. 获取key2，判断key2不存在

  若测试通过，不返回值

```
func TestGet(t *testing.T) {
 lru := New(int64(0), nil)
 lru.Add("key1", String("1234"))
 if v, ok := lru.Get("key1"); !ok || string(v.(String)) != "1234" {
  t.Fatalf("cache hit key1=1234 failed")
 }
 if _, ok := lru.Get("key2"); ok {
  t.Fatalf("cache miss key2 failed")
 }
}
```

TestRemoveoldest

测试逻辑

  测试步骤：

  1. 声明3个key1，key2，key3，key1的value为value1，key2的value为value2，key3的value为value3

  2. 设置缓存容量只能缓存key1和key2

  3. 添加key1和key2，此时key1和key2都存在

  3. 添加key3，key3的value为value3，此时key1由于缓存空间不足，key1被删除

  4. 获取key1，判断key1不存在

  若测试通过，不返回值

```
func TestRemoveoldest(t *testing.T) {
 k1, k2, k3 := "key1", "key2", "k3"
 v1, v2, v3 := "value1", "value2", "v3"
 cap := len(k1 + k2 + v1 + v2)
 lru := New(int64(cap), nil)
 lru.Add(k1, String(v1))
 lru.Add(k2, String(v2))
 lru.Add(k3, String(v3))

 if _, ok := lru.Get("key1"); ok || lru.Len() != 2 {
  t.Fatalf("Removeoldest key1 failed")
 }
}
```

TestOnEvicted

测试逻辑

  测试步骤：

  1. 声明一个回调函数，用于存储被删除的key

  2. 设置缓存容量只能缓存key1和k2

  3. 添加key1，key1的value为123456

  4. 添加k2，k2的value为k2

  5. 添加k3，k3的value为k3

  6. 添加k4，k4的value为k4

  7. 由于缓存空间不足，key1和k2应被删除

  8. keys 应存储被删除的key1和k2

  9. 设置keys的期望值expect = ["key1", "k2"]

  10. 比较keys和expect，若不相等，测试失败，返回语句

  若测试通过，不返回值

```
func TestOnEvicted(t *testing.T) {
 keys := make([]string, 0)
 callback := func(key string, value Value) {
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
```

## 3.测试命令

go test            运行所有测试函数

go test -v       运行所有测试函数并且会显示每个用例的测试结果

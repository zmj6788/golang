# 动手写分布式缓存 - GeeCache第二天 单机并发缓存

- ## 介绍 sync.Mutex 互斥锁的使用，并实现 LRU 缓存的并发控制

- ## 实现 GeeCache 核心数据结构 Group，缓存不存在时，调用回调函数获取源数据

## 1.byteview.go

ByteView 是我们规定缓存值的缓存具体类型

```
type ByteView struct {
 b []byte
}
```

添加Len()方法使其实现了Value接口，与底层entry保存类型保持一致

```
func (v ByteView) Len() int {
 return len(v.b)
}
```

ByteSlice方法用于调用cloneByte函数

```
func (v ByteView) ByteSlice() []byte {
 return cloneBytes(v.b)
}
func cloneBytes(b []byte) []byte {
 c := make([]byte, len(b))
 copy(c, b)
 return c
}
```

cloneBytes用于复制缓存值切片，保证底层缓存值，不被轻易改变

```
func cloneBytes(b []byte) []byte {
 c := make([]byte, len(b))
 copy(c, b)
 return c
}
```

## 2.cache.go

嵌套底层Cache结构体，封装加以改进使其可以并发

```
type cache struct {
 mu         sync.Mutex
 lru        *lru.Cache
 cacheBytes int64
}
```

add函数对底层Add函数进行加锁处理，实现可并发

```
func (c *cache) add(key string, value ByteView) {
 c.mu.Lock()
 defer c.mu.Unlock()
 if c.lru == nil {
  c.lru = lru.New(c.cacheBytes, nil)
 }
 c.lru.Add(key, value)
}
```

get函数对底层Get函数进行加锁处理，实现可并发

```
func (c *cache) get(key string) (value ByteView, ok bool) {
 c.mu.Lock()
 defer c.mu.Unlock()
 if c.lru == nil {
  return
 }

 if v, ok := c.lru.Get(key); ok {
  return v.(ByteView), ok
 }

 return
}
```

## 3.geecache.go

实现接口性函数类型GetterFunc

函数类型实现某一个接口，称之为接口型函数，方便使用者在调用时既能够传入函数作为参数，也能够传入实现了该接口的结构体作为参数。

```
type Getter interface {
 Get(key string) ([]byte, error)
}

// A GetterFunc implements Getter with a function.
type GetterFunc func(key string) ([]byte, error)

// Get implements Getter interface function
func (f GetterFunc) Get(key string) ([]byte, error) {
 return f(key)
}
```

Group结构体规定缓存组 属性name，getter回调函数 ，mainCache 缓存

声明groups  用groups[name]获取缓存组    用groups存储*Group 映射关系

声明mu        用mu进行加锁和解锁

NewGroup   实例化一个Group对象

GetGroup    根据name获取缓存组

```
type Group struct {
 name      string
 getter    Getter
 mainCache cache
}

var (
 mu     sync.RWMutex
 groups = make(map[string]*Group)
)

// NewGroup create a new instance of Group
func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
 if getter == nil {
  panic("nil Getter")
 }
 mu.Lock()
 defer mu.Unlock()
 g := &Group{
  name:      name,
  getter:    getter,
  mainCache: cache{cacheBytes: cacheBytes},
 }
 groups[name] = g
 return g
}

// GetGroup returns the named group previously created with NewGroup, or
// nil if there's no such group.
func GetGroup(name string) *Group {
 mu.RLock()
 g := groups[name]
 mu.RUnlock()
 return g
}
```

geecache核心方法Get，实现我们缓存获取的主要逻辑

```
                           是
接收 key --> 检查是否被缓存 -----> 返回缓存值 ⑴
                |  否                         是
                |-----> 是否应当从远程节点获取 -----> 与远程节点交互 --> 返回缓存值 ⑵
                            |  否
                            |-----> 调用`回调函数`，获取值并添加到缓存 --> 返回缓存值 ⑶
                           
func (g *Group) Get(key string) (ByteView, error) {
 if key == "" {
  return ByteView{}, fmt.Errorf("key is required")
 }

 if v, ok := g.mainCache.get(key); ok {
  log.Println("[GeeCache] hit")
  return v, nil
 }

 return g.load(key)
}
```

load在缓存中获取不到，是否到远程节点获取

暂时完成 否结果 调用回调函数，获取值并添加

```
func (g *Group) load(key string) (value ByteView, err error) {
 return g.getLocally(key)
}

```

getLocally函数调用回调函数，获取值并添加

```
func (g *Group) getLocally(key string) (ByteView, error) {
 bytes, err := g.getter.Get(key)
 if err != nil {
  return ByteView{}, err

 }
 value := ByteView{b: cloneBytes(bytes)}
 g.populateCache(key, value)
 return value, nil
}
```

populateCache函数实现增加值

```
func (g *Group) populateCache(key string, value ByteView) {
 g.mainCache.add(key, value)
}
```

## 4.geecache_test.go

测试Get逻辑功能是否成功实现

回调函数 实现从db数据源中获取数据并添加到缓存中 用loadCounts记录key的调用次数

遍历db 验证gee.Get能否根据key获取值，以及是否正确

验证key的访问次数是否大于1，期待值为1，不大于1，因为初始时缓存中一无所有，

但是gee.Get调用后，缓存中存储对应值，并且访问次数更改为1

验证gee.Get获取未声明值时，是否返回空值

若验证成功，均不反回错误信息

```
func TestGet(t *testing.T) {
 loadCounts := make(map[string]int, len(db))
 gee := NewGroup("scores", 2<<10, GetterFunc(
  func(key string) ([]byte, error) {
   log.Println("[SlowDB] search key", key)
   if v, ok := db[key]; ok {
    if _, ok := loadCounts[key]; !ok {
     loadCounts[key] = 0
    }
    loadCounts[key] += 1
    return []byte(v), nil
   }
   return nil, fmt.Errorf("%s not exist", key)
  }))

 for k, v := range db {
  if view, err := gee.Get(k); err != nil || view.String() != v {
   t.Fatal("failed to get value of Tom")
  } // load from callback function
  if _, err := gee.Get(k); err != nil || loadCounts[k] > 1 {
   t.Fatalf("cache %s miss", k)
  } // cache hit
 }

 if view, err := gee.Get("unknown"); err == nil {
  t.Fatalf("the value of unknow should be empty, but %s got", view)
 }
}
```

测试Getter回调函数是否成功能调用自身函数或实现了接口的结构体的函数

简单实现回调函数返回 key的切片

判断f.Get(key)返回结果与期待值expect是否相同

相同测试成功

```
func TestGetter(t *testing.T) {
 var f Getter = GetterFunc(func(key string) ([]byte, error) {
  return []byte(key), nil
 })

 expect := []byte("key")
 if v, _ := f.Get("key"); !reflect.DeepEqual(v, expect) {
  t.Errorf("callback failed")
 }
}
```

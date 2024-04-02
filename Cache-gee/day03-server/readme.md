# 动手写分布式缓存 - GeeCache第三天 HTTP 服务端

- ## 介绍如何使用 Go 语言标准库 `http` 搭建 HTTP Server

- ## 并实现 main 函数启动 HTTP Server 测试 API

今天的功能实现意味着其他服务器也可以获取我们服务器上的内存中的缓存值了

## GeeCache HTTP 服务端

分布式缓存需要实现节点间通信，建立基于 HTTP 的通信机制是比较常见和简单的做法。如果一个节点启动了 HTTP 服务，那么这个节点就可以被其他节点访问。今天我们就为单机节点搭建 HTTP Server。

不与其他部分耦合，我们将这部分代码放在新的 `http.go` 文件中，当前的代码结构如下：

```
geecache/
    |--lru/
        |--lru.go  // lru 缓存淘汰策略
    |--byteview.go // 缓存值的抽象与封装
    |--cache.go    // 并发控制
    |--geecache.go // 负责与外部交互，控制缓存存储和获取的主流程
 |--http.go     // 提供被其他节点访问的能力(基于http)
```

## 1.http.go

声明HTTPPool结构体，结构体实现ServeHTTP方法，就实现了Handle接口，可以将结构体作为参数传递给

http.ListenAndServe的第二个参数

而该结构体的其他属性，也就可以在ServeHTTP方法中使用了，更好的实现节点的http通信服务

```
const defaultBasePath = "/_geecache/"

// HTTPPool implements PeerPicker for a pool of HTTP peers.
type HTTPPool struct {
 // this peer's base URL, e.g. "https://example.net:8000"
 self     string
 basePath string
}
```

NewHTTPPool实例化HTTPPool对象

```
func NewHTTPPool(self string) *HTTPPool {
 return &HTTPPool{
  self:     self,
  basePath: defaultBasePath,
 }
}
```

Log打印日志函数，对log的加工处理，更好的表明是什么请求，什么请求路径，便于我们开发者查看

```
func (p *HTTPPool) Log(format string, v ...interface{}) {
 log.Printf("[Server %s] %s", p.self, fmt.Sprintf(format, v...))
}
```

ServeHTTP方法，用于开启一个http服务，内部实现了我们自己的请求处理返回响应逻辑

首先查看请求是否是节点通信请求，是继续向下

打印请求类型和请求路径，分割请求成功，继续向下

分割请求路径为组名和键，根据组名获取缓存组

在缓存组中根据键获取值，获取成功

设置响应头，响应返回转换值的复制切片

至此实现http服务获取缓存值

```
func (p *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
 if !strings.HasPrefix(r.URL.Path, p.basePath) {
  panic("HTTPPool serving unexpected path: " + r.URL.Path)
 }
 p.Log("%s %s", r.Method, r.URL.Path)
 // /<basepath>/<groupname>/<key> required
 parts := strings.SplitN(r.URL.Path[len(p.basePath):], "/", 2)
 if len(parts) != 2 {
  http.Error(w, "bad request", http.StatusBadRequest)
  return
 }

 groupName := parts[0]
 key := parts[1]

 group := GetGroup(groupName)
 if group == nil {
  http.Error(w, "no such group: "+groupName, http.StatusNotFound)
  return
 }

 view, err := group.Get(key)
 if err != nil {
  http.Error(w, err.Error(), http.StatusInternalServerError)
  return
 }

 w.Header().Set("Content-Type", "application/octet-stream")
 w.Write(view.ByteSlice())
}
```

## 2.执行流程分析

首先设置本地数据源

```
var db = map[string]string{
 "Tom":  "630",
 "Jack": "589",
 "Sam":  "567",
}
```

main.go示例代码

```
func main() {
 geecache.NewGroup("scores", 2<<10, geecache.GetterFunc(
  func(key string) ([]byte, error) {
   log.Println("[SlowDB] search key", key)
   if v, ok := db[key]; ok {
    return []byte(v), nil
   }
   return nil, fmt.Errorf("%s not exist", key)
  }))

 addr := "localhost:9999"
 peers := geecache.NewHTTPPool(addr)
 log.Println("geecache is running at", addr)
 log.Fatal(http.ListenAndServe(addr, peers))
}
```

首先创建缓存组，为缓存组规定回调函数，回调函数作用，从数据源中获取数据并返回

接着实例化HTTPPool对象，设置节点地址

最后开启http服务监听

当监听到addr地址发出请求后，会执行结构体内的ServeHTTP请求响应方法

ServeHTTP方法内部

首先查看请求是否是节点通信请求，是继续向下

打印请求类型和请求路径，分割请求成功，继续向下

分割请求路径为组名和键，根据组名获取缓存组

在缓存组中根据键获取值，获取成功

设置响应头，响应返回转换值的复制切片

响应返回到addr端口

至此实现缓存值的http通信

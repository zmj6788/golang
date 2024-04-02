# 动手写分布式缓存 - GeeCache第五天 分布式节点

- 注册节点(Register Peers)，借助一致性哈希算法选择节点。
- 实现 HTTP 客户端，与远程节点的服务端通信，

今日实现流程（2），之前以实现流程（1）（3）

```
                            是
接收 key --> 检查是否被缓存 -----> 返回缓存值 ⑴
                |  否                         是
                |-----> 是否应当从远程节点获取 -----> 与远程节点交互 --> 返回缓存值 ⑵
                            |  否
                            |-----> 调用`回调函数`，获取值并添加到缓存 --> 返回缓存值 ⑶
```

## 1.peers.go

封装HTTPPool结构体，使用在Group中

让HTTPPool实现PeerPicker接口，这样在Group中新增peers PeerPicker 时可以将HTTPPool结构体传入

让HTTPPool中httpGetters map[string]*httpGetter的httpGetter实现PeerGetter接口

httpGetter的Get为 `HTTPPool` 实现客户端的功能

- 在这里，抽象出 2 个接口，PeerPicker 的 `PickPeer()` 方法用于根据传入的 key 选择相应节点 PeerGetter。
- 接口 PeerGetter 的 `Get()` 方法用于从对应 group 查找缓存值。PeerGetter 就对应于上述流程中的 HTTP 客户端。

```
type PeerPicker interface {
 PickPeer(key string) (peer PeerGetter, ok bool)
}

// PeerGetter is the interface that must be implemented by a peer.
type PeerGetter interface {
 Get(group string, key string) ([]byte, error)
}
```

## 2.http.go

定义httpGetter结构体，作用：为 `HTTPPool` 实现客户端的功能

Get方法，简单实现客户端，根据缓存组名与键再与基本路径拼接URL请求获取返回值

```
type httpGetter struct {
 baseURL string
}

func (h *httpGetter) Get(group string, key string) ([]byte, error) {
 u := fmt.Sprintf(
  "%v%v/%v",
  h.baseURL,
  url.QueryEscape(group),
  url.QueryEscape(key),
 )
 res, err := http.Get(u)
 if err != nil {
  return nil, err
 }
 defer res.Body.Close()

 if res.StatusCode != http.StatusOK {
  return nil, fmt.Errorf("server returned: %v", res.Status)
 }

 bytes, err := ioutil.ReadAll(res.Body)
 if err != nil {
  return nil, fmt.Errorf("reading response body: %v", err)
 }

 return bytes, nil
}

var _ PeerGetter = (*httpGetter)(nil)
```

HTTPPool结构体新增属性

mu                      用于实现并发

peers                  使HTTPPool结构体嵌套Map,便于封装其中方法，实现节点匹配和增加节点

httpGetters          用于映射远程节点与其客户端的关系

```
const (
 defaultBasePath = "/_geecache/"
 defaultReplicas = 50
)
// HTTPPool implements PeerPicker for a pool of HTTP peers.
type HTTPPool struct {
 // this peer's base URL, e.g. "https://example.net:8000"
 self        string
 basePath    string
 mu          sync.Mutex // guards peers and httpGetters
 peers       *consistenthash.Map
 httpGetters map[string]*httpGetter // keyed by e.g. "http://10.0.0.2:8008"
}
```

Set方法实例化了一致性哈希算法，并且添加了传入的节点。

 并为每一个节点创建了一个 HTTP 客户端 httpGetter。

```
func (p *HTTPPool) Set(peers ...string) {
 p.mu.Lock()
 defer p.mu.Unlock()
 p.peers = consistenthash.New(defaultReplicas, nil)
 p.peers.Add(peers...)
 p.httpGetters = make(map[string]*httpGetter, len(peers))
 for _, peer := range peers {
  p.httpGetters[peer] = &httpGetter{baseURL: peer + p.basePath}
 }
}

// PickPeer picks a peer according to key
func (p *HTTPPool) PickPeer(key string) (PeerGetter, bool) {
 p.mu.Lock()
 defer p.mu.Unlock()
 if peer := p.peers.Get(key); peer != "" && peer != p.self {
  p.Log("Pick peer %s", peer)
  return p.httpGetters[peer], true
 }
 return nil, false
}

var _ PeerPicker = (*HTTPPool)(nil)
```

PickerPeer()包装了一致性哈希算法的 Get() 方法，根据具体的 key，选择节点，返回节点对应的 HTTP 客户端。

```
func (p *HTTPPool) PickPeer(key string) (PeerGetter, bool) {
 p.mu.Lock()
 defer p.mu.Unlock()
 if peer := p.peers.Get(key); peer != "" && peer != p.self {
  p.Log("Pick peer %s", peer)
  return p.httpGetters[peer], true
 }
 return nil, false
}

var _ PeerPicker = (*HTTPPool)(nil)
```

至此，HTTPPool 既具备了提供 HTTP 服务的能力，也具备了根据具体的 key，创建 HTTP 客户端从远程节点获取缓存值的能力。

3.geecache.go

将前面实现的功能最后一次封装，便于用户的直接使用

Group中新增属性

peers     是PeerPicker接口类型的

那么实现了PeerPicker接口的HTTPPool结构体，就可以传入，就可以进行最后一次的封装

```
type Group struct {
 name      string
 getter    Getter
 mainCache cache
 peers     PeerPicker
}
```

RegisterPeers方法用于为一个缓存组注册节点，即将处理好的HTTPPool结构体传参给缓存组的peers

```
func (g *Group) RegisterPeers(peers PeerPicker) {
 if g.peers != nil {
  panic("RegisterPeerPicker called more than once")
 }
 g.peers = peers
}

```

getFromPeer方法使用PeerGetter接口的Get方法，即使用客户端获取信息

```
func (g *Group) getFromPeer(peer PeerGetter, key string) (ByteView, error) {
 bytes, err := peer.Get(g.name, key)
 if err != nil {
  return ByteView{}, err
 }
 return ByteView{b: bytes}, nil
}
```

load更新逻辑， g.peers不为空，说明已经注册过HTTPPool对象信息，即可以链接远程节点，

那么再去根据键匹配节点，匹配节点成功，再根据键和客户端获取缓存值

```
func (g *Group) load(key string) (value ByteView, err error) {
 if g.peers != nil {
  if peer, ok := g.peers.PickPeer(key); ok {
   if value, err = g.getFromPeer(peer, key); err == nil {
    return value, nil
   }
   log.Println("[GeeCache] Failed to get from peer", err)
  }
 }

 return g.getLocally(key)
}
```

4.main.go

main.go示例代码,geecache的简单使用

```
var db = map[string]string{
 "Tom":  "630",
 "Jack": "589",
 "Sam":  "567",
}

func createGroup() *geecache.Group {
 return geecache.NewGroup("scores", 2<<10, geecache.GetterFunc(
  func(key string) ([]byte, error) {
   log.Println("[SlowDB] search key", key)
   if v, ok := db[key]; ok {
    return []byte(v), nil
   }
   return nil, fmt.Errorf("%s not exist", key)
  }))
}

func startCacheServer(addr string, addrs []string, gee *geecache.Group) {
 peers := geecache.NewHTTPPool(addr)
 peers.Set(addrs...)
 gee.RegisterPeers(peers)
 log.Println("geecache is running at", addr)
 log.Fatal(http.ListenAndServe(addr[7:], peers))
}

func startAPIServer(apiAddr string, gee *geecache.Group) {
 http.Handle("/api", http.HandlerFunc(
  func(w http.ResponseWriter, r *http.Request) {
   key := r.URL.Query().Get("key")
   view, err := gee.Get(key)
   if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
   }
   w.Header().Set("Content-Type", "application/octet-stream")
   w.Write(view.ByteSlice())

  }))
 log.Println("fontend server is running at", apiAddr)
 log.Fatal(http.ListenAndServe(apiAddr[7:], nil))

}

func main() {
 var port int
 var api bool
 flag.IntVar(&port, "port", 8001, "Geecache server port")
 flag.BoolVar(&api, "api", false, "Start a api server?")
 flag.Parse()

 apiAddr := "http://localhost:9999"
 addrMap := map[int]string{
  8001: "http://localhost:8001",
  8002: "http://localhost:8002",
  8003: "http://localhost:8003",
 }

 var addrs []string
 for _, v := range addrMap {
  addrs = append(addrs, v)
 }

 gee := createGroup()
 if api {
  go startAPIServer(apiAddr, gee)
 }
 startCacheServer(addrMap[port], []string(addrs), gee)
}
```

主要步骤流程：

创建缓存组，

```
func createGroup() *geecache.Group {
 return geecache.NewGroup("scores", 2<<10, geecache.GetterFunc(
  func(key string) ([]byte, error) {
   log.Println("[SlowDB] search key", key)
   if v, ok := db[key]; ok {
    return []byte(v), nil
   }
   return nil, fmt.Errorf("%s not exist", key)
  }))
}
```

创建已添加远程节点的的匹配机制和远程节点与其客户端映射的HTTPPool对象

注册HTTPPool对象到缓存组

通过HTTPPool开启缓存服务

```
func startCacheServer(addr string, addrs []string, gee *geecache.Group) {
 peers := geecache.NewHTTPPool(addr)
 peers.Set(addrs...)
 gee.RegisterPeers(peers)
 log.Println("geecache is running at", addr)
 log.Fatal(http.ListenAndServe(addr[7:], peers))
}
```

注册一个/api服务，监听到以/api开头会执行http.Handle内绑定函数，

函数作用：获取请求路径key值，根据key进入缓存组的Get逻辑

最终返回Get逻辑最终值作为响应

```
func startAPIServer(apiAddr string, gee *geecache.Group) {
 http.Handle("/api", http.HandlerFunc(
  func(w http.ResponseWriter, r *http.Request) {
   key := r.URL.Query().Get("key")
   view, err := gee.Get(key)
   if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
   }
   w.Header().Set("Content-Type", "application/octet-stream")
   w.Write(view.ByteSlice())

  }))
 log.Println("fontend server is running at", apiAddr)
 log.Fatal(http.ListenAndServe(apiAddr[7:], nil))

}
```

缓存组的Get,分布式缓存的逻辑处理机制

其中远程节点是通过数据源获取数据的，其余的是从内存中获取数据的

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

至此分布式缓存实现。即分布式缓存与检索

# 动手写分布式缓存 - GeeCache第七天 使用 Protobuf 通信

- ## 为什么要使用 protobuf？

- ## 使用 protobuf 进行节点间通信，编码报文，提高效率。

## 1.为什么使用protobuf?

protobuf 广泛地应用于远程过程调用(RPC) 的二进制传输，使用 protobuf 的目的非常简单，为了获得更高的性能。传输前使用 protobuf 编码，接收方再进行解码，可以显著地降低二进制传输的大小。另外一方面，protobuf 可非常适合传输结构化数据，便于通信字段的扩展。

使用 protobuf 一般分为以下 2 步：

- 按照 protobuf 的语法，在 `.proto` 文件中定义数据结构，并使用 `protoc` 生成 Go 代码（`.proto` 文件是跨平台的，还可以生成 C、Java 等其他源码文件）。
- 在项目代码中引用生成的 Go 代码。

## 2.如何使用protobuf?

2.1 protoc的安装

protobuf的github发布地址： https://github.com/protocolbuffers/protobuf/releases

2.2 protoc-gen-go插件的安装。

```
go get -u github.com/golang/protobuf/protoc-gen-go
```

2.3 定义消息类型

```
syntax = "proto3";
 
package geecachepb;
//用于规定.go文件生成位置
option go_package = ".";

message Request {
  string group = 1;
  string key = 2;
}

message Response {
  bytes value = 1;
}
```

2.4 定义服务

如果消息类型是用来远程通信的(Remote Procedure Call, RPC)，可以在 .proto 文件中定义 RPC 服务接口。

```
service GroupCache {
  rpc Get(Request) returns (Response);
}
```

2.5 命令生成go代码

```
protoc --go_out=. *.proto
```

## 3.使用 protobuf 进行节点间通信

我们使用protobuf进行节点通信的传输位置，缓存服务的客户端和服务端

```
服务端-->将信息二进制处理写入响应-->客户端-->获取响应解码到pb.对象中-->pb.对象-->得到想要的信息
```

peers.go的PeerGetter接口方法参数改变

```
import pb "geecache/geecachepb"

type PeerGetter interface {
	Get(in *pb.Request, out *pb.Response) error
}
```

geecache.go的改变，跟PeerGetter接口有关的改变参数

获取节点中搜索到的数据，直接从`pb.Response`对象中获取

```
func (g *Group) getFromPeer(peer PeerGetter, key string) (ByteView, error) {
	req := &pb.Request{
		Group: g.name,
		Key:   key,
	}
	res := &pb.Response{}
	err := peer.Get(req, res)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{b: res.Value}, nil
}
```

http.go的改变

服务端编写响应前，进行proto编码，将序列化后的缓存值写入响应体中返回给客户端。

客户端获取响应信息后，进行`proto.Unmarshal()`方法反序列化到`pb.Response`对象中

```
func (p *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    // ...
	// Write the value to the response body as a proto message.
	body, err := proto.Marshal(&pb.Response{Value: view.ByteSlice()})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(body)
}

func (h *httpGetter) Get(in *pb.Request, out *pb.Response) error {
	u := fmt.Sprintf(
		"%v%v/%v",
		h.baseURL,
		url.QueryEscape(in.GetGroup()),
		url.QueryEscape(in.GetKey()),
	)
    res, err := http.Get(u)
	// ...
	if err = proto.Unmarshal(bytes, out); err != nil {
		return fmt.Errorf("decoding response body: %v", err)
	}

	return nil
}
```

## 4.运行结果

```
$ ./run.sh
2024/04/08 14:52:21 geecache is running at  http://localhost:8001
2024/04/08 14:52:21 geecache is running at  http://localhost:8003
2024/04/08 14:52:21 geecache is running at  http://localhost:8002
2024/04/08 14:52:21 fontend server is running at  http://localhost:9999
>>> start test
2024/04/08 14:52:22 [Server http://localhost:8003] Pick peer http://localhost:8001
2024/04/08 14:52:22 [Server http://localhost:8001] GET /_geecache/scoresTom
2024/04/08 14:52:22 [GeeCache] Failed to get from peer server returned: 400 Bad Request
2024/04/08 14:52:22 [SlowDB] search key Tom
666666666666666666666666666

```


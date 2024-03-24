package main

import (
	"fmt"
	"log"
	"net/http"
)

type Engine struct{}

func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	switch req.URL.Path {
	case "/":
		fmt.Fprintf(w, "URL.Path = %q\n", req.URL.Path)
	case "/hello":
		for k, v := range req.Header {
			fmt.Fprintf(w, "%s: %v\n", k, v)
		}
		fmt.Fprintf(w, "URL.Path = %q\n", req.URL.Path)
	default:
		fmt.Fprintf(w, "404 NOT FOUND : %s\n", req.URL)
	}
}

// 实现了接口方法的 struct 都可以强制转换为接口类型。

func main() {
	// 在实现Engine之后，我们拦截了所有的HTTP请求，
	// 拥有了统一的控制入口。在这里我们可以自由定义路由映射的规则，
	// 也可以统一添加一些处理逻辑，例如日志、异常处理等。

	/*
		第二个参数类型是接口类型 http.Handler ，
		Handler 的定义博文中已经贴了，是从 http 的源码中找到的。

		type Handler interface {
		    ServeHTTP(w ResponseWriter, r *Request)
		}

		func ListenAndServe(address string, h Handler) error
		在 Go 语言中，实现了接口方法的 struct 都可以强制转换为接口类型。你可以这么写：

		handler := (http.Handler)(engine) // 手动转换为借口类型
		log.Fatal(http.ListenAndServe(":9999", handler))
		然后，ListenAndServe 方法里面会去调用 handler.ServeHTTP() 方法，
		你感兴趣，可以在 http 的源码中找到调用的地方。但是这么写是多余的，
		传参时，会自动进行参数转换的。所以直接传入engine 即可。
	*/
	engine := new(Engine)
	log.Fatal(http.ListenAndServe(":9999", engine))
}

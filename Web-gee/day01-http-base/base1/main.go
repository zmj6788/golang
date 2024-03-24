package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	// 在实现Engine之前，
	// 我们调用 http.HandleFunc 实现了路由和Handler的映射，
	// 也就是只能针对具体的路由写处理逻辑。

	// http.HandleFunc("/", indexHandler)是一个Go语言的方法，
	// 用于设置URL路径和对应的处理函数。
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/hello", helloHandler)
	log.Fatal(http.ListenAndServe(":9999", nil))
}

func indexHandler(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "URL.Path = %q\n", req.URL.Path)
}

func helloHandler(w http.ResponseWriter, req *http.Request) {
	for k, v := range req.Header {
		fmt.Fprintf(w, "Header[%q] = %q\n", k, v)
	}
}

package main

import (
	"fmt"
	"gee"
	"net/http"
	"text/template"
	"time"
)

type student struct {
	Name string
	Age  int8
}

func FormatAsDate(t time.Time) string {
	year, month, day := t.Date()
	return fmt.Sprintf("%d-%02d-%02d", year, month, day)
}
func main() {
	r := gee.New()
	r.Use(gee.Logger())
	r.SetFuncMap(template.FuncMap{
		"FormatAsDate": FormatAsDate,
	})
	r.LoadHTMLGlob("templates/*")
	// 相当于是在本地./static目录下注册了一个（虚拟的）静态文件服务器函数
	// 并将其绑定给路由assets，注册到路由组中，
	// 这样当访问/assets/index.html时，
	// 就会调用这个函数并读取在本地的./static目录下的文件
	r.Static("/assets", "./static")

	stu1 := &student{Name: "Geektutu", Age: 20}
	stu2 := &student{Name: "Jack", Age: 22}
	r.GET("/", func(c *gee.Context) {
		c.HTML(http.StatusOK, "css.tmpl", nil)
	})
	r.GET("/students", func(c *gee.Context) {
		c.HTML(http.StatusOK, "arr.tmpl", gee.H{
			"title":  "gee",
			"stuArr": [2]*student{stu1, stu2},
		})
	})
	// 当我们访问 http://localhost:9999/panic的时候，页面会打不开
	// 因为发生了panic，程序崩溃.
	// 我们会在day07-PanicRecover中实现手动捕获panic并且不让程序崩溃
	// 而是转向了我们自己的错误处理机制
	// http://localhost:9999/panic仍然能够正常打开，
	// 但是会显示我们自己的错误处理机制处理后的页面
	// 其他接口请求能够正常调用的原因，你不是说发生panic后整个程序会崩溃吗？
	// 哦，你连这都不知道
	// 其实这真的很难知道
	// 一句话因为并发
	/*
		在 Go 语言中，当某个 goroutine 遇到 panic 时，
		该 goroutine 会立即停止其正常执行流程，并开始执行 panic 息传递链，
		即沿着调用栈向上回溯，直到遇到 recover 函数或者所有 goroutine 上层函数
		均因 panic 导致返回。如果没有被捕获，最终会导致程序崩溃。

		然而，在 Web 服务等并发环境中，
		通常每个 HTTP 请求都是在一个独立的 goroutine 中处理的。
		因此，如果 /panic 接口触发了一个 panic，只会导致处理这个请求的 goroutine 崩溃，
		并不会直接影响到其他正在处理不同请求的 goroutine。
	*/
	r.GET("/panic", func(c *gee.Context) {
		names := []string{"geektutu"}
		c.String(http.StatusOK, names[100])
	})

	r.GET("/date", func(c *gee.Context) {
		c.HTML(http.StatusOK, "custom_func.tmpl", gee.H{
			"title": "gee",
			"now":   time.Date(2019, 8, 17, 0, 0, 0, 0, time.UTC),
		})
	})

	r.Run(":9999")
}

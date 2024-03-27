package main

import (
	"gee"
	"net/http"
)

type student struct {
	Name string
	Age  uint8
}

func main() {
	r := gee.Default()
	r.LoadHTMLGlob("templates/*")
	r.GET("/panic", func(c *gee.Context) {
		names := []string{"geektutu"}
		c.String(http.StatusOK, names[100])
	})
	r.GET("/", func(c *gee.Context) {
		c.HTML(http.StatusOK, "arr.tmpl", gee.H{
			"title": "gee",
			"stuArr": []student{
				{Name: "张三", Age: 18},
				{Name: "李四", Age: 19},
				{Name: "王五", Age: 20},
			},
		})
	})
	r.Run(":9999")
}

package gee

import (
	"log"
	"time"
)

func Logger() HandlerFunc {
	return func(c *Context) {
		// 开启计时
		t := time.Now()
		// 控制中间件和路由匹配函数的执行顺序
		// 在这里的执行顺序为
		// 首先运行中间件函数Logger()
		// 待t := time.Now()运行结束后，由于c.Next()存在会去运行路由匹配函数
		// 等待路由匹配函数执行完毕后，再运行Logger()中的log.Printf
		c.Next()
		// 计时结束
		log.Printf("[%d] %s in %v", c.StatusCode, c.Req.RequestURI, time.Since(t))

	}
}

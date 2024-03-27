package gee

import (
	"fmt"
	"log"
	"net/http"
	"runtime"
	"strings"
)

// 用来获取触发 panic 的堆栈信息
func trace(message string) string {
	var pcs [32]uintptr
	n := runtime.Callers(3, pcs[:])

	var str strings.Builder
	str.WriteString(message + "\nTraceback:")
	for _, pc := range pcs[:n] {
		fn := runtime.FuncForPC(pc)
		file, line := fn.FileLine(pc)
		str.WriteString(fmt.Sprintf("\n\t%s:%d", file, line))
	}
	return str.String()
}

// 定义错误处理的默认中间件
// 作用是实现我们手动捕获panic,执行我们自己的错误处理机制
func Recovery() HandlerFunc {
	return func(c *Context) {
		defer func() {
			if err := recover(); err != nil {
				message := fmt.Sprintf("%s", err)
				log.Printf("%s\n\n", trace(message))
				c.Fail(http.StatusInternalServerError, "Internal Server Error")
			}
		}()
		c.Next()
		// 错误处理机制将要进行错误处理的程序与错误处理函数放在一起
		/*
			func test_recover() {
				defer func() {
					fmt.Println("defer func")
					if err := recover(); err != nil {
						fmt.Println("recover success")
					}
				}()
					//当我们要处理的程序出现错误时，程序想要中断
					//由于defer 的机制
					//会在程序中断前先运行其后的自执行函数
					//这个自执行函数会捕获到panic实现我们手动捕获panic,
					//执行我们自己的错误处理机制
					//然后当前程序中断
					//但主程序中的其他部分不受影响
				arr := []int{1, 2, 3}
				fmt.Println(arr[4])
				//到这里当前程序中出现错误，程序中断，后续代码不执行
				fmt.Println("after panic")
			}

			func main() {
				test_recover()
				// 主程序中的其他部分不受影响，仍正常运行
				fmt.Println("after recover")
			}
		*/
	}
}

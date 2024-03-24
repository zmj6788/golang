package gee

import (
	"encoding/json"
	"fmt"
	"net/http"
)

/*
这段代码定义了一个名为H的类型，
它是map[string]interface{}类型的别名。
map[string]interface{}是一个Go语言中的映射类型，
其中键是字符串类型，值是任意类型。
这个类型的定义可以用来简化代码中对map[string]interface{}的多次使用，
提高代码的可读性和可维护性。
*/
type H map[string]interface{}

type Context struct {
	Writer http.ResponseWriter
	Req    *http.Request
	// 请求信息
	Path   string
	Method string
	// 响应信息
	StatusCode int
}

func newContext(w http.ResponseWriter, req *http.Request) *Context {
	return &Context{
		Writer: w,
		Req:    req,
		Path:   req.URL.Path,
		Method: req.Method,
	}
}

func (c *Context) PostForm(key string) string {
	return c.Req.FormValue(key)
}

func (c *Context) Query(key string) string {
	return c.Req.URL.Query().Get(key)
}

func (c *Context) Status(code int) {
	c.StatusCode = code
	c.Writer.WriteHeader(code)
}

func (c *Context) SetHeader(key string, value string) {
	c.Writer.Header().Set(key, value)
}

func (c *Context) String(code int, format string, values ...interface{}) {
	c.SetHeader("Content-Type", "text/plain")
	c.Status(code)
	c.Writer.Write([]byte(fmt.Sprintf(format, values...)))
}

func (c *Context) JSON(code int, obj interface{}) {
	c.SetHeader("Content-Type", "application/json")
	c.Status(code)
	// 接着创建一个json.NewEncoder(c.Writer)，
	// 用于将obj对象编码为JSON格式，并写入到响应体中。
	encoder := json.NewEncoder(c.Writer)
	if err := encoder.Encode(obj); err != nil {
		http.Error(c.Writer, err.Error(), 500)
	}
	/* encoder := json.NewEncoder(c.Writer) 语句
	创建一个新的 JSON 编码器并将其分配给 encoder 变量。
	json.NewEncoder 函数创建一个新的 JSON 编码器，
	它可以将 Go 值编码为 JSON 格式。该函数接受一个 io.Writer 接口作为参数，
	该接口表示一个写入流，编码器将 JSON 格式的数据写入该流。
	在本例中，c.Writer 是一个 http.ResponseWriter 接口，它表示一个 HTTP 响应的写入流。
	json.NewEncoder 函数使用 c.Writer 作为参数创建一个新的 JSON 编码器，
	该编码器将 JSON 格式的数据写入 HTTP 响应的正文。
	encoder 变量是一个指向 JSON 编码器的指针。
	它用于将 Go 值编码为 JSON 格式并将其写入 HTTP 响应的正文。
	if err := encoder.Encode(obj); err != nil 语句
	使用 encoder 编码 obj 值并将其写入 HTTP 响应的正文。
	如果编码过程中发生错误，则该语句将使用 http.Error 函数向客户端发送一个
	HTTP 500 内部服务器错误响应。
	*/

}

func (c *Context) Data(code int, data []byte) {
	c.Status(code)
	c.Writer.Write(data)
}
func (c *Context) HTML(code int, html string) {
	c.SetHeader("Content-Type", "text/html")
	c.Status(code)
	c.Writer.Write([]byte(html))
}

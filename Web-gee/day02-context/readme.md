# Go语言动手写Web框架 - Gee第二天 上下文Context

## 今天代码实现的新功能或特性

- 将`路由(router)`独立出来，方便之后增强。
- 设计`上下文(Context)`，封装 Request 和 Response ，提供对 JSON、HTML 等返回类型的支持。
- 动手写 Gee 框架的第二天，**框架代码140行，新增代码约90行**

针对使用场景，封装`*http.Request`和`http.ResponseWriter`的方法，简化相关接口的调用，只是设计 Context 的原因之一。对于框架来说，还需要支撑额外的功能。例如，将来解析动态路由`/hello/:name`，参数`:name`的值放在哪呢？再比如，框架需要支持中间件，那中间件产生的信息放在哪呢？Context 随着每一个请求的出现而产生，请求的结束而销毁，和当前请求强相关的信息都应由 Context 承载。因此，设计 Context 结构，扩展性和复杂性留在了内部，而对外简化了接口。路由的处理函数，以及将要实现的中间件，参数都统一使用 Context 实例， Context 就像一次会话的百宝箱，可以找到任何东西。

## 1.context.go

代码最开头，给`map[string]interface{}`起了一个别名`gee.H`，构建JSON数据时，显得更简洁。

直接通过gee.H调用

```
type H map[string]interface{}
```

Context`目前只包含了`http.ResponseWriter`和`*http.Request`，另外提供了对 Method 和 Path 这两个常用属性的直接访问。

```
type Context struct {
 Writer http.ResponseWriter
 Req    *http.Request
 // 请求信息
 Path   string
 Method string
 // 响应信息
 StatusCode int
}
```

提供了访问Query和PostForm参数的方法。

```
func (c *Context) PostForm(key string) string {
 return c.Req.FormValue(key)
}

func (c *Context) Query(key string) string {
 return c.Req.URL.Query().Get(key)
}
```

Query方法用于从HTTP请求的URL查询参数中获取指定键(key)对应的值(value)

```
r.GET("/hello", func(c *gee.Context) {
  // expect /hello?name=geektutu
  c.String(http.StatusOK, "hello %s, you're at %s\n", c.Query("name"), c.Path)
 })
curl "http://localhost:9999/hello?name=geektutu"
hello geektutu, you're at /hello
```

PostForm方法具体功能是从HTTP请求中获取POST表单中指定键(`key`)的值，并将其作为结果返回。

```
r.POST("/login", func(c *gee.Context) {
  c.JSON(http.StatusOK, gee.H{
   "username": c.PostForm("username"),
   "password": c.PostForm("password"),
  })
 })
curl "http://localhost:9999/login" -X POST -d 'username=geektutu&password=1234'
{"password":"1234","username":"geektutu"}
```

提供了快速构造String/Data/JSON/HTML响应的方法。

## 2.router.go

我们将和路由相关的方法和结构提取了出来，放到了一个新的文件中`router.go`，方便我们下一次对 router 的功能进行增强，例如提供动态路由的支持。 router 的 handle 方法作了一个细微的调整，即 handler 的参数，变成了 Context。

```
type router struct {
 handlers map[string]HandlerFunc
}

func newRouter() *router {
 return &router{handlers: make(map[string]HandlerFunc)}
}

func (r *router) addRoute(method string, pattern string, handler HandlerFunc) {
 log.Printf("Route %4s - %s", method, pattern)
 key := method + "-" + pattern
 r.handlers[key] = handler
}

func (r *router) handle(c *Context) {
 key := c.Method + "-" + c.Path
 if handler, ok := r.handlers[key]; ok {
  handler(c)
 } else {
  c.String(http.StatusNotFound, "404 NOT FOUND: %s\n", c.Path)
 }
}
```

## 3,gee.go

将`router`相关的代码独立后，`gee.go`简单了不少。最重要的还是通过实现了 ServeHTTP 接口，接管了所有的 HTTP 请求。相比第一天的代码，这个方法也有细微的调整，在调用 router.handle 之前，构造了一个 Context 对象。这个对象目前还非常简单，仅仅是包装了原来的两个参数，之后我们会慢慢地给Context插上翅膀。

```
func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
 c := newContext(w, req)
 engine.router.handle(c)
}
```

## 4.执行逻辑

main.go代码

```
 r := gee.New()
 r.GET("/", func(c *gee.Context) {
  c.HTML(http.StatusOK, "<h1>Hello Gee</h1>")
 })
 r.GET("/hello", func(c *gee.Context) {
  // expect /hello?name=geektutu
  c.String(http.StatusOK, "hello %s, you're at %s\n", c.Query("name"), c.Path)
 })

 r.POST("/login", func(c *gee.Context) {
  c.JSON(http.StatusOK, gee.H{
   "username": c.PostForm("username"),
   "password": c.PostForm("password"),
  })
 })

 r.Run(":9999")
```

首先在main,go中实例化engine对象，接着调用它的GET函数

GET函数会去调用addRoute函数

```
func (engine *Engine) GET(pattern string, handler HandlerFunc) {
 engine.addRoute("GET", pattern, handler)
}
```

addRoute函数会去调用router中的addRoute方法

```
func (engine *Engine) addRoute(method string, pattern string, handler HandlerFunc) {
 engine.router.addRoute(method, pattern, handler)
}
```

router.addRoute方法将会根据传参注册路由

```
func (r *router) addRoute(method string, pattern string, handler HandlerFunc) {
 log.Printf("Route %4s - %s", method, pattern)
 key := method + "-" + pattern
 r.handlers[key] = handler
}
```

GET函数的第二个参数应该为HandlerFunc类型的

但是我们传入了func(c *gee.Context)指针类型的作为参数，为什么呢？

因为我们定义了一个函数令牌，满足func(*Context)的我们就将其定义为HandlerFunc类型

因此我们再次的传参不但没有错误，更加帮助我们实现了多态

```
type HandlerFunc func(*Context)
```

我们的gee.*Context指针指向的结构体实现了很多方法

因此我们定义路由绑定的函数内部也就可以根据传入的结构体指针直接调用我们实现的结构体方法

更好的根据请求，构造响应

最后我们通过这个engine实例化对象调用.Run函数，

```
func (engine *Engine) Run(addr string) (err error) {
 return http.ListenAndServe(addr, engine)
}

func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
 c := newContext(w, req)
 engine.router.handle(c)
}
func (r *router) handle(c *Context) {
 key := c.Method + "-" + c.Path
 // 代码中的 if 语句检查给定 key 的 handler 是否存在于 r.handlers 映射中。
 // 如果存在，则 ok 变量被设置为 true，并且 handler 被分配给 handler 变量。
 // 否则，ok 变量被设置为 false，并且 handler 变量被设置为 nil。
 if handler, ok := r.handlers[key]; ok {
  handler(c)
 } else {
  c.String(http.StatusNotFound, "404 NOT FOUND: %s\n", c.Path)
 }
}
```

将所有的HTTP请求转向了我们自己的处理逻辑

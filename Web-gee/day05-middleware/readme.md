# Go语言动手写Web框架 - Gee第五天 中间件Middleware

- ## 设计并实现 Web 框架的中间件(Middlewares)机制

- ## 实现通用的`Logger`中间件，能够记录请求到响应所花费的时间，代码约50行

## 那么中间件是什么？

中间件(middlewares)，简单说，就是非业务的技术类组件。Web 框架本身不可能去理解所有的业务，因而不可能实现所有的功能。因此，框架需要有一个插口，允许用户自己定义功能，嵌入到框架中，仿佛这个功能是框架原生支持的一样。

## 中间件设计及其作用？

Gee 的中间件的定义与路由映射的 Handler 一致，处理的输入是`Context`对象。插入点是框架接收到请求初始化`Context`对象后，允许用户使用自己定义的中间件做一些额外的处理，例如记录日志等，以及对`Context`进行二次加工。另外通过调用`(*Context).Next()`函数，中间件可等待用户自己定义的 `Handler`处理结束后，再做一些额外的操作，例如计算本次处理所用时间等。即 Gee 的中间件支持用户在请求被处理的前后，做一些额外的操作。举个例子，我们希望最终能够支持如下定义的中间件，`c.Next()`表示等待执行其他的中间件或用户的`Handler`

# 1.gee.go

定义Use函数，作用让路由组使用中间件函数，即添加中间件函数到路由组的中间件函数参数中

```
func (group *RouterGroup) Use(middlewares ...HandlerFunc) {
 group.middlewares = append(group.middlewares, middlewares...)
}
```

改变ServeHTTP函数让它在根据请求去匹配响应函数之前，先根据当前请求去匹配路由组，接着获取路由组的中间件函数

传参给Context实例化对象的handlers属性，进而再去执行我们自己的http请求处理逻辑，构造响应

```
func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
 var middlewares []HandlerFunc
 for _, group := range engine.groups {
  if strings.HasPrefix(req.URL.Path, group.prefix) {
   middlewares = append(middlewares, group.middlewares...)
  }
 }
 c := newContext(w, req)
 c.handlers = middlewares
 engine.router.handle(c)
}
```

## 2.router.go

改变handler函数，让它不是直接去运行匹配到路由规则对应的函数，而是将它跟传参过来的中间件函数一起重新赋值给

Context实例化对象的handlers属性，最后同意由Next函数来控制它们的执行顺序，

以实现中间件可以在我们响应前后做处理的功能

```
func (r *router) handle(c *Context) {
 n, params := r.getRoute(c.Method, c.Path)
 if n != nil {
  c.Params = params
  key := c.Method + "-" + n.pattern
  c.handlers = append(c.handlers, r.handlers[key])
 } else {
  c.handlers = append(c.handlers, func(c *Context) {
   c.String(http.StatusNotFound, "404 NOT FOUND: %s\n", c.Path)
  })
 }
 c.Next()
}
```

## 3.context.go

Context结构体属性增加

handlers属性用来接收路由组使用的中间件函数

index属性用来控制路由组函数以及中间件函数的执行顺序

```
type Context struct {
 // origin objects
 Writer http.ResponseWriter
 Req    *http.Request
 // request info
 Path   string
 Method string
 Params map[string]string
 // response info
 StatusCode int
 // middleware
 handlers []HandlerFunc
 index    int
}
```

newContext函数的改变新增初始化index属性值为-1

```
func newContext(w http.ResponseWriter, req *http.Request) *Context {
 return &Context{
  Path:   req.URL.Path,
  Method: req.Method,
  Req:    req,
  Writer: w,
  index:  -1,
 }
}
```

新增Next方法，用来控制路由组函数以及中间件函数的执行顺序

```
func (c *Context) Next() {
 c.index++
 s := len(c.handlers)
 for ; c.index < s; c.index++ {
  c.handlers[c.index](c)
 }
}
```

## 新增中间件功能的执行逻辑

main.go代码

```
r := gee.New()
r.Use(gee.Logger()) // global midlleware
r.Run(":9999")
```

logger.go代码

```
func Logger() HandlerFunc {
 return func(c *Context) {
  // Start timer
  t := time.Now()
  // Process request
  c.Next()
  // Calculate resolution time
  log.Printf("[%d] %s in %v", c.StatusCode, c.Req.RequestURI, time.Since(t))
 }
}
```

首先实例化engine对象去调用Use方法

Use方法作用将我们定义的函数Logger添加到当前group实例化对象的中间件middlewares中

```
func (group *RouterGroup) Use(middlewares ...HandlerFunc) {
 group.middlewares = append(group.middlewares, middlewares...)
}
```

然后调用Run函数会去执行http.ListenAndServe函数，该函数的第二个参数是一个接口类型的，会自动去执行其中的

ServeHTTP方法

```
func (engine *Engine) Run(addr string) (err error) {
 return http.ListenAndServe(addr, engine)
}
```

ServeHTTP函数作用则是根据当前的请求去匹配路由组，然后将路由组的中间件函数存在middlewares中

接着根据请求实例化Context对象，并将当前请求包含的路由组的中间件函数添加到实例化Context对象的handles属性中

最后将处理好的Context实例化对象当作传参去运行engine.router.handle函数

```
func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
 var middlewares []HandlerFunc
 for _, group := range engine.groups {
  if strings.HasPrefix(req.URL.Path, group.prefix) {
   middlewares = append(middlewares, group.middlewares...)
  }
 }
 c := newContext(w, req)
 c.handlers = middlewares
 engine.router.handle(c)
}
```

engine.router.handle函数的作用是将路由匹配到的函数handlers[key]和路由组匹配到的中间件函数一起添加到Context

实例化对象的handle属性中，最后执行Next方法

```
func (r *router) handle(c *Context) {
 n, params := r.getRoute(c.Method, c.Path)

 if n != nil {
  key := c.Method + "-" + n.pattern
  c.Params = params
  c.handlers = append(c.handlers, r.handlers[key])
 } else {
  c.handlers = append(c.handlers, func(c *Context) {
   c.String(http.StatusNotFound, "404 NOT FOUND: %s\n", c.Path)
  })
 }
 c.Next()
}
```

Next函数作用是控制handles中的函数的执行顺序

两种情况：

1.中间件函数中没有Next函数，执行顺序为

先运行中间件函数，在运行路由规则匹配到的函数

2.中间件函数中有Next函数，我们先将中间件函数以Next函数为分界区分为上下两部分，

执行顺序仍然为先运行中间件函数，当运行完中间件函数上部分后，会执行Next函数，

此时在Next函数中由于我们已经执行了中间件函数，index已经增加过了那么我们再次增加后，

会去执行路由规则匹配到的handle[key]函数，执行完后会接着执行我们上次执行的中间件函数的下半部分

```
func (c *Context) Next() {
 c.index++
 s := len(c.handlers)
 for ; c.index < s; c.index++ {
  c.handlers[c.index](c)
 }
}
```

中间件函数的下半部分执行完毕后，就此响应结束

中间件完美的运行出我们想要的结果，

即能够在路由匹配到的函数的执行前后，运行我们想要运行的代码用来完成完善我们的业务

## 特殊情况当一个路由组有多个中间件函数时的执行顺序？

当有两个中间件时

假设我们应用了中间件 A 和 B，和路由映射的 Handler。`c.handlers`是这样的[A, B, Handler]，`c.index`初始化为-1。调用`c.Next()`，接下来的流程是这样的：

- c.index++，c.index 变为 0
- 0 < 3，调用 c.handlers[0]，即 A
- 执行 part1，调用 c.Next()
- c.index++，c.index 变为 1
- 1 < 3，调用 c.handlers[1]，即 B
- 执行 part3，调用 c.Next()
- c.index++，c.index 变为 2
- 2 < 3，调用 c.handlers[2]，即Handler
- Handler 调用完毕，返回到 B 中的 part4，执行 part4
- part4 执行完毕，返回到 A 中的 part2，执行 part2
- part2 执行完毕，结束。

一句话说清楚重点，最终的顺序是`part1 -> part3 -> Handler -> part 4 -> part2`。恰恰满足了我们对中间件的要求，接下来看调用部分的代码，就能全部串起来了。

```
func A(c *Context) {
    part1
    c.Next()
    part2
}
func B(c *Context) {
    part3
    c.Next()
    part4
}
```

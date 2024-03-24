# web框架-gee-day01

## 1.base1-main.go

### 要点：http.HandleFunc设置路由绑定函数

我们通过http.HandleFunc设置了2个路由，`/`和`/hello`，分别绑定 *indexHandler* 和 *helloHandler函数* ， 根据不同的HTTP请求会调用不同的处理函数。

```
http.HandleFunc("/", indexHandler)
http.HandleFunc("/hello", helloHandler)
```

## 2.base2-main.go

### 要点：将所有的HTTP请求转向了我们自己的处理逻辑

我们定义了一个空的结构体`Engine`，实现了方法`ServeHTTP`。

```
type Engine struct{}
func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
 switch req.URL.Path {
 case "/":
  fmt.Fprintf(w, "URL.Path = %q\n", req.URL.Path)
 case "/hello":
  for k, v := range req.Header {
   fmt.Fprintf(w, "Header[%q] = %q\n", k, v)
  }
 default:
  fmt.Fprintf(w, "404 NOT FOUND: %s\n", req.URL)
 }
}
```

在 *main* 函数中，我们给 *ListenAndServe* 方法的第二个参数传入了刚才创建的`engine`实例。

```
engine := new(Engine)
log.Fatal(http.ListenAndServe(":9999", engine))
```

至此，我们走出了实现Web框架的第一步，即，将所有的HTTP请求转向了我们自己的处理逻辑。

在实现`Engine`之后，我们拦截了所有的HTTP请求，拥有了统一的控制入口。在这里我们可以自由定义路由映射的规则，也可以统一添加一些处理逻辑，例如日志、异常处理等。

### 难疑点

ListenAndServe函数第二个参数是一个interface类型，里面有有一个ServeHTTP()方法，但是在我们的代码中传入了一个结构体里面有一个ServeHTTP()方法，但是并没有调用啊，怎么就执行了。结构体不是应该先赋值给接口，接口才能够调用结构体中的方法

```
http.ListenAndServe(":9999", engine)
```

### 难疑点解析

#### 1.确认这个问题在go语言语法逻辑上是没有出错的（问题本身是没有错误的）

第二个参数类型是接口类型 `http.Handler`，`Handler` 的定义博文中已经贴了，是从 `http` 的源码中找到的。

```
type Handler interface {
    ServeHTTP(w ResponseWriter, r *Request)
}

func ListenAndServe(address string, h Handler) error
```

含义：ListenAndServe函数第二个参数确实应该传入一个接口类型

#### 2.根据这个问题给出回答

在 Go 语言中，实现了接口方法的 struct 都可以强制转换为接口类型。你可以这么写：

```
handler := (http.Handler)(engine) // 手动转换为借口类型
log.Fatal(http.ListenAndServe(":9999", handler))
```

含义：由于ListenAndServe函数第二个参数确实应该传入一个接口类型，那么按照理想状态下我们就应当将engine转换为接口类型传入或者提前声明接口直接传入接口。然后，`ListenAndServe` 方法里面会去调用 `handler.ServeHTTP()` 方法，如果你感兴趣，可以在 http 的源码中找到调用的地方。

### 3.为我们的做法解释

然而，这么写是多余的，传参时，会自动进行参数转换的。所以直接传入engine 即可。

## 3.base3

### base3-gee.go

首先定义了类型`HandlerFunc`(自定义函数)，这是提供给框架用户的，用来定义路由映射的处理方法。我们在`Engine`中，添加了一张路由映射表`router`，key 由请求方法和静态路由地址构成，例如`GET-/`、`GET-/hello`、`POST-/hello`，这样针对相同的路由，如果请求方法不同,可以映射不同的处理方法(Handler)，value 是用户映射的处理方法。

```
type HandlerFunc func(http.ResponseWriter, *http.Request)

// Engine implement the interface of ServeHTTP
type Engine struct {
 router map[string]HandlerFunc
}
```

当用户调用`(*Engine).GET()`方法时，会将路由和处理方法注册到映射表 *router* 中，`(*Engine).Run()`方法，是 *ListenAndServe* 的包装。

```
func (engine *Engine) GET(pattern string, handler HandlerFunc) {
 engine.addRoute("GET", pattern, handler)
}
```

Engine`实现的 *ServeHTTP* 方法的作用就是，解析请求的路径，查找路由映射表，如果查到，就执行注册的处理方法。如果查不到，就返回 *404 NOT FOUND* 。

```
func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
 key := req.Method + "-" + req.URL.Path
 if handler, ok := engine.router[key]; ok {
  handler(w, req)
 } else {
  fmt.Fprintf(w, "404 NOT FOUND: %s\n", req.URL)
 }
}
```

### base3-main.go

使用`New()`创建 gee 的实例，使用 `GET()`方法添加路由，最后使用`Run()`启动Web服务。这里的路由，只是静态路由，不支持`/hello/:name`这样的动态路由，动态路由我们将在下一次实现。

```
r := gee.New()
r.GET("/", func(w http.ResponseWriter, req *http.Request) {
  fmt.Fprintf(w, "URL.Path = %q\n", req.URL.Path)
})

r.GET("/hello", func(w http.ResponseWriter, req *http.Request) {
  for k, v := range req.Header {
   fmt.Fprintf(w, "Header[%q] = %q\n", k, v)
  }
})

r.Run(":9999")
```

## 4.当前gee简易web框架使用流程分析

首先在main.go中，实例化对象

```
r := gee.New()
```

去调用gee包中的New方法，创建并返回一个新的Engine实例。

```
func New() *Engine {
 return &Engine{router: make(map[string]HandlerFunc)}
}
```

接着在main.go中通过r.GET()去调用gee包中的GET函数

```
func (engine *Engine) GET(pattern string, handler HandlerFunc) {
 engine.addRoute("GET", pattern, handler)
}
```

GET函数，会通过当前实例化对象去调用addRoute()函数，将请求方法、请求路径和处理函数注册到路由表中。

```
func (engine *Engine) addRoute(method string, pattern string, handler HandlerFunc) {
 key := method + "-" + pattern
 engine.router[key] = handler
}
```

最后在main.go中通过r.Run()去调用gee包中的Run方法

```
func (engine *Engine) Run(addr string) (err error) {
 return http.ListenAndServe(addr, engine)
}
```

Run函数会将当前实例化对象作为第二个参数传入http.ListenAndServe，r.Run()中的参数作为第一个参数传入。接着engine会自动转换为接口类型，并且去执行其中的ServeHTTP()方法

```
func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
 key := req.Method + "-" + req.URL.Path
 if handler, ok := engine.router[key]; ok {
  handler(w, req)
 } else {
  http.Error(w, "404 not found", http.StatusNotFound)
 }
}
```

ServeHTTP函数是一个HTTP请求处理函数，它属于`engine`类型。当有HTTP请求到来时，该函数会根据请求的方法和URL路径生成一个键值`key`，然后在`engine.router`中查找对应的处理程序。如果找到了处理程序，则调用它来处理请求；如果没有找到，则返回一个404状态码的错误信息。

至此，整个`Gee`框架的原型已经出来了。实现了路由映射表，提供了用户注册静态路由的方法，包装了启动服务的函数。

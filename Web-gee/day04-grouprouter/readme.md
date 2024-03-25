# Go语言动手写Web框架 - Gee第四天 分组控制Group

## 实现路由分组控制

分组控制(Group Control)是 Web 框架应提供的基础功能之一。所谓分组，是指路由的分组。如果没有路由分组，我们需要针对每一个路由进行控制。但是真实的业务场景中，往往某一组路由需要相似的处理。例如：

- 以`/post`开头的路由匿名可访问。
- 以`/admin`开头的路由需要鉴权。
- 以`/api`开头的路由是 RESTful 接口，可以对接第三方平台，需要三方平台鉴权。

大部分情况下的路由分组，是以相同的前缀来区分的。因此，我们今天实现的分组控制也是以前缀来区分，并且支持分组的嵌套。例如`/post`是一个分组，`/post/a`和`/post/b`可以是该分组下的子分组。作用在`/post`分组上的中间件(middleware)，也都会作用在子分组，子分组还可以应用自己特有的中间件。

## 1.gee.go

### 嵌套结构体Engine和RouterGroup

那么一个group都需要什么属性呢

```
RouterGroup struct {
	prefix      string        // 分组名
	middlewares []HandlerFunc // 当前分组支持的中间件
	engine      *Engine       // 所有的分组共享一个engine接口
}
```

engine的进一步抽象，将`Engine`作为最顶层的分组，也就是说`Engine`拥有`RouterGroup`所有的能力。

```
Engine struct {
	*RouterGroup          //结构体之间的匿名嵌套，可以直接调用其中的方法 例如r.group
	router *router        //路由
	groups []*RouterGroup //保存全部的路由组
}
```

我们在这两个结构体之间使用了go语言嵌套的用法，那我们就可以将和路由有关的函数，都交给`RouterGroup`实现了。

这样做的好处：

实例化engine对象后，我们可以用这个对象去添加路由等；   

 r.addRoute  仍然是被允许的，因为匿名嵌套结构体

我们也可以先用这个实例化对象去调用Group方法生成分组，后再添加路由等；

group 实现了addRoute等方法

### 对于Engine的改变和RouterGroup的声明一些重要方法的定义和改变

New函数的改变，新增两个属性的初始化

```
func New() *Engine {
	engine := &Engine{router: newRouter()}
	engine.RouterGroup = &RouterGroup{engine: engine}
	engine.groups = []*RouterGroup{engine.RouterGroup}
	return engine
}

func (group *RouterGroup) addRoute(method string, comp string, handler HandlerFunc) {
	pattern := group.prefix + comp
	log.Printf("Route %4s - %s", method, pattern)
	group.engine.router.addRoute(method, pattern, handler)
}
```

Group方法的声明，用于创建一个新的路由组

该方法的两种具体用法

1.实例化对象engine调用时，创建一个路由组

2.group调用时，给当前分组创建一个子分组

```
// Group is defined to create a new RouterGroup
// remember all groups share the same Engine instance
func (group *RouterGroup) Group(prefix string) *RouterGroup {
	engine := group.engine
	newGroup := &RouterGroup{
		prefix: group.prefix + prefix,
		parent: group,
		engine: engine,
	}
	engine.groups = append(engine.groups, newGroup)
	return newGroup
}
```

addRoute函数的改变，pattern由拼接而成

两种情况：

1.实例化对象engine调用时，pattern = nil + comp 

2.group调用时，拼接组名和路由规则作为新路由规则去注册路由

```
func (group *RouterGroup) addRoute(method string, comp string, handler HandlerFunc) {
	pattern := group.prefix + comp
	log.Printf("Route %4s - %s", method, pattern)
	group.engine.router.addRoute(method, pattern, handler)
}
```

至此路由分组功能完成

## 2.执行逻辑

main.go代码

```
r := gee.New()
	r.GET("/index", func(c *gee.Context) {
		c.HTML(http.StatusOK, "<h1>Index Page</h1>")
	})
	v1 := r.Group("/v1")
	{
		v1.GET("/", func(c *gee.Context) {
			c.HTML(http.StatusOK, "<h1>Hello Gee</h1>")
		})

		v1.GET("/hello", func(c *gee.Context) {
			// expect /hello?name=geektutu
			c.String(http.StatusOK, "hello %s, you're at %s\n", c.Query("name"), c.Path)
		})
	}
	v2 := r.Group("/v2")
	{
		v2.GET("/hello/:name", func(c *gee.Context) {
			// expect /hello/geektutu
			c.String(http.StatusOK, "hello %s, you're at %s\n", c.Param("name"), c.Path)
		})
		v2.POST("/login", func(c *gee.Context) {
			c.JSON(http.StatusOK, gee.H{
				"username": c.PostForm("username"),
				"password": c.PostForm("password"),
			})
		})

	}

	r.Run(":9999")
```

去其中的一部分来做示例，进行流程分析

```
v1 := r.Group("/v1")

v1.GET("/", func(c *gee.Context) {
	c.HTML(http.StatusOK, "<h1>Hello Gee</h1>")
})
```

首先实例化engine对象调用Group方法，创建了一个路由组并返回给参数v1

```
v1 := r.Group("/v1")
```

然后v1这个group去调用GET方法去注册路由

```
func (group *RouterGroup) GET(pattern string, handler HandlerFunc) {
	group.addRoute("GET", pattern, handler)
}
```

group.addRoute拼接组名和路由规则生成新路由规则pattern后去调用group.engine.router.addRoute注册路由

```
func (group *RouterGroup) addRoute(method string, comp string, handler HandlerFunc) {
	pattern := group.prefix + comp
	log.Printf("Route %4s - %s", method, pattern)
	group.engine.router.addRoute(method, pattern, handler)
}
```

group.engine.router.addRoute函数注册路由成功

```
func (r *router) addRoute(method string, pattern string, handler HandlerFunc) {
	parts := parsePattern(pattern)

	key := method + "-" + pattern
	_, ok := r.roots[method]
	if !ok {
		r.roots[method] = &node{}
	}
	r.roots[method].insert(pattern, parts, 0)
	r.handlers[key] = handler
}
```

至此路由组和路由组中的路由注册完毕

后续根据请求，返回响应，与普通路由无异

### 路由组和其中的路由到底是啥意思了？

首先我们知道我们是根据路由组名和其中的路由共同注册的路由，所以我们在请求路由组中的路由时应该如此做

http://localhost:9999/路由组名/其中的路由

才能得到我们想要的响应

### 路由组对我们的用处是啥呢？

便于我们统一处理，在后续中间件的使用中非常关键，可谓一个致命大杀器，太它🐎好用了
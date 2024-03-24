# Go语言动手写Web框架 - Gee第三天 前缀树路由Router

## 使用trie树实现动态路由解析

## 支持两种模式：name 和 *filepath

## 1.trie.go

### 动态路由，即一条路由规则可以匹配某一类型而非某一条固定的路由。例如`/hello/:name`，可以匹配`/hello/geektutu`、`hello/jack`等

### 参数匹配`:`。例如 `/p/:lang/doc`，可以匹配 `/p/c/doc` 和 `/p/go/doc`

### 通配`*`。例如 `/static/*filepath`，可以匹配`/static/fav.ico`，也可以匹配`/static/js/jQuery.js`，这种模式常用于静态服务器，能够递归地匹配子路径

节点的定义

一个节点应该包含那些内容呢

```
type node struct {
 pattern  string // 路由规则，例如 “/users/：id”  pattern /users/:id
 part     string // 匹配到的具体部分，例如 "/users/123"时，":id"对应的部分时"123"，part 123
 children []*node // 子节点,用于存储更深层次的路由规则  /users/:id/username  /users/:id/password
 isWild   bool // 是否精确匹配，part 含有 : 或 * 时为true
}
```

如何注册一个路由规则

matchChild函数用于在节点的子节点中查找匹配指定部分的节点。

查找成功返回子节点，否则返回nil用于后续增加子节点给当前节点

insert函数的作用是在路由树中插入一个新的模式，并根据模式的结构将它分解并插入到树的相应位置。

```
func (n *node) matchChild(part string) *node {
 for _, child := range n.children {
  if child.part == part || child.isWild {
   return child
  }
 }
 return nil
}
func (n *node) insert(pattern string, parts []string, height int) {
 if len(parts) == height {
  n.pattern = pattern
  return
 }

 part := parts[height]
 child := n.matchChild(part)
 if child == nil {
  child = &node{part: part, isWild: part[0] == ':' || part[0] == '*'}
  n.children = append(n.children, child)
 }
 child.insert(pattern, parts, height+1)
}
```

如何匹配一个路由规则

matchChildren函数用于匹配给定部分字符串的子节点。

匹配成功返回添加完子节点的节点，用于下一层匹配，否则返回空指针

search函数是一个递归搜索函数，用于在树形结构中根据给定的路径parts查找匹配的节点。

递归匹配成功返回非空原指针n，匹配失败返回nil

```
func (n *node) matchChildren(part string) []*node {
 nodes := make([]*node, 0)
 for _, child := range n.children {
  if child.part == part || child.isWild {
   nodes = append(nodes, child)
  }
 }
 return nodes
}
func (n *node) search(parts []string, height int) *node {
 if len(parts) == height || strings.HasPrefix(n.part, "*") {
  if n.pattern == "" {
   return nil
  }
  return n
 }

 part := parts[height]
 children := n.matchChildren(part)

 for _, child := range children {
  result := child.search(parts, height+1)
  if result != nil {
   return result
  }
 }

 return nil
}
```

## 2.router.go

### Trie 树的插入与查找都成功实现了，接下来我们将 Trie 树应用到路由中去吧

我们使用 roots 来存储每种请求方式的Trie 树根节点。使用 handlers 存储每种请求方式的 HandlerFunc 。

```
type router struct {
 roots    map[string]*node
 handlers map[string]HandlerFunc
}
```

存储路由，即存储每种请求方式的Trie树根节点和存储每种请求方式的HandlerFunc

addRoute函数为一个路由器对象添加一条路由记录，包括请求方法、请求路径和处理函数。

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

parsePattern函数在addRoute函数中的用法，用于将我们输入的路由规则解析为parts []string切片

便于后续我们使用Trie树存储节点

parsePattern函数本身用于解析一个字符串pattern（路由规则），并将其按照"/"分割成多个部分，然后将非空的部分存入一个字符串切片中并返回。如果某个部分以"*"开头，则会提前结束循环并返回结果。

```
func parsePattern(pattern string) []string {
 vs := strings.Split(pattern, "/")

 parts := make([]string, 0)
 for _, item := range vs {
  if item != "" {
   parts = append(parts, item)
   if item[0] == '*' {
    break
   }
  }
 }
 return parts
}
```

匹配路由，首先匹配根节点，然后向下匹配

getRoute函数用于根据HTTP方法和路径获取匹配的路由参数和节点。

匹配路由成功返回节点和存储有路径参数的参数映射。匹配失败则返回nil和nil。

```
func (r *router) getRoute(method string, path string) (*node, map[string]string) {
 searchParts := parsePattern(path)
 params := make(map[string]string)
 root, ok := r.roots[method]

 if !ok {
  return nil, nil
 }

 n := root.search(searchParts, 0)

 if n != nil {
  parts := parsePattern(n.pattern)
  for index, part := range parts {
   if part[0] == ':' {
    params[part[1:]] = searchParts[index]
   }
   if part[0] == '*' && len(part) > 1 {
    params[part[1:]] = strings.Join(searchParts[index:], "/")
    break
   }
  }
  return n, params
 }

 return nil, nil
}
```

handle函数，在调用匹配到的`handler`前，将解析出来的路由参数赋值给了`c.Params`。这样就能够在`handler`中，通过`Context`对象访问到具体的值和使用Context对象的方法了。

```
func (r *router) handle(c *Context) {
 n, params := r.getRoute(c.Method, c.Path)
 if n != nil {
  c.Params = params
  key := c.Method + "-" + n.pattern
  r.handlers[key](c)
 } else {
  c.String(http.StatusNotFound, "404 NOT FOUND: %s\n", c.Path)
 }
}
```

## 3.执行流程

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

 r.GET("/hello/:name", func(c *gee.Context) {
  // expect /hello/geektutu
  c.String(http.StatusOK, "hello %s, you're at %s\n", c.Param("name"), c.Path)
 })

 r.GET("/assets/*filepath", func(c *gee.Context) {
  c.JSON(http.StatusOK, gee.H{"filepath": c.Param("filepath")})
 })

 r.Run(":9999")
```

以以下代码的执行流程为例

```
r.GET("/hello/:name", func(c *gee.Context) {
  // expect /hello/geektutu
  c.String(http.StatusOK, "hello %s, you're at %s\n", c.Param("name"), c.Path)
})
```

实例化engine对象调用GET方法

```
func (engine *Engine) GET(pattern string, handler HandlerFunc) {
 engine.addRoute("GET", pattern, handler)
}
```

GET方法会去调用实例化对象的addRoute函数

```
func (engine *Engine) addRoute(method string, pattern string, handler HandlerFunc) {
 log.Printf("Route %4s - %s", method, pattern)
 engine.router.addRoute(method, pattern, handler)
}
```

实例化对象的addRoute函数会调用router下的addRoute函数注册路由

路由注册完毕

开启监听，并将所有的HTTP请求转向了我们自己的处理逻辑

```
r.Run(":9999")
```

当我们在路由器发送请求如下时

```
curl "http://localhost:9999/hello/geektutu"
```

当服务器监听到浏览器的请求时，构造响应

根据请求和响应实例化Context，然后去调用router下的handle函数

```
func (engine *Engine) Run(addr string) (err error) {
 return http.ListenAndServe(addr, engine)
}

func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
 c := newContext(w, req)
 engine.router.handle(c)
}
```

router下的handle函数

首先会去匹配（动态）路由，匹配成功返回匹配成功的路由规则的节点指针和存储有路径参数的参数映射

```
func (r *router) handle(c *Context) {
 n, params := r.getRoute(c.Method, c.Path)
 if n != nil {
  c.Params = params
  key := c.Method + "-" + n.pattern
  r.handlers[key](c)
 } else {
  c.String(http.StatusNotFound, "404 NOT FOUND: %s\n", c.Path)
 }
}
```

然后先将存储有路径参数的参数映射赋值给Context实例化对象的Params，

这样后续我们在绑定的执行函数中才能c.Param("name")获取对应的值

至此动态路由的参数我们也能够获取到了，动态路由彻底完善

接着在根据结点指针获取key值，去执行对应路由规则绑定的函数

至此响应结束

## 动态路由彻底完善是什么意思？

1.我们根据trie树完成了动态路由的注册和匹配

2.我们能够获取动态路由中的信息

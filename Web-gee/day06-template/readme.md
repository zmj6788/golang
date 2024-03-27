# Go语言动手写Web框架 - Gee第六天 模板(HTML Template)

- ## 实现静态资源服务(Static Resource)。请求获取服务器本地的静态文件内容

- ## 支持HTML模板渲染。服务端渲染本地HTML模板返回给客户端

## 今天的内容便是介绍 Web 框架如何支持服务端渲染的场景

要做到服务端渲染，第一步便是要支持 JS、CSS 等静态文件。

我们之前设计动态路由的时候，支持通配符`*`匹配多级子路径。比如路由规则`/assets/*filepath`，可以匹配`/assets/`开头的所有的地址。例如`/assets/js/geektutu.js`，匹配后，参数`filepath`就赋值为`js/geektutu.js`。

那如果我么将所有的静态文件放在`static`目录下，那么`filepath`的值即是该目录下文件的相对地址。映射到真实的文件后，将文件返回，静态服务器就实现了。

找到文件后，如何返回这一步，`net/http`库已经实现了。因此，gee 框架要做的，仅仅是解析请求的地址，映射到服务器上文件的真实地址，交给`http.FileServer`处理就好了。

## 静态资源服务

## 1.gee.go

createStaticHandler函数是一个生成静态文件处理程序的方法

该程序内部功能包括：

1.根据获得的文件路径打开文件，查看文件是否存在

2.通过http.FileServer将文件返回

```
func (group *RouterGroup) createStaticHandler(relativePath string, fs http.FileSystem) HandlerFunc {
 absolutePath := path.Join(group.prefix, relativePath)
 fileServer := http.StripPrefix(absolutePath, http.FileServer(fs))
 return func(c *Context) {
  file := c.Param("filepath")
  // Check if file exists and/or if we have permission to access it
  if _, err := fs.Open(file); err != nil {
   c.Status(http.StatusNotFound)
   return
  }

  fileServer.ServeHTTP(c.Writer, c.Req)
 }
}
```

Static方法将静态文件处理程序注册到路由中，使得可以通过HTTP请求访问指定根目录下的静态文件

```
func (group *RouterGroup) Static(relativePath string, root string) {
 handler := group.createStaticHandler(relativePath, http.Dir(root))
 urlPattern := path.Join(relativePath, "/*filepath")
 // Register GET handlers
 group.GET(urlPattern, handler)
}
```

至此我们实现了获取服务器本地的静态资源

## HTML模板渲染

## 1.gee.go

给engine添加新属性

htmlTemplates 用来存储HTML模板

funcMap 用来添加自定义函数映射，后续将添加至HTML模板中帮助我们处理模板

```
Engine struct {
 *RouterGroup
 router        *router
 groups        []*RouterGroup     // store all groups
 htmlTemplates *template.Template // for html render
 funcMap       template.FuncMap   // for html render
}
```

SetFuncMap函数添加自定义函数映射给engine

```
func (engine *Engine) SetFuncMap(funcMap template.FuncMap) {
 engine.funcMap = funcMap
}
```

LoadHTMLGlob函数根据指定的模式pattern加载并存储所有HTML模板在engine的htmlTemplates中

```
func (engine *Engine) LoadHTMLGlob(pattern string) {
 engine.htmlTemplates = template.Must(template.New("").Funcs(engine.funcMap).ParseGlob(pattern))
}
```

至此我们能够构建含有我们自定义模板处理函数的HTML模板

接下来只需要渲染HTML模板

2.context.go

在Context结构体中添加engine指针用于传递加载好的HTML模板

```
type Context struct {
    // ...
 // engine pointer
 engine *Engine
}
```

HTML函数根据我们提供的name在模板库中查找，存在则渲染，否则结束HTTP请求，并返回错误状态码及错误信息。

```
func (c *Context) HTML(code int, name string, data interface{}) {
 c.SetHeader("Content-Type", "text/html")
 c.Status(code)
 if err := c.engine.htmlTemplates.ExecuteTemplate(c.Writer, name, data); err != nil {
  c.Fail(500, err.Error())
 }
}
```

## 3.gee.go

在路由匹配函数执行前将engine指针传递给Context的engine属性

便于后续我们在HTML方法中调用已经加载好的HTML模板

为什么要在路由匹配函数调用之前？

因为HTML()方法通常在路由调用函数内部使用

```
func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
 // ...
 c := newContext(w, req)
 c.handlers = middlewares
 c.engine = engine
 engine.router.handle(c)
}
```

至此HTML模板渲染功能实现

## 执行流程

最终的目录结构

```

---gee/
---static/
   |---css/
        |---geektutu.css
   |---file1.txt
---templates/
   |---arr.tmpl
   |---css.tmpl
   |---custom_func.tmpl
---main.go
```

main.go代码示例

```
type student struct {
 Name string
 Age  int8
}
//自定义模板处理函数在模板中使用
/*
<!-- templates/arr.tmpl -->
<html>
<body>
    <p>hello, {{.title}}</p>
    <p>Date: {{.now | FormatAsDate}}</p>
</body>
</html>
*/
func FormatAsDate(t time.Time) string {
 year, month, day := t.Date()
 return fmt.Sprintf("%d-%02d-%02d", year, month, day)
}

func main() {
 r := gee.New()
 r.Use(gee.Logger())
 //将自定义模板函数添加到映射中
 r.SetFuncMap(template.FuncMap{
  "FormatAsDate": FormatAsDate,
 })
 //加载HTML模板以及映射函数
 r.LoadHTMLGlob("templates/*")、
 //绑定路由与本地文件夹开启文件处理服务器
 r.Static("/assets", "./static")

 stu1 := &student{Name: "Geektutu", Age: 20}
 stu2 := &student{Name: "Jack", Age: 22}
 r.GET("/", func(c *gee.Context) {
  c.HTML(http.StatusOK, "css.tmpl", nil)
 })
 r.GET("/students", func(c *gee.Context) {
  //根据传入name去渲染HTML模板
  c.HTML(http.StatusOK, "arr.tmpl", gee.H{
   "title":  "gee",
   "stuArr": [2]*student{stu1, stu2},
  })
 })

 r.GET("/date", func(c *gee.Context) {
  c.HTML(http.StatusOK, "custom_func.tmpl", gee.H{
   "title": "gee",
   "now":   time.Date(2019, 8, 17, 0, 0, 0, 0, time.UTC),
  })
 })

 r.Run(":9999")
}
```

各种请求结果

```
Z:\Work\Program\go\src\project\Web-gee>curl http://localhost:9999/assets/css/geektutu.css
p {
    color: orange;
    font-weight: 700;
    font-size: 20px;
}

Z:\Work\Program\go\src\project\Web-gee>curl http://localhost:9999/                       
<html>
    <link rel="stylesheet" href="/assets/css/geektutu.css">
    <p>geektutu.css is loaded</p>
</html>

Z:\Work\Program\go\src\project\Web-gee>curl http://localhost:9999/students
<!-- templates/arr.tmpl -->
<html>
<body>
    <p>hello, gee</p>
    
    <p>0: Geektutu is 20 years old</p>
    
    <p>1: Jack is 22 years old</p>
    
</body>
</html>

Z:\Work\Program\go\src\project\Web-gee>curl http://localhost:9999/date
<!-- templates/arr.tmpl -->
<html>
<body>
    <p>hello, gee</p>
    <p>Date: 2019-08-17</p>
</body>
</html>
```

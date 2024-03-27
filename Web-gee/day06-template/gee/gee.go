package gee

import (
	"log"
	"net/http"
	"path"
	"strings"
	"text/template"
)

// HandlerFunc defines the request handler used by gee
type HandlerFunc func(*Context)

type RouterGroup struct {
	prefix      string
	middlewares []HandlerFunc
	engine      *Engine
}

// Engine implement the interface of ServeHTTP
type Engine struct {
	// 匿名字段可以直接引用该字段的方法(也可以指定字段名引用字段的方法)
	*RouterGroup
	router *router
	groups []*RouterGroup
	// 用于存储html模板
	htmlTemplates *template.Template
	// 用于存储自定义函数映射
	funcMap template.FuncMap
}

// 添加自定义函数映射给engine
func (engine *Engine) SetFuncMap(funcMap template.FuncMap) {
	engine.funcMap = funcMap
}

// 根据指定的模式pattern加载并存储所有HTML模板在engine的htmlTemplates中
func (engine *Engine) LoadHTMLGlob(pattern string) {
	// template.New("")创建一个新的模板对象，模板名称为空
	// .Funcs(engine.funcMap)将指定模板函数映射添加到模板对象中
	// .ParseGlob(pattern)根据指定的模式pattern
	// 加载所有的HTML模板文件，并返回模板对象
	// template.Must()将上述模板对象转换为*template.Template类型，并返回
	engine.htmlTemplates = template.Must(template.New("").Funcs(engine.funcMap).ParseGlob(pattern))
}

// New is the constructor of gee.Engine
func New() *Engine {
	engine := &Engine{router: newRouter()}
	engine.RouterGroup = &RouterGroup{engine: engine}
	engine.groups = []*RouterGroup{engine.RouterGroup}
	return engine
}

func (group *RouterGroup) Group(prefix string) *RouterGroup {
	engine := group.engine
	newGroup := &RouterGroup{
		prefix: group.prefix + prefix,
		engine: engine,
	}
	engine.groups = append(engine.groups, newGroup)
	return newGroup
}

func (group *RouterGroup) Use(middlewares ...HandlerFunc) {
	group.middlewares = append(group.middlewares, middlewares...)
}

// 该函数是一个生成静态文件处理程序的方法
func (group *RouterGroup) createStaticHandler(relativePath string, fs http.FileSystem) HandlerFunc {
	absolutePath := path.Join(group.prefix, relativePath)
	fileServer := http.StripPrefix(absolutePath, http.FileServer(fs))
	return func(c *Context) {
		file := c.Param("filepath")
		if _, err := fs.Open(file); err != nil {
			c.Status(http.StatusNotFound)
			return
		}

		fileServer.ServeHTTP(c.Writer, c.Req)
	}
}

// 将静态文件处理程序注册到路由中
// 使得可以通过HTTP请求访问指定根目录下的静态文件
func (group *RouterGroup) Static(relativePath string, root string) {
	// http.Dir(root) 创建了一个指向实际磁盘目录的虚拟文件系统对象，
	// 这样当客户端发起对应静态资源的请求时，
	// 服务器可以从这个指定的根目录 root 中查找并返回相应的文件内容。
	handler := group.createStaticHandler(relativePath, http.Dir(root))
	urlPattern := path.Join(relativePath, "/*filepath")
	group.GET(urlPattern, handler)
}

func (group *RouterGroup) addRoute(method string, comp string, handler HandlerFunc) {
	pattern := group.prefix + comp
	log.Printf("Route %4s - %s", method, pattern)
	group.engine.router.addRoute(method, pattern, handler)
}

// GET defines the method to add GET request
func (group *RouterGroup) GET(pattern string, handler HandlerFunc) {
	group.addRoute("GET", pattern, handler)
}

// POST defines the method to add POST request
func (group *RouterGroup) POST(pattern string, handler HandlerFunc) {
	group.addRoute("POST", pattern, handler)
}

// Run defines the method to start a http server
func (engine *Engine) Run(addr string) (err error) {
	return http.ListenAndServe(addr, engine)
}

func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var middlewares []HandlerFunc
	for _, group := range engine.groups {
		// 筛选出当前请求的路由组，将它的正在使用的中间件加入到middlewares参数中
		if strings.HasPrefix(req.URL.Path, group.prefix) {
			middlewares = append(middlewares, group.middlewares...)
		}
	}
	c := newContext(w, req)
	// 将正在使用的路由组的中间件传入到c.handlers中
	c.handlers = middlewares
	// 实例化 Context 时，还需要给 c.engine 赋值。
	// 这样就能够通过 Context 访问 Engine 中的 HTML 模板。
	c.engine = engine
	engine.router.handle(c)
}

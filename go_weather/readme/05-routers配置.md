# gvb_server后端项目编写流程

## 实现功能：为初始化路由添加路由，即添加接口访问路径和接口响应函数

## 5.routers配置

api目录下新建settings_api目录

settings_api目录下创建enter.go和settings_info.go文件

settings_info.go中定义对应settingsinfo路由的接口函数SettingsInfoView

```
func (SettingsApi) SettingsInfoView(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "pong",
	})
}
```

enter.go目的，实例化后可以调用所有settings相关路由接口函数

```
type SettingsApi struct {
}
```

api目录下新建enter.go

enter.go中定义结构体ApiGroup，将所有路由接口函数的调用结构体嵌套在其中

目的：通过全局变量实例化相应对象后，调用路由函数

```
type ApiGroup struct {
	SettingsApi settings_api.SettingsApi
}

var ApiGroupApp = new(ApiGroup)
```

routers目录下修改enter.go新增settings_router.go

settings_router.go

SettingsRouter方法添加settings相关路由进路由组中

settingsApi实例化，用于调用路由函数

```
func SettingsRouter(router *gin.RouterGroup) {
	settingsApi := api.ApiGroupApp.SettingsApi
	router.GET("/settingsinfo", settingsApi.SettingsInfoView)
}
```

enter.go

新增路由组，将settings相关路由添加到路由组中

```
routerGroup := router.Group("/api")
SettingsRouter(routerGroup)
```


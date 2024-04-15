package main

import (
	"gvb_server/core"
	"gvb_server/global"
	"gvb_server/models"
	"gvb_server/routers"
)

func main() {
	// 配置信息读取
	core.InitConf()
	//日志初始化
	global.Log = core.InitLogger()
	//数据库连接
	global.DB = core.Initgorm()
	global.DB.AutoMigrate(&models.User{})
	//路由初始化
	router := routers.InitRouter()
	//启动服务
	addr := global.Config.System.Addr()
	global.Log.Infof("gvb_server运行在: %s", addr)
	router.Run(addr)
}

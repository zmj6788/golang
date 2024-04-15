package routers

import (
	"gvb_server/global"
	"gvb_server/models"

	"github.com/gin-gonic/gin"
)

func InitRouter() *gin.Engine {
	gin.SetMode(global.Config.System.Env)
	router := gin.Default()
	//在这里读取json错误码文件
	models.InitCityListJSon()
	models.IninWeatherData()
	models.IninWeatherCityData()
	routerGroup := router.Group("/api")
	WeatherRouter(routerGroup)
	UserRouter(routerGroup)
	routerGroup_citylist := router.Group("/jwt")
	CityRouter(routerGroup_citylist)
	return router
}

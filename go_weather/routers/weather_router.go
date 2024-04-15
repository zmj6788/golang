package routers

import (
	"gvb_server/api"

	"github.com/gin-gonic/gin"
)

func WeatherRouter(router *gin.RouterGroup) {
	weatherApi := api.ApiGroupApp.WeatherApi
	//'http://hmajax.itheima.net/api/weather?city=110100'
	router.GET("/weather", weatherApi.WeatherInfoView)
	router.GET("/weathercity", weatherApi.WeatherCityInfoView)
}

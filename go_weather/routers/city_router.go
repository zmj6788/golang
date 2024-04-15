package routers

import (
	"gvb_server/api"
	"gvb_server/middleware"

	"github.com/gin-gonic/gin"
)

func CityRouter(router *gin.RouterGroup) {
	cityApi := api.ApiGroupApp.CityApi
	router.Use(middleware.JWTMiddleware()).GET("/weather/citylist", cityApi.CityListInfoView)
}

package routers

import (
	"gvb_server/api"

	"github.com/gin-gonic/gin"
)

func UserRouter(router *gin.RouterGroup) {
	userApi := api.ApiGroupApp.UserApi
	router.POST("/login", userApi.Login)
	router.POST("/register", userApi.Register)
}

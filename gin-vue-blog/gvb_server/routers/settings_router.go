package routers

import (
	"gvb_server/api"

	"github.com/gin-gonic/gin"
)

func SettingsRouter(router *gin.RouterGroup) {
	settingsApi := api.ApiGroupApp.SettingsApi
	router.GET("/settingsinfo", settingsApi.SettingsInfoView)
}

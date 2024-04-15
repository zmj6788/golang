package weather_api

import (
	"gvb_server/models"
	"gvb_server/models/res"

	"github.com/gin-gonic/gin"
)

func (WeatherApi) WeatherCityInfoView(c *gin.Context) {
	res.Ok(models.WeatherCityData, "获取成功", c)
}

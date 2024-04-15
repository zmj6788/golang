package city_api

import (
	"gvb_server/models"
	"gvb_server/models/res"

	"github.com/gin-gonic/gin"
)

func (CityApi) CityListInfoView(c *gin.Context) {
	res.Ok(models.Citys, "获取城市列表成功", c)
}

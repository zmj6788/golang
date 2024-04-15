package global

import (
	"gvb_server/config"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// 配置信息存储全局变量，便于全局使用
var (
	Config *config.Config
	DB     *gorm.DB
	Log    *logrus.Logger
)

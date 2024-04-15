package core

import (
	"fmt"
	"gvb_server/global"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Initgorm() *gorm.DB {

	if global.Config.MySQL.Host == "" {
		global.Log.Warn("未配置mysql，取消gorm连接数据库")
	}

	var mysqlLogger logger.Interface
	if global.Config.System.Env == "debug" {
		//开发环境显示所有的sql日志
		mysqlLogger = logger.Default.LogMode(logger.Info)
	} else {
		//只打印错误的sql日志
		mysqlLogger = logger.Default.LogMode(logger.Error)
	}

	dsn := global.Config.MySQL.Dsn()

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: mysqlLogger,
	})
	if err != nil {
		global.Log.Fatalf(fmt.Sprintf("[%s] 连接数据库失败", dsn))
	}

	sqlDB, _ := db.DB()

	sqlDB.SetMaxOpenConns(100)              //最多可容纳
	sqlDB.SetMaxIdleConns(20)               //最大空闲连接数
	sqlDB.SetConnMaxLifetime(time.Hour * 4) //连接最大复用时间，不能超过mysql的wait_timeout
	global.Log.Info("连接数据库成功")
	return db
}

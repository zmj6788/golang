package user_api

import (
	"errors"
	"gvb_server/global"
	"gvb_server/middleware"
	"gvb_server/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func (UserApi) Login(c *gin.Context) {
	type LoginRequest struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	var loginReq LoginRequest
	if err := c.ShouldBindJSON(&loginReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "用户名或密码不能为空"})
		return
	}

	var user models.User
	if err := global.DB.Where("username = ?", loginReq.Username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusUnauthorized, gin.H{"msg": "用户名不存在"})
			return
		}
		//对于除记录未找到以外的其他数据库错误，输出错误日志
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			logrus.Error("数据库操作失败: ", err)
		}
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "数据库错误"})
		return
	}

	// 假设密码已正确哈希存储，此处使用bcrypt进行比较
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginReq.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"msg": "密码错误"})
		return
	}
	// 生成JWT
	var claims = middleware.CustomClaims{Username: user.Username}
	token, err := middleware.GenerateToken(claims)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "无法生成JWT"})
		logrus.Error("生成JWT失败: ", err)
		return
	}

	// 返回JWT
	c.JSON(http.StatusOK, gin.H{
		"msg":      "登录成功",
		"token":    token,
		"username": user.Username,
	})
}

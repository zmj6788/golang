package user_api

import (
	"gvb_server/global"
	"gvb_server/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

func (UserApi) Register(c *gin.Context) {
	type RegisterRequest struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	var regReq RegisterRequest
	if err := c.ShouldBindJSON(&regReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": -1,
			"msg":  "用户名或密码不能为空",
		})
		return
	}

	if exists := checkIfUsernameExists(regReq.Username); exists {
		c.JSON(http.StatusConflict, gin.H{
			"code": -1,
			"msg":  "该用户名已被注册",
		})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(regReq.Password), bcrypt.DefaultCost)
	if err != nil {
		logrus.Errorf("密码加密失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": -1,
			"msg":  "服务器内部错误",
		})
		return
	}

	user := models.User{Username: regReq.Username, Password: string(hashedPassword)}
	result := global.DB.Create(&user)
	if result.Error != nil {
		logrus.Errorf("创建用户失败: %v", result.Error)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": -1,
			"msg":  "服务器内部错误",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"code": 200,
		"msg":  "注册成功",
	})
}

// 假设checkIfUsernameExists是一个用于检查用户名是否存在的辅助函数
func checkIfUsernameExists(username string) bool {
	var count int64
	global.DB.Model(models.User{}).Where("username = ?", username).Count(&count)
	return count > 0
}

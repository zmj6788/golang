package middleware

import (
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

// 自定义Claims结构体，添加自定义字段
type CustomClaims struct {
	jwt.StandardClaims
	// 添加其他你需要的自定义字段，例如Username、Role等
	Username string
}

// GenerateToken 签发JWT令牌函数
func GenerateToken(claims CustomClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	secretKey := []byte("your_secret_key_here") // 替换为你的密钥
	return token.SignedString(secretKey)
}

// JWTMiddleware 验证JWT中间件
func JWTMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
			c.Abort()
			return
		}

		tokenStr := authHeader[len("Bearer "):]
		token, err := jwt.ParseWithClaims(tokenStr, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte("your_secret_key_here"), nil // 同样替换为你的密钥
		})

		if err != nil {
			if ve, ok := err.(*jwt.ValidationError); ok && ve.Errors&jwt.ValidationErrorMalformed != 0 {
				c.JSON(http.StatusBadRequest, gin.H{"message": "Malformed token"})
			} else if ve.Errors&(jwt.ValidationErrorExpired|jwt.ValidationErrorNotValidYet) != 0 {
				c.JSON(http.StatusUnauthorized, gin.H{"message": "Token is either expired or not yet valid"})
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
			}
			c.Abort()
			return
		}

		// 如果token验证成功，那么我们可以通过claims获取相关信息
		if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
			c.Set("username", claims.Username) // 将用户名设置到上下文中供后续路由使用
			c.Next()                           // 调用下一个中间件或路由处理函数
			return
		}

		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		c.Abort()
	}
}

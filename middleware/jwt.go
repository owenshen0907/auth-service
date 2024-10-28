package middleware

import (
	"github.com/gin-gonic/gin"
	// 添加以下导入
	"github.com/golang-jwt/jwt/v5"

	"net/http"
	"os"
	// "strings"
)

func JWTMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, err := c.Cookie("auth-session")
		if err != nil {
			c.Redirect(http.StatusFound, "http://B-website.com/login")
			c.Abort()
			return
		}

		// 解析和验证 JWT 令牌
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("JWT_SECRET")), nil
		}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))

		if err != nil || !token.Valid {
			c.Redirect(http.StatusFound, "http://B-website.com/login")
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.Redirect(http.StatusFound, "http://B-website.com/login")
			c.Abort()
			return
		}

		c.Set("userID", claims["userID"])
		c.Next()
	}
}

// handlers/auth.go
package handlers

import (
	"auth-service/db"
	"database/sql"
	"net/http"

	"auth-service/models"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

// RegisterRequest 定义注册请求的结构
type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginRequest 定义登录请求的结构
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Register 处理用户注册
func Register(database *db.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req RegisterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			logrus.Errorf("Invalid registration request: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		// 检查用户名是否已存在
		var existingUserID int
		err := database.DB.QueryRow("SELECT id FROM users WHERE username = ?", req.Username).Scan(&existingUserID)
		if err != sql.ErrNoRows {
			if err == nil {
				c.JSON(http.StatusConflict, gin.H{"error": "Username already exists"})
				return
			}
			logrus.Errorf("Database error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		// 哈希密码
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			logrus.Errorf("Password hashing failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		// 创建用户
		_, err = database.DB.Exec("INSERT INTO users (username, password) VALUES (?, ?)", req.Username, string(hashedPassword))
		if err != nil {
			logrus.Errorf("User creation failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully"})
	}
}

// Login 处理用户登录
func Login(database *db.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			logrus.Errorf("Invalid login request: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		// 查找用户
		var user models.User
		err := database.DB.QueryRow("SELECT id, password FROM users WHERE username = ?", req.Username).Scan(&user.ID, &user.Password)
		if err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
				return
			}
			logrus.Errorf("Database error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		// 比较密码
		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}

		// 设置会话
		session := sessions.Default(c)
		session.Set("user", user.ID) // 存储用户ID
		if err := session.Save(); err != nil {
			logrus.Errorf("Failed to save session: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Login successful"})
	}
}

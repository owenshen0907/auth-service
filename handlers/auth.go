package handlers

import (
	"auth-service/db"
	"database/sql"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"auth-service/models"

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

		logrus.Infof("Registering user: %s", req.Username)

		// 检查用户名是否已存在
		var existingUserID int
		err := database.DB.QueryRow("SELECT id FROM users WHERE username = ?", req.Username).Scan(&existingUserID)
		if err == sql.ErrNoRows {
			// 用户名不存在，继续注册流程
		} else if err != nil {
			// 数据库查询出错
			logrus.Errorf("Database error while checking username: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		} else {
			// 用户名已存在
			logrus.Warnf("Username already exists: %s", req.Username)
			c.JSON(http.StatusConflict, gin.H{"error": "Username already exists"})
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
		_, err = database.DB.Exec("INSERT INTO users (username, password, created_at, updated_at) VALUES (?, ?, ?, ?)", req.Username, string(hashedPassword), time.Now(), time.Now())
		if err != nil {
			logrus.Errorf("User creation failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		logrus.Infof("User registered successfully: %s", req.Username)
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

		logrus.Infof("User attempting to login: %s", req.Username)

		// 查找用户
		var user models.User
		err := database.DB.QueryRow("SELECT password FROM users WHERE username = ?", req.Username).Scan(&user.Password)
		if err != nil {
			if err == sql.ErrNoRows {
				logrus.Warnf("Invalid credentials for user: %s", req.Username)
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
				return
			}
			logrus.Errorf("Database error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		// 比较密码
		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
			logrus.Warnf("Invalid credentials for user: %s", req.Username)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}

		// 生成 JWT
		claims := jwt.MapClaims{
			"userName": req.Username,
			"exp":      jwt.NewNumericDate(time.Now().Add(720 * time.Hour)),
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

		tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
			return
		}

		// 设置 Cookie
		c.SetCookie(
			"jwtToken",
			tokenString,
			720*60*60, // 30 天，单位：秒
			"/",       // Cookie 有效路径
			"",        // 不设置 Domain，默认为当前域
			false,     // Secure，开发环境下为 false
			true,      // HttpOnly
		)

		// 设置 SameSite 属性
		c.SetSameSite(http.SameSiteLaxMode) // 或者使用 http.SameSiteNoneMode

		c.JSON(http.StatusOK, gin.H{"message": "登录成功"})
	}
}

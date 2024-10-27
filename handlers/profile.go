package handlers

import (
	"auth-service/db"
	"database/sql"
	"net/http"

	"auth-service/models"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// ProfileResponse 定义返回的用户信息
type ProfileResponse struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
}

// Profile 处理获取当前用户信息
func Profile(database *db.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		userID := session.Get("user")

		if userID == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		var user models.User
		err := database.DB.QueryRow("SELECT id, username FROM users WHERE id = ?", userID).Scan(&user.ID, &user.Username)
		if err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
				return
			}
			logrus.Errorf("Database error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		response := ProfileResponse{
			ID:       user.ID,
			Username: user.Username,
		}

		c.JSON(http.StatusOK, response)
	}
}

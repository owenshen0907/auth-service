package main

import (
	"log"
	"strings"

	// "net/http"
	"os"

	"auth-service/db"
	"auth-service/handlers"
	"auth-service/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

func main() {
	// 检查是否是生产环境，如果不是则加载 .env 文件
	if os.Getenv("ENV") != "production" {
		err := godotenv.Load()
		if err != nil {
			log.Fatalf("Error loading .env file: %v", err)
		}
	}
	// 从环境变量中获取允许的来源
	allowOrigins := os.Getenv("ALLOW_ORIGINS")
	origins := strings.Split(allowOrigins, ",")

	// 初始化日志
	configureLogger()

	// 初始化数据库
	database, err := db.NewDatabase()
	if err != nil {
		logrus.Fatalf("Failed to initialize database: %v", err)
	}

	// 初始化 Gin 路由器
	router := gin.Default()

	// 配置 CORS 中间件
	router.Use(cors.New(cors.Config{
		AllowOrigins:     origins, // 使用配置文件中的 CORS 来源
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// 定义公开路由，并传递数据库实例
	public := router.Group("/api")
	{
		public.POST("/register", handlers.Register(database))
		public.POST("/login", handlers.Login(database))
	}

	// 定义受保护路由
	protected := router.Group("/api")
	protected.Use(middleware.JWTMiddleware())
	{
		protected.GET("/profile", handlers.Profile(database))
	}

	// 监听端口
	port := os.Getenv("PORT")
	if port == "" {
		port = "5050"
	}
	if err := router.Run(":" + port); err != nil {
		logrus.Fatalf("Failed to run server: %v", err)
	}
}

// configureLogger 设置日志配置
func configureLogger() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.InfoLevel)
}

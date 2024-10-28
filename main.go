package main

import (
	"log"
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
	// 加载环境变量
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

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
		AllowOrigins:     []string{"http://localhost:3030"}, // 修改为实际允许的域
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
		port = "5000"
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

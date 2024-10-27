// db/db.go
package db

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/sirupsen/logrus"
)

type Database struct {
	DB *sql.DB
}

func NewDatabase() (*Database, error) {
	// 获取环境变量
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")

	if dbUser == "" || dbPassword == "" || dbHost == "" || dbPort == "" || dbName == "" {
		logrus.Fatal("Database configuration is incomplete")
	}

	// 构建 DSN
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		dbUser, dbPassword, dbHost, dbPort, dbName)

	// 连接数据库
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	// 设置数据库连接池参数（可选）
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	// 检查数据库连接
	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &Database{DB: db}, nil
}

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

// NewDatabase 初始化数据库连接并确保所需表存在
func NewDatabase() (*Database, error) {
	// 获取环境变量
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")

	// 检查环境变量是否完整
	if dbUser == "" || dbPassword == "" || dbHost == "" || dbPort == "" || dbName == "" {
		logrus.Fatal("Database configuration is incomplete")
	}

	// 构建 DSN（数据源名称）
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		dbUser, dbPassword, dbHost, dbPort, dbName)

	// 连接数据库
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// 设置数据库连接池参数（可选）
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	// 检查数据库连接
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	database := &Database{DB: db}

	// 确保用户表存在
	if err := database.ensureUsersTable(); err != nil {
		return nil, fmt.Errorf("failed to ensure users table: %w", err)
	}

	return database, nil
}

// ensureUsersTable 检查 users 表是否存在，如果不存在则创建
func (d *Database) ensureUsersTable() error {
	// 定义创建表的 SQL 语句
	createTableQuery := `
	CREATE TABLE IF NOT EXISTS users (
		id INT AUTO_INCREMENT PRIMARY KEY,
		username VARCHAR(255) NOT NULL UNIQUE,
		password VARCHAR(255) NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`

	// 执行创建表的 SQL 语句
	_, err := d.DB.Exec(createTableQuery)
	if err != nil {
		logrus.Errorf("Failed to create users table: %v", err)
		return err
	}

	logrus.Info("Ensured that users table exists")
	return nil
}

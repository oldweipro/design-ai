package database

import (
	"log"
	"os"
	"path/filepath"

	"github.com/oldweipro/design-ai/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitDatabase() {
	dbPath := getDBPath()
	
	// 确保数据库目录存在
	if err := ensureDBDir(dbPath); err != nil {
		log.Fatal("Failed to create database directory:", err)
	}

	var err error
	DB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	err = DB.AutoMigrate(&models.User{}, &models.Portfolio{})
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	log.Printf("Database connected and migrated successfully at: %s", dbPath)
}

// getDBPath 获取数据库路径，支持环境变量配置
func getDBPath() string {
	if dbPath := os.Getenv("DATABASE_URL"); dbPath != "" {
		return dbPath
	}
	if dbPath := os.Getenv("DB_PATH"); dbPath != "" {
		return dbPath
	}
	// 默认路径
	return "design_ai.db"
}

// ensureDBDir 确保数据库文件的目录存在
func ensureDBDir(dbPath string) error {
	dir := filepath.Dir(dbPath)
	if dir != "." {
		return os.MkdirAll(dir, 0755)
	}
	return nil
}

func GetDB() *gorm.DB {
	return DB
}

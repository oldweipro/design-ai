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

	err = DB.AutoMigrate(&models.User{}, &models.Portfolio{}, &models.PortfolioVersion{}, &models.MinIOConfig{}, &models.FileObject{}, &models.AdminSettings{})
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// 确保存在默认管理员设置
	if err := ensureDefaultAdminSettings(); err != nil {
		log.Fatal("Failed to create default admin settings:", err)
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

// ensureDefaultAdminSettings 确保存在默认的管理员设置
func ensureDefaultAdminSettings() error {
	var count int64
	err := DB.Model(&models.AdminSettings{}).Count(&count).Error
	if err != nil {
		return err
	}

	// 如果没有设置记录，创建默认设置
	if count == 0 {
		defaultSettings := &models.AdminSettings{
			UserApprovalRequired:      false,
			PortfolioApprovalRequired: false,
		}
		return DB.Create(defaultSettings).Error
	}

	return nil
}

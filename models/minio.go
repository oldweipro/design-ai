package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MinIOConfig MinIO配置模型
type MinIOConfig struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"size:100;not null;unique" binding:"required"` // 配置名称
	Endpoint    string    `json:"endpoint" gorm:"size:255;not null" binding:"required"`    // MinIO服务端点
	AccessKey   string    `json:"access_key" gorm:"size:100;not null" binding:"required"`  // 访问密钥
	SecretKey   string    `json:"secret_key" gorm:"size:255;not null" binding:"required"`  // 秘密密钥
	BucketName  string    `json:"bucket_name" gorm:"size:100;not null" binding:"required"` // 存储桶名称
	UseSSL      bool      `json:"use_ssl" gorm:"default:true"`                             // 是否使用HTTPS
	IsPrivate   bool      `json:"is_private" gorm:"default:false"`                         // 是否为私有存储桶
	Region      string    `json:"region" gorm:"size:50;default:'us-east-1'"`               // 区域
	URLExpiry   int       `json:"url_expiry" gorm:"default:3600"`                          // URL过期时间(秒)
	IsActive    bool      `json:"is_active" gorm:"default:false"`                          // 是否为当前激活配置
	Description string    `json:"description" gorm:"size:500"`                             // 配置描述
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// FileObject 文件对象模型
type FileObject struct {
	ID           string         `json:"id" gorm:"primaryKey;size:36"`           // 对象ID (UUID)
	OriginalName string         `json:"original_name" gorm:"size:255;not null"` // 原始文件名
	StoragePath  string         `json:"storage_path" gorm:"size:500;not null"`  // MinIO存储路径
	ContentType  string         `json:"content_type" gorm:"size:100"`           // MIME类型
	FileSize     int64          `json:"file_size" gorm:"not null"`              // 文件大小(字节)
	MD5Hash      string         `json:"md5_hash" gorm:"size:32"`                // MD5哈希值
	ConfigID     uint           `json:"config_id" gorm:"not null"`              // 所属MinIO配置ID
	Config       MinIOConfig    `json:"config" gorm:"foreignKey:ConfigID"`      // 关联的MinIO配置
	IsPublic     bool           `json:"is_public" gorm:"default:false"`         // 是否为公开文件
	Tags         string         `json:"tags" gorm:"size:500"`                   // 标签(JSON格式)
	Metadata     string         `json:"metadata" gorm:"type:text"`              // 元数据(JSON格式)
	UploadedBy   string         `json:"uploaded_by" gorm:"not null"`            // 上传者用户ID
	User         User           `json:"user" gorm:"foreignKey:UploadedBy"`      // 上传者用户信息
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"deleted_at" gorm:"index"` // 软删除
}

// BeforeCreate 创建前钩子，生成UUID
func (f *FileObject) BeforeCreate(tx *gorm.DB) error {
	if f.ID == "" {
		f.ID = uuid.New().String()
	}
	return nil
}

// TableName 指定表名
func (MinIOConfig) TableName() string {
	return "minio_configs"
}

// TableName 指定表名
func (FileObject) TableName() string {
	return "file_objects"
}

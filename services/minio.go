package services

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/oldweipro/design-ai/database"
	"github.com/oldweipro/design-ai/models"
	"gorm.io/gorm"
)

var minioClient *minio.Client
var activeConfig *models.MinIOConfig

// MinIOService MinIO服务接口
type MinIOService interface {
	InitializeClient(config *models.MinIOConfig) error
	UploadFile(file *multipart.FileHeader, userID uint, isPublic bool, tags map[string]string) (*models.FileObject, error)
	GetFileURL(objectID string) (string, error)
	DeleteFile(objectID string) error
	GetActiveConfig() *models.MinIOConfig
	SetActiveConfig(configID uint) error
	TestConnection(config *models.MinIOConfig) error
}

type minioService struct{}

// NewMinIOService 创建MinIO服务实例
func NewMinIOService() MinIOService {
	return &minioService{}
}

// InitializeClient 初始化MinIO客户端
func (s *minioService) InitializeClient(config *models.MinIOConfig) error {
	if config == nil {
		return errors.New("minio config is required")
	}

	client, err := minio.New(config.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.AccessKey, config.SecretKey, ""),
		Secure: config.UseSSL,
		Region: config.Region,
	})
	if err != nil {
		return fmt.Errorf("failed to create minio client: %w", err)
	}

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	exists, err := client.BucketExists(ctx, config.BucketName)
	if err != nil {
		return fmt.Errorf("failed to check bucket existence: %w", err)
	}

	if !exists {
		// 创建存储桶
		err = client.MakeBucket(ctx, config.BucketName, minio.MakeBucketOptions{
			Region: config.Region,
		})
		if err != nil {
			return fmt.Errorf("failed to create bucket: %w", err)
		}
		log.Printf("Created bucket: %s", config.BucketName)
	}

	minioClient = client
	activeConfig = config
	log.Printf("MinIO client initialized successfully. Endpoint: %s, Bucket: %s",
		config.Endpoint, config.BucketName)

	return nil
}

// TestConnection 测试MinIO连接
func (s *minioService) TestConnection(config *models.MinIOConfig) error {
	client, err := minio.New(config.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.AccessKey, config.SecretKey, ""),
		Secure: config.UseSSL,
		Region: config.Region,
	})
	if err != nil {
		return fmt.Errorf("failed to create minio client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = client.BucketExists(ctx, config.BucketName)
	if err != nil {
		return fmt.Errorf("connection test failed: %w", err)
	}

	return nil
}

// UploadFile 上传文件
func (s *minioService) UploadFile(file *multipart.FileHeader, userID uint, isPublic bool, tags map[string]string) (*models.FileObject, error) {
	if minioClient == nil || activeConfig == nil {
		return nil, errors.New("minio client not initialized")
	}

	// 打开上传的文件
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	// 生成唯一的文件名
	objectID := uuid.New().String()
	ext := filepath.Ext(file.Filename)
	objectName := fmt.Sprintf("%s%s", objectID, ext)

	// 读取文件内容计算MD5
	fileContent, err := io.ReadAll(src)
	if err != nil {
		return nil, fmt.Errorf("failed to read file content: %w", err)
	}

	hasher := md5.New()
	hasher.Write(fileContent)
	md5Hash := hex.EncodeToString(hasher.Sum(nil))

	// 重新打开文件用于上传
	src, err = file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to reopen file for upload: %w", err)
	}
	defer src.Close()

	// 设置上传选项
	uploadOptions := minio.PutObjectOptions{
		ContentType:  file.Header.Get("Content-Type"),
		UserMetadata: make(map[string]string),
	}

	// 添加自定义标签
	for k, v := range tags {
		uploadOptions.UserMetadata[k] = v
	}
	uploadOptions.UserMetadata["uploaded-by"] = fmt.Sprintf("%d", userID)
	uploadOptions.UserMetadata["original-name"] = file.Filename

	// 上传文件
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	_, err = minioClient.PutObject(ctx, activeConfig.BucketName, objectName, src, file.Size, uploadOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file to minio: %w", err)
	}

	// 创建文件对象记录
	fileObject := &models.FileObject{
		ID:           objectID,
		OriginalName: file.Filename,
		StoragePath:  objectName,
		ContentType:  file.Header.Get("Content-Type"),
		FileSize:     file.Size,
		MD5Hash:      md5Hash,
		ConfigID:     activeConfig.ID,
		IsPublic:     isPublic,
		UploadedBy:   userID,
		Tags:         mapToJSON(tags),
		Metadata:     "",
	}

	// 保存到数据库
	db := database.GetDB()
	if err := db.Create(fileObject).Error; err != nil {
		// 如果数据库保存失败，尝试删除MinIO中的文件
		s.removeFromMinIO(objectName)
		return nil, fmt.Errorf("failed to save file record: %w", err)
	}

	log.Printf("File uploaded successfully. ObjectID: %s, OriginalName: %s", objectID, file.Filename)
	return fileObject, nil
}

// GetFileURL 根据对象ID获取文件URL
func (s *minioService) GetFileURL(objectID string) (string, error) {
	if minioClient == nil || activeConfig == nil {
		return "", errors.New("minio client not initialized")
	}

	// 从数据库获取文件对象信息
	var fileObject models.FileObject
	db := database.GetDB()
	if err := db.Where("id = ?", objectID).First(&fileObject).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", errors.New("file not found")
		}
		return "", fmt.Errorf("failed to get file record: %w", err)
	}

	// 如果是公开文件且存储桶不是私有的，直接返回公共URL
	if fileObject.IsPublic && !activeConfig.IsPrivate {
		protocol := "http"
		if activeConfig.UseSSL {
			protocol = "https"
		}
		return fmt.Sprintf("%s://%s/%s/%s", protocol, activeConfig.Endpoint,
			activeConfig.BucketName, fileObject.StoragePath), nil
	}

	// 生成预签名URL
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	expiry := time.Duration(activeConfig.URLExpiry) * time.Second
	url, err := minioClient.PresignedGetObject(ctx, activeConfig.BucketName,
		fileObject.StoragePath, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return url.String(), nil
}

// DeleteFile 删除文件
func (s *minioService) DeleteFile(objectID string) error {
	if minioClient == nil || activeConfig == nil {
		return errors.New("minio client not initialized")
	}

	// 从数据库获取文件对象信息
	var fileObject models.FileObject
	db := database.GetDB()
	if err := db.Where("id = ?", objectID).First(&fileObject).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("file not found")
		}
		return fmt.Errorf("failed to get file record: %w", err)
	}

	// 从MinIO删除文件
	if err := s.removeFromMinIO(fileObject.StoragePath); err != nil {
		log.Printf("Warning: failed to delete file from MinIO: %v", err)
	}

	// 软删除数据库记录
	if err := db.Delete(&fileObject).Error; err != nil {
		return fmt.Errorf("failed to delete file record: %w", err)
	}

	log.Printf("File deleted successfully. ObjectID: %s", objectID)
	return nil
}

// GetActiveConfig 获取当前激活的配置
func (s *minioService) GetActiveConfig() *models.MinIOConfig {
	return activeConfig
}

// SetActiveConfig 设置激活配置
func (s *minioService) SetActiveConfig(configID uint) error {
	db := database.GetDB()

	// 取消当前激活配置
	if err := db.Model(&models.MinIOConfig{}).Where("is_active = ?", true).
		Update("is_active", false).Error; err != nil {
		return fmt.Errorf("failed to deactivate current config: %w", err)
	}

	// 获取新配置
	var config models.MinIOConfig
	if err := db.Where("id = ?", configID).First(&config).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("config not found")
		}
		return fmt.Errorf("failed to get config: %w", err)
	}

	// 设置为激活状态
	if err := db.Model(&config).Update("is_active", true).Error; err != nil {
		return fmt.Errorf("failed to activate config: %w", err)
	}

	// 初始化客户端
	return s.InitializeClient(&config)
}

// removeFromMinIO 从MinIO删除文件
func (s *minioService) removeFromMinIO(objectName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return minioClient.RemoveObject(ctx, activeConfig.BucketName, objectName, minio.RemoveObjectOptions{})
}

// mapToJSON 将map转换为JSON字符串
func mapToJSON(m map[string]string) string {
	if len(m) == 0 {
		return "{}"
	}

	parts := make([]string, 0, len(m))
	for k, v := range m {
		parts = append(parts, fmt.Sprintf(`"%s":"%s"`, k, v))
	}
	return fmt.Sprintf("{%s}", strings.Join(parts, ","))
}

// LoadActiveConfig 从数据库加载激活配置
func LoadActiveConfig() error {
	db := database.GetDB()
	var config models.MinIOConfig

	if err := db.Where("is_active = ?", true).First(&config).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Println("No active MinIO configuration found")
			return nil
		}
		return fmt.Errorf("failed to load active config: %w", err)
	}

	service := NewMinIOService()
	return service.InitializeClient(&config)
}

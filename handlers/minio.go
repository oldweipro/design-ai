package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/oldweipro/design-ai/database"
	"github.com/oldweipro/design-ai/middleware"
	"github.com/oldweipro/design-ai/models"
	"github.com/oldweipro/design-ai/services"
)

// CreateMinIOConfig 创建MinIO配置
func CreateMinIOConfig(c *gin.Context) {
	var req models.MinIOConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format", "details": err.Error()})
		return
	}

	// 测试连接
	minioService := services.NewMinIOService()
	if err := minioService.TestConnection(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to connect to MinIO", "details": err.Error()})
		return
	}

	db := database.GetDB()

	// 如果设置为激活状态，先取消其他配置的激活状态
	if req.IsActive {
		db.Model(&models.MinIOConfig{}).Where("is_active = ?", true).Update("is_active", false)
	}

	// 创建配置
	if err := db.Create(&req).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create config", "details": err.Error()})
		return
	}

	// 如果是激活配置，初始化客户端
	if req.IsActive {
		if err := minioService.InitializeClient(&req); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initialize MinIO client", "details": err.Error()})
			return
		}
	}

	// 清除敏感信息
	req.SecretKey = "******"

	c.JSON(http.StatusCreated, gin.H{"message": "MinIO config created successfully", "config": req})
}

// GetMinIOConfigs 获取MinIO配置列表
func GetMinIOConfigs(c *gin.Context) {
	db := database.GetDB()
	var configs []models.MinIOConfig

	if err := db.Select("id, name, endpoint, access_key, bucket_name, use_ssl, is_private, region, url_expiry, is_active, description, created_at, updated_at").Find(&configs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get configs", "details": err.Error()})
		return
	}

	// 清除敏感信息
	for i := range configs {
		configs[i].SecretKey = "******"
	}

	c.JSON(http.StatusOK, gin.H{"configs": configs})
}

// GetMinIOConfig 获取单个MinIO配置
func GetMinIOConfig(c *gin.Context) {
	id := c.Param("id")
	configID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid config ID"})
		return
	}

	db := database.GetDB()
	var config models.MinIOConfig

	if err := db.Where("id = ?", uint(configID)).First(&config).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Config not found"})
		return
	}

	// 为了测试连接，我们保留完整的配置信息，但在返回给前端时清除敏感信息
	responseConfig := config
	responseConfig.SecretKey = "******"

	c.JSON(http.StatusOK, gin.H{"config": responseConfig})
}

// UpdateMinIOConfig 更新MinIO配置
func UpdateMinIOConfig(c *gin.Context) {
	id := c.Param("id")
	configID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid config ID"})
		return
	}

	var req models.MinIOConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format", "details": err.Error()})
		return
	}

	db := database.GetDB()
	var existingConfig models.MinIOConfig

	if err := db.Where("id = ?", uint(configID)).First(&existingConfig).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Config not found"})
		return
	}

	// 如果SecretKey为空或为掩码，保持原有值
	if req.SecretKey == "" || req.SecretKey == "******" {
		req.SecretKey = existingConfig.SecretKey
	}

	// 测试连接
	minioService := services.NewMinIOService()
	if err := minioService.TestConnection(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to connect to MinIO", "details": err.Error()})
		return
	}

	// 如果设置为激活状态，先取消其他配置的激活状态
	if req.IsActive {
		db.Model(&models.MinIOConfig{}).Where("is_active = ? AND id != ?", true, uint(configID)).Update("is_active", false)
	}

	// 更新配置
	req.ID = uint(configID)
	if err := db.Save(&req).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update config", "details": err.Error()})
		return
	}

	// 如果是激活配置，重新初始化客户端
	if req.IsActive {
		if err := minioService.InitializeClient(&req); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initialize MinIO client", "details": err.Error()})
			return
		}
	}

	// 清除敏感信息
	req.SecretKey = "******"

	c.JSON(http.StatusOK, gin.H{"message": "Config updated successfully", "config": req})
}

// DeleteMinIOConfig 删除MinIO配置
func DeleteMinIOConfig(c *gin.Context) {
	id := c.Param("id")
	configID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid config ID"})
		return
	}

	db := database.GetDB()
	var config models.MinIOConfig

	if err := db.Where("id = ?", uint(configID)).First(&config).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Config not found"})
		return
	}

	// 检查是否有文件使用此配置
	var fileCount int64
	if err := db.Model(&models.FileObject{}).Where("config_id = ?", uint(configID)).Count(&fileCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check file usage"})
		return
	}

	if fileCount > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete config with associated files"})
		return
	}

	// 删除配置
	if err := db.Delete(&config).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete config"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Config deleted successfully"})
}

// ActivateMinIOConfig 激活MinIO配置
func ActivateMinIOConfig(c *gin.Context) {
	id := c.Param("id")
	configID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid config ID"})
		return
	}

	minioService := services.NewMinIOService()
	if err := minioService.SetActiveConfig(uint(configID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to activate config", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Config activated successfully"})
}

// TestMinIOConnection 测试MinIO连接
func TestMinIOConnection(c *gin.Context) {
	var req models.MinIOConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format", "details": err.Error()})
		return
	}

	minioService := services.NewMinIOService()
	if err := minioService.TestConnection(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Connection test failed", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Connection test successful"})
}

// TestMinIOConfigConnection 测试指定ID的MinIO配置连接
func TestMinIOConfigConnection(c *gin.Context) {
	id := c.Param("id")
	configID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid config ID"})
		return
	}

	db := database.GetDB()
	var config models.MinIOConfig

	if err := db.Where("id = ?", uint(configID)).First(&config).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Config not found"})
		return
	}

	minioService := services.NewMinIOService()
	if err := minioService.TestConnection(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Connection test failed", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Connection test successful"})
}

// UploadFile 上传文件
func UploadFile(c *gin.Context) {
	// 获取用户ID
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// 获取上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded", "details": err.Error()})
		return
	}

	// 获取参数
	isPublic := c.PostForm("is_public") == "true"
	tags := make(map[string]string)

	// 可以添加更多自定义标签
	if category := c.PostForm("category"); category != "" {
		tags["category"] = category
	}
	if purpose := c.PostForm("purpose"); purpose != "" {
		tags["purpose"] = purpose
	}

	// 上传文件
	minioService := services.NewMinIOService()
	fileObject, err := minioService.UploadFile(file, userID, isPublic, tags)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload file", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "File uploaded successfully", "file": fileObject})
}

// GetFileURL 获取文件URL
func GetFileURL(c *gin.Context) {
	objectID := c.Param("id")
	if objectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Object ID is required"})
		return
	}

	minioService := services.NewMinIOService()
	url, err := minioService.GetFileURL(objectID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Failed to get file URL", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"url": url})
}

// DeleteFile 删除文件
func DeleteFile(c *gin.Context) {
	objectID := c.Param("id")
	if objectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Object ID is required"})
		return
	}

	minioService := services.NewMinIOService()
	if err := minioService.DeleteFile(objectID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete file", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "File deleted successfully"})
}

// GetFiles 获取文件列表
func GetFiles(c *gin.Context) {
	db := database.GetDB()
	var files []models.FileObject

	// 分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	offset := (page - 1) * pageSize

	query := db.Preload("User").Preload("Config")

	// 过滤参数
	if userID := c.Query("user_id"); userID != "" {
		query = query.Where("uploaded_by = ?", userID)
	}
	if contentType := c.Query("content_type"); contentType != "" {
		query = query.Where("content_type LIKE ?", "%"+contentType+"%")
	}
	if isPublic := c.Query("is_public"); isPublic != "" {
		query = query.Where("is_public = ?", isPublic == "true")
	}

	var total int64
	query.Model(&models.FileObject{}).Count(&total)

	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&files).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get files", "details": err.Error()})
		return
	}

	// 生成URL
	minioService := services.NewMinIOService()
	for i := range files {
		if url, err := minioService.GetFileURL(files[i].ID); err == nil {
			files[i].StoragePath = url // 临时使用StoragePath字段返回URL
		}
		// 清除敏感配置信息
		if files[i].Config.ID > 0 {
			files[i].Config.SecretKey = "******"
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"files": files,
		"pagination": gin.H{
			"page":        page,
			"page_size":   pageSize,
			"total":       total,
			"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}

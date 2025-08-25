package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/oldweipro/design-ai/database"
	"github.com/oldweipro/design-ai/middleware"
	"github.com/oldweipro/design-ai/models"
	"github.com/oldweipro/design-ai/services"
	"gorm.io/gorm"
)

// CreatePortfolioVersion 创建作品版本
func CreatePortfolioVersion(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	portfolioID := c.Param("id")
	if portfolioID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Portfolio ID is required"})
		return
	}

	var req models.CreateVersionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := database.GetDB()

	// 验证作品存在且用户有权限
	var portfolio models.Portfolio
	if err := db.Where("id = ?", portfolioID).First(&portfolio).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Portfolio not found"})
		return
	}

	// 权限检查：只有作品所有者或管理员可以创建版本
	isAdmin := middleware.IsAdmin(c)
	if !isAdmin && portfolio.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Permission denied"})
		return
	}

	// 获取下一个版本号
	nextVersion, err := getNextVersionNumber(db, portfolioID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate version number"})
		return
	}

	// 生成缩略图
	thumbnail := services.ThumbnailSvc.GenerateHTMLThumbnail(req.HTMLContent)

	// 创建新版本
	version := models.PortfolioVersion{
		PortfolioID: portfolioID,
		Version:     nextVersion,
		Title:       req.Title,
		Description: req.Description,
		HTMLContent: req.HTMLContent,
		Thumbnail:   thumbnail,
		ChangeLog:   req.ChangeLog,
		IsActive:    false, // 默认不激活
	}

	if err := db.Create(&version).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create version"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Version created successfully",
		"version": version.ToResponse(),
	})
}

// GetPortfolioVersions 获取作品的所有版本
func GetPortfolioVersions(c *gin.Context) {
	portfolioID := c.Param("id")
	if portfolioID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Portfolio ID is required"})
		return
	}

	db := database.GetDB()

	// 验证作品存在
	var portfolio models.Portfolio
	if err := db.Where("id = ?", portfolioID).First(&portfolio).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Portfolio not found"})
		return
	}

	var versions []models.PortfolioVersion
	if err := db.Where("portfolio_id = ?", portfolioID).
		Order("created_at DESC").
		Find(&versions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get versions"})
		return
	}

	// 转换为响应格式
	responses := make([]models.PortfolioVersionResponse, len(versions))
	for i, version := range versions {
		responses[i] = version.ToResponse()
	}

	c.JSON(http.StatusOK, gin.H{
		"data": responses,
	})
}

// GetPortfolioVersion 获取特定版本
func GetPortfolioVersion(c *gin.Context) {
	portfolioID := c.Param("id")
	versionID := c.Param("versionId")

	if portfolioID == "" || versionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Portfolio ID and Version ID are required"})
		return
	}

	db := database.GetDB()

	var version models.PortfolioVersion
	if err := db.Where("portfolio_id = ? AND id = ?", portfolioID, versionID).
		First(&version).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Version not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": version.ToResponse(),
	})
}

// UpdatePortfolioVersion 更新版本
func UpdatePortfolioVersion(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	portfolioID := c.Param("id")
	versionID := c.Param("versionId")

	if portfolioID == "" || versionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Portfolio ID and Version ID are required"})
		return
	}

	var req models.UpdateVersionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := database.GetDB()

	// 获取版本和相关的作品
	var version models.PortfolioVersion
	if err := db.Preload("Portfolio").
		Where("portfolio_id = ? AND id = ?", portfolioID, versionID).
		First(&version).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Version not found"})
		return
	}

	// 权限检查
	isAdmin := middleware.IsAdmin(c)
	if !isAdmin && version.Portfolio.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Permission denied"})
		return
	}

	// 使用事务来处理版本激活逻辑
	err := db.Transaction(func(tx *gorm.DB) error {
		// 如果要设置为激活版本，先取消其他版本的激活状态
		if req.IsActive != nil && *req.IsActive {
			if err := tx.Model(&models.PortfolioVersion{}).
				Where("portfolio_id = ? AND id != ?", portfolioID, versionID).
				Update("is_active", false).Error; err != nil {
				return err
			}
		}

		// 更新版本信息
		updates := make(map[string]interface{})
		if req.Title != "" {
			updates["title"] = req.Title
		}
		if req.Description != "" {
			updates["description"] = req.Description
		}
		if req.HTMLContent != "" {
			updates["html_content"] = req.HTMLContent
			updates["thumbnail"] = services.ThumbnailSvc.GenerateHTMLThumbnail(req.HTMLContent)
		}
		if req.ChangeLog != "" {
			updates["change_log"] = req.ChangeLog
		}
		if req.IsActive != nil {
			updates["is_active"] = *req.IsActive
		}

		if len(updates) > 0 {
			if err := tx.Model(&version).Updates(updates).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update version"})
		return
	}

	// 重新加载更新后的版本
	db.Where("id = ?", versionID).First(&version)

	c.JSON(http.StatusOK, gin.H{
		"message": "Version updated successfully",
		"version": version.ToResponse(),
	})
}

// DeletePortfolioVersion 删除版本
func DeletePortfolioVersion(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	portfolioID := c.Param("id")
	versionID := c.Param("versionId")

	if portfolioID == "" || versionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Portfolio ID and Version ID are required"})
		return
	}

	db := database.GetDB()

	// 获取版本和相关的作品
	var version models.PortfolioVersion
	if err := db.Preload("Portfolio").
		Where("portfolio_id = ? AND id = ?", portfolioID, versionID).
		First(&version).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Version not found"})
		return
	}

	// 权限检查
	isAdmin := middleware.IsAdmin(c)
	if !isAdmin && version.Portfolio.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Permission denied"})
		return
	}

	// 不允许删除激活的版本
	if version.IsActive {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete active version"})
		return
	}

	if err := db.Delete(&version).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete version"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Version deleted successfully",
	})
}

// SetActiveVersion 设置激活版本
func SetActiveVersion(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	portfolioID := c.Param("id")
	versionID := c.Param("versionId")

	if portfolioID == "" || versionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Portfolio ID and Version ID are required"})
		return
	}

	db := database.GetDB()

	// 获取版本和相关的作品
	var version models.PortfolioVersion
	if err := db.Preload("Portfolio").
		Where("portfolio_id = ? AND id = ?", portfolioID, versionID).
		First(&version).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Version not found"})
		return
	}

	// 权限检查
	isAdmin := middleware.IsAdmin(c)
	if !isAdmin && version.Portfolio.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Permission denied"})
		return
	}

	// 使用事务处理激活逻辑
	err := db.Transaction(func(tx *gorm.DB) error {
		// 取消所有版本的激活状态
		if err := tx.Model(&models.PortfolioVersion{}).
			Where("portfolio_id = ?", portfolioID).
			Update("is_active", false).Error; err != nil {
			return err
		}

		// 设置当前版本为激活状态
		if err := tx.Model(&version).Update("is_active", true).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set active version"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Active version set successfully",
	})
}

// 辅助函数：获取下一个版本号
func getNextVersionNumber(db *gorm.DB, portfolioID string) (string, error) {
	var maxVersion string
	err := db.Model(&models.PortfolioVersion{}).
		Where("portfolio_id = ?", portfolioID).
		Select("version").
		Order("created_at DESC").
		Limit(1).
		Pluck("version", &maxVersion).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "v1.0", nil
		}
		return "", err
	}

	if maxVersion == "" {
		return "v1.0", nil
	}

	// 解析版本号（假设格式为 v1.0, v1.1, v2.0 等）
	if strings.HasPrefix(maxVersion, "v") {
		versionParts := strings.Split(maxVersion[1:], ".")
		if len(versionParts) >= 2 {
			major, err1 := strconv.Atoi(versionParts[0])
			minor, err2 := strconv.Atoi(versionParts[1])
			if err1 == nil && err2 == nil {
				minor++
				return fmt.Sprintf("v%d.%d", major, minor), nil
			}
		}
	}

	// 如果解析失败，返回一个默认的新版本
	return "v1.0", nil
}

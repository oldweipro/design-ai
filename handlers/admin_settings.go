package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/oldweipro/design-ai/database"
	"github.com/oldweipro/design-ai/models"
)

// GetAdminSettings 获取管理员设置
func GetAdminSettings(c *gin.Context) {
	db := database.GetDB()
	var settings models.AdminSettings

	if err := db.First(&settings).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get admin settings"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": settings.ToResponse()})
}

// UpdateAdminSettings 更新管理员设置
func UpdateAdminSettings(c *gin.Context) {
	var req models.AdminSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := database.GetDB()
	var settings models.AdminSettings

	if err := db.First(&settings).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get admin settings"})
		return
	}

	// 只更新提供的字段
	if req.UserApprovalRequired != nil {
		settings.UserApprovalRequired = *req.UserApprovalRequired
	}

	if req.PortfolioApprovalRequired != nil {
		settings.PortfolioApprovalRequired = *req.PortfolioApprovalRequired
	}

	if err := db.Save(&settings).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update admin settings"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Admin settings updated successfully",
		"data":    settings.ToResponse(),
	})
}

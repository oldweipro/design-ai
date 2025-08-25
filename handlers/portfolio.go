package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/oldweipro/design-ai/database"
	"github.com/oldweipro/design-ai/middleware"
	"github.com/oldweipro/design-ai/models"
	"github.com/oldweipro/design-ai/services"
	"gorm.io/gorm"
)

// buildPortfolioResponse 构建Portfolio响应数据，包含图片URL
func buildPortfolioResponse(portfolio models.Portfolio) models.PortfolioResponse {
	var tags []string
	if portfolio.Tags != "" {
		json.Unmarshal([]byte(portfolio.Tags), &tags)
	}

	response := models.PortfolioResponse{
		ID:            portfolio.ID,
		UserID:        portfolio.UserID,
		Title:         portfolio.Title,
		Author:        portfolio.Author,
		AuthorInitial: getAuthorInitial(portfolio.Author),
		Description:   portfolio.Description,
		Content:       portfolio.Content,
		Category:      portfolio.Category,
		Tags:          tags,
		Image:         portfolio.ImageObjectID, // 保持向后兼容
		ImageURL:      "",                      // 将被下面的代码填充
		AILevel:       portfolio.AILevel,
		Likes:         portfolio.Likes,
		Views:         portfolio.Views,
		Status:        portfolio.Status,
		CreatedAt:     portfolio.CreatedAt,
		UpdatedAt:     portfolio.UpdatedAt,
	}

	// 生成图片URL
	if portfolio.ImageObjectID != "" {
		minioService := services.NewMinIOService()
		if url, err := minioService.GetFileURL(portfolio.ImageObjectID); err == nil {
			response.ImageURL = url
		} else {
			log.Printf("Failed to generate URL for object %s: %v", portfolio.ImageObjectID, err)
		}
	}

	// 添加用户信息
	if portfolio.User != nil {
		userResponse := portfolio.User.ToResponse()
		response.User = &userResponse
	}

	// 添加版本信息
	if len(portfolio.Versions) > 0 {
		response.Versions = make([]models.PortfolioVersionResponse, len(portfolio.Versions))
		for i, version := range portfolio.Versions {
			response.Versions[i] = version.ToResponse()
		}
	}

	// 添加活跃版本信息和缩略图
	if portfolio.ActiveVersion != nil {
		activeVersion := portfolio.ActiveVersion.ToResponse()
		response.ActiveVersion = &activeVersion
	} else if len(portfolio.Versions) > 0 {
		// 如果没有明确的活跃版本，使用最新版本
		latestVersion := portfolio.Versions[0].ToResponse()
		response.ActiveVersion = &latestVersion
	}

	return response
}

// buildPortfolioResponses 批量构建Portfolio响应数据
func buildPortfolioResponses(portfolios []models.Portfolio) []models.PortfolioResponse {
	responses := make([]models.PortfolioResponse, len(portfolios))
	for i, portfolio := range portfolios {
		responses[i] = buildPortfolioResponse(portfolio)
	}
	return responses
}

func GetPortfolios(c *gin.Context) {
	var query models.PortfolioQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := database.GetDB()
	var portfolios []models.Portfolio
	var total int64

	dbQuery := db.Model(&models.Portfolio{}).
		Preload("User").
		Preload("Versions", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at DESC")
		}).
		Preload("ActiveVersion")

	// 根据用户角色决定可见性
	isAdmin := middleware.IsAdmin(c)
	if !isAdmin {
		// 普通用户只能看到已发布的作品
		dbQuery = dbQuery.Where("status = ?", "published")
	} else if query.Status != "" {
		// 管理员可以按状态过滤
		dbQuery = dbQuery.Where("status = ?", query.Status)
	}

	// 用户过滤（查看特定用户的作品）
	if query.UserID != "" {
		dbQuery = dbQuery.Where("user_id = ?", query.UserID)
	}

	if query.Category != "" && query.Category != "all" {
		dbQuery = dbQuery.Where("category = ?", query.Category)
	}

	if query.Search != "" {
		searchTerm := "%" + query.Search + "%"
		dbQuery = dbQuery.Where("title LIKE ? OR author LIKE ? OR description LIKE ? OR tags LIKE ?",
			searchTerm, searchTerm, searchTerm, searchTerm)
	}

	dbQuery.Count(&total)

	offset := (query.Page - 1) * query.PageSize

	// Whitelisting allowed sort columns and orders to prevent SQL injection
	allowedSortBy := map[string]bool{
		"created_at": true,
		"updated_at": true,
		"title":      true,
		"author":     true,
		"id":         true,
	}
	allowedOrder := map[string]bool{
		"ASC":  true,
		"DESC": true,
	}

	sortBy := "created_at"
	if allowedSortBy[query.SortBy] {
		sortBy = query.SortBy
	}

	order := "DESC"
	orderUpper := ""
	if query.Order != "" {
		orderUpper = strings.ToUpper(query.Order)
	}
	if allowedOrder[orderUpper] {
		order = orderUpper
	}

	orderBy := sortBy + " " + order

	err := dbQuery.Order(orderBy).Offset(offset).Limit(query.PageSize).Find(&portfolios).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch portfolios"})
		return
	}

	responses := make([]models.PortfolioResponse, 0, len(portfolios))
	for _, portfolio := range portfolios {
		response := buildPortfolioResponse(portfolio)
		responses = append(responses, response)
	}

	c.JSON(http.StatusOK, gin.H{
		"data":        responses,
		"total":       total,
		"page":        query.Page,
		"page_size":   query.PageSize,
		"total_pages": (total + int64(query.PageSize) - 1) / int64(query.PageSize),
	})
}

func GetPortfolioByID(c *gin.Context) {
	id := c.Param("id")

	db := database.GetDB()
	var portfolio models.Portfolio

	// 根据用户角色决定可见性
	isAdmin := middleware.IsAdmin(c)
	userID, hasUser := middleware.GetCurrentUserID(c)

	query := db.Preload("User").
		Preload("Versions", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at DESC")
		}).
		Preload("ActiveVersion").
		Where("id = ?", id)

	if !isAdmin {
		// 普通用户只能看到已发布的作品，或者自己的作品
		if hasUser {
			query = query.Where("(status = ? OR user_id = ?)", "published", userID)
		} else {
			query = query.Where("status = ?", "published")
		}
	}

	err := query.First(&portfolio).Error
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Portfolio not found"})
		return
	}

	// 只有发布的作品才增加浏览量
	if portfolio.Status == "published" {
		portfolio.Views++
		db.Save(&portfolio)
	}

	response := buildPortfolioResponse(portfolio)
	c.JSON(http.StatusOK, gin.H{"data": response})
}

func CreatePortfolio(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req models.CreatePortfolioRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tagsJSON, _ := json.Marshal(req.Tags)

	db := database.GetDB()

	// 获取当前用户信息，用于设置作者昵称
	var currentUser models.User
	if err := db.Where("id = ?", userID).First(&currentUser).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户未找到"})
		return
	}

	// 获取管理员设置以确定作品状态
	var adminSettings models.AdminSettings
	if err := db.First(&adminSettings).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get admin settings"})
		return
	}

	// 根据管理员设置决定作品初始状态
	portfolioStatus := "published" // 默认直接发布
	if adminSettings.PortfolioApprovalRequired {
		portfolioStatus = "draft" // 需要管理员审核
	}

	// 使用用户昵称作为作者，如果昵称为空则使用用户名
	authorName := currentUser.Nickname
	if authorName == "" {
		authorName = currentUser.Username
	}

	portfolio := models.Portfolio{
		UserID:        userID,
		Title:         req.Title,
		Author:        authorName, // 使用用户昵称作为作者
		Description:   req.Description,
		Category:      req.Category,
		Tags:          string(tagsJSON),
		ImageObjectID: req.ImageObjectID,
		AILevel:       req.AILevel,
		Status:        portfolioStatus,
	}

	// 使用事务来创建作品和版本
	err := db.Transaction(func(tx *gorm.DB) error {
		// 创建作品
		if err := tx.Create(&portfolio).Error; err != nil {
			return err
		}

		// 创建版本
		if len(req.Versions) > 0 {
			var activeVersionCount int64
			
			// 统计活跃版本数量，确保只有一个活跃版本
			for _, versionReq := range req.Versions {
				if versionReq.IsActive {
					activeVersionCount++
				}
			}
			
			// 如果没有活跃版本，默认第一个为活跃版本
			if activeVersionCount == 0 && len(req.Versions) > 0 {
				req.Versions[0].IsActive = true
			} else if activeVersionCount > 1 {
				// 如果有多个活跃版本，只保留第一个
				for i := range req.Versions {
					if req.Versions[i].IsActive {
						if activeVersionCount > 1 {
							req.Versions[i].IsActive = false
							activeVersionCount--
						}
					}
				}
			}

			// 创建版本
			for _, versionReq := range req.Versions {
				// 生成缩略图
				thumbnail := ""
				if versionReq.HTMLContent != "" {
					thumbnail = services.ThumbnailSvc.GenerateHTMLThumbnail(versionReq.HTMLContent)
				}

				version := models.PortfolioVersion{
					PortfolioID: portfolio.ID,
					Version:     versionReq.Name,
					Title:       versionReq.Title,
					Description: versionReq.Description,
					HTMLContent: versionReq.HTMLContent,
					Thumbnail:   thumbnail,
					IsActive:    versionReq.IsActive,
					ChangeLog:   versionReq.ChangeLog,
				}

				if err := tx.Create(&version).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create portfolio"})
		return
	}

	// 预加载用户信息和版本信息
	db.Preload("User").Preload("Versions").Preload("ActiveVersion").First(&portfolio, portfolio.ID)

	response := buildPortfolioResponse(portfolio)
	c.JSON(http.StatusCreated, gin.H{"data": response})
}

func UpdatePortfolio(c *gin.Context) {
	id := c.Param("id")

	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req models.UpdatePortfolioRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := database.GetDB()
	var portfolio models.Portfolio

	if err := db.Where("id = ?", id).First(&portfolio).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Portfolio not found"})
		return
	}

	// 权限检查：只有作品所有者或管理员可以修改
	isAdmin := middleware.IsAdmin(c)
	if !isAdmin && portfolio.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Permission denied"})
		return
	}

	if req.Title != "" {
		portfolio.Title = req.Title
	}

	// 更新作者信息：从用户信息获取最新的昵称
	var currentUser models.User
	if err := db.Where("id = ?", portfolio.UserID).First(&currentUser).Error; err == nil {
		authorName := currentUser.Nickname
		if authorName == "" {
			authorName = currentUser.Username
		}
		portfolio.Author = authorName
	}

	if req.Description != "" {
		portfolio.Description = req.Description
	}
	// 移除Content字段更新，因为现在使用版本系统
	if req.Category != "" {
		portfolio.Category = req.Category
	}
	if len(req.Tags) > 0 {
		tagsJSON, _ := json.Marshal(req.Tags)
		portfolio.Tags = string(tagsJSON)
	}
	if req.ImageObjectID != "" {
		portfolio.ImageObjectID = req.ImageObjectID
	}
	if req.AILevel != "" {
		portfolio.AILevel = req.AILevel
	}

	// 状态更新：普通用户只能设为草稿，管理员可以设任何状态
	if req.Status != "" {
		if isAdmin {
			portfolio.Status = req.Status
		} else if req.Status == "draft" {
			portfolio.Status = req.Status
		}
	}

	if err := db.Save(&portfolio).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update portfolio"})
		return
	}

	// 预加载用户信息和版本信息
	db.Preload("User").
		Preload("Versions", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at DESC")
		}).
		Preload("ActiveVersion").
		First(&portfolio, portfolio.ID)

	response := buildPortfolioResponse(portfolio)
	c.JSON(http.StatusOK, gin.H{"data": response})
}

func DeletePortfolio(c *gin.Context) {
	id := c.Param("id")

	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	db := database.GetDB()
	var portfolio models.Portfolio

	if err := db.Where("id = ?", id).First(&portfolio).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Portfolio not found"})
		return
	}

	// 权限检查：只有作品所有者或管理员可以删除
	isAdmin := middleware.IsAdmin(c)
	if !isAdmin && portfolio.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Permission denied"})
		return
	}

	portfolio.Status = "deleted"
	if err := db.Save(&portfolio).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete portfolio"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Portfolio deleted successfully"})
}

func LikePortfolio(c *gin.Context) {
	id := c.Param("id")

	db := database.GetDB()
	var portfolio models.Portfolio

	if err := db.Where("id = ? AND status = ?", id, "published").First(&portfolio).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Portfolio not found"})
		return
	}

	portfolio.Likes++
	if err := db.Save(&portfolio).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to like portfolio"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Portfolio liked successfully",
		"likes":   portfolio.Likes,
	})
}

func GetCategories(c *gin.Context) {
	categories := []map[string]interface{}{
		{"value": "all", "label": "全部作品"},
		{"value": "ai", "label": "AI生成"},
		{"value": "ui", "label": "UI/UX"},
		{"value": "web", "label": "网页设计"},
		{"value": "mobile", "label": "移动应用"},
		{"value": "brand", "label": "品牌设计"},
		{"value": "3d", "label": "3D渲染"},
	}

	c.JSON(http.StatusOK, gin.H{"data": categories})
}

// 管理员：审核作品
func ApprovePortfolio(c *gin.Context) {
	id := c.Param("id")

	var req models.AdminPortfolioRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := database.GetDB()
	var portfolio models.Portfolio

	if err := db.Where("id = ?", id).First(&portfolio).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Portfolio not found"})
		return
	}

	portfolio.Status = req.Status
	if err := db.Save(&portfolio).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update portfolio"})
		return
	}

	// 预加载用户信息
	db.Preload("User").First(&portfolio, portfolio.ID)

	response := buildPortfolioResponse(portfolio)
	c.JSON(http.StatusOK, gin.H{
		"message": "Portfolio status updated successfully",
		"data":    response,
	})
}

// 获取用户自己的作品
func GetMyPortfolios(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var query models.PortfolioQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 强制设置用户ID为当前用户
	query.UserID = userID

	db := database.GetDB()
	var portfolios []models.Portfolio
	var total int64

	dbQuery := db.Model(&models.Portfolio{}).
		Preload("User").
		Preload("Versions", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at DESC")
		}).
		Preload("ActiveVersion").
		Where("user_id = ?", userID)

	if query.Status != "" {
		dbQuery = dbQuery.Where("status = ?", query.Status)
	}

	if query.Category != "" && query.Category != "all" {
		dbQuery = dbQuery.Where("category = ?", query.Category)
	}

	if query.Search != "" {
		searchTerm := "%" + query.Search + "%"
		dbQuery = dbQuery.Where("title LIKE ? OR description LIKE ? OR tags LIKE ?",
			searchTerm, searchTerm, searchTerm)
	}

	dbQuery.Count(&total)

	offset := (query.Page - 1) * query.PageSize

	// Whitelist allowed sorting columns and ordering directions
	allowedSortBy := map[string]bool{
		"created_at": true,
		"title":      true,
		"category":   true,
		"status":     true,
		// Add any other allowed fields here, must match model/DB column names
	}
	allowedOrder := map[string]string{
		"asc":  "ASC",
		"ASC":  "ASC",
		"desc": "DESC",
		"DESC": "DESC",
	}
	sortBy := "created_at" // default field
	order := "DESC"        // default order
	if allowedSortBy[query.SortBy] {
		sortBy = query.SortBy
	}
	if val, ok := allowedOrder[query.Order]; ok {
		order = val
	}
	orderBy := sortBy + " " + order

	err := dbQuery.Order(orderBy).Offset(offset).Limit(query.PageSize).Find(&portfolios).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch portfolios"})
		return
	}

	responses := make([]models.PortfolioResponse, 0, len(portfolios))
	for _, portfolio := range portfolios {
		response := buildPortfolioResponse(portfolio)
		responses = append(responses, response)
	}

	c.JSON(http.StatusOK, gin.H{
		"data":        responses,
		"total":       total,
		"page":        query.Page,
		"page_size":   query.PageSize,
		"total_pages": (total + int64(query.PageSize) - 1) / int64(query.PageSize),
	})
}

// getAuthorInitial 获取作者名字首字母
func getAuthorInitial(author string) string {
	if len(author) > 0 {
		runes := []rune(author)
		return string(runes[0])
	}
	return ""
}

// 管理员：获取所有作品
func GetAllPortfolios(c *gin.Context) {
	db := database.GetDB()
	var portfolios []models.Portfolio
	var total int64

	query := models.PortfolioQuery{
		Page:     1,
		PageSize: 50,
	}
	if err := c.ShouldBindQuery(&query); err == nil {
		if query.Page <= 0 {
			query.Page = 1
		}
		if query.PageSize <= 0 || query.PageSize > 100 {
			query.PageSize = 50
		}
	}

	dbQuery := db.Preload("User").
		Preload("Versions", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at DESC")
		}).
		Preload("ActiveVersion").
		Model(&models.Portfolio{})

	// 状态过滤
	if query.Status != "" {
		dbQuery = dbQuery.Where("status = ?", query.Status)
	}

	// 分类过滤
	if query.Category != "" {
		dbQuery = dbQuery.Where("category = ?", query.Category)
	}

	// 搜索过滤
	if query.Search != "" {
		searchPattern := "%" + query.Search + "%"
		dbQuery = dbQuery.Where("title LIKE ? OR description LIKE ? OR author LIKE ?", searchPattern, searchPattern, searchPattern)
	}

	// 获取总数
	dbQuery.Count(&total)

	// 分页
	offset := (query.Page - 1) * query.PageSize
	if err := dbQuery.Offset(offset).Limit(query.PageSize).Order("created_at DESC").Find(&portfolios).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch portfolios"})
		return
	}

	// 转换为响应格式
	responses := make([]models.PortfolioResponse, 0, len(portfolios))
	for _, portfolio := range portfolios {
		responses = append(responses, buildPortfolioResponse(portfolio))
	}

	c.JSON(http.StatusOK, gin.H{
		"data":        responses,
		"total":       total,
		"page":        query.Page,
		"page_size":   query.PageSize,
		"total_pages": (total + int64(query.PageSize) - 1) / int64(query.PageSize),
	})
}

// 管理员：更新作品状态
func UpdatePortfolioStatus(c *gin.Context) {
	portfolioID := c.Param("id")

	var req models.AdminPortfolioRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := database.GetDB()
	var portfolio models.Portfolio

	if err := db.Where("id = ?", portfolioID).First(&portfolio).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Portfolio not found"})
		return
	}

	portfolio.Status = req.Status

	if err := db.Save(&portfolio).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update portfolio"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Portfolio updated successfully",
		"data":    buildPortfolioResponse(portfolio),
	})
}

// 管理员：删除作品
func AdminDeletePortfolio(c *gin.Context) {
	portfolioID := c.Param("id")

	db := database.GetDB()
	var portfolio models.Portfolio

	if err := db.Where("id = ?", portfolioID).First(&portfolio).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Portfolio not found"})
		return
	}

	// 硬删除作品
	if err := db.Delete(&portfolio).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete portfolio"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Portfolio deleted successfully"})
}

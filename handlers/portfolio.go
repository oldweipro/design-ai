package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/oldweipro/design-ai/database"
	"github.com/oldweipro/design-ai/middleware"
	"github.com/oldweipro/design-ai/models"
)

func GetPortfolios(c *gin.Context) {
	var query models.PortfolioQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := database.GetDB()
	var portfolios []models.Portfolio
	var total int64

	dbQuery := db.Model(&models.Portfolio{}).Preload("User")

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
		dbQuery = dbQuery.Where("title LIKE ? OR author LIKE ? OR description LIKE ? OR tags LIKE ? OR content LIKE ?",
			searchTerm, searchTerm, searchTerm, searchTerm, searchTerm)
	}

	dbQuery.Count(&total)

	offset := (query.Page - 1) * query.PageSize
	orderBy := query.SortBy + " " + query.Order

	err := dbQuery.Order(orderBy).Offset(offset).Limit(query.PageSize).Find(&portfolios).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch portfolios"})
		return
	}

	responses := make([]models.PortfolioResponse, 0, len(portfolios))
	for _, portfolio := range portfolios {
		response := convertToPortfolioResponse(portfolio)
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

	query := db.Preload("User").Where("id = ?", id)

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

	response := convertToPortfolioResponse(portfolio)
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

	portfolio := models.Portfolio{
		UserID:      userID,
		Title:       req.Title,
		Author:      req.Author,
		Description: req.Description,
		Content:     req.Content,
		Category:    req.Category,
		Tags:        string(tagsJSON),
		ImageURL:    req.ImageURL,
		AILevel:     req.AILevel,
		Status:      "draft", // 默认为草稿状态，需要管理员审核
	}

	db := database.GetDB()
	if err := db.Create(&portfolio).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create portfolio"})
		return
	}

	// 预加载用户信息
	db.Preload("User").First(&portfolio, portfolio.ID)

	response := convertToPortfolioResponse(portfolio)
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
	if req.Author != "" {
		portfolio.Author = req.Author
	}
	if req.Description != "" {
		portfolio.Description = req.Description
	}
	if req.Content != "" {
		portfolio.Content = req.Content
	}
	if req.Category != "" {
		portfolio.Category = req.Category
	}
	if len(req.Tags) > 0 {
		tagsJSON, _ := json.Marshal(req.Tags)
		portfolio.Tags = string(tagsJSON)
	}
	if req.ImageURL != "" {
		portfolio.ImageURL = req.ImageURL
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

	// 预加载用户信息
	db.Preload("User").First(&portfolio, portfolio.ID)

	response := convertToPortfolioResponse(portfolio)
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

	response := convertToPortfolioResponse(portfolio)
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

	dbQuery := db.Model(&models.Portfolio{}).Preload("User").Where("user_id = ?", userID)

	if query.Status != "" {
		dbQuery = dbQuery.Where("status = ?", query.Status)
	}

	if query.Category != "" && query.Category != "all" {
		dbQuery = dbQuery.Where("category = ?", query.Category)
	}

	if query.Search != "" {
		searchTerm := "%" + query.Search + "%"
		dbQuery = dbQuery.Where("title LIKE ? OR description LIKE ? OR content LIKE ?",
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
	order := "DESC"         // default order
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
		response := convertToPortfolioResponse(portfolio)
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

func convertToPortfolioResponse(portfolio models.Portfolio) models.PortfolioResponse {
	var tags []string
	if portfolio.Tags != "" {
		if err := json.Unmarshal([]byte(portfolio.Tags), &tags); err != nil {
			log.Printf("Failed to unmarshal tags for portfolio %s: %v", portfolio.ID, err)
		}
	}

	authorInitial := ""
	if len(portfolio.Author) > 0 {
		runes := []rune(portfolio.Author)
		authorInitial = string(runes[0])
	}

	imageDisplay := portfolio.ImageURL
	if imageDisplay == "" {
		imageDisplay = fmt.Sprintf("🎨 %s", portfolio.Title)
	}

	response := models.PortfolioResponse{
		ID:            portfolio.ID,
		UserID:        portfolio.UserID,
		Title:         portfolio.Title,
		Author:        portfolio.Author,
		AuthorInitial: authorInitial,
		Description:   portfolio.Description,
		Content:       portfolio.Content,
		Category:      portfolio.Category,
		Tags:          tags,
		Image:         imageDisplay,
		ImageURL:      portfolio.ImageURL,
		AILevel:       portfolio.AILevel,
		Likes:         portfolio.Likes,
		Views:         portfolio.Views,
		Status:        portfolio.Status,
		CreatedAt:     portfolio.CreatedAt,
		UpdatedAt:     portfolio.UpdatedAt,
	}

	// 如果预加载了用户信息，则添加到响应中
	if portfolio.User != nil {
		userResponse := portfolio.User.ToResponse()
		response.User = &userResponse
	}

	return response
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

	dbQuery := db.Preload("User").Model(&models.Portfolio{})

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
		responses = append(responses, convertToPortfolioResponse(portfolio))
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
		"data":    convertToPortfolioResponse(portfolio),
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

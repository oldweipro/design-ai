package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/oldweipro/design-ai/database"
	"github.com/oldweipro/design-ai/middleware"
	"github.com/oldweipro/design-ai/models"
	"github.com/oldweipro/design-ai/utils"
)

// 用户注册
func Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := database.GetDB()

	// 检查邮箱是否已存在
	var existingUser models.User
	if err := db.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "邮箱已存在"})
		return
	}

	// 检查用户名是否已存在
	if err := db.Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "用户名已存在"})
		return
	}

	// 获取管理员设置以确定用户状态
	var adminSettings models.AdminSettings
	if err := db.First(&adminSettings).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法获取管理员设置"})
		return
	}

	// 根据管理员设置决定用户初始状态
	userStatus := "approved" // 默认直接批准
	if adminSettings.UserApprovalRequired {
		userStatus = "pending" // 需要管理员审核
	}

	// 如果没有提供昵称，使用用户名作为默认昵称
	nickname := req.Nickname
	if nickname == "" {
		nickname = req.Username
	}

	// 创建新用户
	user := models.User{
		Email:    req.Email,
		Username: req.Username,
		Nickname: nickname,
		Role:     "user",
		Status:   userStatus,
	}

	if err := user.HashPassword(req.Password); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法哈希密码"})
		return
	}

	if err := db.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法创建用户"})
		return
	}

	// 根据用户状态返回不同的消息
	message := "注册成功."
	if userStatus == "pending" {
		message = "注册成功。请等待管理员批准."
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": message,
		"user":    user.ToResponse(),
	})
}

// 用户登录
func Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := database.GetDB()
	var user models.User

	if err := db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	if !user.CheckPassword(req.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	if user.Status != "approved" {
		var message string
		switch user.Status {
		case "pending":
			message = "Account is pending approval"
		case "rejected":
			message = "Account has been rejected"
		case "banned":
			message = "Account has been banned"
		default:
			message = "Account is not active"
		}
		c.JSON(http.StatusForbidden, gin.H{"error": message})
		return
	}

	// 生成JWT token
	token, err := utils.GenerateToken(user.ID, user.Email, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	response := models.LoginResponse{
		User:  user.ToResponse(),
		Token: token,
	}

	c.JSON(http.StatusOK, gin.H{"data": response})
}

// 获取当前用户信息
func GetProfile(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	db := database.GetDB()
	var user models.User

	if err := db.Where("id = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": user.ToResponse()})
}

// 更新用户资料
func UpdateProfile(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req models.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := database.GetDB()
	var user models.User

	if err := db.Where("id = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// 检查用户名是否被其他用户使用
	if req.Username != "" && req.Username != user.Username {
		var existingUser models.User
		if err := db.Where("username = ? AND id != ?", req.Username, userID).First(&existingUser).Error; err == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "用户名已存在"})
			return
		}
		user.Username = req.Username
	}

	if req.Nickname != "" {
		user.Nickname = req.Nickname
	}

	if req.Avatar != "" {
		user.Avatar = req.Avatar
	}

	if req.Bio != "" {
		user.Bio = req.Bio
	}

	if err := db.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Profile updated successfully",
		"data":    user.ToResponse(),
	})
}

// 管理员：获取所有用户
func GetUsers(c *gin.Context) {
	var query models.UserQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := database.GetDB()
	var users []models.User
	var total int64

	dbQuery := db.Model(&models.User{})

	if query.Status != "" {
		dbQuery = dbQuery.Where("status = ?", query.Status)
	}

	if query.Role != "" {
		dbQuery = dbQuery.Where("role = ?", query.Role)
	}

	if query.Search != "" {
		searchTerm := "%" + query.Search + "%"
		dbQuery = dbQuery.Where("username LIKE ? OR email LIKE ?", searchTerm, searchTerm)
	}

	dbQuery.Count(&total)

	offset := (query.Page - 1) * query.PageSize
	err := dbQuery.Order("created_at desc").Offset(offset).Limit(query.PageSize).Find(&users).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}

	responses := make([]models.UserResponse, 0, len(users))
	for _, user := range users {
		responses = append(responses, user.ToResponse())
	}

	c.JSON(http.StatusOK, gin.H{
		"data":        responses,
		"total":       total,
		"page":        query.Page,
		"page_size":   query.PageSize,
		"total_pages": (total + int64(query.PageSize) - 1) / int64(query.PageSize),
	})
}

// 管理员：更新用户状态
func UpdateUserStatus(c *gin.Context) {
	userID := c.Param("id")

	var req models.AdminUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := database.GetDB()
	var user models.User

	if err := db.Where("id = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	user.Status = req.Status
	if req.Role != "" {
		user.Role = req.Role
	}

	if err := db.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User updated successfully",
		"data":    user.ToResponse(),
	})
}

// 管理员：删除用户
func DeleteUser(c *gin.Context) {
	userID := c.Param("id")

	// 防止删除自己
	currentUserID, _ := middleware.GetCurrentUserID(c)
	if userID == currentUserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Cannot delete your own account"})
		return
	}

	db := database.GetDB()
	var user models.User

	if err := db.Where("id = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// 软删除用户的所有作品
	if err := db.Model(&models.Portfolio{}).Where("user_id = ?", userID).Update("status", "deleted").Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user portfolios"})
		return
	}

	// 删除用户
	if err := db.Delete(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

// 管理员：重置用户密码
func ResetUserPassword(c *gin.Context) {
	userID := c.Param("id")

	db := database.GetDB()
	var user models.User

	if err := db.Where("id = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// 生成新密码（这里简化为固定密码，实际应该生成随机密码并发送邮件）
	newPassword := "Reset123456!"

	if err := user.HashPassword(newPassword); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	if err := db.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reset password"})
		return
	}

	// 实际应用中应该发送邮件，这里返回密码仅用于演示
	c.JSON(http.StatusOK, gin.H{
		"message":      "Password reset successfully",
		"new_password": newPassword, // 实际应用中不应该返回密码
	})
}

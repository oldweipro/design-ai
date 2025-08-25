package models

import (
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	ID        string    `json:"id" gorm:"type:char(36);primary_key"`
	Email     string    `json:"email" gorm:"unique;not null;size:255"`
	Username  string    `json:"username" gorm:"unique;not null;size:100"`
	Nickname  string    `json:"nickname" gorm:"size:100"`   // 昵称，用于显示
	Password  string    `json:"-" gorm:"not null;size:255"` // 不在JSON中返回密码
	Avatar    string    `json:"avatar" gorm:"size:500"`
	Bio       string    `json:"bio" gorm:"type:text"`
	Role      string    `json:"role" gorm:"default:'user';size:20"`      // user, admin
	Status    string    `json:"status" gorm:"default:'pending';size:20"` // pending, approved, rejected, banned
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	// 关联作品
	Portfolios []Portfolio `json:"portfolios,omitempty" gorm:"foreignKey:UserID"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == "" {
		u.ID = uuid.New().String()
	}
	return nil
}

// 加密密码
func (u *User) HashPassword(password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	return nil
}

// 验证密码
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}

// 用户响应结构（不包含敏感信息）
type UserResponse struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Username  string    `json:"username"`
	Nickname  string    `json:"nickname"`
	Avatar    string    `json:"avatar"`
	Bio       string    `json:"bio"`
	Role      string    `json:"role"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// 注册请求
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Username string `json:"username" binding:"required,min=3,max=20"`
	Nickname string `json:"nickname" binding:"max=50"`
	Password string `json:"password" binding:"required,min=6"`
}

// 登录请求
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// 登录响应
type LoginResponse struct {
	User  UserResponse `json:"user"`
	Token string       `json:"token"`
}

// 更新用户资料请求
type UpdateUserRequest struct {
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
	Bio      string `json:"bio"`
}

// 管理员操作请求
type AdminUserRequest struct {
	Status string `json:"status" binding:"required,oneof=approved rejected banned"`
	Role   string `json:"role,omitempty" binding:"omitempty,oneof=user admin"`
}

// 用户查询参数
type UserQuery struct {
	Status   string `form:"status"`
	Role     string `form:"role"`
	Search   string `form:"search"`
	Page     int    `form:"page,default=1"`
	PageSize int    `form:"page_size,default=20"`
}

// 转换为用户响应结构
func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:        u.ID,
		Email:     u.Email,
		Username:  u.Username,
		Nickname:  u.Nickname,
		Avatar:    u.Avatar,
		Bio:       u.Bio,
		Role:      u.Role,
		Status:    u.Status,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Portfolio struct {
	ID          string    `json:"id" gorm:"type:char(36);primary_key"`
	UserID      string    `json:"user_id" gorm:"type:char(36);index"` // 关联用户
	Title       string    `json:"title" gorm:"not null;size:255"`
	Author      string    `json:"author" gorm:"not null;size:100"`
	Description string    `json:"description" gorm:"type:text"`
	Content     string    `json:"content" gorm:"type:longtext"` // 新增详细内容字段
	Category    string    `json:"category" gorm:"not null;size:50;index"`
	Tags        string    `json:"tags" gorm:"type:text"` // JSON格式存储标签数组
	ImageURL    string    `json:"image_url" gorm:"size:500"`
	AILevel     string    `json:"ai_level" gorm:"size:50"` // AI完全生成, AI辅助设计, 手工设计
	Likes       int       `json:"likes" gorm:"default:0"`
	Views       int       `json:"views" gorm:"default:0"`
	Status      string    `json:"status" gorm:"default:'draft';size:20"` // draft, published, rejected, deleted
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	
	// 关联用户
	User *User `json:"user,omitempty" gorm:"foreignKey:UserID;references:ID"`
}

func (p *Portfolio) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	return nil
}

type PortfolioResponse struct {
	ID            string        `json:"id"`
	UserID        string        `json:"user_id"`
	Title         string        `json:"title"`
	Author        string        `json:"author"`
	AuthorInitial string        `json:"authorInitial"`
	Description   string        `json:"description"`
	Content       string        `json:"content"`
	Category      string        `json:"category"`
	Tags          []string      `json:"tags"`
	Image         string        `json:"image"`
	ImageURL      string        `json:"image_url"`
	AILevel       string        `json:"aiLevel"`
	Likes         int           `json:"likes"`
	Views         int           `json:"views"`
	Status        string        `json:"status"`
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
	User          *UserResponse `json:"user,omitempty"`
}

type CreatePortfolioRequest struct {
	Title       string   `json:"title" binding:"required"`
	Author      string   `json:"author" binding:"required"`
	Description string   `json:"description"`
	Content     string   `json:"content"`
	Category    string   `json:"category" binding:"required"`
	Tags        []string `json:"tags"`
	ImageURL    string   `json:"image_url"`
	AILevel     string   `json:"ai_level" binding:"required"`
}

type UpdatePortfolioRequest struct {
	Title       string   `json:"title"`
	Author      string   `json:"author"`
	Description string   `json:"description"`
	Content     string   `json:"content"`
	Category    string   `json:"category"`
	Tags        []string `json:"tags"`
	ImageURL    string   `json:"image_url"`
	AILevel     string   `json:"ai_level"`
	Status      string   `json:"status"`
}

// 管理员审核作品请求
type AdminPortfolioRequest struct {
	Status string `json:"status" binding:"required,oneof=published rejected"`
}

type PortfolioQuery struct {
	Category string `form:"category"`
	Search   string `form:"search"`
	Status   string `form:"status"`
	UserID   string `form:"user_id"`
	Page     int    `form:"page,default=1"`
	PageSize int    `form:"page_size,default=12"`
	SortBy   string `form:"sort_by,default=created_at"`
	Order    string `form:"order,default=desc"`
}

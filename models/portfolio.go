package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Portfolio struct {
	ID            string    `json:"id" gorm:"type:char(36);primary_key"`
	UserID        string    `json:"userId" gorm:"type:char(36);index"` // 关联用户
	Title         string    `json:"title" gorm:"not null;size:255"`
	Author        string    `json:"author" gorm:"not null;size:100"`
	Description   string    `json:"description" gorm:"type:text"`
	Content       string    `json:"content" gorm:"type:longtext"` // 新增详细内容字段
	Category      string    `json:"category" gorm:"not null;size:50;index"`
	Tags          string    `json:"tags" gorm:"type:text"`        // JSON格式存储标签数组
	ImageObjectID string    `json:"imageObjectId" gorm:"size:36"` // MinIO对象ID
	ImageURL      string    `json:"imageUrl" gorm:"-"`            // 运行时生成的URL，不存储到数据库
	AILevel       string    `json:"aiLevel" gorm:"size:50"`       // AI完全生成, AI辅助设计, 手工设计
	Likes         int       `json:"likes" gorm:"default:0"`
	Views         int       `json:"views" gorm:"default:0"`
	Status        string    `json:"status" gorm:"default:'draft';size:20"` // draft, published, rejected, deleted
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`

	// 关联用户
	User *User `json:"user,omitempty" gorm:"foreignKey:UserID;references:ID"`

	// 关联版本
	Versions      []PortfolioVersion `json:"versions,omitempty" gorm:"foreignKey:PortfolioID"`
	ActiveVersion *PortfolioVersion  `json:"activeVersion,omitempty" gorm:"foreignKey:PortfolioID;constraint:OnDelete:SET NULL;where:is_active = true"`
}

func (p *Portfolio) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	return nil
}

type PortfolioResponse struct {
	ID            string                     `json:"id"`
	UserID        string                     `json:"user_id"`
	Title         string                     `json:"title"`
	Author        string                     `json:"author"`
	AuthorInitial string                     `json:"authorInitial"`
	Description   string                     `json:"description"`
	Content       string                     `json:"content"`
	Category      string                     `json:"category"`
	Tags          []string                   `json:"tags"`
	Image         string                     `json:"image"`
	ImageURL      string                     `json:"imageUrl"`
	AILevel       string                     `json:"aiLevel"`
	Likes         int                        `json:"likes"`
	Views         int                        `json:"views"`
	Status        string                     `json:"status"`
	CreatedAt     time.Time                  `json:"createdAt"`
	UpdatedAt     time.Time                  `json:"updatedAt"`
	User          *UserResponse              `json:"user,omitempty"`
	Versions      []PortfolioVersionResponse `json:"versions,omitempty"`
	ActiveVersion *PortfolioVersionResponse  `json:"activeVersion,omitempty"`
	Thumbnail     string                     `json:"thumbnail,omitempty"` // 从活跃版本获取的缩略图
}

type CreatePortfolioRequest struct {
	Title         string                      `json:"title" binding:"required"`
	Description   string                      `json:"description"`
	Category      string                      `json:"category" binding:"required"`
	Tags          []string                    `json:"tags"`
	ImageObjectID string                      `json:"imageObjectId"`
	AILevel       string                      `json:"aiLevel" binding:"required"`
	Versions      []CreatePortfolioVersionReq `json:"versions"` // 版本信息
}

// CreatePortfolioVersionReq 创建作品时的版本请求
type CreatePortfolioVersionReq struct {
	Name        string `json:"name" binding:"required"`        // 版本名称，如 "v1.0"
	Title       string `json:"title" binding:"required"`       // 版本标题
	Description string `json:"description"`                    // 版本描述
	HTMLContent string `json:"htmlContent" binding:"required"` // HTML内容
	ChangeLog   string `json:"changeLog"`                      // 版本变更日志
	IsActive    bool   `json:"isActive"`                       // 是否为活跃版本
}

type UpdatePortfolioRequest struct {
	Title         string                         `json:"title"`
	Description   string                         `json:"description"`
	Content       string                         `json:"content"`
	Category      string                         `json:"category"`
	Tags          []string                       `json:"tags"`
	ImageObjectID string                         `json:"imageObjectId"`
	AILevel       string                         `json:"aiLevel"`
	Status        string                         `json:"status"`
	Versions      []UpdatePortfolioVersionReq    `json:"versions"` // 最终的版本列表
}

// UpdatePortfolioVersionReq 更新作品时的版本请求
type UpdatePortfolioVersionReq struct {
	ID          string `json:"id,omitempty"`                   // 版本ID，空表示新版本
	Name        string `json:"name" binding:"required"`        // 版本名称，如 "v1.0"
	Title       string `json:"title" binding:"required"`       // 版本标题
	Description string `json:"description"`                    // 版本描述
	HTMLContent string `json:"htmlContent" binding:"required"` // HTML内容
	ChangeLog   string `json:"changeLog"`                      // 版本变更日志
	IsActive    bool   `json:"isActive"`                       // 是否为活跃版本
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

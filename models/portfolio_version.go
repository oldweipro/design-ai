package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PortfolioVersion 作品版本模型
type PortfolioVersion struct {
	ID          string    `json:"id" gorm:"type:char(36);primary_key"`
	PortfolioID string    `json:"portfolioId" gorm:"type:char(36);index;not null"` // 关联作品
	Version     string    `json:"version" gorm:"size:20;not null"`                 // 版本号，如 "v1.0", "v1.1"
	Title       string    `json:"title" gorm:"size:255;not null"`                  // 版本标题
	HTMLContent string    `json:"htmlContent" gorm:"type:longtext;not null"`       // HTML内容
	ChangeLog   string    `json:"changeLog" gorm:"type:text"`                      // 版本变更日志
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`

	// 关联作品
	Portfolio *Portfolio `json:"portfolio,omitempty" gorm:"foreignKey:PortfolioID;references:ID"`
}

// PortfolioVersionResponse 版本响应结构
type PortfolioVersionResponse struct {
	ID          string    `json:"id"`
	PortfolioID string    `json:"portfolioId"`
	Version     string    `json:"version"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	HTMLContent string    `json:"htmlContent"`
	Thumbnail   string    `json:"thumbnail"`
	IsActive    bool      `json:"isActive"`
	ChangeLog   string    `json:"changeLog"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// CreateVersionRequest 创建版本请求
type CreateVersionRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	HTMLContent string `json:"htmlContent" binding:"required"`
	ChangeLog   string `json:"changeLog"`
}

// UpdateVersionRequest 更新版本请求
type UpdateVersionRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	HTMLContent string `json:"htmlContent"`
	ChangeLog   string `json:"changeLog"`
	IsActive    *bool  `json:"isActive"` // 使用指针以区分false和未设置
}

// ToResponse 转换为响应结构
func (pv *PortfolioVersion) ToResponse() PortfolioVersionResponse {
	return PortfolioVersionResponse{
		ID:          pv.ID,
		PortfolioID: pv.PortfolioID,
		Version:     pv.Version,
		Title:       pv.Title,
		HTMLContent: pv.HTMLContent,
		ChangeLog:   pv.ChangeLog,
		CreatedAt:   pv.CreatedAt,
		UpdatedAt:   pv.UpdatedAt,
	}
}

// BeforeCreate 创建前钩子，生成ID
func (pv *PortfolioVersion) BeforeCreate(tx *gorm.DB) error {
	if pv.ID == "" {
		pv.ID = uuid.New().String()
	}
	return nil
}

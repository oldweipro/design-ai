package models

import (
	"time"

	"gorm.io/gorm"
)

// AdminSettings 系统管理设置
type AdminSettings struct {
	ID                        uint      `json:"id" gorm:"primaryKey"`
	UserApprovalRequired      bool      `json:"userApprovalRequired" gorm:"default:false"`      // 新用户是否需要审核
	PortfolioApprovalRequired bool      `json:"portfolioApprovalRequired" gorm:"default:false"` // 新作品是否需要审核
	CreatedAt                 time.Time `json:"createdAt"`
	UpdatedAt                 time.Time `json:"updatedAt"`
}

// BeforeCreate 创建前的钩子
func (s *AdminSettings) BeforeCreate(tx *gorm.DB) error {
	return nil
}

// AdminSettingsRequest 更新设置请求
type AdminSettingsRequest struct {
	UserApprovalRequired      *bool `json:"userApprovalRequired"`
	PortfolioApprovalRequired *bool `json:"portfolioApprovalRequired"`
}

// AdminSettingsResponse 设置响应
type AdminSettingsResponse struct {
	ID                        uint      `json:"id"`
	UserApprovalRequired      bool      `json:"userApprovalRequired"`
	PortfolioApprovalRequired bool      `json:"portfolioApprovalRequired"`
	CreatedAt                 time.Time `json:"createdAt"`
	UpdatedAt                 time.Time `json:"updatedAt"`
}

// ToResponse 转换为响应结构
func (s *AdminSettings) ToResponse() AdminSettingsResponse {
	return AdminSettingsResponse{
		ID:                        s.ID,
		UserApprovalRequired:      s.UserApprovalRequired,
		PortfolioApprovalRequired: s.PortfolioApprovalRequired,
		CreatedAt:                 s.CreatedAt,
		UpdatedAt:                 s.UpdatedAt,
	}
}

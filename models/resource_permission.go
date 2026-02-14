package models

import (
	"time"
)

// ResourcePermission 资源权限表
type ResourcePermission struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	UserID       uint       `gorm:"index" json:"user_id"`
	ResourceType string     `gorm:"index;size:50" json:"resource_type"` // documents, models
	ResourceID   uint       `gorm:"index" json:"resource_id"`
	Permission   string     `gorm:"size:50" json:"permission"` // read, write, delete
	
	// 授权信息
	GrantedBy    uint       `json:"granted_by"` // 授权人ID
	ExpiresAt    *time.Time `json:"expires_at,omitempty"` // 过期时间
	
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	
	// 关联
	User         User       `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// TableName 指定表名
func (ResourcePermission) TableName() string {
	return "resource_permissions"
}

// IsExpired 是否已过期
func (rp *ResourcePermission) IsExpired() bool {
	if rp.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*rp.ExpiresAt)
}

package models

import (
	"time"
)

// PermissionGroup 权限组表
type PermissionGroup struct {
	ID          uint         `gorm:"primaryKey" json:"id"`
	Code        string       `gorm:"uniqueIndex;size:100" json:"code"` // document_manager
	Name        string       `gorm:"size:100" json:"name"`
	Description string       `gorm:"size:500" json:"description,omitempty"`
	IsSystem    bool         `gorm:"default:false;index" json:"is_system"`
	
	// 关联
	Permissions []Permission `gorm:"many2many:permission_group_permissions;" json:"permissions,omitempty"`
	
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

// TableName 指定表名
func (PermissionGroup) TableName() string {
	return "permission_groups"
}

// 系统权限组常量
const (
	PermGroupDocumentViewer   = "document_viewer"
	PermGroupDocumentManager  = "document_manager"
	PermGroupModelManager     = "model_manager"
	PermGroupAssetManager     = "asset_manager"
	PermGroupResourceViewer   = "resource_viewer"
	PermGroupResourceEditor   = "resource_editor"
	PermGroupSystemAdmin      = "system_admin"
)

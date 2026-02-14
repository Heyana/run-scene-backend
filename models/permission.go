package models

import (
	"time"
)

// Permission 权限表
type Permission struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Code        string    `gorm:"uniqueIndex;size:100" json:"code"` // documents:read
	Name        string    `gorm:"size:100" json:"name"`
	Resource    string    `gorm:"index;size:50" json:"resource"` // documents, models
	Action      string    `gorm:"index;size:50" json:"action"` // read, write, delete
	Description string    `gorm:"size:500" json:"description,omitempty"`
	IsSystem    bool      `gorm:"default:false;index" json:"is_system"`
	
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TableName 指定表名
func (Permission) TableName() string {
	return "permissions"
}

// 资源类型常量
const (
	ResourceDocuments   = "documents"
	ResourceModels      = "models"
	ResourceAssets      = "assets"
	ResourceTextures    = "textures"
	ResourceProjects    = "projects"
	ResourceAI3D        = "ai3d"
	ResourceUsers       = "users"
	ResourceRoles       = "roles"
	ResourcePermissions = "permissions"
	ResourceAll         = "*" // 所有资源
)

// 操作类型常量（使用 audit_log.go 中已定义的常量）
// ActionRead, ActionCreate, ActionUpdate, ActionDelete, ActionDownload, ActionUpload, ActionShare 已在 audit_log.go 中定义
const (
	ActionAdmin = "admin"
	ActionAll   = "*" // 所有操作
)

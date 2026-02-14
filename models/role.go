package models

import (
	"time"
)

// Role 角色表
type Role struct {
	ID               uint              `gorm:"primaryKey" json:"id"`
	Code             string            `gorm:"uniqueIndex;size:100" json:"code"` // super_admin, admin, editor
	Name             string            `gorm:"size:100" json:"name"`
	Description      string            `gorm:"size:500" json:"description,omitempty"`
	IsSystem         bool              `gorm:"default:false;index" json:"is_system"` // 系统预设不可删除
	
	// 关联
	Permissions      []Permission      `gorm:"many2many:role_permissions;" json:"permissions,omitempty"`
	PermissionGroups []PermissionGroup `gorm:"many2many:role_permission_groups;" json:"permission_groups,omitempty"`
	
	CreatedAt        time.Time         `json:"created_at"`
	UpdatedAt        time.Time         `json:"updated_at"`
}

// TableName 指定表名
func (Role) TableName() string {
	return "roles"
}

// 系统角色常量
const (
	RoleSuperAdmin = "super_admin" // 超级管理员
	RoleAdmin      = "admin"       // 管理员
	RoleEditor     = "editor"      // 编辑者
	RoleViewer     = "viewer"      // 查看者
)

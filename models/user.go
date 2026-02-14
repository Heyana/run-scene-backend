package models

import (
	"time"
)

// User 用户表
type User struct {
	ID               uint              `gorm:"primaryKey" json:"id"`
	Username         string            `gorm:"uniqueIndex;size:50" json:"username"`
	Password         string            `gorm:"size:255" json:"-"` // 不返回给前端
	Email            string            `gorm:"uniqueIndex;size:100" json:"email"`
	Phone            string            `gorm:"size:20" json:"phone,omitempty"`
	RealName         string            `gorm:"size:50" json:"real_name,omitempty"`
	Avatar           string            `gorm:"size:512" json:"avatar,omitempty"`
	
	// 状态
	Status           string            `gorm:"default:'active';index" json:"status"` // active, disabled, locked
	
	// 关联
	Roles            []Role            `gorm:"many2many:user_roles;" json:"roles,omitempty"`
	Permissions      []Permission      `gorm:"many2many:user_permissions;" json:"permissions,omitempty"`
	PermissionGroups []PermissionGroup `gorm:"many2many:user_permission_groups;" json:"permission_groups,omitempty"`
	
	// 登录信息
	LastLoginAt      *time.Time        `json:"last_login_at,omitempty"`
	LastLoginIP      string            `gorm:"size:50" json:"last_login_ip,omitempty"`
	LoginFailCount   int               `gorm:"default:0" json:"-"` // 登录失败次数
	LockedUntil      *time.Time        `json:"-"` // 锁定到期时间
	
	CreatedAt        time.Time         `json:"created_at"`
	UpdatedAt        time.Time         `json:"updated_at"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

// 用户状态常量
const (
	UserStatusActive   = "active"   // 正常
	UserStatusDisabled = "disabled" // 禁用
	UserStatusLocked   = "locked"   // 锁定（登录失败过多）
)

// IsActive 是否激活状态
func (u *User) IsActive() bool {
	return u.Status == UserStatusActive
}

// IsLocked 是否被锁定
func (u *User) IsLocked() bool {
	if u.Status == UserStatusLocked && u.LockedUntil != nil {
		return time.Now().Before(*u.LockedUntil)
	}
	return false
}

// UserResponse 用户响应（不包含敏感信息）
type UserResponse struct {
	ID          uint       `json:"id"`
	Username    string     `json:"username"`
	Email       string     `json:"email"`
	Phone       string     `json:"phone,omitempty"`
	RealName    string     `json:"real_name,omitempty"`
	Avatar      string     `json:"avatar,omitempty"`
	Status      string     `json:"status"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

// ToResponse 转换为响应格式
func (u *User) ToResponse() *UserResponse {
	return &UserResponse{
		ID:          u.ID,
		Username:    u.Username,
		Email:       u.Email,
		Phone:       u.Phone,
		RealName:    u.RealName,
		Avatar:      u.Avatar,
		Status:      u.Status,
		LastLoginAt: u.LastLoginAt,
		CreatedAt:   u.CreatedAt,
	}
}

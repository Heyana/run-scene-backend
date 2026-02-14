package requirement

import (
	"time"
	parentModels "go_wails_project_manager/models"
)

// Project 项目
type Project struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	CompanyID   uint       `gorm:"not null;index" json:"company_id"`
	Name        string     `gorm:"size:100;not null" json:"name"`
	Key         string     `gorm:"size:20;not null;uniqueIndex:idx_company_key" json:"key"` // 项目标识，如 PRJ
	Description string     `gorm:"type:text" json:"description"`
	OwnerID     uint       `gorm:"not null;index" json:"owner_id"`
	Status      string     `gorm:"size:20;default:active;index" json:"status"` // active/archived
	StartDate   *time.Time `json:"start_date"`
	EndDate     *time.Time `json:"end_date"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`

	Company      Company           `gorm:"foreignKey:CompanyID" json:"company,omitempty"`
	Owner        parentModels.User `gorm:"foreignKey:OwnerID" json:"owner,omitempty"`
	Members      []ProjectMember   `gorm:"foreignKey:ProjectID" json:"members,omitempty"`
	MissionLists []MissionList     `gorm:"foreignKey:ProjectID" json:"mission_lists,omitempty"`
}

// ProjectMember 项目成员
type ProjectMember struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ProjectID uint      `gorm:"not null;index:idx_project_user" json:"project_id"`
	UserID    uint      `gorm:"not null;index:idx_project_user" json:"user_id"`
	Role      string    `gorm:"size:50;not null" json:"role"` // project_admin/developer/viewer
	JoinedAt  time.Time `json:"joined_at"`

	Project Project           `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
	User    parentModels.User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// TableName 指定表名
func (Project) TableName() string {
	return "requirement_projects"
}

// TableName 指定表名
func (ProjectMember) TableName() string {
	return "requirement_project_members"
}

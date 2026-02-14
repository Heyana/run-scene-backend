package requirement

import (
	"time"
	parentModels "go_wails_project_manager/models"
)

// Company 公司/组织
type Company struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:100;not null" json:"name"`
	Logo        string    `gorm:"size:500" json:"logo"`
	Description string    `gorm:"type:text" json:"description"`
	OwnerID     uint      `gorm:"not null;index" json:"owner_id"`
	Status      string    `gorm:"size:20;default:active;index" json:"status"` // active/archived
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	Owner   parentModels.User `gorm:"foreignKey:OwnerID" json:"owner,omitempty"`
	Members []CompanyMember   `gorm:"foreignKey:CompanyID" json:"members,omitempty"`
	Projects []Project        `gorm:"foreignKey:CompanyID" json:"projects,omitempty"`
}

// CompanyMember 公司成员
type CompanyMember struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CompanyID uint      `gorm:"not null;index:idx_company_user" json:"company_id"`
	UserID    uint      `gorm:"not null;index:idx_company_user" json:"user_id"`
	Role      string    `gorm:"size:50;not null" json:"role"` // company_admin/member/viewer
	JoinedAt  time.Time `json:"joined_at"`

	Company Company           `gorm:"foreignKey:CompanyID" json:"company,omitempty"`
	User    parentModels.User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// TableName 指定表名
func (Company) TableName() string {
	return "requirement_companies"
}

// TableName 指定表名
func (CompanyMember) TableName() string {
	return "requirement_company_members"
}

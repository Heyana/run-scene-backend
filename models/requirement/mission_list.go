package requirement

import "time"

// MissionList 任务列表/迭代（看板列）
type MissionList struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	ProjectID    uint       `gorm:"not null;index" json:"project_id"`
	Name         string     `gorm:"size:100;not null" json:"name"`
	Type         string     `gorm:"size:20;default:sprint" json:"type"` // sprint/version/module
	Description  string     `gorm:"type:text" json:"description"`
	Color        string     `gorm:"size:20;default:#1890ff" json:"color"` // 列颜色
	StartDate    *time.Time `json:"start_date"`
	EndDate      *time.Time `json:"end_date"`
	Status       string     `gorm:"size:20;default:planning;index" json:"status"` // planning/active/completed
	SortOrder    int        `gorm:"default:0" json:"sort_order"`
	MissionCount int        `gorm:"-" json:"mission_count,omitempty"` // 任务数量（不存数据库）
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`

	Project  Project   `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
	Missions []Mission `gorm:"foreignKey:MissionListID" json:"missions,omitempty"`
}

// TableName 指定表名
func (MissionList) TableName() string {
	return "requirement_mission_lists"
}

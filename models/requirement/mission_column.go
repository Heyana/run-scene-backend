package requirement

import "time"

// MissionColumn 任务列（看板列）
type MissionColumn struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	MissionListID uint      `gorm:"not null;index" json:"mission_list_id"`
	Name          string    `gorm:"size:100;not null" json:"name"`
	Color         string    `gorm:"size:20;default:#1890ff" json:"color"` // 列的颜色
	SortOrder     int       `gorm:"default:0" json:"sort_order"`           // 排序
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`

	MissionList MissionList `gorm:"foreignKey:MissionListID" json:"mission_list,omitempty"`
	Missions    []Mission   `gorm:"foreignKey:MissionColumnID" json:"missions,omitempty"`
}

// TableName 指定表名
func (MissionColumn) TableName() string {
	return "requirement_mission_columns"
}

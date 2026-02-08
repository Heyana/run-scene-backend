package models

import "time"

// Activity 活动记录表
type Activity struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Type      string    `gorm:"size:20;index" json:"type"`      // texture/project/model/asset
	Name      string    `gorm:"size:200" json:"name"`
	Action    string    `gorm:"size:20" json:"action"`          // upload/update/delete/version_upload
	User      string    `gorm:"size:100" json:"user"`
	Version   string    `gorm:"size:20" json:"version,omitempty"` // 项目版本号（可选）
	CreatedAt time.Time `gorm:"index" json:"created_at"`
}

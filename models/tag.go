package models

import "time"

// Tag 标签表
type Tag struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"uniqueIndex;size:100" json:"name"`
	Type      string    `gorm:"size:20;index" json:"type"`
	UseCount  int       `gorm:"default:0;index" json:"use_count"`
	CreatedAt time.Time `json:"created_at"`
}

// TextureTag 材质标签关联表
type TextureTag struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	TextureID uint      `gorm:"index" json:"texture_id"`
	TagID     uint      `gorm:"index" json:"tag_id"`
	CreatedAt time.Time `json:"created_at"`
}

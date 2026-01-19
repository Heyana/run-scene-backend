// Package models 定义媒体管理器的数据模型
package models

import (
	"time"

	"gorm.io/gorm"
)

// BaseModel 所有模型的基础结构
type BaseModel struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// MediaLibrary 媒体库表
type MediaLibrary struct {
	BaseModel
	Name        string `gorm:"size:200;not null" json:"name"`
	Path        string `gorm:"size:512;not null" json:"path"` // 媒体库路径
	Description string `gorm:"type:text" json:"description"`
	IsCurrent   bool   `gorm:"default:false" json:"is_current"` // 是否为当前活动库
}

// MediaFile 媒体文件表
type MediaFile struct {
	BaseModel
	LibraryID uint   `gorm:"index;not null" json:"library_id"` // 所属媒体库
	FileName  string `gorm:"size:255;not null" json:"file_name"`
	FilePath  string `gorm:"size:512;not null" json:"file_path"` // 文件完整路径
	FileSize  int64  `json:"file_size"`
	MimeType  string `gorm:"size:100" json:"mime_type"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	Duration  int    `json:"duration"`                       // 视频时长（秒）
	HashCode  string `gorm:"size:64;index" json:"hash_code"` // 文件哈希值，用于去重

	// 关联
	Library MediaLibrary `gorm:"foreignKey:LibraryID" json:"library,omitempty"`
}

// MediaTag 媒体标签表
type MediaTag struct {
	BaseModel
	Name  string `gorm:"size:100;not null;uniqueIndex" json:"name"`
	Color string `gorm:"size:20" json:"color"` // 标签颜色
}

// MediaFileTag 媒体文件-标签关联表
type MediaFileTag struct {
	BaseModel
	MediaFileID uint `gorm:"not null;index:idx_media_tag" json:"media_file_id"`
	TagID       uint `gorm:"not null;index:idx_media_tag" json:"tag_id"`

	// 关联
	MediaFile MediaFile `gorm:"foreignKey:MediaFileID" json:"media_file,omitempty"`
	Tag       MediaTag  `gorm:"foreignKey:TagID" json:"tag,omitempty"`
}

// SystemConfig 系统配置表
type SystemConfig struct {
	BaseModel
	ConfigKey   string `gorm:"uniqueIndex;size:100;not null" json:"config_key"`
	ConfigValue string `gorm:"type:text" json:"config_value"`
	Description string `gorm:"type:text" json:"description"`
}

// ResourceFile 资源文件表（通用文件管理）
type ResourceFile struct {
	BaseModel
	LibraryID   *uint  `gorm:"index" json:"library_id"`              // 可选：关联媒体库
	Type        string `gorm:"size:20;not null" json:"type"`         // image, video, audio, document, text, file
	StorageType string `gorm:"size:10;not null" json:"storage_type"` // db, file, cdn
	FileName    string `gorm:"size:255;not null" json:"file_name"`
	FileSize    int64  `json:"file_size"`
	MimeType    string `gorm:"size:100" json:"mime_type"`
	FilePath    string `gorm:"size:512" json:"file_path"`      // 相对路径或CDN URL
	Content     []byte `gorm:"type:blob" json:"-"`             // 小文件直接存储内容
	Metadata    string `gorm:"type:json" json:"metadata"`      // JSON元数据
	HashCode    string `gorm:"size:64;index" json:"hash_code"` // 文件内容哈希，用于去重

	// 关联（可选）
	Library *MediaLibrary `gorm:"foreignKey:LibraryID" json:"library,omitempty"`
}

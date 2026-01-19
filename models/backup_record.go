// Package models 备份记录模型
package models

import (
	"time"

	"gorm.io/gorm"
)

// BackupRecord 备份记录模型
type BackupRecord struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	BackupType   string    `gorm:"type:varchar(20);not null;comment:备份类型(database/cdn)" json:"backup_type"`
	BackupTime   time.Time `gorm:"not null;comment:备份时间" json:"backup_time"`
	Status       string    `gorm:"type:varchar(20);not null;default:uploading;comment:状态(uploading/success/failed)" json:"status"`
	FilePath     string    `gorm:"type:varchar(500);comment:本地文件路径" json:"file_path"`
	RemotePath   string    `gorm:"type:varchar(500);comment:远程文件路径" json:"remote_path"`
	FileSize     int64     `gorm:"comment:文件大小(字节)" json:"file_size"`
	MD5Hash      string    `gorm:"type:varchar(32);comment:文件MD5值" json:"md5_hash"`
	Duration     int       `gorm:"comment:备份耗时(秒)" json:"duration"`
	ErrorMessage string    `gorm:"type:text;comment:错误信息" json:"error_message"`
	Environment  string    `gorm:"type:varchar(20);comment:环境标识" json:"environment"`
}

// TableName 指定表名
func (BackupRecord) TableName() string {
	return "backup_records"
}

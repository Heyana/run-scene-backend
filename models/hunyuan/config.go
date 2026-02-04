package hunyuan

import (
	"time"
)

// HunyuanConfig 混元3D配置表
type HunyuanConfig struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UpdatedAt time.Time `json:"updatedAt"`

	// API配置
	SecretID  string `gorm:"size:200;not null" json:"secretId"`
	SecretKey string `gorm:"size:200;not null" json:"secretKey"`
	Region    string `gorm:"size:50;default:'ap-guangzhou'" json:"region"`

	// 默认参数
	DefaultModel        string `gorm:"size:10;default:'3.1'" json:"defaultModel"`
	DefaultFaceCount    int    `gorm:"default:500000" json:"defaultFaceCount"`
	DefaultGenerateType string `gorm:"size:20;default:'Normal'" json:"defaultGenerateType"`
	DefaultEnablePBR    bool   `gorm:"default:true" json:"defaultEnablePbr"`
	DefaultResultFormat string `gorm:"size:20;default:'GLB'" json:"defaultResultFormat"`

	// 任务控制
	MaxConcurrent int `gorm:"default:3" json:"maxConcurrent"`
	PollInterval  int `gorm:"default:5" json:"pollInterval"` // 秒

	// 存储配置
	LocalStorageEnabled bool   `gorm:"default:true" json:"localStorageEnabled"`
	StorageDir          string `gorm:"size:200;default:'static/hunyuan'" json:"storageDir"`
	BaseURL             string `gorm:"size:200" json:"baseUrl"`
	NASEnabled          bool   `gorm:"default:false" json:"nasEnabled"`
	NASPath             string `gorm:"size:200" json:"nasPath"`
	DefaultCategory     string `gorm:"size:100;default:'AI生成'" json:"defaultCategory"`
}

// TableName 指定表名
func (HunyuanConfig) TableName() string {
	return "hunyuan_config"
}

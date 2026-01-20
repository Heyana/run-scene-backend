package models

import "time"

// Texture 材质表
type Texture struct {
	ID                uint       `gorm:"primaryKey" json:"id"`
	AssetID           string     `gorm:"uniqueIndex;size:100" json:"asset_id"`
	Name              string     `gorm:"size:200;index" json:"name"`
	Description       string     `gorm:"type:text" json:"description"`
	Type              int        `json:"type"`
	Authors           string     `gorm:"type:text" json:"authors"`
	MaxResolution     string     `gorm:"size:50" json:"max_resolution"`
	FilesHash         string     `gorm:"size:100;index" json:"files_hash"`
	DatePublished     int64      `gorm:"index" json:"date_published"`
	DownloadCount     int        `json:"download_count"`
	UseCount          int        `gorm:"default:0;index" json:"use_count"`
	LastUsedAt        *time.Time `json:"last_used_at"`
	Priority          int        `gorm:"default:0;index" json:"priority"`
	SyncStatus        int        `gorm:"default:0;index" json:"sync_status"`
	DownloadCompleted bool       `gorm:"default:false;index" json:"download_completed"` // 是否已完成下载
	Source            string     `gorm:"size:20;index;default:'polyhaven'" json:"source"` // 数据来源: polyhaven, ambientcg
	TextureTypes      string     `gorm:"type:text" json:"texture_types"`                // 包含的贴图类型，逗号分隔，如: "Diffuse,Rough,Normal"
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

// TextureFile 材质文件关联表
type TextureFile struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	TextureID  uint      `gorm:"index" json:"texture_id"`
	FileID     uint      `gorm:"index" json:"file_id"`
	MapType    string    `gorm:"size:50;index" json:"map_type"`
	Resolution string    `gorm:"size:20" json:"resolution"`
	CreatedAt  time.Time `json:"created_at"`
}

// TextureSyncLog 同步记录表
type TextureSyncLog struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	SyncType       string    `gorm:"size:20" json:"sync_type"`
	Status         int       `gorm:"default:0;index" json:"status"`
	TotalCount     int       `json:"total_count"`
	ProcessedCount int       `json:"processed_count"`
	SuccessCount   int       `json:"success_count"`
	FailCount      int       `json:"fail_count"`
	SkipCount      int       `json:"skip_count"`
	CurrentAsset   string    `gorm:"size:100" json:"current_asset"`
	Progress       float64   `json:"progress"`
	DownloadSpeed  float64   `json:"download_speed"`
	StartTime      time.Time `json:"start_time"`
	EndTime        time.Time `json:"end_time"`
	ErrorMsg       string    `gorm:"type:text" json:"error_msg"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// DownloadQueue 下载队列表
type DownloadQueue struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	FileID      uint       `gorm:"index" json:"file_id"`
	TextureID   uint       `gorm:"index" json:"texture_id"`
	Priority    int        `gorm:"default:5;index" json:"priority"`
	Status      int        `gorm:"default:0;index" json:"status"`
	RetryCount  int        `gorm:"default:0" json:"retry_count"`
	MaxRetry    int        `gorm:"default:3" json:"max_retry"`
	ErrorMsg    string     `gorm:"type:text" json:"error_msg"`
	ScheduledAt time.Time  `gorm:"index" json:"scheduled_at"`
	StartedAt   *time.Time `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// TextureMetrics 系统指标表
type TextureMetrics struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	Date            string    `gorm:"uniqueIndex;size:10" json:"date"`
	TotalTextures   int       `json:"total_textures"`
	TotalFiles      int       `json:"total_files"`
	TotalSize       int64     `json:"total_size"`
	DownloadCount   int       `json:"download_count"`
	DownloadSize    int64     `json:"download_size"`
	FailedCount     int       `json:"failed_count"`
	AvgDownloadTime float64   `json:"avg_download_time"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

package models

import (
	"go_wails_project_manager/config"
	"strings"
	"time"
	
	"gorm.io/gorm"
)

// Asset 资产表
type Asset struct {
	ID            uint       `gorm:"primaryKey" json:"id"`
	Name          string     `gorm:"size:200;index" json:"name"`
	Description   string     `gorm:"type:text" json:"description"`
	Category      string     `gorm:"size:50;index" json:"category"`
	Tags          string     `gorm:"type:text" json:"tags"` // 逗号分隔
	Type          string     `gorm:"size:20;index" json:"type"` // image, video
	
	// 文件信息
	FileSize      int64      `json:"file_size"` // 字节
	FilePath      string     `gorm:"size:512" json:"file_path"` // 资产文件相对路径
	FileHash      string     `gorm:"size:64;index" json:"file_hash"` // MD5
	Format        string     `gorm:"size:20" json:"format"` // jpg, png, webp, mp4, webm
	
	// 预览图
	ThumbnailPath string     `gorm:"size:512" json:"thumbnail_path"` // 缩略图路径
	
	// 使用统计
	UseCount      int        `gorm:"default:0;index" json:"use_count"`
	LastUsedAt    *time.Time `json:"last_used_at"`
	
	// 上传信息
	UploadedBy    string     `gorm:"size:100" json:"uploaded_by"`
	UploadIP      string     `gorm:"size:50" json:"upload_ip"`
	
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	
	// 虚拟字段，不存储到数据库
	FileURL       string     `gorm:"-" json:"file_url"`
	ThumbnailURL  string     `gorm:"-" json:"thumbnail_url"`
}

// AfterFind GORM 钩子：查询后自动拼接完整 URL
func (a *Asset) AfterFind(tx *gorm.DB) error {
	// 拼接资产文件 URL
	if a.FilePath != "" {
		a.FileURL = buildAssetURL(a.FilePath)
	}
	
	// 拼接缩略图 URL
	if a.ThumbnailPath != "" {
		a.ThumbnailURL = buildAssetURL(a.ThumbnailPath)
	}
	
	return nil
}

// buildAssetURL 构建资产文件的完整 URL
func buildAssetURL(path string) string {
	// 如果已经是完整 URL（以 http 开头），直接使用
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return path
	}
	
	// 统一将反斜杠转换为正斜杠
	path = strings.ReplaceAll(path, "\\", "/")
	
	// 否则拼接 base_url + path
	if config.AppConfig != nil && config.AppConfig.Asset.BaseURL != "" {
		// 确保 base_url 和 path 之间有且只有一个斜杠
		baseURL := strings.TrimSuffix(config.AppConfig.Asset.BaseURL, "/")
		filePath := strings.TrimPrefix(path, "/")
		// 移除 "static/assets/" 前缀（如果存在）
		filePath = strings.TrimPrefix(filePath, "static/assets/")
		return baseURL + "/" + filePath
	}
	
	// 如果没有配置 base_url，使用相对路径
	filePath := strings.TrimPrefix(path, "/")
	filePath = strings.TrimPrefix(filePath, "static/assets/")
	return "/assets/" + filePath
}

// AssetMetadata 资产元数据表
type AssetMetadata struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	AssetID   uint      `gorm:"uniqueIndex" json:"asset_id"`
	
	// 图片元数据
	Width     int       `json:"width,omitempty"`
	Height    int       `json:"height,omitempty"`
	ColorMode string    `gorm:"size:20" json:"color_mode,omitempty"` // RGB, RGBA, Grayscale
	
	// 视频元数据
	Duration  float64   `json:"duration,omitempty"` // 秒
	FrameRate float64   `json:"frame_rate,omitempty"`
	Codec     string    `gorm:"size:50" json:"codec,omitempty"`
	Bitrate   int       `json:"bitrate,omitempty"` // kbps
	
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// AssetTag 资产-标签关联表
type AssetTag struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	AssetID   uint      `gorm:"index" json:"asset_id"`
	TagID     uint      `gorm:"index" json:"tag_id"` // 复用 tags 表
	CreatedAt time.Time `json:"created_at"`
}

// AssetMetrics 资产统计指标表
type AssetMetrics struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Date        string    `gorm:"uniqueIndex:idx_date_type;size:10" json:"date"` // YYYY-MM-DD
	Type        string    `gorm:"uniqueIndex:idx_date_type;size:20" json:"type"` // image, video
	TotalAssets int       `json:"total_assets"`
	TotalSize   int64     `json:"total_size"` // 字节
	UploadCount int       `json:"upload_count"` // 当日上传数
	UploadSize  int64     `json:"upload_size"` // 当日上传大小
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

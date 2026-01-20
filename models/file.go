package models

import (
	"go_wails_project_manager/config"
	"strings"
	"time"

	"gorm.io/gorm"
)

// File 通用文件表
type File struct {
	ID            uint       `gorm:"primaryKey" json:"id"`
	FileType      string     `gorm:"size:20;index" json:"file_type"`
	RelatedID     uint       `gorm:"index" json:"related_id"`
	RelatedType   string     `gorm:"size:50" json:"related_type"`
	OriginalURL   string     `gorm:"size:500" json:"original_url"`
	LocalPath     string     `gorm:"size:500;index" json:"local_path"`
	CDNPath       string     `gorm:"size:500" json:"cdn_path"`
	FileName      string     `gorm:"size:200" json:"file_name"`
	FileSize      int64      `json:"file_size"`
	Width         int        `json:"width"`
	Height        int        `json:"height"`
	Format        string     `gorm:"size:10" json:"format"`
	MD5           string     `gorm:"size:50;index" json:"md5"`
	Version       int        `gorm:"default:1" json:"version"`
	Status        int        `gorm:"default:0;index" json:"status"`
	DownloadRetry int        `gorm:"default:0" json:"download_retry"`
	LastError     string     `gorm:"type:text" json:"last_error"`
	TextureType   string     `gorm:"size:50;index" json:"texture_type"` // 贴图类型：Diffuse, Rough, Normal 等
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	DeletedAt     *time.Time `gorm:"index" json:"deleted_at"`
	
	// 虚拟字段，不存储到数据库
	FullURL       string     `gorm:"-" json:"full_url"`
}

// AfterFind GORM 钩子：查询后自动拼接完整 URL
func (f *File) AfterFind(tx *gorm.DB) error {
	// 如果 CDNPath 不为空，拼接完整 URL
	if f.CDNPath != "" {
		// 如果 CDNPath 已经是完整 URL（以 http 开头），直接使用
		if strings.HasPrefix(f.CDNPath, "http://") || strings.HasPrefix(f.CDNPath, "https://") {
			f.FullURL = f.CDNPath
			return nil
		}
		
		// 否则拼接 base_url + cdn_path
		if config.AppConfig != nil && config.AppConfig.Texture.BaseURL != "" {
			// 确保 base_url 和 cdn_path 之间有且只有一个斜杠
			baseURL := strings.TrimSuffix(config.AppConfig.Texture.BaseURL, "/")
			cdnPath := strings.TrimPrefix(f.CDNPath, "/")
			f.FullURL = baseURL + "/" + cdnPath
			return nil
		}
		
		// 如果没有配置 base_url，使用相对路径
		f.FullURL = "/textures/" + strings.TrimPrefix(f.CDNPath, "/")
		return nil
	}
	
	// 如果 CDNPath 为空，尝试使用 OriginalURL
	if f.OriginalURL != "" {
		f.FullURL = f.OriginalURL
		return nil
	}
	
	// 最后回退到空字符串
	f.FullURL = ""
	return nil
}

// ExtractTextureType 从文件名提取贴图类型
// 例如: "nor_dx_2k.jpg" -> "nor_dx"
//      "Diffuse_2k.jpg" -> "Diffuse"
//      "Rough_4k.jpg" -> "Rough"
func ExtractTextureType(fileName string) string {
	// 移除扩展名
	name := fileName
	if idx := strings.LastIndex(name, "."); idx != -1 {
		name = name[:idx]
	}
	
	// 移除常见的分辨率后缀
	resolutionSuffixes := []string{"_1k", "_2k", "_4k", "_8k", "_16k", "_1K", "_2K", "_4K", "_8K", "_16K"}
	for _, suffix := range resolutionSuffixes {
		if strings.HasSuffix(name, suffix) {
			name = strings.TrimSuffix(name, suffix)
			break
		}
	}
	
	// 移除最后一个下划线（如果存在）
	if idx := strings.LastIndex(name, "_"); idx != -1 {
		// 检查下划线后面是否是数字（可能是其他格式的分辨率）
		afterUnderscore := name[idx+1:]
		isNumber := true
		for _, c := range afterUnderscore {
			if c < '0' || c > '9' {
				isNumber = false
				break
			}
		}
		if isNumber {
			name = name[:idx]
		}
	}
	
	return name
}

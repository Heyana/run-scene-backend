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
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	DeletedAt     *time.Time `gorm:"index" json:"deleted_at"`
	
	// 虚拟字段，不存储到数据库
	FullURL       string     `gorm:"-" json:"full_url"`
}

// AfterFind GORM 钩子：查询后自动拼接完整 URL
// 注意：GORM v2 需要接收 *gorm.DB 参数
func (f *File) AfterFind(tx *gorm.DB) error {
	// 如果启用了 NAS，优先使用 NAS 路径
	if config.AppConfig != nil && config.AppConfig.Texture.NASEnabled && config.AppConfig.Texture.NASPath != "" {
		var relativePath string
		
		// 1. 优先从 CDN 路径中提取相对路径
		if f.CDNPath != "" {
			// CDN 路径格式: http://192.168.3.39:23359/textures/stone_floor/arm_2k.jpg
			// 提取 stone_floor/arm_2k.jpg 部分
			parts := strings.Split(f.CDNPath, "/textures/")
			if len(parts) > 1 {
				relativePath = parts[1]
			}
		}
		
		// 2. 如果 CDN 路径没有，从本地路径提取
		if relativePath == "" && f.LocalPath != "" {
			// 将 Windows 路径分隔符转换为正斜杠
			cleanPath := strings.ReplaceAll(f.LocalPath, "\\", "/")
			
			// 移除开头的 ./ 或 /
			cleanPath = strings.TrimPrefix(cleanPath, "./")
			cleanPath = strings.TrimPrefix(cleanPath, "/")
			
			// 移除 "static/" 或 "static/textures/" 前缀
			cleanPath = strings.TrimPrefix(cleanPath, "static/textures/")
			cleanPath = strings.TrimPrefix(cleanPath, "static/")
			cleanPath = strings.TrimPrefix(cleanPath, "textures/")
			
			relativePath = cleanPath
		}
		
		// 3. 拼接 NAS 路径
		if relativePath != "" {
			// 不直接返回 NAS 路径，而是返回通过后端代理的 HTTP 路径
			// 这样浏览器可以正常访问
			f.FullURL = "/textures/" + relativePath
			return nil
		}
	}
	
	// 如果没有启用 NAS，使用 CDN 路径
	if f.CDNPath != "" {
		f.FullURL = f.CDNPath
		return nil
	}
	
	// 回退到本地路径
	if f.LocalPath != "" {
		cleanPath := strings.ReplaceAll(f.LocalPath, "\\", "/")
		cleanPath = strings.TrimPrefix(cleanPath, "./")
		cleanPath = strings.TrimPrefix(cleanPath, "/")
		cleanPath = strings.TrimPrefix(cleanPath, "static/textures/")
		cleanPath = strings.TrimPrefix(cleanPath, "static/")
		cleanPath = strings.TrimPrefix(cleanPath, "textures/")
		
		// 回退到默认的网络共享路径
		f.FullURL = "//192.168.3.10/project/editor_v2/static/textures/" + cleanPath
		return nil
	}
	
	// 最后使用原始 URL
	if f.OriginalURL != "" {
		f.FullURL = f.OriginalURL
	}
	
	return nil
}

package models

import (
	"go_wails_project_manager/config"
	"strings"
	"time"
	
	"gorm.io/gorm"
)

// Project 项目表
type Project struct {
	ID              uint       `gorm:"primaryKey" json:"id"`
	Name            string     `gorm:"size:200;uniqueIndex" json:"name"`
	Description     string     `gorm:"type:text" json:"description"`
	CurrentVersion  string     `gorm:"size:20" json:"current_version"`
	LatestVersionID uint       `json:"latest_version_id"`
	
	// 缩略图路径（最新版本的截图）
	ThumbnailPath   string     `gorm:"size:512" json:"thumbnail_path"`
	
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	
	// 虚拟字段
	ThumbnailURL    string     `gorm:"-" json:"thumbnail_url"`
	
	// 关联
	Versions        []ProjectVersion `gorm:"foreignKey:ProjectID" json:"versions,omitempty"`
}

// ProjectVersion 项目版本表
type ProjectVersion struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	ProjectID   uint       `gorm:"index" json:"project_id"`
	Version     string     `gorm:"size:20;index" json:"version"`
	Username    string     `gorm:"size:100" json:"username"`
	Description string     `gorm:"type:text" json:"description"`
	
	// 文件信息
	FilePath    string     `gorm:"size:512" json:"file_path"`
	FileSize    int64      `json:"file_size"`
	FileHash    string     `gorm:"size:64" json:"file_hash"`
	
	// 元数据
	FileCount   int        `json:"file_count"`
	UploadIP    string     `gorm:"size:50" json:"upload_ip"`
	
	// 解压后的目录路径（用于预览）
	ExtractedPath string   `gorm:"size:512" json:"extracted_path"`
	
	// 截图路径
	ThumbnailPath string   `gorm:"size:512" json:"thumbnail_path"`
	
	CreatedAt   time.Time  `json:"created_at"`
	
	// 虚拟字段
	FileURL      string     `gorm:"-" json:"file_url"`
	PreviewURL   string     `gorm:"-" json:"preview_url"`
	ThumbnailURL string     `gorm:"-" json:"thumbnail_url"`
}

// AfterFind 查询后钩子 - Project
func (p *Project) AfterFind(tx *gorm.DB) error {
	// 构建缩略图URL
	if p.ThumbnailPath != "" {
		p.ThumbnailURL = buildProjectURL(p.ThumbnailPath)
	}
	
	return nil
}

// AfterFind 查询后钩子 - ProjectVersion
func (pv *ProjectVersion) AfterFind(tx *gorm.DB) error {
	if pv.FilePath != "" {
		pv.FileURL = buildProjectURL(pv.FilePath)
	}
	
	// 构建预览URL
	if pv.ExtractedPath != "" {
		pv.PreviewURL = buildPreviewURL(pv.ExtractedPath)
	}
	
	// 构建缩略图URL
	if pv.ThumbnailPath != "" {
		pv.ThumbnailURL = buildProjectURL(pv.ThumbnailPath)
	}
	
	return nil
}

// buildProjectURL 构建项目文件的完整 URL
func buildProjectURL(path string) string {
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return path
	}
	
	// 统一路径分隔符
	path = strings.ReplaceAll(path, "\\", "/")
	
	// 提取相对路径（移除NAS路径前缀）
	relativePath := extractProjectRelativePath(path)
	
	if config.ProjectAppConfig != nil && config.ProjectAppConfig.BaseURL != "" {
		baseURL := strings.TrimSuffix(config.ProjectAppConfig.BaseURL, "/")
		return baseURL + "/" + relativePath
	}
	
	return "/projects/" + relativePath
}

// buildPreviewURL 构建预览URL
func buildPreviewURL(extractedPath string) string {
	if extractedPath == "" {
		return ""
	}
	
	// 统一路径分隔符
	path := strings.ReplaceAll(extractedPath, "\\", "/")
	
	// 提取相对路径
	relativePath := extractProjectRelativePath(path)
	
	if config.ProjectAppConfig != nil && config.ProjectAppConfig.BaseURL != "" {
		baseURL := strings.TrimSuffix(config.ProjectAppConfig.BaseURL, "/")
		return baseURL + "/" + relativePath + "/index.html"
	}
	
	return "/projects/" + relativePath + "/index.html"
}

// extractProjectRelativePath 从完整路径提取相对路径
// 例如: \\192.168.3.10\project\editor_v2\static\projects\123\v1.0.1.zip -> 123/v1.0.1.zip
//      static/projects/123/v1.0.1.zip -> 123/v1.0.1.zip
func extractProjectRelativePath(fullPath string) string {
	// 统一为正斜杠
	path := strings.ReplaceAll(fullPath, "\\", "/")
	
	// 查找 projects/ 的位置
	projectsIndex := strings.LastIndex(path, "/projects/")
	if projectsIndex != -1 {
		// 提取 projects/ 之后的部分
		return path[projectsIndex+10:] // 跳过 "/projects/"
	}
	
	// 如果没找到，尝试查找 projects\ 的位置（Windows路径）
	projectsIndex = strings.LastIndex(path, "projects/")
	if projectsIndex != -1 {
		return path[projectsIndex+9:] // 跳过 "projects/"
	}
	
	// 如果都没找到，返回文件名
	parts := strings.Split(path, "/")
	if len(parts) >= 2 {
		// 返回最后两段路径（项目名/文件名）
		return parts[len(parts)-2] + "/" + parts[len(parts)-1]
	}
	
	return path
}

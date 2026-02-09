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
	
	// 项目类型：upload（上传文件）或 external（外部链接）
	ProjectType     string     `gorm:"size:20;default:upload" json:"project_type"`
	
	// 外部链接（当 project_type 为 external 时使用）
	ExternalURL     string     `gorm:"size:512" json:"external_url"`
	
	// 缩略图路径（最新版本的截图）
	ThumbnailPath   string     `gorm:"size:512" json:"thumbnail_path"`
	
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	
	// 虚拟字段
	ThumbnailURL    string     `gorm:"-" json:"thumbnail_url"`
	PreviewURL      string     `gorm:"-" json:"preview_url"` // 预览 URL（外部链接或最新版本）
	
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
	
	// 历史版本解压路径
	HistoryPath   string   `gorm:"size:512" json:"history_path"`
	
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
	
	// 构建预览URL
	if p.ProjectType == "external" {
		// 外部链接项目，直接使用外部 URL
		p.PreviewURL = p.ExternalURL
	} else {
		// 上传文件项目，构建最新版本的预览 URL
		if p.Name != "" {
			p.PreviewURL = buildProjectPreviewURL(p.Name)
		}
	}
	
	return nil
}

// AfterFind 查询后钩子 - ProjectVersion
func (pv *ProjectVersion) AfterFind(tx *gorm.DB) error {
	if pv.FilePath != "" {
		pv.FileURL = buildProjectURL(pv.FilePath)
	}
	
	// 构建预览URL
	// 优先使用 HistoryPath（历史版本），如果没有则使用 ExtractedPath（当前版本）
	previewPath := pv.HistoryPath
	if previewPath == "" {
		previewPath = pv.ExtractedPath
	}
	
	if previewPath != "" {
		pv.PreviewURL = buildPreviewURL(previewPath)
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
	
	// 如果是相对路径，直接使用
	relativePath := path
	
	// 如果是绝对路径，提取相对路径
	if strings.Contains(path, "/static/") || strings.Contains(path, "\\static\\") {
		relativePath = extractProjectRelativePath(path)
	}
	
	if config.ProjectAppConfig != nil && config.ProjectAppConfig.BaseURL != "" {
		baseURL := strings.TrimSuffix(config.ProjectAppConfig.BaseURL, "/")
		
		// 判断相对路径是否已经包含 projects/ 或 project_histories/
		if strings.HasPrefix(relativePath, "projects/") {
			// 已经包含 projects/，需要去掉 baseURL 中的 /projects
			baseURLWithoutProjects := strings.TrimSuffix(baseURL, "/projects")
			return baseURLWithoutProjects + "/" + relativePath
		} else if strings.HasPrefix(relativePath, "project_histories/") {
			// 已经包含 project_histories/
			baseURLWithoutProjects := strings.TrimSuffix(baseURL, "/projects")
			return baseURLWithoutProjects + "/" + relativePath
		} else {
			// 不包含前缀，直接拼接
			return baseURL + "/" + relativePath
		}
	}
	
	// 如果没有配置 baseURL，判断是否需要添加前缀
	if strings.HasPrefix(relativePath, "projects/") || strings.HasPrefix(relativePath, "project_histories/") {
		return "/" + relativePath
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
	
	// 判断是当前版本还是历史版本
	// 相对路径格式:
	// - 当前版本: projects/项目名称/
	// - 历史版本: project_histories/项目名称/v1.0.0/extracted/
	
	var relativePath string
	var urlPrefix string
	
	// 如果是绝对路径，先提取相对路径
	if strings.Contains(path, "/static/") || strings.Contains(path, "\\static\\") {
		staticIdx := strings.LastIndex(path, "/static/")
		if staticIdx == -1 {
			staticIdx = strings.LastIndex(path, "static/")
			if staticIdx != -1 {
				path = path[staticIdx+7:] // 跳过 "static/"
			}
		} else {
			path = path[staticIdx+8:] // 跳过 "/static/"
		}
	}
	
	if strings.HasPrefix(path, "project_histories/") || strings.Contains(path, "/project_histories/") {
		// 历史版本
		urlPrefix = "/project_histories"
		// 提取 project_histories/ 之后的部分
		idx := strings.Index(path, "project_histories/")
		if idx != -1 {
			relativePath = path[idx+18:] // 跳过 "project_histories/"
		} else {
			relativePath = path
		}
	} else {
		// 当前版本
		urlPrefix = "/projects"
		// 提取 projects/ 之后的部分
		idx := strings.Index(path, "projects/")
		if idx != -1 {
			relativePath = path[idx+9:] // 跳过 "projects/"
		} else {
			relativePath = path
		}
	}
	
	if config.ProjectAppConfig != nil && config.ProjectAppConfig.BaseURL != "" {
		baseURL := strings.TrimSuffix(config.ProjectAppConfig.BaseURL, "/")
		// 替换 /projects 为实际的 baseURL
		if urlPrefix == "/projects" {
			return baseURL + "/" + relativePath + "/index.html"
		} else {
			// 历史版本使用不同的 baseURL
			historyBaseURL := strings.Replace(baseURL, "/projects", "/project_histories", 1)
			return historyBaseURL + "/" + relativePath + "/index.html"
		}
	}

	return urlPrefix + "/" + relativePath + "/index.html"
}

// buildProjectPreviewURL 构建项目的预览 URL（当前版本）
func buildProjectPreviewURL(projectName string) string {
	if config.ProjectAppConfig != nil && config.ProjectAppConfig.BaseURL != "" {
		baseURL := strings.TrimSuffix(config.ProjectAppConfig.BaseURL, "/")
		return baseURL + "/" + projectName + "/index.html"
	}
	
	return "/projects/" + projectName + "/index.html"
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

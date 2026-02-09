package models

import (
	"go_wails_project_manager/config"
	"strings"
	"time"

	"gorm.io/gorm"
)

// Document 文档主表（统一文件和文件夹）
type Document struct {
	ID            uint       `gorm:"primaryKey" json:"id"`
	Name          string     `gorm:"size:200;index" json:"name"`
	Description   string     `gorm:"type:text" json:"description"`
	Category      string     `gorm:"size:50;index" json:"category"`
	Tags          string     `gorm:"type:text" json:"tags"` // 逗号分隔
	Type          string     `gorm:"size:20;index" json:"type"` // folder, document, video, archive, other

	// 文件夹支持
	ParentID      *uint      `gorm:"index" json:"parent_id"`        // 父文件夹ID，null表示根目录
	IsFolder      bool       `gorm:"default:false;index" json:"is_folder"` // 是否是文件夹
	ChildCount    int        `gorm:"default:0" json:"child_count"`  // 子项数量（文件+文件夹）

	// 文件信息（仅文件有效，相对于 static 目录的路径）
	FileSize      int64  `json:"file_size,omitempty"`                      // 字节
	FilePath      string `gorm:"size:512" json:"file_path,omitempty"`      // 相对路径：documents/2026/02/09/1/file.pdf
	FileHash      string `gorm:"size:64;index" json:"file_hash,omitempty"` // MD5
	Format        string `gorm:"size:20" json:"format,omitempty"`          // pdf, docx, mp4, zip

	// 预览（相对路径，仅文件有效）
	ThumbnailPath string `gorm:"size:512" json:"thumbnail_path,omitempty"` // documents/2026/02/09/1/thumbnail.jpg
	PreviewPath   string `gorm:"size:512" json:"preview_path,omitempty"`   // documents/2026/02/09/1/preview/

	// 版本控制（仅文件有效）
	Version  string `gorm:"size:20" json:"version,omitempty"`
	ParentVersionID *uint  `json:"parent_version_id,omitempty"` // 父版本ID
	IsLatest bool   `gorm:"default:true;index" json:"is_latest"`

	// 权限
	Department string `gorm:"size:50;index" json:"department"`
	Project    string `gorm:"size:100;index" json:"project"`
	IsPublic   bool   `gorm:"default:false" json:"is_public"`

	// 统计
	DownloadCount int        `gorm:"default:0;index" json:"download_count"`
	ViewCount     int        `gorm:"default:0" json:"view_count"`
	LastViewedAt  *time.Time `json:"last_viewed_at,omitempty"`

	// 上传信息
	UploadedBy string `gorm:"size:100" json:"uploaded_by"`
	UploadIP   string `gorm:"size:50" json:"upload_ip"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// 虚拟字段，不存储到数据库
	FileURL      string `gorm:"-" json:"file_url,omitempty"`
	ThumbnailURL string `gorm:"-" json:"thumbnail_url,omitempty"`
	PreviewURL   string `gorm:"-" json:"preview_url,omitempty"`
}

// AfterFind GORM 钩子：查询后自动拼接完整 URL
func (d *Document) AfterFind(tx *gorm.DB) error {
	// 只有文件才拼接 URL
	if !d.IsFolder {
		// 拼接文件 URL
		if d.FilePath != "" {
			d.FileURL = buildDocumentURL(d.FilePath)
		}

		// 拼接缩略图 URL
		if d.ThumbnailPath != "" {
			d.ThumbnailURL = buildDocumentURL(d.ThumbnailPath)
		}

		// 拼接预览 URL
		if d.PreviewPath != "" {
			d.PreviewURL = buildDocumentURL(d.PreviewPath)
		}
	}

	return nil
}

// AfterCreate GORM 钩子：创建后更新父文件夹统计
func (d *Document) AfterCreate(tx *gorm.DB) error {
	if d.ParentID != nil {
		return updateParentChildCount(tx, *d.ParentID)
	}
	return nil
}

// AfterDelete GORM 钩子：删除后更新父文件夹统计
func (d *Document) AfterDelete(tx *gorm.DB) error {
	if d.ParentID != nil {
		return updateParentChildCount(tx, *d.ParentID)
	}
	return nil
}

// updateParentChildCount 更新父文件夹的子项数量
func updateParentChildCount(tx *gorm.DB, parentID uint) error {
	var count int64
	if err := tx.Model(&Document{}).Where("parent_id = ?", parentID).Count(&count).Error; err != nil {
		return err
	}
	return tx.Model(&Document{}).Where("id = ?", parentID).Update("child_count", count).Error
}

// buildDocumentURL 构建文档文件的完整 URL
func buildDocumentURL(path string) string {
	// 如果已经是完整 URL（以 http 开头），直接使用
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return path
	}

	// 统一将反斜杠转换为正斜杠
	path = strings.ReplaceAll(path, "\\", "/")

	// 获取文件库配置
	docConfig, err := config.LoadDocumentConfig()
	if err == nil && docConfig.BaseURL != "" {
		// 确保 base_url 和 path 之间有且只有一个斜杠
		baseURL := strings.TrimSuffix(docConfig.BaseURL, "/")
		filePath := strings.TrimPrefix(path, "/")
		// 移除 "static/documents/" 前缀（如果存在）
		filePath = strings.TrimPrefix(filePath, "static/documents/")
		filePath = strings.TrimPrefix(filePath, "documents/")
		return baseURL + "/" + filePath
	}

	// 如果没有配置 base_url，使用相对路径
	filePath := strings.TrimPrefix(path, "/")
	filePath = strings.TrimPrefix(filePath, "static/documents/")
	filePath = strings.TrimPrefix(filePath, "documents/")
	return "/documents/" + filePath
}

// DocumentMetadata 文档元数据表
type DocumentMetadata struct {
	ID         uint   `gorm:"primaryKey" json:"id"`
	DocumentID uint   `gorm:"uniqueIndex" json:"document_id"`

	// 文档元数据
	PageCount int    `json:"page_count,omitempty"` // PDF页数
	Author    string `gorm:"size:100" json:"author,omitempty"`
	Title     string `gorm:"size:200" json:"title,omitempty"`
	Subject   string `gorm:"size:200" json:"subject,omitempty"`

	// 视频元数据
	Duration  float64 `json:"duration,omitempty"`
	Width     int     `json:"width,omitempty"`
	Height    int     `json:"height,omitempty"`
	FrameRate float64 `json:"frame_rate,omitempty"`
	Codec     string  `gorm:"size:50" json:"codec,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// DocumentAccessLog 访问日志表
type DocumentAccessLog struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	DocumentID uint      `gorm:"index" json:"document_id"`
	Action     string    `gorm:"size:20;index" json:"action"` // view, download, upload, delete, create_folder
	UserName   string    `gorm:"size:100;index" json:"user_name"`
	UserIP     string    `gorm:"size:50" json:"user_ip"`
	CreatedAt  time.Time `gorm:"index" json:"created_at"`
}

// DocumentMetrics 统计指标表
type DocumentMetrics struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	Date          string    `gorm:"uniqueIndex:idx_document_date_type;size:10" json:"date"` // YYYY-MM-DD
	Type          string    `gorm:"uniqueIndex:idx_document_date_type;size:20" json:"type"` // folder, document, video, archive, other
	TotalDocs     int       `json:"total_docs"`
	TotalSize     int64     `json:"total_size"`
	UploadCount   int       `json:"upload_count"`
	DownloadCount int       `json:"download_count"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// 文档类型常量
const (
	TypeFolder   = "folder"   // 文件夹
	TypeDocument = "document" // PDF, Word, Excel, PPT, TXT
	TypeVideo    = "video"    // MP4, WebM, AVI, MOV
	TypeArchive  = "archive"  // ZIP, RAR, 7Z
	TypeOther    = "other"    // 其他类型
)

// FormatTypeMap 格式到类型的映射
var FormatTypeMap = map[string]string{
	"pdf":  TypeDocument,
	"doc":  TypeDocument,
	"docx": TypeDocument,
	"xls":  TypeDocument,
	"xlsx": TypeDocument,
	"ppt":  TypeDocument,
	"pptx": TypeDocument,
	"txt":  TypeDocument,
	"md":   TypeDocument,

	"mp4":  TypeVideo,
	"webm": TypeVideo,
	"avi":  TypeVideo,
	"mov":  TypeVideo,

	"zip": TypeArchive,
	"rar": TypeArchive,
	"7z":  TypeArchive,
}

// GetDocumentType 根据格式获取文档类型
func GetDocumentType(format string) string {
	if docType, ok := FormatTypeMap[strings.ToLower(format)]; ok {
		return docType
	}
	return TypeOther
}

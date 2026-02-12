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
	
	// 递归统计（仅文件夹有效）
	TotalSize     int64      `gorm:"default:0" json:"total_size"`   // 递归统计所有子文件的总大小（字节）
	TotalCount    int        `gorm:"default:0" json:"total_count"`  // 递归统计所有子文件的总数量（不含文件夹）
	StatsUpdatedAt *time.Time `json:"stats_updated_at,omitempty"`   // 统计信息更新时间

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
	FileURL           string   `gorm:"-" json:"file_url,omitempty"`
	ThumbnailURL      string   `gorm:"-" json:"thumbnail_url,omitempty"`
	PreviewURL        string   `gorm:"-" json:"preview_url,omitempty"`
	FolderThumbnails  []string `gorm:"-" json:"folder_thumbnails,omitempty"` // 文件夹缩略图（前4个文件）
}

// AfterFind GORM 钩子：查询后自动拼接完整 URL
func (d *Document) AfterFind(tx *gorm.DB) error {
	// 文件：拼接文件 URL
	if !d.IsFolder {
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
	// 注意：FolderThumbnails 已在 loadFolderThumbnails 中处理，无需在此处理

	return nil
}

// AfterCreate GORM 钩子：创建后更新父文件夹统计
func (d *Document) AfterCreate(tx *gorm.DB) error {
	if d.ParentID != nil {
		// 更新直接父文件夹的子项数量
		if err := updateParentChildCount(tx, *d.ParentID); err != nil {
			return err
		}
		// 更新所有祖先文件夹的递归统计
		if !d.IsFolder {
			return updateAncestorStats(tx, *d.ParentID, d.FileSize, 1)
		}
	}
	return nil
}

// AfterDelete GORM 钩子：删除后更新父文件夹统计
func (d *Document) AfterDelete(tx *gorm.DB) error {
	if d.ParentID != nil {
		// 更新直接父文件夹的子项数量
		if err := updateParentChildCount(tx, *d.ParentID); err != nil {
			return err
		}
		// 更新所有祖先文件夹的递归统计
		if !d.IsFolder {
			return updateAncestorStats(tx, *d.ParentID, -d.FileSize, -1)
		} else {
			// 如果删除的是文件夹，需要减去该文件夹的统计
			return updateAncestorStats(tx, *d.ParentID, -d.TotalSize, -d.TotalCount)
		}
	}
	return nil
}

// BeforeUpdate GORM 钩子：更新前处理（用于移动文件）
func (d *Document) BeforeUpdate(tx *gorm.DB) error {
	// 检查 parent_id 是否变化（文件移动）
	if tx.Statement.Changed("ParentID") {
		var oldDoc Document
		if err := tx.Model(&Document{}).Where("id = ?", d.ID).First(&oldDoc).Error; err != nil {
			return err
		}
		
		// 如果 parent_id 发生变化，需要更新新旧父文件夹的统计
		if oldDoc.ParentID != d.ParentID {
			// 从旧父文件夹减去
			if oldDoc.ParentID != nil {
				if !d.IsFolder {
					updateAncestorStats(tx, *oldDoc.ParentID, -d.FileSize, -1)
				} else {
					updateAncestorStats(tx, *oldDoc.ParentID, -d.TotalSize, -d.TotalCount)
				}
			}
			
			// 向新父文件夹添加
			if d.ParentID != nil {
				if !d.IsFolder {
					updateAncestorStats(tx, *d.ParentID, d.FileSize, 1)
				} else {
					updateAncestorStats(tx, *d.ParentID, d.TotalSize, d.TotalCount)
				}
			}
		}
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

// updateAncestorStats 递归更新所有祖先文件夹的统计信息
func updateAncestorStats(tx *gorm.DB, folderID uint, sizeChange int64, countChange int) error {
	now := time.Now()
	
	// 更新当前文件夹
	if err := tx.Model(&Document{}).Where("id = ? AND is_folder = ?", folderID, true).
		Updates(map[string]interface{}{
			"total_size":       gorm.Expr("total_size + ?", sizeChange),
			"total_count":      gorm.Expr("total_count + ?", countChange),
			"stats_updated_at": now,
		}).Error; err != nil {
		return err
	}
	
	// 查找父文件夹
	var folder Document
	if err := tx.Select("parent_id").Where("id = ?", folderID).First(&folder).Error; err != nil {
		return err
	}
	
	// 如果有父文件夹，递归更新
	if folder.ParentID != nil {
		return updateAncestorStats(tx, *folder.ParentID, sizeChange, countChange)
	}
	
	return nil
}

// RecalculateFolderStats 重新计算文件夹的递归统计（用于修复数据或手动刷新）
func RecalculateFolderStats(tx *gorm.DB, folderID uint) error {
	var totalSize int64
	var totalCount int64
	
	// 递归计算所有子文件的大小和数量
	if err := calculateFolderStatsRecursive(tx, folderID, &totalSize, &totalCount); err != nil {
		return err
	}
	
	// 更新文件夹统计
	now := time.Now()
	return tx.Model(&Document{}).Where("id = ?", folderID).Updates(map[string]interface{}{
		"total_size":       totalSize,
		"total_count":      totalCount,
		"stats_updated_at": now,
	}).Error
}

// calculateFolderStatsRecursive 递归计算文件夹统计
func calculateFolderStatsRecursive(tx *gorm.DB, folderID uint, totalSize *int64, totalCount *int64) error {
	// 查询直接子项
	var children []Document
	if err := tx.Select("id", "is_folder", "file_size").
		Where("parent_id = ?", folderID).
		Find(&children).Error; err != nil {
		return err
	}
	
	for _, child := range children {
		if child.IsFolder {
			// 递归计算子文件夹
			if err := calculateFolderStatsRecursive(tx, child.ID, totalSize, totalCount); err != nil {
				return err
			}
		} else {
			// 累加文件大小和数量
			*totalSize += child.FileSize
			*totalCount++
		}
	}
	
	return nil
}

// RecalculateAllFolderStats 重新计算所有文件夹的统计（用于数据修复）
func RecalculateAllFolderStats(tx *gorm.DB) error {
	// 查询所有文件夹
	var folders []Document
	if err := tx.Select("id").Where("is_folder = ?", true).Find(&folders).Error; err != nil {
		return err
	}
	
	// 从最深层开始计算（先计算子文件夹，再计算父文件夹）
	// 这里简化处理，直接遍历所有文件夹
	for _, folder := range folders {
		if err := RecalculateFolderStats(tx, folder.ID); err != nil {
			return err
		}
	}
	
	return nil
}

// BuildDocumentURL 构建文档文件的完整 URL（导出供其他包使用）
func BuildDocumentURL(path string) string {
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

// buildDocumentURL 内部使用的别名（保持向后兼容）
func buildDocumentURL(path string) string {
	return BuildDocumentURL(path)
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

package document

import (
	"fmt"
	"go_wails_project_manager/models"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gorm.io/gorm"
)

// QueryService 查询服务
type QueryService struct {
	db *gorm.DB
}

// NewQueryService 创建查询服务
func NewQueryService(db *gorm.DB) *QueryService {
	return &QueryService{
		db: db,
	}
}

// QueryFilters 查询过滤器
type QueryFilters struct {
	Type       string   // folder, document, video, archive, other
	Category   string
	Tags       []string
	Format     string
	Department string
	Project    string
	IsPublic   *bool
	Keyword    string
	SortBy     string // name, created_at, download_count, file_size
	SortOrder  string // asc, desc
	ParentID   *uint  // 父文件夹ID，nil表示查询所有，0表示根目录
}

// List 分页查询
func (q *QueryService) List(page, pageSize int, filters QueryFilters) ([]*models.Document, int64, error) {
	query := q.db.Model(&models.Document{}).Where("is_latest = ?", true)

	// 父文件夹过滤
	if filters.ParentID != nil {
		if *filters.ParentID == 0 {
			// 查询根目录
			query = query.Where("parent_id IS NULL")
		} else {
			// 查询指定文件夹下的内容
			query = query.Where("parent_id = ?", *filters.ParentID)
		}
	}

	// 应用过滤条件
	if filters.Type != "" {
		query = query.Where("type = ?", filters.Type)
	}

	if filters.Category != "" {
		query = query.Where("category = ?", filters.Category)
	}

	if filters.Format != "" {
		query = query.Where("format = ?", filters.Format)
	}

	if filters.Department != "" {
		query = query.Where("department = ?", filters.Department)
	}

	if filters.Project != "" {
		query = query.Where("project = ?", filters.Project)
	}

	if filters.IsPublic != nil {
		query = query.Where("is_public = ?", *filters.IsPublic)
	}

	if filters.Keyword != "" {
		query = query.Where("name LIKE ? OR description LIKE ?",
			"%"+filters.Keyword+"%", "%"+filters.Keyword+"%")
	}

	// 标签过滤
	if len(filters.Tags) > 0 {
		for _, tag := range filters.Tags {
			query = query.Where("tags LIKE ?", "%"+tag+"%")
		}
	}

	// 排序
	sortBy := "created_at"
	if filters.SortBy != "" {
		sortBy = filters.SortBy
	}

	sortOrder := "DESC"
	if filters.SortOrder == "asc" {
		sortOrder = "ASC"
	}

	query = query.Order(fmt.Sprintf("%s %s", sortBy, sortOrder))

	// 统计总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	var documents []*models.Document
	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Find(&documents).Error
	
	if err != nil {
		return nil, 0, err
	}

	// 为文件夹查询前4个文件的缩略图（性能优化：批量查询）
	q.loadFolderThumbnails(documents)

	return documents, total, err
}

// loadFolderThumbnails 批量加载文件夹缩略图（性能优化）
func (q *QueryService) loadFolderThumbnails(documents []*models.Document) {
	// 收集所有文件夹ID
	folderIDs := make([]uint, 0)
	for _, doc := range documents {
		if doc.IsFolder {
			folderIDs = append(folderIDs, doc.ID)
		}
	}
	
	if len(folderIDs) == 0 {
		return
	}
	
	// 批量查询所有文件夹的前4个缩略图（一次查询，性能优化）
	type FolderThumbnail struct {
		ParentID      uint
		ThumbnailPath string
	}
	
	var thumbnails []FolderThumbnail
	q.db.Model(&models.Document{}).
		Select("parent_id, thumbnail_path").
		Where("parent_id IN ? AND is_folder = ? AND thumbnail_path IS NOT NULL AND thumbnail_path != ''", folderIDs, false).
		Order("created_at DESC").
		Limit(len(folderIDs) * 4). // 每个文件夹最多4个
		Find(&thumbnails)
	
	// 按文件夹ID分组，并转换为完整URL
	thumbnailMap := make(map[uint][]string)
	for _, thumb := range thumbnails {
		if len(thumbnailMap[thumb.ParentID]) < 4 {
			// 调用 buildDocumentURL 转换为完整 URL
			fullURL := models.BuildDocumentURL(thumb.ThumbnailPath)
			thumbnailMap[thumb.ParentID] = append(thumbnailMap[thumb.ParentID], fullURL)
		}
	}
	
	// 将缩略图赋值给文件夹（使用 FolderThumbnails 字段）
	for _, doc := range documents {
		if doc.IsFolder {
			if thumbs, ok := thumbnailMap[doc.ID]; ok {
				doc.FolderThumbnails = thumbs
			}
		}
	}
}

// GetDetail 获取详情
func (q *QueryService) GetDetail(id uint) (*models.Document, *models.DocumentMetadata, error) {
	var document models.Document
	if err := q.db.First(&document, id).Error; err != nil {
		return nil, nil, err
	}

	var metadata models.DocumentMetadata
	q.db.Where("document_id = ?", id).First(&metadata)

	return &document, &metadata, nil
}

// GetVersions 获取版本列表
func (q *QueryService) GetVersions(documentID uint) ([]*models.Document, error) {
	var document models.Document
	if err := q.db.First(&document, documentID).Error; err != nil {
		return nil, err
	}

	var versions []*models.Document

	// 如果是最新版本，查找所有相关版本
	if document.IsLatest {
		// 查找所有父版本
		q.db.Where("id = ? OR parent_id = ?", documentID, documentID).
			Order("created_at DESC").
			Find(&versions)
	} else {
		// 如果不是最新版本，查找最新版本和所有相关版本
		if document.ParentID != nil {
			q.db.Where("id = ? OR parent_id = ?", *document.ParentID, *document.ParentID).
				Order("created_at DESC").
				Find(&versions)
		} else {
			versions = append(versions, &document)
		}
	}

	return versions, nil
}

// Delete 删除文档（硬删除，包括物理文件）
func (q *QueryService) Delete(id uint) error {
	var document models.Document
	if err := q.db.First(&document, id).Error; err != nil {
		return err
	}

	// 如果是文件夹，递归删除所有子项
	if document.IsFolder {
		var children []*models.Document
		if err := q.db.Where("parent_id = ?", id).Find(&children).Error; err != nil {
			return fmt.Errorf("查询子项失败: %w", err)
		}

		// 递归删除所有子项
		for _, child := range children {
			if err := q.Delete(child.ID); err != nil {
				return fmt.Errorf("删除子项失败 (ID: %d): %w", child.ID, err)
			}
		}
	}

	// 删除物理文件
	if document.FilePath != "" && !document.IsFolder {
		// FilePath 格式: static/documents/2026/02/09/123/file.zip
		// 提取目录路径: static/documents/2026/02/09/123
		// 使用正斜杠分割（数据库中存储的是正斜杠）
		parts := strings.Split(document.FilePath, "/")
		if len(parts) >= 2 {
			// 重新组合为目录路径（去掉最后的文件名）
			dirPath := filepath.Join(parts[:len(parts)-1]...)
			
			// 删除整个目录（包含文件和可能的其他资源）
			if err := os.RemoveAll(dirPath); err != nil && !os.IsNotExist(err) {
				// 文件删除失败只记录警告，不阻止数据库删除
				fmt.Printf("警告: 删除文件目录失败 %s: %v\n", dirPath, err)
			} else {
				fmt.Printf("已删除文件目录: %s\n", dirPath)
			}
		}
	}

	// 删除缩略图
	if document.ThumbnailPath != "" {
		if err := os.Remove(document.ThumbnailPath); err != nil && !os.IsNotExist(err) {
			fmt.Printf("警告: 删除缩略图失败 %s: %v\n", document.ThumbnailPath, err)
		}
	}

	// 删除预览
	if document.PreviewPath != "" {
		if err := os.RemoveAll(document.PreviewPath); err != nil && !os.IsNotExist(err) {
			fmt.Printf("警告: 删除预览失败 %s: %v\n", document.PreviewPath, err)
		}
	}

	// 删除元数据
	q.db.Unscoped().Where("document_id = ?", id).Delete(&models.DocumentMetadata{})

	// 删除访问日志
	q.db.Unscoped().Where("document_id = ?", id).Delete(&models.DocumentAccessLog{})

	// 硬删除文档记录（使用 Unscoped 跳过软删除）
	return q.db.Unscoped().Delete(&document).Error
}

// Update 更新文档信息
func (q *QueryService) Update(id uint, updates map[string]interface{}) error {
	return q.db.Model(&models.Document{}).Where("id = ?", id).Updates(updates).Error
}

// IncrementDownloadCount 增加下载次数
func (q *QueryService) IncrementDownloadCount(id uint) error {
	return q.db.Model(&models.Document{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"download_count": gorm.Expr("download_count + 1"),
		}).Error
}

// IncrementViewCount 增加查看次数
func (q *QueryService) IncrementViewCount(id uint) error {
	now := time.Now()
	return q.db.Model(&models.Document{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"view_count":     gorm.Expr("view_count + 1"),
			"last_viewed_at": now,
		}).Error
}

// GetStatistics 获取统计信息
func (q *QueryService) GetStatistics(filters QueryFilters) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	query := q.db.Model(&models.Document{}).Where("is_latest = ?", true)

	// 应用过滤条件
	if filters.Type != "" {
		query = query.Where("type = ?", filters.Type)
	}
	if filters.Department != "" {
		query = query.Where("department = ?", filters.Department)
	}
	if filters.Project != "" {
		query = query.Where("project = ?", filters.Project)
	}

	// 总数和总大小
	var totalDocs int64
	var totalSize int64
	query.Count(&totalDocs)
	query.Select("COALESCE(SUM(file_size), 0)").Scan(&totalSize)

	stats["total_documents"] = totalDocs
	stats["total_size"] = totalSize

	// 按类型统计
	var typeStats []struct {
		Type  string
		Count int64
	}
	q.db.Model(&models.Document{}).
		Where("is_latest = ?", true).
		Select("type, COUNT(*) as count").
		Group("type").
		Scan(&typeStats)

	typeDistribution := make(map[string]int64)
	for _, stat := range typeStats {
		typeDistribution[stat.Type] = stat.Count
	}
	stats["type_distribution"] = typeDistribution

	// 按格式统计
	var formatStats []struct {
		Format string
		Count  int64
	}
	q.db.Model(&models.Document{}).
		Where("is_latest = ?", true).
		Select("format, COUNT(*) as count").
		Group("format").
		Scan(&formatStats)

	formatDistribution := make(map[string]int64)
	for _, stat := range formatStats {
		formatDistribution[stat.Format] = stat.Count
	}
	stats["format_distribution"] = formatDistribution

	// 最近上传数（7天内）
	sevenDaysAgo := time.Now().AddDate(0, 0, -7)
	var recentUploads int64
	q.db.Model(&models.Document{}).
		Where("created_at >= ?", sevenDaysAgo).
		Count(&recentUploads)
	stats["recent_uploads"] = recentUploads

	return stats, nil
}

// GetPopular 获取热门文档
func (q *QueryService) GetPopular(limit int, filters QueryFilters) ([]*models.Document, error) {
	query := q.db.Model(&models.Document{}).
		Where("is_latest = ?", true).
		Order("download_count DESC")

	if filters.Type != "" {
		query = query.Where("type = ?", filters.Type)
	}
	if filters.Department != "" {
		query = query.Where("department = ?", filters.Department)
	}
	if filters.Project != "" {
		query = query.Where("project = ?", filters.Project)
	}

	var documents []*models.Document
	err := query.Limit(limit).Find(&documents).Error

	return documents, err
}

// GetAccessLogs 获取访问日志
func (q *QueryService) GetAccessLogs(documentID uint, page, pageSize int, action string) ([]*models.DocumentAccessLog, int64, error) {
	query := q.db.Model(&models.DocumentAccessLog{}).Where("document_id = ?", documentID)

	if action != "" {
		query = query.Where("action = ?", action)
	}

	// 统计总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	var logs []*models.DocumentAccessLog
	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&logs).Error

	return logs, total, err
}


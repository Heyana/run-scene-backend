package document

import (
	"fmt"
	"go_wails_project_manager/models"
	"os"
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
	Type       string   // document, video, archive, other
	Category   string
	Tags       []string
	Format     string
	Department string
	Project    string
	IsPublic   *bool
	Keyword    string
	SortBy     string // name, created_at, download_count, file_size
	SortOrder  string // asc, desc
}

// List 分页查询
func (q *QueryService) List(page, pageSize int, filters QueryFilters) ([]*models.Document, int64, error) {
	query := q.db.Model(&models.Document{}).Where("is_latest = ?", true)

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

	return documents, total, err
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

// Delete 删除文档
func (q *QueryService) Delete(id uint) error {
	var document models.Document
	if err := q.db.First(&document, id).Error; err != nil {
		return err
	}

	// 删除文件
	if document.FilePath != "" {
		os.Remove(document.FilePath)
	}

	// 删除缩略图
	if document.ThumbnailPath != "" {
		os.Remove(document.ThumbnailPath)
	}

	// 删除预览
	if document.PreviewPath != "" {
		os.RemoveAll(document.PreviewPath)
	}

	// 删除元数据
	q.db.Where("document_id = ?", id).Delete(&models.DocumentMetadata{})

	// 删除访问日志（可选，根据需求决定是否保留）
	// q.db.Where("document_id = ?", id).Delete(&models.DocumentAccessLog{})

	// 删除文档记录
	return q.db.Delete(&document).Error
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


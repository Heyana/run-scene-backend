package asset

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
	Type      string   // image, video
	Category  string
	Tags      []string
	Format    string
	Keyword   string
	SortBy    string // name, created_at, use_count, file_size
	SortOrder string // asc, desc
}

// List 分页查询
func (q *QueryService) List(page, pageSize int, filters QueryFilters) ([]*models.Asset, int64, error) {
	query := q.db.Model(&models.Asset{})
	
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
	
	if filters.Keyword != "" {
		query = query.Where("name LIKE ?", "%"+filters.Keyword+"%")
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
	var assets []*models.Asset
	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Find(&assets).Error
	
	return assets, total, err
}

// GetDetail 获取详情
func (q *QueryService) GetDetail(id uint) (*models.Asset, *models.AssetMetadata, error) {
	var asset models.Asset
	if err := q.db.First(&asset, id).Error; err != nil {
		return nil, nil, err
	}
	
	var metadata models.AssetMetadata
	q.db.Where("asset_id = ?", id).First(&metadata)
	
	return &asset, &metadata, nil
}

// ListByType 按类型查询
func (q *QueryService) ListByType(assetType string, page, pageSize int) ([]*models.Asset, int64, error) {
	return q.List(page, pageSize, QueryFilters{Type: assetType})
}

// ListByCategory 按分类查询
func (q *QueryService) ListByCategory(category string, page, pageSize int) ([]*models.Asset, int64, error) {
	return q.List(page, pageSize, QueryFilters{Category: category})
}

// Search 搜索
func (q *QueryService) Search(keyword string, page, pageSize int) ([]*models.Asset, int64, error) {
	return q.List(page, pageSize, QueryFilters{Keyword: keyword})
}

// IncrementUseCount 记录使用次数
func (q *QueryService) IncrementUseCount(id uint) error {
	now := time.Now()
	return q.db.Model(&models.Asset{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"use_count":    gorm.Expr("use_count + 1"),
			"last_used_at": now,
		}).Error
}

// Delete 删除资产
func (q *QueryService) Delete(id uint) error {
	var asset models.Asset
	if err := q.db.First(&asset, id).Error; err != nil {
		return err
	}
	
	// 删除文件
	if asset.FilePath != "" {
		os.Remove(asset.FilePath)
	}
	
	// 删除缩略图
	if asset.ThumbnailPath != "" {
		os.Remove(asset.ThumbnailPath)
	}
	
	// 删除元数据
	q.db.Where("asset_id = ?", id).Delete(&models.AssetMetadata{})
	
	// 删除标签关联
	q.db.Where("asset_id = ?", id).Delete(&models.AssetTag{})
	
	// 删除资产记录
	return q.db.Delete(&asset).Error
}

// Update 更新资产信息
func (q *QueryService) Update(id uint, updates map[string]interface{}) error {
	return q.db.Model(&models.Asset{}).Where("id = ?", id).Updates(updates).Error
}

// GetStatistics 获取统计信息
func (q *QueryService) GetStatistics(assetType string) (map[string]interface{}, error) {
	stats := make(map[string]interface{})
	
	query := q.db.Model(&models.Asset{})
	if assetType != "" {
		query = query.Where("type = ?", assetType)
	}
	
	// 总数和总大小
	var totalAssets int64
	var totalSize int64
	query.Count(&totalAssets)
	query.Select("COALESCE(SUM(file_size), 0)").Scan(&totalSize)
	
	stats["total_assets"] = totalAssets
	stats["total_size"] = totalSize
	
	// 按类型统计
	var typeStats []struct {
		Type  string
		Count int64
	}
	q.db.Model(&models.Asset{}).
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
	q.db.Model(&models.Asset{}).
		Select("format, COUNT(*) as count").
		Group("format").
		Scan(&formatStats)
	
	formatDistribution := make(map[string]int64)
	for _, stat := range formatStats {
		formatDistribution[stat.Format] = stat.Count
	}
	stats["format_distribution"] = formatDistribution
	
	// 按分类统计
	var categoryStats []struct {
		Category string
		Count    int64
	}
	q.db.Model(&models.Asset{}).
		Where("category != ''").
		Select("category, COUNT(*) as count").
		Group("category").
		Scan(&categoryStats)
	
	categoryDistribution := make(map[string]int64)
	for _, stat := range categoryStats {
		categoryDistribution[stat.Category] = stat.Count
	}
	stats["category_distribution"] = categoryDistribution
	
	// 最近上传数（7天内）
	sevenDaysAgo := time.Now().AddDate(0, 0, -7)
	var recentUploads int64
	q.db.Model(&models.Asset{}).
		Where("created_at >= ?", sevenDaysAgo).
		Count(&recentUploads)
	stats["recent_uploads"] = recentUploads
	
	return stats, nil
}

// GetPopular 获取热门资产
func (q *QueryService) GetPopular(limit int, assetType string) ([]*models.Asset, error) {
	query := q.db.Model(&models.Asset{}).Order("use_count DESC")
	
	if assetType != "" {
		query = query.Where("type = ?", assetType)
	}
	
	var assets []*models.Asset
	err := query.Limit(limit).Find(&assets).Error
	
	return assets, err
}

// GetStatisticsByType 按类型获取统计信息
func (q *QueryService) GetStatisticsByType() (map[string]interface{}, error) {
	result := make(map[string]interface{})
	
	types := []string{"image", "video"}
	
	for _, assetType := range types {
		var count int64
		var totalSize int64
		
		q.db.Model(&models.Asset{}).
			Where("type = ?", assetType).
			Count(&count)
		
		q.db.Model(&models.Asset{}).
			Where("type = ?", assetType).
			Select("COALESCE(SUM(file_size), 0)").
			Scan(&totalSize)
		
		avgSize := int64(0)
		if count > 0 {
			avgSize = totalSize / count
		}
		
		result[assetType] = map[string]interface{}{
			"count":      count,
			"total_size": totalSize,
			"avg_size":   avgSize,
		}
	}
	
	return result, nil
}

package model

import (
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
	"go_wails_project_manager/config"
	"go_wails_project_manager/logger"
	"go_wails_project_manager/models"
	"go_wails_project_manager/services/storage"
)

type QueryService struct {
	db             *gorm.DB
	storageService *storage.FileStorageService
}

type QueryFilters struct {
	Category  string
	Tags      []string
	Type      string
	SortBy    string // name, created_at, use_count
	SortOrder string // asc, desc
}

type ModelStatistics struct {
	TotalModels          int64          `json:"total_models"`
	TotalSize            int64          `json:"total_size"`
	TypeDistribution     map[string]int `json:"type_distribution"`
	CategoryDistribution map[string]int `json:"category_distribution"`
	RecentUploads        int64          `json:"recent_uploads"`
}

func NewQueryService(db *gorm.DB) *QueryService {
	// 创建存储服务配置
	storageConfig := &storage.StorageConfig{
		LocalStorageEnabled: config.AppConfig.Model.LocalStorageEnabled,
		StorageDir:          config.AppConfig.Model.StorageDir,
		NASEnabled:          config.AppConfig.Model.NASEnabled,
		NASPath:             config.AppConfig.Model.NASPath,
	}

	return &QueryService{
		db:             db,
		storageService: storage.NewFileStorageService(storageConfig, logger.Log),
	}
}

// List 分页查询
func (q *QueryService) List(page, pageSize int, filters QueryFilters) ([]*models.Model, int64, error) {
	var modelList []*models.Model
	var total int64

	query := q.db.Model(&models.Model{})

	// 应用过滤条件
	if filters.Category != "" {
		query = query.Where("category = ?", filters.Category)
	}

	if filters.Type != "" {
		query = query.Where("type = ?", filters.Type)
	}

	if len(filters.Tags) > 0 {
		// 标签过滤（简单实现：包含任一标签）
		conditions := make([]string, len(filters.Tags))
		args := make([]interface{}, len(filters.Tags))
		for i, tag := range filters.Tags {
			conditions[i] = "tags LIKE ?"
			args[i] = "%" + tag + "%"
		}
		query = query.Where(strings.Join(conditions, " OR "), args...)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 排序
	sortBy := filters.SortBy
	if sortBy == "" {
		sortBy = "created_at"
	}
	sortOrder := filters.SortOrder
	if sortOrder == "" {
		sortOrder = "desc"
	}
	query = query.Order(fmt.Sprintf("%s %s", sortBy, sortOrder))

	// 分页
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Find(&modelList).Error; err != nil {
		return nil, 0, err
	}

	return modelList, total, nil
}

// ListByCategory 按分类查询
func (q *QueryService) ListByCategory(category string, page, pageSize int) ([]*models.Model, int64, error) {
	filters := QueryFilters{
		Category: category,
	}
	return q.List(page, pageSize, filters)
}

// ListByTags 按标签查询
func (q *QueryService) ListByTags(tagNames []string, page, pageSize int) ([]*models.Model, int64, error) {
	filters := QueryFilters{
		Tags: tagNames,
	}
	return q.List(page, pageSize, filters)
}

// Search 搜索（名称、标签）
func (q *QueryService) Search(keyword string, page, pageSize int) ([]*models.Model, int64, error) {
	var modelList []*models.Model
	var total int64

	query := q.db.Model(&models.Model{}).Where(
		"name LIKE ? OR description LIKE ? OR tags LIKE ?",
		"%"+keyword+"%",
		"%"+keyword+"%",
		"%"+keyword+"%",
	)

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&modelList).Error; err != nil {
		return nil, 0, err
	}

	return modelList, total, nil
}

// GetDetail 获取详情
func (q *QueryService) GetDetail(id uint) (*models.Model, []*models.Tag, error) {
	var model models.Model
	if err := q.db.First(&model, id).Error; err != nil {
		return nil, nil, err
	}

	// 获取标签
	var tags []*models.Tag
	q.db.Table("tags").
		Joins("INNER JOIN model_tags ON tags.id = model_tags.tag_id").
		Where("model_tags.model_id = ?", id).
		Find(&tags)

	return &model, tags, nil
}

// IncrementUseCount 记录使用次数
func (q *QueryService) IncrementUseCount(id uint) error {
	now := time.Now()
	return q.db.Model(&models.Model{}).Where("id = ?", id).Updates(map[string]interface{}{
		"use_count":    gorm.Expr("use_count + 1"),
		"last_used_at": now,
	}).Error
}

// GetStatistics 获取统计信息
func (q *QueryService) GetStatistics() (*ModelStatistics, error) {
	stats := &ModelStatistics{
		TypeDistribution:     make(map[string]int),
		CategoryDistribution: make(map[string]int),
	}

	// 总数和总大小
	q.db.Model(&models.Model{}).Count(&stats.TotalModels)
	q.db.Model(&models.Model{}).Select("COALESCE(SUM(file_size), 0)").Scan(&stats.TotalSize)

	// 类型分布
	var typeStats []struct {
		Type  string
		Count int
	}
	q.db.Model(&models.Model{}).Select("type, COUNT(*) as count").Group("type").Scan(&typeStats)
	for _, stat := range typeStats {
		stats.TypeDistribution[stat.Type] = stat.Count
	}

	// 分类分布
	var categoryStats []struct {
		Category string
		Count    int
	}
	q.db.Model(&models.Model{}).
		Where("category != ''").
		Select("category, COUNT(*) as count").
		Group("category").
		Scan(&categoryStats)
	for _, stat := range categoryStats {
		stats.CategoryDistribution[stat.Category] = stat.Count
	}

	// 最近上传数（24小时内）
	yesterday := time.Now().Add(-24 * time.Hour)
	q.db.Model(&models.Model{}).Where("created_at > ?", yesterday).Count(&stats.RecentUploads)

	return stats, nil
}

// Delete 删除模型
func (q *QueryService) Delete(id uint) error {
	// 1. 查询模型
	var model models.Model
	if err := q.db.First(&model, id).Error; err != nil {
		return err
	}

	// 2. 删除文件（使用通用存储服务）
	subPath := fmt.Sprintf("%d", id)
	if err := q.storageService.DeleteFile(subPath); err != nil {
		// 文件删除失败只记录日志，不影响数据库删除
		fmt.Printf("删除文件失败: %v\n", err)
	}

	// 3. 删除数据库记录
	if err := q.db.Delete(&model).Error; err != nil {
		return err
	}

	// 4. 删除标签关联
	q.db.Where("model_id = ?", id).Delete(&models.ModelTag{})

	return nil
}

// GetPopular 获取热门模型
func (q *QueryService) GetPopular(limit int) ([]*models.Model, error) {
	var models []*models.Model
	err := q.db.Order("use_count DESC").Limit(limit).Find(&models).Error
	return models, err
}

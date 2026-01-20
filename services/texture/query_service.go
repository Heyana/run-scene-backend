package texture

import (
	"go_wails_project_manager/config"
	"go_wails_project_manager/models"
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
	return &QueryService{db: db}
}

// List 分页查询材质列表
func (s *QueryService) List(page, pageSize int, filters map[string]interface{}) ([]models.Texture, int64, error) {
	var textures []models.Texture
	var total int64

	query := s.db.Model(&models.Texture{})

	// 应用过滤条件
	if keyword, ok := filters["keyword"].(string); ok && keyword != "" {
		// 支持按名称、描述或 asset_id 搜索
		query = query.Where("name LIKE ? OR description LIKE ? OR asset_id LIKE ?", 
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}

	if syncStatus, ok := filters["sync_status"].(int); ok {
		query = query.Where("sync_status = ?", syncStatus)
	}

	// 按贴图类型筛选
	if textureType, ok := filters["texture_type"].(string); ok && textureType != "" {
		// 使用 LIKE 查询，因为 texture_types 是逗号分隔的字符串
		query = query.Where("texture_types LIKE ?", "%"+textureType+"%")
	}

	// 按 Three.js 类型筛选（需要映射到原始类型）
	if threeJSType, ok := filters["threejs_type"].(string); ok && threeJSType != "" {
		// 从配置中获取对应的原始类型列表
		if config.TextureMapping != nil && config.TextureMapping.ThreeJS != nil {
			if originalTypes, exists := config.TextureMapping.ThreeJS[threeJSType]; exists && len(originalTypes) > 0 {
				// 构建 OR 查询：texture_types LIKE '%Diffuse%' OR texture_types LIKE '%col%' ...
				orConditions := make([]string, len(originalTypes))
				args := make([]interface{}, len(originalTypes))
				for i, originalType := range originalTypes {
					orConditions[i] = "texture_types LIKE ?"
					args[i] = "%" + originalType + "%"
				}
				query = query.Where(strings.Join(orConditions, " OR "), args...)
			}
		}
	}

	// 排序
	if sortBy, ok := filters["sort_by"].(string); ok {
		switch sortBy {
		case "use_count":
			query = query.Order("use_count DESC")
		case "date_published":
			query = query.Order("date_published DESC")
		default:
			query = query.Order("created_at DESC")
		}
	} else {
		query = query.Order("created_at DESC")
	}

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Find(&textures).Error; err != nil {
		return nil, 0, err
	}

	return textures, total, nil
}

// ListByTag 按标签查询材质
func (s *QueryService) ListByTag(tagID uint, page, pageSize int) ([]models.Texture, int64, error) {
	var textures []models.Texture
	var total int64

	query := s.db.Table("texture").
		Joins("JOIN texture_tag ON texture_tag.texture_id = texture.id").
		Where("texture_tag.tag_id = ?", tagID)

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Find(&textures).Error; err != nil {
		return nil, 0, err
	}

	return textures, total, nil
}

// ListByCategory 按分类查询材质
func (s *QueryService) ListByCategory(category string, page, pageSize int) ([]models.Texture, int64, error) {
	// 先找到分类标签
	var tag models.Tag
	if err := s.db.Where("name = ? AND type = ?", category, "category").First(&tag).Error; err != nil {
		return nil, 0, err
	}

	return s.ListByTag(tag.ID, page, pageSize)
}

// Search 搜索材质
func (s *QueryService) Search(keyword string, page, pageSize int) ([]models.Texture, int64, error) {
	filters := map[string]interface{}{
		"keyword": keyword,
	}
	return s.List(page, pageSize, filters)
}

// GetDetail 获取材质详情
func (s *QueryService) GetDetail(assetID string) (*models.Texture, []models.Tag, []models.File, error) {
	var texture models.Texture
	if err := s.db.Where("asset_id = ?", assetID).First(&texture).Error; err != nil {
		return nil, nil, nil, err
	}

	// 获取标签
	var tags []models.Tag
	s.db.Table("tag").
		Joins("JOIN texture_tag ON texture_tag.tag_id = tag.id").
		Where("texture_tag.texture_id = ?", texture.ID).
		Find(&tags)

	// 获取文件
	var files []models.File
	s.db.Where("related_id = ? AND related_type = ?", texture.ID, "Texture").
		Find(&files)

	return &texture, tags, files, nil
}

// GetByAssetID 根据AssetID获取材质
func (s *QueryService) GetByAssetID(assetID string) (*models.Texture, error) {
	var texture models.Texture
	err := s.db.Where("asset_id = ?", assetID).First(&texture).Error
	return &texture, err
}

// IncrementUseCount 记录使用次数
func (s *QueryService) IncrementUseCount(textureID uint) error {
	now := time.Now()
	return s.db.Model(&models.Texture{}).Where("id = ?", textureID).
		Updates(map[string]interface{}{
			"use_count":    gorm.Expr("use_count + ?", 1),
			"last_used_at": now,
		}).Error
}

// GetSyncProgress 获取同步进度
func (s *QueryService) GetSyncProgress(logID uint) (*models.TextureSyncLog, error) {
	var log models.TextureSyncLog
	err := s.db.Where("id = ?", logID).First(&log).Error
	return &log, err
}

// GetLatestSyncLog 获取最新的同步日志
func (s *QueryService) GetLatestSyncLog() (*models.TextureSyncLog, error) {
	var log models.TextureSyncLog
	err := s.db.Order("created_at DESC").First(&log).Error
	return &log, err
}

// ListSyncLogs 获取同步日志列表
func (s *QueryService) ListSyncLogs(page, pageSize int, filters map[string]interface{}) ([]models.TextureSyncLog, int64, error) {
	var logs []models.TextureSyncLog
	var total int64

	query := s.db.Model(&models.TextureSyncLog{})

	// 应用过滤条件
	if status, ok := filters["status"].(int); ok {
		query = query.Where("status = ?", status)
	}

	if syncType, ok := filters["sync_type"].(string); ok && syncType != "" {
		query = query.Where("sync_type = ?", syncType)
	}

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// GetAllTextureTypes 获取所有唯一的贴图类型
func (s *QueryService) GetAllTextureTypes() ([]string, error) {
	var files []models.File
	
	// 查询所有不同的 texture_type
	if err := s.db.Model(&models.File{}).
		Where("file_type = ? AND texture_type != ?", "texture", "").
		Distinct("texture_type").
		Find(&files).Error; err != nil {
		return nil, err
	}

	// 提取类型列表
	types := make([]string, 0, len(files))
	for _, file := range files {
		if file.TextureType != "" {
			types = append(types, file.TextureType)
		}
	}

	return types, nil
}

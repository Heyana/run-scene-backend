package texture

import (
	"go_wails_project_manager/models"

	"gorm.io/gorm"
)

// TagService 标签服务
type TagService struct {
	db *gorm.DB
}

// NewTagService 创建标签服务
func NewTagService(db *gorm.DB) *TagService {
	return &TagService{db: db}
}

// GetOrCreateTag 创建或获取标签
func (s *TagService) GetOrCreateTag(name string, tagType string) (*models.Tag, error) {
	var tag models.Tag
	err := s.db.Where("name = ? AND type = ?", name, tagType).First(&tag).Error
	if err == gorm.ErrRecordNotFound {
		// 标签不存在，创建新标签
		tag = models.Tag{
			Name:     name,
			Type:     tagType,
			UseCount: 0,
		}
		if err := s.db.Create(&tag).Error; err != nil {
			return nil, err
		}
		return &tag, nil
	}
	if err != nil {
		return nil, err
	}
	return &tag, nil
}

// AssociateTextureTags 关联材质和标签
func (s *TagService) AssociateTextureTags(textureID uint, tagIDs []uint) error {
	// 先删除旧的关联
	if err := s.db.Where("texture_id = ?", textureID).Delete(&models.TextureTag{}).Error; err != nil {
		return err
	}

	// 创建新的关联
	for _, tagID := range tagIDs {
		textureTag := models.TextureTag{
			TextureID: textureID,
			TagID:     tagID,
		}
		if err := s.db.Create(&textureTag).Error; err != nil {
			return err
		}
	}

	return nil
}

// IncrementTagUseCount 更新标签使用次数
func (s *TagService) IncrementTagUseCount(tagID uint) error {
	return s.db.Model(&models.Tag{}).Where("id = ?", tagID).
		UpdateColumn("use_count", gorm.Expr("use_count + ?", 1)).Error
}

// GetPopularTags 获取热门标签
func (s *TagService) GetPopularTags(limit int) ([]models.Tag, error) {
	var tags []models.Tag
	err := s.db.Order("use_count DESC").Limit(limit).Find(&tags).Error
	return tags, err
}

// ListByType 按类型获取标签列表
func (s *TagService) ListByType(tagType string) ([]models.Tag, error) {
	var tags []models.Tag
	err := s.db.Where("type = ?", tagType).Order("use_count DESC").Find(&tags).Error
	return tags, err
}

// GetTextureTagIDs 获取材质的所有标签ID
func (s *TagService) GetTextureTagIDs(textureID uint) ([]uint, error) {
	var tagIDs []uint
	err := s.db.Model(&models.TextureTag{}).
		Where("texture_id = ?", textureID).
		Pluck("tag_id", &tagIDs).Error
	return tagIDs, err
}

// GetTextureTags 获取材质的所有标签
func (s *TagService) GetTextureTags(textureID uint) ([]models.Tag, error) {
	var tags []models.Tag
	err := s.db.Table("tag").
		Joins("JOIN texture_tag ON texture_tag.tag_id = tag.id").
		Where("texture_tag.texture_id = ?", textureID).
		Find(&tags).Error
	return tags, err
}

package texture

import (
	"fmt"
	"go_wails_project_manager/models"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// UnifiedDownloadService 统一下载服务（支持所有数据源）
type UnifiedDownloadService struct {
	db                       *gorm.DB
	logger                   *logrus.Logger
	polyhavenDownloadService *DownloadService
	ambientcgDownloadService *AmbientCGDownloadService
}

// NewUnifiedDownloadService 创建统一下载服务
func NewUnifiedDownloadService(db *gorm.DB, logger *logrus.Logger) *UnifiedDownloadService {
	return &UnifiedDownloadService{
		db:                       db,
		logger:                   logger,
		polyhavenDownloadService: NewDownloadService(db, logger),
		ambientcgDownloadService: NewAmbientCGDownloadService(db, logger),
	}
}

// DownloadTexture 下载材质（自动识别数据源）
func (s *UnifiedDownloadService) DownloadTexture(assetID string, opts DownloadOptions) ([]models.File, error) {
	// 1. 查询材质，获取数据源
	var texture models.Texture
	if err := s.db.Where("asset_id = ?", assetID).First(&texture).Error; err != nil {
		return nil, fmt.Errorf("材质不存在: %w", err)
	}

	// 2. 检查是否已下载
	if texture.DownloadCompleted {
		s.logInfo("材质已下载，返回现有文件: %s (来源: %s)", assetID, texture.Source)
		return s.getExistingFiles(texture.ID)
	}

	// 3. 根据数据源调用对应的下载服务
	s.logInfo("开始下载材质: %s (来源: %s)", assetID, texture.Source)

	var files []models.File
	var err error

	switch texture.Source {
	case "ambientcg":
		files, err = s.ambientcgDownloadService.DownloadTexture(assetID, opts)
	case "polyhaven", "":
		// PolyHaven 或未标记来源的（兼容旧数据）
		files, err = s.downloadPolyhavenTexture(assetID, &texture)
	default:
		return nil, fmt.Errorf("不支持的数据源: %s", texture.Source)
	}

	if err != nil {
		return nil, err
	}

	s.logInfo("材质下载完成: %s (文件数: %d)", assetID, len(files))
	return files, nil
}

// downloadPolyhavenTexture 下载 PolyHaven 材质
func (s *UnifiedDownloadService) downloadPolyhavenTexture(assetID string, texture *models.Texture) ([]models.File, error) {
	s.logInfo("下载 PolyHaven 材质: %s", assetID)

	// 获取文件列表
	syncService := GetGlobalSyncService()
	if syncService == nil {
		return nil, fmt.Errorf("同步服务未初始化")
	}

	filesDetail, err := syncService.FetchTextureFiles(assetID)
	if err != nil {
		return nil, fmt.Errorf("获取文件列表失败: %w", err)
	}

	// 下载并转换
	if err := s.polyhavenDownloadService.DownloadAndConvert(texture.ID, assetID, filesDetail); err != nil {
		return nil, fmt.Errorf("下载贴图失败: %w", err)
	}

	// 更新状态
	texture.DownloadCompleted = true
	texture.SyncStatus = 2
	if texture.Source == "" {
		texture.Source = "polyhaven" // 兼容旧数据
	}
	if err := s.db.Save(texture).Error; err != nil {
		return nil, fmt.Errorf("更新状态失败: %w", err)
	}

	// 返回文件列表
	return s.getExistingFiles(texture.ID)
}

// getExistingFiles 获取已存在的文件
func (s *UnifiedDownloadService) getExistingFiles(textureID uint) ([]models.File, error) {
	var files []models.File
	err := s.db.Where("related_id = ? AND related_type = ? AND file_type = ?",
		textureID, "Texture", "texture").Find(&files).Error
	return files, err
}

// 日志方法
func (s *UnifiedDownloadService) logInfo(format string, args ...interface{}) {
	s.logger.Infof("[Unified Download] "+format, args...)
}

func (s *UnifiedDownloadService) logError(format string, err error, args ...interface{}) {
	allArgs := append([]interface{}{err}, args...)
	s.logger.Errorf("[Unified Download] "+format, allArgs...)
}

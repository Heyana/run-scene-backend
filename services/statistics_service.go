package services

import (
	"fmt"
	"go_wails_project_manager/config"
	"go_wails_project_manager/models"
	"os"
	"time"

	"gorm.io/gorm"
)

type StatisticsService struct {
	db *gorm.DB
}

func NewStatisticsService(db *gorm.DB) *StatisticsService {
	return &StatisticsService{db: db}
}

// ResourceStats 资源统计信息
type ResourceStats struct {
	Total       int64   `json:"total"`
	Trend       float64 `json:"trend"`
	RecentCount int64   `json:"recent_count"`
}

// OverviewStats 统计概览
type OverviewStats struct {
	Textures ResourceStats `json:"textures"`
	Projects ResourceStats `json:"projects"`
	Models   ResourceStats `json:"models"`
	Assets   ResourceStats `json:"assets"`
	AI3D     ResourceStats `json:"ai3d"`
}

// StorageInfo 存储信息
type StorageInfo struct {
	Total        uint64  `json:"total"`
	Used         uint64  `json:"used"`
	UsagePercent float64 `json:"usage_percent"`
}

// SystemStatus 系统状态
type SystemStatus struct {
	Service struct {
		Status string `json:"status"`
		Uptime int64  `json:"uptime"`
	} `json:"service"`
	Database struct {
		Status string `json:"status"`
		Size   int64  `json:"size"`
	} `json:"database"`
	Storage StorageInfo `json:"storage"`
	Sync    struct {
		LastSyncAt time.Time `json:"last_sync_at"`
		Status     string    `json:"status"`
	} `json:"sync"`
}

// GetOverview 获取资源统计概览
func (ss *StatisticsService) GetOverview() (*OverviewStats, error) {
	stats := &OverviewStats{}

	// 贴图统计
	textureStats, err := ss.getResourceStats("texture")
	if err != nil {
		return nil, err
	}
	stats.Textures = textureStats

	// 项目统计
	projectStats, err := ss.getResourceStats("project")
	if err != nil {
		return nil, err
	}
	stats.Projects = projectStats

	// 模型统计
	modelStats, err := ss.getResourceStats("model")
	if err != nil {
		return nil, err
	}
	stats.Models = modelStats

	// 资产统计
	assetStats, err := ss.getResourceStats("asset")
	if err != nil {
		return nil, err
	}
	stats.Assets = assetStats

	// AI3D 统计
	ai3dStats, err := ss.getResourceStats("ai3d")
	if err != nil {
		// AI3D 统计失败不影响整体，使用空值
		ai3dStats = ResourceStats{}
	}
	stats.AI3D = ai3dStats

	return stats, nil
}

// getResourceStats 获取单个资源类型的统计
func (ss *StatisticsService) getResourceStats(resourceType string) (ResourceStats, error) {
	stats := ResourceStats{}

	// 获取总数
	total, err := ss.GetResourceCount(resourceType)
	if err != nil {
		return stats, err
	}
	stats.Total = total

	// 计算趋势
	trend, err := ss.CalculateTrend(resourceType)
	if err != nil {
		// 趋势计算失败不影响整体，设为0
		trend = 0
	}
	stats.Trend = trend

	// 获取最近7天新增数量
	recentCount, err := ss.GetRecentCount(resourceType, 7)
	if err != nil {
		recentCount = 0
	}
	stats.RecentCount = recentCount

	return stats, nil
}

// GetResourceCount 获取资源总数
func (ss *StatisticsService) GetResourceCount(resourceType string) (int64, error) {
	var count int64
	var err error

	switch resourceType {
	case "texture":
		err = ss.db.Model(&models.Texture{}).Count(&count).Error
	case "project":
		err = ss.db.Model(&models.Project{}).Count(&count).Error
	case "model":
		err = ss.db.Model(&models.Model{}).Count(&count).Error
	case "asset":
		err = ss.db.Model(&models.Asset{}).Count(&count).Error
	case "ai3d":
		// AI3D 任务统计 - 使用 Model 以支持软删除过滤
		var task struct {
			gorm.Model
		}
		err = ss.db.Model(&task).Table("ai3d_tasks").Count(&count).Error
	default:
		return 0, fmt.Errorf("未知的资源类型: %s", resourceType)
	}

	return count, err
}

// CalculateTrend 计算本月增长趋势（百分比）
func (ss *StatisticsService) CalculateTrend(resourceType string) (float64, error) {
	now := time.Now()
	
	// 本月第一天
	currentMonthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	
	// 上月第一天
	lastMonthStart := currentMonthStart.AddDate(0, -1, 0)

	// 本月数量
	currentCount, err := ss.getCountSince(resourceType, currentMonthStart)
	if err != nil {
		return 0, err
	}

	// 上月数量
	lastCount, err := ss.getCountBetween(resourceType, lastMonthStart, currentMonthStart)
	if err != nil {
		return 0, err
	}

	// 计算增长率
	if lastCount == 0 {
		if currentCount > 0 {
			return 100.0, nil // 从0增长视为100%
		}
		return 0, nil
	}

	trend := (float64(currentCount-lastCount) / float64(lastCount)) * 100
	return trend, nil
}

// getCountSince 获取指定时间之后的数量
func (ss *StatisticsService) getCountSince(resourceType string, since time.Time) (int64, error) {
	var count int64
	var err error

	switch resourceType {
	case "texture":
		err = ss.db.Model(&models.Texture{}).Where("created_at >= ?", since).Count(&count).Error
	case "project":
		err = ss.db.Model(&models.Project{}).Where("created_at >= ?", since).Count(&count).Error
	case "model":
		err = ss.db.Model(&models.Model{}).Where("created_at >= ?", since).Count(&count).Error
	case "asset":
		err = ss.db.Model(&models.Asset{}).Where("created_at >= ?", since).Count(&count).Error
	case "ai3d":
		var task struct {
			gorm.Model
		}
		err = ss.db.Model(&task).Table("ai3d_tasks").Where("created_at >= ?", since).Count(&count).Error
	}

	return count, err
}

// getCountBetween 获取指定时间段内的数量
func (ss *StatisticsService) getCountBetween(resourceType string, start, end time.Time) (int64, error) {
	var count int64
	var err error

	switch resourceType {
	case "texture":
		err = ss.db.Model(&models.Texture{}).Where("created_at >= ? AND created_at < ?", start, end).Count(&count).Error
	case "project":
		err = ss.db.Model(&models.Project{}).Where("created_at >= ? AND created_at < ?", start, end).Count(&count).Error
	case "model":
		err = ss.db.Model(&models.Model{}).Where("created_at >= ? AND created_at < ?", start, end).Count(&count).Error
	case "asset":
		err = ss.db.Model(&models.Asset{}).Where("created_at >= ? AND created_at < ?", start, end).Count(&count).Error
	case "ai3d":
		var task struct {
			gorm.Model
		}
		err = ss.db.Model(&task).Table("ai3d_tasks").Where("created_at >= ? AND created_at < ?", start, end).Count(&count).Error
	}

	return count, err
}

// GetRecentCount 获取最近N天新增数量
func (ss *StatisticsService) GetRecentCount(resourceType string, days int) (int64, error) {
	since := time.Now().AddDate(0, 0, -days)
	return ss.getCountSince(resourceType, since)
}

// GetRecentActivities 获取最近活动记录
func (ss *StatisticsService) GetRecentActivities(limit int) ([]models.Activity, error) {
	var activities []models.Activity
	err := ss.db.Order("created_at DESC").Limit(limit).Find(&activities).Error
	return activities, err
}

// GetStorageInfo 获取存储空间信息
func (ss *StatisticsService) GetStorageInfo() (StorageInfo, error) {
	info := StorageInfo{}

	// 获取主存储目录
	storageDir := "./static"
	if config.AppConfig.Texture.NASEnabled && config.AppConfig.Texture.NASPath != "" {
		storageDir = config.AppConfig.Texture.NASPath
	}

	// 获取磁盘空间（跨平台）
	total, used, err := getDiskUsage(storageDir)
	if err != nil {
		// 获取失败不影响整体，返回空值
		return info, nil
	}
	
	info.Total = total
	info.Used = used

	if info.Total > 0 {
		info.UsagePercent = float64(info.Used) / float64(info.Total) * 100
	}

	return info, nil
}

// 服务启动时间
var startTime = time.Now()
func (ss *StatisticsService) GetSystemStatus() (*SystemStatus, error) {
	status := &SystemStatus{}

	// 服务状态
	status.Service.Status = "running"
	status.Service.Uptime = int64(time.Since(startTime).Seconds())

	// 数据库状态
	status.Database.Status = "healthy"
	dbSize, _ := ss.getDatabaseSize()
	status.Database.Size = dbSize

	// 存储信息
	storageInfo, _ := ss.GetStorageInfo()
	status.Storage = storageInfo

	// 同步状态（从贴图同步日志获取）
	var lastSync models.TextureSyncLog
	err := ss.db.Order("created_at DESC").First(&lastSync).Error
	if err == nil {
		status.Sync.LastSyncAt = lastSync.CreatedAt
		if lastSync.Status == 2 {
			status.Sync.Status = "success"
		} else if lastSync.Status == 3 {
			status.Sync.Status = "failed"
		} else {
			status.Sync.Status = "running"
		}
	} else {
		status.Sync.Status = "unknown"
	}

	return status, nil
}

// getDatabaseSize 获取数据库大小
func (ss *StatisticsService) getDatabaseSize() (int64, error) {
	dbPath := config.AppConfig.DBPath
	fileInfo, err := os.Stat(dbPath)
	if err != nil {
		return 0, err
	}
	return fileInfo.Size(), nil
}

// RecordActivity 记录活动
func (ss *StatisticsService) RecordActivity(activityType, name, action, user, version string) error {
	activity := models.Activity{
		Type:      activityType,
		Name:      name,
		Action:    action,
		User:      user,
		Version:   version,
		CreatedAt: time.Now(),
	}
	return ss.db.Create(&activity).Error
}

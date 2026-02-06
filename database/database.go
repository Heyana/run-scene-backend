// Package database 提供数据库连接和操作
package database

import (
	"fmt"
	"go_wails_project_manager/config"
	"go_wails_project_manager/database/migrations"
	"go_wails_project_manager/logger"
	"go_wails_project_manager/models"
	"os"
	"path/filepath"

	"github.com/glebarez/sqlite" // 纯Go实现的SQLite驱动，无需CGO
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

var (
	db *gorm.DB
)

// Init 初始化数据库
func Init() error {
	var err error

	// 确保数据库目录存在
	dbDir := filepath.Dir(config.AppConfig.DBPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return fmt.Errorf("创建数据库目录失败: %w", err)
	}

	logger.Log.Infof("正在连接数据库: %s", config.AppConfig.DBPath)

	// 初始化数据库连接 (使用glebarez/sqlite驱动)
	db, err = gorm.Open(sqlite.Open(config.AppConfig.DBPath), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true, // 使用单数表名
		},
	})

	if err != nil {
		return fmt.Errorf("连接数据库失败: %w", err)
	}

	// 运行数据库迁移
	if err := RunMigrations(); err != nil {
		return fmt.Errorf("数据库迁移失败: %w", err)
	}

	logger.Log.Info("数据库初始化完成")
	return nil
}

// GetDB 获取数据库连接
func GetDB() (*gorm.DB, error) {
	if db == nil {
		return nil, fmt.Errorf("数据库未初始化")
	}
	return db, nil
}

// MustGetDB 获取数据库连接，如果未初始化则panic
func MustGetDB() *gorm.DB {
	if db == nil {
		logger.Log.Fatal("数据库未初始化")
	}
	return db
}

// AutoMigrate 自动迁移表结构
func AutoMigrate(models ...interface{}) error {
	if db == nil {
		return fmt.Errorf("数据库未初始化")
	}
	return db.AutoMigrate(models...)
}

// RunMigrations 运行数据库迁移
func RunMigrations() error {
	logger.Log.Info("开始运行数据库迁移...")

	// 执行AI3D表合并迁移（必须在AutoMigrate之前）
	if err := migrations.MergeAI3DTasks(db); err != nil {
		logger.Log.Errorf("AI3D表合并失败: %v", err)
		// 不返回错误，允许系统继续运行
	}

	// 自动迁移所有模型
	err := db.AutoMigrate(
		&models.MediaLibrary{},
		&models.MediaFile{},
		&models.MediaTag{},
		&models.MediaFileTag{},
		&models.SystemConfig{},
		&models.ResourceFile{}, // 通用资源文件表
		&models.BackupRecord{}, // 备份记录表
		// 贴图库相关表
		&models.File{},
		&models.Texture{},
		&models.TextureFile{},
		&models.Tag{},
		&models.TextureTag{},
		&models.TextureSyncLog{},
		&models.DownloadQueue{},
		&models.TextureMetrics{},
		// 模型库相关表
		&models.Model{},
		&models.ModelTag{},
		&models.ModelMetrics{},
		// 资产库相关表
		&models.Asset{},
		&models.AssetMetadata{},
		&models.AssetTag{},
		&models.AssetMetrics{},
		// 项目管理相关表
		&models.Project{},
		&models.ProjectVersion{},
		// 注意：不再迁移旧的 hunyuan_tasks 和 meshy_tasks 表
		// 它们已被合并到 ai3d_tasks 表中
		// &hunyuan.HunyuanTask{},
		// &hunyuan.HunyuanConfig{},
		// &meshy.MeshyTask{},
		// AI蓝图相关表
		&models.BlueprintHistory{},
	)
	if err != nil {
		return err
	}

	// 创建贴图库索引
	if err := createTextureIndexes(); err != nil {
		logger.Log.Warnf("创建贴图库索引失败: %v", err)
	}

	// 创建默认系统配置
	if err := createDefaultSystemConfigs(); err != nil {
		logger.Log.Warnf("创建默认系统配置失败: %v", err)
	}

	// 创建模型库存储目录
	if err := os.MkdirAll(config.AppConfig.Model.StorageDir, 0755); err != nil {
		logger.Log.Warnf("创建模型库存储目录失败: %v", err)
	} else {
		logger.Log.Infof("模型库存储目录: %s", config.AppConfig.Model.StorageDir)
	}
	
	// 创建资产库存储目录
	if err := os.MkdirAll(config.AppConfig.Asset.StorageDir, 0755); err != nil {
		logger.Log.Warnf("创建资产库存储目录失败: %v", err)
	} else {
		logger.Log.Infof("资产库存储目录: %s", config.AppConfig.Asset.StorageDir)
	}

	// 创建项目管理存储目录
	if config.ProjectAppConfig != nil {
		if err := os.MkdirAll(config.ProjectAppConfig.StorageDir, 0755); err != nil {
			logger.Log.Warnf("创建项目管理存储目录失败: %v", err)
		} else {
			logger.Log.Infof("项目管理存储目录: %s", config.ProjectAppConfig.StorageDir)
		}
	}

	// 运行一次性升级任务
	if err := RunOnceUpgrade(db); err != nil {
		logger.Log.Warnf("运行升级任务失败: %v", err)
	}

	logger.Log.Info("数据库迁移完成")
	return nil
}

// createTextureIndexes 创建贴图库相关索引
func createTextureIndexes() error {
	// 创建 texture_tag 联合唯一索引
	if err := db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_texture_tag_unique ON texture_tag(texture_id, tag_id)").Error; err != nil {
		return err
	}
	return nil
}

// createDefaultSystemConfigs 创建默认系统配置
func createDefaultSystemConfigs() error {
	configs := []models.SystemConfig{
		{
			ConfigKey:   "system_name",
			ConfigValue: "媒体管理器",
			Description: "系统名称",
		},
		{
			ConfigKey:   "system_version",
			ConfigValue: "1.0.0",
			Description: "系统版本",
		},
		{
			ConfigKey:   "max_file_size",
			ConfigValue: "104857600", // 100MB
			Description: "单个文件最大大小（字节）",
		},
		{
			ConfigKey:   "supported_image_formats",
			ConfigValue: "jpg,jpeg,png,gif,webp,bmp",
			Description: "支持的图片格式",
		},
		{
			ConfigKey:   "supported_video_formats",
			ConfigValue: "mp4,avi,mov,wmv,flv,webm",
			Description: "支持的视频格式",
		},
		{
			ConfigKey:   "thumbnail_width",
			ConfigValue: "300",
			Description: "缩略图宽度",
		},
		{
			ConfigKey:   "thumbnail_height",
			ConfigValue: "300",
			Description: "缩略图高度",
		},
	}

	for _, config := range configs {
		var existing models.SystemConfig
		if db.Where("config_key = ?", config.ConfigKey).First(&existing).Error != nil {
			if err := db.Create(&config).Error; err != nil {
				return err
			}
			logger.Log.Infof("创建系统配置: %s", config.ConfigKey)
		}
	}

	return nil
}

// Close 关闭数据库连接
func Close() error {
	if db == nil {
		return nil
	}

	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	return sqlDB.Close()
}

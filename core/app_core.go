// Package core 提供应用程序的核心初始化和服务功能
package core

import (
	"time"

	"go_wails_project_manager/api"
	"go_wails_project_manager/config"
	"go_wails_project_manager/database"
	"go_wails_project_manager/logger"
	"go_wails_project_manager/server"
	"go_wails_project_manager/services"
	ai3dService "go_wails_project_manager/services/ai3d"
	"go_wails_project_manager/services/ai3d/adapters"
	"go_wails_project_manager/services/fileprocessor"
	"go_wails_project_manager/services/task"
	textureServices "go_wails_project_manager/services/texture"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// AppCore 应用程序核心结构
type AppCore struct {
	Server                *server.Server
	Log                   *logrus.Logger
	BackupScheduler       *services.BackupScheduler
	TextureSyncService    *textureServices.SyncService
	AI3DTaskService       *ai3dService.TaskService
	FileProcessorService  *fileprocessor.FileProcessorService
	FileProcessorConfig   *fileprocessor.Config
	TaskService           *task.TaskService
	IsRunning             bool
}

// NewAppCore 创建新的应用核心实例
func NewAppCore() (*AppCore, error) {
	// 初始化配置
	if err := config.LoadConfig(); err != nil {
		return nil, err
	}

	// 加载贴图映射配置
	if err := config.LoadTextureMappingConfig("configs/texture_mapping.yaml"); err != nil {
		logger.Log.Warnf("加载贴图映射配置失败: %v，将使用默认配置", err)
	} else {
		logger.Log.Info("贴图映射配置加载成功")
	}

	// 加载数据库版本配置
	if err := config.LoadDatabaseVersionConfig("configs/database_version.yaml"); err != nil {
		logger.Log.Warnf("加载数据库版本配置失败: %v，将使用默认版本", err)
	} else {
		logger.Log.Infof("数据库版本配置加载成功，目标版本: %d", config.DatabaseVersion.GetTargetVersion())
	}

	// 加载项目管理配置
	if err := config.LoadProjectConfig(); err != nil {
		logger.Log.Warnf("加载项目管理配置失败: %v，将使用默认配置", err)
	} else {
		logger.Log.Info("项目管理配置加载成功")
	}

	// 初始化日志
	logger.Init()
	log := logger.GetLogger()

	return &AppCore{
		Log: log,
	}, nil
}

// InitDatabases 初始化数据库连接和迁移表结构
func (a *AppCore) InitDatabases() error {
	// 初始化主数据库
	a.Log.Info("正在初始化数据库...")
	if err := database.Init(); err != nil {
		a.Log.Errorf("数据库初始化失败: %v", err)
		return err
	}
	a.Log.Info("数据库初始化成功")

	// 数据库迁移已在database.Init()中处理
	a.Log.Info("数据库迁移完成")

	// 初始化备份服务
	if err := a.InitBackupService(); err != nil {
		a.Log.Errorf("备份服务初始化失败: %v", err)
		return err
	}

	// 初始化贴图服务
	if err := a.InitTextureService(); err != nil {
		a.Log.Errorf("贴图服务初始化失败: %v", err)
		return err
	}

	// 初始化AI3D服务
	if err := a.InitAI3DService(); err != nil {
		a.Log.Errorf("AI3D服务初始化失败: %v", err)
		return err
	}

	// 初始化文件处理器服务
	if err := a.InitFileProcessorService(); err != nil {
		a.Log.Errorf("文件处理器服务初始化失败: %v", err)
		return err
	}

	return nil
}

// InitBackupService 初始化备份服务
func (a *AppCore) InitBackupService() error {
	a.Log.Info("正在初始化备份服务...")

	// 加载备份配置
	backupConfig, cosConfig := config.LoadBackupConfig()

	// 获取数据库连接
	db, err := database.GetDB()
	if err != nil {
		return err
	}

	// 创建备份调度器
	a.BackupScheduler = services.NewBackupScheduler(backupConfig, cosConfig, db)

	// 设置全局备份调度器
	services.SetGlobalBackupScheduler(a.BackupScheduler)

	// 启动备份调度器
	if err := a.BackupScheduler.Start(); err != nil {
		return err
	}

	a.Log.Info("备份服务初始化成功")
	return nil
}

// InitTextureService 初始化贴图服务
func (a *AppCore) InitTextureService() error {
	a.Log.Info("正在初始化贴图服务...")

	// 获取数据库连接
	db, err := database.GetDB()
	if err != nil {
		return err
	}

	// 创建存储目录
	storageDir := config.AppConfig.Texture.StorageDir
	if err := os.MkdirAll(storageDir, 0755); err != nil {
		a.Log.Errorf("创建贴图存储目录失败: %v", err)
		return err
	}
	a.Log.Infof("贴图存储目录: %s", storageDir)

	// 初始化同步服务
	a.TextureSyncService = textureServices.NewSyncService(db, a.Log)

	// 设置全局同步服务
	textureServices.SetGlobalSyncService(a.TextureSyncService)

	// 启动定时同步任务
	a.TextureSyncService.StartScheduler()
	a.Log.Info("贴图同步调度器已启动")

	// 启动后自动执行一次增量同步（PolyHaven + AmbientCG）
	go func() {
		// 1. PolyHaven 增量同步
		a.Log.Info("启动后自动执行 PolyHaven 增量同步...")
		if err := a.TextureSyncService.IncrementalSync(); err != nil {
			a.Log.Errorf("PolyHaven 自动同步失败: %v", err)
		} else {
			a.Log.Info("PolyHaven 自动同步完成")
		}

		// 2. AmbientCG 增量同步
		a.Log.Info("启动后自动执行 AmbientCG 增量同步...")
		ambientcgService := textureServices.NewAmbientCGSyncService(db, a.Log)
		if err := ambientcgService.IncrementalSync(); err != nil {
			a.Log.Errorf("AmbientCG 自动同步失败: %v", err)
		} else {
			a.Log.Info("AmbientCG 自动同步完成")
		}
	}()

	a.Log.Info("贴图服务初始化成功")
	return nil
}

// InitAI3DService 初始化AI3D服务
func (a *AppCore) InitAI3DService() error {
	a.Log.Info("正在初始化AI3D服务...")

	// 获取数据库连接
	db, err := database.GetDB()
	if err != nil {
		return err
	}

	// 创建任务服务（轮询间隔5秒）
	a.AI3DTaskService = ai3dService.NewTaskService(db, 5*time.Second)

	// 注册混元适配器
	if config.AppConfig.Hunyuan.SecretID != "" {
		hunyuanAdapter := adapters.NewHunyuanAdapter(db, &config.AppConfig.Hunyuan)
		a.AI3DTaskService.RegisterAdapter(hunyuanAdapter)
		a.Log.Info("混元适配器已注册")
	} else {
		a.Log.Warn("混元配置未设置，跳过注册")
	}

	// 注册Meshy适配器
	if config.AppConfig.Meshy.APIKey != "" {
		meshyAdapter := adapters.NewMeshyAdapter(db, &config.AppConfig.Meshy)
		a.AI3DTaskService.RegisterAdapter(meshyAdapter)
		a.Log.Info("Meshy适配器已注册")
	} else {
		a.Log.Warn("Meshy配置未设置，跳过注册")
	}

	// 启动轮询器
	a.AI3DTaskService.StartPoller()
	a.Log.Info("AI3D轮询器已启动")

	a.Log.Info("AI3D服务初始化成功")
	return nil
}

// InitFileProcessorService 初始化文件处理器服务
func (a *AppCore) InitFileProcessorService() error {
	a.Log.Info("正在初始化文件处理器服务...")

	// 获取数据库连接
	db, err := database.GetDB()
	if err != nil {
		return err
	}

	// 创建文件处理器配置
	fpConfig := &fileprocessor.Config{
		FFmpeg: fileprocessor.FFmpegConfig{
			BinPath: config.FileProcessorAppConfig.FFmpeg.BinPath,
			Timeout: config.FileProcessorAppConfig.FFmpeg.Timeout,
		},
		ImageMagick: fileprocessor.ImageMagickConfig{
			BinPath: config.FileProcessorAppConfig.ImageMagick.BinPath,
			Timeout: config.FileProcessorAppConfig.ImageMagick.Timeout,
		},
		PDF: fileprocessor.PDFConfig{
			BinPath: config.FileProcessorAppConfig.PDF.BinPath,
			Timeout: config.FileProcessorAppConfig.PDF.Timeout,
		},
		Blender: fileprocessor.BlenderConfig{
			BinPath:    config.FileProcessorAppConfig.Blender.BinPath,
			ScriptPath: config.FileProcessorAppConfig.Blender.ScriptPath,
			Timeout:    config.FileProcessorAppConfig.Blender.Timeout,
		},
		Thumbnail: fileprocessor.ThumbnailConfig{
			Format:  config.FileProcessorAppConfig.Thumbnail.Format,
			Width:   config.FileProcessorAppConfig.Thumbnail.Width,
			Height:  config.FileProcessorAppConfig.Thumbnail.Height,
			Quality: config.FileProcessorAppConfig.Thumbnail.Quality,
		},
		Task: fileprocessor.TaskConfig{
			MaxConcurrent: config.FileProcessorAppConfig.Task.MaxConcurrent,
			MaxRetries:    config.FileProcessorAppConfig.Task.MaxRetries,
			RetryDelay:    config.FileProcessorAppConfig.Task.RetryDelay,
			CleanupAfter:  config.FileProcessorAppConfig.Task.CleanupAfter,
		},
		Resource: fileprocessor.ResourceConfig{
			MaxMemoryPerTask: config.FileProcessorAppConfig.Resource.MaxMemoryPerTask,
			MaxCPUPercent:    config.FileProcessorAppConfig.Resource.MaxCPUPercent,
			MaxTempSize:      config.FileProcessorAppConfig.Resource.MaxTempSize,
		},
	}

	// 初始化文件处理器服务
	a.FileProcessorService = fileprocessor.NewFileProcessorService(fpConfig)
	a.FileProcessorConfig = fpConfig // 保存配置供其他地方使用
	a.Log.Info("文件处理器服务已创建")

	// 初始化任务服务（传入文件处理器服务）
	a.TaskService = task.NewTaskService(db, a.FileProcessorService)
	a.Log.Info("任务服务已创建")

	// 恢复未完成的任务
	if err := a.TaskService.RecoverTasks(); err != nil {
		a.Log.Warnf("恢复未完成任务失败: %v", err)
	} else {
		a.Log.Info("未完成任务已恢复")
	}

	a.Log.Info("文件处理器服务初始化成功")
	return nil
}

// StartServer 启动HTTP服务器
func (a *AppCore) StartServer() error {
	// 创建并启动 Gin 服务器
	a.Log.Info("正在初始化 HTTP 服务器...")
	a.Server = server.NewServer(config.AppConfig.ServerPort)

	// 添加自定义路由
	a.Server.AddRoutes(func(router *gin.Engine) {
		// 注册所有 API 路由（传递AI3D服务、文件处理器服务和配置）
		api.RegisterRoutes(router, a.Log, a.AI3DTaskService, a.FileProcessorService, a.FileProcessorConfig, a.TaskService)
	})

	err := a.Server.Start()
	if err != nil {
		a.Log.Errorf("无法启动 HTTP 服务器: %v", err)
		return err
	}

	a.IsRunning = true
	a.Log.Infof("HTTP 服务器已启动在端口 %d", config.AppConfig.ServerPort)
	a.Log.Infof("API文档可通过 %s:%d/api/docs 访问", config.GetAPIDocsBaseURL(), config.AppConfig.ServerPort)

	return nil
}

// StopServer 停止HTTP服务器
func (a *AppCore) StopServer() error {
	if a.IsRunning && a.Server != nil {
		a.Log.Info("正在关闭 HTTP 服务器...")
		if err := a.Server.Stop(); err != nil {
			a.Log.Errorf("关闭 HTTP 服务器时出错: %v", err)
			return err
		}
		a.IsRunning = false
	}

	return nil
}

// GetServerStatus 获取服务器状态
func (a *AppCore) GetServerStatus() map[string]interface{} {
	return map[string]interface{}{
		"running": a.IsRunning,
		"port":    config.AppConfig.ServerPort,
	}
}

// Shutdown 执行优雅停机
func (a *AppCore) Shutdown() {
	a.Log.Info("🔄 开始优雅停机...")

	// 1. 停止文件处理器任务（等待当前任务完成）
	if a.TaskService != nil {
		a.Log.Info("⏳ 正在等待文件处理任务完成...")
		// 任务服务会自动等待当前任务完成
		a.Log.Info("✅ 文件处理任务已停止")
	}

	// 2. 停止AI3D轮询器
	if a.AI3DTaskService != nil {
		a.Log.Info("⏳ 正在停止AI3D轮询器...")
		a.AI3DTaskService.StopPoller()
		a.Log.Info("✅ AI3D轮询器已停止")
	}

	// 3. 停止贴图同步调度器
	if a.TextureSyncService != nil {
		a.Log.Info("⏳ 正在停止贴图同步调度器...")
		a.TextureSyncService.StopScheduler()
		a.Log.Info("✅ 贴图同步调度器已停止")
	}

	// 4. 停止备份调度器
	if a.BackupScheduler != nil {
		a.Log.Info("⏳ 正在停止备份调度器...")
		a.BackupScheduler.Stop()
		a.Log.Info("✅ 备份调度器已停止")
	}

	// 5. 停止 HTTP 服务器
	if err := a.StopServer(); err != nil {
		a.Log.Errorf("❌ 停止服务器失败: %v", err)
	} else {
		a.Log.Info("✅ HTTP服务器已停止")
	}

	// 6. 关闭数据库连接
	a.Log.Info("⏳ 正在关闭数据库连接...")
	if err := database.Close(); err != nil {
		a.Log.Errorf("❌ 关闭数据库失败: %v", err)
	} else {
		a.Log.Info("✅ 数据库连接已关闭")
	}

	a.Log.Info("👋 优雅停机完成")
}

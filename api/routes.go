// Package api 提供REST API实现
package api

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"go_wails_project_manager/config"
	"go_wails_project_manager/controllers"
	"go_wails_project_manager/database"
	"go_wails_project_manager/logger"
	"go_wails_project_manager/middleware"
	"go_wails_project_manager/response"
	ai3dService "go_wails_project_manager/services/ai3d"
	"go_wails_project_manager/services/audit"
	"go_wails_project_manager/services/fileprocessor"
	"go_wails_project_manager/services/task"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// Ping 健康检查
// @Summary 健康检查
// @Description 检查API服务是否正常运行
// @Tags 系统
// @Accept json
// @Produce json
// @Success 200 {object} response.Response "服务正常"
// @Router /api/ping [get]
func Ping(c *gin.Context) {
	response.SuccessWithMsg(c, "服务正常运行", gin.H{
		"timestamp": time.Now().Unix(),
		"status":    "ok",
	})
}

// RegisterRoutes 注册API路由
func RegisterRoutes(router *gin.Engine, log *logrus.Logger, ai3dTaskService interface{}, fileProcessorService interface{}, fileProcessorConfig interface{}, taskService interface{}) {
	// 初始化审计服务
	auditConfig, _ := config.LoadAuditConfig()
	var auditService *audit.AuditService
	var auditQueryService *audit.QueryService
	var auditArchiveService *audit.ArchiveService
	var auditController *controllers.AuditController
	
	if auditConfig != nil && auditConfig.Enabled {
		auditService = audit.NewAuditService(database.MustGetDB(), auditConfig)
		auditQueryService = audit.NewQueryService(database.MustGetDB(), auditConfig)
		auditArchiveService = audit.NewArchiveService(database.MustGetDB(), auditConfig)
		auditController = controllers.NewAuditController(auditQueryService, auditArchiveService)
		
		logger.Log.Info("审计服务已启动")
	}
	
	// 初始化有用的控制器
	backupController := controllers.NewBackupController()
	securityController := controllers.NewSecurityController()
	textureController := controllers.NewTextureController()
	modelController := controllers.NewModelController(database.MustGetDB())
	assetController := controllers.NewAssetController(database.MustGetDB())
	
	// 创建文档控制器（传递文件处理器服务和配置）
	var documentController *controllers.DocumentController
	if fileProcessorService != nil && fileProcessorConfig != nil {
		documentController = controllers.NewDocumentController(
			database.MustGetDB(), 
			fileProcessorService.(*fileprocessor.FileProcessorService),
			fileProcessorConfig.(*fileprocessor.Config),
		)
	} else {
		documentController = controllers.NewDocumentController(database.MustGetDB(), nil, nil)
	}
	
	imageController := controllers.NewImageController()
	blueprintController := controllers.NewBlueprintController(database.MustGetDB())
	projectController := controllers.NewProjectController(database.MustGetDB())
	statisticsController := controllers.NewStatisticsController(database.MustGetDB())
	
	// 创建AI3D统一控制器（如果服务已初始化）
	var ai3dUnifiedController *controllers.AI3DUnifiedController
	if ai3dTaskService != nil {
		if service, ok := ai3dTaskService.(*ai3dService.TaskService); ok {
			ai3dUnifiedController = controllers.NewAI3DUnifiedController(service)
		}
	}

	// 创建文件处理器控制器（如果服务已初始化）
	var fileProcessorController *controllers.FileProcessorController
	if fileProcessorService != nil && taskService != nil {
		fileProcessorController = controllers.NewFileProcessorController(
			fileProcessorService.(*fileprocessor.FileProcessorService),
			taskService.(*task.TaskService),
			log,
		)
	}

	// 设置API文档（仅开发环境）
	if config.IsDev() {
		SetupAPIDocs(router)
	}

	// ==================== 全局安全中间件（按优先级顺序）====================
	// 0. Panic恢复中间件（最先添加，确保捕获所有panic）
	router.Use(middleware.RecoveryWithLog(config.IsDev()))

	// 1. 请求ID追踪（便于日志追踪）
	router.Use(requestid.New())

	// 1. CORS中间件（支持白名单）
	router.Use(CorsMiddleware())

	// 2. 安全响应头
	router.Use(SecurityHeadersMiddleware())

	// 3. IP黑名单过滤（最高优先级 - 直接拒绝已知恶意IP）
	router.Use(IPFilterMiddleware())

	// 4. DDoS防护（并发连接限制 + 请求频率检测）
	router.Use(DDoSProtectionMiddleware())

	// 5. 连接频率限制
	router.Use(ConnectionRateLimitMiddleware())

	// 6. 速率限制（令牌桶算法）
	router.Use(RateLimitMiddleware())

	// 7. 可疑活动检测（SQL注入/XSS/路径遍历）
	router.Use(DetectSuspiciousActivityMiddleware())

	// 8. 请求大小限制（全局100MB，文件上传路由除外）
	router.Use(RequestSizeLimitMiddleware(100 * 1024 * 1024))

	// 9. 敏感路径保护
	router.Use(ProtectSensitivePathsMiddleware())

	// 10. 审计中间件（如果启用）
	if auditService != nil {
		router.Use(middleware.AuditMiddleware(auditService))
		logger.Log.Info("审计中间件已启用")
	}

	// 根级别健康检查端点（供 Electron 检测）
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":    "ok",
			"timestamp": time.Now().Unix(),
			"service":   "media-manager-server",
		})
	})

	// 贴图文件静态服务
	// 如果启用了 NAS，使用 NAS 路径；否则使用本地路径
	textureDir := "./static/textures"
	if config.AppConfig.Texture.NASEnabled && config.AppConfig.Texture.NASPath != "" {
		// 使用 NAS 路径
		textureDir = config.AppConfig.Texture.NASPath
		logger.Log.Infof("使用 NAS 路径提供静态文件服务: %s", textureDir)
	} else {
		logger.Log.Infof("使用本地路径提供静态文件服务: %s", textureDir)
	}
	
	// 使用自定义处理器来支持 UNC 路径
	router.GET("/textures/*filepath", func(c *gin.Context) {
		filepath := c.Param("filepath")
		// 移除开头的斜杠
		if len(filepath) > 0 && filepath[0] == '/' {
			filepath = filepath[1:]
		}
		
		// 拼接完整路径（使用 path.Join 而不是 filepath.Join，保持正斜杠）
		fullPath := textureDir
		if !strings.HasSuffix(fullPath, "/") && !strings.HasSuffix(fullPath, "\\") {
			fullPath += "/"
		}
		fullPath += filepath
		
		// 不再强制转换路径分隔符，保持原样
		// Linux 使用正斜杠，Windows 使用反斜杠都能正常工作
		
		logger.Log.Infof("请求文件: %s -> %s", filepath, fullPath)
		
		// 检查文件是否存在
		fileInfo, err := os.Stat(fullPath)
		if os.IsNotExist(err) {
			logger.Log.Warnf("文件不存在: %s", fullPath)
			c.JSON(404, gin.H{
				"error": "文件不存在",
				"path":  fullPath,
			})
			return
		}
		if err != nil {
			logger.Log.Errorf("访问文件失败: %s, 错误: %v", fullPath, err)
			c.JSON(500, gin.H{
				"error": "访问文件失败",
				"path":  fullPath,
				"msg":   err.Error(),
			})
			return
		}
		
		logger.Log.Infof("文件存在，大小: %d bytes", fileInfo.Size())
		
		// 返回文件
		c.File(fullPath)
	})

	// 模型文件静态服务
	// 如果启用了 NAS，使用 NAS 路径；否则使用本地路径
	modelDir := "./static/models"
	if config.AppConfig.Model.NASEnabled && config.AppConfig.Model.NASPath != "" {
		// 使用 NAS 路径
		modelDir = config.AppConfig.Model.NASPath
		logger.Log.Infof("模型库使用 NAS 路径提供静态文件服务: %s", modelDir)
	} else {
		logger.Log.Infof("模型库使用本地路径提供静态文件服务: %s", modelDir)
	}
	
	// 使用自定义处理器来支持 UNC 路径
	router.GET("/models/*filepath", func(c *gin.Context) {
		filepath := c.Param("filepath")
		// 移除开头的斜杠
		if len(filepath) > 0 && filepath[0] == '/' {
			filepath = filepath[1:]
		}
		
		// 拼接完整路径
		fullPath := modelDir
		if !strings.HasSuffix(fullPath, "/") && !strings.HasSuffix(fullPath, "\\") {
			fullPath += "/"
		}
		fullPath += filepath
		
		logger.Log.Infof("请求模型文件: %s -> %s", filepath, fullPath)
		
		// 检查文件是否存在
		fileInfo, err := os.Stat(fullPath)
		if os.IsNotExist(err) {
			logger.Log.Warnf("模型文件不存在: %s", fullPath)
			c.JSON(404, gin.H{
				"error": "文件不存在",
				"path":  fullPath,
			})
			return
		}
		if err != nil {
			logger.Log.Errorf("访问模型文件失败: %s, 错误: %v", fullPath, err)
			c.JSON(500, gin.H{
				"error": "访问文件失败",
				"path":  fullPath,
				"msg":   err.Error(),
			})
			return
		}
		
		logger.Log.Infof("模型文件存在，大小: %d bytes", fileInfo.Size())
		
		// 返回文件
		c.File(fullPath)
	})

	// 资产文件静态服务
	// 如果启用了 NAS，使用 NAS 路径；否则使用本地路径
	assetDir := "./static/assets"
	if config.AppConfig.Asset.NASEnabled && config.AppConfig.Asset.NASPath != "" {
		// 使用 NAS 路径
		assetDir = config.AppConfig.Asset.NASPath
		logger.Log.Infof("资产库使用 NAS 路径提供静态文件服务: %s", assetDir)
	} else {
		logger.Log.Infof("资产库使用本地路径提供静态文件服务: %s", assetDir)
	}
	
	// 使用自定义处理器来支持 UNC 路径
	router.GET("/assets/*filepath", func(c *gin.Context) {
		filepath := c.Param("filepath")
		// 移除开头的斜杠
		if len(filepath) > 0 && filepath[0] == '/' {
			filepath = filepath[1:]
		}
		
		// 拼接完整路径
		fullPath := assetDir
		if !strings.HasSuffix(fullPath, "/") && !strings.HasSuffix(fullPath, "\\") {
			fullPath += "/"
		}
		fullPath += filepath
		
		logger.Log.Infof("请求资产文件: %s -> %s", filepath, fullPath)
		
		// 检查文件是否存在
		fileInfo, err := os.Stat(fullPath)
		if os.IsNotExist(err) {
			logger.Log.Warnf("资产文件不存在: %s", fullPath)
			c.JSON(404, gin.H{
				"error": "文件不存在",
				"path":  fullPath,
			})
			return
		}
		if err != nil {
			logger.Log.Errorf("访问资产文件失败: %s, 错误: %v", fullPath, err)
			c.JSON(500, gin.H{
				"error": "访问文件失败",
				"path":  fullPath,
				"msg":   err.Error(),
			})
			return
		}
		
		logger.Log.Infof("资产文件存在，大小: %d bytes", fileInfo.Size())
		
		// 返回文件
		c.File(fullPath)
	})

	// 混元3D文件静态服务
	// 如果启用了 NAS，使用 NAS 路径；否则使用本地路径
	hunyuanDir := "./static/hunyuan"
	if config.AppConfig.Hunyuan.NASEnabled && config.AppConfig.Hunyuan.NASPath != "" {
		// 使用 NAS 路径
		hunyuanDir = config.AppConfig.Hunyuan.NASPath
		logger.Log.Infof("混元3D使用 NAS 路径提供静态文件服务: %s", hunyuanDir)
	} else {
		logger.Log.Infof("混元3D使用本地路径提供静态文件服务: %s", hunyuanDir)
	}
	
	// 使用自定义处理器来支持 UNC 路径
	router.GET("/hunyuan/*filepath", func(c *gin.Context) {
		filepath := c.Param("filepath")
		// 移除开头的斜杠
		if len(filepath) > 0 && filepath[0] == '/' {
			filepath = filepath[1:]
		}
		
		// 拼接完整路径
		fullPath := hunyuanDir
		if !strings.HasSuffix(fullPath, "/") && !strings.HasSuffix(fullPath, "\\") {
			fullPath += "/"
		}
		fullPath += filepath
		
		logger.Log.Infof("请求混元3D文件: %s -> %s", filepath, fullPath)
		
		// 检查文件是否存在
		fileInfo, err := os.Stat(fullPath)
		if os.IsNotExist(err) {
			logger.Log.Warnf("混元3D文件不存在: %s", fullPath)
			c.JSON(404, gin.H{
				"error": "文件不存在",
				"path":  fullPath,
			})
			return
		}
		if err != nil {
			logger.Log.Errorf("访问混元3D文件失败: %s, 错误: %v", fullPath, err)
			c.JSON(500, gin.H{
				"error": "访问文件失败",
				"path":  fullPath,
				"msg":   err.Error(),
			})
			return
		}
		
		logger.Log.Infof("混元3D文件存在，大小: %d bytes", fileInfo.Size())
		
		// 返回文件
		c.File(fullPath)
	})

	// Meshy文件静态服务
	meshyDir := "./static/meshy"
	if config.AppConfig.Meshy.NASEnabled && config.AppConfig.Meshy.NASPath != "" {
		meshyDir = config.AppConfig.Meshy.NASPath
		logger.Log.Infof("Meshy使用 NAS 路径提供静态文件服务: %s", meshyDir)
	} else {
		logger.Log.Infof("Meshy使用本地路径提供静态文件服务: %s", meshyDir)
	}
	
	router.GET("/meshy/*filepath", func(c *gin.Context) {
		filepath := c.Param("filepath")
		if len(filepath) > 0 && filepath[0] == '/' {
			filepath = filepath[1:]
		}
		
		fullPath := meshyDir
		if !strings.HasSuffix(fullPath, "/") && !strings.HasSuffix(fullPath, "\\") {
			fullPath += "/"
		}
		fullPath += filepath
		
		logger.Log.Infof("请求Meshy文件: %s -> %s", filepath, fullPath)
		
		fileInfo, err := os.Stat(fullPath)
		if os.IsNotExist(err) {
			logger.Log.Warnf("Meshy文件不存在: %s", fullPath)
			c.JSON(404, gin.H{
				"error": "文件不存在",
				"path":  fullPath,
			})
			return
		}
		if err != nil {
			logger.Log.Errorf("访问Meshy文件失败: %s, 错误: %v", fullPath, err)
			c.JSON(500, gin.H{
				"error": "访问文件失败",
				"path":  fullPath,
				"msg":   err.Error(),
			})
			return
		}
		
		logger.Log.Infof("Meshy文件存在，大小: %d bytes", fileInfo.Size())
		c.File(fullPath)
	})

	// 项目文件静态服务（当前版本）
	projectDir := "./static/projects"
	if config.ProjectAppConfig != nil && config.ProjectAppConfig.NASEnabled && config.ProjectAppConfig.NASPath != "" {
		projectDir = config.ProjectAppConfig.NASPath
		logger.Log.Infof("项目管理使用 NAS 路径提供静态文件服务: %s", projectDir)
	} else {
		logger.Log.Infof("项目管理使用本地路径提供静态文件服务: %s", projectDir)
	}
	
	router.Static("/projects", projectDir)
	
	// 项目历史版本静态服务
	projectHistoryDir := "./static/project_histories"
	if config.ProjectAppConfig != nil && config.ProjectAppConfig.NASEnabled && config.ProjectAppConfig.NASHistoryPath != "" {
		projectHistoryDir = config.ProjectAppConfig.NASHistoryPath
		logger.Log.Infof("项目历史版本使用 NAS 路径提供静态文件服务: %s", projectHistoryDir)
	} else {
		logger.Log.Infof("项目历史版本使用本地路径提供静态文件服务: %s", projectHistoryDir)
	}
	
	router.Static("/project_histories", projectHistoryDir)

	// 文件库静态服务
	docConfig, _ := config.LoadDocumentConfig()
	documentDir := "./static/documents"
	if docConfig != nil && docConfig.NASEnabled && docConfig.NASPath != "" {
		documentDir = docConfig.NASPath
		logger.Log.Infof("文件库使用 NAS 路径提供静态文件服务: %s", documentDir)
	} else {
		logger.Log.Infof("文件库使用本地路径提供静态文件服务: %s", documentDir)
	}
	
	router.GET("/documents/*filepath", func(c *gin.Context) {
		requestPath := c.Param("filepath")
		if len(requestPath) > 0 && requestPath[0] == '/' {
			requestPath = requestPath[1:]
		}
		
		// 使用 filepath.Join 正确拼接路径
		fullPath := filepath.Join(documentDir, requestPath)
		
		logger.Log.Infof("请求文件库文件: %s -> %s", requestPath, fullPath)
		
		fileInfo, err := os.Stat(fullPath)
		if os.IsNotExist(err) {
			logger.Log.Warnf("文件库文件不存在: %s", fullPath)
			c.JSON(404, gin.H{
				"error": "文件不存在",
				"path":  fullPath,
			})
			return
		}
		if err != nil {
			logger.Log.Errorf("访问文件库文件失败: %s, 错误: %v", fullPath, err)
			c.JSON(500, gin.H{
				"error": "访问文件失败",
				"path":  fullPath,
				"msg":   err.Error(),
			})
			return
		}
		
		logger.Log.Infof("文件库文件存在，大小: %d bytes", fileInfo.Size())
		c.File(fullPath)
	})

	// 设置API路由组
	api := router.Group("/api")
	{
		api.GET("/ping", Ping)

		// 备份管理API
		backup := api.Group("/backup")
		{
			backup.GET("/status", backupController.GetStatus)                             // 获取备份状态
			backup.POST("/trigger", backupController.TriggerManualBackup)                 // 手动触发全量备份
			backup.POST("/database", backupController.TriggerDatabaseBackup)              // 手动触发数据库备份
			backup.POST("/cdn", backupController.TriggerCDNBackup)                        // 手动触发CDN备份
			backup.GET("/history", backupController.GetBackupHistory)                     // 获取备份历史
			backup.POST("/restore/cdn/:backup_id", backupController.RestoreCDNFromBackup) // CDN文件恢复
		}

		// 安全管理API
		security := api.Group("/security")
		{
			security.GET("/status", securityController.GetStatus)                     // 获取安全状态
			security.GET("/blocked-ips", securityController.GetBlockedIPs)            // 获取被封禁IP列表
			security.POST("/unblock/:ip", securityController.UnblockIP)               // 解封IP地址
			security.GET("/ip-stats", securityController.GetIPStats)                  // 获取IP统计信息
			security.POST("/block/:ip", securityController.BlockIP)                   // 封禁IP地址
			security.POST("/whitelist/:ip", securityController.AddToWhitelist)        // 添加IP到白名单
			security.DELETE("/whitelist/:ip", securityController.RemoveFromWhitelist) // 从白名单移除IP
			security.GET("/connections", securityController.GetConnections)           // 获取连接统计
		}

		// 贴图库管理API
		textures := api.Group("/textures")
		{
			textures.GET("", textureController.List)                                    // 获取贴图列表
			textures.GET("/types", textureController.GetTextureTypes)                   // 获取所有贴图类型
			textures.GET("/types/threejs", textureController.GetThreeJSTypes)           // 获取 Three.js 贴图类型
			textures.GET("/analyze-types", textureController.AnalyzeTextureTypes)       // 分析所有贴图类型
			textures.GET("/:assetId", textureController.GetDetail)                      // 获取贴图详情
			textures.POST("/:assetId/use", textureController.RecordUse)                 // 记录使用次数
			textures.POST("/sync", textureController.TriggerSync)                       // 触发同步
			textures.GET("/sync/progress", textureController.GetSyncProgress)           // 获取同步进度
			textures.GET("/sync/status/:logId", textureController.GetSyncStatus)        // 获取同步状态
			textures.GET("/sync/logs", textureController.GetSyncLogs)                   // 获取同步日志
			textures.POST("/download/:assetId", textureController.DownloadTexture)      // 触发材质下载（AmbientCG）
			textures.GET("/download-status/:assetId", textureController.CheckDownloadStatus) // 检查下载状态
		}

		// 标签管理API
		tags := api.Group("/tags")
		{
			tags.GET("", textureController.GetTags)                          // 获取标签列表
			tags.GET("/:tagId/textures", textureController.GetTexturesByTag) // 根据标签获取贴图
		}

		// 模型库管理API
		models := api.Group("/models")
		{
			models.POST("/upload", modelController.Upload)                  // 上传模型
			models.GET("", modelController.List)                            // 获取模型列表
			models.GET("/search", modelController.Search)                   // 搜索模型
			models.GET("/statistics", modelController.GetStatistics)        // 获取统计信息
			models.GET("/popular", modelController.GetPopular)              // 获取热门模型
			models.GET("/:id", modelController.GetDetail)                   // 获取模型详情
			models.POST("/:id/use", modelController.IncrementUseCount)      // 记录使用次数
			models.DELETE("/:id", modelController.Delete)                   // 删除模型
		}

		// 资产库管理API
		assets := api.Group("/assets")
		{
			assets.POST("/upload", assetController.Upload)                     // 上传资产
			assets.GET("", assetController.List)                               // 获取资产列表
			assets.GET("/statistics", assetController.GetStatistics)           // 获取统计信息
			assets.GET("/statistics/by-type", assetController.GetStatisticsByType) // 按类型统计
			assets.GET("/popular", assetController.GetPopular)                 // 获取热门资产
			assets.GET("/:id", assetController.GetDetail)                      // 获取资产详情
			assets.PUT("/:id", assetController.Update)                         // 更新资产信息
			assets.DELETE("/:id", assetController.Delete)                      // 删除资产
			assets.POST("/:id/use", assetController.IncrementUseCount)         // 记录使用次数
		}

		// 文件库管理API（统一文件和文件夹）
		documents := api.Group("/documents")
		{
			documents.POST("/upload", documentController.Upload)               // 上传文档
			documents.POST("/upload-folder", documentController.UploadFolder)  // 上传文件夹（保持结构）
			documents.POST("/folder", documentController.CreateFolder)         // 创建文件夹
			documents.GET("", documentController.List)                         // 获取文档列表（支持parent_id过滤）
			documents.GET("/statistics", documentController.GetStatistics)     // 获取统计信息
			documents.GET("/popular", documentController.GetPopular)           // 获取热门文档
			documents.GET("/:id", documentController.GetDetail)                // 获取文档详情
			documents.PUT("/:id", documentController.Update)                   // 更新文档信息
			documents.DELETE("/:id", documentController.Delete)                // 删除文档（支持级联删除文件夹）
			documents.GET("/:id/download", documentController.Download)        // 下载文档
			documents.POST("/:id/refresh-thumbnail", documentController.RefreshThumbnail) // 刷新缩略图
			documents.GET("/:id/versions", documentController.GetVersions)     // 获取版本列表
			documents.GET("/:id/logs", documentController.GetAccessLogs)       // 获取访问日志
		}

		// AI 3D生成统一API（支持多平台）
		ai3d := api.Group("/ai3d")
		{
			// 使用统一控制器
			if ai3dUnifiedController != nil {
				ai3d.POST("/tasks", ai3dUnifiedController.SubmitTask)           // 提交任务（支持provider参数）
				ai3d.GET("/tasks", ai3dUnifiedController.ListTasks)             // 任务列表
				ai3d.GET("/tasks/:id", ai3dUnifiedController.GetTask)           // 获取任务详情
				ai3d.POST("/tasks/:id/poll", ai3dUnifiedController.PollTask)    // 轮询任务
				ai3d.DELETE("/tasks/:id", ai3dUnifiedController.DeleteTask)     // 删除任务
				ai3d.GET("/config", ai3dUnifiedController.GetConfig)            // 获取配置
			} else {
				// 如果服务未初始化，返回错误
				ai3d.POST("/tasks", func(c *gin.Context) {
					response.Error(c, 500, "AI3D服务未初始化")
				})
			}
		}

		// 图片处理API
		image := api.Group("/image")
		{
			image.POST("/flipy-webp", imageController.FlipYAndToWebp) // 图片翻转并转换为WebP
		}

		// AI蓝图生成API
		blueprint := api.Group("/blueprint")
		{
			blueprint.POST("/generate", blueprintController.Generate)   // 生成蓝图
			blueprint.GET("/history", blueprintController.GetHistory)   // 获取生成历史
		}

		// 项目管理API
		projects := api.Group("/projects")
		{
			projects.GET("", projectController.GetProjects)                          // 获取项目列表
			projects.POST("", projectController.CreateProject)                       // 创建项目
			projects.GET("/:id", projectController.GetProject)                       // 获取项目详情
			projects.DELETE("/:id", projectController.DeleteProject)                 // 删除项目
			projects.POST("/:id/versions", projectController.UploadVersion)          // 上传版本
			projects.GET("/:id/versions", projectController.GetVersionHistory)       // 获取版本历史
			projects.POST("/:id/refresh-thumbnail", projectController.RefreshThumbnail) // 刷新缩略图
			projects.GET("/versions/:versionId/download", projectController.DownloadVersion) // 下载版本
			projects.POST("/versions/:versionId/rollback", projectController.RollbackVersion) // 回滚版本
		}

		// 统计API
		statistics := api.Group("/statistics")
		{
			statistics.GET("/overview", statisticsController.GetOverview)                   // 获取统计概览
			statistics.GET("/recent-activities", statisticsController.GetRecentActivities) // 获取最近活动
			statistics.GET("/system-status", statisticsController.GetSystemStatus)         // 获取系统状态
		}

		// 文件处理器API
		if fileProcessorController != nil {
			fileprocessor := api.Group("/fileprocessor")
			{
				fileprocessor.GET("/formats", fileProcessorController.GetSupportedFormats)      // 获取支持的格式
				fileprocessor.POST("/metadata", fileProcessorController.ExtractMetadata)        // 提取元数据
				fileprocessor.POST("/thumbnail", fileProcessorController.GenerateThumbnail)     // 生成缩略图
				fileprocessor.POST("/tasks", fileProcessorController.CreateTask)                // 创建任务
				fileprocessor.GET("/tasks", fileProcessorController.ListTasks)                  // 列出任务
				fileprocessor.GET("/tasks/:id", fileProcessorController.GetTask)                // 获取任务详情
				fileprocessor.POST("/tasks/:id/cancel", fileProcessorController.CancelTask)     // 取消任务
				fileprocessor.POST("/tasks/:id/retry", fileProcessorController.RetryTask)       // 重试任务
			}
		}

		// 审计日志API
		if auditController != nil {
			audit := api.Group("/audit")
			{
				audit.GET("/logs", auditController.ListLogs)                                    // 查询审计日志列表
				audit.GET("/logs/:id", auditController.GetLog)                                  // 获取单条审计日志
				audit.GET("/users/:user_id/logs", auditController.GetUserLogs)                  // 获取用户的审计日志
				audit.GET("/resources/:resource/:resource_id/logs", auditController.GetResourceLogs) // 获取资源的审计日志
				audit.GET("/statistics", auditController.GetStatistics)                         // 获取统计信息
				audit.POST("/archive", auditController.TriggerArchive)                          // 手动触发归档
				audit.GET("/archive/statistics", auditController.GetArchiveStatistics)          // 获取归档统计信息
				audit.GET("/archive/files", auditController.ListArchiveFiles)                   // 列出归档文件
			}
		}

		// TODO: 添加其他业务控制器和路由

		// ==================== 认证路由（示例）====================
		// 初始化JWT认证器
		jwtAuth := middleware.NewJWTAuth()

		auth := api.Group("/auth")
		{
			// 登录（需要实现validateUser函数）
			// auth.POST("/login", jwtAuth.LoginHandler(validateUser))
			auth.POST("/refresh", jwtAuth.RefreshHandler()) // 刷新token
			auth.POST("/logout", jwtAuth.LogoutHandler())   // 登出
		}

		// 需要认证的路由示例
		_ = jwtAuth // 使用jwtAuth.AuthMiddleware()保护需要认证的路由
		// protectedGroup := api.Group("/protected")
		// protectedGroup.Use(jwtAuth.AuthMiddleware())
		// {
		//     protectedGroup.GET("/profile", getProfile)
		// }

	}
}

package controllers

import (
	"fmt"
	"go_wails_project_manager/config"
	"go_wails_project_manager/database"
	"go_wails_project_manager/logger"
	"go_wails_project_manager/models"
	"go_wails_project_manager/response"
	"go_wails_project_manager/services/texture"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// TextureController 贴图控制器
type TextureController struct {
	db           *gorm.DB
	syncService  *texture.SyncService
	queryService *texture.QueryService
	tagService   *texture.TagService
}

// NewTextureController 创建贴图控制器
func NewTextureController() *TextureController {
	db := database.MustGetDB()
	return &TextureController{
		db:           db,
		syncService:  texture.GetGlobalSyncService(),
		queryService: texture.NewQueryService(db),
		tagService:   texture.NewTagService(db),
	}
}

// List 获取材质列表
// @Summary 获取材质列表
// @Tags Texture
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(20)
// @Param keyword query string false "搜索关键词（支持名称、描述、Asset ID）"
// @Param sortBy query string false "排序方式: use_count|date_published"
// @Param syncStatus query int false "同步状态: 0=未同步 1=同步中 2=已同步 3=失败"
// @Param textureType query string false "原始贴图类型: Diffuse|Rough|Normal 等"
// @Param threeJSType query string false "Three.js 贴图类型: map|normalMap|roughnessMap 等"
// @Success 200 {object} response.Response
// @Router /api/textures [get]
func (c *TextureController) List(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("pageSize", "20"))
	keyword := ctx.Query("keyword")
	sortBy := ctx.Query("sortBy")
	syncStatusStr := ctx.Query("syncStatus")
	textureType := ctx.Query("textureType")
	threeJSType := ctx.Query("threeJSType")

	filters := map[string]interface{}{}
	if keyword != "" {
		filters["keyword"] = keyword
	}
	if sortBy != "" {
		filters["sort_by"] = sortBy
	}
	if syncStatusStr != "" {
		if syncStatus, err := strconv.Atoi(syncStatusStr); err == nil {
			filters["sync_status"] = syncStatus
		}
	}
	if textureType != "" {
		filters["texture_type"] = textureType
	}
	if threeJSType != "" {
		filters["threejs_type"] = threeJSType
	}

	textures, total, err := c.queryService.List(page, pageSize, filters)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, "查询失败")
		return
	}

	// 为每个材质加载文件信息
	type TextureWithFiles struct {
		models.Texture
		Files []models.File `json:"files"`
	}

	result := make([]TextureWithFiles, len(textures))
	db := database.MustGetDB()
	
	for i, texture := range textures {
		result[i].Texture = texture
		
		// 查询关联的文件
		var files []models.File
		db.Where("related_id = ? AND related_type = ?", texture.ID, "Texture").
			Order("file_type ASC").
			Find(&files)
		result[i].Files = files
	}

	response.Success(ctx, gin.H{
		"list":     result,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

// GetDetail 获取材质详情
// @Summary 获取材质详情
// @Tags Texture
// @Param assetId path string true "材质ID"
// @Success 200 {object} response.Response
// @Router /api/textures/{assetId} [get]
func (c *TextureController) GetDetail(ctx *gin.Context) {
	assetID := ctx.Param("assetId")

	texture, tags, files, err := c.queryService.GetDetail(assetID)
	if err != nil {
		response.Error(ctx, http.StatusNotFound, "材质不存在")
		return
	}

	response.Success(ctx, gin.H{
		"texture": texture,
		"tags":    tags,
		"files":   files,
	})
}

// RecordUse 记录使用
// @Summary 记录材质使用
// @Tags Texture
// @Param assetId path string true "材质ID"
// @Success 200 {object} response.Response
// @Router /api/textures/{assetId}/use [post]
func (c *TextureController) RecordUse(ctx *gin.Context) {
	assetID := ctx.Param("assetId")

	texture, err := c.queryService.GetByAssetID(assetID)
	if err != nil {
		response.Error(ctx, http.StatusNotFound, "材质不存在")
		return
	}

	if err := c.queryService.IncrementUseCount(texture.ID); err != nil {
		response.Error(ctx, http.StatusInternalServerError, "记录失败")
		return
	}

	// 重新获取更新后的数据
	texture, _ = c.queryService.GetByAssetID(assetID)

	response.Success(ctx, gin.H{
		"success":   true,
		"use_count": texture.UseCount,
	})
}

// GetTags 获取标签列表
// @Summary 获取标签列表
// @Tags Texture
// @Param type query string false "标签类型: tag|category"
// @Success 200 {object} response.Response
// @Router /api/tags [get]
func (c *TextureController) GetTags(ctx *gin.Context) {
	tagType := ctx.Query("type")

	var tags []models.Tag
	var err error

	if tagType != "" {
		tags, err = c.tagService.ListByType(tagType)
	} else {
		// 获取所有标签
		tagTags, _ := c.tagService.ListByType("tag")
		categoryTags, _ := c.tagService.ListByType("category")
		tags = append(tagTags, categoryTags...)
	}

	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, "查询失败")
		return
	}

	response.Success(ctx, tags)
}

// GetTexturesByTag 按标签查询材质
// @Summary 按标签查询材质
// @Tags Texture
// @Param tagId path int true "标签ID"
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(20)
// @Success 200 {object} response.Response
// @Router /api/tags/{tagId}/textures [get]
func (c *TextureController) GetTexturesByTag(ctx *gin.Context) {
	tagID, err := strconv.ParseUint(ctx.Param("tagId"), 10, 32)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "无效的标签ID")
		return
	}

	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("pageSize", "20"))

	textures, total, err := c.queryService.ListByTag(uint(tagID), page, pageSize)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, "查询失败")
		return
	}

	response.Success(ctx, gin.H{
		"textures": textures,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

// TriggerSync 手动触发同步
// @Summary 手动触发同步
// @Tags Texture
// @Param body body object true "同步类型"
// @Success 200 {object} response.Response
// @Router /api/textures/sync [post]
func (c *TextureController) TriggerSync(ctx *gin.Context) {
	var req struct {
		Type   string `json:"type" binding:"required"`   // full | incremental | ambientcg
		Source string `json:"source"`                    // polyhaven | ambientcg (可选，用于区分数据源)
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "参数错误")
		return
	}

	// 检查同步服务是否已初始化
	if c.syncService == nil {
		response.Error(ctx, http.StatusServiceUnavailable, "同步服务未初始化")
		return
	}

	// 异步执行同步任务
	go func() {
		var err error
		
		// 如果是 AmbientCG 同步
		if req.Type == "ambientcg" || req.Source == "ambientcg" {
			logger.Log.Info("开始 AmbientCG 元数据同步...")
			ambientcgService := texture.NewAmbientCGSyncService(c.db, logger.Log)
			err = ambientcgService.SyncMetadata()
		} else if req.Type == "ambientcg_incremental" {
			// AmbientCG 增量同步
			logger.Log.Info("开始 AmbientCG 增量同步...")
			ambientcgService := texture.NewAmbientCGSyncService(c.db, logger.Log)
			err = ambientcgService.IncrementalSync()
		} else {
			// PolyHaven 同步
			if req.Type == "full" {
				err = c.syncService.FullSync()
			} else {
				err = c.syncService.IncrementalSync()
			}
		}

		if err != nil {
			logger.Log.Errorf("同步任务失败: %v", err)
		}
	}()

	response.Success(ctx, gin.H{
		"message": "同步任务已启动",
		"type":    req.Type,
		"source":  req.Source,
	})
}

// GetSyncProgress 获取同步进度
// @Summary 获取同步进度
// @Tags Texture
// @Success 200 {object} response.Response
// @Router /api/textures/sync/progress [get]
func (c *TextureController) GetSyncProgress(ctx *gin.Context) {
	log, err := c.queryService.GetLatestSyncLog()
	if err != nil {
		response.Error(ctx, http.StatusNotFound, "暂无同步记录")
		return
	}

	isRunning := log.Status == 0

	response.Success(ctx, gin.H{
		"is_running":      isRunning,
		"progress":        log.Progress,
		"current_asset":   log.CurrentAsset,
		"processed_count": log.ProcessedCount,
		"total_count":     log.TotalCount,
		"success_count":   log.SuccessCount,
		"fail_count":      log.FailCount,
		"start_time":      log.StartTime,
	})
}

// GetSyncStatus 获取同步状态
// @Summary 获取同步状态
// @Tags Texture
// @Param logId path int true "日志ID"
// @Success 200 {object} response.Response
// @Router /api/textures/sync/status/{logId} [get]
func (c *TextureController) GetSyncStatus(ctx *gin.Context) {
	logID, err := strconv.ParseUint(ctx.Param("logId"), 10, 32)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "无效的日志ID")
		return
	}

	log, err := c.queryService.GetSyncProgress(uint(logID))
	if err != nil {
		response.Error(ctx, http.StatusNotFound, "日志不存在")
		return
	}

	response.Success(ctx, log)
}

// GetSyncLogs 获取同步日志列表
// @Summary 获取同步日志列表
// @Tags Texture
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(20)
// @Param status query int false "状态: 0=进行中 1=成功 2=失败"
// @Param syncType query string false "类型: full|incremental"
// @Success 200 {object} response.Response
// @Router /api/textures/sync/logs [get]
func (c *TextureController) GetSyncLogs(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("pageSize", "20"))

	filters := map[string]interface{}{}
	if status := ctx.Query("status"); status != "" {
		if s, err := strconv.Atoi(status); err == nil {
			filters["status"] = s
		}
	}
	if syncType := ctx.Query("syncType"); syncType != "" {
		filters["sync_type"] = syncType
	}

	logs, total, err := c.queryService.ListSyncLogs(page, pageSize, filters)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, "查询失败")
		return
	}

	response.Success(ctx, gin.H{
		"logs":     logs,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

// GetTextureTypes 获取所有贴图类型
// @Summary 获取所有贴图类型
// @Tags Texture
// @Success 200 {object} response.Response
// @Router /api/textures/types [get]
func (c *TextureController) GetTextureTypes(ctx *gin.Context) {
	types, err := c.queryService.GetAllTextureTypes()
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, "查询失败")
		return
	}

	response.Success(ctx, gin.H{
		"types": types,
		"count": len(types),
	})
}

// GetThreeJSTypes 获取 Three.js 贴图类型
// @Summary 获取 Three.js 贴图类型
// @Tags Texture
// @Success 200 {object} response.Response
// @Router /api/textures/types/threejs [get]
func (c *TextureController) GetThreeJSTypes(ctx *gin.Context) {
	if config.TextureMapping == nil {
		response.Error(ctx, http.StatusServiceUnavailable, "贴图映射配置未加载")
		return
	}

	typeInfo := config.TextureMapping.GetThreeJSTypeInfo()

	response.Success(ctx, gin.H{
		"types": typeInfo,
		"count": len(typeInfo),
	})
}


// DownloadTexture 触发材质下载（统一按需下载）
// @Summary 触发材质下载
// @Tags Texture
// @Param assetId path string true "Asset ID"
// @Param body body object false "下载选项"
// @Success 200 {object} response.Response
// @Router /api/textures/download/:assetId [post]
func (c *TextureController) DownloadTexture(ctx *gin.Context) {
	assetID := ctx.Param("assetId")
	if assetID == "" {
		response.Error(ctx, http.StatusBadRequest, "Asset ID 不能为空")
		return
	}

	// 解析下载选项
	var opts texture.DownloadOptions
	opts.Resolution = ctx.DefaultQuery("resolution", "2K")
	opts.Format = ctx.DefaultQuery("format", "JPG")

	// 也支持从 body 中获取
	if ctx.Request.ContentLength > 0 {
		ctx.ShouldBindJSON(&opts)
	}

	// 创建统一下载服务（自动识别数据源）
	downloadService := texture.NewUnifiedDownloadService(c.db, logger.Log)

	// 执行下载
	files, err := downloadService.DownloadTexture(assetID, opts)
	if err != nil {
		logger.Log.Errorf("下载材质失败: %v", err)
		response.Error(ctx, http.StatusInternalServerError, fmt.Sprintf("下载失败: %v", err))
		return
	}

	response.Success(ctx, gin.H{
		"message":            "下载成功",
		"asset_id":           assetID,
		"download_completed": true,
		"files":              files,
		"file_count":         len(files),
	})
}

// CheckDownloadStatus 检查下载状态
// @Summary 检查材质下载状态
// @Tags Texture
// @Param assetId path string true "Asset ID"
// @Success 200 {object} response.Response
// @Router /api/textures/download-status/:assetId [get]
func (c *TextureController) CheckDownloadStatus(ctx *gin.Context) {
	assetID := ctx.Param("assetId")
	if assetID == "" {
		response.Error(ctx, http.StatusBadRequest, "Asset ID 不能为空")
		return
	}

	// 查询材质
	var texture models.Texture
	if err := c.db.Where("asset_id = ?", assetID).First(&texture).Error; err != nil {
		response.Error(ctx, http.StatusNotFound, "材质不存在")
		return
	}

	// 如果已下载，获取文件列表
	var files []models.File
	if texture.DownloadCompleted {
		c.db.Where("related_id = ? AND related_type = ? AND file_type = ?",
			texture.ID, "Texture", "texture").Find(&files)
	}

	response.Success(ctx, gin.H{
		"asset_id":           texture.AssetID,
		"download_completed": texture.DownloadCompleted,
		"has_preview":        true, // 元数据同步时已下载预览图
		"source":             texture.Source,
		"files":              files,
		"file_count":         len(files),
	})
}

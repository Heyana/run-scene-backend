package controllers

import (
	"go_wails_project_manager/database"
	"go_wails_project_manager/logger"
	"go_wails_project_manager/models"
	"go_wails_project_manager/response"
	"go_wails_project_manager/services/texture"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// TextureController 贴图控制器
type TextureController struct {
	syncService  *texture.SyncService
	queryService *texture.QueryService
	tagService   *texture.TagService
}

// NewTextureController 创建贴图控制器
func NewTextureController() *TextureController {
	db := database.MustGetDB()
	return &TextureController{
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
// @Param keyword query string false "搜索关键词"
// @Param sortBy query string false "排序方式: use_count|date_published"
// @Success 200 {object} response.Response
// @Router /api/textures [get]
func (c *TextureController) List(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("pageSize", "20"))
	keyword := ctx.Query("keyword")
	sortBy := ctx.Query("sortBy")

	filters := map[string]interface{}{}
	if keyword != "" {
		filters["keyword"] = keyword
	}
	if sortBy != "" {
		filters["sort_by"] = sortBy
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
		Type string `json:"type" binding:"required"` // full | incremental
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
		if req.Type == "full" {
			err = c.syncService.FullSync()
		} else {
			err = c.syncService.IncrementalSync()
		}

		if err != nil {
			logger.Log.Errorf("同步任务失败: %v", err)
		}
	}()

	response.Success(ctx, gin.H{
		"message": "同步任务已启动",
		"type":    req.Type,
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

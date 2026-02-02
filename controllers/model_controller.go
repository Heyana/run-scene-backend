package controllers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"go_wails_project_manager/config"
	"go_wails_project_manager/response"
	modelService "go_wails_project_manager/services/model"
)

type ModelController struct {
	uploadService *modelService.UploadService
	queryService  *modelService.QueryService
}

func NewModelController(db *gorm.DB) *ModelController {
	return &ModelController{
		uploadService: modelService.NewUploadService(db, &config.AppConfig.Model),
		queryService:  modelService.NewQueryService(db),
	}
}

// Upload 上传模型
func (c *ModelController) Upload(ctx *gin.Context) {
	// 获取文件
	modelFile, err := ctx.FormFile("model")
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "缺少模型文件")
		return
	}

	thumbnailFile, err := ctx.FormFile("thumbnail")
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "缺少预览图")
		return
	}

	// 获取元数据
	name := ctx.PostForm("name")
	if name == "" {
		response.Error(ctx, http.StatusBadRequest, "缺少模型名称")
		return
	}

	fileType := ctx.PostForm("type")
	if fileType == "" {
		response.Error(ctx, http.StatusBadRequest, "缺少文件类型")
		return
	}

	description := ctx.PostForm("description")
	category := ctx.PostForm("category")
	tagsStr := ctx.PostForm("tags")
	
	var tags []string
	if tagsStr != "" {
		tags = strings.Split(tagsStr, ",")
		// 去除空格
		for i := range tags {
			tags[i] = strings.TrimSpace(tags[i])
		}
	}

	// 获取上传者信息
	uploadedBy := ctx.GetString("username")
	if uploadedBy == "" {
		uploadedBy = "anonymous"
	}
	uploadIP := ctx.ClientIP()

	// 上传
	metadata := modelService.UploadMetadata{
		Name:        name,
		Description: description,
		Category:    category,
		Tags:        tags,
		Type:        fileType,
		UploadedBy:  uploadedBy,
		UploadIP:    uploadIP,
	}

	model, err := c.uploadService.UploadSingle(modelFile, thumbnailFile, metadata)
	if err != nil {
		if err == modelService.ErrDuplicateFile {
			response.Success(ctx, gin.H{
				"message": "文件已存在",
				"model":   model,
			})
			return
		}
		if err == modelService.ErrInvalidType {
			response.Error(ctx, http.StatusUnsupportedMediaType, "不支持的文件类型")
			return
		}
		if err == modelService.ErrFileTooLarge {
			response.Error(ctx, http.StatusRequestEntityTooLarge, "文件过大")
			return
		}
		response.Error(ctx, http.StatusInternalServerError, "上传失败")
		return
	}

	response.Success(ctx, model)
}

// List 模型列表
func (c *ModelController) List(ctx *gin.Context) {
	// 分页参数
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("pageSize", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// 过滤参数
	filters := modelService.QueryFilters{
		Category:  ctx.Query("category"),
		Type:      ctx.Query("type"),
		SortBy:    ctx.DefaultQuery("sortBy", "created_at"),
		SortOrder: ctx.DefaultQuery("sortOrder", "desc"),
	}

	// 标签过滤
	if tagsStr := ctx.Query("tags"); tagsStr != "" {
		filters.Tags = strings.Split(tagsStr, ",")
		for i := range filters.Tags {
			filters.Tags[i] = strings.TrimSpace(filters.Tags[i])
		}
	}

	// 查询
	models, total, err := c.queryService.List(page, pageSize, filters)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, "查询失败")
		return
	}

	response.Success(ctx, gin.H{
		"list":     models,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

// GetDetail 模型详情
func (c *ModelController) GetDetail(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "无效的ID")
		return
	}

	model, tags, err := c.queryService.GetDetail(uint(id))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Error(ctx, http.StatusNotFound, "模型不存在")
			return
		}
		response.Error(ctx, http.StatusInternalServerError, "查询失败")
		return
	}

	response.Success(ctx, gin.H{
		"model": model,
		"tags":  tags,
	})
}

// Search 搜索模型
func (c *ModelController) Search(ctx *gin.Context) {
	keyword := ctx.Query("keyword")
	if keyword == "" {
		response.Error(ctx, http.StatusBadRequest, "缺少搜索关键词")
		return
	}

	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("pageSize", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	models, total, err := c.queryService.Search(keyword, page, pageSize)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, "搜索失败")
		return
	}

	response.Success(ctx, gin.H{
		"list":     models,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

// IncrementUseCount 记录使用
func (c *ModelController) IncrementUseCount(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "无效的ID")
		return
	}

	if err := c.queryService.IncrementUseCount(uint(id)); err != nil {
		response.Error(ctx, http.StatusInternalServerError, "记录失败")
		return
	}

	// 获取更新后的使用次数
	model, _, _ := c.queryService.GetDetail(uint(id))
	
	response.Success(ctx, gin.H{
		"message":   "记录成功",
		"use_count": model.UseCount,
	})
}

// Delete 删除模型
func (c *ModelController) Delete(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "无效的ID")
		return
	}

	if err := c.queryService.Delete(uint(id)); err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Error(ctx, http.StatusNotFound, "模型不存在")
			return
		}
		response.Error(ctx, http.StatusInternalServerError, "删除失败")
		return
	}

	response.Success(ctx, gin.H{
		"message": "删除成功",
	})
}

// GetStatistics 获取统计信息
func (c *ModelController) GetStatistics(ctx *gin.Context) {
	stats, err := c.queryService.GetStatistics()
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, "查询失败")
		return
	}

	response.Success(ctx, stats)
}

// GetPopular 获取热门模型
func (c *ModelController) GetPopular(ctx *gin.Context) {
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))
	if limit < 1 || limit > 50 {
		limit = 10
	}

	models, err := c.queryService.GetPopular(limit)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, "查询失败")
		return
	}

	response.Success(ctx, models)
}

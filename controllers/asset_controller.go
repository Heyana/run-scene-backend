package controllers

import (
	"go_wails_project_manager/config"
	"go_wails_project_manager/response"
	"go_wails_project_manager/services/asset"
	"net/http"
	"strconv"
	"strings"
	
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// AssetController 资产控制器
type AssetController struct {
	uploadService *asset.UploadService
	queryService  *asset.QueryService
}

// NewAssetController 创建资产控制器
func NewAssetController(db *gorm.DB) *AssetController {
	return &AssetController{
		uploadService: asset.NewUploadService(db, &config.AppConfig.Asset),
		queryService:  asset.NewQueryService(db),
	}
}

// Upload 上传资产
// @Summary 上传资产
// @Tags 资产管理
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "资产文件"
// @Param name formData string true "资产名称"
// @Param description formData string false "描述"
// @Param category formData string false "分类"
// @Param tags formData string false "标签(逗号分隔)"
// @Success 200 {object} response.Response
// @Router /api/assets/upload [post]
func (c *AssetController) Upload(ctx *gin.Context) {
	// 获取上传文件
	file, err := ctx.FormFile("file")
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "未找到上传文件")
		return
	}
	
	// 获取元数据
	metadata := asset.UploadMetadata{
		Name:        ctx.PostForm("name"),
		Description: ctx.PostForm("description"),
		Category:    ctx.PostForm("category"),
		UploadedBy:  ctx.GetString("username"), // 从上下文获取用户名
		UploadIP:    ctx.ClientIP(),
	}
	
	// 解析标签
	if tagsStr := ctx.PostForm("tags"); tagsStr != "" {
		metadata.Tags = strings.Split(tagsStr, ",")
	}
	
	// 验证必填字段
	if metadata.Name == "" {
		response.Error(ctx, http.StatusBadRequest, "资产名称不能为空")
		return
	}
	
	// 上传文件（type 由后端自动检测）
	uploadedAsset, err := c.uploadService.Upload(file, metadata)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, "上传失败: "+err.Error())
		return
	}
	
	response.SuccessWithMsg(ctx, "上传成功", uploadedAsset)
}

// List 资产列表
// @Summary 资产列表
// @Tags 资产管理
// @Produce json
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(20)
// @Param type query string false "资产类型"
// @Param category query string false "分类"
// @Param format query string false "格式"
// @Param keyword query string false "关键词"
// @Param sortBy query string false "排序字段"
// @Param sortOrder query string false "排序方向"
// @Success 200 {object} response.Response
// @Router /api/assets [get]
func (c *AssetController) List(ctx *gin.Context) {
	// 获取分页参数
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("pageSize", "20"))
	
	// 构建过滤器
	filters := asset.QueryFilters{
		Type:      ctx.Query("type"),
		Category:  ctx.Query("category"),
		Format:    ctx.Query("format"),
		Keyword:   ctx.Query("keyword"),
		SortBy:    ctx.Query("sortBy"),
		SortOrder: ctx.Query("sortOrder"),
	}
	
	// 解析标签
	if tagsStr := ctx.Query("tags"); tagsStr != "" {
		filters.Tags = strings.Split(tagsStr, ",")
	}
	
	// 查询
	assets, total, err := c.queryService.List(page, pageSize, filters)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, "查询失败: "+err.Error())
		return
	}
	
	response.SuccessWithPagination(ctx, assets, total, page, pageSize)
}

// GetDetail 获取资产详情
// @Summary 获取资产详情
// @Tags 资产管理
// @Produce json
// @Param id path int true "资产ID"
// @Success 200 {object} response.Response
// @Router /api/assets/{id} [get]
func (c *AssetController) GetDetail(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "无效的资产ID")
		return
	}
	
	asset, metadata, err := c.queryService.GetDetail(uint(id))
	if err != nil {
		response.Error(ctx, http.StatusNotFound, "资产不存在")
		return
	}
	
	result := map[string]interface{}{
		"asset":    asset,
		"metadata": metadata,
	}
	
	response.Success(ctx, result)
}

// Delete 删除资产
// @Summary 删除资产
// @Tags 资产管理
// @Produce json
// @Param id path int true "资产ID"
// @Success 200 {object} response.Response
// @Router /api/assets/{id} [delete]
func (c *AssetController) Delete(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "无效的资产ID")
		return
	}
	
	if err := c.queryService.Delete(uint(id)); err != nil {
		response.Error(ctx, http.StatusInternalServerError, "删除失败: "+err.Error())
		return
	}
	
	response.SuccessWithMsg(ctx, "删除成功", nil)
}

// Update 更新资产信息
// @Summary 更新资产信息
// @Tags 资产管理
// @Accept json
// @Produce json
// @Param id path int true "资产ID"
// @Param body body object true "更新内容"
// @Success 200 {object} response.Response
// @Router /api/assets/{id} [put]
func (c *AssetController) Update(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "无效的资产ID")
		return
	}
	
	var updates map[string]interface{}
	if err := ctx.ShouldBindJSON(&updates); err != nil {
		response.Error(ctx, http.StatusBadRequest, "请求参数错误")
		return
	}
	
	if err := c.queryService.Update(uint(id), updates); err != nil {
		response.Error(ctx, http.StatusInternalServerError, "更新失败: "+err.Error())
		return
	}
	
	response.SuccessWithMsg(ctx, "更新成功", nil)
}

// IncrementUseCount 记录使用
// @Summary 记录使用
// @Tags 资产管理
// @Produce json
// @Param id path int true "资产ID"
// @Success 200 {object} response.Response
// @Router /api/assets/{id}/use [post]
func (c *AssetController) IncrementUseCount(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "无效的资产ID")
		return
	}
	
	if err := c.queryService.IncrementUseCount(uint(id)); err != nil {
		response.Error(ctx, http.StatusInternalServerError, "记录失败: "+err.Error())
		return
	}
	
	// 获取更新后的使用次数
	asset, _, _ := c.queryService.GetDetail(uint(id))
	
	response.SuccessWithMsg(ctx, "记录成功", map[string]interface{}{
		"use_count": asset.UseCount,
	})
}

// GetStatistics 获取统计信息
// @Summary 获取统计信息
// @Tags 资产管理
// @Produce json
// @Param type query string false "资产类型"
// @Success 200 {object} response.Response
// @Router /api/assets/statistics [get]
func (c *AssetController) GetStatistics(ctx *gin.Context) {
	assetType := ctx.Query("type")
	
	stats, err := c.queryService.GetStatistics(assetType)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, "查询失败: "+err.Error())
		return
	}
	
	response.Success(ctx, stats)
}

// GetPopular 获取热门资产
// @Summary 获取热门资产
// @Tags 资产管理
// @Produce json
// @Param limit query int false "数量限制" default(10)
// @Param type query string false "资产类型"
// @Success 200 {object} response.Response
// @Router /api/assets/popular [get]
func (c *AssetController) GetPopular(ctx *gin.Context) {
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))
	assetType := ctx.Query("type")
	
	assets, err := c.queryService.GetPopular(limit, assetType)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, "查询失败: "+err.Error())
		return
	}
	
	response.Success(ctx, assets)
}

// GetStatisticsByType 按类型获取统计信息
// @Summary 按类型获取统计信息
// @Tags 资产管理
// @Produce json
// @Success 200 {object} response.Response
// @Router /api/assets/statistics/by-type [get]
func (c *AssetController) GetStatisticsByType(ctx *gin.Context) {
	stats, err := c.queryService.GetStatisticsByType()
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, "查询失败: "+err.Error())
		return
	}
	
	response.Success(ctx, stats)
}

// RegisterAssetRoutes 注册资产路由
func RegisterAssetRoutes(router *gin.Engine, db *gorm.DB) {
	controller := NewAssetController(db)
	
	api := router.Group("/api/assets")
	{
		api.POST("/upload", controller.Upload)
		api.GET("", controller.List)
		api.GET("/:id", controller.GetDetail)
		api.PUT("/:id", controller.Update)
		api.DELETE("/:id", controller.Delete)
		api.POST("/:id/use", controller.IncrementUseCount)
		api.GET("/statistics", controller.GetStatistics)
		api.GET("/statistics/by-type", controller.GetStatisticsByType)
		api.GET("/popular", controller.GetPopular)
	}
}

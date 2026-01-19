package api

import (
	"fmt"
	"go_wails_project_manager/database"
	"go_wails_project_manager/logger"
	"go_wails_project_manager/models"
	"go_wails_project_manager/response"
	"go_wails_project_manager/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ResourceFile 资源文件结构
type ResourceFile struct {
	ID          uint   `json:"id" example:"1"`
	FileName    string `json:"file_name" example:"example.jpg"`
	FileSize    int64  `json:"file_size" example:"1048576"`
	MimeType    string `json:"mime_type" example:"image/jpeg"`
	Type        string `json:"type" example:"image"`
	StorageType string `json:"storage_type" example:"cdn"`
	URL         string `json:"url" example:"https://cdn.example.com/files/example.jpg"`
	CreatedAt   int64  `json:"created_at" example:"1640995200"`
}

// UploadResourceResponse 资源上传响应
type UploadResourceResponse struct {
	ResourceID uint   `json:"resource_id" example:"1"`
	FileName   string `json:"file_name" example:"example.jpg"`
	FileSize   int64  `json:"file_size" example:"1048576"`
	MimeType   string `json:"mime_type" example:"image/jpeg"`
	Type       string `json:"type" example:"image"`
	URL        string `json:"url" example:"https://cdn.example.com/files/example.jpg"`
}

// ResourceListResponse 资源列表响应
type ResourceListResponse struct {
	Total     int64          `json:"total" example:"100"`
	Page      int            `json:"page" example:"1"`
	PageSize  int            `json:"page_size" example:"10"`
	Resources []ResourceFile `json:"resources"`
}

// ResourceController 资源控制器
type ResourceController struct {
	resourceService *services.ResourceService
}

// NewResourceController 创建资源控制器
func NewResourceController() *ResourceController {
	return &ResourceController{
		resourceService: services.NewResourceService(),
	}
}

// UploadResource 上传资源
// @Summary 上传资源文件
// @Description 上传各类资源文件（图片、文档、音频、视频等）到系统
// @Tags 资源管理
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "资源文件（最大10MB）"
// @Param library_id query int false "媒体库ID，可选"
// @Param resource_type formData string false "资源类型（系统自动识别）"
// @Success 200 {object} response.Response{data=UploadResourceResponse} "上传成功"
// @Failure 400 {object} response.Response "文件上传失败或文件过大"
// @Failure 500 {object} response.Response "保存失败"
// @Router /api/resources/upload [post]
func (c *ResourceController) UploadResource(ctx *gin.Context) {
	// 获取可选的媒体库ID
	var libraryID *uint
	if libIDStr := ctx.DefaultQuery("library_id", ""); libIDStr != "" {
		if id, err := strconv.ParseUint(libIDStr, 10, 32); err == nil {
			uid := uint(id)
			libraryID = &uid
		}
	}

	// 获取上传的文件
	file, header, err := ctx.Request.FormFile("file")
	if err != nil {
		response.BadRequest(ctx, "文件上传失败: "+err.Error())
		return
	}
	defer file.Close()

	// 检查文件大小
	if header.Size > 10*1024*1024 { // 10MB
		response.BadRequest(ctx, "文件大小超过限制(10MB)")
		return
	}

	// 保存资源
	resource, err := c.resourceService.SaveResource(ctx, file, header, libraryID)
	if err != nil {
		logger.Log.Errorf("资源保存失败: %v", err)
		response.InternalServerError(ctx, "资源保存失败: "+err.Error())
		return
	}

	// 动态拼接完整 URL
	fullURL := c.resourceService.GetFullURL(resource.FilePath)

	response.Success(ctx, gin.H{
		"resource_id": resource.ID,
		"file_name":   resource.FileName,
		"file_size":   resource.FileSize,
		"mime_type":   resource.MimeType,
		"type":        resource.Type,
		"url":         fullURL,
	})
}

// GetResource 获取资源
// @Summary 获取资源文件
// @Description 根据ID获取资源文件，返回文件内容或重定向到CDN地址
// @Tags 资源管理
// @Produce application/octet-stream,image/*,text/*
// @Param id path string true "资源ID（示例: 1）"
// @Success 200 {file} binary "资源文件内容"
// @Success 302 {string} string "CDN重定向"
// @Failure 400 {object} response.Response "无效的资源ID"
// @Failure 404 {object} response.Response "资源不存在或内容不可用"
// @Router /api/resources/{id} [get]
func (c *ResourceController) GetResource(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(ctx, "无效的资源ID")
		return
	}

	// 获取资源
	resource, err := c.resourceService.GetResource(uint(id))
	if err != nil {
		response.NotFound(ctx, "资源不存在")
		return
	}

	// 如果是数据库存储的内容，直接返回
	if resource.StorageType == "db" && len(resource.Content) > 0 {
		ctx.Data(http.StatusOK, resource.MimeType, resource.Content)
		return
	}

	// 如果是CDN存储，重定向到CDN URL
	if resource.StorageType == "cdn" && resource.FilePath != "" {
		ctx.Redirect(http.StatusFound, resource.FilePath)
		return
	}

	response.NotFound(ctx, "资源内容不可用")
}

// DeleteResource 删除资源
// @Summary 删除资源文件
// @Description 永久删除指定的资源文件，包括本地和CDN存储
// @Tags 资源管理
// @Accept json
// @Produce json
// @Param id path string true "资源ID（示例: 1）"
// @Success 200 {object} response.Response "删除成功"
// @Failure 400 {object} response.Response "无效的资源ID"
// @Failure 404 {object} response.Response "资源不存在"
// @Failure 500 {object} response.Response "删除失败"
// @Router /api/resources/{id} [delete]
func (c *ResourceController) DeleteResource(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(ctx, "无效的资源ID")
		return
	}

	// 获取资源（简化权限检查）
	resource, err := c.resourceService.GetResource(uint(id))
	if err != nil {
		response.NotFound(ctx, "资源不存在")
		return
	}

	// TODO: 根据需要添加权限检查
	_ = resource

	// 删除资源
	if err := c.resourceService.DeleteResource(ctx, uint(id)); err != nil {
		response.InternalServerError(ctx, "删除资源失败: "+err.Error())
		return
	}

	response.Success(ctx, gin.H{"message": "资源删除成功"})
}

// ListResources 列出资源
// @Summary 获取资源文件列表
// @Description 获取系统中所有资源文件的列表，支持分页和类型筛选
// @Tags 资源管理
// @Accept json
// @Produce json
// @Param page query int false "页码，默认1" default(1)
// @Param page_size query int false "每页数量，默认10，最大100" default(10)
// @Param type query string false "资源类型筛选：image | video | audio | document"
// @Param library_id query int false "媒体库ID筛选"
// @Success 200 {object} response.Response{data=ResourceListResponse} "获取成功"
// @Failure 500 {object} response.Response "查询失败"
// @Router /api/resources [get]
func (c *ResourceController) ListResources(ctx *gin.Context) {
	// 获取查询参数
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "10"))
	resourceType := ctx.DefaultQuery("type", "")

	// 分页参数校验
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// 查询资源
	db, err := getCurrentDB(ctx)
	if err != nil {
		response.InternalServerError(ctx, "数据库连接失败")
		return
	}

	var resources []models.ResourceFile
	query := db.Model(&models.ResourceFile{})

	// 可选：按媒体库筛选
	if libIDStr := ctx.DefaultQuery("library_id", ""); libIDStr != "" {
		if libID, err := strconv.ParseUint(libIDStr, 10, 32); err == nil {
			query = query.Where("library_id = ?", uint(libID))
		}
	}

	if resourceType != "" {
		query = query.Where("type = ?", resourceType)
	}

	var total int64
	query.Model(&models.ResourceFile{}).Count(&total)

	offset := (page - 1) * pageSize
	err = query.Order("created_at desc").
		Offset(offset).
		Limit(pageSize).
		Find(&resources).Error
	if err != nil {
		response.InternalServerError(ctx, "查询资源失败")
		return
	}

	// 构建响应
	var result []gin.H
	for _, resource := range resources {
		// 动态拼接完整 URL
		fullURL := c.resourceService.GetFullURL(resource.FilePath)

		result = append(result, gin.H{
			"id":           resource.ID,
			"file_name":    resource.FileName,
			"file_size":    resource.FileSize,
			"mime_type":    resource.MimeType,
			"type":         resource.Type,
			"storage_type": resource.StorageType,
			"url":          fullURL,
			"created_at":   resource.CreatedAt,
		})
	}

	response.Success(ctx, gin.H{
		"total":     total,
		"page":      page,
		"page_size": pageSize,
		"resources": result,
	})
}

// getCurrentEnterpriseID 获取当前企业ID
func getCurrentEnterpriseID(ctx *gin.Context) (uint, error) {
	// 从认证中获取企业ID
	value, exists := ctx.Get("enterprise_id")
	if !exists {
		// 临时方案：如果没有认证信息，使用请求参数中的enterprise_id
		idStr := ctx.DefaultQuery("enterprise_id", "")
		if idStr == "" {
			idStr = ctx.PostForm("enterprise_id")
		}

		if idStr != "" {
			id, err := strconv.ParseUint(idStr, 10, 32)
			if err != nil {
				return 0, fmt.Errorf("无效的企业ID")
			}
			return uint(id), nil
		}

		return 0, fmt.Errorf("未找到企业ID")
	}

	id, ok := value.(uint)
	if !ok {
		return 0, fmt.Errorf("无效的企业ID类型")
	}

	return id, nil
}

// getCurrentDB 获取当前数据库连接
func getCurrentDB(ctx *gin.Context) (*gorm.DB, error) {
	// 从上下文获取数据库连接
	value, exists := ctx.Get("db")
	if !exists {
		// 如果上下文中没有，使用默认连接
		db, err := database.GetDB()
		if err != nil {
			return nil, err
		}
		return db, nil
	}

	db, ok := value.(*gorm.DB)
	if !ok {
		return nil, fmt.Errorf("无效的数据库连接类型")
	}

	return db, nil
}

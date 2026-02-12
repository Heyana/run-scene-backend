package controllers

import (
	"encoding/json"
	"go_wails_project_manager/config"
	"go_wails_project_manager/response"
	"go_wails_project_manager/services/document"
	"go_wails_project_manager/services/fileprocessor"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// DocumentController 文档控制器
type DocumentController struct {
	uploadService        *document.UploadService
	queryService         *document.QueryService
	config               *config.DocumentConfig
	fileProcessorService *fileprocessor.FileProcessorService
}

// NewDocumentController 创建文档控制器
func NewDocumentController(db *gorm.DB, fpService *fileprocessor.FileProcessorService, fpConfig *fileprocessor.Config) *DocumentController {
	docConfig, err := config.LoadDocumentConfig()
	if err != nil {
		// 使用默认配置
		docConfig = &config.DocumentConfig{}
	}

	uploadService := document.NewUploadService(db, docConfig)
	
	// 设置文件处理器服务和配置
	if fpService != nil {
		uploadService.SetFileProcessorService(fpService)
	}
	if fpConfig != nil {
		uploadService.SetFileProcessorConfig(fpConfig)
	}

	return &DocumentController{
		uploadService:        uploadService,
		queryService:         document.NewQueryService(db),
		config:               docConfig,
		fileProcessorService: fpService,
	}
}

// Upload 上传文档
// @Summary 上传文档
// @Tags 文件库
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "文档文件"
// @Param name formData string true "文档名称"
// @Param description formData string false "描述"
// @Param category formData string false "分类"
// @Param tags formData string false "标签(逗号分隔)"
// @Param department formData string false "部门"
// @Param project formData string false "项目"
// @Param is_public formData boolean false "是否公开"
// @Param version formData string false "版本号"
// @Success 200 {object} response.Response
// @Router /api/documents/upload [post]
func (c *DocumentController) Upload(ctx *gin.Context) {
	// 获取上传文件
	file, err := ctx.FormFile("file")
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "未找到上传文件")
		return
	}

	// 获取元数据
	metadata := document.UploadMetadata{
		Name:        ctx.PostForm("name"),
		Description: ctx.PostForm("description"),
		Category:    ctx.PostForm("category"),
		Department:  ctx.PostForm("department"),
		Project:     ctx.PostForm("project"),
		Version:     ctx.PostForm("version"),
		UploadedBy:  ctx.GetString("username"), // 从上下文获取用户名
		UploadIP:    ctx.ClientIP(),
	}

	// 解析 parent_id
	if parentIDStr := ctx.PostForm("parent_id"); parentIDStr != "" {
		parentID, err := strconv.ParseUint(parentIDStr, 10, 32)
		if err == nil {
			pid := uint(parentID)
			metadata.ParentID = &pid
		}
	}

	// 解析标签
	if tagsStr := ctx.PostForm("tags"); tagsStr != "" {
		metadata.Tags = strings.Split(tagsStr, ",")
	}

	// 解析是否公开
	if isPublicStr := ctx.PostForm("is_public"); isPublicStr != "" {
		metadata.IsPublic = isPublicStr == "true" || isPublicStr == "1"
	}

	// 验证必填字段
	if metadata.Name == "" {
		response.Error(ctx, http.StatusBadRequest, "文档名称不能为空")
		return
	}

	// 上传文件
	uploadedDoc, err := c.uploadService.Upload(file, metadata)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, "上传失败: "+err.Error())
		return
	}

	response.SuccessWithMsg(ctx, "上传成功", uploadedDoc)
}

// UploadFolder 上传文件夹（保持结构）
// @Summary 上传文件夹
// @Tags 文件库
// @Accept multipart/form-data
// @Produce json
// @Param files formData file true "文件列表"
// @Param file_paths formData string true "文件路径列表(JSON数组)"
// @Param parent_id formData int false "父文件夹ID"
// @Param description formData string false "描述"
// @Param category formData string false "分类"
// @Param department formData string false "部门"
// @Param project formData string false "项目"
// @Success 200 {object} response.Response
// @Router /api/documents/upload-folder [post]
func (c *DocumentController) UploadFolder(ctx *gin.Context) {
	// 解析文件路径列表
	filePathsJSON := ctx.PostForm("file_paths")
	var filePaths []string
	if err := json.Unmarshal([]byte(filePathsJSON), &filePaths); err != nil {
		response.Error(ctx, http.StatusBadRequest, "文件路径解析失败: "+err.Error())
		return
	}

	// 获取上传的文件
	form, err := ctx.MultipartForm()
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "获取文件失败: "+err.Error())
		return
	}

	files := form.File["files"]
	if len(files) == 0 {
		response.Error(ctx, http.StatusBadRequest, "没有上传文件")
		return
	}

	if len(files) != len(filePaths) {
		response.Error(ctx, http.StatusBadRequest, "文件数量与路径数量不匹配")
		return
	}

	// 解析 parent_id
	var parentID *uint
	if parentIDStr := ctx.PostForm("parent_id"); parentIDStr != "" {
		pid, err := strconv.ParseUint(parentIDStr, 10, 32)
		if err == nil {
			p := uint(pid)
			parentID = &p
		}
	}

	// 获取元数据
	metadata := document.FolderUploadMetadata{
		Description: ctx.PostForm("description"),
		Category:    ctx.PostForm("category"),
		Department:  ctx.PostForm("department"),
		Project:     ctx.PostForm("project"),
		UploadedBy:  ctx.GetString("username"),
		UploadIP:    ctx.ClientIP(),
		ParentID:    parentID,
	}

	// 解析标签
	if tagsStr := ctx.PostForm("tags"); tagsStr != "" {
		metadata.Tags = strings.Split(tagsStr, ",")
	}

	// 上传文件夹
	result, err := c.uploadService.UploadFolder(files, filePaths, metadata)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, "上传失败: "+err.Error())
		return
	}

	response.SuccessWithMsg(ctx, "上传成功", result)
}

// CreateFolder 创建文件夹
// @Summary 创建文件夹
// @Tags 文件库
// @Accept json
// @Produce json
// @Param body body object true "文件夹信息"
// @Success 200 {object} response.Response
// @Router /api/documents/folder [post]
func (c *DocumentController) CreateFolder(ctx *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
		ParentID    *uint  `json:"parent_id"`
		Department  string `json:"department"`
		Project     string `json:"project"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	// 创建文件夹记录
	folder, err := c.uploadService.CreateFolder(req.Name, req.Description, req.ParentID, req.Department, req.Project, ctx.GetString("username"), ctx.ClientIP())
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, "创建失败: "+err.Error())
		return
	}

	response.SuccessWithMsg(ctx, "创建成功", folder)
}

// List 文档列表
// @Summary 文档列表
// @Tags 文件库
// @Produce json
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(20)
// @Param type query string false "文档类型"
// @Param category query string false "分类"
// @Param format query string false "格式"
// @Param department query string false "部门"
// @Param project query string false "项目"
// @Param keyword query string false "关键词"
// @Param sortBy query string false "排序字段"
// @Param sortOrder query string false "排序方向"
// @Success 200 {object} response.Response
// @Router /api/documents [get]
func (c *DocumentController) List(ctx *gin.Context) {
	// 获取分页参数
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("pageSize", "20"))

	// 构建过滤器
	filters := document.QueryFilters{
		Type:       ctx.Query("type"),
		Category:   ctx.Query("category"),
		Format:     ctx.Query("format"),
		Department: ctx.Query("department"),
		Project:    ctx.Query("project"),
		Keyword:    ctx.Query("keyword"),
		SortBy:     ctx.Query("sortBy"),
		SortOrder:  ctx.Query("sortOrder"),
	}

	// 解析 parent_id
	if parentIDStr := ctx.Query("parent_id"); parentIDStr != "" {
		parentID, err := strconv.ParseUint(parentIDStr, 10, 32)
		if err == nil {
			pid := uint(parentID)
			filters.ParentID = &pid
		}
	}

	// 解析标签
	if tagsStr := ctx.Query("tags"); tagsStr != "" {
		filters.Tags = strings.Split(tagsStr, ",")
	}

	// 解析是否公开
	if isPublicStr := ctx.Query("is_public"); isPublicStr != "" {
		isPublic := isPublicStr == "true" || isPublicStr == "1"
		filters.IsPublic = &isPublic
	}

	// 查询
	documents, total, err := c.queryService.List(page, pageSize, filters)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, "查询失败: "+err.Error())
		return
	}

	response.SuccessWithPagination(ctx, documents, total, page, pageSize)
}

// GetDetail 获取文档详情
// @Summary 获取文档详情
// @Tags 文件库
// @Produce json
// @Param id path int true "文档ID"
// @Success 200 {object} response.Response
// @Router /api/documents/{id} [get]
func (c *DocumentController) GetDetail(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "无效的文档ID")
		return
	}

	doc, metadata, err := c.queryService.GetDetail(uint(id))
	if err != nil {
		response.Error(ctx, http.StatusNotFound, "文档不存在")
		return
	}

	// 增加查看次数
	c.queryService.IncrementViewCount(uint(id))

	// 获取版本列表
	versions, _ := c.queryService.GetVersions(uint(id))

	result := map[string]interface{}{
		"document": doc,
		"metadata": metadata,
		"versions": versions,
	}

	response.Success(ctx, result)
}

// Delete 删除文档
// @Summary 删除文档
// @Tags 文件库
// @Produce json
// @Param id path int true "文档ID"
// @Success 200 {object} response.Response
// @Router /api/documents/{id} [delete]
func (c *DocumentController) Delete(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "无效的文档ID")
		return
	}

	if err := c.queryService.Delete(uint(id)); err != nil {
		response.Error(ctx, http.StatusInternalServerError, "删除失败: "+err.Error())
		return
	}

	response.SuccessWithMsg(ctx, "删除成功", nil)
}

// Update 更新文档信息
// @Summary 更新文档信息
// @Tags 文件库
// @Accept json
// @Produce json
// @Param id path int true "文档ID"
// @Param body body object true "更新内容"
// @Success 200 {object} response.Response
// @Router /api/documents/{id} [put]
func (c *DocumentController) Update(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "无效的文档ID")
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

// Download 下载文档
// @Summary 下载文档
// @Tags 文件库
// @Produce octet-stream
// @Param id path int true "文档ID"
// @Success 200 {file} binary
// @Router /api/documents/{id}/download [get]
func (c *DocumentController) Download(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "无效的文档ID")
		return
	}

	doc, _, err := c.queryService.GetDetail(uint(id))
	if err != nil {
		response.Error(ctx, http.StatusNotFound, "文档不存在")
		return
	}

	// 增加下载次数
	c.queryService.IncrementDownloadCount(uint(id))

	// 返回文件
	ctx.FileAttachment(doc.FilePath, doc.Name)
}

// GetStatistics 获取统计信息
// @Summary 获取统计信息
// @Tags 文件库
// @Produce json
// @Param type query string false "文档类型"
// @Param department query string false "部门"
// @Param project query string false "项目"
// @Success 200 {object} response.Response
// @Router /api/documents/statistics [get]
func (c *DocumentController) GetStatistics(ctx *gin.Context) {
	filters := document.QueryFilters{
		Type:       ctx.Query("type"),
		Department: ctx.Query("department"),
		Project:    ctx.Query("project"),
	}

	stats, err := c.queryService.GetStatistics(filters)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, "查询失败: "+err.Error())
		return
	}

	response.Success(ctx, stats)
}

// GetPopular 获取热门文档
// @Summary 获取热门文档
// @Tags 文件库
// @Produce json
// @Param limit query int false "数量限制" default(10)
// @Param type query string false "文档类型"
// @Success 200 {object} response.Response
// @Router /api/documents/popular [get]
func (c *DocumentController) GetPopular(ctx *gin.Context) {
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))

	filters := document.QueryFilters{
		Type:       ctx.Query("type"),
		Department: ctx.Query("department"),
		Project:    ctx.Query("project"),
	}

	documents, err := c.queryService.GetPopular(limit, filters)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, "查询失败: "+err.Error())
		return
	}

	response.Success(ctx, documents)
}

// GetAccessLogs 获取访问日志
// @Summary 获取访问日志
// @Tags 文件库
// @Produce json
// @Param id path int true "文档ID"
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(20)
// @Param action query string false "操作类型"
// @Success 200 {object} response.Response
// @Router /api/documents/{id}/logs [get]
func (c *DocumentController) GetAccessLogs(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "无效的文档ID")
		return
	}

	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("pageSize", "20"))
	action := ctx.Query("action")

	logs, total, err := c.queryService.GetAccessLogs(uint(id), page, pageSize, action)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, "查询失败: "+err.Error())
		return
	}

	response.SuccessWithPagination(ctx, logs, total, page, pageSize)
}
// RefreshThumbnail 刷新文档缩略图
// @Summary 刷新文档缩略图
// @Tags 文件库
// @Produce json
// @Param id path int true "文档ID"
// @Success 200 {object} response.Response
// @Router /api/documents/:id/refresh-thumbnail [post]
func (c *DocumentController) RefreshThumbnail(ctx *gin.Context) {
	// 获取文档ID
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "无效的文档ID")
		return
	}

	// 调用 uploadService 的 RegenerateThumbnail 方法
	if err := c.uploadService.RegenerateThumbnail(uint(id)); err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	response.SuccessWithMsg(ctx, "缩略图刷新任务已启动", nil)
}

// GetVersions 获取版本列表
// @Summary 获取版本列表
// @Tags 文件库
// @Produce json
// @Param id path int true "文档ID"
// @Success 200 {object} response.Response
// @Router /api/documents/{id}/versions [get]
func (c *DocumentController) GetVersions(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "无效的文档ID")
		return
	}

	versions, err := c.queryService.GetVersions(uint(id))
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, "查询失败: "+err.Error())
		return
	}

	response.Success(ctx, versions)
}

// RefreshFolderStats 刷新文件夹统计信息
// @Summary 刷新文件夹统计信息
// @Tags 文件库
// @Produce json
// @Param id path int true "文件夹ID"
// @Success 200 {object} response.Response
// @Router /api/documents/{id}/refresh-stats [post]
func (c *DocumentController) RefreshFolderStats(ctx *gin.Context) {
	// 获取文件夹ID
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "无效的文件夹ID")
		return
	}

	// 调用 queryService 的刷新统计方法
	if err := c.queryService.RefreshFolderStats(uint(id)); err != nil {
		response.Error(ctx, http.StatusInternalServerError, "刷新统计失败: "+err.Error())
		return
	}

	response.SuccessWithMsg(ctx, "文件夹统计已刷新", nil)
}

// RegisterDocumentRoutes 注册文档路由
func RegisterDocumentRoutes(router *gin.Engine, db *gorm.DB, fpService *fileprocessor.FileProcessorService, fpConfig *fileprocessor.Config) {
	controller := NewDocumentController(db, fpService, fpConfig)

	api := router.Group("/api/documents")
	{
		// 基础操作
		api.POST("/upload", controller.Upload)
		api.POST("/upload-folder", controller.UploadFolder)
		api.POST("/folder", controller.CreateFolder) // 创建文件夹
		api.GET("", controller.List)
		api.GET("/:id", controller.GetDetail)
		api.PUT("/:id", controller.Update)
		api.DELETE("/:id", controller.Delete)

		// 文件操作
		api.GET("/:id/download", controller.Download)
		api.POST("/:id/refresh-thumbnail", controller.RefreshThumbnail) // 刷新缩略图
		api.POST("/:id/refresh-stats", controller.RefreshFolderStats)   // 刷新文件夹统计

		// 版本管理
		api.GET("/:id/versions", controller.GetVersions)

		// 统计和日志
		api.GET("/statistics", controller.GetStatistics)
		api.GET("/popular", controller.GetPopular)
		api.GET("/:id/logs", controller.GetAccessLogs)
	}
}

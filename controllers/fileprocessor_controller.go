// Package controllers 文件处理器控制器
package controllers

import (
	"go_wails_project_manager/models"
	"go_wails_project_manager/response"
	"go_wails_project_manager/services/fileprocessor"
	"go_wails_project_manager/services/fileprocessor/processors"
	"go_wails_project_manager/services/task"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// FileProcessorController 文件处理器控制器
type FileProcessorController struct {
	fpService   *fileprocessor.FileProcessorService
	taskService *task.TaskService
	log         *logrus.Logger
}

// NewFileProcessorController 创建文件处理器控制器
func NewFileProcessorController(fpService *fileprocessor.FileProcessorService, taskService *task.TaskService, log *logrus.Logger) *FileProcessorController {
	return &FileProcessorController{
		fpService:   fpService,
		taskService: taskService,
		log:         log,
	}
}

// ExtractMetadata 提取文件元数据
func (c *FileProcessorController) ExtractMetadata(ctx *gin.Context) {
	var req struct {
		FilePath string `json:"file_path" binding:"required"`
		Format   string `json:"format"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	// 如果没有指定格式，从文件扩展名推断
	if req.Format == "" {
		ext := filepath.Ext(req.FilePath)
		req.Format = strings.TrimPrefix(ext, ".")
	}

	// 提取元数据
	metadata, err := c.fpService.ExtractMetadata(req.FilePath, req.Format)
	if err != nil {
		c.log.Errorf("提取元数据失败: %v", err)
		response.Error(ctx, http.StatusInternalServerError, "提取元数据失败: "+err.Error())
		return
	}

	response.Success(ctx, metadata)
}

// GenerateThumbnail 生成缩略图
func (c *FileProcessorController) GenerateThumbnail(ctx *gin.Context) {
	var req struct {
		FilePath   string `json:"file_path" binding:"required"`
		Format     string `json:"format"`
		OutputPath string `json:"output_path" binding:"required"`
		Size       int    `json:"size"`
		Quality    int    `json:"quality"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	// 默认值
	if req.Format == "" {
		ext := filepath.Ext(req.FilePath)
		req.Format = strings.TrimPrefix(ext, ".")
	}
	if req.Size == 0 {
		req.Size = 200
	}
	if req.Quality == 0 {
		req.Quality = 85
	}

	// 生成缩略图
	options := processors.ThumbnailOptions{
		Size:       req.Size,
		Quality:    req.Quality,
		OutputPath: req.OutputPath,
	}

	thumbnailPath, err := c.fpService.GenerateThumbnail(req.FilePath, req.Format, options)
	if err != nil {
		c.log.Errorf("生成缩略图失败: %v", err)
		response.Error(ctx, http.StatusInternalServerError, "生成缩略图失败: "+err.Error())
		return
	}

	response.Success(ctx, gin.H{
		"thumbnail_path": thumbnailPath,
	})
}

// CreateTask 创建处理任务
func (c *FileProcessorController) CreateTask(ctx *gin.Context) {
	var task models.Task

	if err := ctx.ShouldBindJSON(&task); err != nil {
		response.Error(ctx, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	// 创建任务
	if err := c.taskService.CreateTask(&task); err != nil {
		c.log.Errorf("创建任务失败: %v", err)
		response.Error(ctx, http.StatusInternalServerError, "创建任务失败: "+err.Error())
		return
	}

	response.Success(ctx, task)
}

// GetTask 获取任务详情
func (c *FileProcessorController) GetTask(ctx *gin.Context) {
	taskID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "无效的任务ID")
		return
	}

	task, err := c.taskService.GetTask(uint(taskID))
	if err != nil {
		c.log.Errorf("获取任务失败: %v", err)
		response.Error(ctx, http.StatusNotFound, "任务不存在")
		return
	}

	response.Success(ctx, task)
}

// ListTasks 列出任务
func (c *FileProcessorController) ListTasks(ctx *gin.Context) {
	filters := task.TaskFilters{
		Status:   ctx.Query("status"),
		Type:     ctx.Query("type"),
		Page:     1,
		PageSize: 20,
	}

	if page := ctx.Query("page"); page != "" {
		if p, err := strconv.Atoi(page); err == nil {
			filters.Page = p
		}
	}

	if pageSize := ctx.Query("page_size"); pageSize != "" {
		if ps, err := strconv.Atoi(pageSize); err == nil {
			filters.PageSize = ps
		}
	}

	tasks, total, err := c.taskService.ListTasks(filters)
	if err != nil {
		c.log.Errorf("列出任务失败: %v", err)
		response.Error(ctx, http.StatusInternalServerError, "列出任务失败: "+err.Error())
		return
	}

	response.Success(ctx, gin.H{
		"tasks": tasks,
		"total": total,
	})
}

// CancelTask 取消任务
func (c *FileProcessorController) CancelTask(ctx *gin.Context) {
	taskID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "无效的任务ID")
		return
	}

	if err := c.taskService.CancelTask(uint(taskID)); err != nil {
		c.log.Errorf("取消任务失败: %v", err)
		response.Error(ctx, http.StatusInternalServerError, "取消任务失败: "+err.Error())
		return
	}

	response.Success(ctx, gin.H{
		"message": "任务已取消",
	})
}

// RetryTask 重试任务
func (c *FileProcessorController) RetryTask(ctx *gin.Context) {
	taskID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "无效的任务ID")
		return
	}

	if err := c.taskService.RetryTask(uint(taskID)); err != nil {
		c.log.Errorf("重试任务失败: %v", err)
		response.Error(ctx, http.StatusInternalServerError, "重试任务失败: "+err.Error())
		return
	}

	response.Success(ctx, gin.H{
		"message": "任务已重新加入队列",
	})
}

// GetSupportedFormats 获取支持的文件格式
func (c *FileProcessorController) GetSupportedFormats(ctx *gin.Context) {
	formats := c.fpService.ListSupportedFormats()
	response.Success(ctx, formats)
}

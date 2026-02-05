package controllers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"go_wails_project_manager/config"
	"go_wails_project_manager/models/ai3d"
	ai3dService "go_wails_project_manager/services/ai3d"
	"go_wails_project_manager/response"
)

type AI3DUnifiedController struct {
	taskService *ai3dService.TaskService
}

func NewAI3DUnifiedController(taskService *ai3dService.TaskService) *AI3DUnifiedController {
	return &AI3DUnifiedController{
		taskService: taskService,
	}
}

// SubmitTask 提交任务
func (c *AI3DUnifiedController) SubmitTask(ctx *gin.Context) {
	var req struct {
		Provider         string                 `json:"provider" binding:"required"` // hunyuan | meshy
		InputType        string                 `json:"inputType" binding:"required"` // text | image
		Prompt           *string                `json:"prompt"`
		ImageURL         *string                `json:"imageUrl"`
		ImageBase64      *string                `json:"imageBase64"`
		GenerationParams map[string]interface{} `json:"generationParams"`
		Name             *string                `json:"name"`
		Description      *string                `json:"description"`
		Tags             *string                `json:"tags"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	// 验证输入
	if req.InputType == "text" && req.Prompt == nil {
		response.Error(ctx, http.StatusBadRequest, "文生3D需要提供prompt")
		return
	}
	if req.InputType == "image" && req.ImageURL == nil && req.ImageBase64 == nil {
		response.Error(ctx, http.StatusBadRequest, "图生3D需要提供imageUrl或imageBase64")
		return
	}

	// 获取用户信息
	username := ctx.GetString("username")
	if username == "" {
		username = "anonymous"
	}

	// 构建任务
	task := &ai3d.Task{
		Provider:         req.Provider,
		InputType:        req.InputType,
		Prompt:           req.Prompt,
		GenerationParams: req.GenerationParams,
		Category:         "AI生成",
		Tags:             req.Tags,
		CreatedBy:        username,
		CreatedIP:        ctx.ClientIP(),
	}

	// 处理图片输入：分离base64 data URI和普通URL
	if req.ImageURL != nil && *req.ImageURL != "" {
		imageURL := *req.ImageURL
		// 检查是否是base64 data URI格式
		if len(imageURL) > 100 && (imageURL[:22] == "data:image/png;base64," || 
			imageURL[:23] == "data:image/jpeg;base64," ||
			imageURL[:22] == "data:image/jpg;base64," ||
			imageURL[:22] == "data:image/webp;base64,") {
			// 是base64 data URI，提取base64部分保存到ImageBase64
			// 找到逗号位置
			commaIdx := 0
			for i, c := range imageURL {
				if c == ',' {
					commaIdx = i
					break
				}
			}
			if commaIdx > 0 && commaIdx < len(imageURL)-1 {
				base64Data := imageURL[commaIdx+1:]
				task.ImageBase64 = &base64Data
				// ImageURL设为空，避免保存大数据
				task.ImageURL = nil
			}
		} else {
			// 是普通URL，直接保存
			task.ImageURL = req.ImageURL
		}
	} else if req.ImageBase64 != nil {
		task.ImageBase64 = req.ImageBase64
	}

	// 设置任务名称
	if req.Name != nil && *req.Name != "" {
		task.Name = *req.Name
	} else {
		task.Name = fmt.Sprintf("任务_%s", time.Now().Format("20060102_150405"))
	}
	task.Description = req.Description

	// 创建任务
	if err := c.taskService.CreateTask(ctx, task); err != nil {
		response.Error(ctx, http.StatusInternalServerError, fmt.Sprintf("提交任务失败: %v", err))
		return
	}

	response.Success(ctx, gin.H{
		"task": task,
	})
}

// ListTasks 任务列表
func (c *AI3DUnifiedController) ListTasks(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("pageSize", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	filters := map[string]string{
		"provider": ctx.Query("provider"),
		"status":   ctx.Query("status"),
		"keyword":  ctx.Query("keyword"),
	}

	tasks, total, err := c.taskService.ListTasks(page, pageSize, filters)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, "查询失败")
		return
	}

	// 生成文件URL
	baseURL := fmt.Sprintf("http://%s:%d", config.AppConfig.LocalIP, config.AppConfig.ServerPort)
	for i := range tasks {
		// 动态生成fileUrl和thumbnailUrl
		if fileURL := tasks[i].GetFileURL(baseURL); fileURL != "" {
			tasks[i].ModelURL = &fileURL
		}
		if thumbURL := tasks[i].GetThumbnailURL(baseURL); thumbURL != "" {
			tasks[i].ThumbnailURL = &thumbURL
		}
		
		// 清理大字段：如果imageUrl包含base64数据，清空它以减少传输量
		if tasks[i].ImageURL != nil && len(*tasks[i].ImageURL) > 200 {
			// 如果是base64 data URI（通常很长），清空它
			tasks[i].ImageURL = nil
		}
	}

	response.Success(ctx, gin.H{
		"list":     tasks,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

// GetTask 获取任务详情
func (c *AI3DUnifiedController) GetTask(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "无效的ID")
		return
	}

	task, err := c.taskService.GetTask(uint(id))
	if err != nil {
		response.Error(ctx, http.StatusNotFound, "任务不存在")
		return
	}

	// 生成文件URL
	baseURL := fmt.Sprintf("http://%s:%d", config.AppConfig.LocalIP, config.AppConfig.ServerPort)
	if fileURL := task.GetFileURL(baseURL); fileURL != "" {
		task.ModelURL = &fileURL
	}
	if thumbURL := task.GetThumbnailURL(baseURL); thumbURL != "" {
		task.ThumbnailURL = &thumbURL
	}

	response.Success(ctx, task)
}

// PollTask 轮询任务
func (c *AI3DUnifiedController) PollTask(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "无效的ID")
		return
	}

	if err := c.taskService.PollTask(ctx, uint(id)); err != nil {
		response.Error(ctx, http.StatusInternalServerError, fmt.Sprintf("轮询失败: %v", err))
		return
	}

	task, _ := c.taskService.GetTask(uint(id))
	
	// 生成文件URL
	baseURL := fmt.Sprintf("http://%s:%d", config.AppConfig.LocalIP, config.AppConfig.ServerPort)
	if fileURL := task.GetFileURL(baseURL); fileURL != "" {
		task.ModelURL = &fileURL
	}
	if thumbURL := task.GetThumbnailURL(baseURL); thumbURL != "" {
		task.ThumbnailURL = &thumbURL
	}
	
	response.Success(ctx, task)
}

// DeleteTask 删除任务
func (c *AI3DUnifiedController) DeleteTask(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "无效的ID")
		return
	}

	if err := c.taskService.DeleteTask(uint(id)); err != nil {
		response.Error(ctx, http.StatusInternalServerError, "删除失败")
		return
	}

	response.Success(ctx, nil)
}

// GetConfig 获取配置
func (c *AI3DUnifiedController) GetConfig(ctx *gin.Context) {
	response.Success(ctx, gin.H{
		"hunyuan": gin.H{
			"enabled":             config.AppConfig.Hunyuan.SecretID != "",
			"defaultModel":        config.AppConfig.Hunyuan.DefaultModel,
			"defaultGenerateType": config.AppConfig.Hunyuan.DefaultGenerateType,
			"defaultFaceCount":    config.AppConfig.Hunyuan.DefaultFaceCount,
			"defaultEnablePBR":    config.AppConfig.Hunyuan.DefaultEnablePBR,
			"defaultResultFormat": config.AppConfig.Hunyuan.DefaultResultFormat,
		},
		"meshy": gin.H{
			"enabled":                config.AppConfig.Meshy.APIKey != "",
			"defaultAIModel":         config.AppConfig.Meshy.DefaultAIModel,
			"defaultEnablePBR":       config.AppConfig.Meshy.DefaultEnablePBR,
			"defaultTopology":        config.AppConfig.Meshy.DefaultTopology,
			"defaultTargetPolycount": config.AppConfig.Meshy.DefaultTargetPolycount,
			"defaultShouldRemesh":    config.AppConfig.Meshy.DefaultShouldRemesh,
			"defaultShouldTexture":   config.AppConfig.Meshy.DefaultShouldTexture,
			"defaultSavePreRemeshed": config.AppConfig.Meshy.DefaultSavePreRemeshed,
		},
	})
}

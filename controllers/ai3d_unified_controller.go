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
	// 先解析为通用map，以便处理平铺的参数
	var reqMap map[string]interface{}
	if err := ctx.ShouldBindJSON(&reqMap); err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	// 提取必需字段
	provider, _ := reqMap["provider"].(string)
	inputType, _ := reqMap["inputType"].(string)
	
	if provider == "" || inputType == "" {
		response.Error(ctx, http.StatusBadRequest, "provider和inputType为必填项")
		return
	}

	// 验证输入
	if inputType == "text" {
		if prompt, ok := reqMap["prompt"].(string); !ok || prompt == "" {
			response.Error(ctx, http.StatusBadRequest, "文生3D需要提供prompt")
			return
		}
	}
	if inputType == "image" {
		hasImage := false
		if imageUrl, ok := reqMap["imageUrl"].(string); ok && imageUrl != "" {
			hasImage = true
		}
		if imageBase64, ok := reqMap["imageBase64"].(string); ok && imageBase64 != "" {
			hasImage = true
		}
		if !hasImage {
			response.Error(ctx, http.StatusBadRequest, "图生3D需要提供imageUrl或imageBase64")
			return
		}
	}

	// 获取用户信息
	username := ctx.GetString("username")
	if username == "" {
		username = "anonymous"
	}

	// 构建GenerationParams：收集所有非标准字段
	standardFields := map[string]bool{
		"provider": true, "inputType": true, "prompt": true,
		"imageUrl": true, "imageBase64": true, "name": true,
		"description": true, "tags": true, "generationParams": true,
	}
	
	generationParams := make(map[string]interface{})
	
	// 如果有generationParams字段，先合并它
	if gp, ok := reqMap["generationParams"].(map[string]interface{}); ok {
		for k, v := range gp {
			generationParams[k] = v
		}
	}
	
	// 收集平铺的参数（如enablePbr, model, faceCount等）
	for key, value := range reqMap {
		if !standardFields[key] {
			generationParams[key] = value
		}
	}

	// 处理图片输入：提取base64数据，临时存储在GenerationParams中
	if imageUrl, ok := reqMap["imageUrl"].(string); ok && imageUrl != "" {
		// 检查是否是base64 data URI格式
		if len(imageUrl) > 100 && (imageUrl[:22] == "data:image/png;base64," || 
			imageUrl[:23] == "data:image/jpeg;base64," ||
			imageUrl[:22] == "data:image/jpg;base64," ||
			imageUrl[:23] == "data:image/webp;base64,") {
			// 提取base64部分
			commaIdx := 0
			for i, c := range imageUrl {
				if c == ',' {
					commaIdx = i
					break
				}
			}
			if commaIdx > 0 && commaIdx < len(imageUrl)-1 {
				base64Data := imageUrl[commaIdx+1:]
				generationParams["_imageBase64"] = base64Data
			}
		} else {
			// 普通URL
			generationParams["_imageUrl"] = imageUrl
		}
	} else if imageBase64, ok := reqMap["imageBase64"].(string); ok && imageBase64 != "" {
		generationParams["_imageBase64"] = imageBase64
	}

	// 提取prompt
	var prompt *string
	if p, ok := reqMap["prompt"].(string); ok && p != "" {
		prompt = &p
	}

	// 构建任务
	task := &ai3d.Task{
		Provider:         provider,
		InputType:        inputType,
		Prompt:           prompt,
		GenerationParams: generationParams,
		Category:         "AI生成",
		CreatedBy:        username,
		CreatedIP:        ctx.ClientIP(),
	}

	// 设置任务名称
	if name, ok := reqMap["name"].(string); ok && name != "" {
		task.Name = name
	} else {
		task.Name = fmt.Sprintf("任务_%s", time.Now().Format("20060102_150405"))
	}
	
	// 设置描述
	if desc, ok := reqMap["description"].(string); ok && desc != "" {
		task.Description = &desc
	}
	
	// 设置标签
	if tags, ok := reqMap["tags"].(string); ok && tags != "" {
		task.Tags = &tags
	}

	// 创建任务（适配器会更新GenerationParams为实际使用的参数）
	if err := c.taskService.CreateTask(ctx, task); err != nil {
		response.Error(ctx, http.StatusInternalServerError, fmt.Sprintf("提交任务失败: %v", err))
		return
	}

	response.Success(ctx, gin.H{
		"task": task, // 返回的task.GenerationParams已包含实际使用的参数
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

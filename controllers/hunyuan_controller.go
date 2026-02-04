package controllers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"go_wails_project_manager/config"
	"go_wails_project_manager/models/hunyuan"
	"go_wails_project_manager/response"
	hunyuanService "go_wails_project_manager/services/hunyuan"
)

// HunyuanController 混元3D控制器
type HunyuanController struct {
	taskService    *hunyuanService.TaskService
	configService  *hunyuanService.ConfigService
	storageService *hunyuanService.StorageService
	poller         *hunyuanService.TaskPoller
}

// NewHunyuanController 创建控制器
func NewHunyuanController(db *gorm.DB) *HunyuanController {
	configService := hunyuanService.NewConfigService(db)
	storageService := hunyuanService.NewStorageService(db, &config.AppConfig.Hunyuan)
	taskService := hunyuanService.NewTaskService(db, configService, storageService)

	// 创建并启动轮询器
	pollInterval := time.Duration(config.AppConfig.Hunyuan.PollInterval) * time.Second
	poller := hunyuanService.NewTaskPoller(db, taskService, pollInterval)
	poller.Start()

	return &HunyuanController{
		taskService:    taskService,
		configService:  configService,
		storageService: storageService,
		poller:         poller,
	}
}

// SubmitTaskRequest 提交任务请求
type SubmitTaskRequest struct {
	InputType   string                          `json:"inputType" binding:"required"`
	Prompt      *string                         `json:"prompt"`
	ImageURL    *string                         `json:"imageUrl"`
	ImageBase64 *string                         `json:"imageBase64"`
	Model       string                          `json:"model"`
	FaceCount   *int                            `json:"faceCount"`
	EnablePBR   *bool                           `json:"enablePbr"`
	Name        *string                         `json:"name"`
	Description *string                         `json:"description"`
	Tags        []string                        `json:"tags"`
}

// SubmitTask 提交任务
func (c *HunyuanController) SubmitTask(ctx *gin.Context) {
	var req SubmitTaskRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	// 验证输入
	if req.InputType == "text" && (req.Prompt == nil || *req.Prompt == "") {
		response.Error(ctx, http.StatusBadRequest, "文生3D需要提供prompt")
		return
	}
	if req.InputType == "image" && req.ImageURL == nil && req.ImageBase64 == nil {
		response.Error(ctx, http.StatusBadRequest, "图生3D需要提供imageUrl或imageBase64")
	}

	// 获取配置
	cfg, err := c.configService.GetConfig()
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, "获取配置失败")
		return
	}

	// 构建参数
	params := &hunyuanService.GenerateParams{
		Model:        req.Model,
		Prompt:       req.Prompt,
		ImageURL:     req.ImageURL,
		ImageBase64:  req.ImageBase64,
		FaceCount:    req.FaceCount,
		GenerateType: cfg.DefaultGenerateType,
	}

	// 设置默认值
	if params.Model == "" {
		params.Model = cfg.DefaultModel
	}
	if params.FaceCount == nil {
		faceCount := cfg.DefaultFaceCount
		params.FaceCount = &faceCount
	}
	if req.EnablePBR == nil {
		enablePBR := cfg.DefaultEnablePBR
		params.EnablePBR = &enablePBR
	} else {
		params.EnablePBR = req.EnablePBR
	}

	resultFormat := cfg.DefaultResultFormat
	params.ResultFormat = &resultFormat

	// 获取用户信息
	username := ctx.GetString("username")
	if username == "" {
		username = "anonymous"
	}

	userInfo := hunyuanService.UserInfo{
		Username: username,
		IP:       ctx.ClientIP(),
	}

	// 创建任务
	task, err := c.taskService.CreateTask(params, userInfo)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	// 更新名称和描述
	if req.Name != nil {
		task.Name = *req.Name
	}
	if req.Description != nil {
		task.Description = req.Description
	}

	response.Success(ctx, task)
}

// GetTask 获取任务详情
func (c *HunyuanController) GetTask(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "无效的ID")
		return
	}

	task, err := c.taskService.GetTaskStatus(uint(id))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Error(ctx, http.StatusNotFound, "任务不存在")
			return
		}
		response.Error(ctx, http.StatusInternalServerError, "查询失败")
		return
	}

	response.Success(ctx, task)
}

// ListTasks 任务列表
func (c *HunyuanController) ListTasks(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("pageSize", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	filters := hunyuanService.TaskFilters{
		Status:    ctx.Query("status"),
		InputType: ctx.Query("inputType"),
		Keyword:   ctx.Query("keyword"),
	}

	tasks, total, err := c.taskService.ListTasks(page, pageSize, filters)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, "查询失败")
		return
	}

	response.Success(ctx, gin.H{
		"list":     tasks,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

// PollTask 轮询任务
func (c *HunyuanController) PollTask(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "无效的ID")
		return
	}

	if err := c.taskService.PollTask(uint(id)); err != nil {
		response.Error(ctx, http.StatusInternalServerError, "轮询失败: "+err.Error())
		return
	}

	// 返回更新后的任务
	task, _ := c.taskService.GetTaskStatus(uint(id))
	response.Success(ctx, task)
}

// CancelTask 取消任务
func (c *HunyuanController) CancelTask(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "无效的ID")
		return
	}

	if err := c.taskService.CancelTask(uint(id)); err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(ctx, gin.H{"message": "任务已取消"})
}

// RetryTask 重试任务
func (c *HunyuanController) RetryTask(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "无效的ID")
		return
	}

	if err := c.taskService.RetryTask(uint(id)); err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	// 返回更新后的任务
	task, _ := c.taskService.GetTaskStatus(uint(id))
	response.Success(ctx, task)
}

// DeleteTask 删除任务
func (c *HunyuanController) DeleteTask(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "无效的ID")
		return
	}

	if err := c.taskService.DeleteTask(uint(id)); err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Error(ctx, http.StatusNotFound, "任务不存在")
			return
		}
		response.Error(ctx, http.StatusInternalServerError, "删除失败")
		return
	}

	response.Success(ctx, gin.H{"message": "删除成功"})
}

// GetStatistics 获取统计信息
func (c *HunyuanController) GetStatistics(ctx *gin.Context) {
	stats, err := c.taskService.GetStatistics()
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, "查询失败")
		return
	}

	response.Success(ctx, stats)
}

// GetConfig 获取配置（从 config.yaml 读取）
func (c *HunyuanController) GetConfig(ctx *gin.Context) {
	cfg, err := c.configService.GetConfig()
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	// 脱敏处理
	if cfg.SecretKey != "" {
		cfg.SecretKey = "******"
	}

	response.Success(ctx, cfg)
}

// UpdateConfig 更新配置（已禁用，请修改 config.yaml）
func (c *HunyuanController) UpdateConfig(ctx *gin.Context) {
	response.Error(ctx, http.StatusForbidden, "配置更新功能已禁用，请直接修改 config.yaml 文件并重启服务")
}

// ValidateConfig 验证配置
func (c *HunyuanController) ValidateConfig(ctx *gin.Context) {
	var cfg hunyuan.HunyuanConfig
	if err := ctx.ShouldBindJSON(&cfg); err != nil {
		response.Error(ctx, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	if err := c.configService.ValidateConfig(&cfg); err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(ctx, gin.H{
		"valid":   true,
		"message": "配置有效",
	})
}

// GetPollerStatus 获取轮询器状态
func (c *HunyuanController) GetPollerStatus(ctx *gin.Context) {
	status := c.poller.GetStatus()
	response.Success(ctx, status)
}

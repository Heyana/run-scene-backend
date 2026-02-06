package adapters

import (
	"context"
	"fmt"

	"go_wails_project_manager/config"
	"go_wails_project_manager/models/ai3d"
	ai3dService "go_wails_project_manager/services/ai3d"
	hunyuanService "go_wails_project_manager/services/hunyuan"

	"gorm.io/gorm"
)

type HunyuanAdapter struct {
	db      *gorm.DB
	config  *config.HunyuanConfig
	client  *hunyuanService.HunyuanClient
	storage *hunyuanService.StorageService
}

func NewHunyuanAdapter(db *gorm.DB, cfg *config.HunyuanConfig) *HunyuanAdapter {
	client := hunyuanService.NewHunyuanClient(cfg.SecretID, cfg.SecretKey, cfg.Region)
	storage := hunyuanService.NewStorageService(db, cfg)

	return &HunyuanAdapter{
		db:      db,
		config:  cfg,
		client:  client,
		storage: storage,
	}
}

func (a *HunyuanAdapter) GetName() string {
	return "hunyuan"
}

func (a *HunyuanAdapter) SubmitTask(ctx context.Context, task *ai3d.Task) (string, error) {
	// 从 GenerationParams 提取混元参数（使用默认值填充）
	model := getStringParam(task.GenerationParams, "model", a.config.DefaultModel)
	generateType := getStringParam(task.GenerationParams, "generateType", a.config.DefaultGenerateType)
	faceCount := getIntParam(task.GenerationParams, "faceCount", a.config.DefaultFaceCount)
	enablePBR := getBoolParam(task.GenerationParams, "enablePbr", a.config.DefaultEnablePBR)
	resultFormat := getStringParam(task.GenerationParams, "resultFormat", a.config.DefaultResultFormat)

	// 构建API参数
	params := &hunyuanService.GenerateParams{
		Model:        model,
		GenerateType: generateType,
		FaceCount:    &faceCount,
		EnablePBR:    &enablePBR,
		ResultFormat: &resultFormat,
	}

	// 设置输入（从临时字段获取图片数据）
	if task.InputType == "text" && task.Prompt != nil {
		params.Prompt = task.Prompt
	} else if task.InputType == "image" {
		// 从GenerationParams临时获取图片数据
		if imageURL, ok := task.GenerationParams["_imageUrl"].(string); ok && imageURL != "" {
			params.ImageURL = &imageURL
		} else if imageBase64, ok := task.GenerationParams["_imageBase64"].(string); ok && imageBase64 != "" {
			params.ImageBase64 = &imageBase64
		}
	}

	// 保存实际使用的参数到GenerationParams（不包括图片数据）
	task.GenerationParams = ai3d.GenerationParams{
		"model":        model,
		"generateType": generateType,
		"faceCount":    faceCount,
		"enablePbr":    enablePBR,
		"resultFormat": resultFormat,
	}

	// 调用混元API
	jobID, err := a.client.SubmitJob(params)
	if err != nil {
		return "", fmt.Errorf("提交混元任务失败: %w", err)
	}

	return jobID, nil
}

func (a *HunyuanAdapter) QueryTask(ctx context.Context, providerTaskID string) (*ai3dService.TaskStatus, error) {
	// 调用混元API查询
	resp, err := a.client.QueryJob(providerTaskID)
	if err != nil {
		return nil, fmt.Errorf("查询混元任务失败: %w", err)
	}

	status := &ai3dService.TaskStatus{
		Status:   resp.Response.Status, // 混元已经使用统一状态
		Progress: 100,                  // 混元没有进度，默认100
	}

	// 从ResultFiles提取模型和缩略图URL
	if len(resp.Response.ResultFiles) > 0 {
		for _, file := range resp.Response.ResultFiles {
			if file.Type == "glb" || file.Type == "fbx" {
				status.ModelURL = file.URL
			}
			if file.PreviewImageURL != "" {
				status.ThumbnailURL = file.PreviewImageURL
			}
		}
	}

	if resp.Response.ErrorCode != "" {
		status.ErrorCode = resp.Response.ErrorCode
	}
	if resp.Response.ErrorMessage != "" {
		status.ErrorMessage = resp.Response.ErrorMessage
	}

	return status, nil
}

func (a *HunyuanAdapter) DownloadResult(ctx context.Context, task *ai3d.Task) (*ai3dService.DownloadResult, error) {
	if task.ModelURL == nil || *task.ModelURL == "" {
		return nil, fmt.Errorf("模型URL为空")
	}

	// 简化实现：直接返回URL，不下载
	// 文件已经在轮询时由旧系统下载了
	result := &ai3dService.DownloadResult{
		FileSize: 0,
		FileHash: "",
	}

	// 如果任务已有路径信息，直接使用
	if task.LocalPath != nil {
		result.LocalPath = *task.LocalPath
	}
	if task.NASPath != nil {
		result.NASPath = *task.NASPath
	}
	if task.ThumbnailPath != nil {
		result.ThumbnailPath = *task.ThumbnailPath
	}
	if task.FileSize != nil {
		result.FileSize = *task.FileSize
	}
	if task.FileHash != nil {
		result.FileHash = *task.FileHash
	}

	return result, nil
}

func (a *HunyuanAdapter) CancelTask(ctx context.Context, providerTaskID string) error {
	// 混元不支持取消任务
	return fmt.Errorf("混元不支持取消任务")
}

// 辅助函数
func getStringParam(params ai3d.GenerationParams, key, defaultValue string) string {
	if val, ok := params[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return defaultValue
}

func getIntParam(params ai3d.GenerationParams, key string, defaultValue int) int {
	if val, ok := params[key]; ok {
		switch v := val.(type) {
		case int:
			return v
		case float64:
			return int(v)
		}
	}
	return defaultValue
}

func getBoolParam(params ai3d.GenerationParams, key string, defaultValue bool) bool {
	if val, ok := params[key]; ok {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return defaultValue
}

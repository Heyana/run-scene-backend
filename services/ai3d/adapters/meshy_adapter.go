package adapters

import (
	"context"
	"fmt"

	"go_wails_project_manager/config"
	"go_wails_project_manager/models/ai3d"
	"go_wails_project_manager/models/meshy"
	ai3dService "go_wails_project_manager/services/ai3d"
	meshyService "go_wails_project_manager/services/meshy"

	"gorm.io/gorm"
)

type MeshyAdapter struct {
	db      *gorm.DB
	config  *config.MeshyConfig
	client  *meshyService.Client
	storage *meshyService.StorageService
}

func NewMeshyAdapter(db *gorm.DB, cfg *config.MeshyConfig) *MeshyAdapter {
	client := meshyService.NewClient(cfg.APIKey, cfg.BaseURL)
	storage := meshyService.NewStorageService(db, cfg)

	return &MeshyAdapter{
		db:      db,
		config:  cfg,
		client:  client,
		storage: storage,
	}
}

func (a *MeshyAdapter) GetName() string {
	return "meshy"
}

func (a *MeshyAdapter) SubmitTask(ctx context.Context, task *ai3d.Task) (string, error) {
	// 从 GenerationParams 提取Meshy参数，并填充默认值
	aiModel := getStringParam(task.GenerationParams, "aiModel", a.config.DefaultAIModel)
	enablePBR := getBoolParam(task.GenerationParams, "enablePbr", a.config.DefaultEnablePBR)
	topology := getStringParam(task.GenerationParams, "topology", a.config.DefaultTopology)
	targetPolycount := getIntParam(task.GenerationParams, "targetPolycount", a.config.DefaultTargetPolycount)
	shouldRemesh := getBoolParam(task.GenerationParams, "shouldRemesh", a.config.DefaultShouldRemesh)
	shouldTexture := getBoolParam(task.GenerationParams, "shouldTexture", a.config.DefaultShouldTexture)
	savePreRemeshed := getBoolParam(task.GenerationParams, "savePreRemeshed", a.config.DefaultSavePreRemeshed)

	// 构建图片URL（从临时字段获取）
	imageURL := ""
	if url, ok := task.GenerationParams["_imageUrl"].(string); ok && url != "" {
		imageURL = url
	} else if base64, ok := task.GenerationParams["_imageBase64"].(string); ok && base64 != "" {
		// 从ImageBase64重新构建Data URI格式
		imageURL = "data:image/png;base64," + base64
	}

	if imageURL == "" {
		return "", fmt.Errorf("Meshy需要提供图片")
	}

	// 构建API请求参数
	params := &meshyService.ImageTo3DRequest{
		ImageURL:             imageURL,
		AIModel:              aiModel,
		EnablePBR:            enablePBR,
		Topology:             topology,
		TargetPolycount:      targetPolycount,
		ShouldRemesh:         shouldRemesh,
		ShouldTexture:        shouldTexture,
		SavePreRemeshedModel: savePreRemeshed,
	}

	// 保存实际使用的参数到GenerationParams（不包括图片数据）
	task.GenerationParams = ai3d.GenerationParams{
		"aiModel":         aiModel,
		"enablePbr":       enablePBR,
		"topology":        topology,
		"targetPolycount": targetPolycount,
		"shouldRemesh":    shouldRemesh,
		"shouldTexture":   shouldTexture,
		"savePreRemeshed": savePreRemeshed,
	}

	fmt.Printf("Meshy适配器更新GenerationParams: %+v\n", task.GenerationParams)

	// 调用Meshy API
	taskID, err := a.client.SubmitImageTo3D(params)
	if err != nil {
		return "", fmt.Errorf("提交Meshy任务失败: %w", err)
	}

	return taskID, nil
}

func (a *MeshyAdapter) QueryTask(ctx context.Context, providerTaskID string) (*ai3dService.TaskStatus, error) {
	// 调用Meshy API查询
	resp, err := a.client.GetTask(providerTaskID)
	if err != nil {
		return nil, fmt.Errorf("查询Meshy任务失败: %w", err)
	}

	// 转换Meshy状态为统一状态
	status := &ai3dService.TaskStatus{
		Status:   convertMeshyStatus(resp.Status),
		Progress: resp.Progress,
	}

	// 提取模型URL（优先使用优化后的模型）
	if resp.ModelURLs != nil {
		if resp.ModelURLs.GLB != "" {
			status.ModelURL = resp.ModelURLs.GLB
		} else if resp.ModelURLs.FBX != "" {
			status.ModelURL = resp.ModelURLs.FBX
		} else if resp.ModelURLs.OBJ != "" {
			status.ModelURL = resp.ModelURLs.OBJ
		}
		
		// 保存PreRemeshed模型URL（如果存在）
		if resp.ModelURLs.PreRemeshedGLB != "" {
			status.PreRemeshedURL = resp.ModelURLs.PreRemeshedGLB
		}
	}

	if resp.ThumbnailURL != "" {
		status.ThumbnailURL = resp.ThumbnailURL
	}

	if resp.TaskError != nil && resp.TaskError.Message != "" {
		status.ErrorMessage = resp.TaskError.Message
	}

	return status, nil
}

func (a *MeshyAdapter) DownloadResult(ctx context.Context, task *ai3d.Task) (*ai3dService.DownloadResult, error) {
	if task.ModelURL == nil || *task.ModelURL == "" {
		return nil, fmt.Errorf("模型URL为空")
	}

	thumbnailURL := ""
	if task.ThumbnailURL != nil {
		thumbnailURL = *task.ThumbnailURL
	}
	
	preRemeshedURL := ""
	if task.PreRemeshedURL != nil {
		preRemeshedURL = *task.PreRemeshedURL
	}

	// 创建临时任务结构用于下载
	tempTask := &meshy.MeshyTask{
		TaskID: task.ProviderTaskID,
	}

	// 调用存储服务下载（包括PreRemeshed模型）
	info, err := a.storage.SaveTaskResult(tempTask, *task.ModelURL, thumbnailURL, preRemeshedURL)
	if err != nil {
		return nil, fmt.Errorf("下载Meshy文件失败: %w", err)
	}

	result := &ai3dService.DownloadResult{
		FileSize: info.FileSize,
		FileHash: info.FileHash,
	}

	if info.LocalPath != "" {
		result.LocalPath = info.LocalPath
	}
	if info.NASPath != "" {
		result.NASPath = info.NASPath
	}
	if info.ThumbnailPath != "" {
		result.ThumbnailPath = info.ThumbnailPath
	}
	if info.PreRemeshedPath != "" {
		result.PreRemeshedPath = info.PreRemeshedPath
	}
	if info.PreRemeshedNASPath != "" {
		result.PreRemeshedNASPath = info.PreRemeshedNASPath
	}

	return result, nil
}

func (a *MeshyAdapter) CancelTask(ctx context.Context, providerTaskID string) error {
	// Meshy不支持取消任务
	return nil
}

// convertMeshyStatus 转换Meshy状态为统一状态
func convertMeshyStatus(meshyStatus string) string {
	switch meshyStatus {
	case "PENDING":
		return "WAIT"
	case "IN_PROGRESS":
		return "RUN"
	case "SUCCEEDED":
		return "DONE"
	case "FAILED", "CANCELED":
		return "FAIL"
	default:
		return "WAIT"
	}
}

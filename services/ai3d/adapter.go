package ai3d

import (
	"context"
	"go_wails_project_manager/models/ai3d"
)

// ProviderAdapter 平台适配器接口
type ProviderAdapter interface {
	// GetName 获取平台名称
	GetName() string

	// SubmitTask 提交任务到平台
	SubmitTask(ctx context.Context, task *ai3d.Task) (providerTaskID string, err error)

	// QueryTask 查询平台任务状态
	QueryTask(ctx context.Context, providerTaskID string) (*TaskStatus, error)

	// DownloadResult 下载任务结果
	DownloadResult(ctx context.Context, task *ai3d.Task) (*DownloadResult, error)

	// CancelTask 取消任务（可选，不支持的平台返回nil）
	CancelTask(ctx context.Context, providerTaskID string) error
}

// TaskStatus 任务状态
type TaskStatus struct {
	Status         string // WAIT | RUN | DONE | FAIL
	Progress       int    // 0-100
	ModelURL       string
	PreRemeshedURL string // PreRemeshed模型URL（Meshy专用，高精度原始模型）
	ThumbnailURL   string
	ErrorCode      string
	ErrorMessage   string
}

// DownloadResult 下载结果
type DownloadResult struct {
	LocalPath          string
	NASPath            string
	ThumbnailPath      string
	PreRemeshedPath    string // PreRemeshed模型本地路径（如果有）
	PreRemeshedNASPath string // PreRemeshed模型NAS路径（如果有）
	FileSize           int64
	FileHash           string
}

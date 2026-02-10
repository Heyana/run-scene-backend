// Package models 任务数据模型
package models

import (
	"time"

	"gorm.io/gorm"
)

// Task 任务模型
type Task struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// 任务信息
	Type     string `json:"type"`     // 任务类型
	Status   string `json:"status"`   // 状态
	Priority int    `json:"priority"` // 优先级 0-10

	// 输入输出
	InputFile  string `json:"input_file"`
	OutputFile string `json:"output_file"`
	Options    string `json:"options"` // JSON

	// 进度信息
	Progress float64 `json:"progress"` // 0-100
	Message  string  `json:"message"`
	Error    string  `json:"error"`

	// 执行信息
	StartedAt   *time.Time `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at"`
	Duration    int64      `json:"duration"` // 秒

	// 重试机制
	RetryCount int    `json:"retry_count"`
	MaxRetries int    `json:"max_retries"`
	RetryDelay int    `json:"retry_delay"` // 秒
	LastError  string `json:"last_error"`

	// 断点续传
	Checkpoint string `json:"checkpoint"` // JSON
	Resumable  bool   `json:"resumable"`

	// 依赖关系
	ParentTaskID *uint  `json:"parent_task_id"`
	DependsOn    string `json:"depends_on"` // JSON数组

	// 通知
	NotifyEmail   string `json:"notify_email"`
	NotifyWebhook string `json:"notify_webhook"`
	CallbackURL   string `json:"callback_url"`

	// 关联
	UserID     uint `json:"user_id"`
	DocumentID uint `json:"document_id"`
}

// 任务状态常量
const (
	TaskStatusPending   = "pending"    // 等待中
	TaskStatusRunning   = "running"    // 运行中
	TaskStatusCompleted = "completed"  // 已完成
	TaskStatusFailed    = "failed"     // 失败
	TaskStatusCancelled = "cancelled"  // 已取消
	TaskStatusRetrying  = "retrying"   // 重试中
)

// 任务类型常量
const (
	TaskTypeVideoPreview    = "video_preview"
	TaskTypeVideoThumbnail  = "video_thumbnail"
	TaskTypeVideoConvert    = "video_convert"
	TaskTypeImageThumbnail  = "image_thumbnail"
	TaskTypeImageConvert    = "image_convert"
	TaskTypeDocumentPreview = "document_preview"
	TaskTypeModelPreview    = "model_preview"
)

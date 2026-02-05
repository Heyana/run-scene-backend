package meshy

import (
	"time"

	"gorm.io/gorm"
)

// MeshyTask Meshy任务模型
type MeshyTask struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
	
	// 任务信息
	TaskID      string `gorm:"uniqueIndex;not null" json:"taskId"`
	Status      string `gorm:"index;not null" json:"status"` // PENDING, IN_PROGRESS, SUCCEEDED, FAILED
	Progress    int    `json:"progress"`
	
	// 输入参数
	ImageURL           string `json:"imageUrl"`
	EnablePBR          bool   `json:"enablePbr"`
	ShouldRemesh       bool   `json:"shouldRemesh"`
	ShouldTexture      bool   `json:"shouldTexture"`
	SavePreRemeshed    bool   `json:"savePreRemeshed"`
	
	// 结果
	ModelURL      string `json:"modelUrl,omitempty"`
	ThumbnailURL  string `json:"thumbnailUrl,omitempty"`
	LocalPath     string `json:"localPath,omitempty"`
	NASPath       string `json:"nasPath,omitempty"`
	ThumbnailPath string `json:"thumbnailPath,omitempty"`
	FileSize      int64  `json:"fileSize,omitempty"`
	FileHash      string `json:"fileHash,omitempty"`
	
	// 错误信息
	ErrorMessage string `json:"errorMessage,omitempty"`
	
	// 元数据
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	Category    string  `gorm:"index" json:"category"`
	Tags        string  `json:"tags,omitempty"` // JSON数组字符串
	
	// 用户信息
	CreatedBy string `gorm:"index" json:"createdBy"`
	CreatedIP string `json:"createdIp"`
}

// TableName 指定表名
func (MeshyTask) TableName() string {
	return "meshy_tasks"
}

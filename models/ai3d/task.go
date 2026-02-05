package ai3d

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// Task 统一AI3D任务模型
type Task struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// 平台信息
	Provider       string `gorm:"index;not null" json:"provider"`             // hunyuan | meshy
	ProviderTaskID string `gorm:"not null" json:"providerTaskId"`             // 平台的任务ID
	
	// 任务状态（统一状态）
	Status   string `gorm:"index;not null" json:"status"` // WAIT | RUN | DONE | FAIL
	Progress int    `json:"progress"`                     // 0-100

	// 输入参数
	InputType   string  `gorm:"not null" json:"inputType"`      // text | image | multi_view
	Prompt      *string `json:"prompt,omitempty"`               // 文本提示词
	ImageURL    *string `json:"imageUrl,omitempty"`             // 图片URL
	ImageBase64 *string `gorm:"type:text" json:"-"`             // Base64图片（不返回给前端）

	// 生成参数（JSON存储平台特定参数）
	GenerationParams GenerationParams `gorm:"type:text" json:"generationParams"`

	// 结果文件
	ModelURL      *string `json:"modelUrl,omitempty"`      // 平台返回的模型URL
	ThumbnailURL  *string `json:"thumbnailUrl,omitempty"`  // 平台返回的缩略图URL
	LocalPath     *string `json:"localPath,omitempty"`     // 本地存储路径
	NASPath       *string `json:"nasPath,omitempty"`       // NAS存储路径
	ThumbnailPath *string `json:"thumbnailPath,omitempty"` // 缩略图路径
	FileSize      *int64  `json:"fileSize,omitempty"`      // 文件大小（字节）
	FileHash      *string `json:"fileHash,omitempty"`      // MD5哈希

	// 错误信息
	ErrorCode    *string `json:"errorCode,omitempty"`
	ErrorMessage *string `json:"errorMessage,omitempty"`

	// 元数据
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	Category    string  `gorm:"index" json:"category"`
	Tags        *string `json:"tags,omitempty"` // JSON数组字符串

	// 用户信息
	CreatedBy string `gorm:"index" json:"createdBy"`
	CreatedIP string `json:"createdIp"`
}

// GenerationParams 生成参数（平台特定）
type GenerationParams map[string]interface{}

// Scan 实现 sql.Scanner 接口
func (g *GenerationParams) Scan(value interface{}) error {
	if value == nil {
		*g = make(GenerationParams)
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		str, ok := value.(string)
		if !ok {
			return fmt.Errorf("failed to unmarshal JSON value: invalid type")
		}
		bytes = []byte(str)
	}
	if len(bytes) == 0 {
		*g = make(GenerationParams)
		return nil
	}
	return json.Unmarshal(bytes, g)
}

// Value 实现 driver.Valuer 接口
func (g GenerationParams) Value() (driver.Value, error) {
	if g == nil || len(g) == 0 {
		return "{}", nil
	}
	return json.Marshal(g)
}

// TableName 指定表名
func (Task) TableName() string {
	return "ai3d_tasks"
}

// GetFileURL 生成文件访问URL
func (t *Task) GetFileURL(baseURL string) string {
	if t.LocalPath != nil && *t.LocalPath != "" {
		return generateURL(baseURL, t.Provider, *t.LocalPath)
	}
	if t.NASPath != nil && *t.NASPath != "" {
		return generateURL(baseURL, t.Provider, *t.NASPath)
	}
	return ""
}

// GetThumbnailURL 生成缩略图访问URL
func (t *Task) GetThumbnailURL(baseURL string) string {
	if t.ThumbnailPath != nil && *t.ThumbnailPath != "" {
		return generateURL(baseURL, t.Provider, *t.ThumbnailPath)
	}
	return ""
}

// generateURL 生成访问URL
func generateURL(baseURL, provider, path string) string {
	if path == "" {
		return ""
	}
	
	// 提取相对路径（从provider目录开始）
	// 例如: "static\\hunyuan\\2026\\02\\file.glb" -> "2026/02/file.glb"
	relativePath := path
	
	// 查找provider目录的位置
	providerPrefix := "static\\" + provider + "\\"
	if idx := findSubstring(path, providerPrefix); idx >= 0 {
		relativePath = path[idx+len(providerPrefix):]
	} else {
		// 尝试正斜杠
		providerPrefix = "static/" + provider + "/"
		if idx := findSubstring(path, providerPrefix); idx >= 0 {
			relativePath = path[idx+len(providerPrefix):]
		}
	}
	
	// 统一使用正斜杠
	relativePath = replaceBackslash(relativePath)
	
	return fmt.Sprintf("%s/%s/%s", baseURL, provider, relativePath)
}

// findSubstring 查找子串位置
func findSubstring(s, substr string) int {
	if len(substr) == 0 {
		return 0
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// replaceBackslash 替换反斜杠为正斜杠
func replaceBackslash(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		if s[i] == '\\' {
			result[i] = '/'
		} else {
			result[i] = s[i]
		}
	}
	return string(result)
}

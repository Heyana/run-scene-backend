package hunyuan

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"go_wails_project_manager/config"

	"gorm.io/gorm"
)

// HunyuanTask 混元3D任务表
type HunyuanTask struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	// 任务信息
	JobID  string `gorm:"size:64;uniqueIndex" json:"jobId"`
	Status string `gorm:"size:20;index" json:"status"` // WAIT/RUN/DONE/FAIL

	// 输入参数
	InputType string  `gorm:"size:20" json:"inputType"` // text/image/multi_view
	Prompt    *string `gorm:"type:text" json:"prompt"`
	ImageURL  *string `gorm:"size:512" json:"imageUrl"`

	// 生成配置
	Model        string  `gorm:"size:10;default:'3.1'" json:"model"`
	FaceCount    *int    `json:"faceCount"`
	GenerateType string  `gorm:"size:20;default:'Normal'" json:"generateType"`
	EnablePBR    bool    `gorm:"default:true" json:"enablePbr"`
	ResultFormat string  `gorm:"size:20;default:'GLB'" json:"resultFormat"`

	// 结果
	ErrorCode    *string `gorm:"size:100" json:"errorCode"`
	ErrorMessage *string `gorm:"type:text" json:"errorMessage"`
	ResultFiles  *string `gorm:"type:text" json:"resultFiles"` // JSON数组

	// 文件存储
	LocalPath     *string `gorm:"size:512" json:"localPath"`     // 本地存储路径
	NASPath       *string `gorm:"size:512" json:"nasPath"`       // NAS存储路径
	ThumbnailPath *string `gorm:"size:512" json:"thumbnailPath"` // 预览图路径
	FileSize      *int64  `json:"fileSize"`                      // 文件大小
	FileHash      *string `gorm:"size:64" json:"fileHash"`       // 文件哈希

	// 元数据
	Name        string  `gorm:"size:200" json:"name"`                         // 模型名称
	Description *string `gorm:"type:text" json:"description"`                 // 描述
	Tags        *string `gorm:"type:text" json:"tags"`                        // 标签（JSON数组）
	Category    string  `gorm:"size:100;default:'AI生成'" json:"category"`     // 分类

	// 用户信息
	CreatedBy string `gorm:"size:100" json:"createdBy"`
	CreatedIP string `gorm:"size:50" json:"createdIp"`

	// 动态生成的URL字段（不存储到数据库）
	FileURL      string `gorm:"-" json:"fileUrl"`
	ThumbnailURL string `gorm:"-" json:"thumbnailUrl"`
}

// TableName 指定表名
func (HunyuanTask) TableName() string {
	return "hunyuan_tasks"
}

// AfterFind GORM钩子：查询后自动生成URL
func (t *HunyuanTask) AfterFind(tx *gorm.DB) error {
	// 生成文件URL
	if t.LocalPath != nil && *t.LocalPath != "" {
		t.FileURL = buildHunyuanURL(*t.LocalPath)
	} else if t.NASPath != nil && *t.NASPath != "" {
		t.FileURL = buildHunyuanURL(*t.NASPath)
	}

	// 生成缩略图URL
	if t.ThumbnailPath != nil && *t.ThumbnailPath != "" {
		t.ThumbnailURL = buildHunyuanURL(*t.ThumbnailPath)
	}

	return nil
}

// buildHunyuanURL 构建混元3D文件的访问URL
func buildHunyuanURL(path string) string {
	if path == "" {
		return ""
	}

	// 处理NAS路径（在标准化之前）
	// \\192.168.3.10\project\editor_v2\static\hunyuan\2026\02\file.glb
	// -> 2026/02/file.glb
	if strings.HasPrefix(path, "\\\\192.168.3.10\\") || strings.HasPrefix(path, "//192.168.3.10/") {
		// 查找 \hunyuan\ 或 /hunyuan/ 后面的部分
		idx := strings.Index(path, "\\hunyuan\\")
		if idx == -1 {
			idx = strings.Index(path, "/hunyuan/")
		}
		if idx != -1 {
			// 提取hunyuan后面的路径
			if strings.Contains(path, "\\hunyuan\\") {
				parts := strings.Split(path, "\\hunyuan\\")
				if len(parts) > 1 {
					path = parts[1]
				}
			} else {
				parts := strings.Split(path, "/hunyuan/")
				if len(parts) > 1 {
					path = parts[1]
				}
			}
		}
	}

	// 标准化路径分隔符
	path = filepath.ToSlash(path)

	// 移除可能的前缀
	path = strings.TrimPrefix(path, "static/hunyuan/")
	path = strings.TrimPrefix(path, "static\\hunyuan\\")
	path = strings.TrimPrefix(path, "./static/hunyuan/")
	path = strings.TrimPrefix(path, ".\\static\\hunyuan\\")

	// 确保路径使用正斜杠
	path = strings.ReplaceAll(path, "\\", "/")

	// 构建完整URL
	baseURL := config.AppConfig.Hunyuan.BaseURL
	if baseURL == "" {
		// 如果没有配置base_url，使用默认值
		if config.AppConfig.AppEnv == "production" {
			baseURL = fmt.Sprintf("http://%s:%d/hunyuan",
				config.AppConfig.PublicIP,
				config.AppConfig.ServerPort)
		} else {
			baseURL = fmt.Sprintf("http://%s:%d/hunyuan",
				config.AppConfig.LocalIP,
				config.AppConfig.ServerPort)
		}
	}

	return fmt.Sprintf("%s/%s", strings.TrimSuffix(baseURL, "/"), path)
}

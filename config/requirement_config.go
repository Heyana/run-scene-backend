package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// RequirementConfig 需求管理平台配置
type RequirementConfig struct {
	Requirement struct {
		Enabled bool `yaml:"enabled"`

		Mission struct {
			KeyPrefixLength   int      `yaml:"key_prefix_length"`
			DefaultPriority   string   `yaml:"default_priority"`
			DefaultStatus     string   `yaml:"default_status"`
			Statuses          []string `yaml:"statuses"`
			Priorities        []string `yaml:"priorities"`
			Types             []string `yaml:"types"`
			AutoAdjustPriority struct {
				Enabled       bool `yaml:"enabled"`
				DaysThreshold int  `yaml:"days_threshold"`
			} `yaml:"auto_adjust_priority"`
		} `yaml:"mission"`

		Attachment struct {
			AllowedTypes []string `yaml:"allowed_types"`
			MaxSize      int64    `yaml:"max_size"`
			StoragePath  string   `yaml:"storage_path"`
		} `yaml:"attachment"`

		Comment struct {
			MaxLength      int  `yaml:"max_length"`
			MentionEnabled bool `yaml:"mention_enabled"`
		} `yaml:"comment"`

		Permission struct {
			CompanyRoles []string `yaml:"company_roles"`
			ProjectRoles []string `yaml:"project_roles"`
		} `yaml:"permission"`

		Notification struct {
			Enabled bool     `yaml:"enabled"`
			Types   []string `yaml:"types"`
		} `yaml:"notification"`

		Statistics struct {
			CacheTTL       int `yaml:"cache_ttl"`
			BurndownPoints int `yaml:"burndown_points"`
		} `yaml:"statistics"`

		Performance struct {
			DefaultPageSize int `yaml:"default_page_size"`
			MaxPageSize     int `yaml:"max_page_size"`
			QueryTimeout    int `yaml:"query_timeout"`
			Cache           struct {
				Enabled bool `yaml:"enabled"`
				TTL     int  `yaml:"ttl"`
			} `yaml:"cache"`
		} `yaml:"performance"`
	} `yaml:"requirement"`
}

var RequirementCfg *RequirementConfig

// LoadRequirementConfig 加载需求管理配置
func LoadRequirementConfig() error {
	configPath := filepath.Join("configs", "requirement.yaml")

	data, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	RequirementCfg = &RequirementConfig{}
	if err := yaml.Unmarshal(data, RequirementCfg); err != nil {
		return err
	}

	return nil
}

// IsRequirementEnabled 检查需求管理功能是否启用
func IsRequirementEnabled() bool {
	if RequirementCfg == nil {
		return false
	}
	return RequirementCfg.Requirement.Enabled
}

// GetMissionStatuses 获取任务状态列表
func GetMissionStatuses() []string {
	if RequirementCfg == nil {
		return []string{"todo", "in_progress", "done", "closed"}
	}
	return RequirementCfg.Requirement.Mission.Statuses
}

// GetMissionPriorities 获取任务优先级列表
func GetMissionPriorities() []string {
	if RequirementCfg == nil {
		return []string{"P0", "P1", "P2", "P3"}
	}
	return RequirementCfg.Requirement.Mission.Priorities
}

// GetMissionTypes 获取任务类型列表
func GetMissionTypes() []string {
	if RequirementCfg == nil {
		return []string{"feature", "enhancement", "bug"}
	}
	return RequirementCfg.Requirement.Mission.Types
}

// IsFileTypeAllowed 检查文件类型是否允许
func IsFileTypeAllowed(fileExt string) bool {
	if RequirementCfg == nil {
		return false
	}
	for _, allowed := range RequirementCfg.Requirement.Attachment.AllowedTypes {
		if allowed == fileExt {
			return true
		}
	}
	return false
}

// GetMaxFileSize 获取最大文件大小
func GetMaxFileSize() int64 {
	if RequirementCfg == nil {
		return 10485760 // 10MB
	}
	return RequirementCfg.Requirement.Attachment.MaxSize
}

// GetAttachmentStoragePath 获取附件存储路径
func GetAttachmentStoragePath() string {
	if RequirementCfg == nil {
		return "uploads/requirement/attachments"
	}
	return RequirementCfg.Requirement.Attachment.StoragePath
}

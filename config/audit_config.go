// Package config 审计系统配置
package config

import (
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// AuditYAMLConfig 审计系统 YAML 配置结构
type AuditYAMLConfig struct {
	Audit struct {
		Enabled       bool `yaml:"enabled"`
		RetentionDays int  `yaml:"retention_days"`
		Archive       struct {
			Enabled              bool   `yaml:"enabled"`
			Schedule             string `yaml:"schedule"`
			NASEnabled           bool   `yaml:"nas_enabled"`
			NASPath              string `yaml:"nas_path"`
			LocalStorageEnabled  bool   `yaml:"local_storage_enabled"`
			StorageDir           string `yaml:"storage_dir"`
			Format               string `yaml:"format"`
			Compression          bool   `yaml:"compression"`
		} `yaml:"archive"`
		Log struct {
			Actions              []string `yaml:"actions"`
			Resources            []string `yaml:"resources"`
			ExcludePaths         []string `yaml:"exclude_paths"`
			SensitiveFields      []string `yaml:"sensitive_fields"`
			RecordRequestBody    bool     `yaml:"record_request_body"`
			MaxRequestBodySize   int      `yaml:"max_request_body_size"`
			RecordResponseBody   bool     `yaml:"record_response_body"`
			MaxResponseBodySize  int      `yaml:"max_response_body_size"`
		} `yaml:"log"`
		Performance struct {
			AsyncWrite    bool `yaml:"async_write"`
			BufferSize    int  `yaml:"buffer_size"`
			BatchSize     int  `yaml:"batch_size"`
			FlushInterval int  `yaml:"flush_interval"`
		} `yaml:"performance"`
		Query struct {
			DefaultPageSize       int  `yaml:"default_page_size"`
			MaxPageSize           int  `yaml:"max_page_size"`
			EnableFullTextSearch  bool `yaml:"enable_full_text_search"`
		} `yaml:"query"`
	} `yaml:"audit"`
}

// AuditConfig 审计系统配置结构
type AuditConfig struct {
	// 基础配置
	Enabled       bool
	RetentionDays int

	// 归档配置
	ArchiveEnabled          bool
	ArchiveSchedule         string
	ArchiveNASEnabled       bool
	ArchiveNASPath          string
	ArchiveLocalEnabled     bool
	ArchiveStorageDir       string
	ArchiveFormat           string
	ArchiveCompression      bool

	// 日志记录配置
	LogActions             []string
	LogResources           []string
	LogExcludePaths        []string
	LogSensitiveFields     []string
	RecordRequestBody      bool
	MaxRequestBodySize     int
	RecordResponseBody     bool
	MaxResponseBodySize    int

	// 性能配置
	AsyncWrite            bool
	BufferSize            int
	BatchSize             int
	FlushInterval         int

	// 查询配置
	DefaultPageSize       int
	MaxPageSize           int
	EnableFullTextSearch  bool
}

// LoadAuditConfig 加载审计系统配置
func LoadAuditConfig() (*AuditConfig, error) {
	// 1. 加载默认配置
	defaultConfig := getDefaultAuditConfig()

	// 2. 尝试从 YAML 文件加载
	yamlConfig := loadAuditYAML()
	if yamlConfig != nil {
		mergeAuditConfig(defaultConfig, yamlConfig)
	}

	// 3. 环境变量覆盖
	applyAuditEnvOverrides(defaultConfig)

	return defaultConfig, nil
}

// getDefaultAuditConfig 获取默认配置
func getDefaultAuditConfig() *AuditConfig {
	return &AuditConfig{
		// 基础配置
		Enabled:       true,
		RetentionDays: 7,

		// 归档配置
		ArchiveEnabled:      true,
		ArchiveSchedule:     "0 2 * * *", // 每天凌晨2点
		ArchiveNASEnabled:   true,
		ArchiveNASPath:      "",
		ArchiveLocalEnabled: false,
		ArchiveStorageDir:   "data/audit_archives",
		ArchiveFormat:       "json",
		ArchiveCompression:  true,

		// 日志记录配置
		LogActions: []string{
			"login", "logout", "create", "update", "delete",
			"download", "upload", "move", "rename", "share",
			"export", "import",
		},
		LogResources: []string{
			"document", "folder", "user", "project", "model", "texture",
		},
		LogExcludePaths: []string{
			"/api/health", "/api/ping", "/metrics",
			"/api/statistics/*", "/api/audit/*",
			"/website/*", "/documents/*", "/models/*", "/textures/*",
			"/assets/*", "/projects/*", "/hunyuan/*", "/meshy/*",
		},
		LogSensitiveFields: []string{
			"password", "token", "secret", "api_key",
			"access_token", "refresh_token",
		},
		RecordRequestBody:   true,
		MaxRequestBodySize:  10240, // 10KB
		RecordResponseBody:  false,
		MaxResponseBodySize: 10240,

		// 性能配置
		AsyncWrite:    true,
		BufferSize:    1000,
		BatchSize:     100,
		FlushInterval: 5,

		// 查询配置
		DefaultPageSize:      20,
		MaxPageSize:          100,
		EnableFullTextSearch: false,
	}
}

// loadAuditYAML 从 YAML 文件加载配置
func loadAuditYAML() *AuditYAMLConfig {
	configFile := "configs/audit.yaml"
	if _, err := os.Stat(configFile); err != nil {
		return nil
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil
	}

	var yamlConfig AuditYAMLConfig
	if err := yaml.Unmarshal(data, &yamlConfig); err != nil {
		return nil
	}

	return &yamlConfig
}

// mergeAuditConfig 合并 YAML 配置到默认配置
func mergeAuditConfig(config *AuditConfig, yamlConfig *AuditYAMLConfig) {
	audit := yamlConfig.Audit

	// 基础配置
	config.Enabled = audit.Enabled
	if audit.RetentionDays > 0 {
		config.RetentionDays = audit.RetentionDays
	}

	// 归档配置
	config.ArchiveEnabled = audit.Archive.Enabled
	if audit.Archive.Schedule != "" {
		config.ArchiveSchedule = audit.Archive.Schedule
	}
	config.ArchiveNASEnabled = audit.Archive.NASEnabled
	if audit.Archive.NASPath != "" {
		config.ArchiveNASPath = audit.Archive.NASPath
	}
	config.ArchiveLocalEnabled = audit.Archive.LocalStorageEnabled
	if audit.Archive.StorageDir != "" {
		config.ArchiveStorageDir = audit.Archive.StorageDir
	}
	if audit.Archive.Format != "" {
		config.ArchiveFormat = audit.Archive.Format
	}
	config.ArchiveCompression = audit.Archive.Compression

	// 日志记录配置
	if len(audit.Log.Actions) > 0 {
		config.LogActions = audit.Log.Actions
	}
	if len(audit.Log.Resources) > 0 {
		config.LogResources = audit.Log.Resources
	}
	if len(audit.Log.ExcludePaths) > 0 {
		config.LogExcludePaths = audit.Log.ExcludePaths
	}
	if len(audit.Log.SensitiveFields) > 0 {
		config.LogSensitiveFields = audit.Log.SensitiveFields
	}
	config.RecordRequestBody = audit.Log.RecordRequestBody
	if audit.Log.MaxRequestBodySize > 0 {
		config.MaxRequestBodySize = audit.Log.MaxRequestBodySize
	}
	config.RecordResponseBody = audit.Log.RecordResponseBody
	if audit.Log.MaxResponseBodySize > 0 {
		config.MaxResponseBodySize = audit.Log.MaxResponseBodySize
	}

	// 性能配置
	config.AsyncWrite = audit.Performance.AsyncWrite
	if audit.Performance.BufferSize > 0 {
		config.BufferSize = audit.Performance.BufferSize
	}
	if audit.Performance.BatchSize > 0 {
		config.BatchSize = audit.Performance.BatchSize
	}
	if audit.Performance.FlushInterval > 0 {
		config.FlushInterval = audit.Performance.FlushInterval
	}

	// 查询配置
	if audit.Query.DefaultPageSize > 0 {
		config.DefaultPageSize = audit.Query.DefaultPageSize
	}
	if audit.Query.MaxPageSize > 0 {
		config.MaxPageSize = audit.Query.MaxPageSize
	}
	config.EnableFullTextSearch = audit.Query.EnableFullTextSearch
}

// applyAuditEnvOverrides 应用环境变量覆盖
func applyAuditEnvOverrides(config *AuditConfig) {
	// 基础配置
	if val := os.Getenv("AUDIT_ENABLED"); val != "" {
		if b, err := strconv.ParseBool(val); err == nil {
			config.Enabled = b
		}
	}
	if val := os.Getenv("AUDIT_RETENTION_DAYS"); val != "" {
		if days, err := strconv.Atoi(val); err == nil {
			config.RetentionDays = days
		}
	}

	// 归档配置
	if val := os.Getenv("AUDIT_ARCHIVE_ENABLED"); val != "" {
		if b, err := strconv.ParseBool(val); err == nil {
			config.ArchiveEnabled = b
		}
	}
	if val := os.Getenv("AUDIT_ARCHIVE_SCHEDULE"); val != "" {
		config.ArchiveSchedule = val
	}
	if val := os.Getenv("AUDIT_ARCHIVE_NAS_ENABLED"); val != "" {
		if b, err := strconv.ParseBool(val); err == nil {
			config.ArchiveNASEnabled = b
		}
	}
	if val := os.Getenv("AUDIT_ARCHIVE_NAS_PATH"); val != "" {
		config.ArchiveNASPath = val
	}
	if val := os.Getenv("AUDIT_ARCHIVE_LOCAL_ENABLED"); val != "" {
		if b, err := strconv.ParseBool(val); err == nil {
			config.ArchiveLocalEnabled = b
		}
	}
	if val := os.Getenv("AUDIT_ARCHIVE_STORAGE_DIR"); val != "" {
		config.ArchiveStorageDir = val
	}
	if val := os.Getenv("AUDIT_ARCHIVE_FORMAT"); val != "" {
		config.ArchiveFormat = val
	}
	if val := os.Getenv("AUDIT_ARCHIVE_COMPRESSION"); val != "" {
		if b, err := strconv.ParseBool(val); err == nil {
			config.ArchiveCompression = b
		}
	}

	// 日志记录配置
	if val := os.Getenv("AUDIT_LOG_ACTIONS"); val != "" {
		config.LogActions = strings.Split(val, ",")
	}
	if val := os.Getenv("AUDIT_LOG_RESOURCES"); val != "" {
		config.LogResources = strings.Split(val, ",")
	}
	if val := os.Getenv("AUDIT_LOG_EXCLUDE_PATHS"); val != "" {
		config.LogExcludePaths = strings.Split(val, ",")
	}
	if val := os.Getenv("AUDIT_RECORD_REQUEST_BODY"); val != "" {
		if b, err := strconv.ParseBool(val); err == nil {
			config.RecordRequestBody = b
		}
	}
	if val := os.Getenv("AUDIT_MAX_REQUEST_BODY_SIZE"); val != "" {
		if size, err := strconv.Atoi(val); err == nil {
			config.MaxRequestBodySize = size
		}
	}

	// 性能配置
	if val := os.Getenv("AUDIT_ASYNC_WRITE"); val != "" {
		if b, err := strconv.ParseBool(val); err == nil {
			config.AsyncWrite = b
		}
	}
	if val := os.Getenv("AUDIT_BUFFER_SIZE"); val != "" {
		if size, err := strconv.Atoi(val); err == nil {
			config.BufferSize = size
		}
	}
	if val := os.Getenv("AUDIT_BATCH_SIZE"); val != "" {
		if size, err := strconv.Atoi(val); err == nil {
			config.BatchSize = size
		}
	}

	// 查询配置
	if val := os.Getenv("AUDIT_DEFAULT_PAGE_SIZE"); val != "" {
		if size, err := strconv.Atoi(val); err == nil {
			config.DefaultPageSize = size
		}
	}
	if val := os.Getenv("AUDIT_MAX_PAGE_SIZE"); val != "" {
		if size, err := strconv.Atoi(val); err == nil {
			config.MaxPageSize = size
		}
	}
}

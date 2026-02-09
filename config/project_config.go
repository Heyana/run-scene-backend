package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// ProjectConfig 项目管理配置
type ProjectConfig struct {
	LocalStorageEnabled   bool
	StorageDir            string
	BaseURL               string
	NASEnabled            bool
	NASPath               string
	NASHistoryPath        string // 历史版本存储路径
	MaxFileSize           int64
	MaxProjectSize        int64
	DefaultInitialVersion string
	VersionFormat         string
	EnableCompression     bool
	CompressionFormat     string
	PreviewEnabled        bool
	PreviewIndexFile      string
}

// ProjectYAMLConfig YAML配置结构
type ProjectYAMLConfig struct {
	Project struct {
		LocalStorageEnabled   bool   `yaml:"local_storage_enabled"`
		StorageDir            string `yaml:"storage_dir"`
		BaseURL               string `yaml:"base_url"`
		NASEnabled            bool   `yaml:"nas_enabled"`
		NASPath               string `yaml:"nas_path"`
		NASHistoryPath        string `yaml:"nas_history_path"`
		MaxFileSize           int64  `yaml:"max_file_size"`
		MaxProjectSize        int64  `yaml:"max_project_size"`
		DefaultInitialVersion string `yaml:"default_initial_version"`
		VersionFormat         string `yaml:"version_format"`
		EnableCompression     bool   `yaml:"enable_compression"`
		CompressionFormat     string `yaml:"compression_format"`
		PreviewEnabled        bool   `yaml:"preview_enabled"`
		PreviewIndexFile      string `yaml:"preview_index_file"`
	} `yaml:"project"`
}

// ProjectAppConfig 全局项目配置实例
var ProjectAppConfig *ProjectConfig

// LoadProjectConfig 加载项目配置
func LoadProjectConfig() error {
	// 加载YAML配置
	yamlConfig := loadProjectYAMLConfig()

	// 使用YAML配置初始化，支持环境变量覆盖
	ProjectAppConfig = &ProjectConfig{
		LocalStorageEnabled:   getEnvAsBoolOrDefault("PROJECT_LOCAL_STORAGE_ENABLED", yamlConfig.Project.LocalStorageEnabled),
		StorageDir:            getEnvOrDefault("PROJECT_STORAGE_DIR", yamlConfig.Project.StorageDir),
		BaseURL:               getEnvOrDefault("PROJECT_BASE_URL", yamlConfig.Project.BaseURL),
		NASEnabled:            getEnvAsBoolOrDefault("PROJECT_NAS_ENABLED", yamlConfig.Project.NASEnabled),
		NASPath:               getEnvOrDefault("PROJECT_NAS_PATH", yamlConfig.Project.NASPath),
		NASHistoryPath:        getEnvOrDefault("PROJECT_NAS_HISTORY_PATH", yamlConfig.Project.NASHistoryPath),
		MaxFileSize:           getEnvAsInt64OrDefault("PROJECT_MAX_FILE_SIZE", yamlConfig.Project.MaxFileSize),
		MaxProjectSize:        getEnvAsInt64OrDefault("PROJECT_MAX_PROJECT_SIZE", yamlConfig.Project.MaxProjectSize),
		DefaultInitialVersion: getEnvOrDefault("PROJECT_DEFAULT_INITIAL_VERSION", yamlConfig.Project.DefaultInitialVersion),
		VersionFormat:         getEnvOrDefault("PROJECT_VERSION_FORMAT", yamlConfig.Project.VersionFormat),
		EnableCompression:     getEnvAsBoolOrDefault("PROJECT_ENABLE_COMPRESSION", yamlConfig.Project.EnableCompression),
		CompressionFormat:     getEnvOrDefault("PROJECT_COMPRESSION_FORMAT", yamlConfig.Project.CompressionFormat),
		PreviewEnabled:        getEnvAsBoolOrDefault("PROJECT_PREVIEW_ENABLED", yamlConfig.Project.PreviewEnabled),
		PreviewIndexFile:      getEnvOrDefault("PROJECT_PREVIEW_INDEX_FILE", yamlConfig.Project.PreviewIndexFile),
	}

	return nil
}

// loadProjectYAMLConfig 从YAML文件加载项目配置
func loadProjectYAMLConfig() *ProjectYAMLConfig {
	// 默认配置
	defaultConfig := &ProjectYAMLConfig{}
	defaultConfig.Project.LocalStorageEnabled = false
	defaultConfig.Project.StorageDir = "static/projects"
	defaultConfig.Project.BaseURL = "http://192.168.3.10:23359/projects"
	defaultConfig.Project.NASEnabled = true
	defaultConfig.Project.NASPath = "/vol1/1003/project/editor_v2/static/projects"
	defaultConfig.Project.NASHistoryPath = "/vol1/1003/project/editor_v2/static/project_histories"
	defaultConfig.Project.MaxFileSize = 524288000      // 500MB
	defaultConfig.Project.MaxProjectSize = 1073741824  // 1GB
	defaultConfig.Project.DefaultInitialVersion = "1.0.0"
	defaultConfig.Project.VersionFormat = "semantic"
	defaultConfig.Project.EnableCompression = true
	defaultConfig.Project.CompressionFormat = "zip"
	defaultConfig.Project.PreviewEnabled = true
	defaultConfig.Project.PreviewIndexFile = "index.html"

	// 尝试读取YAML配置文件
	configFile := "configs/project_config.yaml"
	if _, err := os.Stat(configFile); err == nil {
		data, err := os.ReadFile(configFile)
		if err == nil {
			var yamlConfig ProjectYAMLConfig
			if err := yaml.Unmarshal(data, &yamlConfig); err == nil {
				// 用YAML配置覆盖默认值
				if yamlConfig.Project.StorageDir != "" {
					defaultConfig.Project = yamlConfig.Project
				}
			}
		}
	}

	return defaultConfig
}

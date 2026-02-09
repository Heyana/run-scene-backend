// Package config 文件库配置
package config

import (
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

// DocumentYAMLConfig 文件库 YAML 配置结构
type DocumentYAMLConfig struct {
	Document struct {
		LocalStorageEnabled bool                    `yaml:"local_storage_enabled"`
		StorageDir          string                  `yaml:"storage_dir"`
		BaseURL             string                  `yaml:"base_url"`
		NASEnabled          bool                    `yaml:"nas_enabled"`
		NASPath             string                  `yaml:"nas_path"`
		MaxFileSize         int64                   `yaml:"max_file_size"`
		AllowAllFormats     bool                    `yaml:"allow_all_formats"`
		AllowedFormats      map[string][]string     `yaml:"allowed_formats"`
		Preview             struct {
			Enabled         bool `yaml:"enabled"`
			PDFToImage      bool `yaml:"pdf_to_image"`
			MaxPreviewPages int  `yaml:"max_preview_pages"`
			PreviewQuality  int  `yaml:"preview_quality"`
			PreviewWidth    int  `yaml:"preview_width"`
		} `yaml:"preview"`
		Video struct {
			FFmpegPath      string  `yaml:"ffmpeg_path"`
			ThumbnailTime   float64 `yaml:"thumbnail_time"`
			ThumbnailCount  int     `yaml:"thumbnail_count"`
		} `yaml:"video"`
		VersionControl struct {
			Enabled     bool `yaml:"enabled"`
			MaxVersions int  `yaml:"max_versions"`
			AutoVersion bool `yaml:"auto_version"`
		} `yaml:"version_control"`
		Permission struct {
			EnableDepartment bool `yaml:"enable_department"`
			EnableProject    bool `yaml:"enable_project"`
			DefaultPublic    bool `yaml:"default_public"`
		} `yaml:"permission"`
		Log struct {
			Enabled          bool `yaml:"enabled"`
			LogAccess        bool `yaml:"log_access"`
			LogRetentionDays int  `yaml:"log_retention_days"`
		} `yaml:"log"`
	} `yaml:"document"`
}

// DocumentConfig 文件库配置结构
type DocumentConfig struct {
	// 存储配置
	LocalStorageEnabled bool
	StorageDir          string
	BaseURL             string
	NASEnabled          bool
	NASPath             string

	// 文件限制
	MaxFileSize     int64               // 统一文件大小限制（字节）
	AllowAllFormats bool                // 是否允许所有文件格式（开启后不验证文件类型）
	AllowedFormats  map[string][]string

	// 预览配置
	PreviewEnabled      bool
	PDFToImage          bool
	MaxPreviewPages     int
	PreviewQuality      int
	PreviewWidth        int

	// 视频配置
	FFmpegPath          string
	VideoThumbnailTime  float64
	VideoThumbnailCount int

	// 版本控制
	VersionControlEnabled bool
	MaxVersions           int
	AutoVersion           bool

	// 权限配置
	EnableDepartment bool
	EnableProject    bool
	DefaultPublic    bool

	// 日志配置
	LogEnabled       bool
	LogAccess        bool
	LogRetentionDays int
}

// LoadDocumentConfig 加载文件库配置
func LoadDocumentConfig() (*DocumentConfig, error) {
	// 1. 加载默认配置
	defaultConfig := getDefaultDocumentConfig()

	// 2. 尝试从 YAML 文件加载
	yamlConfig := loadDocumentYAML()
	if yamlConfig != nil {
		mergeDocumentConfig(defaultConfig, yamlConfig)
	}

	// 3. 环境变量覆盖
	applyDocumentEnvOverrides(defaultConfig)

	return defaultConfig, nil
}

// getDefaultDocumentConfig 获取默认配置
func getDefaultDocumentConfig() *DocumentConfig {
	return &DocumentConfig{
		// 存储配置
		LocalStorageEnabled: false,
		StorageDir:          "static/documents",
		BaseURL:             "",
		NASEnabled:          true,
		NASPath:             "",

		// 文件限制
		MaxFileSize: 10737418240, // 10GB
		AllowAllFormats: true, // 默认允许所有格式
		AllowedFormats: map[string][]string{
			"document": {"pdf", "doc", "docx", "xls", "xlsx", "ppt", "pptx", "txt", "md"},
			"video":    {"mp4", "webm", "avi", "mov"},
			"archive":  {"zip", "rar", "7z"},
			"other":    {"jpg", "png", "gif", "mp3", "wav"},
		},

		// 预览配置
		PreviewEnabled:  true,
		PDFToImage:      true,
		MaxPreviewPages: 10,
		PreviewQuality:  85,
		PreviewWidth:    800,

		// 视频配置
		FFmpegPath:          "ffmpeg",
		VideoThumbnailTime:  1.0,
		VideoThumbnailCount: 3,

		// 版本控制
		VersionControlEnabled: true,
		MaxVersions:           10,
		AutoVersion:           true,

		// 权限配置
		EnableDepartment: true,
		EnableProject:    true,
		DefaultPublic:    false,

		// 日志配置
		LogEnabled:       true,
		LogAccess:        true,
		LogRetentionDays: 90,
	}
}

// loadDocumentYAML 从 YAML 文件加载配置
func loadDocumentYAML() *DocumentYAMLConfig {
	configFile := "configs/document.yaml"
	if _, err := os.Stat(configFile); err != nil {
		return nil
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil
	}

	var yamlConfig DocumentYAMLConfig
	if err := yaml.Unmarshal(data, &yamlConfig); err != nil {
		return nil
	}

	return &yamlConfig
}

// mergeDocumentConfig 合并 YAML 配置到默认配置
func mergeDocumentConfig(config *DocumentConfig, yamlConfig *DocumentYAMLConfig) {
	doc := yamlConfig.Document

	// 存储配置
	if doc.StorageDir != "" {
		config.LocalStorageEnabled = doc.LocalStorageEnabled
		config.StorageDir = doc.StorageDir
		config.BaseURL = doc.BaseURL
		config.NASEnabled = doc.NASEnabled
		config.NASPath = doc.NASPath
	}

	// 文件限制
	if doc.MaxFileSize > 0 {
		config.MaxFileSize = doc.MaxFileSize
	}
	config.AllowAllFormats = doc.AllowAllFormats
	if len(doc.AllowedFormats) > 0 {
		config.AllowedFormats = doc.AllowedFormats
	}

	// 预览配置
	config.PreviewEnabled = doc.Preview.Enabled
	config.PDFToImage = doc.Preview.PDFToImage
	if doc.Preview.MaxPreviewPages > 0 {
		config.MaxPreviewPages = doc.Preview.MaxPreviewPages
	}
	if doc.Preview.PreviewQuality > 0 {
		config.PreviewQuality = doc.Preview.PreviewQuality
	}
	if doc.Preview.PreviewWidth > 0 {
		config.PreviewWidth = doc.Preview.PreviewWidth
	}

	// 视频配置
	if doc.Video.FFmpegPath != "" {
		config.FFmpegPath = doc.Video.FFmpegPath
	}
	if doc.Video.ThumbnailTime > 0 {
		config.VideoThumbnailTime = doc.Video.ThumbnailTime
	}
	if doc.Video.ThumbnailCount > 0 {
		config.VideoThumbnailCount = doc.Video.ThumbnailCount
	}

	// 版本控制
	config.VersionControlEnabled = doc.VersionControl.Enabled
	if doc.VersionControl.MaxVersions > 0 {
		config.MaxVersions = doc.VersionControl.MaxVersions
	}
	config.AutoVersion = doc.VersionControl.AutoVersion

	// 权限配置
	config.EnableDepartment = doc.Permission.EnableDepartment
	config.EnableProject = doc.Permission.EnableProject
	config.DefaultPublic = doc.Permission.DefaultPublic

	// 日志配置
	config.LogEnabled = doc.Log.Enabled
	config.LogAccess = doc.Log.LogAccess
	if doc.Log.LogRetentionDays > 0 {
		config.LogRetentionDays = doc.Log.LogRetentionDays
	}
}

// applyDocumentEnvOverrides 应用环境变量覆盖
func applyDocumentEnvOverrides(config *DocumentConfig) {
	// 存储配置
	if val := os.Getenv("DOCUMENT_LOCAL_STORAGE_ENABLED"); val != "" {
		if b, err := strconv.ParseBool(val); err == nil {
			config.LocalStorageEnabled = b
		}
	}
	if val := os.Getenv("DOCUMENT_STORAGE_DIR"); val != "" {
		config.StorageDir = val
	}
	if val := os.Getenv("DOCUMENT_BASE_URL"); val != "" {
		config.BaseURL = val
	}
	if val := os.Getenv("DOCUMENT_NAS_ENABLED"); val != "" {
		if b, err := strconv.ParseBool(val); err == nil {
			config.NASEnabled = b
		}
	}
	if val := os.Getenv("DOCUMENT_NAS_PATH"); val != "" {
		config.NASPath = val
	}

	// 文件格式限制
	if val := os.Getenv("DOCUMENT_ALLOW_ALL_FORMATS"); val != "" {
		if b, err := strconv.ParseBool(val); err == nil {
			config.AllowAllFormats = b
		}
	}

	// 文件大小限制
	if val := os.Getenv("DOCUMENT_MAX_FILE_SIZE"); val != "" {
		if size, err := strconv.ParseInt(val, 10, 64); err == nil {
			config.MaxFileSize = size
		}
	}

	// 预览配置
	if val := os.Getenv("DOCUMENT_PREVIEW_ENABLED"); val != "" {
		if b, err := strconv.ParseBool(val); err == nil {
			config.PreviewEnabled = b
		}
	}
	if val := os.Getenv("DOCUMENT_PDF_TO_IMAGE"); val != "" {
		if b, err := strconv.ParseBool(val); err == nil {
			config.PDFToImage = b
		}
	}

	// 视频配置
	if val := os.Getenv("DOCUMENT_FFMPEG_PATH"); val != "" {
		config.FFmpegPath = val
	}

	// 版本控制
	if val := os.Getenv("DOCUMENT_VERSION_CONTROL_ENABLED"); val != "" {
		if b, err := strconv.ParseBool(val); err == nil {
			config.VersionControlEnabled = b
		}
	}

	// 权限配置
	if val := os.Getenv("DOCUMENT_ENABLE_DEPARTMENT"); val != "" {
		if b, err := strconv.ParseBool(val); err == nil {
			config.EnableDepartment = b
		}
	}
	if val := os.Getenv("DOCUMENT_ENABLE_PROJECT"); val != "" {
		if b, err := strconv.ParseBool(val); err == nil {
			config.EnableProject = b
		}
	}

	// 日志配置
	if val := os.Getenv("DOCUMENT_LOG_ENABLED"); val != "" {
		if b, err := strconv.ParseBool(val); err == nil {
			config.LogEnabled = b
		}
	}
	if val := os.Getenv("DOCUMENT_LOG_ACCESS"); val != "" {
		if b, err := strconv.ParseBool(val); err == nil {
			config.LogAccess = b
		}
	}
}


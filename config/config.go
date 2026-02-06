// Package config 提供应用程序配置
package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

// ==================== 安全配置 ====================

// SecurityConfig 安全相关配置
type SecurityConfig struct {
	// 速率限制配置
	RateLimitPerSecond int // 每秒允许的请求数
	RateLimitBurst     int // 突发请求数

	// 文件上传限制
	MaxUploadSize      int64 // 单个文件最大大小（字节）
	MaxRequestBodySize int64 // 最大请求体大小（字节）

	// 连接限制
	MaxConnectionsPerIP      int   // 单个IP最大并发连接数
	MaxConcurrentConnections int   // 最大并发连接数
	ConnectionRatePerMinute  int   // 每分钟最大连接次数
	QuickBanThreshold        int   // 快速封禁阈值（每分钟请求数）
	QuickBanDuration         int64 // 快速封禁时长（秒）

	// IP封禁配置
	AutoBlockThreshold int   // 自动封禁阈值（可疑活动次数）
	AutoBlockDuration  int64 // 自动封禁时长（秒）

	// 慢速攻击检测
	SlowAttackTimeout      int64 // 慢速攻击超时（秒）
	EnableSlowAttackDetect bool  // 是否启用慢速攻击检测

	// CORS配置
	AllowedOrigins []string // 允许的源列表
}

// DefaultSecurityConfig 默认安全配置
var DefaultSecurityConfig = SecurityConfig{
	// 速率限制：适中的限制，平衡性能和安全
	RateLimitPerSecond: 100,
	RateLimitBurst:     200,

	// 文件上传限制
	MaxUploadSize:      50 * 1024 * 1024,  // 50MB
	MaxRequestBodySize: 100 * 1024 * 1024, // 100MB

	// 连接限制
	MaxConnectionsPerIP:      100,   // 单IP最大100个并发连接
	MaxConcurrentConnections: 100,   // 最大并发连接数
	ConnectionRatePerMinute:  600,   // 每分钟最多600次连接
	QuickBanThreshold:        1000,  // 每分钟超过1000次请求立即封禁
	QuickBanDuration:         600,   // 快速封禁10分钟
	SlowAttackTimeout:        30,    // 慢速攻击30秒超时
	EnableSlowAttackDetect:   false, // 默认禁用慢速攻击检测

	// IP封禁：保守的自动封禁策略
	AutoBlockThreshold: 50,           // 1小时内50次可疑活动封禁
	AutoBlockDuration:  24 * 60 * 60, // 封禁24小时

	// CORS配置：默认允许所有源（开发环境）
	AllowedOrigins: []string{"*"},
}

// YAMLConfig YAML配置文件结构
type YAMLConfig struct {
	App struct {
		Env  string `yaml:"env"`
		Name string `yaml:"name"`
		Port int    `yaml:"port"`
	} `yaml:"app"`

	Network struct {
		Local struct {
			IP      string `yaml:"ip"`
			CDNPort int    `yaml:"cdn_port"`
		} `yaml:"local"`
		Public struct {
			IP      string `yaml:"ip"`
			CDNPort int    `yaml:"cdn_port"`
		} `yaml:"public"`
	} `yaml:"network"`

	CDN struct {
		BaseURL string `yaml:"base_url"`
	} `yaml:"cdn"`

	APIDocs struct {
		BaseURL string `yaml:"base_url"`
	} `yaml:"api_docs"`

	Storage struct {
		DBPath        string `yaml:"db_path"`
		ProductDBPath string `yaml:"product_db_path"`
		ResourcePath  string `yaml:"resource_path"`
	} `yaml:"storage"`

	Logging struct {
		Level string `yaml:"level"`
	} `yaml:"logging"`

	Backup struct {
		Enabled       bool   `yaml:"enabled"`
		LocalPath     string `yaml:"local_path"`
		RetentionDays int    `yaml:"retention_days"`
		Environment   string `yaml:"environment"`
		AutoCleanup   bool   `yaml:"auto_cleanup"`
	} `yaml:"backup"`

	COS struct {
		Enabled   bool   `yaml:"enabled"`
		SecretID  string `yaml:"secret_id"`
		SecretKey string `yaml:"secret_key"`
		BucketURL string `yaml:"bucket_url"`
		Region    string `yaml:"region"`
	} `yaml:"cos"`

	Misc struct {
		MaxUploadSize  int64 `yaml:"max_upload_size"`
		SessionTimeout int64 `yaml:"session_timeout"`
	} `yaml:"misc"`

	Texture struct {
		LocalStorageEnabled bool   `yaml:"local_storage_enabled"`
		StorageDir          string `yaml:"storage_dir"`
		BaseURL             string `yaml:"base_url"`
		NASEnabled          bool   `yaml:"nas_enabled"`
		NASPath             string `yaml:"nas_path"`
		WebDAVEnabled       bool   `yaml:"webdav_enabled"`
		WebDAVURL           string `yaml:"webdav_url"`
		WebDAVUsername      string `yaml:"webdav_username"`
		WebDAVPassword      string `yaml:"webdav_password"`
		FTPEnabled          bool   `yaml:"ftp_enabled"`
		FTPHost             string `yaml:"ftp_host"`
		FTPPort             int    `yaml:"ftp_port"`
		FTPUsername         string `yaml:"ftp_username"`
		FTPPassword         string `yaml:"ftp_password"`
		FTPBasePath         string `yaml:"ftp_base_path"`
		SyncInterval        string `yaml:"sync_interval"`
		DownloadConcurrency int    `yaml:"download_concurrency"`
		RetryTimes          int    `yaml:"retry_times"`
		ThumbnailSize       int    `yaml:"thumbnail_size"`
		TextureResolution   int    `yaml:"texture_resolution"`
		WebPQuality         int    `yaml:"webp_quality"`
		DownloadThumbnail   bool   `yaml:"download_thumbnail"`
		DownloadTextures    bool   `yaml:"download_textures"`
		APIBaseURL          string `yaml:"api_base_url"`
		APITimeout          int    `yaml:"api_timeout"`
		ProxyEnabled        bool   `yaml:"proxy_enabled"`
		ProxyURL            string `yaml:"proxy_url"`
		LogEnabled          bool   `yaml:"log_enabled"`
		LogLevel            string `yaml:"log_level"`
		LogToFile           bool   `yaml:"log_to_file"`
		LogFilePath         string `yaml:"log_file_path"`
	} `yaml:"texture"`

	Model struct {
		LocalStorageEnabled bool     `yaml:"local_storage_enabled"`
		StorageDir          string   `yaml:"storage_dir"`
		BaseURL             string   `yaml:"base_url"`
		NASEnabled          bool     `yaml:"nas_enabled"`
		NASPath             string   `yaml:"nas_path"`
		MaxFileSize         int64    `yaml:"max_file_size"`
		MaxThumbnailSize    int64    `yaml:"max_thumbnail_size"`
		AllowedTypes        []string `yaml:"allowed_types"`
	} `yaml:"model"`

	Asset struct {
		LocalStorageEnabled bool `yaml:"local_storage_enabled"`
		StorageDir          string `yaml:"storage_dir"`
		BaseURL             string `yaml:"base_url"`
		NASEnabled          bool `yaml:"nas_enabled"`
		NASPath             string `yaml:"nas_path"`
		MaxFileSize         map[string]int64 `yaml:"max_file_size"`
		AllowedFormats      map[string][]string `yaml:"allowed_formats"`
		Thumbnail           struct {
			Width   int `yaml:"width"`
			Height  int `yaml:"height"`
			Quality int `yaml:"quality"`
		} `yaml:"thumbnail"`
		Video struct {
			FFmpegPath    string  `yaml:"ffmpeg_path"`
			ThumbnailTime float64 `yaml:"thumbnail_time"`
		} `yaml:"video"`
	} `yaml:"asset"`

	AI3D struct {
		DefaultProvider string `yaml:"default_provider"`
	} `yaml:"ai3d"`

	Hunyuan struct {
		SecretID            string `yaml:"secret_id"`
		SecretKey           string `yaml:"secret_key"`
		Region              string `yaml:"region"`
		DefaultModel        string `yaml:"default_model"`
		DefaultFaceCount    int    `yaml:"default_face_count"`
		DefaultGenerateType string `yaml:"default_generate_type"`
		DefaultEnablePBR    bool   `yaml:"default_enable_pbr"`
		DefaultResultFormat string `yaml:"default_result_format"`
		MaxConcurrent       int    `yaml:"max_concurrent"`
		PollInterval        int    `yaml:"poll_interval"`
		TaskTimeout         int    `yaml:"task_timeout"`
		LocalStorageEnabled bool   `yaml:"local_storage_enabled"`
		StorageDir          string `yaml:"storage_dir"`
		BaseURL             string `yaml:"base_url"`
		NASEnabled          bool   `yaml:"nas_enabled"`
		NASPath             string `yaml:"nas_path"`
		DefaultCategory     string `yaml:"default_category"`
		MaxRetryTimes       int    `yaml:"max_retry_times"`
		RetryInterval       int    `yaml:"retry_interval"`
	} `yaml:"hunyuan"`

	Meshy struct {
		APIKey                 string `yaml:"api_key"`
		BaseURL                string `yaml:"base_url"`
		DefaultAIModel         string `yaml:"default_ai_model"`
		DefaultEnablePBR       bool   `yaml:"default_enable_pbr"`
		DefaultTopology        string `yaml:"default_topology"`
		DefaultTargetPolycount int    `yaml:"default_target_polycount"`
		DefaultShouldRemesh    bool   `yaml:"default_should_remesh"`
		DefaultShouldTexture   bool   `yaml:"default_should_texture"`
		DefaultSavePreRemeshed bool   `yaml:"default_save_pre_remeshed"`
		DefaultResultFormat    string `yaml:"default_result_format"`
		MaxConcurrent          int    `yaml:"max_concurrent"`
		PollInterval           int    `yaml:"poll_interval"`
		TaskTimeout            int    `yaml:"task_timeout"`
		LocalStorageEnabled    bool   `yaml:"local_storage_enabled"`
		StorageDir             string `yaml:"storage_dir"`
		BaseURLCDN             string `yaml:"base_url_cdn"`
		NASEnabled             bool   `yaml:"nas_enabled"`
		NASPath                string `yaml:"nas_path"`
		DefaultCategory        string `yaml:"default_category"`
		MaxRetryTimes          int    `yaml:"max_retry_times"`
		RetryInterval          int    `yaml:"retry_interval"`
	} `yaml:"meshy"`
}

// Config 应用程序配置结构
type Config struct {
	AppEnv        string         // 应用环境 (development, production)
	ProjectName   string         // 项目名称
	ServerPort    int            // 服务器端口
	LogLevel      logrus.Level   // 日志级别
	DBPath        string         // 数据库路径
	ProductDBPath string         // 产品数据库路径
	CDNBasePath   string         // CDN基础路径
	ResourcePath  string         // 资源文件存储路径
	LocalIP       string         // 内网IP
	LocalCDNPort  int            // 内网CDN端口
	PublicIP      string         // 公网IP
	PublicCDNPort int            // 公网CDN端口
	Security      SecurityConfig // 安全配置
	Texture       TextureConfig  // 贴图库配置
	Model         ModelConfig    // 模型库配置
	Asset         AssetConfig    // 资产库配置
	AI3D          AI3DConfig     // AI 3D平台配置
	Hunyuan       HunyuanConfig  // 混元3D配置
	Meshy         MeshyConfig    // Meshy配置
}

// AI3DConfig AI 3D平台配置
type AI3DConfig struct {
	DefaultProvider string // 默认平台
}

// MeshyConfig Meshy配置
type MeshyConfig struct {
	APIKey                 string
	BaseURL                string
	DefaultAIModel         string
	DefaultEnablePBR       bool
	DefaultTopology        string
	DefaultTargetPolycount int
	DefaultShouldRemesh    bool
	DefaultShouldTexture   bool
	DefaultSavePreRemeshed bool
	DefaultResultFormat    string
	MaxConcurrent          int
	PollInterval           int
	TaskTimeout            int
	LocalStorageEnabled    bool
	StorageDir             string
	BaseURLCDN             string
	NASEnabled             bool
	NASPath                string
	DefaultCategory        string
	MaxRetryTimes          int
	RetryInterval          int
}

// TextureConfig 贴图库配置
type TextureConfig struct {
	LocalStorageEnabled bool
	StorageDir          string
	BaseURL             string
	NASEnabled          bool
	NASPath             string
	WebDAVEnabled       bool
	WebDAVURL           string
	WebDAVUsername      string
	WebDAVPassword      string
	FTPEnabled          bool
	FTPHost             string
	FTPPort             int
	FTPUsername         string
	FTPPassword         string
	FTPBasePath         string
	SyncInterval        string
	DownloadConcurrency int
	RetryTimes          int
	ThumbnailSize       int
	TextureResolution   int
	WebPQuality         int
	DownloadThumbnail   bool
	DownloadTextures    bool
	APIBaseURL          string
	APITimeout          int
	ProxyEnabled        bool
	ProxyURL            string
	LogEnabled          bool
	LogLevel            string
	LogToFile           bool
	LogFilePath         string
}

// ModelConfig 模型库配置
type ModelConfig struct {
	// 本地存储配置
	LocalStorageEnabled bool     // 是否保存到本地
	StorageDir          string   // 存储目录
	BaseURL             string   // 网络访问地址
	
	// NAS存储配置
	NASEnabled bool   // 是否启用NAS存储
	NASPath    string // NAS SMB共享路径
	
	// 文件限制
	MaxFileSize      int64    // 最大文件大小（字节）
	MaxThumbnailSize int64    // 最大预览图大小（字节）
	AllowedTypes     []string // 允许的文件类型
}

// AssetConfig 资产库配置
type AssetConfig struct {
	// 本地存储配置
	LocalStorageEnabled bool   // 是否保存到本地
	StorageDir          string // 存储目录
	BaseURL             string // 网络访问地址
	
	// NAS存储配置
	NASEnabled bool   // 是否启用NAS存储
	NASPath    string // NAS SMB共享路径
	
	// 文件限制（按类型）
	MaxFileSize    map[string]int64    // 最大文件大小（字节）
	AllowedFormats map[string][]string // 允许的文件格式
	
	// 缩略图配置
	ThumbnailWidth   int // 缩略图宽度
	ThumbnailHeight  int // 缩略图高度
	ThumbnailQuality int // 缩略图质量
	
	// 视频处理配置
	FFmpegPath         string  // FFmpeg可执行文件路径
	VideoThumbnailTime float64 // 视频截图时间点（秒）
}

// HunyuanConfig 混元3D配置
type HunyuanConfig struct {
	// API配置
	SecretID  string // 腾讯云SecretId
	SecretKey string // 腾讯云SecretKey
	Region    string // 地域
	
	// 默认参数
	DefaultModel        string // 默认模型版本
	DefaultFaceCount    int    // 默认面数
	DefaultGenerateType string // 默认生成类型
	DefaultEnablePBR    bool   // 默认PBR材质
	DefaultResultFormat string // 默认结果格式
	
	// 任务控制
	MaxConcurrent int // 最大并发任务数
	PollInterval  int // 轮询间隔（秒）
	TaskTimeout   int // 任务超时（秒）
	
	// 存储配置
	LocalStorageEnabled bool   // 是否启用本地存储
	StorageDir          string // 存储目录
	BaseURL             string // 网络访问地址
	NASEnabled          bool   // 是否启用NAS存储
	NASPath             string // NAS SMB共享路径
	DefaultCategory     string // 默认分类
	
	// 重试配置
	MaxRetryTimes int // 最大重试次数
	RetryInterval int // 重试间隔（秒）
}

// AppConfig 全局配置实例
var AppConfig *Config

// LoadConfig 加载配置
func LoadConfig() error {
	// 优先级: 环境变量 > YAML配置文件 > 默认值

	// 1. 从YAML文件加载基础配置
	yamlConfig := loadYAMLConfig()

	// 2. 使用YAML配置初始化，支持环境变量覆盖
	AppConfig = &Config{
		AppEnv:        getEnvOrDefault("APP_ENV", yamlConfig.App.Env),
		ProjectName:   getEnvOrDefault("PROJECT_NAME", yamlConfig.App.Name),
		ServerPort:    getEnvAsIntOrDefault("SERVER_PORT", yamlConfig.App.Port),
		DBPath:        getEnvOrDefault("DB_PATH", yamlConfig.Storage.DBPath),
		ProductDBPath: getEnvOrDefault("PRODUCT_DB_PATH", yamlConfig.Storage.ProductDBPath),
		ResourcePath:  getEnvOrDefault("RESOURCE_PATH", yamlConfig.Storage.ResourcePath),
		LocalIP:       getEnvOrDefault("LOCAL_IP", yamlConfig.Network.Local.IP),
		LocalCDNPort:  getEnvAsIntOrDefault("LOCAL_CDN_PORT", yamlConfig.Network.Local.CDNPort),
		PublicIP:      getEnvOrDefault("PUBLIC_IP", yamlConfig.Network.Public.IP),
		PublicCDNPort: getEnvAsIntOrDefault("PUBLIC_CDN_PORT", yamlConfig.Network.Public.CDNPort),
		Security:      DefaultSecurityConfig, // 使用默认安全配置
		Texture: TextureConfig{
			LocalStorageEnabled: getEnvAsBoolOrDefault("TEXTURE_LOCAL_STORAGE_ENABLED", yamlConfig.Texture.LocalStorageEnabled),
			StorageDir:          getEnvOrDefault("TEXTURE_STORAGE_DIR", yamlConfig.Texture.StorageDir),
			BaseURL:             getEnvOrDefault("TEXTURE_BASE_URL", yamlConfig.Texture.BaseURL),
			NASEnabled:          getEnvAsBoolOrDefault("TEXTURE_NAS_ENABLED", yamlConfig.Texture.NASEnabled),
			NASPath:             getEnvOrDefault("TEXTURE_NAS_PATH", yamlConfig.Texture.NASPath),
			WebDAVEnabled:       getEnvAsBoolOrDefault("TEXTURE_WEBDAV_ENABLED", yamlConfig.Texture.WebDAVEnabled),
			WebDAVURL:           getEnvOrDefault("TEXTURE_WEBDAV_URL", yamlConfig.Texture.WebDAVURL),
			WebDAVUsername:      getEnvOrDefault("TEXTURE_WEBDAV_USERNAME", yamlConfig.Texture.WebDAVUsername),
			WebDAVPassword:      getEnvOrDefault("TEXTURE_WEBDAV_PASSWORD", yamlConfig.Texture.WebDAVPassword),
			FTPEnabled:          getEnvAsBoolOrDefault("TEXTURE_FTP_ENABLED", yamlConfig.Texture.FTPEnabled),
			FTPHost:             getEnvOrDefault("TEXTURE_FTP_HOST", yamlConfig.Texture.FTPHost),
			FTPPort:             getEnvAsIntOrDefault("TEXTURE_FTP_PORT", yamlConfig.Texture.FTPPort),
			FTPUsername:         getEnvOrDefault("TEXTURE_FTP_USERNAME", yamlConfig.Texture.FTPUsername),
			FTPPassword:         getEnvOrDefault("TEXTURE_FTP_PASSWORD", yamlConfig.Texture.FTPPassword),
			FTPBasePath:         getEnvOrDefault("TEXTURE_FTP_BASE_PATH", yamlConfig.Texture.FTPBasePath),
			SyncInterval:        getEnvOrDefault("TEXTURE_SYNC_INTERVAL", yamlConfig.Texture.SyncInterval),
			DownloadConcurrency: getEnvAsIntOrDefault("TEXTURE_DOWNLOAD_CONCURRENCY", yamlConfig.Texture.DownloadConcurrency),
			RetryTimes:          getEnvAsIntOrDefault("TEXTURE_RETRY_TIMES", yamlConfig.Texture.RetryTimes),
			ThumbnailSize:       getEnvAsIntOrDefault("TEXTURE_THUMBNAIL_SIZE", yamlConfig.Texture.ThumbnailSize),
			TextureResolution:   getEnvAsIntOrDefault("TEXTURE_RESOLUTION", yamlConfig.Texture.TextureResolution),
			WebPQuality:         getEnvAsIntOrDefault("TEXTURE_WEBP_QUALITY", yamlConfig.Texture.WebPQuality),
			DownloadThumbnail:   getEnvAsBoolOrDefault("TEXTURE_DOWNLOAD_THUMBNAIL", yamlConfig.Texture.DownloadThumbnail),
			DownloadTextures:    getEnvAsBoolOrDefault("TEXTURE_DOWNLOAD_TEXTURES", yamlConfig.Texture.DownloadTextures),
			APIBaseURL:          getEnvOrDefault("TEXTURE_API_BASE_URL", yamlConfig.Texture.APIBaseURL),
			APITimeout:          getEnvAsIntOrDefault("TEXTURE_API_TIMEOUT", yamlConfig.Texture.APITimeout),
			ProxyEnabled:        getEnvAsBoolOrDefault("TEXTURE_PROXY_ENABLED", yamlConfig.Texture.ProxyEnabled),
			ProxyURL:            getEnvOrDefault("TEXTURE_PROXY_URL", yamlConfig.Texture.ProxyURL),
			LogEnabled:          getEnvAsBoolOrDefault("TEXTURE_LOG_ENABLED", yamlConfig.Texture.LogEnabled),
			LogLevel:            getEnvOrDefault("TEXTURE_LOG_LEVEL", yamlConfig.Texture.LogLevel),
			LogToFile:           getEnvAsBoolOrDefault("TEXTURE_LOG_TO_FILE", yamlConfig.Texture.LogToFile),
			LogFilePath:         getEnvOrDefault("TEXTURE_LOG_FILE_PATH", yamlConfig.Texture.LogFilePath),
		},
		Model: ModelConfig{
			LocalStorageEnabled: getEnvAsBoolOrDefault("MODEL_LOCAL_STORAGE_ENABLED", yamlConfig.Model.LocalStorageEnabled),
			StorageDir:          getEnvOrDefault("MODEL_STORAGE_DIR", yamlConfig.Model.StorageDir),
			BaseURL:             getEnvOrDefault("MODEL_BASE_URL", yamlConfig.Model.BaseURL),
			NASEnabled:          getEnvAsBoolOrDefault("MODEL_NAS_ENABLED", yamlConfig.Model.NASEnabled),
			NASPath:             getEnvOrDefault("MODEL_NAS_PATH", yamlConfig.Model.NASPath),
			MaxFileSize:         getEnvAsInt64OrDefault("MODEL_MAX_FILE_SIZE", yamlConfig.Model.MaxFileSize),
			MaxThumbnailSize:    getEnvAsInt64OrDefault("MODEL_MAX_THUMBNAIL_SIZE", yamlConfig.Model.MaxThumbnailSize),
			AllowedTypes:        yamlConfig.Model.AllowedTypes,
		},
		Asset: AssetConfig{
			LocalStorageEnabled: getEnvAsBoolOrDefault("ASSET_LOCAL_STORAGE_ENABLED", yamlConfig.Asset.LocalStorageEnabled),
			StorageDir:          getEnvOrDefault("ASSET_STORAGE_DIR", yamlConfig.Asset.StorageDir),
			BaseURL:             getEnvOrDefault("ASSET_BASE_URL", yamlConfig.Asset.BaseURL),
			NASEnabled:          getEnvAsBoolOrDefault("ASSET_NAS_ENABLED", yamlConfig.Asset.NASEnabled),
			NASPath:             getEnvOrDefault("ASSET_NAS_PATH", yamlConfig.Asset.NASPath),
			MaxFileSize:         yamlConfig.Asset.MaxFileSize,
			AllowedFormats:      yamlConfig.Asset.AllowedFormats,
			ThumbnailWidth:      getEnvAsIntOrDefault("ASSET_THUMBNAIL_WIDTH", yamlConfig.Asset.Thumbnail.Width),
			ThumbnailHeight:     getEnvAsIntOrDefault("ASSET_THUMBNAIL_HEIGHT", yamlConfig.Asset.Thumbnail.Height),
			ThumbnailQuality:    getEnvAsIntOrDefault("ASSET_THUMBNAIL_QUALITY", yamlConfig.Asset.Thumbnail.Quality),
			FFmpegPath:          getEnvOrDefault("ASSET_FFMPEG_PATH", yamlConfig.Asset.Video.FFmpegPath),
			VideoThumbnailTime:  yamlConfig.Asset.Video.ThumbnailTime,
		},
		AI3D: AI3DConfig{
			DefaultProvider: getEnvOrDefault("AI3D_DEFAULT_PROVIDER", yamlConfig.AI3D.DefaultProvider),
		},
		Hunyuan: HunyuanConfig{
			SecretID:            getEnvOrDefault("HUNYUAN_SECRET_ID", yamlConfig.Hunyuan.SecretID),
			SecretKey:           getEnvOrDefault("HUNYUAN_SECRET_KEY", yamlConfig.Hunyuan.SecretKey),
			Region:              getEnvOrDefault("HUNYUAN_REGION", yamlConfig.Hunyuan.Region),
			DefaultModel:        getEnvOrDefault("HUNYUAN_DEFAULT_MODEL", yamlConfig.Hunyuan.DefaultModel),
			DefaultFaceCount:    getEnvAsIntOrDefault("HUNYUAN_DEFAULT_FACE_COUNT", yamlConfig.Hunyuan.DefaultFaceCount),
			DefaultGenerateType: getEnvOrDefault("HUNYUAN_DEFAULT_GENERATE_TYPE", yamlConfig.Hunyuan.DefaultGenerateType),
			DefaultEnablePBR:    getEnvAsBoolOrDefault("HUNYUAN_DEFAULT_ENABLE_PBR", yamlConfig.Hunyuan.DefaultEnablePBR),
			DefaultResultFormat: getEnvOrDefault("HUNYUAN_DEFAULT_RESULT_FORMAT", yamlConfig.Hunyuan.DefaultResultFormat),
			MaxConcurrent:       getEnvAsIntOrDefault("HUNYUAN_MAX_CONCURRENT", yamlConfig.Hunyuan.MaxConcurrent),
			PollInterval:        getEnvAsIntOrDefault("HUNYUAN_POLL_INTERVAL", yamlConfig.Hunyuan.PollInterval),
			TaskTimeout:         getEnvAsIntOrDefault("HUNYUAN_TASK_TIMEOUT", yamlConfig.Hunyuan.TaskTimeout),
			LocalStorageEnabled: getEnvAsBoolOrDefault("HUNYUAN_LOCAL_STORAGE_ENABLED", yamlConfig.Hunyuan.LocalStorageEnabled),
			StorageDir:          getEnvOrDefault("HUNYUAN_STORAGE_DIR", yamlConfig.Hunyuan.StorageDir),
			BaseURL:             getEnvOrDefault("HUNYUAN_BASE_URL", yamlConfig.Hunyuan.BaseURL),
			NASEnabled:          getEnvAsBoolOrDefault("HUNYUAN_NAS_ENABLED", yamlConfig.Hunyuan.NASEnabled),
			NASPath:             getEnvOrDefault("HUNYUAN_NAS_PATH", yamlConfig.Hunyuan.NASPath),
			DefaultCategory:     getEnvOrDefault("HUNYUAN_DEFAULT_CATEGORY", yamlConfig.Hunyuan.DefaultCategory),
			MaxRetryTimes:       getEnvAsIntOrDefault("HUNYUAN_MAX_RETRY_TIMES", yamlConfig.Hunyuan.MaxRetryTimes),
			RetryInterval:       getEnvAsIntOrDefault("HUNYUAN_RETRY_INTERVAL", yamlConfig.Hunyuan.RetryInterval),
		},
		Meshy: MeshyConfig{
			APIKey:                 getEnvOrDefault("MESHY_API_KEY", yamlConfig.Meshy.APIKey),
			BaseURL:                getEnvOrDefault("MESHY_BASE_URL", yamlConfig.Meshy.BaseURL),
			DefaultAIModel:         getEnvOrDefault("MESHY_DEFAULT_AI_MODEL", yamlConfig.Meshy.DefaultAIModel),
			DefaultEnablePBR:       getEnvAsBoolOrDefault("MESHY_DEFAULT_ENABLE_PBR", yamlConfig.Meshy.DefaultEnablePBR),
			DefaultTopology:        getEnvOrDefault("MESHY_DEFAULT_TOPOLOGY", yamlConfig.Meshy.DefaultTopology),
			DefaultTargetPolycount: getEnvAsIntOrDefault("MESHY_DEFAULT_TARGET_POLYCOUNT", yamlConfig.Meshy.DefaultTargetPolycount),
			DefaultShouldRemesh:    getEnvAsBoolOrDefault("MESHY_DEFAULT_SHOULD_REMESH", yamlConfig.Meshy.DefaultShouldRemesh),
			DefaultShouldTexture:   getEnvAsBoolOrDefault("MESHY_DEFAULT_SHOULD_TEXTURE", yamlConfig.Meshy.DefaultShouldTexture),
			DefaultSavePreRemeshed: getEnvAsBoolOrDefault("MESHY_DEFAULT_SAVE_PRE_REMESHED", yamlConfig.Meshy.DefaultSavePreRemeshed),
			DefaultResultFormat:    getEnvOrDefault("MESHY_DEFAULT_RESULT_FORMAT", yamlConfig.Meshy.DefaultResultFormat),
			MaxConcurrent:          getEnvAsIntOrDefault("MESHY_MAX_CONCURRENT", yamlConfig.Meshy.MaxConcurrent),
			PollInterval:           getEnvAsIntOrDefault("MESHY_POLL_INTERVAL", yamlConfig.Meshy.PollInterval),
			TaskTimeout:            getEnvAsIntOrDefault("MESHY_TASK_TIMEOUT", yamlConfig.Meshy.TaskTimeout),
			LocalStorageEnabled:    getEnvAsBoolOrDefault("MESHY_LOCAL_STORAGE_ENABLED", yamlConfig.Meshy.LocalStorageEnabled),
			StorageDir:             getEnvOrDefault("MESHY_STORAGE_DIR", yamlConfig.Meshy.StorageDir),
			BaseURLCDN:             getEnvOrDefault("MESHY_BASE_URL_CDN", yamlConfig.Meshy.BaseURLCDN),
			NASEnabled:             getEnvAsBoolOrDefault("MESHY_NAS_ENABLED", yamlConfig.Meshy.NASEnabled),
			NASPath:                getEnvOrDefault("MESHY_NAS_PATH", yamlConfig.Meshy.NASPath),
			DefaultCategory:        getEnvOrDefault("MESHY_DEFAULT_CATEGORY", yamlConfig.Meshy.DefaultCategory),
			MaxRetryTimes:          getEnvAsIntOrDefault("MESHY_MAX_RETRY_TIMES", yamlConfig.Meshy.MaxRetryTimes),
			RetryInterval:          getEnvAsIntOrDefault("MESHY_RETRY_INTERVAL", yamlConfig.Meshy.RetryInterval),
		},
	}

	// 3. 设置CDN基础路径（支持YAML和环境变量覆盖）
	if customCDN := os.Getenv("CDN_BASE_PATH"); customCDN != "" {
		AppConfig.CDNBasePath = customCDN
	} else if yamlConfig.CDN.BaseURL != "" {
		AppConfig.CDNBasePath = yamlConfig.CDN.BaseURL
	} else {
		AppConfig.CDNBasePath = getCDNBaseURL()
	}

	// 4. 设置日志级别（支持YAML和环境变量覆盖）
	logLevelStr := getEnvOrDefault("LOG_LEVEL", yamlConfig.Logging.Level)
	switch logLevelStr {
	case "debug":
		AppConfig.LogLevel = logrus.DebugLevel
	case "info":
		AppConfig.LogLevel = logrus.InfoLevel
	case "warn":
		AppConfig.LogLevel = logrus.WarnLevel
	case "error":
		AppConfig.LogLevel = logrus.ErrorLevel
	default:
		AppConfig.LogLevel = logrus.InfoLevel
	}

	// 5. 开发环境默认使用调试级别和放宽安全限制
	if AppConfig.AppEnv == "development" {
		if logLevelStr == "info" {
			AppConfig.LogLevel = logrus.DebugLevel
		}
		// 开发环境放宽安全限制
		AppConfig.Security.RateLimitPerSecond = 1000
		AppConfig.Security.RateLimitBurst = 2000
		AppConfig.Security.QuickBanThreshold = 10000
	}

	// 6. 尝试加载 .env 文件作为备选（向后兼容）
	if err := godotenv.Load(); err != nil {
		// .env 文件不存在时不报错，生产环境通常使用系统环境变量
	}

	// 7. 加载AI配置
	if err := LoadAIConfig(); err != nil {
		// AI配置加载失败不影响主程序启动
		logrus.Warnf("加载AI配置失败: %v", err)
	}

	return nil
}

// loadYAMLConfig 从YAML文件加载配置
func loadYAMLConfig() *YAMLConfig {
	// 默认配置
	defaultConfig := &YAMLConfig{}
	defaultConfig.App.Env = "development"
	defaultConfig.App.Name = "go_wails_project_manager"
	defaultConfig.App.Port = 23347
	defaultConfig.Network.Local.IP = "192.168.3.39"
	defaultConfig.Network.Local.CDNPort = 23357
	defaultConfig.Network.Public.IP = "111.229.160.27"
	defaultConfig.Network.Public.CDNPort = 23357
	defaultConfig.Storage.DBPath = "./data/app.db"
	defaultConfig.Storage.ProductDBPath = "./data/product.db"
	defaultConfig.Storage.ResourcePath = "./static/cdn"
	defaultConfig.Logging.Level = "info"
	defaultConfig.Backup.Enabled = true
	defaultConfig.Backup.LocalPath = "./backups"
	defaultConfig.Backup.RetentionDays = 7
	defaultConfig.Backup.Environment = "dev"
	defaultConfig.Backup.AutoCleanup = false
	defaultConfig.COS.Enabled = false
	defaultConfig.COS.Region = "ap-shanghai"
	defaultConfig.Misc.MaxUploadSize = 104857600 // 100MB
	defaultConfig.Misc.SessionTimeout = 86400    // 24小时
	
	// 贴图库默认配置
	defaultConfig.Texture.StorageDir = "static/textures"
	defaultConfig.Texture.SyncInterval = "6h"
	defaultConfig.Texture.DownloadConcurrency = 10
	defaultConfig.Texture.RetryTimes = 3
	defaultConfig.Texture.ThumbnailSize = 256
	defaultConfig.Texture.TextureResolution = 1024
	defaultConfig.Texture.WebPQuality = 80
	defaultConfig.Texture.DownloadThumbnail = true
	defaultConfig.Texture.DownloadTextures = true
	defaultConfig.Texture.APIBaseURL = "https://api.polyhaven.com"
	defaultConfig.Texture.APITimeout = 30
	defaultConfig.Texture.LogEnabled = true
	defaultConfig.Texture.LogLevel = "info"
	defaultConfig.Texture.LogToFile = true
	defaultConfig.Texture.LogFilePath = "./logs/texture_sync.log"
	
	// 模型库默认配置
	defaultConfig.Model.LocalStorageEnabled = true
	defaultConfig.Model.StorageDir = "static/models"
	defaultConfig.Model.BaseURL = ""
	defaultConfig.Model.NASEnabled = false
	defaultConfig.Model.NASPath = ""
	defaultConfig.Model.MaxFileSize = 104857600   // 100MB
	defaultConfig.Model.MaxThumbnailSize = 5242880 // 5MB
	defaultConfig.Model.AllowedTypes = []string{"glb", "glt"}
	
	// 资产库默认配置
	defaultConfig.Asset.LocalStorageEnabled = true
	defaultConfig.Asset.StorageDir = "static/assets"
	defaultConfig.Asset.BaseURL = ""
	defaultConfig.Asset.NASEnabled = false
	defaultConfig.Asset.NASPath = ""
	defaultConfig.Asset.MaxFileSize = map[string]int64{
		"image": 52428800,  // 50MB
		"video": 524288000, // 500MB
	}
	defaultConfig.Asset.AllowedFormats = map[string][]string{
		"image": {"jpg", "jpeg", "png", "webp"},
		"video": {"mp4", "webm"},
	}
	defaultConfig.Asset.Thumbnail.Width = 512
	defaultConfig.Asset.Thumbnail.Height = 512
	defaultConfig.Asset.Thumbnail.Quality = 85
	defaultConfig.Asset.Video.FFmpegPath = "ffmpeg"
	defaultConfig.Asset.Video.ThumbnailTime = 1.0
	
	// AI 3D平台默认配置
	defaultConfig.AI3D.DefaultProvider = "hunyuan"
	
	// 混元3D默认配置
	defaultConfig.Hunyuan.Region = "ap-guangzhou"
	defaultConfig.Hunyuan.DefaultModel = "3.1"
	defaultConfig.Hunyuan.DefaultFaceCount = 500000
	defaultConfig.Hunyuan.DefaultGenerateType = "Normal"
	defaultConfig.Hunyuan.DefaultEnablePBR = false
	defaultConfig.Hunyuan.DefaultResultFormat = "GLB"
	defaultConfig.Hunyuan.MaxConcurrent = 3
	defaultConfig.Hunyuan.PollInterval = 5
	defaultConfig.Hunyuan.TaskTimeout = 86400
	defaultConfig.Hunyuan.LocalStorageEnabled = true
	defaultConfig.Hunyuan.StorageDir = "static/hunyuan"
	defaultConfig.Hunyuan.NASEnabled = false
	defaultConfig.Hunyuan.DefaultCategory = "AI生成"
	defaultConfig.Hunyuan.MaxRetryTimes = 3
	defaultConfig.Hunyuan.RetryInterval = 10
	
	// Meshy默认配置
	defaultConfig.Meshy.BaseURL = "https://api.meshy.ai"
	defaultConfig.Meshy.DefaultAIModel = "meshy-6"
	defaultConfig.Meshy.DefaultEnablePBR = true
	defaultConfig.Meshy.DefaultTopology = "triangle"
	defaultConfig.Meshy.DefaultTargetPolycount = 30000
	defaultConfig.Meshy.DefaultShouldRemesh = true
	defaultConfig.Meshy.DefaultShouldTexture = true
	defaultConfig.Meshy.DefaultSavePreRemeshed = true
	defaultConfig.Meshy.DefaultResultFormat = "GLB"
	defaultConfig.Meshy.MaxConcurrent = 3
	defaultConfig.Meshy.PollInterval = 5
	defaultConfig.Meshy.TaskTimeout = 86400
	defaultConfig.Meshy.LocalStorageEnabled = false
	defaultConfig.Meshy.StorageDir = "static/meshy"
	defaultConfig.Meshy.NASEnabled = false
	defaultConfig.Meshy.DefaultCategory = "AI生成"
	defaultConfig.Meshy.MaxRetryTimes = 3
	defaultConfig.Meshy.RetryInterval = 10

	// 尝试读取YAML配置文件
	configFile := "config.yaml"
	if _, err := os.Stat(configFile); err == nil {
		data, err := os.ReadFile(configFile)
		if err == nil {
			var yamlConfig YAMLConfig
			if err := yaml.Unmarshal(data, &yamlConfig); err == nil {
				// 用YAML配置覆盖默认值（只覆盖非零值）
				if yamlConfig.App.Env != "" {
					defaultConfig.App.Env = yamlConfig.App.Env
				}
				if yamlConfig.App.Name != "" {
					defaultConfig.App.Name = yamlConfig.App.Name
				}
				if yamlConfig.App.Port != 0 {
					defaultConfig.App.Port = yamlConfig.App.Port
				}
				if yamlConfig.Network.Local.IP != "" {
					defaultConfig.Network.Local.IP = yamlConfig.Network.Local.IP
				}
				if yamlConfig.Network.Local.CDNPort != 0 {
					defaultConfig.Network.Local.CDNPort = yamlConfig.Network.Local.CDNPort
				}
				if yamlConfig.Network.Public.IP != "" {
					defaultConfig.Network.Public.IP = yamlConfig.Network.Public.IP
				}
				if yamlConfig.Network.Public.CDNPort != 0 {
					defaultConfig.Network.Public.CDNPort = yamlConfig.Network.Public.CDNPort
				}
				if yamlConfig.CDN.BaseURL != "" {
					defaultConfig.CDN.BaseURL = yamlConfig.CDN.BaseURL
				}
				if yamlConfig.APIDocs.BaseURL != "" {
					defaultConfig.APIDocs.BaseURL = yamlConfig.APIDocs.BaseURL
				}
				if yamlConfig.Storage.DBPath != "" {
					defaultConfig.Storage.DBPath = yamlConfig.Storage.DBPath
				}
				if yamlConfig.Storage.ProductDBPath != "" {
					defaultConfig.Storage.ProductDBPath = yamlConfig.Storage.ProductDBPath
				}
				if yamlConfig.Storage.ResourcePath != "" {
					defaultConfig.Storage.ResourcePath = yamlConfig.Storage.ResourcePath
				}
				if yamlConfig.Logging.Level != "" {
					defaultConfig.Logging.Level = yamlConfig.Logging.Level
				}
				// 备份配置
				defaultConfig.Backup = yamlConfig.Backup
				// COS配置
				defaultConfig.COS = yamlConfig.COS
				// 其他配置
				if yamlConfig.Misc.MaxUploadSize != 0 {
					defaultConfig.Misc.MaxUploadSize = yamlConfig.Misc.MaxUploadSize
				}
				if yamlConfig.Misc.SessionTimeout != 0 {
					defaultConfig.Misc.SessionTimeout = yamlConfig.Misc.SessionTimeout
				}
				// 贴图库配置
				if yamlConfig.Texture.StorageDir != "" {
					defaultConfig.Texture = yamlConfig.Texture
				}
				// 模型库配置
				if yamlConfig.Model.StorageDir != "" {
					defaultConfig.Model = yamlConfig.Model
				}
				// 资产库配置
				if yamlConfig.Asset.StorageDir != "" {
					defaultConfig.Asset = yamlConfig.Asset
				}
				// AI 3D平台配置
				if yamlConfig.AI3D.DefaultProvider != "" {
					defaultConfig.AI3D = yamlConfig.AI3D
				}
				// 混元3D配置
				if yamlConfig.Hunyuan.StorageDir != "" {
					defaultConfig.Hunyuan = yamlConfig.Hunyuan
				}
				// Meshy配置
				if yamlConfig.Meshy.StorageDir != "" {
					defaultConfig.Meshy = yamlConfig.Meshy
				}
			}
		}
	}

	return defaultConfig
}

// GetSecurityConfig 获取安全配置
func GetSecurityConfig() SecurityConfig {
	return AppConfig.Security
}

// UpdateSecurityConfig 更新安全配置
func UpdateSecurityConfig(cfg SecurityConfig) {
	AppConfig.Security = cfg
}

// getEnvOrDefault 获取环境变量，如果不存在则返回默认值
func getEnvOrDefault(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// getEnvAsIntOrDefault 获取环境变量并转换为int，如果不存在或转换失败则返回默认值
func getEnvAsIntOrDefault(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvAsInt64OrDefault 获取环境变量并转换为int64，如果不存在或转换失败则返回默认值
func getEnvAsInt64OrDefault(key string, defaultValue int64) int64 {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvAsBoolOrDefault 获取环境变量并转换为bool，如果不存在或转换失败则返回默认值
func getEnvAsBoolOrDefault(key string, defaultValue bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

// getCDNBaseURL 根据环境返回CDN基础URL
func getCDNBaseURL() string {
	// 如果设置了CDN_BASE_PATH环境变量，则优先使用该值
	if customCDN := os.Getenv("CDN_BASE_PATH"); customCDN != "" {
		return customCDN
	}

	// 根据环境和配置动态生成CDN URL
	env := getEnvOrDefault("APP_ENV", "development")
	if env == "development" {
		// 开发环境使用内网配置
		localIP := getEnvOrDefault("LOCAL_IP", "192.168.3.39")
		localPort := getEnvAsIntOrDefault("LOCAL_CDN_PORT", 23357)
		return "http://" + localIP + ":" + strconv.Itoa(localPort) + "/cdn/"
	}

	// 生产环境使用公网配置
	publicIP := getEnvOrDefault("PUBLIC_IP", "111.229.160.27")
	publicPort := getEnvAsIntOrDefault("PUBLIC_CDN_PORT", 23357)
	return "http://" + publicIP + ":" + strconv.Itoa(publicPort) + "/cdn/"
}

// IsDev 判断是否为开发环境
func IsDev() bool {
	return AppConfig.AppEnv == "development"
}

// IsProd 判断是否为生产环境
func IsProd() bool {
	return AppConfig.AppEnv == "production"
}

// GetCDNBaseURL 获取CDN基础URL
func GetCDNBaseURL() string {
	// 动态获取CDN基础URL，以便环境变化时能正确返回
	return getCDNBaseURL()
}

// GetResourcePath 获取资源文件存储路径
func GetResourcePath() string {
	return AppConfig.ResourcePath
}

// GetAPIDocsBaseURL 获取API文档基础URL
func GetAPIDocsBaseURL() string {
	// 如果设置了API_DOCS_BASE_URL环境变量，则优先使用该值
	if customURL := os.Getenv("API_DOCS_BASE_URL"); customURL != "" {
		return customURL
	}

	// 根据环境和配置动态生成API文档URL
	env := getEnvOrDefault("APP_ENV", "development")
	if env == "development" {
		// 开发环境使用内网配置
		localIP := getEnvOrDefault("LOCAL_IP", "192.168.3.39")
		return "http://" + localIP
	}

	// 生产环境使用公网配置
	publicIP := getEnvOrDefault("PUBLIC_IP", "111.229.160.27")
	return "http://" + publicIP
}

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

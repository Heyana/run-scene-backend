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

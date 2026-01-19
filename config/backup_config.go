// Package config 备份配置管理
package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// BackupConfig 备份配置结构
type BackupConfig struct {
	Enabled       bool   `yaml:"enabled"`        // 是否启用备份
	LocalPath     string `yaml:"local_path"`     // 本地备份路径
	RetentionDays int    `yaml:"retention_days"` // 本地保留天数
	Environment   string `yaml:"environment"`    // 环境标识: "dev" | "build"
	AutoCleanup   bool   `yaml:"auto_cleanup"`   // 是否启用自动清理
}

// COSConfig 腾讯云COS配置
type COSConfig struct {
	Enabled   bool   `yaml:"enabled"`    // 是否启用COS上传
	SecretID  string `yaml:"secret_id"`  // 腾讯云SecretID
	SecretKey string `yaml:"secret_key"` // 腾讯云SecretKey
	BucketURL string `yaml:"bucket_url"` // COS存储桶URL
	Region    string `yaml:"region"`     // COS区域
}

// BackupFullConfig 完整备份配置（包含COS）
type BackupFullConfig struct {
	Backup BackupConfig `yaml:"backup"`
	COS    COSConfig    `yaml:"cos"`
}

// LoadBackupConfig 加载备份配置
// 现在从主配置文件(config.yaml)加载，优先级: 环境变量 > YAML配置文件 > 默认值
func LoadBackupConfig() (*BackupConfig, *COSConfig) {
	// 从主配置系统加载YAML配置
	yamlConfig := loadYAMLConfigForBackup()

	// 使用YAML配置初始化，支持环境变量覆盖
	backup := &BackupConfig{
		Enabled:       getBoolEnvOrDefault("BACKUP_ENABLED", yamlConfig.Backup.Enabled),
		LocalPath:     getEnvOrDefault("BACKUP_LOCAL_PATH", yamlConfig.Backup.LocalPath),
		RetentionDays: getEnvAsIntOrDefault("BACKUP_RETENTION_DAYS", yamlConfig.Backup.RetentionDays),
		Environment:   getEnvOrDefault("APP_ENV", yamlConfig.Backup.Environment),
		AutoCleanup:   getBoolEnvOrDefault("BACKUP_AUTO_CLEANUP", yamlConfig.Backup.AutoCleanup),
	}

	cos := &COSConfig{
		Enabled:   getBoolEnvOrDefault("COS_ENABLED", yamlConfig.COS.Enabled),
		SecretID:  getEnvOrDefault("COS_SECRET_ID", yamlConfig.COS.SecretID),
		SecretKey: getEnvOrDefault("COS_SECRET_KEY", yamlConfig.COS.SecretKey),
		BucketURL: getEnvOrDefault("COS_BUCKET_URL", yamlConfig.COS.BucketURL),
		Region:    getEnvOrDefault("COS_REGION", yamlConfig.COS.Region),
	}

	return backup, cos
}

// loadYAMLConfigForBackup 从主配置文件加载备份相关配置
func loadYAMLConfigForBackup() *YAMLConfig {
	// 默认配置
	defaultConfig := &YAMLConfig{}
	defaultConfig.Backup.Enabled = true
	defaultConfig.Backup.LocalPath = "./backups"
	defaultConfig.Backup.RetentionDays = 7
	defaultConfig.Backup.Environment = "dev"
	defaultConfig.Backup.AutoCleanup = false
	defaultConfig.COS.Enabled = false
	defaultConfig.COS.Region = "ap-shanghai"

	// 尝试读取主配置文件
	configFile := "config.yaml"
	if _, err := os.Stat(configFile); err == nil {
		data, err := os.ReadFile(configFile)
		if err == nil {
			var yamlConfig YAMLConfig
			if err := yaml.Unmarshal(data, &yamlConfig); err == nil {
				// 备份配置
				if yamlConfig.Backup.LocalPath != "" {
					defaultConfig.Backup = yamlConfig.Backup
				}
				// COS配置
				if yamlConfig.COS.Region != "" {
					defaultConfig.COS = yamlConfig.COS
				}
			}
		}
	}

	return defaultConfig
}

// getBoolEnvOrDefault 获取布尔类型环境变量，如果不存在则返回默认值
func getBoolEnvOrDefault(key string, defaultValue bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		return value == "true"
	}
	return defaultValue
}

// loadFromYAML 从YAML文件加载配置
func loadFromYAML() (*BackupConfig, *COSConfig) {
	// 默认配置
	backup := &BackupConfig{
		Enabled:       true,
		LocalPath:     "./backups",
		RetentionDays: 7,
		Environment:   "dev",
		AutoCleanup:   false, // 默认禁用自动清理
	}

	cos := &COSConfig{
		Enabled: false,
		Region:  "ap-shanghai",
	}

	// 尝试读取YAML配置文件
	configFile := "config.backup.yaml"
	if _, err := os.Stat(configFile); err == nil {
		data, err := os.ReadFile(configFile)
		if err == nil {
			var fullConfig BackupFullConfig
			if err := yaml.Unmarshal(data, &fullConfig); err == nil {
				// 使用YAML配置覆盖默认值
				if fullConfig.Backup.Enabled {
					backup.Enabled = fullConfig.Backup.Enabled
				}

				if fullConfig.Backup.LocalPath != "" {
					backup.LocalPath = fullConfig.Backup.LocalPath
				}
				if fullConfig.Backup.RetentionDays > 0 {
					backup.RetentionDays = fullConfig.Backup.RetentionDays
				}
				if fullConfig.Backup.Environment != "" {
					backup.Environment = fullConfig.Backup.Environment
				}
				// 设置自动清理配置
				backup.AutoCleanup = fullConfig.Backup.AutoCleanup

				// COS配置
				cos = &fullConfig.COS
			}
		}
	}

	return backup, cos
}

// GetEnvironmentPrefix 获取环境前缀
func (c *BackupConfig) GetEnvironmentPrefix() string {
	if c.Environment == "build" {
		return "build"
	}
	return "dev"
}

package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// DatabaseVersionConfig 数据库版本配置
type DatabaseVersionConfig struct {
	Version int                      `yaml:"version"`
	History []DatabaseVersionHistory `yaml:"history"`
}

// DatabaseVersionHistory 版本历史记录
type DatabaseVersionHistory struct {
	Version     int      `yaml:"version"`
	Date        string   `yaml:"date"`
	Description string   `yaml:"description"`
	Tasks       []string `yaml:"tasks"`
	Affected    struct {
		Files    int `yaml:"files"`
		Textures int `yaml:"textures"`
	} `yaml:"affected"`
}

// DatabaseVersion 全局数据库版本配置
var DatabaseVersion *DatabaseVersionConfig

// LoadDatabaseVersionConfig 加载数据库版本配置
func LoadDatabaseVersionConfig(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var config DatabaseVersionConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return err
	}

	DatabaseVersion = &config
	return nil
}

// GetTargetVersion 获取目标版本号
func (c *DatabaseVersionConfig) GetTargetVersion() int {
	if c == nil {
		return 0
	}
	return c.Version
}

// GetVersionHistory 获取指定版本的历史记录
func (c *DatabaseVersionConfig) GetVersionHistory(version int) *DatabaseVersionHistory {
	if c == nil || c.History == nil {
		return nil
	}

	for _, history := range c.History {
		if history.Version == version {
			return &history
		}
	}
	return nil
}

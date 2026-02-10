// Package config 文件处理器配置
package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// FileProcessorConfig 文件处理器配置
type FileProcessorConfig struct {
	FFmpeg struct {
		BinPath string `yaml:"bin_path"`
		Timeout int    `yaml:"timeout"`
	} `yaml:"ffmpeg"`

	ImageMagick struct {
		BinPath string `yaml:"bin_path"`
		Timeout int    `yaml:"timeout"`
	} `yaml:"imagemagick"`

	PDF struct {
		BinPath string `yaml:"bin_path"`
		Timeout int    `yaml:"timeout"`
	} `yaml:"pdf"`

	Blender struct {
		BinPath    string `yaml:"bin_path"`
		ScriptPath string `yaml:"script_path"`
		Timeout    int    `yaml:"timeout"`
	} `yaml:"blender"`

	Thumbnail struct {
		Format  string `yaml:"format"`
		Width   int    `yaml:"width"`
		Height  int    `yaml:"height"`
		Quality int    `yaml:"quality"`
	} `yaml:"thumbnail"`

	Task struct {
		MaxConcurrent int `yaml:"max_concurrent"`
		MaxRetries    int `yaml:"max_retries"`
		RetryDelay    int `yaml:"retry_delay"`
		CleanupAfter  int `yaml:"cleanup_after"`
	} `yaml:"task"`

	Resource struct {
		MaxMemoryPerTask int64   `yaml:"max_memory_per_task"`
		MaxCPUPercent    float64 `yaml:"max_cpu_percent"`
		MaxTempSize      int64   `yaml:"max_temp_size"`
	} `yaml:"resource"`
}

// LoadFileProcessorConfig 加载文件处理器配置
func LoadFileProcessorConfig() (*FileProcessorConfig, error) {
	// 默认配置
	config := &FileProcessorConfig{}
	config.FFmpeg.BinPath = "ffmpeg"
	config.FFmpeg.Timeout = 300
	config.ImageMagick.BinPath = "convert"
	config.ImageMagick.Timeout = 60
	config.PDF.BinPath = "pdftoppm"
	config.PDF.Timeout = 120
	config.Blender.BinPath = "blender"
	config.Blender.ScriptPath = "deploy/scripts/render_fbx.py"
	config.Blender.Timeout = 300
	config.Thumbnail.Format = "webp"
	config.Thumbnail.Width = 1280
	config.Thumbnail.Height = 720
	config.Thumbnail.Quality = 85
	config.Task.MaxConcurrent = 5
	config.Task.MaxRetries = 3
	config.Task.RetryDelay = 60
	config.Task.CleanupAfter = 86400
	config.Resource.MaxMemoryPerTask = 2147483648
	config.Resource.MaxCPUPercent = 80.0
	config.Resource.MaxTempSize = 10737418240

	// 尝试从文件加载
	configFile := "configs/fileprocessor.yaml"
	if _, err := os.Stat(configFile); err == nil {
		data, err := os.ReadFile(configFile)
		if err == nil {
			var yamlConfig struct {
				FileProcessor FileProcessorConfig `yaml:"fileprocessor"`
			}
			if err := yaml.Unmarshal(data, &yamlConfig); err == nil {
				config = &yamlConfig.FileProcessor
			}
		}
	}

	return config, nil
}

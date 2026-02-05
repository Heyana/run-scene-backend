package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// AIConfig AI配置
type AIConfig struct {
	Provider string `yaml:"provider"` // "deepseek" | "openai"
	
	DeepSeek DeepSeekConfig `yaml:"deepseek"`
	OpenAI   OpenAIConfig   `yaml:"openai"`
	
	// 通用配置
	MaxRetries    int `yaml:"max_retries"`
	RetryInterval int `yaml:"retry_interval"` // 秒
	Timeout       int `yaml:"timeout"`        // 秒
}

// DeepSeekConfig DeepSeek配置
type DeepSeekConfig struct {
	APIKey      string  `yaml:"api_key"`
	Model       string  `yaml:"model"`
	BaseURL     string  `yaml:"base_url"`
	Temperature float64 `yaml:"temperature"`
	MaxTokens   int     `yaml:"max_tokens"`
}

// OpenAIConfig OpenAI配置
type OpenAIConfig struct {
	APIKey      string  `yaml:"api_key"`
	Model       string  `yaml:"model"`
	BaseURL     string  `yaml:"base_url"`
	Temperature float64 `yaml:"temperature"`
	MaxTokens   int     `yaml:"max_tokens"`
}

// YAMLAIConfig YAML中的AI配置结构
type YAMLAIConfig struct {
	AI AIConfig `yaml:"ai"`
}

// AppAIConfig 全局AI配置实例
var AppAIConfig *AIConfig

// LoadAIConfig 加载AI配置
func LoadAIConfig() error {
	// 默认配置
	AppAIConfig = &AIConfig{
		Provider: "deepseek",
		DeepSeek: DeepSeekConfig{
			APIKey:      "",
			Model:       "deepseek-chat",
			BaseURL:     "https://api.deepseek.com/v1",
			Temperature: 0.7,
			MaxTokens:   4000,
		},
		OpenAI: OpenAIConfig{
			APIKey:      "",
			Model:       "gpt-4",
			BaseURL:     "https://api.openai.com/v1",
			Temperature: 0.7,
			MaxTokens:   4000,
		},
		MaxRetries:    3,
		RetryInterval: 5,
		Timeout:       60,
	}

	// 尝试从YAML文件加载
	configFile := "configs/ai_config.yaml"
	if _, err := os.Stat(configFile); err == nil {
		data, err := os.ReadFile(configFile)
		if err == nil {
			var yamlConfig YAMLAIConfig
			if err := yaml.Unmarshal(data, &yamlConfig); err == nil {
				// 用YAML配置覆盖默认值
				if yamlConfig.AI.Provider != "" {
					AppAIConfig.Provider = yamlConfig.AI.Provider
				}
				
				// DeepSeek配置
				if yamlConfig.AI.DeepSeek.APIKey != "" {
					AppAIConfig.DeepSeek.APIKey = yamlConfig.AI.DeepSeek.APIKey
				}
				if yamlConfig.AI.DeepSeek.Model != "" {
					AppAIConfig.DeepSeek.Model = yamlConfig.AI.DeepSeek.Model
				}
				if yamlConfig.AI.DeepSeek.BaseURL != "" {
					AppAIConfig.DeepSeek.BaseURL = yamlConfig.AI.DeepSeek.BaseURL
				}
				if yamlConfig.AI.DeepSeek.Temperature > 0 {
					AppAIConfig.DeepSeek.Temperature = yamlConfig.AI.DeepSeek.Temperature
				}
				if yamlConfig.AI.DeepSeek.MaxTokens > 0 {
					AppAIConfig.DeepSeek.MaxTokens = yamlConfig.AI.DeepSeek.MaxTokens
				}
				
				// OpenAI配置
				if yamlConfig.AI.OpenAI.APIKey != "" {
					AppAIConfig.OpenAI.APIKey = yamlConfig.AI.OpenAI.APIKey
				}
				if yamlConfig.AI.OpenAI.Model != "" {
					AppAIConfig.OpenAI.Model = yamlConfig.AI.OpenAI.Model
				}
				if yamlConfig.AI.OpenAI.BaseURL != "" {
					AppAIConfig.OpenAI.BaseURL = yamlConfig.AI.OpenAI.BaseURL
				}
				if yamlConfig.AI.OpenAI.Temperature > 0 {
					AppAIConfig.OpenAI.Temperature = yamlConfig.AI.OpenAI.Temperature
				}
				if yamlConfig.AI.OpenAI.MaxTokens > 0 {
					AppAIConfig.OpenAI.MaxTokens = yamlConfig.AI.OpenAI.MaxTokens
				}
				
				// 通用配置
				if yamlConfig.AI.MaxRetries > 0 {
					AppAIConfig.MaxRetries = yamlConfig.AI.MaxRetries
				}
				if yamlConfig.AI.RetryInterval > 0 {
					AppAIConfig.RetryInterval = yamlConfig.AI.RetryInterval
				}
				if yamlConfig.AI.Timeout > 0 {
					AppAIConfig.Timeout = yamlConfig.AI.Timeout
				}
			}
		}
	}

	// 环境变量覆盖
	if apiKey := os.Getenv("DEEPSEEK_API_KEY"); apiKey != "" {
		AppAIConfig.DeepSeek.APIKey = apiKey
	}
	if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" {
		AppAIConfig.OpenAI.APIKey = apiKey
	}
	if provider := os.Getenv("AI_PROVIDER"); provider != "" {
		AppAIConfig.Provider = provider
	}

	return nil
}

// GetCurrentConfig 获取当前使用的AI配置
func (c *AIConfig) GetCurrentConfig() interface{} {
	switch c.Provider {
	case "openai":
		return c.OpenAI
	case "deepseek":
		return c.DeepSeek
	default:
		return c.DeepSeek
	}
}

// GetAPIKey 获取当前使用的API Key
func (c *AIConfig) GetAPIKey() string {
	switch c.Provider {
	case "openai":
		return c.OpenAI.APIKey
	case "deepseek":
		return c.DeepSeek.APIKey
	default:
		return c.DeepSeek.APIKey
	}
}

// GetModel 获取当前使用的模型
func (c *AIConfig) GetModel() string {
	switch c.Provider {
	case "openai":
		return c.OpenAI.Model
	case "deepseek":
		return c.DeepSeek.Model
	default:
		return c.DeepSeek.Model
	}
}

// GetBaseURL 获取当前使用的BaseURL
func (c *AIConfig) GetBaseURL() string {
	switch c.Provider {
	case "openai":
		return c.OpenAI.BaseURL
	case "deepseek":
		return c.DeepSeek.BaseURL
	default:
		return c.DeepSeek.BaseURL
	}
}

// GetTemperature 获取当前使用的Temperature
func (c *AIConfig) GetTemperature() float64 {
	switch c.Provider {
	case "openai":
		return c.OpenAI.Temperature
	case "deepseek":
		return c.DeepSeek.Temperature
	default:
		return c.DeepSeek.Temperature
	}
}

// GetMaxTokens 获取当前使用的MaxTokens
func (c *AIConfig) GetMaxTokens() int {
	switch c.Provider {
	case "openai":
		return c.OpenAI.MaxTokens
	case "deepseek":
		return c.DeepSeek.MaxTokens
	default:
		return c.DeepSeek.MaxTokens
	}
}

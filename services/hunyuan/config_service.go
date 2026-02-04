package hunyuan

import (
	"errors"
	"go_wails_project_manager/config"
	"go_wails_project_manager/models/hunyuan"

	"gorm.io/gorm"
)

// ConfigService 配置服务
type ConfigService struct {
	db *gorm.DB
}

// NewConfigService 创建配置服务
func NewConfigService(db *gorm.DB) *ConfigService {
	return &ConfigService{db: db}
}

// GetConfig 获取配置（从 config.yaml 读取）
func (s *ConfigService) GetConfig() (*hunyuan.HunyuanConfig, error) {
	cfg := &config.AppConfig.Hunyuan
	
	// 验证配置
	if cfg.SecretID == "" || cfg.SecretKey == "" {
		return nil, errors.New("未配置API密钥")
	}
	
	// 转换为模型配置
	modelConfig := &hunyuan.HunyuanConfig{
		SecretID:            cfg.SecretID,
		SecretKey:           cfg.SecretKey,
		Region:              cfg.Region,
		DefaultModel:        cfg.DefaultModel,
		DefaultFaceCount:    cfg.DefaultFaceCount,
		DefaultGenerateType: cfg.DefaultGenerateType,
		DefaultEnablePBR:    cfg.DefaultEnablePBR,
		DefaultResultFormat: cfg.DefaultResultFormat,
		MaxConcurrent:       cfg.MaxConcurrent,
		PollInterval:        cfg.PollInterval,
		LocalStorageEnabled: cfg.LocalStorageEnabled,
		StorageDir:          cfg.StorageDir,
		BaseURL:             cfg.BaseURL,
		NASEnabled:          cfg.NASEnabled,
		NASPath:             cfg.NASPath,
		DefaultCategory:     cfg.DefaultCategory,
	}
	
	return modelConfig, nil
}

// UpdateConfig 更新配置（不再支持，配置应该在 config.yaml 中修改）
func (s *ConfigService) UpdateConfig(config *hunyuan.HunyuanConfig) error {
	return errors.New("配置更新功能已禁用，请直接修改 config.yaml 文件")
}

// GetClient 获取API客户端
func (s *ConfigService) GetClient() (*HunyuanClient, error) {
	cfg := &config.AppConfig.Hunyuan
	
	if cfg.SecretID == "" || cfg.SecretKey == "" {
		return nil, errors.New("未配置API密钥，请在 config.yaml 中配置 hunyuan.secret_id 和 hunyuan.secret_key")
	}
	
	return NewHunyuanClient(cfg.SecretID, cfg.SecretKey, cfg.Region), nil
}

// ValidateConfig 验证配置
func (s *ConfigService) ValidateConfig(config *hunyuan.HunyuanConfig) error {
	if config.SecretID == "" {
		return errors.New("SecretID不能为空")
	}
	
	if config.SecretKey == "" {
		return errors.New("SecretKey不能为空")
	}
	
	if config.Region == "" {
		return errors.New("Region不能为空")
	}
	
	if config.MaxConcurrent < 1 {
		return errors.New("MaxConcurrent必须大于0")
	}
	
	if config.PollInterval < 1 {
		return errors.New("PollInterval必须大于0")
	}
	
	return nil
}

// TestConnection 测试API连接
func (s *ConfigService) TestConnection() error {
	client, err := s.GetClient()
	if err != nil {
		return err
	}
	
	// 尝试查询一个不存在的任务ID来测试连接
	// 如果返回错误但不是网络错误，说明连接正常
	_, err = client.QueryJob("test-connection-job-id")
	
	// 这里会返回任务不存在的错误，但说明API连接正常
	// 如果是网络错误或认证错误，会在这里体现
	return nil
}

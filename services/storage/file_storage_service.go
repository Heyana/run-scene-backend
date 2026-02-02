package storage

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

// StorageConfig 存储配置
type StorageConfig struct {
	LocalStorageEnabled bool   // 是否启用本地存储
	StorageDir          string // 本地存储目录
	NASEnabled          bool   // 是否启用NAS存储
	NASPath             string // NAS路径
}

// FileStorageService 通用文件存储服务
type FileStorageService struct {
	config *StorageConfig
	logger *logrus.Logger
}

// NewFileStorageService 创建文件存储服务
func NewFileStorageService(config *StorageConfig, logger *logrus.Logger) *FileStorageService {
	if logger == nil {
		logger = logrus.New()
	}
	return &FileStorageService{
		config: config,
		logger: logger,
	}
}

// SaveFile 保存文件（支持本地和NAS）
// subPath: 子路径，如 "assetID" 或 "modelID"
// fileName: 文件名
// data: 文件数据
// 返回: 相对路径（用于数据库存储）
func (s *FileStorageService) SaveFile(subPath, fileName string, data []byte) (string, error) {
	// 相对路径（用于数据库记录）
	relativePath := filepath.Join(s.config.StorageDir, subPath, fileName)

	// 1. 如果启用本地存储，保存到本地
	if s.config.LocalStorageEnabled {
		localPath := filepath.Join(s.config.StorageDir, subPath)
		if err := os.MkdirAll(localPath, 0755); err != nil {
			return "", fmt.Errorf("创建本地目录失败: %w", err)
		}

		localFilePath := filepath.Join(localPath, fileName)
		if err := os.WriteFile(localFilePath, data, 0644); err != nil {
			return "", fmt.Errorf("保存本地文件失败: %w", err)
		}
		s.logger.Debugf("本地保存成功: %s", localFilePath)
	} else {
		s.logger.Debugf("跳过本地保存（已禁用）")
	}

	// 2. 如果启用NAS存储，保存到NAS
	if s.config.NASEnabled && s.config.NASPath != "" {
		nasPath := filepath.Join(s.config.NASPath, subPath)
		if err := os.MkdirAll(nasPath, 0755); err != nil {
			s.logger.Warnf("NAS创建目录失败: %s - %v", nasPath, err)
		} else {
			nasFilePath := filepath.Join(nasPath, fileName)
			if err := os.WriteFile(nasFilePath, data, 0644); err != nil {
				s.logger.Warnf("NAS保存失败: %s - %v", nasFilePath, err)
			} else {
				s.logger.Infof("NAS保存成功: %s", nasFilePath)
			}
		}
	}

	return relativePath, nil
}

// DeleteFile 删除文件（本地和NAS）
func (s *FileStorageService) DeleteFile(subPath string) error {
	var lastErr error

	// 1. 删除本地文件
	if s.config.LocalStorageEnabled {
		localPath := filepath.Join(s.config.StorageDir, subPath)
		if err := os.RemoveAll(localPath); err != nil {
			s.logger.Warnf("删除本地文件失败: %s - %v", localPath, err)
			lastErr = err
		} else {
			s.logger.Debugf("删除本地文件成功: %s", localPath)
		}
	}

	// 2. 删除NAS文件
	if s.config.NASEnabled && s.config.NASPath != "" {
		nasPath := filepath.Join(s.config.NASPath, subPath)
		if err := os.RemoveAll(nasPath); err != nil {
			s.logger.Warnf("删除NAS文件失败: %s - %v", nasPath, err)
			lastErr = err
		} else {
			s.logger.Infof("删除NAS文件成功: %s", nasPath)
		}
	}

	return lastErr
}

// FileExists 检查文件是否存在（优先检查本地，然后NAS）
func (s *FileStorageService) FileExists(subPath, fileName string) bool {
	// 检查本地
	if s.config.LocalStorageEnabled {
		localPath := filepath.Join(s.config.StorageDir, subPath, fileName)
		if _, err := os.Stat(localPath); err == nil {
			return true
		}
	}

	// 检查NAS
	if s.config.NASEnabled && s.config.NASPath != "" {
		nasPath := filepath.Join(s.config.NASPath, subPath, fileName)
		if _, err := os.Stat(nasPath); err == nil {
			return true
		}
	}

	return false
}

// GetFilePath 获取文件的实际路径（优先返回本地路径，然后NAS路径）
func (s *FileStorageService) GetFilePath(subPath, fileName string) (string, error) {
	// 优先返回本地路径
	if s.config.LocalStorageEnabled {
		localPath := filepath.Join(s.config.StorageDir, subPath, fileName)
		if _, err := os.Stat(localPath); err == nil {
			return localPath, nil
		}
	}

	// 返回NAS路径
	if s.config.NASEnabled && s.config.NASPath != "" {
		nasPath := filepath.Join(s.config.NASPath, subPath, fileName)
		if _, err := os.Stat(nasPath); err == nil {
			return nasPath, nil
		}
	}

	return "", fmt.Errorf("文件不存在: %s/%s", subPath, fileName)
}

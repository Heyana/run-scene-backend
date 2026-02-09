package storage

import (
	"fmt"
	"io"
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

// SaveFileStream 流式保存文件（支持大文件，避免内存占用）
// subPath: 子路径，如 "assetID" 或 "modelID"
// fileName: 文件名
// reader: 文件数据流
// fileSize: 文件大小（用于日志）
// 返回: 相对路径（用于数据库存储）
func (s *FileStorageService) SaveFileStream(subPath, fileName string, reader io.Reader, fileSize int64) (string, error) {
	s.logger.Infof("开始流式保存文件: subPath=%s, fileName=%s, size=%.2fMB", 
		subPath, fileName, float64(fileSize)/1024/1024)
	
	// 相对路径（用于数据库记录）- 统一使用正斜杠
	relativePath := s.config.StorageDir + "/" + subPath + "/" + fileName
	s.logger.Debugf("相对路径: %s", relativePath)

	// 1. 如果启用本地存储，保存到本地
	if s.config.LocalStorageEnabled {
		localPath := filepath.Join(s.config.StorageDir, subPath)
		s.logger.Debugf("本地存储路径: %s", localPath)
		
		if err := os.MkdirAll(localPath, 0755); err != nil {
			s.logger.Errorf("创建本地目录失败: %s - %v", localPath, err)
			return "", fmt.Errorf("创建本地目录失败: %w", err)
		}

		localFilePath := filepath.Join(localPath, fileName)
		s.logger.Infof("本地文件完整路径: %s", localFilePath)
		
		// 创建目标文件
		dst, err := os.Create(localFilePath)
		if err != nil {
			s.logger.Errorf("创建本地文件失败: %s - %v", localFilePath, err)
			return "", fmt.Errorf("创建本地文件失败: %w", err)
		}
		defer dst.Close()

		// 如果同时启用了NAS，使用TeeReader同时写入两个目标
		if s.config.NASEnabled && s.config.NASPath != "" {
			nasPath := filepath.Join(s.config.NASPath, subPath)
			s.logger.Debugf("NAS存储路径: %s", nasPath)
			
			if err := os.MkdirAll(nasPath, 0755); err != nil {
				s.logger.Warnf("NAS创建目录失败: %s - %v", nasPath, err)
				// NAS失败不影响本地保存，继续
			} else {
				nasFilePath := filepath.Join(nasPath, fileName)
				s.logger.Infof("NAS文件完整路径: %s", nasFilePath)
				
				nasDst, err := os.Create(nasFilePath)
				if err != nil {
					s.logger.Warnf("NAS创建文件失败: %s - %v", nasFilePath, err)
				} else {
					defer nasDst.Close()
					
					// 使用TeeReader同时写入本地和NAS
					teeReader := io.TeeReader(reader, nasDst)
					written, err := io.Copy(dst, teeReader)
					if err != nil {
						os.Remove(localFilePath)
						os.Remove(nasFilePath)
						s.logger.Errorf("保存文件失败: %v", err)
						return "", fmt.Errorf("保存文件失败: %w", err)
					}
					
					s.logger.Infof("本地保存成功: %s (%.2f MB)", localFilePath, float64(written)/1024/1024)
					s.logger.Infof("NAS保存成功: %s (%.2f MB)", nasFilePath, float64(written)/1024/1024)
					return relativePath, nil
				}
			}
		}

		// 只保存到本地（NAS未启用或失败）
		written, err := io.Copy(dst, reader)
		if err != nil {
			os.Remove(localFilePath)
			s.logger.Errorf("保存本地文件失败: %v", err)
			return "", fmt.Errorf("保存本地文件失败: %w", err)
		}

		s.logger.Infof("本地保存成功: %s (%.2f MB)", localFilePath, float64(written)/1024/1024)
	} else if s.config.NASEnabled && s.config.NASPath != "" {
		// 只启用了NAS，不启用本地存储
		nasPath := filepath.Join(s.config.NASPath, subPath)
		s.logger.Debugf("仅NAS存储路径: %s", nasPath)
		
		if err := os.MkdirAll(nasPath, 0755); err != nil {
			s.logger.Errorf("NAS创建目录失败: %s - %v", nasPath, err)
			return "", fmt.Errorf("NAS创建目录失败: %w", err)
		}
		
		nasFilePath := filepath.Join(nasPath, fileName)
		s.logger.Infof("仅NAS文件完整路径: %s", nasFilePath)
		
		dst, err := os.Create(nasFilePath)
		if err != nil {
			s.logger.Errorf("NAS创建文件失败: %s - %v", nasFilePath, err)
			return "", fmt.Errorf("NAS创建文件失败: %w", err)
		}
		defer dst.Close()

		written, err := io.Copy(dst, reader)
		if err != nil {
			os.Remove(nasFilePath)
			s.logger.Errorf("NAS保存失败: %v", err)
			return "", fmt.Errorf("NAS保存失败: %w", err)
		}

		s.logger.Infof("NAS保存成功: %s (%.2f MB)", nasFilePath, float64(written)/1024/1024)
	} else {
		s.logger.Error("本地存储和NAS存储均未启用")
		return "", fmt.Errorf("本地存储和NAS存储均未启用")
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

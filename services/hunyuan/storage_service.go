package hunyuan

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"time"

	"go_wails_project_manager/config"
	"go_wails_project_manager/logger"
	"go_wails_project_manager/models/hunyuan"
	"go_wails_project_manager/services/storage"

	"gorm.io/gorm"
)

// StorageService 存储服务
type StorageService struct {
	db             *gorm.DB
	config         *config.HunyuanConfig
	storageService *storage.FileStorageService
}

// NewStorageService 创建存储服务
func NewStorageService(db *gorm.DB, cfg *config.HunyuanConfig) *StorageService {
	// 转换配置为通用存储配置
	storageConfig := &storage.StorageConfig{
		LocalStorageEnabled: cfg.LocalStorageEnabled,
		StorageDir:          cfg.StorageDir,
		NASEnabled:          cfg.NASEnabled,
		NASPath:             cfg.NASPath,
	}
	
	return &StorageService{
		db:             db,
		config:         cfg,
		storageService: storage.NewFileStorageService(storageConfig, logger.Log),
	}
}

// StorageInfo 存储信息
type StorageInfo struct {
	LocalPath     string
	NASPath       string
	ThumbnailPath string
	FileSize      int64
	FileHash      string
	AccessURL     string
}

// SaveTaskResult 保存任务结果
func (s *StorageService) SaveTaskResult(task *hunyuan.HunyuanTask, files []File3D) (*StorageInfo, error) {
	if len(files) == 0 {
		return nil, fmt.Errorf("没有文件可保存")
	}

	// 找到GLB文件、OBJ文件和预览图
	var glbFile *File3D
	var objFile *File3D
	var previewFile *File3D
	
	for i := range files {
		if files[i].Type == "GLB" {
			glbFile = &files[i]
		}
		if files[i].Type == "OBJ" {
			objFile = &files[i]
		}
		if files[i].PreviewImageURL != "" && previewFile == nil {
			previewFile = &files[i]
		}
	}
	
	if glbFile == nil {
		return nil, fmt.Errorf("未找到GLB文件")
	}

	// 下载GLB文件
	glbData, err := s.downloadFile(glbFile.URL)
	if err != nil {
		return nil, fmt.Errorf("下载GLB文件失败: %w", err)
	}

	// 计算文件哈希
	fileHash := s.calculateHash(glbData)

	// 检查是否已存在
	existing, err := s.fileExists(fileHash)
	if err == nil && existing != nil {
		// 文件已存在，直接返回现有信息
		return &StorageInfo{
			LocalPath:     stringValue(existing.LocalPath),
			NASPath:       stringValue(existing.NASPath),
			ThumbnailPath: stringValue(existing.ThumbnailPath),
			FileSize:      *existing.FileSize,
			FileHash:      fileHash,
		}, nil
	}

	// 生成文件名（使用时间戳和哈希）
	now := time.Now()
	yearMonth := now.Format("2006/02") // 使用两位数月份
	filename := fmt.Sprintf("%s.glb", fileHash[:16])
	thumbnailFilename := fmt.Sprintf("%s.png", fileHash[:16])
	objZipFilename := fmt.Sprintf("%s_obj.zip", fileHash[:16])

	info := &StorageInfo{
		FileSize: int64(len(glbData)),
		FileHash: fileHash,
	}

	// 使用通用存储服务保存GLB文件
	subPath := yearMonth
	relativePath, err := s.storageService.SaveFile(subPath, filename, glbData)
	if err != nil {
		return nil, fmt.Errorf("保存GLB文件失败: %w", err)
	}
	
	// 保存相对路径（格式：static\hunyuan\2026\02\xxx.glb）
	info.LocalPath = relativePath
	info.NASPath = relativePath

	// 下载并保存OBJ ZIP文件（包含贴图）
	if objFile != nil {
		objData, err := s.downloadFile(objFile.URL)
		if err == nil {
			_, err := s.storageService.SaveFile(subPath, objZipFilename, objData)
			if err != nil {
				logger.Log.Warnf("保存OBJ ZIP文件失败: %v", err)
			} else {
				logger.Log.Infof("OBJ ZIP文件已保存（包含贴图）: %s", objZipFilename)
			}
		} else {
			logger.Log.Warnf("下载OBJ ZIP文件失败: %v", err)
		}
	}

	// 下载并保存预览图
	if previewFile != nil {
		thumbnailData, err := s.downloadFile(previewFile.PreviewImageURL)
		if err == nil {
			thumbnailPath, err := s.storageService.SaveFile(subPath, thumbnailFilename, thumbnailData)
			if err == nil {
				info.ThumbnailPath = thumbnailPath
			}
		}
	}

	return info, nil
}

// downloadFile 下载文件
func (s *StorageService) downloadFile(url string) ([]byte, error) {
	client := &http.Client{Timeout: 60 * time.Second}
	
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP错误: %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

// calculateHash 计算文件哈希
func (s *StorageService) calculateHash(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// fileExists 检查文件是否存在
func (s *StorageService) fileExists(hash string) (*hunyuan.HunyuanTask, error) {
	var task hunyuan.HunyuanTask
	err := s.db.Where("file_hash = ? AND status = ?", hash, "DONE").First(&task).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

// DeleteFiles 删除文件
func (s *StorageService) DeleteFiles(task *hunyuan.HunyuanTask) error {
	// 提取子路径（如 2026/02）
	var subPath string
	if task.LocalPath != nil && *task.LocalPath != "" {
		// 从 static\hunyuan\2026\02\xxx.glb 提取 2026\02
		path := *task.LocalPath
		path = filepath.ToSlash(path)
		parts := filepath.SplitList(path)
		if len(parts) >= 3 {
			subPath = filepath.Join(parts[len(parts)-3], parts[len(parts)-2])
		}
	}
	
	if subPath != "" {
		return s.storageService.DeleteFile(subPath)
	}
	
	return nil
}

// GetAccessURL 获取访问URL（已废弃，使用 AfterFind 钩子）
func (s *StorageService) GetAccessURL(task *hunyuan.HunyuanTask) string {
	// 这个方法已经不需要了，URL 由 AfterFind 钩子自动生成
	return ""
}

// stringValue 安全获取字符串指针的值
func stringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

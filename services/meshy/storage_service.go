package meshy

import (
	"crypto/md5"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"go_wails_project_manager/config"
	"go_wails_project_manager/models/meshy"

	"gorm.io/gorm"
)

// StorageService 存储服务
type StorageService struct {
	db     *gorm.DB
	config *config.MeshyConfig
}

// NewStorageService 创建存储服务
func NewStorageService(db *gorm.DB, cfg *config.MeshyConfig) *StorageService {
	return &StorageService{
		db:     db,
		config: cfg,
	}
}

// FileInfo 文件信息
type FileInfo struct {
	LocalPath          string
	NASPath            string
	ThumbnailPath      string
	PreRemeshedPath    string // PreRemeshed模型本地路径
	PreRemeshedNASPath string // PreRemeshed模型NAS路径
	FileSize           int64
	FileHash           string
}

// SaveTaskResult 保存任务结果
func (s *StorageService) SaveTaskResult(task *meshy.MeshyTask, modelURL, thumbnailURL, preRemeshedURL string) (*FileInfo, error) {
	info := &FileInfo{}

	// 下载主模型文件
	modelData, err := s.downloadFile(modelURL)
	if err != nil {
		return nil, fmt.Errorf("下载模型失败: %w", err)
	}

	// 计算文件哈希
	hash := md5.Sum(modelData)
	info.FileHash = fmt.Sprintf("%x", hash)
	info.FileSize = int64(len(modelData))

	// 生成文件名
	filename := fmt.Sprintf("%s_%s.glb", task.TaskID, info.FileHash[:8])

	// 保存主模型到本地
	if s.config.LocalStorageEnabled {
		localPath := filepath.Join(s.config.StorageDir, filename)
		if err := os.MkdirAll(s.config.StorageDir, 0755); err != nil {
			return nil, fmt.Errorf("创建目录失败: %w", err)
		}
		if err := os.WriteFile(localPath, modelData, 0644); err != nil {
			return nil, fmt.Errorf("保存文件失败: %w", err)
		}
		info.LocalPath = localPath
	}

	// 保存主模型到NAS
	if s.config.NASEnabled && s.config.NASPath != "" {
		nasPath := filepath.Join(s.config.NASPath, filename)
		if err := os.MkdirAll(s.config.NASPath, 0755); err == nil {
			if err := os.WriteFile(nasPath, modelData, 0644); err == nil {
				info.NASPath = nasPath
			}
		}
	}

	// 下载PreRemeshed模型（如果有）
	if preRemeshedURL != "" {
		fmt.Printf("开始下载PreRemeshed模型: %s\n", preRemeshedURL)
		preRemeshedData, err := s.downloadFile(preRemeshedURL)
		if err != nil {
			fmt.Printf("下载PreRemeshed模型失败: %v\n", err)
		} else {
			preRemeshedFilename := fmt.Sprintf("%s_%s_pre_remeshed.glb", task.TaskID, info.FileHash[:8])
			fmt.Printf("PreRemeshed模型下载成功，大小: %d bytes，文件名: %s\n", len(preRemeshedData), preRemeshedFilename)
			
			// 保存到本地
			if s.config.LocalStorageEnabled {
				preRemeshedPath := filepath.Join(s.config.StorageDir, preRemeshedFilename)
				if err := os.WriteFile(preRemeshedPath, preRemeshedData, 0644); err == nil {
					info.PreRemeshedPath = preRemeshedPath
					fmt.Printf("PreRemeshed模型已保存到本地: %s\n", preRemeshedPath)
				} else {
					fmt.Printf("保存PreRemeshed模型到本地失败: %v\n", err)
				}
			}
			
			// 保存到NAS
			if s.config.NASEnabled && s.config.NASPath != "" {
				preRemeshedPath := filepath.Join(s.config.NASPath, preRemeshedFilename)
				fmt.Printf("尝试保存PreRemeshed模型到NAS: %s\n", preRemeshedPath)
				if err := os.WriteFile(preRemeshedPath, preRemeshedData, 0644); err == nil {
					info.PreRemeshedNASPath = preRemeshedPath
					fmt.Printf("PreRemeshed模型已保存到NAS: %s\n", preRemeshedPath)
				} else {
					fmt.Printf("保存PreRemeshed模型到NAS失败: %v\n", err)
				}
			}
		}
	}

	// 下载缩略图
	if thumbnailURL != "" {
		fmt.Printf("开始下载缩略图: %s\n", thumbnailURL)
		thumbData, err := s.downloadFile(thumbnailURL)
		if err != nil {
			fmt.Printf("下载缩略图失败: %v\n", err)
		} else {
			thumbFilename := fmt.Sprintf("%s_%s_thumb.png", task.TaskID, info.FileHash[:8])
			fmt.Printf("缩略图下载成功，大小: %d bytes，文件名: %s\n", len(thumbData), thumbFilename)
			
			// 保存到本地
			if s.config.LocalStorageEnabled {
				thumbPath := filepath.Join(s.config.StorageDir, thumbFilename)
				if err := os.WriteFile(thumbPath, thumbData, 0644); err == nil {
					info.ThumbnailPath = thumbPath
					fmt.Printf("缩略图已保存到本地: %s\n", thumbPath)
				} else {
					fmt.Printf("保存缩略图到本地失败: %v\n", err)
				}
			}
			
			// 保存到NAS
			if s.config.NASEnabled && s.config.NASPath != "" {
				thumbPath := filepath.Join(s.config.NASPath, thumbFilename)
				fmt.Printf("尝试保存缩略图到NAS: %s\n", thumbPath)
				if err := os.WriteFile(thumbPath, thumbData, 0644); err == nil {
					// 如果没有本地路径，使用NAS路径
					if info.ThumbnailPath == "" {
						info.ThumbnailPath = thumbPath
					}
					fmt.Printf("缩略图已保存到NAS: %s\n", thumbPath)
				} else {
					fmt.Printf("保存缩略图到NAS失败: %v\n", err)
				}
			}
			
			if info.ThumbnailPath == "" {
				fmt.Printf("警告：缩略图未保存到任何位置\n")
			}
		}
	} else {
		fmt.Printf("没有缩略图URL，跳过下载\n")
	}

	return info, nil
}

// DeleteFiles 删除文件
func (s *StorageService) DeleteFiles(task *meshy.MeshyTask) error {
	// 删除本地文件
	if task.LocalPath != "" {
		os.Remove(task.LocalPath)
	}
	if task.ThumbnailPath != "" {
		os.Remove(task.ThumbnailPath)
	}

	// 删除NAS文件
	if task.NASPath != "" {
		os.Remove(task.NASPath)
	}

	return nil
}

// downloadFile 下载文件
func (s *StorageService) downloadFile(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("下载失败: %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

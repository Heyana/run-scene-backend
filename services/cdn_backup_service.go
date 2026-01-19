// Package services CDN备份服务
package services

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"go_wails_project_manager/config"
	"go_wails_project_manager/logger"
	"go_wails_project_manager/models"

	"gorm.io/gorm"
)

// CDNBackupService CDN备份服务
type CDNBackupService struct {
	backupConfig  *config.BackupConfig
	cosConfig     *config.COSConfig
	db            *gorm.DB
	backupService *BackupService
}

// CDNBackupManifest CDN备份清单
type CDNBackupManifest struct {
	BackupTime  time.Time     `json:"backup_time"`
	TotalFiles  int           `json:"total_files"`
	TotalSize   int64         `json:"total_size"`
	Files       []CDNFileInfo `json:"files"`
	Environment string        `json:"environment"`
}

// CDNFileInfo CDN文件信息
type CDNFileInfo struct {
	RelativePath string    `json:"relative_path"`
	Size         int64     `json:"size"`
	ModTime      time.Time `json:"mod_time"`
	MD5          string    `json:"md5,omitempty"`
}

// NewCDNBackupService 创建CDN备份服务实例
func NewCDNBackupService(backupConfig *config.BackupConfig, cosConfig *config.COSConfig, db *gorm.DB, backupService *BackupService) *CDNBackupService {
	return &CDNBackupService{
		backupConfig:  backupConfig,
		cosConfig:     cosConfig,
		db:            db,
		backupService: backupService,
	}
}

// BackupCDN 执行CDN增量备份
func (s *CDNBackupService) BackupCDN(ctx context.Context) (*models.BackupRecord, error) {
	startTime := time.Now()

	record := &models.BackupRecord{
		BackupType:  "cdn",
		BackupTime:  startTime,
		Status:      "uploading",
		Environment: s.backupConfig.Environment,
	}

	logger.Log.Info("开始CDN增量备份...")

	// 1. 获取上次备份时间
	lastBackupTime := s.getLastCDNBackupTime()
	logger.Log.Infof("上次CDN备份时间: %s", lastBackupTime.Format("2006-01-02 15:04:05"))

	// 2. 创建备份目录
	timestamp := startTime.Format("20060102_150405")
	backupDir := filepath.Join(s.backupConfig.LocalPath, "cdn", timestamp)
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return nil, fmt.Errorf("创建备份目录失败: %w", err)
	}

	// 3. 扫描CDN目录，查找变更文件
	cdnPath := filepath.Join("static", "cdn") // 使用项目中的CDN路径
	changedFiles, err := s.scanChangedFiles(cdnPath, lastBackupTime)
	if err != nil {
		record.Status = "failed"
		record.ErrorMessage = fmt.Sprintf("扫描CDN目录失败: %v", err)
		s.db.Create(record)
		return record, err
	}

	logger.Log.Infof("发现 %d 个变更文件", len(changedFiles))

	if len(changedFiles) == 0 {
		logger.Log.Info("没有文件变更，跳过备份")
		record.Status = "success"
		record.FileSize = 0
		record.Duration = int(time.Since(startTime).Seconds())
		s.db.Create(record)
		return record, nil
	}

	// 4. 复制变更文件到备份目录
	totalSize := int64(0)
	// 先复制所有文件到本地
	for _, fileInfo := range changedFiles {
		srcPath := filepath.Join(cdnPath, fileInfo.RelativePath)
		dstPath := filepath.Join(backupDir, fileInfo.RelativePath)

		// 确保目标目录存在
		dstDir := filepath.Dir(dstPath)
		if err := os.MkdirAll(dstDir, 0755); err != nil {
			logger.Log.Warnf("创建目录失败: %s, 错误: %v", dstDir, err)
			continue
		}

		// 复制文件
		if err := copyFile(srcPath, dstPath); err != nil {
			logger.Log.Warnf("复制文件失败: %s, 错误: %v", srcPath, err)
			continue
		}

		totalSize += fileInfo.Size
	}

	// 并发上传到COS（如果启用）- 不阻塞主流程
	if s.cosConfig.Enabled {
		go func() {
			// 使用 worker pool 限制并发数
			const maxWorkers = 5
			sem := make(chan struct{}, maxWorkers)
			var wg sync.WaitGroup

			for _, fileInfo := range changedFiles {
				wg.Add(1)
				sem <- struct{}{} // 获取信号量

				go func(fi CDNFileInfo) {
					defer wg.Done()
					defer func() { <-sem }() // 释放信号量

					dstPath := filepath.Join(backupDir, fi.RelativePath)
					envPrefix := s.backupConfig.GetEnvironmentPrefix()
					// 将 Windows 路径分隔符转换为 URL 路径分隔符
					cosPath := strings.ReplaceAll(fi.RelativePath, "\\", "/")
					remotePath := fmt.Sprintf("%s/cdn/%s/%s", envPrefix, timestamp, cosPath)

					uploadCtx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
					defer cancel()

					if err := s.backupService.UploadToCOS(uploadCtx, dstPath, remotePath); err != nil {
						logger.Log.Warnf("上传到COS失败: %s, 错误: %v", remotePath, err)
					}
				}(fileInfo)
			}

			wg.Wait()
			logger.Log.Infof("CDN备份上传完成，共 %d 个文件", len(changedFiles))
		}()
	}

	// 5. 生成备份清单
	manifest := &CDNBackupManifest{
		BackupTime:  startTime,
		TotalFiles:  len(changedFiles),
		TotalSize:   totalSize,
		Files:       changedFiles,
		Environment: s.backupConfig.Environment,
	}

	manifestPath := filepath.Join(backupDir, "manifest.json")
	if err := s.saveManifest(manifest, manifestPath); err != nil {
		logger.Log.Warnf("保存备份清单失败: %v", err)
	}

	// 6. 更新记录
	record.FilePath = backupDir
	record.FileSize = totalSize
	record.Status = "success"
	record.Duration = int(time.Since(startTime).Seconds())

	// 7. 保存备份记录
	if err := s.db.Create(record).Error; err != nil {
		logger.Log.Errorf("保存备份记录失败: %v", err)
	}

	logger.Log.Infof("CDN备份完成，共备份 %d 个文件，总大小: %d 字节，耗时: %d 秒",
		len(changedFiles), totalSize, record.Duration)

	return record, nil
}

// scanChangedFiles 扫描变更的文件
func (s *CDNBackupService) scanChangedFiles(cdnPath string, since time.Time) ([]CDNFileInfo, error) {
	var changedFiles []CDNFileInfo

	err := filepath.Walk(cdnPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过目录
		if info.IsDir() {
			return nil
		}

		// 只备份修改时间晚于上次备份的文件
		if info.ModTime().After(since) {
			// 获取相对路径
			relPath, err := filepath.Rel(cdnPath, path)
			if err != nil {
				return err
			}

			changedFiles = append(changedFiles, CDNFileInfo{
				RelativePath: relPath,
				Size:         info.Size(),
				ModTime:      info.ModTime(),
			})
		}

		return nil
	})

	return changedFiles, err
}

// getLastCDNBackupTime 获取上次CDN备份时间
func (s *CDNBackupService) getLastCDNBackupTime() time.Time {
	var record models.BackupRecord

	err := s.db.Where("backup_type = ? AND status = ?", "cdn", "success").
		Order("backup_time DESC").
		First(&record).Error

	if err != nil {
		// 如果没有备份记录，返回30天前
		return time.Now().AddDate(0, 0, -30)
	}

	return record.BackupTime
}

// saveManifest 保存备份清单
func (s *CDNBackupService) saveManifest(manifest *CDNBackupManifest, path string) error {
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// CleanupOldCDNBackups 清理过期的CDN备份
func (s *CDNBackupService) CleanupOldCDNBackups() error {
	logger.Log.Info("开始清理过期CDN备份...")

	cutoffTime := time.Now().AddDate(0, 0, -s.backupConfig.RetentionDays)
	backupDir := filepath.Join(s.backupConfig.LocalPath, "cdn")

	entries, err := os.ReadDir(backupDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("读取CDN备份目录失败: %w", err)
	}

	deletedCount := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		fullPath := filepath.Join(backupDir, entry.Name())
		info, err := entry.Info()
		if err != nil {
			continue
		}

		// 删除超过保留期限的目录
		if info.ModTime().Before(cutoffTime) {
			if err := os.RemoveAll(fullPath); err != nil {
				logger.Log.Warnf("删除过期CDN备份失败: %s, 错误: %v", fullPath, err)
			} else {
				deletedCount++
				logger.Log.Infof("已删除过期CDN备份: %s", fullPath)
			}
		}
	}

	logger.Log.Infof("CDN备份清理完成，共删除 %d 个过期备份目录", deletedCount)
	return nil
}

// RestoreCDNFromBackup 从备份恢复CDN文件
func (s *CDNBackupService) RestoreCDNFromBackup(backupID uint) error {
	var record models.BackupRecord
	if err := s.db.First(&record, backupID).Error; err != nil {
		return fmt.Errorf("备份记录不存在: %w", err)
	}

	if record.BackupType != "cdn" {
		return fmt.Errorf("不是CDN备份记录")
	}

	backupDir := record.FilePath
	cdnPath := filepath.Join("static", "cdn")

	logger.Log.Infof("开始从备份恢复CDN: %s", backupDir)

	// 读取备份清单
	manifestPath := filepath.Join(backupDir, "manifest.json")
	manifestData, err := os.ReadFile(manifestPath)
	if err != nil {
		return fmt.Errorf("读取备份清单失败: %w", err)
	}

	var manifest CDNBackupManifest
	if err := json.Unmarshal(manifestData, &manifest); err != nil {
		return fmt.Errorf("解析备份清单失败: %w", err)
	}

	// 恢复文件
	for _, fileInfo := range manifest.Files {
		srcPath := filepath.Join(backupDir, fileInfo.RelativePath)
		dstPath := filepath.Join(cdnPath, fileInfo.RelativePath)

		// 确保目标目录存在
		dstDir := filepath.Dir(dstPath)
		if err := os.MkdirAll(dstDir, 0755); err != nil {
			logger.Log.Warnf("创建目录失败: %s, 错误: %v", dstDir, err)
			continue
		}

		// 复制文件
		if err := copyFile(srcPath, dstPath); err != nil {
			logger.Log.Warnf("恢复文件失败: %s, 错误: %v", srcPath, err)
			continue
		}

		logger.Log.Debugf("已恢复文件: %s", fileInfo.RelativePath)
	}

	logger.Log.Infof("CDN恢复完成，共恢复 %d 个文件", len(manifest.Files))
	return nil
}

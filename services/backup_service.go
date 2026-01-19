// Package services 备份服务实现
package services

import (
	"compress/gzip"
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"go_wails_project_manager/config"
	"go_wails_project_manager/logger"
	"go_wails_project_manager/models"

	// "github.com/tencentyun/cos-go-sdk-v5" // TODO: 添加腾讯云COS依赖
	"github.com/tencentyun/cos-go-sdk-v5"
	"gorm.io/gorm"
)

// BackupService 备份服务
type BackupService struct {
	backupConfig *config.BackupConfig
	cosConfig    *config.COSConfig
	db           *gorm.DB
}

// NewBackupService 创建备份服务实例
func NewBackupService(backupConfig *config.BackupConfig, cosConfig *config.COSConfig, db *gorm.DB) *BackupService {
	return &BackupService{
		backupConfig: backupConfig,
		cosConfig:    cosConfig,
		db:           db,
	}
}

// BackupDatabase 备份数据库（直接复制文件）
func (s *BackupService) BackupDatabase(ctx context.Context) (*models.BackupRecord, error) {
	startTime := time.Now()

	record := &models.BackupRecord{
		BackupType:  "database",
		BackupTime:  startTime,
		Status:      "uploading",
		Environment: s.backupConfig.Environment,
	}

	logger.Log.Info("开始数据库备份...")

	// 1. 确保备份目录存在
	backupDir := filepath.Join(s.backupConfig.LocalPath, "database")
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return nil, fmt.Errorf("创建备份目录失败: %w", err)
	}

	// 2. 生成备份文件名
	timestamp := startTime.Format("20060102_150405")
	dbFileName := fmt.Sprintf("app_%s.db", timestamp)
	gzFileName := fmt.Sprintf("app_%s.db.gz", timestamp)

	localDBPath := filepath.Join(backupDir, dbFileName)
	localGZPath := filepath.Join(backupDir, gzFileName)

	// 3. 复制数据库文件
	sourceDB := config.AppConfig.DBPath
	if err := copyFile(sourceDB, localDBPath); err != nil {
		record.Status = "failed"
		record.ErrorMessage = fmt.Sprintf("复制数据库文件失败: %v", err)
		s.db.Create(record)
		return record, err
	}

	logger.Log.Infof("数据库文件复制成功: %s", localDBPath)

	// 4. 压缩备份文件
	if err := gzipFile(localDBPath, localGZPath); err != nil {
		record.Status = "failed"
		record.ErrorMessage = fmt.Sprintf("压缩文件失败: %v", err)
		s.db.Create(record)
		return record, err
	}

	// 删除未压缩的文件，节省空间
	os.Remove(localDBPath)

	logger.Log.Infof("数据库文件压缩成功: %s", localGZPath)

	// 5. 计算MD5
	md5Hash, err := calculateMD5(localGZPath)
	if err != nil {
		logger.Log.Warnf("计算MD5失败: %v", err)
	}

	// 6. 获取文件大小
	fileInfo, _ := os.Stat(localGZPath)
	fileSize := fileInfo.Size()

	// 7. 更新记录信息
	record.FilePath = localGZPath
	record.MD5Hash = md5Hash
	record.FileSize = fileSize
	record.Duration = int(time.Since(startTime).Seconds())

	// 8. 上传到腾讯云COS（如果启用）
	if s.cosConfig.Enabled {
		envPrefix := s.backupConfig.GetEnvironmentPrefix()
		remotePath := fmt.Sprintf("%s/database/%s", envPrefix, gzFileName)

		if err := s.UploadToCOS(ctx, localGZPath, remotePath); err != nil {
			logger.Log.Errorf("上传到COS失败: %v", err)
			record.ErrorMessage = fmt.Sprintf("COS上传失败: %v", err)
			// 不标记为失败，因为本地备份成功
		} else {
			record.RemotePath = remotePath
			logger.Log.Infof("数据库已上传到COS: %s", remotePath)
		}
	}

	// 9. 标记为成功
	record.Status = "success"

	// 10. 保存备份记录
	if err := s.db.Create(record).Error; err != nil {
		logger.Log.Errorf("保存备份记录失败: %v", err)
	}

	logger.Log.Infof("数据库备份完成，耗时: %d秒，大小: %d字节", record.Duration, record.FileSize)

	return record, nil
}

// CleanupOldBackups 清理过期的本地备份
func (s *BackupService) CleanupOldBackups() error {
	logger.Log.Info("开始清理过期备份...")

	cutoffTime := time.Now().AddDate(0, 0, -s.backupConfig.RetentionDays)
	backupDir := filepath.Join(s.backupConfig.LocalPath, "database")

	// 遍历备份目录
	entries, err := os.ReadDir(backupDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // 目录不存在，无需清理
		}
		return fmt.Errorf("读取备份目录失败: %w", err)
	}

	deletedCount := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		fullPath := filepath.Join(backupDir, entry.Name())
		info, err := entry.Info()
		if err != nil {
			continue
		}

		// 删除超过保留期限的文件
		if info.ModTime().Before(cutoffTime) {
			if err := os.Remove(fullPath); err != nil {
				logger.Log.Warnf("删除过期备份失败: %s, 错误: %v", fullPath, err)
			} else {
				deletedCount++
				logger.Log.Infof("已删除过期备份: %s", fullPath)
			}
		}
	}

	logger.Log.Infof("清理完成，共删除 %d 个过期备份文件", deletedCount)
	return nil
}

// UploadToCOS 上传文件到腾讯云COS（使用官方SDK）
func (s *BackupService) UploadToCOS(ctx context.Context, localPath, remotePath string) error {
	if !s.cosConfig.Enabled {
		return fmt.Errorf("COS未启用")
	}

	// 解析存储桶URL
	bucketURL, err := url.Parse(s.cosConfig.BucketURL)
	if err != nil {
		return fmt.Errorf("解析COS URL失败: %w", err)
	}

	// 创建COS客户端
	client := cos.NewClient(&cos.BaseURL{BucketURL: bucketURL}, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  s.cosConfig.SecretID,
			SecretKey: s.cosConfig.SecretKey,
		},
	})

	// 打开本地文件
	file, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("打开文件失败: %w", err)
	}
	defer file.Close()

	// 上传文件
	_, err = client.Object.Put(ctx, remotePath, file, nil)
	if err != nil {
		return fmt.Errorf("上传到COS失败: %w", err)
	}

	logger.Log.Infof("文件已上传到COS: %s", remotePath)
	return nil
}

// GetBackupHistory 获取备份历史记录
func (s *BackupService) GetBackupHistory(page, limit int) ([]models.BackupRecord, int64, error) {
	var records []models.BackupRecord
	var total int64

	offset := (page - 1) * limit

	// 查询总数
	if err := s.db.Model(&models.BackupRecord{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 查询记录
	if err := s.db.Order("backup_time DESC").
		Offset(offset).
		Limit(limit).
		Find(&records).Error; err != nil {
		return nil, 0, err
	}

	return records, total, nil
}

// GetBackupStatus 获取最近的备份状态
func (s *BackupService) GetBackupStatus() (*models.BackupRecord, error) {
	var record models.BackupRecord

	err := s.db.Order("backup_time DESC").First(&record).Error
	if err != nil {
		return nil, err
	}

	return &record, nil
}

// copyFile 复制文件
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

// gzipFile 压缩文件
func gzipFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	gzWriter := gzip.NewWriter(destFile)
	defer gzWriter.Close()

	_, err = io.Copy(gzWriter, sourceFile)
	return err
}

// calculateMD5 计算文件MD5
func calculateMD5(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// GetCOSUploadURL 生成COS上传URL（用于前端直传）
func (s *BackupService) GetCOSUploadURL(filename string) (string, error) {
	if !s.cosConfig.Enabled {
		return "", fmt.Errorf("COS未启用")
	}

	envPrefix := s.backupConfig.GetEnvironmentPrefix()
	remotePath := fmt.Sprintf("%s/manual/%s", envPrefix, filename)

	cosURL := fmt.Sprintf("%s/%s", s.cosConfig.BucketURL, remotePath)

	// 返回预签名URL（简化版）
	u, err := url.Parse(cosURL)
	if err != nil {
		return "", err
	}

	return u.String(), nil
}

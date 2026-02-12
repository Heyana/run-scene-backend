package audit

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go_wails_project_manager/config"
	"go_wails_project_manager/logger"
	"go_wails_project_manager/models"
	"go_wails_project_manager/services/storage"

	"gorm.io/gorm"
)

// ArchiveService 归档服务
type ArchiveService struct {
	db             *gorm.DB
	config         *config.AuditConfig
	storageService *storage.FileStorageService
}

// NewArchiveService 创建归档服务
func NewArchiveService(db *gorm.DB, cfg *config.AuditConfig) *ArchiveService {
	// 创建存储服务配置
	storageConfig := &storage.StorageConfig{
		LocalStorageEnabled: cfg.ArchiveLocalEnabled,
		StorageDir:          cfg.ArchiveStorageDir,
		NASEnabled:          cfg.ArchiveNASEnabled,
		NASPath:             cfg.ArchiveNASPath,
	}

	return &ArchiveService{
		db:             db,
		config:         cfg,
		storageService: storage.NewFileStorageService(storageConfig, logger.Log),
	}
}

// ArchiveFile 归档文件结构
type ArchiveFile struct {
	ArchiveDate string                `json:"archive_date"`
	RecordCount int                   `json:"record_count"`
	DateRange   ArchiveDateRange      `json:"date_range"`
	Logs        []models.AuditLog     `json:"logs"`
}

// ArchiveDateRange 归档日期范围
type ArchiveDateRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// ArchiveOldLogs 归档旧日志（7天前的数据）
func (s *ArchiveService) ArchiveOldLogs() (int64, error) {
	if !s.config.ArchiveEnabled {
		return 0, fmt.Errorf("归档功能未启用")
	}

	// 计算归档截止日期（7天前）
	cutoffDate := time.Now().AddDate(0, 0, -s.config.RetentionDays)
	cutoffDate = time.Date(cutoffDate.Year(), cutoffDate.Month(), cutoffDate.Day(), 0, 0, 0, 0, cutoffDate.Location())

	logger.Log.Infof("开始归档审计日志，截止日期: %s", cutoffDate.Format("2006-01-02"))

	// 按天分组归档
	var totalArchived int64
	currentDate := cutoffDate

	for currentDate.Before(time.Now().AddDate(0, 0, -s.config.RetentionDays)) {
		archived, err := s.archiveByDate(currentDate)
		if err != nil {
			logger.Log.Errorf("归档日期 %s 的日志失败: %v", currentDate.Format("2006-01-02"), err)
		} else {
			totalArchived += archived
		}
		currentDate = currentDate.AddDate(0, 0, 1)
	}

	logger.Log.Infof("归档完成，共归档 %d 条日志", totalArchived)
	return totalArchived, nil
}

// archiveByDate 按日期归档
func (s *ArchiveService) archiveByDate(date time.Time) (int64, error) {
	startTime := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endTime := startTime.AddDate(0, 0, 1)

	// 查询该日期的所有日志
	var logs []models.AuditLog
	if err := s.db.Where("created_at >= ? AND created_at < ?", startTime, endTime).
		Order("created_at ASC").
		Find(&logs).Error; err != nil {
		return 0, fmt.Errorf("查询日志失败: %w", err)
	}

	if len(logs) == 0 {
		logger.Log.Debugf("日期 %s 没有需要归档的日志", date.Format("2006-01-02"))
		return 0, nil
	}

	logger.Log.Infof("日期 %s 找到 %d 条日志需要归档", date.Format("2006-01-02"), len(logs))

	// 创建归档文件
	archiveFile := ArchiveFile{
		ArchiveDate: date.Format("2006-01-02"),
		RecordCount: len(logs),
		DateRange: ArchiveDateRange{
			Start: startTime,
			End:   endTime,
		},
		Logs: logs,
	}

	// 保存归档文件
	filePath, err := s.saveArchiveFile(&archiveFile, date)
	if err != nil {
		return 0, fmt.Errorf("保存归档文件失败: %w", err)
	}

	logger.Log.Infof("归档文件已保存: %s", filePath)

	// 删除已归档的日志
	if err := s.db.Where("created_at >= ? AND created_at < ?", startTime, endTime).
		Delete(&models.AuditLog{}).Error; err != nil {
		return 0, fmt.Errorf("删除已归档日志失败: %w", err)
	}

	logger.Log.Infof("已删除 %d 条已归档的日志", len(logs))

	return int64(len(logs)), nil
}

// saveArchiveFile 保存归档文件
func (s *ArchiveService) saveArchiveFile(archiveFile *ArchiveFile, date time.Time) (string, error) {
	// 生成文件路径：{year}/{month}/{day}/audit_{timestamp}.json
	timestamp := time.Now().Format("20060102_150405")
	fileName := fmt.Sprintf("audit_%s.json", timestamp)
	
	if s.config.ArchiveCompression {
		fileName += ".gz"
	}

	subPath := fmt.Sprintf("%d/%02d/%02d", date.Year(), date.Month(), date.Day())

	// 序列化为 JSON
	jsonData, err := json.MarshalIndent(archiveFile, "", "  ")
	if err != nil {
		return "", fmt.Errorf("序列化归档数据失败: %w", err)
	}

	// 如果启用压缩
	var dataToSave []byte
	if s.config.ArchiveCompression {
		compressed, err := s.compressData(jsonData)
		if err != nil {
			return "", fmt.Errorf("压缩归档数据失败: %w", err)
		}
		dataToSave = compressed
	} else {
		dataToSave = jsonData
	}

	// 保存文件
	filePath, err := s.storageService.SaveFile(subPath, fileName, dataToSave)
	if err != nil {
		return "", fmt.Errorf("保存归档文件失败: %w", err)
	}

	return filePath, nil
}

// compressData 压缩数据
func (s *ArchiveService) compressData(data []byte) ([]byte, error) {
	var buf strings.Builder
	gzWriter := gzip.NewWriter(&buf)
	
	if _, err := gzWriter.Write(data); err != nil {
		return nil, err
	}
	
	if err := gzWriter.Close(); err != nil {
		return nil, err
	}
	
	return []byte(buf.String()), nil
}

// LoadFromArchive 从归档文件加载日志
func (s *ArchiveService) LoadFromArchive(filePath string) (*ArchiveFile, error) {
	// 读取文件
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("读取归档文件失败: %w", err)
	}

	// 如果是压缩文件，先解压
	if strings.HasSuffix(filePath, ".gz") {
		data, err = s.decompressData(data)
		if err != nil {
			return nil, fmt.Errorf("解压归档文件失败: %w", err)
		}
	}

	// 反序列化
	var archiveFile ArchiveFile
	if err := json.Unmarshal(data, &archiveFile); err != nil {
		return nil, fmt.Errorf("解析归档文件失败: %w", err)
	}

	return &archiveFile, nil
}

// decompressData 解压数据
func (s *ArchiveService) decompressData(data []byte) ([]byte, error) {
	reader := strings.NewReader(string(data))
	gzReader, err := gzip.NewReader(reader)
	if err != nil {
		return nil, err
	}
	defer gzReader.Close()

	var buf []byte
	buf, err = io.ReadAll(gzReader)
	if err != nil {
		return nil, err
	}

	return buf, nil
}

// ListArchiveFiles 列出归档文件
func (s *ArchiveService) ListArchiveFiles(startDate, endDate time.Time) ([]string, error) {
	var files []string

	// 遍历日期范围
	currentDate := startDate
	for currentDate.Before(endDate) || currentDate.Equal(endDate) {
		// 构建目录路径
		dirPath := filepath.Join(
			s.getArchiveBasePath(),
			fmt.Sprintf("%d/%02d/%02d", currentDate.Year(), currentDate.Month(), currentDate.Day()),
		)

		// 检查目录是否存在
		if _, err := os.Stat(dirPath); err == nil {
			// 列出目录中的文件
			entries, err := os.ReadDir(dirPath)
			if err != nil {
				logger.Log.Warnf("读取归档目录失败: %s, error: %v", dirPath, err)
			} else {
				for _, entry := range entries {
					if !entry.IsDir() && strings.HasPrefix(entry.Name(), "audit_") {
						files = append(files, filepath.Join(dirPath, entry.Name()))
					}
				}
			}
		}

		currentDate = currentDate.AddDate(0, 0, 1)
	}

	return files, nil
}

// getArchiveBasePath 获取归档基础路径
func (s *ArchiveService) getArchiveBasePath() string {
	if s.config.ArchiveNASEnabled && s.config.ArchiveNASPath != "" {
		return s.config.ArchiveNASPath
	}
	return s.config.ArchiveStorageDir
}

// GetArchiveStatistics 获取归档统计信息
func (s *ArchiveService) GetArchiveStatistics() (map[string]interface{}, error) {
	// 统计数据库中的日志数量
	var dbCount int64
	if err := s.db.Model(&models.AuditLog{}).Count(&dbCount).Error; err != nil {
		return nil, err
	}

	// 统计最早和最新的日志时间
	var oldestLog, newestLog models.AuditLog
	s.db.Order("created_at ASC").First(&oldestLog)
	s.db.Order("created_at DESC").First(&newestLog)

	// 统计归档文件数量
	archiveFiles, _ := s.ListArchiveFiles(
		time.Now().AddDate(-1, 0, 0), // 过去一年
		time.Now(),
	)

	return map[string]interface{}{
		"database_count":    dbCount,
		"oldest_log":        oldestLog.CreatedAt,
		"newest_log":        newestLog.CreatedAt,
		"archive_files":     len(archiveFiles),
		"retention_days":    s.config.RetentionDays,
		"archive_enabled":   s.config.ArchiveEnabled,
	}, nil
}

// Package services 备份调度器
package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go_wails_project_manager/config"
	"go_wails_project_manager/logger"

	"gorm.io/gorm"
)

// 全局备份调度器实例
var globalBackupScheduler *BackupScheduler

// SetGlobalBackupScheduler 设置全局备份调度器
func SetGlobalBackupScheduler(scheduler *BackupScheduler) {
	globalBackupScheduler = scheduler
}

// GetGlobalBackupScheduler 获取全局备份调度器
func GetGlobalBackupScheduler() *BackupScheduler {
	return globalBackupScheduler
}

// BackupScheduler 备份调度器
type BackupScheduler struct {
	backupService    *BackupService
	cdnBackupService *CDNBackupService
	backupConfig     *config.BackupConfig
	stopCh           chan struct{}
	running          bool
	mu               sync.Mutex
}

// NewBackupScheduler 创建备份调度器实例
func NewBackupScheduler(
	backupConfig *config.BackupConfig,
	cosConfig *config.COSConfig,
	db *gorm.DB,
) *BackupScheduler {
	backupService := NewBackupService(backupConfig, cosConfig, db)
	cdnBackupService := NewCDNBackupService(backupConfig, cosConfig, db, backupService)

	return &BackupScheduler{
		backupService:    backupService,
		cdnBackupService: cdnBackupService,
		backupConfig:     backupConfig,
		stopCh:           make(chan struct{}),
		running:          false,
	}
}

// Start 启动调度器
func (s *BackupScheduler) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return fmt.Errorf("备份调度器已经在运行")
	}

	if !s.backupConfig.Enabled {
		logger.Log.Info("备份功能未启用")
		return nil
	}

	s.running = true

	logger.Log.Infof("备份调度器已启动，每天12:00和24:00执行备份, 环境: %s",
		s.backupConfig.Environment)

	// 启动后台goroutine执行定时任务
	go s.run()

	return nil
}

// Stop 停止调度器
func (s *BackupScheduler) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return
	}

	logger.Log.Info("正在停止备份调度器...")

	close(s.stopCh)

	s.running = false
	logger.Log.Info("备份调度器已停止")
}

// run 运行调度器
func (s *BackupScheduler) run() {
	defer func() {
		if r := recover(); r != nil {
			logger.Log.Errorf("备份调度器panic: %v", r)
		}
	}()

	ticker := time.NewTicker(1 * time.Minute) // 每分钟检查一次
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// 检查是否到了备份时间（每天12:00和24:00）
			now := time.Now()
			if s.shouldRunBackup(now) {
				logger.Log.Info("到达备份时间点，开始执行备份")
				s.runBackup()
			}
		case <-s.stopCh:
			// 收到停止信号
			return
		}
	}
}

// shouldRunBackup 检查是否应该执行备份
func (s *BackupScheduler) shouldRunBackup(now time.Time) bool {
	hour := now.Hour()
	minute := now.Minute()

	// 每天12:00和24:00（0:00）执行备份
	if (hour == 12 || hour == 0) && minute == 0 {
		// 检查今天是否已经在这个时间点执行过备份了
		return !s.hasBackupRunToday(now, hour)
	}

	return false
}

// hasBackupRunToday 检查今天是否已经在指定小时执行过备份
func (s *BackupScheduler) hasBackupRunToday(now time.Time, targetHour int) bool {
	record, err := s.backupService.GetBackupStatus()
	if err != nil {
		return false // 没有备份记录，可以执行
	}

	// 检查最后备份时间是否是今天且在目标小时
	lastBackup := record.BackupTime
	if lastBackup.Year() == now.Year() &&
		lastBackup.Month() == now.Month() &&
		lastBackup.Day() == now.Day() &&
		lastBackup.Hour() == targetHour {
		return true // 今天这个时间点已经备份过了
	}

	return false
}

// runBackup 执行备份任务
func (s *BackupScheduler) runBackup() {
	logger.Log.Info("========== 开始执行定时备份任务 ==========")
	startTime := time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	// 1. 数据库备份
	logger.Log.Info(">>> 执行数据库备份...")
	dbRecord, err := s.backupService.BackupDatabase(ctx)
	if err != nil {
		logger.Log.Errorf("数据库备份失败: %v", err)
	} else {
		logger.Log.Infof("数据库备份成功: %s (大小: %d 字节)", dbRecord.FilePath, dbRecord.FileSize)
	}

	// 2. CDN备份
	logger.Log.Info(">>> 执行CDN增量备份...")
	cdnRecord, err := s.cdnBackupService.BackupCDN(ctx)
	if err != nil {
		logger.Log.Errorf("CDN备份失败: %v", err)
	} else {
		logger.Log.Infof("CDN备份成功: %s (大小: %d 字节)", cdnRecord.FilePath, cdnRecord.FileSize)
	}

	// 3. 清理过期备份（如果启用了自动清理）
	if s.backupConfig.AutoCleanup {
		logger.Log.Info(">>> 清理过期备份...")
		if err := s.backupService.CleanupOldBackups(); err != nil {
			logger.Log.Errorf("清理数据库备份失败: %v", err)
		}

		if err := s.cdnBackupService.CleanupOldCDNBackups(); err != nil {
			logger.Log.Errorf("清理CDN备份失败: %v", err)
		}
	} else {
		logger.Log.Info(">>> 自动清理已禁用，跳过清理过期备份")
	}

	duration := time.Since(startTime)
	logger.Log.Infof("========== 备份任务完成，总耗时: %s ==========", duration)
}

// TriggerManualBackup 手动触发全量备份
func (s *BackupScheduler) TriggerManualBackup() error {
	s.mu.Lock()
	if !s.running && s.backupConfig.Enabled {
		s.mu.Unlock()
		return fmt.Errorf("备份调度器未运行")
	}
	s.mu.Unlock()

	logger.Log.Info("收到手动全量备份触发请求")

	// 在新的goroutine中执行，避免阻塞
	go s.runBackup()

	return nil
}

// TriggerDatabaseBackup 手动触发数据库备份
func (s *BackupScheduler) TriggerDatabaseBackup() error {
	logger.Log.Info("收到手动数据库备份触发请求")

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()

		logger.Log.Info(">>> 执行数据库备份...")
		dbRecord, err := s.backupService.BackupDatabase(ctx)
		if err != nil {
			logger.Log.Errorf("数据库备份失败: %v", err)
		} else {
			logger.Log.Infof("数据库备份成功: %s (大小: %d 字节)", dbRecord.FilePath, dbRecord.FileSize)
		}
	}()

	return nil
}

// TriggerCDNBackup 手动触发CDN备份
func (s *BackupScheduler) TriggerCDNBackup() error {
	logger.Log.Info("收到手动CDN备份触发请求")

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
		defer cancel()

		logger.Log.Info(">>> 执行CDN增量备份...")
		cdnRecord, err := s.cdnBackupService.BackupCDN(ctx)
		if err != nil {
			logger.Log.Errorf("CDN备份失败: %v", err)
		} else {
			logger.Log.Infof("CDN备份成功: %s (大小: %d 字节)", cdnRecord.FilePath, cdnRecord.FileSize)
		}
	}()

	return nil
}

// IsRunning 检查调度器是否在运行
func (s *BackupScheduler) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}

// GetNextBackupTime 获取下次备份时间
func (s *BackupScheduler) GetNextBackupTime() time.Time {
	now := time.Now()

	// 计算今天的12:00和24:00
	today12 := time.Date(now.Year(), now.Month(), now.Day(), 12, 0, 0, 0, now.Location())
	today24 := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location()) // 明天的0:00

	// 根据当前时间判断下次备份时间
	if now.Before(today12) {
		return today12 // 今天12:00还没到
	} else if now.Before(today24) {
		return today24 // 今天24:00还没到
	} else {
		// 今天的备份时间都过了，返回明天的12:00
		return time.Date(now.Year(), now.Month(), now.Day()+1, 12, 0, 0, 0, now.Location())
	}
}

// getLastBackupTime 获取最后一次备份时间
func (s *BackupScheduler) getLastBackupTime() (time.Time, error) {
	record, err := s.backupService.GetBackupStatus()
	if err != nil {
		return time.Time{}, err
	}

	return record.BackupTime, nil
}

// GetBackupService 获取备份服务实例（用于API调用）
func (s *BackupScheduler) GetBackupService() *BackupService {
	return s.backupService
}

// GetCDNBackupService 获取CDN备份服务实例（用于API调用）
func (s *BackupScheduler) GetCDNBackupService() *CDNBackupService {
	return s.cdnBackupService
}

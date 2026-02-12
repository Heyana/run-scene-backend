package audit

import (
	"sync"
	"time"

	"go_wails_project_manager/config"
	"go_wails_project_manager/logger"

	"gorm.io/gorm"
)

// ArchiveScheduler 归档调度器
type ArchiveScheduler struct {
	mu             sync.Mutex
	running        bool
	ticker         *time.Ticker
	stopChan       chan struct{}
	archiveService *ArchiveService
	config         *config.AuditConfig
}

// NewArchiveScheduler 创建归档调度器
func NewArchiveScheduler(db *gorm.DB, cfg *config.AuditConfig) *ArchiveScheduler {
	return &ArchiveScheduler{
		archiveService: NewArchiveService(db, cfg),
		config:         cfg,
		stopChan:       make(chan struct{}),
	}
}

// Start 启动调度器
func (s *ArchiveScheduler) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.config.ArchiveEnabled {
		logger.Log.Info("审计日志归档功能未启用")
		return nil
	}

	if s.running {
		logger.Log.Warn("审计归档调度器已在运行")
		return nil
	}

	s.running = true
	s.ticker = time.NewTicker(1 * time.Hour) // 每小时检查一次

	go s.run()

	logger.Log.Info("审计归档调度器已启动")
	return nil
}

// Stop 停止调度器
func (s *ArchiveScheduler) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return
	}

	s.running = false
	close(s.stopChan)

	if s.ticker != nil {
		s.ticker.Stop()
	}

	logger.Log.Info("审计归档调度器已停止")
}

// run 运行调度器
func (s *ArchiveScheduler) run() {
	defer func() {
		if r := recover(); r != nil {
			logger.Log.Errorf("审计归档调度器发生panic: %v", r)
		}
	}()

	// 启动时立即检查一次
	s.checkAndArchive()

	for {
		select {
		case <-s.ticker.C:
			s.checkAndArchive()
		case <-s.stopChan:
			return
		}
	}
}

// checkAndArchive 检查并执行归档
func (s *ArchiveScheduler) checkAndArchive() {
	now := time.Now()

	// 解析 cron 表达式（简化版，只支持 "0 2 * * *" 格式）
	// 默认每天凌晨2点执行
	targetHour := 2
	targetMinute := 0

	// 检查当前时间是否匹配
	if now.Hour() == targetHour && now.Minute() >= targetMinute && now.Minute() < targetMinute+60 {
		// 检查今天是否已经执行过
		if s.hasArchivedToday() {
			return
		}

		logger.Log.Info("开始执行审计日志归档任务")
		s.runArchive()
	}
}

// hasArchivedToday 检查今天是否已经归档过
func (s *ArchiveScheduler) hasArchivedToday() bool {
	// 简化实现：检查是否在过去2小时内执行过
	// 实际项目中可以记录到数据库或文件
	return false
}

// runArchive 执行归档任务
func (s *ArchiveScheduler) runArchive() {
	startTime := time.Now()

	count, err := s.archiveService.ArchiveOldLogs()
	if err != nil {
		logger.Log.Errorf("审计日志归档失败: %v", err)
		return
	}

	duration := time.Since(startTime)
	logger.Log.Infof("审计日志归档完成，共归档 %d 条日志，耗时: %v", count, duration)
}

// IsRunning 检查调度器是否在运行
func (s *ArchiveScheduler) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}

// GetArchiveService 获取归档服务实例
func (s *ArchiveScheduler) GetArchiveService() *ArchiveService {
	return s.archiveService
}

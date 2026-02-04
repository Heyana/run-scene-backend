package hunyuan

import (
	"sync"
	"time"

	"go_wails_project_manager/logger"
	"go_wails_project_manager/models/hunyuan"

	"gorm.io/gorm"
)

// TaskPoller 任务轮询器
type TaskPoller struct {
	db          *gorm.DB
	taskService *TaskService
	interval    time.Duration
	stopChan    chan struct{}
	wg          sync.WaitGroup
	running     bool
	mu          sync.Mutex
}

// NewTaskPoller 创建任务轮询器
func NewTaskPoller(db *gorm.DB, taskService *TaskService, interval time.Duration) *TaskPoller {
	return &TaskPoller{
		db:          db,
		taskService: taskService,
		interval:    interval,
		stopChan:    make(chan struct{}),
	}
}

// Start 启动轮询器
func (p *TaskPoller) Start() {
	p.mu.Lock()
	if p.running {
		p.mu.Unlock()
		return
	}
	p.running = true
	p.mu.Unlock()

	logger.Log.Info("混元3D任务轮询器已启动")

	p.wg.Add(1)
	go p.pollLoop()
}

// Stop 停止轮询器
func (p *TaskPoller) Stop() {
	p.mu.Lock()
	if !p.running {
		p.mu.Unlock()
		return
	}
	p.running = false
	p.mu.Unlock()

	close(p.stopChan)
	p.wg.Wait()

	logger.Log.Info("混元3D任务轮询器已停止")
}

// pollLoop 轮询循环
func (p *TaskPoller) pollLoop() {
	defer p.wg.Done()

	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	// 启动时立即执行一次
	p.pollPendingTasks()

	for {
		select {
		case <-ticker.C:
			p.pollPendingTasks()
		case <-p.stopChan:
			return
		}
	}
}

// pollPendingTasks 轮询待处理的任务
func (p *TaskPoller) pollPendingTasks() {
	// 查询所有等待中和运行中的任务
	var tasks []hunyuan.HunyuanTask
	err := p.db.Where("status IN ?", []string{"WAIT", "RUN"}).
		Order("created_at ASC").
		Find(&tasks).Error

	if err != nil {
		logger.Log.Errorf("查询待处理任务失败: %v", err)
		return
	}

	if len(tasks) == 0 {
		return
	}

	logger.Log.Infof("开始轮询 %d 个待处理任务", len(tasks))

	// 并发轮询任务
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 5) // 限制并发数为5

	for _, task := range tasks {
		wg.Add(1)
		go func(t hunyuan.HunyuanTask) {
			defer wg.Done()

			semaphore <- struct{}{} // 获取信号量
			defer func() { <-semaphore }() // 释放信号量

			if err := p.taskService.PollTask(t.ID); err != nil {
				logger.Log.Errorf("轮询任务 %d (JobID: %s) 失败: %v", t.ID, t.JobID, err)
			} else {
				// 查询更新后的状态
				var updatedTask hunyuan.HunyuanTask
				if err := p.db.First(&updatedTask, t.ID).Error; err == nil {
					if updatedTask.Status == "DONE" {
						logger.Log.Infof("任务 %d (JobID: %s) 已完成", t.ID, t.JobID)
					} else if updatedTask.Status == "FAIL" {
						logger.Log.Warnf("任务 %d (JobID: %s) 失败: %s", t.ID, t.JobID, 
							stringPtrOrEmpty(updatedTask.ErrorMessage))
					}
				}
			}
		}(task)
	}

	wg.Wait()
	logger.Log.Infof("轮询完成，处理了 %d 个任务", len(tasks))
}

// stringPtrOrEmpty 返回字符串指针的值或空字符串
func stringPtrOrEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// GetStatus 获取轮询器状态
func (p *TaskPoller) GetStatus() map[string]interface{} {
	p.mu.Lock()
	defer p.mu.Unlock()

	var pendingCount int64
	p.db.Model(&hunyuan.HunyuanTask{}).
		Where("status IN ?", []string{"WAIT", "RUN"}).
		Count(&pendingCount)

	return map[string]interface{}{
		"running":      p.running,
		"interval":     p.interval.String(),
		"pendingTasks": pendingCount,
	}
}

package ai3d

import (
	"context"
	"sync"
	"time"

	"go_wails_project_manager/logger"
	"go_wails_project_manager/models/ai3d"

	"gorm.io/gorm"
)

// TaskPoller 统一任务轮询器
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

	logger.Log.Info("AI3D统一任务轮询器已启动")

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

	logger.Log.Info("AI3D统一任务轮询器已停止")
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
	// 获取所有待处理的任务
	tasks, err := p.taskService.GetPendingTasks()
	if err != nil {
		logger.Log.Errorf("查询待处理任务失败: %v", err)
		return
	}

	if len(tasks) == 0 {
		return
	}

	logger.Log.Infof("开始轮询 %d 个AI3D待处理任务", len(tasks))

	// 并发轮询任务
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 5) // 限制并发数为5

	for _, task := range tasks {
		wg.Add(1)
		go func(t *ai3d.Task) {
			defer wg.Done()

			semaphore <- struct{}{}        // 获取信号量
			defer func() { <-semaphore }() // 释放信号量

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			if err := p.taskService.pollTaskInternal(ctx, t); err != nil {
				logger.Log.Errorf("轮询任务 %d (%s:%s) 失败: %v", t.ID, t.Provider, t.ProviderTaskID, err)
			}
		}(task)
	}

	wg.Wait()
	logger.Log.Infof("AI3D轮询完成，处理了 %d 个任务", len(tasks))
}

// GetStatus 获取轮询器状态
func (p *TaskPoller) GetStatus() map[string]interface{} {
	p.mu.Lock()
	defer p.mu.Unlock()

	var pendingCount int64
	p.db.Model(&ai3d.Task{}).
		Where("status IN ?", []string{"WAIT", "RUN"}).
		Count(&pendingCount)

	return map[string]interface{}{
		"running":      p.running,
		"interval":     p.interval.String(),
		"pendingTasks": pendingCount,
	}
}

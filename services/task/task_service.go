// Package task 任务管理服务
package task

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go_wails_project_manager/models"
	"go_wails_project_manager/services/fileprocessor"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// TaskService 任务服务
type TaskService struct {
	db       *gorm.DB
	queue    *TaskQueue
	executor *TaskExecutor
	tasks    map[uint]*TaskContext
	mu       sync.RWMutex
}

// TaskContext 任务上下文
type TaskContext struct {
	Task      *models.Task
	Cancel    context.CancelFunc
	Progress  chan float64
	Message   chan string
	MemoryUsed int64
	CPUUsed    float64
}

// NewTaskService 创建任务服务
func NewTaskService(db *gorm.DB, fpService fileprocessor.IFileProcessorService) *TaskService {
	service := &TaskService{
		db:    db,
		queue: NewTaskQueue(),
		tasks: make(map[uint]*TaskContext),
	}

	// 初始化执行器
	service.executor = NewTaskExecutor(5, fpService) // 最大并发5个任务

	// 启动任务处理协程
	go service.processQueue()

	return service
}

// CreateTask 创建任务
func (s *TaskService) CreateTask(task *models.Task) error {
	// 设置默认值
	if task.Status == "" {
		task.Status = models.TaskStatusPending
	}
	if task.Priority == 0 {
		task.Priority = 5
	}
	if task.MaxRetries == 0 {
		task.MaxRetries = 3
	}
	if task.RetryDelay == 0 {
		task.RetryDelay = 60
	}

	// 保存到数据库
	if err := s.db.Create(task).Error; err != nil {
		return err
	}

	// 加入队列
	s.queue.Push(task)

	logrus.Infof("任务创建成功: ID=%d, Type=%s", task.ID, task.Type)
	return nil
}

// StartTask 启动任务
func (s *TaskService) StartTask(taskID uint) error {
	var task models.Task
	if err := s.db.First(&task, taskID).Error; err != nil {
		return err
	}

	// 检查任务状态
	if task.Status != models.TaskStatusPending {
		return fmt.Errorf("任务状态不正确: %s", task.Status)
	}

	// 加入队列
	s.queue.Push(&task)

	return nil
}

// CancelTask 取消任务
func (s *TaskService) CancelTask(taskID uint) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 查找任务上下文
	taskCtx, exists := s.tasks[taskID]
	if !exists {
		// 任务不在运行中，从队列中移除
		s.queue.Remove(taskID)

		// 更新数据库状态
		return s.db.Model(&models.Task{}).
			Where("id = ?", taskID).
			Updates(map[string]interface{}{
				"status":  models.TaskStatusCancelled,
				"message": "任务已取消",
			}).Error
	}

	// 取消正在运行的任务
	if taskCtx.Cancel != nil {
		taskCtx.Cancel()
	}

	// 更新状态
	return s.db.Model(&models.Task{}).
		Where("id = ?", taskID).
		Updates(map[string]interface{}{
			"status":  models.TaskStatusCancelled,
			"message": "任务已取消",
		}).Error
}

// RetryTask 重试任务
func (s *TaskService) RetryTask(taskID uint) error {
	var task models.Task
	if err := s.db.First(&task, taskID).Error; err != nil {
		return err
	}

	// 检查重试次数
	if task.RetryCount >= task.MaxRetries {
		return errors.New("超过最大重试次数")
	}

	// 增加重试计数
	task.RetryCount++
	task.Status = models.TaskStatusRetrying
	task.Error = ""
	task.Progress = 0

	// 延迟重试
	time.Sleep(time.Duration(task.RetryDelay) * time.Second)

	// 重新加入队列
	s.queue.Push(&task)

	return s.db.Save(&task).Error
}

// GetTask 获取任务
func (s *TaskService) GetTask(taskID uint) (*models.Task, error) {
	var task models.Task
	if err := s.db.First(&task, taskID).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

// GetTaskProgress 获取任务进度
func (s *TaskService) GetTaskProgress(taskID uint) (float64, string, error) {
	var task models.Task
	if err := s.db.First(&task, taskID).Error; err != nil {
		return 0, "", err
	}
	return task.Progress, task.Message, nil
}

// ListTasks 列出任务
func (s *TaskService) ListTasks(filters TaskFilters) ([]*models.Task, int64, error) {
	var tasks []*models.Task
	var total int64

	query := s.db.Model(&models.Task{})

	// 应用过滤器
	if filters.Status != "" {
		query = query.Where("status = ?", filters.Status)
	}
	if filters.Type != "" {
		query = query.Where("type = ?", filters.Type)
	}
	if filters.UserID > 0 {
		query = query.Where("user_id = ?", filters.UserID)
	}

	// 统计总数
	query.Count(&total)

	// 分页
	if filters.Page > 0 && filters.PageSize > 0 {
		offset := (filters.Page - 1) * filters.PageSize
		query = query.Offset(offset).Limit(filters.PageSize)
	}

	// 排序
	query = query.Order("created_at DESC")

	if err := query.Find(&tasks).Error; err != nil {
		return nil, 0, err
	}

	return tasks, total, nil
}

// RecoverTasks 恢复未完成任务（服务启动时调用）
func (s *TaskService) RecoverTasks() error {
	var tasks []models.Task

	// 查找未完成的任务
	err := s.db.Where("status IN ?", []string{
		models.TaskStatusPending,
		models.TaskStatusRunning,
		models.TaskStatusRetrying,
	}).Find(&tasks).Error

	if err != nil {
		return err
	}

	logrus.Infof("发现 %d 个未完成任务，开始恢复...", len(tasks))

	for _, task := range tasks {
		// 重置状态
		task.Status = models.TaskStatusPending
		task.Progress = 0
		task.Message = "任务恢复中..."

		// 检查是否支持断点续传
		if task.Resumable && task.Checkpoint != "" {
			task.Message = "从断点恢复..."
		}

		// 重新加入队列
		s.queue.Push(&task)
		s.db.Save(&task)
	}

	return nil
}

// processQueue 处理任务队列
func (s *TaskService) processQueue() {
	for {
		// 从队列获取任务
		task := s.queue.Pop()
		if task == nil {
			time.Sleep(1 * time.Second)
			continue
		}

		// 执行任务
		go s.executeTask(task)
	}
}

// executeTask 执行任务
func (s *TaskService) executeTask(task *models.Task) {
	// 创建任务上下文
	ctx, cancel := context.WithCancel(context.Background())
	taskCtx := &TaskContext{
		Task:     task,
		Cancel:   cancel,
		Progress: make(chan float64, 10),
		Message:  make(chan string, 10),
	}

	// 注册任务上下文
	s.mu.Lock()
	s.tasks[task.ID] = taskCtx
	s.mu.Unlock()

	// 任务完成后清理
	defer func() {
		s.mu.Lock()
		delete(s.tasks, task.ID)
		s.mu.Unlock()
		close(taskCtx.Progress)
		close(taskCtx.Message)
	}()

	// 更新任务状态为运行中
	now := time.Now()
	task.Status = models.TaskStatusRunning
	task.StartedAt = &now
	s.db.Save(task)

	// 监听进度更新
	go s.monitorProgress(task, taskCtx)

	// 执行任务
	err := s.executor.Execute(ctx, taskCtx)

	// 更新任务状态
	completedAt := time.Now()
	task.CompletedAt = &completedAt
	if task.StartedAt != nil {
		task.Duration = int64(completedAt.Sub(*task.StartedAt).Seconds())
	}

	if err != nil {
		task.Status = models.TaskStatusFailed
		task.Error = err.Error()
		task.LastError = err.Error()
		logrus.Errorf("任务执行失败: ID=%d, Error=%v", task.ID, err)

		// 判断是否需要重试
		if task.RetryCount < task.MaxRetries && IsRetryableError(err) {
			logrus.Infof("任务将重试: ID=%d, RetryCount=%d", task.ID, task.RetryCount+1)
			go s.RetryTask(task.ID)
		}
	} else {
		task.Status = models.TaskStatusCompleted
		task.Progress = 100
		task.Message = "任务完成"
		logrus.Infof("任务执行成功: ID=%d", task.ID)
	}

	s.db.Save(task)
}

// monitorProgress 监听进度更新
func (s *TaskService) monitorProgress(task *models.Task, taskCtx *TaskContext) {
	for {
		select {
		case progress, ok := <-taskCtx.Progress:
			if !ok {
				return
			}
			task.Progress = progress
			s.db.Model(task).Update("progress", progress)

		case message, ok := <-taskCtx.Message:
			if !ok {
				return
			}
			task.Message = message
			s.db.Model(task).Update("message", message)
		}
	}
}

// TaskFilters 任务过滤器
type TaskFilters struct {
	Status   string
	Type     string
	UserID   uint
	Page     int
	PageSize int
}

// CheckDependencies 检查任务依赖
func (s *TaskService) CheckDependencies(task *models.Task) (bool, error) {
	if task.DependsOn == "" {
		return true, nil
	}

	var depends []uint
	if err := json.Unmarshal([]byte(task.DependsOn), &depends); err != nil {
		return false, err
	}

	for _, depID := range depends {
		var depTask models.Task
		if err := s.db.First(&depTask, depID).Error; err != nil {
			return false, err
		}

		if depTask.Status != models.TaskStatusCompleted {
			return false, nil
		}
	}

	return true, nil
}

// IsRetryableError 判断错误是否可重试
func IsRetryableError(err error) bool {
	if err == nil {
		return false
	}

	retryableErrors := []string{
		"connection refused",
		"timeout",
		"temporary failure",
		"disk full",
	}

	errMsg := err.Error()
	for _, retryable := range retryableErrors {
		if contains(errMsg, retryable) {
			return true
		}
	}

	return false
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

package ai3d

import (
	"context"
	"fmt"
	"time"

	"go_wails_project_manager/logger"
	"go_wails_project_manager/models/ai3d"

	"gorm.io/gorm"
)

type TaskService struct {
	db       *gorm.DB
	adapters map[string]ProviderAdapter
	poller   *TaskPoller
}

func NewTaskService(db *gorm.DB, pollInterval time.Duration) *TaskService {
	service := &TaskService{
		db:       db,
		adapters: make(map[string]ProviderAdapter),
	}
	
	// 创建轮询器
	service.poller = NewTaskPoller(db, service, pollInterval)
	
	return service
}

// RegisterAdapter 注册平台适配器
func (s *TaskService) RegisterAdapter(adapter ProviderAdapter) {
	s.adapters[adapter.GetName()] = adapter
	logger.Log.Infof("已注册AI3D适配器: %s", adapter.GetName())
}

// StartPoller 启动轮询器
func (s *TaskService) StartPoller() {
	s.poller.Start()
}

// StopPoller 停止轮询器
func (s *TaskService) StopPoller() {
	s.poller.Stop()
}

// CreateTask 创建任务
func (s *TaskService) CreateTask(ctx context.Context, task *ai3d.Task) error {
	// 获取适配器
	adapter, ok := s.adapters[task.Provider]
	if !ok {
		return fmt.Errorf("不支持的平台: %s", task.Provider)
	}

	// 提交到平台
	providerTaskID, err := adapter.SubmitTask(ctx, task)
	if err != nil {
		return fmt.Errorf("提交任务失败: %w", err)
	}

	// 保存到数据库
	task.ProviderTaskID = providerTaskID
	task.Status = "WAIT"
	task.Progress = 0

	return s.db.Create(task).Error
}

// GetTask 获取任务
func (s *TaskService) GetTask(id uint) (*ai3d.Task, error) {
	var task ai3d.Task
	err := s.db.First(&task, id).Error
	return &task, err
}

// GetTaskByProviderID 根据平台任务ID获取任务
func (s *TaskService) GetTaskByProviderID(provider, providerTaskID string) (*ai3d.Task, error) {
	var task ai3d.Task
	err := s.db.Where("provider = ? AND provider_task_id = ?", provider, providerTaskID).First(&task).Error
	return &task, err
}

// ListTasks 任务列表
func (s *TaskService) ListTasks(page, pageSize int, filters map[string]string) ([]*ai3d.Task, int64, error) {
	query := s.db.Model(&ai3d.Task{})

	// 应用过滤器
	if provider := filters["provider"]; provider != "" {
		query = query.Where("provider = ?", provider)
	}
	if status := filters["status"]; status != "" {
		query = query.Where("status = ?", status)
	}
	if keyword := filters["keyword"]; keyword != "" {
		query = query.Where("name LIKE ?", "%"+keyword+"%")
	}

	// 统计总数
	var total int64
	query.Count(&total)

	// 分页查询
	var tasks []*ai3d.Task
	err := query.Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&tasks).Error

	return tasks, total, err
}

// PollTask 轮询单个任务
func (s *TaskService) PollTask(ctx context.Context, id uint) error {
	task, err := s.GetTask(id)
	if err != nil {
		return err
	}

	return s.pollTaskInternal(ctx, task)
}

// pollTaskInternal 内部轮询方法
func (s *TaskService) pollTaskInternal(ctx context.Context, task *ai3d.Task) error {
	// 已完成的任务不再轮询
	if task.Status == "DONE" || task.Status == "FAIL" {
		return nil
	}

	// 获取适配器
	adapter, ok := s.adapters[task.Provider]
	if !ok {
		return fmt.Errorf("不支持的平台: %s", task.Provider)
	}

	// 查询平台状态
	status, err := adapter.QueryTask(ctx, task.ProviderTaskID)
	if err != nil {
		logger.Log.Errorf("查询任务 %d (%s:%s) 失败: %v", task.ID, task.Provider, task.ProviderTaskID, err)
		return err
	}

	// 更新数据库
	updates := map[string]interface{}{
		"status":   status.Status,
		"progress": status.Progress,
	}

	if status.ModelURL != "" {
		updates["model_url"] = status.ModelURL
	}
	if status.ThumbnailURL != "" {
		updates["thumbnail_url"] = status.ThumbnailURL
	}
	if status.ErrorCode != "" {
		updates["error_code"] = status.ErrorCode
	}
	if status.ErrorMessage != "" {
		updates["error_message"] = status.ErrorMessage
	}

	// 如果任务完成，下载文件
	if status.Status == "DONE" && status.ModelURL != "" {
		logger.Log.Infof("任务 %d (%s:%s) 完成，开始下载文件", task.ID, task.Provider, task.ProviderTaskID)
		
		// 更新任务的ModelURL和ThumbnailURL，以便下载
		task.ModelURL = &status.ModelURL
		if status.ThumbnailURL != "" {
			task.ThumbnailURL = &status.ThumbnailURL
		}
		
		result, err := adapter.DownloadResult(ctx, task)
		if err == nil {
			if result.LocalPath != "" {
				updates["local_path"] = result.LocalPath
			}
			if result.NASPath != "" {
				updates["nas_path"] = result.NASPath
			}
			if result.ThumbnailPath != "" {
				updates["thumbnail_path"] = result.ThumbnailPath
			}
			updates["file_size"] = result.FileSize
			updates["file_hash"] = result.FileHash
			
			logger.Log.Infof("任务 %d 文件已保存: %s", task.ID, result.NASPath)
		} else {
			logger.Log.Errorf("任务 %d 下载文件失败: %v", task.ID, err)
		}
	}

	return s.db.Model(task).Updates(updates).Error
}

// DeleteTask 删除任务
func (s *TaskService) DeleteTask(id uint) error {
	return s.db.Delete(&ai3d.Task{}, id).Error
}

// GetPendingTasks 获取待处理的任务
func (s *TaskService) GetPendingTasks() ([]*ai3d.Task, error) {
	var tasks []*ai3d.Task
	err := s.db.Where("status IN ?", []string{"WAIT", "RUN"}).
		Order("created_at ASC").
		Find(&tasks).Error
	return tasks, err
}

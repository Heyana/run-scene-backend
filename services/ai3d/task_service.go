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

	// 提交到平台（适配器会更新task.GenerationParams）
	providerTaskID, err := adapter.SubmitTask(ctx, task)
	if err != nil {
		return fmt.Errorf("提交任务失败: %w", err)
	}

	// 保存到数据库
	task.ProviderTaskID = providerTaskID
	task.Status = "WAIT"
	task.Progress = 0

	// 打印调试信息
	logger.Log.Infof("创建任务，GenerationParams: %+v", task.GenerationParams)

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

	// 分页查询 - 只选择需要的字段，避免加载大字段
	var tasks []*ai3d.Task
	err := query.Select("id", "created_at", "updated_at", "deleted_at", 
		"provider", "provider_task_id", "status", "progress", 
		"input_type", "prompt", "generation_params",
		"model_url", "pre_remeshed_url", "thumbnail_url", 
		"local_path", "pre_remeshed_path", "pre_remeshed_nas_path",
		"nas_path", "thumbnail_path", "file_size", "file_hash",
		"error_code", "error_message", "name", "description", 
		"category", "tags", "created_by", "created_ip").
		Order("created_at DESC").
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

	// 打印任务状态
	logger.Log.Infof("任务 %d (%s:%s) 状态: %s, 进度: %d%%", 
		task.ID, task.Provider, task.ProviderTaskID, status.Status, status.Progress)

	// 检查是否需要更新数据库
	// 只在以下情况更新：
	// 1. 状态变化
	// 2. 进度变化超过10%
	// 3. 任务完成或失败
	shouldUpdate := false
	progressDiff := 0
	
	if task.Status != status.Status {
		shouldUpdate = true
		logger.Log.Infof("任务 %d 状态变化: %s -> %s", task.ID, task.Status, status.Status)
	} else if task.Progress != status.Progress {
		progressDiff = status.Progress - task.Progress
		if progressDiff < 0 {
			progressDiff = -progressDiff
		}
		// 进度变化超过10%或达到完成状态时更新
		if progressDiff >= 10 || status.Status == "DONE" || status.Status == "FAIL" {
			shouldUpdate = true
			logger.Log.Infof("任务 %d 进度变化: %d%% -> %d%% (变化%d%%)", 
				task.ID, task.Progress, status.Progress, progressDiff)
		}
	}

	// 如果不需要更新，直接返回
	if !shouldUpdate && status.Status != "DONE" && status.Status != "FAIL" {
		logger.Log.Debugf("任务 %d 进度变化不足10%%，跳过数据库更新", task.ID)
		return nil
	}

	// 更新数据库
	updates := map[string]interface{}{
		"status":   status.Status,
		"progress": status.Progress,
	}

	if status.ModelURL != "" {
		updates["model_url"] = status.ModelURL
	}
	if status.PreRemeshedURL != "" {
		updates["pre_remeshed_url"] = status.PreRemeshedURL
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
		
		// 更新任务的ModelURL、PreRemeshedURL和ThumbnailURL，以便下载
		task.ModelURL = &status.ModelURL
		if status.PreRemeshedURL != "" {
			task.PreRemeshedURL = &status.PreRemeshedURL
		}
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
			if result.PreRemeshedPath != "" {
				updates["pre_remeshed_path"] = result.PreRemeshedPath
			}
			if result.PreRemeshedNASPath != "" {
				updates["pre_remeshed_nas_path"] = result.PreRemeshedNASPath
			}
			updates["file_size"] = result.FileSize
			updates["file_hash"] = result.FileHash
			
			logger.Log.Infof("任务 %d 文件已保存: %s", task.ID, result.NASPath)
			if result.PreRemeshedNASPath != "" {
				logger.Log.Infof("任务 %d PreRemeshed模型已保存: %s", task.ID, result.PreRemeshedNASPath)
			}
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

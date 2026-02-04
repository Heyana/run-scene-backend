package hunyuan

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"go_wails_project_manager/models/hunyuan"

	"gorm.io/gorm"
)

// TaskService 任务服务
type TaskService struct {
	db      *gorm.DB
	client  *HunyuanClient
	config  *ConfigService
	storage *StorageService
}

// NewTaskService 创建任务服务
func NewTaskService(db *gorm.DB, config *ConfigService, storage *StorageService) *TaskService {
	return &TaskService{
		db:      db,
		config:  config,
		storage: storage,
	}
}

// UserInfo 用户信息
type UserInfo struct {
	Username string
	IP       string
}

// TaskFilters 任务筛选条件
type TaskFilters struct {
	Status    string
	InputType string
	Keyword   string
}

// CreateTask 创建任务
func (s *TaskService) CreateTask(params *GenerateParams, userInfo UserInfo) (*hunyuan.HunyuanTask, error) {
	// 检查并发限制
	cfg, err := s.config.GetConfig()
	if err != nil {
		return nil, err
	}

	var runningCount int64
	s.db.Model(&hunyuan.HunyuanTask{}).
		Where("status IN ?", []string{"WAIT", "RUN"}).
		Count(&runningCount)

	if int(runningCount) >= cfg.MaxConcurrent {
		return nil, fmt.Errorf("已达到最大并发限制(%d)", cfg.MaxConcurrent)
	}

	// 获取客户端
	client, err := s.config.GetClient()
	if err != nil {
		return nil, err
	}
	s.client = client

	// 提交任务到腾讯云
	jobID, err := client.SubmitJob(params)
	if err != nil {
		return nil, fmt.Errorf("提交任务失败: %w", err)
	}

	// 生成任务名称
	name := s.generateTaskName(params)

	// 创建任务记录
	task := &hunyuan.HunyuanTask{
		JobID:        jobID,
		Status:       "WAIT",
		InputType:    s.getInputType(params),
		Prompt:       params.Prompt,
		ImageURL:     params.ImageURL,
		Model:        params.Model,
		FaceCount:    params.FaceCount,
		GenerateType: params.GenerateType,
		EnablePBR:    params.EnablePBR != nil && *params.EnablePBR,
		ResultFormat: stringOrDefault(params.ResultFormat, "GLB"),
		Name:         name,
		Category:     cfg.DefaultCategory,
		CreatedBy:    userInfo.Username,
		CreatedIP:    userInfo.IP,
	}

	if err := s.db.Create(task).Error; err != nil {
		return nil, fmt.Errorf("创建任务记录失败: %w", err)
	}

	return task, nil
}

// GetTaskStatus 查询任务状态
func (s *TaskService) GetTaskStatus(taskID uint) (*hunyuan.HunyuanTask, error) {
	var task hunyuan.HunyuanTask
	if err := s.db.First(&task, taskID).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

// PollTask 轮询任务
func (s *TaskService) PollTask(taskID uint) error {
	// 获取任务
	task, err := s.GetTaskStatus(taskID)
	if err != nil {
		return err
	}

	// 如果已经完成，直接返回
	if task.Status == "DONE" || task.Status == "FAIL" {
		return nil
	}

	// 获取客户端
	client, err := s.config.GetClient()
	if err != nil {
		return err
	}

	// 查询任务状态
	resp, err := client.QueryJob(task.JobID)
	if err != nil {
		return fmt.Errorf("查询任务失败: %w", err)
	}

	// 更新任务状态
	updates := map[string]interface{}{
		"status": resp.Response.Status,
	}

	if resp.Response.ErrorCode != "" {
		updates["error_code"] = resp.Response.ErrorCode
		updates["error_message"] = resp.Response.ErrorMessage
	}

	// 如果任务完成，保存结果
	if resp.Response.Status == "DONE" && len(resp.Response.ResultFiles) > 0 {
		// 序列化结果文件
		filesJSON, _ := json.Marshal(resp.Response.ResultFiles)
		updates["result_files"] = string(filesJSON)

		// 下载并保存文件
		info, err := s.storage.SaveTaskResult(task, resp.Response.ResultFiles)
		if err == nil {
			updates["local_path"] = info.LocalPath
			if info.NASPath != "" {
				updates["nas_path"] = info.NASPath
			}
			if info.ThumbnailPath != "" {
				updates["thumbnail_path"] = info.ThumbnailPath
			}
			updates["file_size"] = info.FileSize
			updates["file_hash"] = info.FileHash
		}
	}

	return s.db.Model(task).Updates(updates).Error
}

// ListTasks 任务列表
func (s *TaskService) ListTasks(page, pageSize int, filters TaskFilters) ([]hunyuan.HunyuanTask, int64, error) {
	var tasks []hunyuan.HunyuanTask
	var total int64

	query := s.db.Model(&hunyuan.HunyuanTask{})

	// 应用筛选条件
	if filters.Status != "" {
		query = query.Where("status = ?", filters.Status)
	}
	if filters.InputType != "" {
		query = query.Where("input_type = ?", filters.InputType)
	}
	if filters.Keyword != "" {
		query = query.Where("name LIKE ? OR job_id LIKE ?",
			"%"+filters.Keyword+"%", "%"+filters.Keyword+"%")
	}

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&tasks).Error; err != nil {
		return nil, 0, err
	}

	return tasks, total, nil
}

// CancelTask 取消任务
func (s *TaskService) CancelTask(taskID uint) error {
	task, err := s.GetTaskStatus(taskID)
	if err != nil {
		return err
	}

	if task.Status != "WAIT" && task.Status != "RUN" {
		return errors.New("只能取消等待中或运行中的任务")
	}

	// 更新状态为失败
	return s.db.Model(task).Updates(map[string]interface{}{
		"status":        "FAIL",
		"error_message": "用户取消",
	}).Error
}

// RetryTask 重试任务
func (s *TaskService) RetryTask(taskID uint) error {
	task, err := s.GetTaskStatus(taskID)
	if err != nil {
		return err
	}

	if task.Status != "FAIL" {
		return errors.New("只能重试失败的任务")
	}

	// 重新提交任务
	client, err := s.config.GetClient()
	if err != nil {
		return err
	}

	params := &GenerateParams{
		Model:        task.Model,
		Prompt:       task.Prompt,
		ImageURL:     task.ImageURL,
		FaceCount:    task.FaceCount,
		GenerateType: task.GenerateType,
	}

	enablePBR := task.EnablePBR
	params.EnablePBR = &enablePBR

	jobID, err := client.SubmitJob(params)
	if err != nil {
		return fmt.Errorf("重新提交任务失败: %w", err)
	}

	// 更新任务记录
	return s.db.Model(task).Updates(map[string]interface{}{
		"job_id":        jobID,
		"status":        "WAIT",
		"error_code":    nil,
		"error_message": nil,
	}).Error
}

// DeleteTask 删除任务
func (s *TaskService) DeleteTask(taskID uint) error {
	task, err := s.GetTaskStatus(taskID)
	if err != nil {
		return err
	}

	// 删除文件
	if err := s.storage.DeleteFiles(task); err != nil {
		// 记录错误但继续删除记录
		fmt.Printf("删除文件失败: %v\n", err)
	}

	// 删除数据库记录
	return s.db.Delete(task).Error
}

// GetStatistics 获取统计信息
func (s *TaskService) GetStatistics() (map[string]interface{}, error) {
	var total, waiting, running, done, failed int64

	s.db.Model(&hunyuan.HunyuanTask{}).Count(&total)
	s.db.Model(&hunyuan.HunyuanTask{}).Where("status = ?", "WAIT").Count(&waiting)
	s.db.Model(&hunyuan.HunyuanTask{}).Where("status = ?", "RUN").Count(&running)
	s.db.Model(&hunyuan.HunyuanTask{}).Where("status = ?", "DONE").Count(&done)
	s.db.Model(&hunyuan.HunyuanTask{}).Where("status = ?", "FAIL").Count(&failed)

	// 今日任务数
	today := time.Now().Format("2006-01-02")
	var todayCount int64
	s.db.Model(&hunyuan.HunyuanTask{}).
		Where("DATE(created_at) = ?", today).
		Count(&todayCount)

	// 成功率
	successRate := 0.0
	if total > 0 {
		successRate = float64(done) / float64(total) * 100
	}

	return map[string]interface{}{
		"total":       total,
		"waiting":     waiting,
		"running":     running,
		"done":        done,
		"failed":      failed,
		"todayCount":  todayCount,
		"successRate": successRate,
	}, nil
}

// generateTaskName 生成任务名称
func (s *TaskService) generateTaskName(params *GenerateParams) string {
	if params.Prompt != nil && *params.Prompt != "" {
		// 截取前20个字符作为名称
		prompt := *params.Prompt
		if len(prompt) > 20 {
			prompt = prompt[:20] + "..."
		}
		return prompt
	}
	return fmt.Sprintf("任务_%s", time.Now().Format("20060102_150405"))
}

// getInputType 获取输入类型
func (s *TaskService) getInputType(params *GenerateParams) string {
	if params.Prompt != nil && *params.Prompt != "" {
		return "text"
	}
	if len(params.MultiViewImages) > 0 {
		return "multi_view"
	}
	return "image"
}

// stringOrDefault 返回字符串或默认值
func stringOrDefault(s *string, defaultValue string) string {
	if s == nil || *s == "" {
		return defaultValue
	}
	return *s
}

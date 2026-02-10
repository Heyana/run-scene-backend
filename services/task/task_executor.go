// Package task 任务执行器
package task

import (
	"context"
	"encoding/json"
	"fmt"
	"go_wails_project_manager/models"
	"go_wails_project_manager/services/fileprocessor"
	"go_wails_project_manager/services/fileprocessor/processors"
	"strings"
)

// TaskExecutor 任务执行器
type TaskExecutor struct {
	maxConcurrent        int
	semaphore            chan struct{}
	fileProcessorService fileprocessor.IFileProcessorService
}

// NewTaskExecutor 创建任务执行器
func NewTaskExecutor(maxConcurrent int, fpService fileprocessor.IFileProcessorService) *TaskExecutor {
	return &TaskExecutor{
		maxConcurrent:        maxConcurrent,
		semaphore:            make(chan struct{}, maxConcurrent),
		fileProcessorService: fpService,
	}
}

// Execute 执行任务
func (e *TaskExecutor) Execute(ctx context.Context, taskCtx *TaskContext) error {
	// 获取信号量（限制并发）
	select {
	case e.semaphore <- struct{}{}:
		defer func() { <-e.semaphore }()
	case <-ctx.Done():
		return ctx.Err()
	}

	// 根据任务类型执行
	return e.executeTask(ctx, taskCtx)
}

// executeTask 执行具体任务
func (e *TaskExecutor) executeTask(ctx context.Context, taskCtx *TaskContext) error {
	task := taskCtx.Task

	// 发送进度更新
	sendProgress := func(progress float64) {
		select {
		case taskCtx.Progress <- progress:
		default:
		}
	}

	// 发送消息更新
	sendMessage := func(message string) {
		select {
		case taskCtx.Message <- message:
		default:
		}
	}

	sendMessage("任务开始执行...")
	sendProgress(0)

	// 根据任务类型执行不同的处理
	switch task.Type {
	case models.TaskTypeVideoPreview:
		return e.executeVideoPreview(ctx, task, sendProgress, sendMessage)
	case models.TaskTypeVideoThumbnail:
		return e.executeVideoThumbnail(ctx, task, sendProgress, sendMessage)
	case models.TaskTypeVideoConvert:
		return e.executeVideoConvert(ctx, task, sendProgress, sendMessage)
	case models.TaskTypeImageThumbnail:
		return e.executeImageThumbnail(ctx, task, sendProgress, sendMessage)
	case models.TaskTypeImageConvert:
		return e.executeImageConvert(ctx, task, sendProgress, sendMessage)
	case models.TaskTypeDocumentPreview:
		return e.executeDocumentPreview(ctx, task, sendProgress, sendMessage)
	default:
		return fmt.Errorf("不支持的任务类型: %s", task.Type)
	}
}

// executeVideoPreview 执行视频预览生成
func (e *TaskExecutor) executeVideoPreview(ctx context.Context, task *models.Task, sendProgress func(float64), sendMessage func(string)) error {
	sendMessage("生成视频预览图...")
	sendProgress(10)

	// 解析任务参数
	var params struct {
		Size    int `json:"size"`
		Quality int `json:"quality"`
	}
	if task.Options != "" {
		if err := json.Unmarshal([]byte(task.Options), &params); err != nil {
			return fmt.Errorf("解析任务参数失败: %w", err)
		}
	}

	// 设置默认值
	if params.Size == 0 {
		params.Size = 256
	}
	if params.Quality == 0 {
		params.Quality = 85
	}

	sendProgress(30)

	// 调用文件处理器生成缩略图
	options := processors.ThumbnailOptions{
		Size:       params.Size,
		Quality:    params.Quality,
		OutputPath: task.OutputFile,
	}

	sendMessage("正在生成预览...")
	sendProgress(50)

	// 检测文件格式
	format := detectFormat(task.InputFile)
	outputPath, err := e.fileProcessorService.GenerateThumbnail(task.InputFile, format, options)
	if err != nil {
		return fmt.Errorf("生成预览失败: %w", err)
	}

	sendProgress(90)

	// 更新输出文件路径
	task.OutputFile = outputPath

	sendProgress(100)
	sendMessage("视频预览图生成完成")
	return nil
}

// executeVideoThumbnail 执行视频缩略图生成
func (e *TaskExecutor) executeVideoThumbnail(ctx context.Context, task *models.Task, sendProgress func(float64), sendMessage func(string)) error {
	sendMessage("生成视频缩略图...")
	sendProgress(10)

	// 解析任务参数
	var params struct {
		Size    int `json:"size"`
		Quality int `json:"quality"`
	}
	if task.Options != "" {
		if err := json.Unmarshal([]byte(task.Options), &params); err != nil {
			return fmt.Errorf("解析任务参数失败: %w", err)
		}
	}

	// 设置默认值
	if params.Size == 0 {
		params.Size = 256
	}
	if params.Quality == 0 {
		params.Quality = 85
	}

	sendProgress(30)

	// 调用文件处理器生成缩略图
	options := processors.ThumbnailOptions{
		Size:       params.Size,
		Quality:    params.Quality,
		OutputPath: task.OutputFile,
	}

	sendMessage("正在生成缩略图...")
	sendProgress(50)

	format := detectFormat(task.InputFile)
	outputPath, err := e.fileProcessorService.GenerateThumbnail(task.InputFile, format, options)
	if err != nil {
		return fmt.Errorf("生成缩略图失败: %w", err)
	}

	sendProgress(90)

	// 更新输出文件路径
	task.OutputFile = outputPath

	sendProgress(100)
	sendMessage("视频缩略图生成完成")
	return nil
}

// executeVideoConvert 执行视频转换
func (e *TaskExecutor) executeVideoConvert(ctx context.Context, task *models.Task, sendProgress func(float64), sendMessage func(string)) error {
	sendMessage("开始视频转换...")
	sendProgress(10)

	// 解析任务参数
	var params struct {
		Quality int `json:"quality"`
		Width   int `json:"width"`
		Height  int `json:"height"`
	}
	if task.Options != "" {
		if err := json.Unmarshal([]byte(task.Options), &params); err != nil {
			return fmt.Errorf("解析任务参数失败: %w", err)
		}
	}

	// 设置默认值
	if params.Quality == 0 {
		params.Quality = 85
	}

	sendProgress(20)
	sendMessage("正在转换视频...")
	sendProgress(30)

	format := detectFormat(task.InputFile)
	outputPath, err := e.fileProcessorService.GenerateThumbnail(task.InputFile, format, processors.ThumbnailOptions{
		Size:       params.Width,
		Quality:    params.Quality,
		OutputPath: task.OutputFile,
	})
	if err != nil {
		return fmt.Errorf("视频转换失败: %w", err)
	}

	sendProgress(95)

	// 更新输出文件路径
	task.OutputFile = outputPath

	sendProgress(100)
	sendMessage("视频转换完成")
	return nil
}

// executeImageThumbnail 执行图片缩略图生成
func (e *TaskExecutor) executeImageThumbnail(ctx context.Context, task *models.Task, sendProgress func(float64), sendMessage func(string)) error {
	sendMessage("生成图片缩略图...")
	sendProgress(10)

	// 解析任务参数
	var params struct {
		Size    int `json:"size"`
		Quality int `json:"quality"`
	}
	if task.Options != "" {
		if err := json.Unmarshal([]byte(task.Options), &params); err != nil {
			return fmt.Errorf("解析任务参数失败: %w", err)
		}
	}

	// 设置默认值
	if params.Size == 0 {
		params.Size = 256
	}
	if params.Quality == 0 {
		params.Quality = 85
	}

	sendProgress(30)

	// 调用文件处理器生成缩略图
	options := processors.ThumbnailOptions{
		Size:       params.Size,
		Quality:    params.Quality,
		OutputPath: task.OutputFile,
	}

	sendMessage("正在生成缩略图...")
	sendProgress(50)

	format := detectFormat(task.InputFile)
	outputPath, err := e.fileProcessorService.GenerateThumbnail(task.InputFile, format, options)
	if err != nil {
		return fmt.Errorf("生成缩略图失败: %w", err)
	}

	sendProgress(90)

	// 更新输出文件路径
	task.OutputFile = outputPath

	sendProgress(100)
	sendMessage("图片缩略图生成完成")
	return nil
}

// executeImageConvert 执行图片转换
func (e *TaskExecutor) executeImageConvert(ctx context.Context, task *models.Task, sendProgress func(float64), sendMessage func(string)) error {
	sendMessage("开始图片转换...")
	sendProgress(10)

	// 解析任务参数
	var params struct {
		Quality int `json:"quality"`
		Width   int `json:"width"`
		Height  int `json:"height"`
	}
	if task.Options != "" {
		if err := json.Unmarshal([]byte(task.Options), &params); err != nil {
			return fmt.Errorf("解析任务参数失败: %w", err)
		}
	}

	sendProgress(30)

	// 调用文件处理器转换图片（使用缩略图功能）
	options := processors.ThumbnailOptions{
		Size:       params.Width,
		Quality:    params.Quality,
		OutputPath: task.OutputFile,
	}

	sendMessage("正在转换图片...")
	sendProgress(50)

	format := detectFormat(task.InputFile)
	outputPath, err := e.fileProcessorService.GenerateThumbnail(task.InputFile, format, options)
	if err != nil {
		return fmt.Errorf("图片转换失败: %w", err)
	}

	sendProgress(90)

	// 更新输出文件路径
	task.OutputFile = outputPath

	sendProgress(100)
	sendMessage("图片转换完成")
	return nil
}

// executeDocumentPreview 执行文档预览生成
func (e *TaskExecutor) executeDocumentPreview(ctx context.Context, task *models.Task, sendProgress func(float64), sendMessage func(string)) error {
	sendMessage("生成文档预览图...")
	sendProgress(10)

	// 解析任务参数
	var params struct {
		Size    int `json:"size"`
		Quality int `json:"quality"`
	}
	if task.Options != "" {
		if err := json.Unmarshal([]byte(task.Options), &params); err != nil {
			return fmt.Errorf("解析任务参数失败: %w", err)
		}
	}

	// 设置默认值
	if params.Size == 0 {
		params.Size = 256
	}
	if params.Quality == 0 {
		params.Quality = 85
	}

	sendProgress(30)

	// 调用文件处理器生成预览
	options := processors.ThumbnailOptions{
		Size:       params.Size,
		Quality:    params.Quality,
		OutputPath: task.OutputFile,
	}

	sendMessage("正在生成预览...")
	sendProgress(50)

	format := detectFormat(task.InputFile)
	outputPath, err := e.fileProcessorService.GenerateThumbnail(task.InputFile, format, options)
	if err != nil {
		return fmt.Errorf("生成预览失败: %w", err)
	}

	sendProgress(90)

	// 更新输出文件路径
	task.OutputFile = outputPath

	sendProgress(100)
	sendMessage("文档预览图生成完成")
	return nil
}

// detectFormat 从文件路径检测格式
func detectFormat(filePath string) string {
	// 简单实现：从文件扩展名获取格式
	parts := strings.Split(filePath, ".")
	if len(parts) > 0 {
		return strings.ToLower(parts[len(parts)-1])
	}
	return ""
}

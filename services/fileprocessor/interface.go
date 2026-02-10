// Package fileprocessor 文件处理器服务接口定义
package fileprocessor

import "go_wails_project_manager/services/fileprocessor/processors"

// IFileProcessorService 文件处理器服务接口（供其他包使用，避免循环依赖）
type IFileProcessorService interface {
	// ExtractMetadata 提取文件元数据
	ExtractMetadata(filePath, format string) (*processors.FileMetadata, error)

	// GenerateThumbnail 生成缩略图
	GenerateThumbnail(filePath, format string, options processors.ThumbnailOptions) (string, error)

	// GetProcessor 根据文件格式获取处理器
	GetProcessor(format string) processors.FileProcessor

	// ListSupportedFormats 列出支持的格式
	ListSupportedFormats() map[string][]string
}

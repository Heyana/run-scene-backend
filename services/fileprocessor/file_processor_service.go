// Package fileprocessor 文件处理器服务
package fileprocessor

import (
	"fmt"
	"go_wails_project_manager/services/fileprocessor/processors"
	"go_wails_project_manager/utils/document"
	"go_wails_project_manager/utils/image"
	"go_wails_project_manager/utils/model"
	"go_wails_project_manager/utils/video"
	"time"
)

// FileProcessorService 文件处理器服务
type FileProcessorService struct {
	processors []processors.FileProcessor
}

// NewFileProcessorService 创建文件处理器服务
func NewFileProcessorService(config *Config) *FileProcessorService {
	service := &FileProcessorService{
		processors: make([]processors.FileProcessor, 0),
	}

	// 初始化工具
	ffmpeg := video.NewFFmpeg(config.FFmpeg.BinPath, config.FFmpeg.Timeout)
	imagemagick := image.NewImageMagick(config.ImageMagick.BinPath, config.ImageMagick.Timeout)
	pdftool := document.NewPDFTool(config.PDF.BinPath, config.PDF.Timeout)

	// 注册处理器
	service.RegisterProcessor(processors.NewVideoProcessor(ffmpeg))
	service.RegisterProcessor(processors.NewImageProcessor(imagemagick))
	service.RegisterProcessor(processors.NewDocumentProcessor(pdftool))

	// 如果配置了 Blender，注册 3D 模型处理器
	if config.Blender.BinPath != "" {
		fmt.Printf("[FileProcessor] 注册 3D 模型处理器: bin_path=%s, script_path=%s\n", 
			config.Blender.BinPath, config.Blender.ScriptPath)
		blenderTimeout := time.Duration(config.Blender.Timeout) * time.Second
		blender := model.NewBlender(config.Blender.BinPath, config.Blender.ScriptPath, blenderTimeout)
		service.RegisterProcessor(processors.NewModelProcessor(blender, imagemagick))
	} else {
		fmt.Printf("[FileProcessor] 未配置 Blender，跳过 3D 模型处理器注册\n")
	}

	return service
}

// RegisterProcessor 注册处理器
func (s *FileProcessorService) RegisterProcessor(processor processors.FileProcessor) {
	s.processors = append(s.processors, processor)
}

// GetProcessor 根据文件格式获取处理器
func (s *FileProcessorService) GetProcessor(format string) processors.FileProcessor {
	for _, processor := range s.processors {
		if processor.Support(format) {
			return processor
		}
	}
	return nil
}

// ExtractMetadata 提取元数据
func (s *FileProcessorService) ExtractMetadata(filePath, format string) (*processors.FileMetadata, error) {
	processor := s.GetProcessor(format)
	if processor == nil {
		return nil, fmt.Errorf("不支持的文件格式: %s", format)
	}
	return processor.ExtractMetadata(filePath)
}

// GeneratePreview 生成预览图
func (s *FileProcessorService) GeneratePreview(filePath, format string, options processors.PreviewOptions) (*processors.PreviewResult, error) {
	processor := s.GetProcessor(format)
	if processor == nil {
		return nil, fmt.Errorf("不支持的文件格式: %s", format)
	}
	return processor.GeneratePreview(filePath, options)
}

// GenerateThumbnail 生成缩略图
func (s *FileProcessorService) GenerateThumbnail(filePath, format string, options processors.ThumbnailOptions) (string, error) {
	processor := s.GetProcessor(format)
	if processor == nil {
		return "", fmt.Errorf("不支持的文件格式: %s", format)
	}
	return processor.GenerateThumbnail(filePath, options)
}

// Convert 格式转换
func (s *FileProcessorService) Convert(filePath, format string, options processors.ConvertOptions) (string, error) {
	processor := s.GetProcessor(format)
	if processor == nil {
		return "", fmt.Errorf("不支持的文件格式: %s", format)
	}
	return processor.Convert(filePath, options)
}

// Validate 验证文件
func (s *FileProcessorService) Validate(filePath, format string) error {
	processor := s.GetProcessor(format)
	if processor == nil {
		return fmt.Errorf("不支持的文件格式: %s", format)
	}
	return processor.Validate(filePath)
}

// ListSupportedFormats 列出支持的格式
func (s *FileProcessorService) ListSupportedFormats() map[string][]string {
	formats := make(map[string][]string)

	// 视频格式
	formats["video"] = []string{"mp4", "avi", "mov", "webm", "mkv", "flv", "wmv", "m4v", "mpg", "mpeg"}

	// 图片格式
	formats["image"] = []string{"jpg", "jpeg", "png", "gif", "webp", "bmp", "tiff", "tif", "svg", "ico"}

	// 文档格式
	formats["document"] = []string{"pdf", "doc", "docx", "ppt", "pptx", "xls", "xlsx", "txt", "md"}

	// 3D 模型格式
	formats["model"] = []string{"fbx", "obj", "glb", "gltf"}

	return formats
}

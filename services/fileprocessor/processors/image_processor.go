// Package processors 文件处理器实现
package processors

import (
	"fmt"
	"go_wails_project_manager/utils/image"
	"path/filepath"
)

// ImageProcessor 图片处理器
type ImageProcessor struct {
	imagemagick *image.ImageMagick
}

// NewImageProcessor 创建图片处理器
func NewImageProcessor(imagemagick *image.ImageMagick) *ImageProcessor {
	return &ImageProcessor{
		imagemagick: imagemagick,
	}
}

// Name 获取处理器名称
func (p *ImageProcessor) Name() string {
	return "ImageProcessor"
}

// Support 检查是否支持该文件格式
func (p *ImageProcessor) Support(format string) bool {
	supported := []string{
		"jpg", "jpeg", "png", "gif", "webp",
		"bmp", "tiff", "tif", "svg", "ico",
	}

	for _, f := range supported {
		if f == format {
			return true
		}
	}
	return false
}

// ExtractMetadata 提取文件元数据
func (p *ImageProcessor) ExtractMetadata(filePath string) (*FileMetadata, error) {
	imageMeta, err := p.imagemagick.ExtractMetadata(filePath)
	if err != nil {
		return nil, err
	}

	return &FileMetadata{
		FileName: filepath.Base(filePath),
		Format:   imageMeta.Format,
		Width:    imageMeta.Width,
		Height:   imageMeta.Height,
		FileSize: imageMeta.FileSize,
		Extra: map[string]interface{}{
			"color_space": imageMeta.ColorSpace,
			"depth":       imageMeta.Depth,
		},
	}, nil
}

// GeneratePreview 生成预览图
func (p *ImageProcessor) GeneratePreview(filePath string, options PreviewOptions) (*PreviewResult, error) {
	// 生成预览图
	err := p.imagemagick.Resize(
		filePath,
		options.OutputPath,
		options.Width,
		options.Height,
		true, // 保持宽高比
	)

	if err != nil {
		return nil, err
	}

	return &PreviewResult{
		ThumbnailPath: options.OutputPath,
		PreviewPaths:  []string{options.OutputPath},
	}, nil
}

// GenerateThumbnail 生成缩略图
func (p *ImageProcessor) GenerateThumbnail(filePath string, options ThumbnailOptions) (string, error) {
	// 添加日志
	fmt.Printf("[ImageProcessor] GenerateThumbnail 开始\n")
	fmt.Printf("[ImageProcessor] 输入文件: %s\n", filePath)
	fmt.Printf("[ImageProcessor] 输出文件: %s\n", options.OutputPath)
	fmt.Printf("[ImageProcessor] 尺寸: %d, 质量: %d\n", options.Size, options.Quality)

	// 检查是否是 SVG 文件
	ext := filepath.Ext(filePath)
	isSVG := ext == ".svg" || ext == ".SVG"

	var err error
	if isSVG {
		// SVG 使用特殊处理：保持宽高比，不裁剪
		err = p.imagemagick.GenerateThumbnailKeepAspect(
			filePath,
			options.OutputPath,
			options.Size,
			options.Quality,
		)
	} else {
		// 其他图片格式使用正方形裁剪
		err = p.imagemagick.GenerateThumbnail(
			filePath,
			options.OutputPath,
			options.Size,
			options.Quality,
		)
	}

	if err != nil {
		fmt.Printf("[ImageProcessor] 生成缩略图失败: %v\n", err)
		return "", err
	}

	fmt.Printf("[ImageProcessor] 生成缩略图成功: %s\n", options.OutputPath)
	return options.OutputPath, nil
}

// Convert 格式转换
func (p *ImageProcessor) Convert(filePath string, options ConvertOptions) (string, error) {
	err := p.imagemagick.Convert(
		filePath,
		options.OutputPath,
		options.Format,
		options.Quality,
	)

	if err != nil {
		return "", err
	}

	return options.OutputPath, nil
}

// Validate 验证文件完整性
func (p *ImageProcessor) Validate(filePath string) error {
	// 尝试提取元数据来验证文件
	_, err := p.imagemagick.ExtractMetadata(filePath)
	return err
}

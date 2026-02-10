// Package processors 3D 模型处理器
package processors

import (
	"context"
	"fmt"
	"go_wails_project_manager/utils/image"
	"go_wails_project_manager/utils/model"
	"os"
	"path/filepath"
	"strings"
)

// ModelProcessor 3D 模型处理器
type ModelProcessor struct {
	blender     *model.Blender
	imagemagick *image.ImageMagick
}

// NewModelProcessor 创建 3D 模型处理器
func NewModelProcessor(blender *model.Blender, imagemagick *image.ImageMagick) *ModelProcessor {
	return &ModelProcessor{
		blender:     blender,
		imagemagick: imagemagick,
	}
}

// Name 获取处理器名称
func (p *ModelProcessor) Name() string {
	return "ModelProcessor"
}

// Support 检查是否支持该文件格式
func (p *ModelProcessor) Support(format string) bool {
	format = strings.ToLower(format)
	supported := []string{
		"fbx", "obj", "glb", "gltf",
	}

	for _, f := range supported {
		if f == format {
			return true
		}
	}
	return false
}

// ExtractMetadata 提取文件元数据
func (p *ModelProcessor) ExtractMetadata(filePath string) (*FileMetadata, error) {
	modelMeta, err := p.blender.ExtractMetadata(filePath)
	if err != nil {
		return nil, err
	}

	return &FileMetadata{
		FileName: filepath.Base(filePath),
		Format:   modelMeta.Format,
		Extra: map[string]interface{}{
			"vertices": modelMeta.Vertices,
			"faces":    modelMeta.Faces,
			"objects":  modelMeta.Objects,
		},
	}, nil
}

// GeneratePreview 生成预览图
func (p *ModelProcessor) GeneratePreview(filePath string, options PreviewOptions) (*PreviewResult, error) {
	// 使用 Blender 渲染预览图
	renderOpts := model.RenderOptions{
		Width:   options.Width,
		Height:  options.Height,
		Quality: "fast", // 默认快速模式
	}

	ctx := context.Background()
	err := p.blender.RenderPreview(ctx, filePath, options.OutputPath, renderOpts)
	if err != nil {
		return nil, err
	}

	return &PreviewResult{
		ThumbnailPath: options.OutputPath,
		PreviewPaths:  []string{options.OutputPath},
	}, nil
}

// GenerateThumbnail 生成缩略图
func (p *ModelProcessor) GenerateThumbnail(filePath string, options ThumbnailOptions) (string, error) {
	fmt.Printf("[ModelProcessor] GenerateThumbnail 开始\n")
	fmt.Printf("[ModelProcessor] 输入文件: %s\n", filePath)
	fmt.Printf("[ModelProcessor] 输出文件: %s\n", options.OutputPath)
	fmt.Printf("[ModelProcessor] 尺寸: %d, 质量: %d\n", options.Size, options.Quality)

	// 步骤1: 使用 Blender 渲染为 PNG（临时文件）
	tempPNG := strings.TrimSuffix(options.OutputPath, filepath.Ext(options.OutputPath)) + "_temp.png"
	
	renderOpts := model.RenderOptions{
		Width:   options.Size,
		Height:  options.Size,
		Quality: "fast", // 快速模式
	}

	ctx := context.Background()
	err := p.blender.RenderPreview(ctx, filePath, tempPNG, renderOpts)
	if err != nil {
		fmt.Printf("[ModelProcessor] Blender 渲染失败: %v\n", err)
		return "", err
	}

	fmt.Printf("[ModelProcessor] Blender 渲染成功: %s\n", tempPNG)

	// 步骤2: 使用 ImageMagick 转换为 WebP
	err = p.imagemagick.Convert(tempPNG, options.OutputPath, "webp", options.Quality)
	if err != nil {
		fmt.Printf("[ModelProcessor] 转换为 WebP 失败: %v\n", err)
		// 清理临时文件
		os.Remove(tempPNG)
		return "", err
	}

	fmt.Printf("[ModelProcessor] 转换为 WebP 成功: %s\n", options.OutputPath)

	// 步骤3: 清理临时 PNG 文件
	err = os.Remove(tempPNG)
	if err != nil {
		fmt.Printf("[ModelProcessor] 警告: 清理临时文件失败: %v\n", err)
	}

	fmt.Printf("[ModelProcessor] 生成缩略图成功: %s\n", options.OutputPath)
	return options.OutputPath, nil
}

// Convert 格式转换（暂不支持）
func (p *ModelProcessor) Convert(filePath string, options ConvertOptions) (string, error) {
	return "", fmt.Errorf("3D 模型格式转换暂不支持")
}

// Validate 验证文件完整性
func (p *ModelProcessor) Validate(filePath string) error {
	// 简单验证：检查文件扩展名
	ext := strings.ToLower(filepath.Ext(filePath))
	if !p.Support(strings.TrimPrefix(ext, ".")) {
		return fmt.Errorf("不支持的 3D 模型格式: %s", ext)
	}
	return nil
}

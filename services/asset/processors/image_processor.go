package processors

import (
	"errors"
	"fmt"
	"go_wails_project_manager/config"
	"go_wails_project_manager/models"
	"image"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	
	"github.com/disintegration/imaging"
	_ "golang.org/x/image/webp"
)

// ImageProcessor 图片处理器
type ImageProcessor struct {
	config *config.AssetConfig
}

// NewImageProcessor 创建图片处理器
func NewImageProcessor(cfg *config.AssetConfig) *ImageProcessor {
	return &ImageProcessor{
		config: cfg,
	}
}

// GenerateThumbnail 生成缩略图
func (p *ImageProcessor) GenerateThumbnail(filePath, outputPath string) error {
	// 检查是否是特殊格式（APNG、GIF 等动画格式）
	ext := strings.ToLower(filepath.Ext(filePath))
	
	// 对于动画格式，直接复制原文件作为缩略图
	// 这样可以保留动画效果
	if ext == ".apng" || ext == ".gif" {
		return p.copyFile(filePath, outputPath)
	}
	
	// 尝试打开图片
	img, err := imaging.Open(filePath)
	if err != nil {
		// 如果 imaging 库无法打开（比如某些特殊格式、全景图等），直接复制原文件
		// 这样可以确保文件能够正常上传和使用
		return p.copyFile(filePath, outputPath)
	}
	
	// 尝试生成缩略图
	thumb := imaging.Fit(img, p.config.ThumbnailWidth, p.config.ThumbnailHeight, imaging.Lanczos)
	
	// 尝试保存为WebP格式
	err = imaging.Save(thumb, outputPath, imaging.JPEGQuality(p.config.ThumbnailQuality))
	if err != nil {
		// 如果保存失败，尝试直接复制原文件
		copyErr := p.copyFile(filePath, outputPath)
		if copyErr != nil {
			return fmt.Errorf("保存缩略图失败且复制原文件也失败: save error: %w, copy error: %v", err, copyErr)
		}
		return nil
	}
	
	return nil
}

// copyFile 复制文件
func (p *ImageProcessor) copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("打开源文件失败: %w", err)
	}
	defer sourceFile.Close()
	
	destFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("创建目标文件失败: %w", err)
	}
	defer destFile.Close()
	
	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return fmt.Errorf("复制文件失败: %w", err)
	}
	
	return nil
}

// ExtractMetadata 提取元数据
func (p *ImageProcessor) ExtractMetadata(filePath string) (*models.AssetMetadata, error) {
	// 尝试打开图片
	img, err := imaging.Open(filePath)
	if err != nil {
		// 如果无法打开（比如 APNG 等特殊格式），返回基本信息
		fileInfo, statErr := os.Stat(filePath)
		if statErr != nil {
			return nil, fmt.Errorf("获取文件信息失败: %w", statErr)
		}
		
		// 返回基本元数据
		metadata := &models.AssetMetadata{
			Width:     0, // 未知
			Height:    0, // 未知
			ColorMode: "Unknown",
		}
		
		// 尝试从文件大小推测是否有效
		if fileInfo.Size() > 0 {
			return metadata, nil
		}
		
		return nil, fmt.Errorf("无效的图片文件")
	}
	
	bounds := img.Bounds()
	
	metadata := &models.AssetMetadata{
		Width:     bounds.Dx(),
		Height:    bounds.Dy(),
		ColorMode: getColorMode(img),
	}
	
	return metadata, nil
}

// Validate 验证文件
func (p *ImageProcessor) Validate(file *multipart.FileHeader) error {
	// 检查文件扩展名
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(file.Filename), "."))
	
	supported := false
	for _, format := range p.SupportedFormats() {
		if ext == format {
			supported = true
			break
		}
	}
	
	if !supported {
		return errors.New("不支持的图片格式")
	}
	
	// 检查文件大小
	maxSize := p.config.MaxFileSize["image"]
	if file.Size > maxSize {
		return fmt.Errorf("图片文件过大，最大允许 %d MB", maxSize/(1024*1024))
	}
	
	return nil
}

// SupportedFormats 获取支持的格式
func (p *ImageProcessor) SupportedFormats() []string {
	if formats, ok := p.config.AllowedFormats["image"]; ok {
		return formats
	}
	return []string{"jpg", "jpeg", "png", "webp"}
}

// getColorMode 获取图片色彩模式
func getColorMode(img image.Image) string {
	switch img.ColorModel() {
	case color.RGBAModel:
		return "RGBA"
	case color.RGBA64Model:
		return "RGBA"
	case color.NRGBAModel:
		return "RGBA"
	case color.NRGBA64Model:
		return "RGBA"
	case color.AlphaModel:
		return "Alpha"
	case color.Alpha16Model:
		return "Alpha"
	case color.GrayModel:
		return "Grayscale"
	case color.Gray16Model:
		return "Grayscale"
	default:
		return "RGB"
	}
}

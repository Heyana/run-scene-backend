// Package utils 提供通用工具函数
package utils

import (
	"bytes"
	"fmt"
	"go_wails_project_manager/logger"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"strings"

	"github.com/HugoSmits86/nativewebp"
	"github.com/nfnt/resize"
)

// ImageProcessor 图片处理器
type ImageProcessor struct {
	maxWidth  uint // 最大宽度
	maxHeight uint // 最大高度
	quality   int  // 图片质量(0-100)
}

// NewImageProcessor 创建图片处理器实例
func NewImageProcessor() *ImageProcessor {
	return &ImageProcessor{
		maxWidth:  1000, // 默认最大宽度1000像素
		maxHeight: 1000, // 默认最大高度1000像素
		quality:   85,   // 默认质量85%
	}
}

// SetMaxWidth 设置最大宽度
func (p *ImageProcessor) SetMaxWidth(width uint) *ImageProcessor {
	p.maxWidth = width
	return p
}

// SetMaxHeight 设置最大高度
func (p *ImageProcessor) SetMaxHeight(height uint) *ImageProcessor {
	p.maxHeight = height
	return p
}

// SetQuality 设置图片质量
func (p *ImageProcessor) SetQuality(quality int) *ImageProcessor {
	if quality < 0 {
		quality = 0
	}
	if quality > 100 {
		quality = 100
	}
	p.quality = quality
	return p
}

// ProcessImage 处理图片
// data: 原始图片数据
// mimeType: 图片MIME类型
// 返回: 处理后的图片数据, 新的MIME类型, 错误
func (p *ImageProcessor) ProcessImage(data []byte, mimeType string) ([]byte, string, error) {
	// 解码图片
	img, format, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, mimeType, fmt.Errorf("解码图片失败: %w", err)
	}

	logger.Log.Debugf("原始图片格式: %s, 大小: %d 字节", format, len(data))

	// 调整大小
	resized := p.resizeImage(img)

	// 转换为WebP
	webpData, err := p.convertToWebP(resized)
	if err != nil {
		// WebP转换失败，尝试保持原格式
		return p.encodeOriginalFormat(resized, format)
	}

	logger.Log.Debugf("处理后图片格式: WebP, 大小: %d 字节", len(webpData))
	return webpData, "image/webp", nil
}

// resizeImage 调整图片大小
func (p *ImageProcessor) resizeImage(img image.Image) image.Image {
	bounds := img.Bounds()
	width := uint(bounds.Dx())
	height := uint(bounds.Dy())

	// 检查是否需要调整大小
	if width <= p.maxWidth && height <= p.maxHeight {
		return img // 不需要调整大小
	}

	// 计算宽高比
	ratio := float64(width) / float64(height)

	var newWidth, newHeight uint
	if width > height {
		// 横向图片
		newWidth = p.maxWidth
		newHeight = uint(float64(newWidth) / ratio)
		if newHeight > p.maxHeight {
			newHeight = p.maxHeight
			newWidth = uint(float64(newHeight) * ratio)
		}
	} else {
		// 纵向图片
		newHeight = p.maxHeight
		newWidth = uint(float64(newHeight) * ratio)
		if newWidth > p.maxWidth {
			newWidth = p.maxWidth
			newHeight = uint(float64(newWidth) / ratio)
		}
	}

	// 调整大小，使用Lanczos3插值算法(最高质量)
	return resize.Resize(newWidth, newHeight, img, resize.Lanczos3)
}

// convertToWebP 转换为WebP格式
func (p *ImageProcessor) convertToWebP(img image.Image) ([]byte, error) {
	// nativewebp使用选项
	var options *nativewebp.Options = nil

	// 创建输出缓冲区
	var buf bytes.Buffer

	// 转换为WebP (默认是lossless格式)
	err := nativewebp.Encode(&buf, img, options)
	if err != nil {
		return nil, fmt.Errorf("WebP编码失败: %w", err)
	}

	return buf.Bytes(), nil
}

// encodeOriginalFormat 以原始格式编码图片
func (p *ImageProcessor) encodeOriginalFormat(img image.Image, format string) ([]byte, string, error) {
	var buf bytes.Buffer
	var mimeType string

	format = strings.ToLower(format)
	switch format {
	case "jpeg", "jpg":
		err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: p.quality})
		mimeType = "image/jpeg"
		if err != nil {
			return nil, mimeType, err
		}
	case "png":
		err := png.Encode(&buf, img)
		mimeType = "image/png"
		if err != nil {
			return nil, mimeType, err
		}
	default:
		// 默认使用JPEG格式
		err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: p.quality})
		mimeType = "image/jpeg"
		if err != nil {
			return nil, mimeType, err
		}
	}

	return buf.Bytes(), mimeType, nil
}

// ProcessImageFromReader 从Reader处理图片
func (p *ImageProcessor) ProcessImageFromReader(reader io.Reader, mimeType string) ([]byte, string, error) {
	// 读取全部数据
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, mimeType, err
	}

	// 处理图片
	return p.ProcessImage(data, mimeType)
}

// GetDefaultProcessor 获取默认的图片处理器
func GetDefaultProcessor() *ImageProcessor {
	return NewImageProcessor()
}

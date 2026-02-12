// Package image 提供图片处理工具封装
package image

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// ImageMagick ImageMagick工具封装
type ImageMagick struct {
	binPath string
	timeout int
}

// NewImageMagick 创建ImageMagick实例
func NewImageMagick(binPath string, timeout int) *ImageMagick {
	if binPath == "" {
		binPath = "convert"
	}
	if timeout == 0 {
		timeout = 60
	}
	return &ImageMagick{
		binPath: binPath,
		timeout: timeout,
	}
}

// GetPath 获取可执行文件路径
func (im *ImageMagick) GetPath() string {
	return im.binPath
}

// CheckInstalled 检查是否已安装
func (im *ImageMagick) CheckInstalled() error {
	cmd := exec.Command(im.binPath, "-version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ImageMagick 未安装或路径错误: %w", err)
	}
	return nil
}

// ImageMetadata 图片元数据
type ImageMetadata struct {
	Width      int    `json:"width"`
	Height     int    `json:"height"`
	Format     string `json:"format"`
	FileSize   int64  `json:"file_size"`
	ColorSpace string `json:"color_space"`
	Depth      int    `json:"depth"`
}

// ExtractMetadata 提取图片元数据
func (im *ImageMagick) ExtractMetadata(filePath string) (*ImageMetadata, error) {
	// ImageMagick 7: magick identify -format "%w %h %b %m %[colorspace] %z" input.jpg
	cmd := exec.Command(im.binPath, "identify",
		"-format", "%w %h %b %m %[colorspace] %z",
		filePath,
	)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("identify 执行失败: %w", err)
	}

	// 解析输出
	parts := strings.Fields(string(output))
	if len(parts) < 6 {
		return nil, fmt.Errorf("identify 输出格式错误")
	}

	metadata := &ImageMetadata{}

	// 宽度
	metadata.Width, _ = strconv.Atoi(parts[0])

	// 高度
	metadata.Height, _ = strconv.Atoi(parts[1])

	// 文件大小
	sizeStr := parts[2]
	if strings.HasSuffix(sizeStr, "KB") {
		size, _ := strconv.ParseFloat(strings.TrimSuffix(sizeStr, "KB"), 64)
		metadata.FileSize = int64(size * 1024)
	} else if strings.HasSuffix(sizeStr, "MB") {
		size, _ := strconv.ParseFloat(strings.TrimSuffix(sizeStr, "MB"), 64)
		metadata.FileSize = int64(size * 1024 * 1024)
	}

	// 格式
	metadata.Format = parts[3]

	// 色彩空间
	metadata.ColorSpace = parts[4]

	// 位深度
	metadata.Depth, _ = strconv.Atoi(parts[5])

	return metadata, nil
}

// Resize 调整大小
func (im *ImageMagick) Resize(input, output string, width, height int, keepAspect bool) error {
	sizeStr := fmt.Sprintf("%dx%d", width, height)
	if keepAspect {
		sizeStr += ">" // 保持宽高比，不放大
	} else {
		sizeStr += "!" // 强制调整到指定大小
	}

	// ImageMagick 7: magick input.jpg -resize 800x600 output.jpg
	cmd := exec.Command(im.binPath,
		input,
		"-resize", sizeStr,
		output,
	)

	return cmd.Run()
}

// Convert 转换格式
func (im *ImageMagick) Convert(input, output string, format string, quality int) error {
	args := []string{input}

	// 质量
	if quality > 0 {
		args = append(args, "-quality", fmt.Sprintf("%d", quality))
	}

	// 输出格式
	if format != "" {
		args = append(args, "-format", format)
	}

	args = append(args, output)

	// ImageMagick 7: magick input.jpg -quality 90 output.png
	cmd := exec.Command(im.binPath, args...)
	return cmd.Run()
}

// GenerateThumbnail 生成缩略图（正方形，居中裁剪）
func (im *ImageMagick) GenerateThumbnail(input, output string, size int, quality int) error {
	// ImageMagick 7: magick input.jpg -thumbnail 256x256^ -gravity center -extent 256x256 -quality 85 output.jpg
	args := []string{
		input,
		"-thumbnail", fmt.Sprintf("%dx%d^", size, size), // ^ 表示填充
		"-gravity", "center",                             // 居中
		"-extent", fmt.Sprintf("%dx%d", size, size),     // 裁剪
		"-quality", fmt.Sprintf("%d", quality),
		output,
	}

	cmd := exec.Command(im.binPath, args...)

	// 捕获标准输出和错误输出
	output_bytes, err := cmd.CombinedOutput()
	
	if err != nil {
		return fmt.Errorf("ImageMagick执行失败: %w\n命令: %s %v\n输入文件: %s\n输出文件: %s\n输出信息: %s", 
			err, im.binPath, args, input, output, string(output_bytes))
	}

	return nil
}

// GenerateThumbnailKeepAspect 生成缩略图（保持宽高比，不裁剪，适用于 SVG）
func (im *ImageMagick) GenerateThumbnailKeepAspect(input, output string, size int, quality int) error {
	// ImageMagick 7: magick input.svg -background white -flatten -resize 256x256 -quality 85 output.webp
	args := []string{
		input,
		"-background", "white",                      // 设置背景色为白色
		"-flatten",                                  // 展平图层
		"-resize", fmt.Sprintf("%dx%d", size, size), // 保持宽高比缩放
		"-quality", fmt.Sprintf("%d", quality),
		output,
	}

	cmd := exec.Command(im.binPath, args...)

	// 捕获标准输出和错误输出
	output_bytes, err := cmd.CombinedOutput()
	
	if err != nil {
		return fmt.Errorf("ImageMagick执行失败: %w\n命令: %s %v\n输入文件: %s\n输出文件: %s\n输出信息: %s", 
			err, im.binPath, args, input, output, string(output_bytes))
	}

	return nil
}

// Crop 裁剪图片
func (im *ImageMagick) Crop(input, output string, x, y, width, height int) error {
	// ImageMagick 7: magick input.jpg -crop 800x600+100+50 +repage output.jpg
	cmd := exec.Command(im.binPath,
		input,
		"-crop", fmt.Sprintf("%dx%d+%d+%d", width, height, x, y),
		"+repage", // 重置画布
		output,
	)

	return cmd.Run()
}

// Rotate 旋转图片
func (im *ImageMagick) Rotate(input, output string, angle int) error {
	// ImageMagick 7: magick input.jpg -rotate 90 output.jpg
	cmd := exec.Command(im.binPath,
		input,
		"-rotate", fmt.Sprintf("%d", angle),
		output,
	)

	return cmd.Run()
}

// AddWatermark 添加水印
func (im *ImageMagick) AddWatermark(input, watermark, output string, position string, opacity int) error {
	// position: NorthWest, North, NorthEast, West, Center, East, SouthWest, South, SouthEast
	// ImageMagick 7: magick input.jpg watermark.png -gravity Center -compose over -composite output.jpg
	cmd := exec.Command(im.binPath,
		input,
		watermark,
		"-gravity", position,
		"-compose", "over",
		"-define", fmt.Sprintf("compose:args=%d", opacity),
		"-composite",
		output,
	)

	return cmd.Run()
}

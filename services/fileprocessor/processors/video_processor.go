// Package processors 文件处理器实现
package processors

import (
	"context"
	"fmt"
	"go_wails_project_manager/utils/video"
	"path/filepath"
)

// VideoProcessor 视频处理器
type VideoProcessor struct {
	ffmpeg *video.FFmpeg
}

// NewVideoProcessor 创建视频处理器
func NewVideoProcessor(ffmpeg *video.FFmpeg) *VideoProcessor {
	return &VideoProcessor{
		ffmpeg: ffmpeg,
	}
}

// Name 获取处理器名称
func (p *VideoProcessor) Name() string {
	return "VideoProcessor"
}

// Support 检查是否支持该文件格式
func (p *VideoProcessor) Support(format string) bool {
	supported := []string{
		"mp4", "avi", "mov", "webm", "mkv",
		"flv", "wmv", "m4v", "mpg", "mpeg",
	}

	for _, f := range supported {
		if f == format {
			return true
		}
	}
	return false
}

// FileMetadata 文件元数据（从父包复制）
type FileMetadata struct {
	FileName  string                 `json:"file_name"`
	Format    string                 `json:"format"`
	Width     int                    `json:"width"`
	Height    int                    `json:"height"`
	Duration  int64                  `json:"duration"`
	Bitrate   int64                  `json:"bitrate"`
	Codec     string                 `json:"codec"`
	FrameRate float64                `json:"frame_rate"`
	FileSize  int64                  `json:"file_size"`
	PageCount int                    `json:"page_count"`
	Author    string                 `json:"author"`
	Title     string                 `json:"title"`
	Extra     map[string]interface{} `json:"extra"`
}

// PreviewOptions 预览图生成选项
type PreviewOptions struct {
	Width      int    `json:"width"`
	Height     int    `json:"height"`
	OutputPath string `json:"output_path"`
	MaxPages   int    `json:"max_pages"`
}

// PreviewResult 预览图生成结果
type PreviewResult struct {
	ThumbnailPath string   `json:"thumbnail_path"`
	PreviewPaths  []string `json:"preview_paths"`
}

// ThumbnailOptions 缩略图生成选项
type ThumbnailOptions struct {
	Size       int    `json:"size"`
	Quality    int    `json:"quality"`
	OutputPath string `json:"output_path"`
}

// ConvertOptions 格式转换选项
type ConvertOptions struct {
	OutputPath string  `json:"output_path"`
	Format     string  `json:"format"`
	Codec      string  `json:"codec"`
	Bitrate    int64   `json:"bitrate"`
	Width      int     `json:"width"`
	Height     int     `json:"height"`
	Quality    int     `json:"quality"`
	FrameRate  float64 `json:"frame_rate"`
}

// ExtractMetadata 提取文件元数据
func (p *VideoProcessor) ExtractMetadata(filePath string) (*FileMetadata, error) {
	videoMeta, err := p.ffmpeg.ExtractMetadata(filePath)
	if err != nil {
		return nil, err
	}

	return &FileMetadata{
		FileName:  filepath.Base(filePath),
		Format:    filepath.Ext(filePath),
		Width:     videoMeta.Width,
		Height:    videoMeta.Height,
		Duration:  int64(videoMeta.Duration),
		Bitrate:   videoMeta.Bitrate,
		Codec:     videoMeta.Codec,
		FrameRate: videoMeta.FrameRate,
		FileSize:  videoMeta.FileSize,
		Extra: map[string]interface{}{
			"audio_codec":   videoMeta.AudioCodec,
			"audio_bitrate": videoMeta.AudioBitrate,
		},
	}, nil
}

// GeneratePreview 生成预览图
func (p *VideoProcessor) GeneratePreview(filePath string, options PreviewOptions) (*PreviewResult, error) {
	// 生成多个时间点的预览图
	timeOffsets := []int{1, 5, 10, 30, 60} // 秒

	var previewPaths []string
	var thumbnailPath string

	for i, offset := range timeOffsets {
		outputPath := fmt.Sprintf("%s_preview_%d.jpg", options.OutputPath, i)

		err := p.ffmpeg.GenerateThumbnail(filePath, outputPath, offset)
		if err != nil {
			continue
		}

		previewPaths = append(previewPaths, outputPath)

		if i == 0 {
			thumbnailPath = outputPath
		}
	}

	if len(previewPaths) == 0 {
		return nil, fmt.Errorf("生成预览图失败")
	}

	return &PreviewResult{
		ThumbnailPath: thumbnailPath,
		PreviewPaths:  previewPaths,
	}, nil
}

// GenerateThumbnail 生成缩略图
func (p *VideoProcessor) GenerateThumbnail(filePath string, options ThumbnailOptions) (string, error) {
	// 添加日志
	fmt.Printf("[VideoProcessor] GenerateThumbnail 开始\n")
	fmt.Printf("[VideoProcessor] 输入文件: %s\n", filePath)
	fmt.Printf("[VideoProcessor] 输出文件: %s\n", options.OutputPath)
	fmt.Printf("[VideoProcessor] 尺寸: %d, 质量: %d\n", options.Size, options.Quality)

	err := p.ffmpeg.GenerateThumbnail(filePath, options.OutputPath, 1)
	if err != nil {
		fmt.Printf("[VideoProcessor] 生成缩略图失败: %v\n", err)
		return "", err
	}

	fmt.Printf("[VideoProcessor] 生成缩略图成功: %s\n", options.OutputPath)
	return options.OutputPath, nil
}

// Convert 格式转换
func (p *VideoProcessor) Convert(filePath string, options ConvertOptions) (string, error) {
	ctx := context.Background()

	convertOpts := video.ConvertOptions{
		Codec:   options.Codec,
		Bitrate: options.Bitrate,
		Width:   options.Width,
		Height:  options.Height,
		Quality: options.Quality,
	}

	err := p.ffmpeg.Convert(ctx, filePath, options.OutputPath, convertOpts, nil)
	if err != nil {
		return "", err
	}

	return options.OutputPath, nil
}

// Validate 验证文件完整性
func (p *VideoProcessor) Validate(filePath string) error {
	// 尝试提取元数据来验证文件
	_, err := p.ffmpeg.ExtractMetadata(filePath)
	return err
}

package processors

import (
	"encoding/json"
	"errors"
	"fmt"
	"go_wails_project_manager/config"
	"go_wails_project_manager/models"
	"mime/multipart"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// VideoProcessor 视频处理器
type VideoProcessor struct {
	config *config.AssetConfig
}

// NewVideoProcessor 创建视频处理器
func NewVideoProcessor(cfg *config.AssetConfig) *VideoProcessor {
	return &VideoProcessor{
		config: cfg,
	}
}

// GenerateThumbnail 生成缩略图
func (p *VideoProcessor) GenerateThumbnail(filePath, outputPath string) error {
	// 使用FFmpeg截取视频帧
	cmd := exec.Command(
		p.config.FFmpegPath,
		"-i", filePath,
		"-ss", fmt.Sprintf("%.1f", p.config.VideoThumbnailTime),
		"-vframes", "1",
		"-vf", fmt.Sprintf("scale=%d:%d:force_original_aspect_ratio=decrease", 
			p.config.ThumbnailWidth, p.config.ThumbnailHeight),
		"-y",
		outputPath,
	)
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("FFmpeg生成缩略图失败: %w, output: %s", err, string(output))
	}
	
	return nil
}

// ExtractMetadata 提取元数据
func (p *VideoProcessor) ExtractMetadata(filePath string) (*models.AssetMetadata, error) {
	// 使用FFprobe提取视频信息
	cmd := exec.Command(
		"ffprobe",
		"-v", "quiet",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		filePath,
	)
	
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("FFprobe提取元数据失败: %w", err)
	}
	
	var info VideoInfo
	if err := json.Unmarshal(output, &info); err != nil {
		return nil, fmt.Errorf("解析视频信息失败: %w", err)
	}
	
	// 查找视频流
	var videoStream *Stream
	for i := range info.Streams {
		if info.Streams[i].CodecType == "video" {
			videoStream = &info.Streams[i]
			break
		}
	}
	
	if videoStream == nil {
		return nil, errors.New("未找到视频流")
	}
	
	// 解析时长
	duration, _ := strconv.ParseFloat(info.Format.Duration, 64)
	
	// 解析帧率
	frameRate := parseFrameRate(videoStream.AvgFrameRate)
	
	// 解析比特率
	bitrate, _ := strconv.Atoi(info.Format.BitRate)
	
	metadata := &models.AssetMetadata{
		Width:     videoStream.Width,
		Height:    videoStream.Height,
		Duration:  duration,
		FrameRate: frameRate,
		Codec:     videoStream.CodecName,
		Bitrate:   bitrate / 1000, // 转换为kbps
	}
	
	return metadata, nil
}

// Validate 验证文件
func (p *VideoProcessor) Validate(file *multipart.FileHeader) error {
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
		return errors.New("不支持的视频格式")
	}
	
	// 检查文件大小
	maxSize := p.config.MaxFileSize["video"]
	if file.Size > maxSize {
		return fmt.Errorf("视频文件过大，最大允许 %d MB", maxSize/(1024*1024))
	}
	
	return nil
}

// SupportedFormats 获取支持的格式
func (p *VideoProcessor) SupportedFormats() []string {
	if formats, ok := p.config.AllowedFormats["video"]; ok {
		return formats
	}
	return []string{"mp4", "webm"}
}

// VideoInfo FFprobe输出结构
type VideoInfo struct {
	Streams []Stream `json:"streams"`
	Format  Format   `json:"format"`
}

// Stream 视频流信息
type Stream struct {
	CodecType    string `json:"codec_type"`
	CodecName    string `json:"codec_name"`
	Width        int    `json:"width"`
	Height       int    `json:"height"`
	AvgFrameRate string `json:"avg_frame_rate"`
}

// Format 格式信息
type Format struct {
	Duration string `json:"duration"`
	BitRate  string `json:"bit_rate"`
}

// parseFrameRate 解析帧率字符串 (例如 "30/1" -> 30.0)
func parseFrameRate(frameRateStr string) float64 {
	parts := strings.Split(frameRateStr, "/")
	if len(parts) != 2 {
		return 0
	}
	
	numerator, err1 := strconv.ParseFloat(parts[0], 64)
	denominator, err2 := strconv.ParseFloat(parts[1], 64)
	
	if err1 != nil || err2 != nil || denominator == 0 {
		return 0
	}
	
	return numerator / denominator
}

// Package video 提供视频处理工具封装
package video

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// FFmpeg FFmpeg工具封装
type FFmpeg struct {
	binPath string
	timeout time.Duration
}

// NewFFmpeg 创建FFmpeg实例
func NewFFmpeg(binPath string, timeout int) *FFmpeg {
	if binPath == "" {
		binPath = "ffmpeg"
	}
	if timeout == 0 {
		timeout = 300
	}
	return &FFmpeg{
		binPath: binPath,
		timeout: time.Duration(timeout) * time.Second,
	}
}

// GetPath 获取可执行文件路径
func (f *FFmpeg) GetPath() string {
	return f.binPath
}

// CheckInstalled 检查是否已安装
func (f *FFmpeg) CheckInstalled() error {
	cmd := exec.Command(f.binPath, "-version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("FFmpeg 未安装或路径错误: %w", err)
	}
	return nil
}

// VideoMetadata 视频元数据
type VideoMetadata struct {
	Width        int     `json:"width"`
	Height       int     `json:"height"`
	Duration     float64 `json:"duration"`      // 秒
	Bitrate      int64   `json:"bitrate"`       // bps
	Codec        string  `json:"codec"`
	FrameRate    float64 `json:"frame_rate"`
	AudioCodec   string  `json:"audio_codec"`
	AudioBitrate int64   `json:"audio_bitrate"`
	FileSize     int64   `json:"file_size"`
}

// ExtractMetadata 提取视频元数据
func (f *FFmpeg) ExtractMetadata(filePath string) (*VideoMetadata, error) {
	// 使用 ffprobe 提取元数据
	cmd := exec.Command("ffprobe",
		"-v", "quiet",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		filePath,
	)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("ffprobe 执行失败: %w", err)
	}

	// 解析 JSON
	var result struct {
		Format struct {
			Duration string `json:"duration"`
			Bitrate  string `json:"bit_rate"`
			Size     string `json:"size"`
		} `json:"format"`
		Streams []struct {
			CodecType  string `json:"codec_type"`
			CodecName  string `json:"codec_name"`
			Width      int    `json:"width"`
			Height     int    `json:"height"`
			RFrameRate string `json:"r_frame_rate"`
			Bitrate    string `json:"bit_rate"`
		} `json:"streams"`
	}

	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("解析 ffprobe 输出失败: %w", err)
	}

	metadata := &VideoMetadata{}

	// 解析时长
	if duration, err := strconv.ParseFloat(result.Format.Duration, 64); err == nil {
		metadata.Duration = duration
	}

	// 解析比特率
	if bitrate, err := strconv.ParseInt(result.Format.Bitrate, 10, 64); err == nil {
		metadata.Bitrate = bitrate
	}

	// 解析文件大小
	if size, err := strconv.ParseInt(result.Format.Size, 10, 64); err == nil {
		metadata.FileSize = size
	}

	// 解析流信息
	for _, stream := range result.Streams {
		if stream.CodecType == "video" {
			metadata.Width = stream.Width
			metadata.Height = stream.Height
			metadata.Codec = stream.CodecName

			// 解析帧率
			if parts := strings.Split(stream.RFrameRate, "/"); len(parts) == 2 {
				num, _ := strconv.ParseFloat(parts[0], 64)
				den, _ := strconv.ParseFloat(parts[1], 64)
				if den != 0 {
					metadata.FrameRate = num / den
				}
			}
		} else if stream.CodecType == "audio" {
			metadata.AudioCodec = stream.CodecName
			if bitrate, err := strconv.ParseInt(stream.Bitrate, 10, 64); err == nil {
				metadata.AudioBitrate = bitrate
			}
		}
	}

	return metadata, nil
}

// GenerateThumbnail 生成缩略图（简单版本）
func (f *FFmpeg) GenerateThumbnail(input, output string, timeOffset int) error {
	args := []string{
		"-i", input,
		"-ss", fmt.Sprintf("%d", timeOffset),
		"-vframes", "1",
		"-q:v", "2", // 质量
		"-y",
		output,
	}

	cmd := exec.Command(f.binPath, args...)

	// 捕获标准输出和错误输出
	output_bytes, err := cmd.CombinedOutput()

	if err != nil {
		return fmt.Errorf("生成缩略图失败: %w\n命令: %s %v\n输入文件: %s\n输出文件: %s\n输出信息: %s",
			err, f.binPath, args, input, output, string(output_bytes))
	}

	return nil
}

// GenerateThumbnailWithContext 生成缩略图（支持取消和进度）
func (f *FFmpeg) GenerateThumbnailWithContext(
	ctx context.Context,
	input, output string,
	timeOffset int,
	progressCallback func(float64),
) error {
	cmd := exec.CommandContext(ctx, f.binPath,
		"-i", input,
		"-ss", fmt.Sprintf("%d", timeOffset),
		"-vframes", "1",
		"-q:v", "2",
		"-progress", "pipe:1", // 输出进度到 stdout
		"-y",
		output,
	)

	// 捕获输出
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	// 解析进度
	if progressCallback != nil {
		go f.parseProgress(stdout, progressCallback)
	}

	// 捕获错误输出
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			// 记录错误日志
			fmt.Println("FFmpeg:", scanner.Text())
		}
	}()

	return cmd.Wait()
}

// parseProgress 解析 FFmpeg 进度输出
func (f *FFmpeg) parseProgress(reader io.Reader, callback func(float64)) {
	scanner := bufio.NewScanner(reader)

	var totalDuration float64
	var currentTime float64

	for scanner.Scan() {
		line := scanner.Text()

		// 解析总时长
		if strings.HasPrefix(line, "Duration: ") {
			re := regexp.MustCompile(`Duration: (\d+):(\d+):(\d+\.\d+)`)
			matches := re.FindStringSubmatch(line)
			if len(matches) == 4 {
				h, _ := strconv.ParseFloat(matches[1], 64)
				m, _ := strconv.ParseFloat(matches[2], 64)
				s, _ := strconv.ParseFloat(matches[3], 64)
				totalDuration = h*3600 + m*60 + s
			}
		}

		// 解析当前时间
		if strings.HasPrefix(line, "out_time_ms=") {
			timeStr := strings.TrimPrefix(line, "out_time_ms=")
			if timeMs, err := strconv.ParseInt(timeStr, 10, 64); err == nil {
				currentTime = float64(timeMs) / 1000000.0

				if totalDuration > 0 {
					progress := (currentTime / totalDuration) * 100
					if progress > 100 {
						progress = 100
					}
					callback(progress)
				}
			}
		}
	}
}

// ConvertOptions 转换选项
type ConvertOptions struct {
	Codec      string  // 视频编码器
	Bitrate    int64   // 比特率
	Width      int     // 宽度
	Height     int     // 高度
	FrameRate  float64 // 帧率
	AudioCodec string  // 音频编码器
	Quality    int     // 质量 (0-51, 越小越好)
}

// Convert 转换视频格式
func (f *FFmpeg) Convert(
	ctx context.Context,
	input, output string,
	options ConvertOptions,
	progressCallback func(float64),
) error {
	args := []string{"-i", input}

	// 视频编码
	if options.Codec != "" {
		args = append(args, "-c:v", options.Codec)
	}

	// 比特率
	if options.Bitrate > 0 {
		args = append(args, "-b:v", fmt.Sprintf("%d", options.Bitrate))
	}

	// 分辨率
	if options.Width > 0 && options.Height > 0 {
		args = append(args, "-s", fmt.Sprintf("%dx%d", options.Width, options.Height))
	}

	// 帧率
	if options.FrameRate > 0 {
		args = append(args, "-r", fmt.Sprintf("%.2f", options.FrameRate))
	}

	// 质量
	if options.Quality > 0 {
		args = append(args, "-crf", fmt.Sprintf("%d", options.Quality))
	}

	// 音频编码
	if options.AudioCodec != "" {
		args = append(args, "-c:a", options.AudioCodec)
	}

	// 进度输出
	args = append(args, "-progress", "pipe:1")

	// 覆盖输出
	args = append(args, "-y", output)

	cmd := exec.CommandContext(ctx, f.binPath, args...)

	// 处理进度
	stdout, _ := cmd.StdoutPipe()
	if err := cmd.Start(); err != nil {
		return err
	}

	if progressCallback != nil {
		go f.parseProgress(stdout, progressCallback)
	}

	return cmd.Wait()
}

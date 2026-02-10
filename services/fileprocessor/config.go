// Package fileprocessor 文件处理器配置
package fileprocessor

// Config 文件处理器配置
type Config struct {
	FFmpeg      FFmpegConfig
	ImageMagick ImageMagickConfig
	PDF         PDFConfig
	Blender     BlenderConfig
	Thumbnail   ThumbnailConfig
	Task        TaskConfig
	Resource    ResourceConfig
}

// FFmpegConfig FFmpeg配置
type FFmpegConfig struct {
	BinPath string // 可执行文件路径
	Timeout int    // 超时时间（秒）
}

// ImageMagickConfig ImageMagick配置
type ImageMagickConfig struct {
	BinPath string // 可执行文件路径
	Timeout int    // 超时时间（秒）
}

// PDFConfig PDF工具配置
type PDFConfig struct {
	BinPath string // 可执行文件路径
	Timeout int    // 超时时间（秒）
}

// BlenderConfig Blender配置
type BlenderConfig struct {
	BinPath    string // 可执行文件路径
	ScriptPath string // 渲染脚本路径
	Timeout    int    // 超时时间（秒）
}

// ThumbnailConfig 预览图配置
type ThumbnailConfig struct {
	Format  string // 预览图格式：webp, jpg, png
	Width   int    // 预览图宽度
	Height  int    // 预览图高度
	Quality int    // 预览图质量（0-100）
}

// TaskConfig 任务配置
type TaskConfig struct {
	MaxConcurrent int // 最大并发数
	MaxRetries    int // 最大重试次数
	RetryDelay    int // 重试延迟（秒）
	CleanupAfter  int // 清理时间（秒）
}

// ResourceConfig 资源限制配置
type ResourceConfig struct {
	MaxMemoryPerTask int64   // 单个任务最大内存（字节）
	MaxCPUPercent    float64 // 最大CPU使用率
	MaxTempSize      int64   // 最大临时文件大小（字节）
}

// DefaultConfig 默认配置
func DefaultConfig() *Config {
	return &Config{
		FFmpeg: FFmpegConfig{
			BinPath: "ffmpeg",
			Timeout: 300,
		},
		ImageMagick: ImageMagickConfig{
			BinPath: "convert",
			Timeout: 60,
		},
		PDF: PDFConfig{
			BinPath: "pdftoppm",
			Timeout: 120,
		},
		Blender: BlenderConfig{
			BinPath:    "blender",
			ScriptPath: "deploy/scripts/render_fbx.py",
			Timeout:    300,
		},
		Thumbnail: ThumbnailConfig{
			Format:  "webp",
			Width:   1280,
			Height:  720,
			Quality: 85,
		},
		Task: TaskConfig{
			MaxConcurrent: 5,
			MaxRetries:    3,
			RetryDelay:    60,
			CleanupAfter:  86400,
		},
		Resource: ResourceConfig{
			MaxMemoryPerTask: 2147483648,  // 2GB
			MaxCPUPercent:    80.0,
			MaxTempSize:      10737418240, // 10GB
		},
	}
}

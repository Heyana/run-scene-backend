// Package model 3D 模型处理工具
package model

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"
)

// Blender Blender 工具包装器
type Blender struct {
	binPath    string
	scriptPath string
	timeout    time.Duration
}

// NewBlender 创建 Blender 工具实例
func NewBlender(binPath, scriptPath string, timeout time.Duration) *Blender {
	if binPath == "" {
		binPath = "blender"
	}
	if scriptPath == "" {
		scriptPath = "deploy/scripts/render_fbx.py"
	}
	if timeout == 0 {
		timeout = 300 * time.Second
	}
	return &Blender{
		binPath:    binPath,
		scriptPath: scriptPath,
		timeout:    timeout,
	}
}

// RenderOptions 渲染选项
type RenderOptions struct {
	Width   int    // 宽度
	Height  int    // 高度
	Quality string // 质量: fast, normal, high
}

// RenderPreview 渲染 3D 模型预览图
func (b *Blender) RenderPreview(ctx context.Context, inputPath, outputPath string, options RenderOptions) error {
	// 构建命令参数
	args := []string{
		"-b",                 // 后台模式
		"-P", b.scriptPath,   // Python 脚本
		"--",                 // 分隔符
		inputPath,            // 输入文件
		outputPath,           // 输出文件
	}

	// 添加可选参数
	if options.Width > 0 {
		args = append(args, strconv.Itoa(options.Width))
	}
	if options.Height > 0 {
		args = append(args, strconv.Itoa(options.Height))
	}
	if options.Quality != "" {
		args = append(args, options.Quality)
	}

	// 创建上下文
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), b.timeout)
		defer cancel()
	}

	// 执行命令
	cmd := exec.CommandContext(ctx, b.binPath, args...)
	output, err := cmd.CombinedOutput()

	if err != nil {
		return fmt.Errorf("Blender 渲染失败: %v, 输出: %s", err, string(output))
	}

	fmt.Printf("[Blender] 渲染成功: %s -> %s\n", inputPath, outputPath)
	fmt.Printf("[Blender] 输出: %s\n", string(output))

	return nil
}

// ModelMetadata 3D 模型元数据
type ModelMetadata struct {
	Format   string  `json:"format"`
	FileSize int64   `json:"file_size"`
	Vertices int     `json:"vertices"`
	Faces    int     `json:"faces"`
	Objects  int     `json:"objects"`
	BoundBox [6]float64 `json:"bound_box"` // [min_x, min_y, min_z, max_x, max_y, max_z]
}

// ExtractMetadata 提取 3D 模型元数据（简化版，仅返回基本信息）
func (b *Blender) ExtractMetadata(filePath string) (*ModelMetadata, error) {
	// 获取文件信息
	ext := filepath.Ext(filePath)
	
	// 简化版：只返回格式信息
	// 完整实现需要运行 Blender Python 脚本提取详细信息
	return &ModelMetadata{
		Format: ext,
	}, nil
}

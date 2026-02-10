// Package document LibreOffice 文档处理工具
package document

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// LibreOffice LibreOffice 工具包装器
type LibreOffice struct {
	binPath string
	timeout time.Duration
}

// NewLibreOffice 创建 LibreOffice 工具实例
func NewLibreOffice(binPath string, timeout time.Duration) *LibreOffice {
	if binPath == "" {
		binPath = "libreoffice"
	}
	if timeout == 0 {
		timeout = 120 * time.Second
	}
	return &LibreOffice{
		binPath: binPath,
		timeout: timeout,
	}
}

// ConvertToPNG 将文档转换为 PNG 图片（第一页）
func (lo *LibreOffice) ConvertToPNG(ctx context.Context, inputPath, outputDir string) (string, error) {
	// 确保输出目录存在
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", fmt.Errorf("创建输出目录失败: %w", err)
	}

	// 创建上下文
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), lo.timeout)
		defer cancel()
	}

	// LibreOffice 命令：转换为 PNG
	// --headless: 无界面模式
	// --convert-to png: 转换为 PNG 格式
	// --outdir: 输出目录
	args := []string{
		"--headless",
		"--convert-to", "png",
		"--outdir", outputDir,
		inputPath,
	}

	cmd := exec.CommandContext(ctx, lo.binPath, args...)
	output, err := cmd.CombinedOutput()

	if err != nil {
		return "", fmt.Errorf("LibreOffice 转换失败: %v, 输出: %s", err, string(output))
	}

	// 生成的文件名：原文件名.png
	baseName := filepath.Base(inputPath)
	ext := filepath.Ext(baseName)
	nameWithoutExt := baseName[:len(baseName)-len(ext)]
	outputFile := filepath.Join(outputDir, nameWithoutExt+".png")

	// 检查文件是否生成
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		return "", fmt.Errorf("转换后的文件不存在: %s", outputFile)
	}

	fmt.Printf("[LibreOffice] 转换成功: %s -> %s\n", inputPath, outputFile)
	fmt.Printf("[LibreOffice] 输出: %s\n", string(output))

	return outputFile, nil
}

// ConvertToPDF 将文档转换为 PDF
func (lo *LibreOffice) ConvertToPDF(ctx context.Context, inputPath, outputDir string) (string, error) {
	// 确保输出目录存在
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", fmt.Errorf("创建输出目录失败: %w", err)
	}

	// 创建上下文
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), lo.timeout)
		defer cancel()
	}

	// LibreOffice 命令：转换为 PDF
	args := []string{
		"--headless",
		"--convert-to", "pdf",
		"--outdir", outputDir,
		inputPath,
	}

	cmd := exec.CommandContext(ctx, lo.binPath, args...)
	output, err := cmd.CombinedOutput()

	if err != nil {
		return "", fmt.Errorf("LibreOffice 转换失败: %v, 输出: %s", err, string(output))
	}

	// 生成的文件名：原文件名.pdf
	baseName := filepath.Base(inputPath)
	ext := filepath.Ext(baseName)
	nameWithoutExt := baseName[:len(baseName)-len(ext)]
	outputFile := filepath.Join(outputDir, nameWithoutExt+".pdf")

	// 检查文件是否生成
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		return "", fmt.Errorf("转换后的文件不存在: %s", outputFile)
	}

	fmt.Printf("[LibreOffice] 转换成功: %s -> %s\n", inputPath, outputFile)

	return outputFile, nil
}

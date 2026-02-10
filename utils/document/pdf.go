// Package document 提供文档处理工具封装
package document

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// PDFTool PDF工具封装
type PDFTool struct {
	binPath string
	timeout int
}

// NewPDFTool 创建PDFTool实例
func NewPDFTool(binPath string, timeout int) *PDFTool {
	if binPath == "" {
		binPath = "pdftoppm"
	}
	if timeout == 0 {
		timeout = 120
	}
	return &PDFTool{
		binPath: binPath,
		timeout: timeout,
	}
}

// GetPath 获取可执行文件路径
func (p *PDFTool) GetPath() string {
	return p.binPath
}

// CheckInstalled 检查是否已安装
func (p *PDFTool) CheckInstalled() error {
	cmd := exec.Command("pdfinfo", "-v")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("PDF工具 未安装或路径错误: %w", err)
	}
	return nil
}

// PDFMetadata PDF 元数据
type PDFMetadata struct {
	PageCount int    `json:"page_count"`
	Title     string `json:"title"`
	Author    string `json:"author"`
	Subject   string `json:"subject"`
	Creator   string `json:"creator"`
}

// ExtractMetadata 提取 PDF 元数据
func (p *PDFTool) ExtractMetadata(filePath string) (*PDFMetadata, error) {
	// pdfinfo input.pdf
	cmd := exec.Command("pdfinfo", filePath)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("pdfinfo 执行失败: %w", err)
	}

	metadata := &PDFMetadata{}
	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "Pages":
			metadata.PageCount, _ = strconv.Atoi(value)
		case "Title":
			metadata.Title = value
		case "Author":
			metadata.Author = value
		case "Subject":
			metadata.Subject = value
		case "Creator":
			metadata.Creator = value
		}
	}

	return metadata, nil
}

// GeneratePreview 生成 PDF 预览图
func (p *PDFTool) GeneratePreview(input, outputPrefix string, pages []int, dpi int) ([]string, error) {
	// pdftoppm -png -r 150 -f 1 -l 5 input.pdf output
	args := []string{
		"-png",
		"-r", fmt.Sprintf("%d", dpi),
	}

	if len(pages) > 0 {
		args = append(args, "-f", fmt.Sprintf("%d", pages[0]))
		args = append(args, "-l", fmt.Sprintf("%d", pages[len(pages)-1]))
	}

	args = append(args, input, outputPrefix)

	cmd := exec.Command(p.binPath, args...)
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("生成预览失败: %w", err)
	}

	// 返回生成的文件列表
	var outputs []string
	for _, page := range pages {
		outputs = append(outputs, fmt.Sprintf("%s-%d.png", outputPrefix, page))
	}

	return outputs, nil
}

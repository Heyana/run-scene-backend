// Package processors 文件处理器实现
package processors

import (
	"fmt"
	"go_wails_project_manager/utils/document"
	"path/filepath"
)

// DocumentProcessor 文档处理器
type DocumentProcessor struct {
	pdftool *document.PDFTool
}

// NewDocumentProcessor 创建文档处理器
func NewDocumentProcessor(pdftool *document.PDFTool) *DocumentProcessor {
	return &DocumentProcessor{
		pdftool: pdftool,
	}
}

// Name 获取处理器名称
func (p *DocumentProcessor) Name() string {
	return "DocumentProcessor"
}

// Support 检查是否支持该文件格式
func (p *DocumentProcessor) Support(format string) bool {
	supported := []string{
		"pdf", "doc", "docx", "ppt", "pptx",
		"xls", "xlsx", "txt", "md",
	}

	for _, f := range supported {
		if f == format {
			return true
		}
	}
	return false
}

// ExtractMetadata 提取文件元数据
func (p *DocumentProcessor) ExtractMetadata(filePath string) (*FileMetadata, error) {
	// 目前只支持 PDF
	ext := filepath.Ext(filePath)
	if ext != ".pdf" {
		return &FileMetadata{
			FileName: filepath.Base(filePath),
			Format:   ext,
		}, nil
	}

	pdfMeta, err := p.pdftool.ExtractMetadata(filePath)
	if err != nil {
		return nil, err
	}

	return &FileMetadata{
		FileName:  filepath.Base(filePath),
		Format:    "pdf",
		PageCount: pdfMeta.PageCount,
		Author:    pdfMeta.Author,
		Title:     pdfMeta.Title,
		Extra: map[string]interface{}{
			"subject": pdfMeta.Subject,
			"creator": pdfMeta.Creator,
		},
	}, nil
}

// GeneratePreview 生成预览图
func (p *DocumentProcessor) GeneratePreview(filePath string, options PreviewOptions) (*PreviewResult, error) {
	// 目前只支持 PDF
	ext := filepath.Ext(filePath)
	if ext != ".pdf" {
		return nil, fmt.Errorf("不支持的文档格式: %s", ext)
	}

	// 生成前几页的预览
	maxPages := options.MaxPages
	if maxPages == 0 {
		maxPages = 5
	}

	pages := make([]int, maxPages)
	for i := 0; i < maxPages; i++ {
		pages[i] = i + 1
	}

	previewPaths, err := p.pdftool.GeneratePreview(
		filePath,
		options.OutputPath,
		pages,
		150, // DPI
	)

	if err != nil {
		return nil, err
	}

	thumbnailPath := ""
	if len(previewPaths) > 0 {
		thumbnailPath = previewPaths[0]
	}

	return &PreviewResult{
		ThumbnailPath: thumbnailPath,
		PreviewPaths:  previewPaths,
	}, nil
}

// GenerateThumbnail 生成缩略图
func (p *DocumentProcessor) GenerateThumbnail(filePath string, options ThumbnailOptions) (string, error) {
	// 生成第一页作为缩略图
	previewPaths, err := p.pdftool.GeneratePreview(
		filePath,
		options.OutputPath,
		[]int{1},
		150,
	)

	if err != nil {
		return "", err
	}

	if len(previewPaths) == 0 {
		return "", fmt.Errorf("生成缩略图失败")
	}

	return previewPaths[0], nil
}

// Convert 格式转换
func (p *DocumentProcessor) Convert(filePath string, options ConvertOptions) (string, error) {
	return "", fmt.Errorf("文档转换功能未实现")
}

// Validate 验证文件完整性
func (p *DocumentProcessor) Validate(filePath string) error {
	// 尝试提取元数据来验证文件
	_, err := p.pdftool.ExtractMetadata(filePath)
	return err
}

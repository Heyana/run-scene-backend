// Package processors 文件处理器实现
package processors

import (
	"context"
	"fmt"
	"go_wails_project_manager/utils/document"
	"go_wails_project_manager/utils/image"
	"os"
	"path/filepath"
	"strings"
)

// DocumentProcessor 文档处理器
type DocumentProcessor struct {
	pdftool     *document.PDFTool
	libreoffice *document.LibreOffice
	imagemagick *image.ImageMagick
}

// NewDocumentProcessor 创建文档处理器
func NewDocumentProcessor(pdftool *document.PDFTool, libreoffice *document.LibreOffice, imagemagick *image.ImageMagick) *DocumentProcessor {
	return &DocumentProcessor{
		pdftool:     pdftool,
		libreoffice: libreoffice,
		imagemagick: imagemagick,
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
	fmt.Printf("[DocumentProcessor] GenerateThumbnail 开始\n")
	fmt.Printf("[DocumentProcessor] 输入文件: %s\n", filePath)
	fmt.Printf("[DocumentProcessor] 输出文件: %s\n", options.OutputPath)

	ext := strings.ToLower(filepath.Ext(filePath))
	
	// 如果是 Office 文档或 PDF，使用 LibreOffice
	if p.libreoffice != nil && (ext == ".doc" || ext == ".docx" || ext == ".ppt" || ext == ".pptx" || 
		ext == ".xls" || ext == ".xlsx" || ext == ".pdf") {
		
		// 步骤1: 使用 LibreOffice 转换为 PNG（临时文件）
		tempDir := filepath.Dir(options.OutputPath)
		tempPNG := strings.TrimSuffix(options.OutputPath, filepath.Ext(options.OutputPath)) + "_temp.png"
		
		ctx := context.Background()
		pngPath, err := p.libreoffice.ConvertToPNG(ctx, filePath, tempDir)
		if err != nil {
			fmt.Printf("[DocumentProcessor] LibreOffice 转换失败: %v\n", err)
			return "", err
		}
		
		// LibreOffice 生成的文件名可能不同，重命名为临时文件
		if pngPath != tempPNG {
			if err := os.Rename(pngPath, tempPNG); err != nil {
				fmt.Printf("[DocumentProcessor] 重命名临时文件失败: %v\n", err)
				os.Remove(pngPath)
				return "", err
			}
		}
		
		fmt.Printf("[DocumentProcessor] LibreOffice 转换成功: %s\n", tempPNG)
		
		// 步骤2: 使用 ImageMagick 转换为 WebP
		if p.imagemagick != nil {
			err = p.imagemagick.Convert(tempPNG, options.OutputPath, "webp", options.Quality)
			if err != nil {
				fmt.Printf("[DocumentProcessor] 转换为 WebP 失败: %v\n", err)
				os.Remove(tempPNG)
				return "", err
			}
			
			fmt.Printf("[DocumentProcessor] 转换为 WebP 成功: %s\n", options.OutputPath)
			
			// 步骤3: 清理临时 PNG 文件
			err = os.Remove(tempPNG)
			if err != nil {
				fmt.Printf("[DocumentProcessor] 警告: 清理临时文件失败: %v\n", err)
			}
			
			return options.OutputPath, nil
		}
		
		// 如果没有 ImageMagick，直接返回 PNG
		if err := os.Rename(tempPNG, options.OutputPath); err != nil {
			os.Remove(tempPNG)
			return "", err
		}
		return options.OutputPath, nil
	}
	
	// 纯文本文件不支持预览
	if ext == ".txt" || ext == ".md" {
		return "", fmt.Errorf("文本文件不支持预览图生成")
	}
	
	return "", fmt.Errorf("不支持的文档格式: %s", ext)
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

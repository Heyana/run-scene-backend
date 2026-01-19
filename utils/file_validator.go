// Package utils 提供工具函数
package utils

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"

	"go_wails_project_manager/config"
	"go_wails_project_manager/logger"
)

// FileValidationConfig 文件验证配置
type FileValidationConfig struct {
	AllowedExtensions []string // 允许的扩展名
	AllowedMimeTypes  []string // 允许的MIME类型
	MaxFileSize       int64    // 最大文件大小（字节）
	CheckMagicNumber  bool     // 是否检查文件头魔数
}

// 预定义的验证配置
var (
	// ImageValidationConfig 图片验证配置
	ImageValidationConfig = FileValidationConfig{
		AllowedExtensions: []string{".jpg", ".jpeg", ".png", ".gif", ".webp", ".bmp", ".svg"},
		AllowedMimeTypes: []string{
			"image/jpeg",
			"image/png", 
			"image/gif",
			"image/webp",
			"image/bmp",
			"image/svg+xml",
		},
		MaxFileSize:      config.DefaultSecurityConfig.MaxUploadSize,
		CheckMagicNumber: true,
	}

	// DocumentValidationConfig 文档验证配置
	DocumentValidationConfig = FileValidationConfig{
		AllowedExtensions: []string{".pdf", ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx", ".txt"},
		AllowedMimeTypes: []string{
			"application/pdf",
			"application/msword",
			"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
			"application/vnd.ms-excel", 
			"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
			"application/vnd.ms-powerpoint",
			"application/vnd.openxmlformats-officedocument.presentationml.presentation",
			"text/plain",
		},
		MaxFileSize:      config.DefaultSecurityConfig.MaxUploadSize,
		CheckMagicNumber: false,
	}

	// VideoValidationConfig 视频验证配置
	VideoValidationConfig = FileValidationConfig{
		AllowedExtensions: []string{".mp4", ".avi", ".mov", ".wmv", ".flv", ".mkv"},
		AllowedMimeTypes: []string{
			"video/mp4",
			"video/x-msvideo", 
			"video/quicktime",
			"video/x-ms-wmv",
			"video/x-flv",
			"video/x-matroska",
		},
		MaxFileSize:      config.DefaultSecurityConfig.MaxUploadSize * 5, // 视频文件可以更大
		CheckMagicNumber: false,
	}
)

// FileMagicNumbers 文件魔数映射（文件头识别）
var FileMagicNumbers = map[string][]byte{
	"image/jpeg":      {0xFF, 0xD8, 0xFF},
	"image/png":       {0x89, 0x50, 0x4E, 0x47},
	"image/gif":       {0x47, 0x49, 0x46},
	"image/webp":      {0x52, 0x49, 0x46, 0x46}, // RIFF
	"image/bmp":       {0x42, 0x4D},
	"application/pdf": {0x25, 0x50, 0x44, 0x46},
}

// FileValidator 文件验证器
type FileValidator struct {
	config FileValidationConfig
}

// NewFileValidator 创建文件验证器
func NewFileValidator(cfg FileValidationConfig) *FileValidator {
	return &FileValidator{config: cfg}
}

// ValidateFile 验证上传的文件
func (fv *FileValidator) ValidateFile(file *multipart.FileHeader) error {
	// 1. 检查文件大小
	if file.Size > fv.config.MaxFileSize {
		return fmt.Errorf("文件大小超过限制：%d MB", fv.config.MaxFileSize/(1024*1024))
	}

	// 2. 检查文件扩展名
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !contains(fv.config.AllowedExtensions, ext) {
		return fmt.Errorf("不支持的文件类型：%s，允许的类型：%v", ext, fv.config.AllowedExtensions)
	}

	// 3. 检查MIME类型（从header获取）
	contentType := file.Header.Get("Content-Type")
	if !contains(fv.config.AllowedMimeTypes, contentType) {
		return fmt.Errorf("不支持的MIME类型：%s", contentType)
	}

	// 4. 检查文件头魔数（防止伪造扩展名）
	if fv.config.CheckMagicNumber {
		if err := fv.validateMagicNumber(file, contentType); err != nil {
			return err
		}
	}

	// 5. 检查文件名安全性
	if err := fv.validateFilename(file.Filename); err != nil {
		return err
	}

	return nil
}

// validateMagicNumber 验证文件头魔数
func (fv *FileValidator) validateMagicNumber(file *multipart.FileHeader, expectedMimeType string) error {
	expectedMagic, exists := FileMagicNumbers[expectedMimeType]
	if !exists {
		// 没有对应的魔数定义，跳过检查
		return nil
	}

	// 打开文件读取前几个字节
	f, err := file.Open()
	if err != nil {
		return fmt.Errorf("无法读取文件：%v", err)
	}
	defer f.Close()

	// 读取文件头（取最长的魔数长度）
	buffer := make([]byte, 12)
	n, err := io.ReadFull(f, buffer)
	if err != nil && err != io.ErrUnexpectedEOF {
		return fmt.Errorf("读取文件头失败：%v", err)
	}

	// 比较魔数
	if n < len(expectedMagic) {
		return fmt.Errorf("文件头不完整")
	}

	if !bytes.HasPrefix(buffer[:n], expectedMagic) {
		logger.Log.Warnf("⚠️ 文件魔数不匹配: 期望=%v, 实际=%v", expectedMagic, buffer[:len(expectedMagic)])
		return fmt.Errorf("文件类型验证失败：文件内容与扩展名不匹配")
	}

	return nil
}

// validateFilename 验证文件名安全性
func (fv *FileValidator) validateFilename(filename string) error {
	// 防止路径遍历攻击
	if strings.Contains(filename, "..") ||
		strings.Contains(filename, "/") ||
		strings.Contains(filename, "\\") {
		return fmt.Errorf("文件名包含非法字符")
	}

	// 防止隐藏文件
	if strings.HasPrefix(filename, ".") {
		return fmt.Errorf("不允许上传隐藏文件")
	}

	// 检查特殊字符
	dangerousChars := []string{"<", ">", ":", "\"", "|", "?", "*", "\x00"}
	for _, char := range dangerousChars {
		if strings.Contains(filename, char) {
			return fmt.Errorf("文件名包含非法字符：%s", char)
		}
	}

	return nil
}

// SanitizeFilename 清理文件名（移除危险字符）
func (fv *FileValidator) SanitizeFilename(filename string) string {
	// 移除路径分隔符
	filename = filepath.Base(filename)

	// 移除特殊字符
	replacer := strings.NewReplacer(
		"<", "",
		">", "",
		":", "",
		"\"", "",
		"|", "",
		"?", "",
		"*", "",
		"\x00", "",
		"..", "",
	)

	return replacer.Replace(filename)
}

// 快捷方法

// ValidateImageFile 验证图片文件
func ValidateImageFile(file *multipart.FileHeader) error {
	validator := NewFileValidator(ImageValidationConfig)
	return validator.ValidateFile(file)
}

// ValidateDocumentFile 验证文档文件
func ValidateDocumentFile(file *multipart.FileHeader) error {
	validator := NewFileValidator(DocumentValidationConfig)
	return validator.ValidateFile(file)
}

// ValidateVideoFile 验证视频文件
func ValidateVideoFile(file *multipart.FileHeader) error {
	validator := NewFileValidator(VideoValidationConfig)
	return validator.ValidateFile(file)
}

// contains 检查切片是否包含元素
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// GetFileTypeByExtension 根据扩展名获取文件类型
func GetFileTypeByExtension(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))

	if contains(ImageValidationConfig.AllowedExtensions, ext) {
		return "image"
	}
	if contains(DocumentValidationConfig.AllowedExtensions, ext) {
		return "document"
	}
	if contains(VideoValidationConfig.AllowedExtensions, ext) {
		return "video"
	}

	return "unknown"
}

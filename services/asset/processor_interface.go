package asset

import (
	"go_wails_project_manager/models"
	"mime/multipart"
)

// AssetProcessor 资产处理器接口
type AssetProcessor interface {
	// GenerateThumbnail 生成缩略图
	GenerateThumbnail(filePath string, outputPath string) error
	
	// ExtractMetadata 提取元数据
	ExtractMetadata(filePath string) (*models.AssetMetadata, error)
	
	// Validate 验证文件
	Validate(file *multipart.FileHeader) error
	
	// SupportedFormats 获取支持的格式
	SupportedFormats() []string
}

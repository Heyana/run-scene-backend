// Package processors 文件处理器接口定义
package processors

// FileProcessor 文件处理器接口
type FileProcessor interface {
	// Support 检查是否支持该文件格式
	Support(format string) bool

	// ExtractMetadata 提取文件元数据
	ExtractMetadata(filePath string) (*FileMetadata, error)

	// GeneratePreview 生成预览图
	GeneratePreview(filePath string, options PreviewOptions) (*PreviewResult, error)

	// GenerateThumbnail 生成缩略图
	GenerateThumbnail(filePath string, options ThumbnailOptions) (string, error)

	// Convert 格式转换（可选）
	Convert(filePath string, options ConvertOptions) (string, error)

	// Validate 验证文件完整性（可选）
	Validate(filePath string) error

	// Name 获取处理器名称
	Name() string
}

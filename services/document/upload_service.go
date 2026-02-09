package document

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"go_wails_project_manager/config"
	"go_wails_project_manager/logger"
	"go_wails_project_manager/models"
	"go_wails_project_manager/services/storage"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"gorm.io/gorm"
)

// UploadService 上传服务
type UploadService struct {
	db             *gorm.DB
	config         *config.DocumentConfig
	storageService *storage.FileStorageService
}

// NewUploadService 创建上传服务
func NewUploadService(db *gorm.DB, cfg *config.DocumentConfig) *UploadService {
	// 创建存储服务配置
	storageConfig := &storage.StorageConfig{
		LocalStorageEnabled: cfg.LocalStorageEnabled,
		StorageDir:          cfg.StorageDir,
		NASEnabled:          cfg.NASEnabled,
		NASPath:             cfg.NASPath,
	}

	return &UploadService{
		db:             db,
		config:         cfg,
		storageService: storage.NewFileStorageService(storageConfig, logger.Log),
	}
}

// UploadMetadata 上传元数据
type UploadMetadata struct {
	Name        string
	Description string
	Category    string
	Tags        []string
	Department  string
	Project     string
	IsPublic    bool
	Version     string
	UploadedBy  string
	UploadIP    string
}

// Upload 上传文档
func (s *UploadService) Upload(file *multipart.FileHeader, metadata UploadMetadata) (*models.Document, error) {
	// 1. 检测文件格式
	format := s.detectFileFormat(file)
	if format == "" {
		return nil, fmt.Errorf("无法识别的文件格式")
	}

	// 2. 根据格式判断文档类型
	docType := models.GetDocumentType(format)

	// 3. 验证文件格式
	if !s.isFormatAllowed(docType, format) {
		return nil, fmt.Errorf("不支持的文件格式: %s", format)
	}

	// 4. 验证文件大小
	if !s.isFileSizeAllowed(docType, file.Size) {
		maxSize := s.config.MaxFileSize[docType]
		return nil, fmt.Errorf("文件大小超过限制: %d MB", maxSize/1024/1024)
	}

	// 5. 计算文件哈希
	fileHash, err := s.calculateHash(file)
	if err != nil {
		return nil, fmt.Errorf("计算文件哈希失败: %w", err)
	}

	// 6. 检查是否重复
	existingDoc, isDuplicate, err := s.checkDuplicate(fileHash)
	if err != nil {
		return nil, err
	}
	if isDuplicate {
		logger.Log.Infof("文件已存在，返回现有记录: %s (ID: %d)", existingDoc.Name, existingDoc.ID)
		return nil, fmt.Errorf("文件已存在: %s (上传于 %s)", existingDoc.Name, existingDoc.CreatedAt.Format("2006-01-02 15:04:05"))
	}

	// 7. 清洗文件名（如果用户没有提供名称，使用清洗后的文件名）
	if metadata.Name == "" {
		metadata.Name = s.sanitizeFileName(file.Filename)
	}

	// 8. 自动生成版本号
	if metadata.Version == "" && s.config.AutoVersion {
		metadata.Version = "v1.0"
	}

	// 9. 创建文档记录
	document := &models.Document{
		Name:        metadata.Name,
		Description: metadata.Description,
		Category:    metadata.Category,
		Tags:        strings.Join(metadata.Tags, ","),
		Type:        docType,
		FileSize:    file.Size,
		FileHash:    fileHash,
		Format:      format,
		Version:     metadata.Version,
		Department:  metadata.Department,
		Project:     metadata.Project,
		IsPublic:    metadata.IsPublic,
		IsLatest:    true,
		UploadedBy:  metadata.UploadedBy,
		UploadIP:    metadata.UploadIP,
	}

	if err := s.db.Create(document).Error; err != nil {
		return nil, fmt.Errorf("创建文档记录失败: %w", err)
	}

	// 10. 保存文件（使用清洗后的文件名）
	filePath, err := s.saveFile(file, document.ID)
	if err != nil {
		s.db.Delete(document) // 回滚
		return nil, err
	}
	document.FilePath = filePath

	// 11. 更新文档路径
	if err := s.db.Save(document).Error; err != nil {
		s.storageService.DeleteFile(fmt.Sprintf("%d", document.ID))
		s.db.Delete(document)
		return nil, err
	}

	// 12. 更新统计信息
	s.updateMetrics(document)

	// 13. 记录访问日志
	s.logAccess(document.ID, "upload", metadata.UploadedBy, metadata.UploadIP)

	return document, nil
}

// saveFile 保存文件到存储
func (s *UploadService) saveFile(file *multipart.FileHeader, documentID uint) (string, error) {
	// 读取文件数据
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("打开上传文件失败: %w", err)
	}
	defer src.Close()

	data, err := io.ReadAll(src)
	if err != nil {
		return "", fmt.Errorf("读取文件数据失败: %w", err)
	}

	// 确定文件名
	ext := filepath.Ext(file.Filename)
	fileName := "file" + ext

	// 使用通用存储服务保存
	subPath := fmt.Sprintf("%d", documentID)

	return s.storageService.SaveFile(subPath, fileName, data)
}

// detectFileFormat 检测文件格式
func (s *UploadService) detectFileFormat(file *multipart.FileHeader) string {
	// 打开文件
	src, err := file.Open()
	if err != nil {
		return ""
	}
	defer src.Close()

	// 读取文件头
	buffer := make([]byte, 512)
	n, err := src.Read(buffer)
	if err != nil && err != io.EOF {
		return ""
	}

	// 使用 http.DetectContentType 检测 MIME 类型
	mimeType := http.DetectContentType(buffer[:n])

	// 将 MIME 类型映射到文件扩展名
	format := s.mimeTypeToFormat(mimeType)

	// 如果无法从 MIME 检测，尝试从文件名获取
	if format == "" {
		format = strings.ToLower(strings.TrimPrefix(filepath.Ext(file.Filename), "."))
	}

	return format
}

// mimeTypeToFormat 将 MIME 类型转换为文件格式
func (s *UploadService) mimeTypeToFormat(mimeType string) string {
	mimeMap := map[string]string{
		"application/pdf":                                                          "pdf",
		"application/msword":                                                       "doc",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document": "docx",
		"application/vnd.ms-excel":                                                 "xls",
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":       "xlsx",
		"application/vnd.ms-powerpoint":                                            "ppt",
		"application/vnd.openxmlformats-officedocument.presentationml.presentation": "pptx",
		"text/plain":                "txt",
		"text/plain; charset=utf-8": "txt",
		"text/markdown":             "md",
		"video/mp4":                 "mp4",
		"video/webm":                "webm",
		"video/x-msvideo":           "avi",
		"video/quicktime":           "mov",
		"application/zip":           "zip",
		"application/x-rar-compressed": "rar",
		"application/x-7z-compressed":  "7z",
		"image/jpeg":                   "jpg",
		"image/png":                    "png",
		"image/gif":                    "gif",
		"audio/mpeg":                   "mp3",
		"audio/wav":                    "wav",
	}

	if format, ok := mimeMap[mimeType]; ok {
		return format
	}

	// 尝试从 MIME 类型中提取格式
	parts := strings.Split(mimeType, "/")
	if len(parts) == 2 {
		// 移除可能的 charset 等参数
		format := strings.Split(parts[1], ";")[0]
		return strings.TrimSpace(format)
	}

	return ""
}

// isFormatAllowed 检查格式是否允许
func (s *UploadService) isFormatAllowed(docType, format string) bool {
	allowedFormats, ok := s.config.AllowedFormats[docType]
	if !ok {
		return false
	}

	for _, allowed := range allowedFormats {
		if strings.EqualFold(allowed, format) {
			return true
		}
	}

	return false
}

// isFileSizeAllowed 检查文件大小是否允许
func (s *UploadService) isFileSizeAllowed(docType string, size int64) bool {
	maxSize, ok := s.config.MaxFileSize[docType]
	if !ok {
		return false
	}

	return size <= maxSize
}

// calculateHash 计算文件哈希
func (s *UploadService) calculateHash(file *multipart.FileHeader) (string, error) {
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, src); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// checkDuplicate 检查文件是否已存在
func (s *UploadService) checkDuplicate(fileHash string) (*models.Document, bool, error) {
	var document models.Document
	err := s.db.Where("file_hash = ?", fileHash).First(&document).Error

	if err == gorm.ErrRecordNotFound {
		return nil, false, nil
	}

	if err != nil {
		return nil, false, err
	}

	return &document, true, nil
}

// updateMetrics 更新统计信息
func (s *UploadService) updateMetrics(document *models.Document) {
	today := time.Now().Format("2006-01-02")

	var metrics models.DocumentMetrics
	err := s.db.Where("date = ? AND type = ?", today, document.Type).
		FirstOrCreate(&metrics, models.DocumentMetrics{
			Date: today,
			Type: document.Type,
		}).Error

	if err != nil {
		return
	}

	metrics.UploadCount++
	metrics.TotalDocs++
	metrics.TotalSize += document.FileSize

	s.db.Save(&metrics)
}

// logAccess 记录访问日志
func (s *UploadService) logAccess(documentID uint, action, userName, userIP string) {
	if !s.config.LogAccess {
		return
	}

	log := &models.DocumentAccessLog{
		DocumentID: documentID,
		Action:     action,
		UserName:   userName,
		UserIP:     userIP,
	}

	s.db.Create(log)
}

// sanitizeFileName 清洗文件名，移除特殊字符
func (s *UploadService) sanitizeFileName(filename string) string {
	// 获取文件名（不含扩展名）
	ext := filepath.Ext(filename)
	nameWithoutExt := strings.TrimSuffix(filename, ext)
	
	// 定义需要移除或替换的特殊字符
	replacements := map[string]string{
		"#": "",
		"@": "",
		"$": "",
		"%": "",
		"&": "",
		"*": "",
		"+": "",
		"=": "",
		"{": "",
		"}": "",
		"[": "",
		"]": "",
		"|": "",
		"\\": "",
		"/": "",
		":": "",
		";": "",
		"\"": "",
		"'": "",
		"<": "",
		">": "",
		"?": "",
		"!": "",
		"~": "",
		"`": "",
	}
	
	// 替换特殊字符
	cleanName := nameWithoutExt
	for old, new := range replacements {
		cleanName = strings.ReplaceAll(cleanName, old, new)
	}
	
	// 移除多余的空格和下划线
	cleanName = strings.TrimSpace(cleanName)
	cleanName = strings.ReplaceAll(cleanName, "  ", " ")
	cleanName = strings.ReplaceAll(cleanName, "__", "_")
	
	// 限制长度（最多100个字符）
	if len(cleanName) > 100 {
		cleanName = cleanName[:100]
	}
	
	// 如果清洗后为空，使用默认名称
	if cleanName == "" {
		cleanName = "未命名文档"
	}
	
	return cleanName
}


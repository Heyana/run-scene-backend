package asset

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"go_wails_project_manager/config"
	"go_wails_project_manager/logger"
	"go_wails_project_manager/models"
	"go_wails_project_manager/services/asset/processors"
	"go_wails_project_manager/services/storage"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
	
	"gorm.io/gorm"
)

// UploadService 上传服务
type UploadService struct {
	db             *gorm.DB
	config         *config.AssetConfig
	processors     map[string]AssetProcessor
	storageService *storage.FileStorageService
}

// NewUploadService 创建上传服务
func NewUploadService(db *gorm.DB, cfg *config.AssetConfig) *UploadService {
	// 创建存储服务配置
	storageConfig := &storage.StorageConfig{
		LocalStorageEnabled: cfg.LocalStorageEnabled,
		StorageDir:          cfg.StorageDir,
		NASEnabled:          cfg.NASEnabled,
		NASPath:             cfg.NASPath,
	}
	
	service := &UploadService{
		db:             db,
		config:         cfg,
		processors:     make(map[string]AssetProcessor),
		storageService: storage.NewFileStorageService(storageConfig, logger.Log),
	}
	
	// 注册处理器
	service.processors["image"] = processors.NewImageProcessor(cfg)
	service.processors["video"] = processors.NewVideoProcessor(cfg)
	
	return service
}

// UploadMetadata 上传元数据
type UploadMetadata struct {
	Name        string
	Description string
	Category    string
	Tags        []string
	UploadedBy  string
	UploadIP    string
}

// Upload 上传资产
func (s *UploadService) Upload(file *multipart.FileHeader, metadata UploadMetadata) (*models.Asset, error) {
	// 1. 检测真实文件类型
	realFormat, err := s.detectFileFormat(file)
	if err != nil {
		return nil, fmt.Errorf("检测文件类型失败: %w", err)
	}
	
	// 2. 根据格式判断资产类型
	assetType := s.getAssetTypeFromFormat(realFormat)
	if assetType == "" {
		return nil, fmt.Errorf("不支持的文件格式: %s", realFormat)
	}
	
	// 3. 验证文件类型
	processor, err := s.getProcessor(assetType)
	if err != nil {
		return nil, err
	}
	
	// 4. 验证文件
	if err := processor.Validate(file); err != nil {
		return nil, err
	}
	
	// 5. 计算文件哈希
	fileHash, err := s.calculateHash(file)
	if err != nil {
		return nil, fmt.Errorf("计算文件哈希失败: %w", err)
	}
	
	// 6. 检查是否重复
	existingAsset, isDuplicate, err := s.checkDuplicate(fileHash)
	if err != nil {
		return nil, err
	}
	if isDuplicate {
		return existingAsset, nil
	}
	
	// 7. 创建资产记录（使用检测到的真实格式和类型）
	asset := &models.Asset{
		Name:        metadata.Name,
		Description: metadata.Description,
		Category:    metadata.Category,
		Tags:        strings.Join(metadata.Tags, ","),
		Type:        assetType,    // 使用检测到的类型
		FileSize:    file.Size,
		FileHash:    fileHash,
		Format:      realFormat,   // 使用检测到的格式
		UploadedBy:  metadata.UploadedBy,
		UploadIP:    metadata.UploadIP,
	}
	
	if err := s.db.Create(asset).Error; err != nil {
		return nil, fmt.Errorf("创建资产记录失败: %w", err)
	}
	
	// 8. 保存文件
	filePath, err := s.saveFile(file, asset.ID)
	if err != nil {
		s.db.Delete(asset) // 回滚
		return nil, err
	}
	asset.FilePath = filePath
	
	// 9. 生成缩略图
	// 对于特殊格式（APNG、GIF 等动画格式），使用原格式作为缩略图扩展名
	// 对于 JPEG/PNG 等格式，如果处理失败会自动回退到复制原文件
	thumbnailExt := ".webp"
	if asset.Format == "apng" || asset.Format == "gif" {
		// 动画格式使用原格式
		thumbnailExt = "." + asset.Format
	} else if asset.Format == "png" || asset.Format == "jpg" || asset.Format == "jpeg" {
		// 对于 PNG/JPEG，先尝试生成 WebP，如果失败会自动复制原文件
		// 但为了保险起见，如果是大文件（可能是全景图），直接使用原格式
		if file.Size > 5*1024*1024 { // 大于 5MB
			thumbnailExt = "." + asset.Format
		}
	}
	
	// 获取实际物理路径用于生成缩略图
	actualFilePath, _ := s.storageService.GetFilePath(fmt.Sprintf("%d", asset.ID), "file"+filepath.Ext(file.Filename))
	actualThumbnailPath := s.getActualThumbnailPath(asset.ID, thumbnailExt)
	
	// 确保目录存在
	if err := os.MkdirAll(filepath.Dir(actualThumbnailPath), 0755); err != nil {
		os.Remove(actualFilePath) // 清理文件
		s.db.Delete(asset)         // 回滚
		return nil, fmt.Errorf("创建缩略图目录失败: %w", err)
	}
	
	if err := processor.GenerateThumbnail(actualFilePath, actualThumbnailPath); err != nil {
		s.storageService.DeleteFile(fmt.Sprintf("%d", asset.ID)) // 清理文件
		s.db.Delete(asset)                                       // 回滚
		return nil, fmt.Errorf("生成缩略图失败: %w", err)
	}
	
	// 保存相对路径到数据库
	asset.ThumbnailPath = s.getThumbnailPathWithExt(asset.ID, thumbnailExt)
	
	// 10. 提取元数据
	assetMetadata, err := processor.ExtractMetadata(actualFilePath)
	if err == nil {
		assetMetadata.AssetID = asset.ID
		s.db.Create(assetMetadata)
	}
	
	// 11. 更新资产路径
	if err := s.db.Save(asset).Error; err != nil {
		s.storageService.DeleteFile(fmt.Sprintf("%d", asset.ID))
		s.db.Delete(asset)
		return nil, err
	}
	
	// 12. 更新统计信息
	s.updateMetrics(asset)
	
	return asset, nil
}

// saveFile 保存文件到磁盘
func (s *UploadService) saveFile(file *multipart.FileHeader, assetID uint) (string, error) {
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
	subPath := fmt.Sprintf("%d", assetID)
	
	return s.storageService.SaveFile(subPath, fileName, data)
}

// getThumbnailPath 获取缩略图路径
func (s *UploadService) getThumbnailPath(assetID uint) string {
	return s.getThumbnailPathWithExt(assetID, ".webp")
}

// getThumbnailPathWithExt 获取指定扩展名的缩略图路径（相对路径）
func (s *UploadService) getThumbnailPathWithExt(assetID uint, ext string) string {
	return filepath.Join(s.config.StorageDir, fmt.Sprintf("%d", assetID), "thumbnail"+ext)
}

// getActualThumbnailPath 获取缩略图的实际物理路径（用于生成缩略图）
func (s *UploadService) getActualThumbnailPath(assetID uint, ext string) string {
	subPath := fmt.Sprintf("%d", assetID)
	fileName := "thumbnail" + ext
	
	// 优先使用本地路径
	if s.config.LocalStorageEnabled {
		return filepath.Join(s.config.StorageDir, subPath, fileName)
	}
	
	// 使用 NAS 路径
	if s.config.NASEnabled {
		return filepath.Join(s.config.NASPath, subPath, fileName)
	}
	
	return filepath.Join(s.config.StorageDir, subPath, fileName)
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
func (s *UploadService) checkDuplicate(fileHash string) (*models.Asset, bool, error) {
	var asset models.Asset
	err := s.db.Where("file_hash = ?", fileHash).First(&asset).Error
	
	if err == gorm.ErrRecordNotFound {
		return nil, false, nil
	}
	
	if err != nil {
		return nil, false, err
	}
	
	return &asset, true, nil
}

// getProcessor 获取对应的处理器
func (s *UploadService) getProcessor(assetType string) (AssetProcessor, error) {
	processor, ok := s.processors[assetType]
	if !ok {
		return nil, errors.New("不支持的资产类型")
	}
	return processor, nil
}

// updateMetrics 更新统计信息
func (s *UploadService) updateMetrics(asset *models.Asset) {
	today := time.Now().Format("2006-01-02")
	
	var metrics models.AssetMetrics
	err := s.db.Where("date = ? AND type = ?", today, asset.Type).
		FirstOrCreate(&metrics, models.AssetMetrics{
			Date: today,
			Type: asset.Type,
		}).Error
	
	if err != nil {
		return
	}
	
	metrics.UploadCount++
	metrics.UploadSize += asset.FileSize
	metrics.TotalAssets++
	metrics.TotalSize += asset.FileSize
	
	s.db.Save(&metrics)
}

// detectFileFormat 检测文件真实格式
func (s *UploadService) detectFileFormat(file *multipart.FileHeader) (string, error) {
	// 打开文件
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()
	
	// 读取文件头（前512字节足够检测大多数格式）
	buffer := make([]byte, 512)
	n, err := src.Read(buffer)
	if err != nil && err != io.EOF {
		return "", err
	}
	
	// 使用 http.DetectContentType 检测 MIME 类型
	mimeType := http.DetectContentType(buffer[:n])
	
	// 将 MIME 类型映射到文件扩展名
	format := s.mimeTypeToFormat(mimeType)
	
	// 如果无法从 MIME 检测，尝试从文件名获取
	if format == "" {
		format = strings.ToLower(strings.TrimPrefix(filepath.Ext(file.Filename), "."))
	}
	
	// 特殊处理：检测 APNG
	if format == "png" {
		// 重置文件指针
		src.Seek(0, 0)
		if s.isAPNG(src) {
			format = "apng"
		}
	}
	
	return format, nil
}

// mimeTypeToFormat 将 MIME 类型转换为文件格式
func (s *UploadService) mimeTypeToFormat(mimeType string) string {
	mimeMap := map[string]string{
		"image/jpeg":               "jpg",
		"image/png":                "png",
		"image/webp":               "webp",
		"image/gif":                "gif",
		"video/mp4":                "mp4",
		"video/webm":               "webm",
		"application/octet-stream": "", // 未知类型
	}
	
	if format, ok := mimeMap[mimeType]; ok {
		return format
	}
	
	// 尝试从 MIME 类型中提取格式
	parts := strings.Split(mimeType, "/")
	if len(parts) == 2 {
		return parts[1]
	}
	
	return ""
}

// isAPNG 检测是否为 APNG 格式
func (s *UploadService) isAPNG(reader io.Reader) bool {
	// APNG 文件在 PNG 签名后会有 acTL chunk
	// PNG 签名: 89 50 4E 47 0D 0A 1A 0A (8 bytes)
	
	buffer := make([]byte, 1024)
	n, err := reader.Read(buffer)
	if err != nil || n < 8 {
		return false
	}
	
	// 检查 PNG 签名
	pngSignature := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	if !bytes.Equal(buffer[:8], pngSignature) {
		return false
	}
	
	// 在前 1024 字节中查找 acTL chunk
	acTLChunk := []byte("acTL")
	return bytes.Contains(buffer, acTLChunk)
}

// getAssetTypeFromFormat 根据文件格式判断资产类型
func (s *UploadService) getAssetTypeFromFormat(format string) string {
	imageFormats := map[string]bool{
		"jpg":  true,
		"jpeg": true,
		"png":  true,
		"webp": true,
		"apng": true,
		"gif":  true,
	}
	
	videoFormats := map[string]bool{
		"mp4":  true,
		"webm": true,
	}
	
	if imageFormats[format] {
		return "image"
	}
	
	if videoFormats[format] {
		return "video"
	}
	
	return ""
}

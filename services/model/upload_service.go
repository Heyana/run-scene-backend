package model

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"

	"gorm.io/gorm"
	"go_wails_project_manager/config"
	"go_wails_project_manager/logger"
	"go_wails_project_manager/models"
	"go_wails_project_manager/services/storage"
)

var (
	ErrInvalidType      = errors.New("不支持的文件类型")
	ErrFileTooLarge     = errors.New("文件大小超过限制")
	ErrDuplicateFile    = errors.New("文件已存在")
	ErrMissingThumbnail = errors.New("缺少预览图")
	ErrInvalidMetadata  = errors.New("元数据格式错误")
)

type UploadService struct {
	db             *gorm.DB
	config         *config.ModelConfig
	storageService *storage.FileStorageService
}

type UploadMetadata struct {
	Name        string
	Description string
	Category    string
	Tags        []string
	Type        string
	UploadedBy  string
	UploadIP    string
}

func NewUploadService(db *gorm.DB, cfg *config.ModelConfig) *UploadService {
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

// UploadSingle 单文件上传（模型 + 预览图）
func (s *UploadService) UploadSingle(
	modelFile, thumbnailFile *multipart.FileHeader,
	metadata UploadMetadata,
) (*models.Model, error) {
	// 1. 验证模型文件
	if err := s.validateFile(modelFile, s.config.AllowedTypes, s.config.MaxFileSize); err != nil {
		return nil, err
	}

	// 2. 验证预览图
	if err := s.validateFile(thumbnailFile, []string{"webp", "jpg", "jpeg", "png"}, s.config.MaxThumbnailSize); err != nil {
		return nil, err
	}

	// 3. 计算哈希
	fileHash, err := s.calculateHash(modelFile)
	if err != nil {
		return nil, fmt.Errorf("计算文件哈希失败: %w", err)
	}

	// 4. 检查重复
	if existing, isDup, err := s.checkDuplicate(fileHash); err != nil {
		return nil, err
	} else if isDup {
		return existing, ErrDuplicateFile
	}

	// 5. 创建模型记录（获取ID）
	model := &models.Model{
		Name:        metadata.Name,
		Description: metadata.Description,
		Category:    metadata.Category,
		Tags:        strings.Join(metadata.Tags, ","),
		Type:        metadata.Type,
		FileSize:    modelFile.Size,
		FileHash:    fileHash,
		UploadedBy:  metadata.UploadedBy,
		UploadIP:    metadata.UploadIP,
	}

	if err := s.db.Create(model).Error; err != nil {
		return nil, fmt.Errorf("创建数据库记录失败: %w", err)
	}

	// 6. 保存模型文件
	filePath, err := s.saveModelFile(modelFile, model.ID, metadata.Type)
	if err != nil {
		s.db.Delete(model) // 回滚
		return nil, fmt.Errorf("保存模型文件失败: %w", err)
	}

	// 7. 保存预览图
	thumbnailPath, err := s.saveThumbnail(thumbnailFile, model.ID)
	if err != nil {
		s.storageService.DeleteFile(fmt.Sprintf("%d", model.ID)) // 清理文件
		s.db.Delete(model)                                       // 回滚
		return nil, fmt.Errorf("保存预览图失败: %w", err)
	}

	// 8. 更新路径
	model.FilePath = filePath
	model.ThumbnailPath = thumbnailPath
	if err := s.db.Save(model).Error; err != nil {
		s.storageService.DeleteFile(fmt.Sprintf("%d", model.ID)) // 清理文件
		return nil, fmt.Errorf("更新模型记录失败: %w", err)
	}

	return model, nil
}

// validateFile 验证文件
func (s *UploadService) validateFile(file *multipart.FileHeader, allowedTypes []string, maxSize int64) error {
	// 检查文件大小
	if file.Size > maxSize {
		return ErrFileTooLarge
	}

	// 检查文件类型
	ext := strings.ToLower(filepath.Ext(file.Filename))
	ext = strings.TrimPrefix(ext, ".")

	allowed := false
	for _, t := range allowedTypes {
		if ext == t {
			allowed = true
			break
		}
	}

	if !allowed {
		return ErrInvalidType
	}

	return nil
}

// calculateHash 计算文件哈希
func (s *UploadService) calculateHash(file *multipart.FileHeader) (string, error) {
	f, err := file.Open()
	if err != nil {
		return "", err
	}
	defer f.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, f); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// checkDuplicate 检查文件是否已存在
func (s *UploadService) checkDuplicate(fileHash string) (*models.Model, bool, error) {
	var model models.Model
	err := s.db.Where("file_hash = ?", fileHash).First(&model).Error
	if err == gorm.ErrRecordNotFound {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	return &model, true, nil
}

// saveModelFile 保存模型文件
func (s *UploadService) saveModelFile(file *multipart.FileHeader, modelID uint, fileType string) (string, error) {
	// 读取文件数据
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	data, err := io.ReadAll(src)
	if err != nil {
		return "", err
	}

	// 使用通用存储服务保存
	subPath := fmt.Sprintf("%d", modelID)
	fileName := fmt.Sprintf("model.%s", fileType)
	
	return s.storageService.SaveFile(subPath, fileName, data)
}

// saveThumbnail 保存预览图
func (s *UploadService) saveThumbnail(file *multipart.FileHeader, modelID uint) (string, error) {
	// 读取文件数据
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	data, err := io.ReadAll(src)
	if err != nil {
		return "", err
	}

	// 获取原始扩展名
	ext := strings.ToLower(filepath.Ext(file.Filename))
	fileName := fmt.Sprintf("thumbnail%s", ext)
	
	// 使用通用存储服务保存
	subPath := fmt.Sprintf("%d", modelID)
	
	return s.storageService.SaveFile(subPath, fileName, data)
}

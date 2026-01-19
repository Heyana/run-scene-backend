package services

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"go_wails_project_manager/database"
	"go_wails_project_manager/logger"
	"go_wails_project_manager/models"
	"go_wails_project_manager/utils"
	"io"
	"mime/multipart"
	"strings"
)

// ResourceService 资源服务
type ResourceService struct {
	cdnService   CDNService
	imageProcess *utils.ImageProcessor
}

// NewResourceService 创建资源服务实例
func NewResourceService() *ResourceService {
	return &ResourceService{
		cdnService:   NewCDNService(),
		imageProcess: utils.GetDefaultProcessor(),
	}
}

// GetFullURL 获取资源的完整URL
func (s *ResourceService) GetFullURL(path string) string {
	// 如果已经是完整的URL，直接返回
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return path
	}
	// 否则使用CDN服务拼接完整URL
	return s.cdnService.GetURL(path)
}

// SaveResource 保存资源
func (s *ResourceService) SaveResource(ctx context.Context, file multipart.File, header *multipart.FileHeader, libraryID *uint) (*models.ResourceFile, error) {
	// 读取文件内容
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	// 计算文件哈希
	hash := calculateFileHash(data)

	// 检查是否已存在相同哈希的资源
	db, err := database.GetDB()
	if err != nil {
		return nil, err
	}

	// 查找相同哈希的资源（支持去重）
	var existingResource models.ResourceFile
	query := db.Where("hash_code = ?", hash)
	if libraryID != nil {
		// 如果指定了媒体库，优先查找同库下的资源
		query = query.Where("library_id = ?", *libraryID)
	}
	if err := query.First(&existingResource).Error; err == nil {
		// 资源已存在，直接返回
		logger.Log.Infof("资源已存在，复用资源: %s (ID=%d)", existingResource.FileName, existingResource.ID)
		return &existingResource, nil
	}

	// 处理资源
	return s.processAndSaveResource(ctx, data, header, libraryID, hash)
}

// processAndSaveResource 处理和保存资源
func (s *ResourceService) processAndSaveResource(ctx context.Context, data []byte, header *multipart.FileHeader, libraryID *uint, hash string) (*models.ResourceFile, error) {
	mimeType := header.Header.Get("Content-Type")
	fileName := header.Filename
	fileSize := len(data)

	// 确定存储类型和内容
	var storageType string
	var filePath string
	var content []byte

	// 初始化处理后的数据为原始数据
	processedData := data

	// 根据文件类型处理
	// 1. 处理图片
	if strings.HasPrefix(mimeType, "image/") {
		// 图片处理
		processedImg, newMimeType, err := s.imageProcess.ProcessImage(data, mimeType)
		if err != nil {
			logger.Log.Warnf("图片处理失败: %v，使用原始图片", err)
		} else {
			processedData = processedImg
			mimeType = newMimeType
		}

		// 处理后重新计算哈希值，再次检查是否存在相同的资源
		processedHash := calculateFileHash(processedData)
		if processedHash != hash {
			// 处理后的图片哈希值发生变化，再次查找是否存在相同的资源
			db, err := database.GetDB()
			if err == nil {
				var existingProcessedResource models.ResourceFile
				query := db.Where("hash_code = ?", processedHash)
				if libraryID != nil {
					query = query.Where("library_id = ?", *libraryID)
				}
				if err := query.First(&existingProcessedResource).Error; err == nil {
					// 处理后的资源已存在，直接返回
					logger.Log.Infof("处理后的资源已存在，复用资源: %s (ID=%d)", existingProcessedResource.FileName, existingProcessedResource.ID)
					return &existingProcessedResource, nil
				}
			}
			// 更新哈希值为处理后的哈希值
			hash = processedHash
		}
	}

	// 2. 处理富文本文件
	if strings.Contains(fileName, "richText") || strings.HasSuffix(fileName, ".html") ||
		mimeType == "application/json" || mimeType == "text/html" {
		// 富文本内容，直接使用原始数据

		// 确保使用正确的MIME类型
		originalMimeType := mimeType
		if mimeType != "text/html" && mimeType != "application/json" {
			mimeType = "text/html"
		}

		// 记录详细日志
		logger.Log.Infof("处理富文本内容: 文件名=%s, 原始MIME类型=%s, 最终MIME类型=%s, 大小=%d 字节",
			fileName, originalMimeType, mimeType, len(processedData))

		// 检测是否是JSON格式
		if strings.HasPrefix(string(data), "{") && strings.HasSuffix(string(data), "}") {
			logger.Log.Infof("检测到JSON格式的富文本内容")
		}
	}

	// 上传到CDN
	path := GenerateFilePath(fileName, mimeType)
	url, err := s.cdnService.Upload(ctx, processedData, path, mimeType)
	if err != nil {
		return nil, err
	}

	storageType = "cdn"
	filePath = url
	content = nil                 // CDN存储不保留内容
	fileSize = len(processedData) // 更新为处理后的文件大小

	// 准备元数据
	metadata := map[string]interface{}{
		"original_name": fileName,
		"original_size": header.Size,
		"mime_type":     mimeType,
	}
	metadataBytes, _ := json.Marshal(metadata)

	// 创建资源记录
	resource := &models.ResourceFile{
		LibraryID:   libraryID,
		Type:        getResourceTypeFromMimeType(mimeType),
		StorageType: storageType,
		FileName:    fileName,
		FileSize:    int64(fileSize),
		MimeType:    mimeType,
		FilePath:    filePath,
		Content:     content,
		Metadata:    string(metadataBytes),
		HashCode:    hash,
	}

	// 保存到数据库
	db, err := database.GetDB()
	if err != nil {
		return nil, err
	}

	if err := db.Create(resource).Error; err != nil {
		// 如果保存失败，尝试删除上传的文件
		if storageType == "cdn" {
			_ = s.cdnService.Delete(ctx, filePath)
		}
		return nil, err
	}

	logger.Log.Infof("资源保存成功: %s, ID: %d", resource.FileName, resource.ID)
	return resource, nil
}

// DeleteResource 删除资源
func (s *ResourceService) DeleteResource(ctx context.Context, resourceID uint) error {
	db, err := database.GetDB()
	if err != nil {
		return err
	}

	var resource models.ResourceFile
	if err := db.First(&resource, resourceID).Error; err != nil {
		return err
	}

	// 如果是CDN存储，删除文件
	if resource.StorageType == "cdn" && resource.FilePath != "" {
		if err := s.cdnService.Delete(ctx, resource.FilePath); err != nil {
			logger.Log.Warnf("删除CDN文件失败: %v", err)
		}
	}

	// 删除数据库记录
	return db.Delete(&resource).Error
}

// GetResource 获取资源
func (s *ResourceService) GetResource(resourceID uint) (*models.ResourceFile, error) {
	db, err := database.GetDB()
	if err != nil {
		return nil, err
	}

	var resource models.ResourceFile
	if err := db.First(&resource, resourceID).Error; err != nil {
		return nil, err
	}

	return &resource, nil
}

// calculateFileHash 计算文件哈希值
func calculateFileHash(data []byte) string {
	hasher := md5.New()
	hasher.Write(data)
	return hex.EncodeToString(hasher.Sum(nil))
}

// getResourceTypeFromMimeType 从MIME类型获取资源类型
func getResourceTypeFromMimeType(mimeType string) string {
	if strings.HasPrefix(mimeType, "image/") {
		return "image"
	} else if strings.HasPrefix(mimeType, "video/") {
		return "video"
	} else if strings.HasPrefix(mimeType, "audio/") {
		return "audio"
	} else if strings.HasPrefix(mimeType, "text/html") ||
		strings.HasPrefix(mimeType, "application/xhtml") ||
		strings.HasPrefix(mimeType, "application/json") {
		return "rich_text"
	} else if strings.HasPrefix(mimeType, "text/") {
		return "text"
	} else if strings.Contains(mimeType, "pdf") {
		return "document"
	} else {
		return "file"
	}
}

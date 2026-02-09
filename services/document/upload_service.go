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
	ParentID    *uint // 父文件夹ID
}

// FolderUploadMetadata 文件夹上传元数据
type FolderUploadMetadata struct {
	Description string
	Category    string
	Tags        []string
	Department  string
	Project     string
	UploadedBy  string
	UploadIP    string
	ParentID    *uint // 父文件夹ID
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

	// 3. 验证文件格式（如果启用了格式限制）
	if !s.config.AllowAllFormats {
		if !s.isFormatAllowed(docType, format) {
			return nil, fmt.Errorf("不支持的文件格式: %s", format)
		}
	}

	// 4. 验证文件大小
	if file.Size > s.config.MaxFileSize {
		return nil, fmt.Errorf("文件大小超过限制: %.2f GB (最大 %.2f GB)", 
			float64(file.Size)/1024/1024/1024,
			float64(s.config.MaxFileSize)/1024/1024/1024)
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
		ParentID:    metadata.ParentID, // 支持父文件夹
		IsFolder:    false,             // 这是文件，不是文件夹
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

	// 10. 保存文件（使用检测到的格式作为扩展名）
	filePath, err := s.saveFile(file, document.ID, format)
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

// saveFile 保存文件到存储（流式处理，避免大文件占用内存）
func (s *UploadService) saveFile(file *multipart.FileHeader, documentID uint, format string) (string, error) {
	logger.Log.Infof("开始保存文件: documentID=%d, format=%s, size=%.2fMB", 
		documentID, format, float64(file.Size)/1024/1024)
	
	// 打开文件
	src, err := file.Open()
	if err != nil {
		logger.Log.Errorf("打开上传文件失败: %v", err)
		return "", fmt.Errorf("打开上传文件失败: %w", err)
	}
	defer src.Close()

	// 使用检测到的格式作为扩展名
	fileName := "file." + format

	// 使用年月日分组，避免单个文件夹文件过多
	// 格式: 2026/02/09/123
	now := time.Now()
	subPath := fmt.Sprintf("%d/%02d/%02d/%d", now.Year(), now.Month(), now.Day(), documentID)

	filePath, err := s.storageService.SaveFileStream(subPath, fileName, src, file.Size)
	if err != nil {
		logger.Log.Errorf("保存文件失败: documentID=%d, error=%v", documentID, err)
		return "", err
	}
	
	logger.Log.Infof("文件保存成功: documentID=%d, path=%s", documentID, filePath)
	return filePath, nil
}

// detectFileFormat 检测文件格式
func (s *UploadService) detectFileFormat(file *multipart.FileHeader) string {
	// 1. 优先从文件名获取扩展名
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(file.Filename), "."))
	if ext != "" {
		// 如果有扩展名，直接使用
		return ext
	}

	// 2. 如果没有扩展名，尝试通过 MIME 类型检测
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


// CreateFolder 创建文件夹
func (s *UploadService) CreateFolder(name, description string, parentID *uint, department, project, createdBy, createdIP string) (*models.Document, error) {
	// 验证父文件夹存在
	if parentID != nil {
		var parent models.Document
		if err := s.db.First(&parent, *parentID).Error; err != nil {
			return nil, fmt.Errorf("父文件夹不存在")
		}
		if !parent.IsFolder {
			return nil, fmt.Errorf("父级不是文件夹")
		}
	}

	// 创建文件夹记录
	folder := &models.Document{
		Name:        name,
		Description: description,
		Type:        models.TypeFolder,
		ParentID:    parentID,
		IsFolder:    true,
		Department:  department,
		Project:     project,
		UploadedBy:  createdBy,
		UploadIP:    createdIP,
	}

	if err := s.db.Create(folder).Error; err != nil {
		return nil, fmt.Errorf("创建文件夹失败: %w", err)
	}

	// 记录访问日志
	s.logAccess(folder.ID, "create_folder", createdBy, createdIP)

	return folder, nil
}


// UploadFolder 上传文件夹（保持结构）
func (s *UploadService) UploadFolder(files []*multipart.FileHeader, filePaths []string, metadata FolderUploadMetadata) (map[string]interface{}, error) {
	if len(files) != len(filePaths) {
		return nil, fmt.Errorf("文件数量与路径数量不匹配")
	}

	// 提取根文件夹名（第一层目录）
	var rootFolderName string
	if len(filePaths) > 0 {
		parts := strings.Split(strings.ReplaceAll(filePaths[0], "\\", "/"), "/")
		if len(parts) > 1 {
			rootFolderName = parts[0]
			logger.Log.Infof("检测到根文件夹: %s", rootFolderName)
		}
	}

	// 创建根文件夹
	rootFolder, err := s.CreateFolder(
		rootFolderName,
		metadata.Description,
		metadata.ParentID,
		metadata.Department,
		metadata.Project,
		metadata.UploadedBy,
		metadata.UploadIP,
	)
	if err != nil {
		return nil, fmt.Errorf("创建根文件夹失败: %w", err)
	}

	// 文件夹映射：路径 -> Document ID
	folderMap := make(map[string]uint)
	folderMap[rootFolderName] = rootFolder.ID

	// 统计信息
	uploadedFiles := 0
	createdFolders := 1 // 已创建根文件夹
	skippedFiles := 0
	var errors []string

	// 遍历所有文件
	for i, file := range files {
		originalPath := filePaths[i]
		
		// 去掉根文件夹前缀
		relativePath := originalPath
		if rootFolderName != "" {
			prefix := rootFolderName + "/"
			relativePath = strings.TrimPrefix(strings.ReplaceAll(originalPath, "\\", "/"), prefix)
		}

		// 解析路径，创建必要的文件夹
		parts := strings.Split(relativePath, "/")
		fileName := parts[len(parts)-1]
		
		// 确定父文件夹ID
		var parentFolderID uint = rootFolder.ID
		
		// 如果有子文件夹，逐层创建
		if len(parts) > 1 {
			currentPath := rootFolderName
			for j := 0; j < len(parts)-1; j++ {
				folderName := parts[j]
				currentPath = currentPath + "/" + folderName
				
				// 检查文件夹是否已创建
				if folderID, exists := folderMap[currentPath]; exists {
					parentFolderID = folderID
				} else {
					// 创建新文件夹
					newFolder, err := s.CreateFolder(
						folderName,
						"",
						&parentFolderID,
						metadata.Department,
						metadata.Project,
						metadata.UploadedBy,
						metadata.UploadIP,
					)
					if err != nil {
						logger.Log.Errorf("创建文件夹失败: %s, 错误: %v", folderName, err)
						errors = append(errors, fmt.Sprintf("创建文件夹 %s 失败: %v", folderName, err))
						continue
					}
					folderMap[currentPath] = newFolder.ID
					parentFolderID = newFolder.ID
					createdFolders++
				}
			}
		}

		// 上传文件
		fileMetadata := UploadMetadata{
			Name:        s.sanitizeFileName(fileName),
			Description: metadata.Description,
			Category:    metadata.Category,
			Tags:        metadata.Tags,
			Department:  metadata.Department,
			Project:     metadata.Project,
			UploadedBy:  metadata.UploadedBy,
			UploadIP:    metadata.UploadIP,
			ParentID:    &parentFolderID,
		}

		_, err := s.Upload(file, fileMetadata)
		if err != nil {
			logger.Log.Errorf("上传文件失败: %s, 错误: %v", fileName, err)
			errors = append(errors, fmt.Sprintf("上传文件 %s 失败: %v", fileName, err))
			skippedFiles++
		} else {
			uploadedFiles++
		}
	}

	// 返回结果
	result := map[string]interface{}{
		"root_folder_id":   rootFolder.ID,
		"root_folder_name": rootFolder.Name,
		"uploaded_files":   uploadedFiles,
		"created_folders":  createdFolders,
		"skipped_files":    skippedFiles,
		"total_files":      len(files),
	}

	if len(errors) > 0 {
		result["errors"] = errors
	}

	return result, nil
}

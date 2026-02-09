package services

import (
	"fmt"
	"go_wails_project_manager/config"
	"go_wails_project_manager/models"
	"go_wails_project_manager/utils"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gorm.io/gorm"
)

type ProjectService struct {
	db *gorm.DB
}

func NewProjectService(db *gorm.DB) *ProjectService {
	return &ProjectService{db: db}
}

// GetProjects 获取项目列表
func (ps *ProjectService) GetProjects(page, pageSize int, keyword string) ([]models.Project, int64, error) {
	var projects []models.Project
	var total int64

	query := ps.db.Model(&models.Project{})
	
	if keyword != "" {
		query = query.Where("name LIKE ? OR description LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("updated_at DESC").Offset(offset).Limit(pageSize).Find(&projects).Error; err != nil {
		return nil, 0, err
	}

	return projects, total, nil
}

// CreateProject 创建项目
func (ps *ProjectService) CreateProject(name, description string) (*models.Project, error) {
	project := &models.Project{
		Name:           name,
		Description:    description,
		CurrentVersion: config.ProjectAppConfig.DefaultInitialVersion,
	}

	if err := ps.db.Create(project).Error; err != nil {
		return nil, err
	}

	return project, nil
}

// GetProject 获取项目详情
func (ps *ProjectService) GetProject(id uint) (*models.Project, error) {
	var project models.Project
	if err := ps.db.First(&project, id).Error; err != nil {
		return nil, err
	}
	return &project, nil
}

// DeleteProject 删除项目
func (ps *ProjectService) DeleteProject(id uint) error {
	// 查询所有版本
	var versions []models.ProjectVersion
	if err := ps.db.Where("project_id = ?", id).Find(&versions).Error; err != nil {
		return err
	}

	// 删除所有版本文件（处理相对路径）
	for _, version := range versions {
		if version.FilePath != "" {
			absolutePath := resolveAbsolutePath(version.FilePath)
			os.Remove(absolutePath)
		}
		if version.ExtractedPath != "" {
			absolutePath := resolveAbsolutePath(version.ExtractedPath)
			os.RemoveAll(absolutePath)
		}
		if version.HistoryPath != "" {
			absolutePath := resolveAbsolutePath(version.HistoryPath)
			os.RemoveAll(absolutePath)
		}
	}

	// 删除版本记录
	if err := ps.db.Where("project_id = ?", id).Delete(&models.ProjectVersion{}).Error; err != nil {
		return err
	}

	// 删除项目
	return ps.db.Delete(&models.Project{}, id).Error
}

// GetVersionHistory 获取版本历史
func (ps *ProjectService) GetVersionHistory(projectID uint) ([]models.ProjectVersion, error) {
	var versions []models.ProjectVersion
	if err := ps.db.Where("project_id = ?", projectID).Order("created_at DESC").Find(&versions).Error; err != nil {
		return nil, err
	}
	return versions, nil
}

// RollbackVersion 回滚版本
func (ps *ProjectService) RollbackVersion(versionID uint) error {
	// 1. 获取版本信息
	var version models.ProjectVersion
	if err := ps.db.First(&version, versionID).Error; err != nil {
		return err
	}

	// 2. 获取项目信息
	var project models.Project
	if err := ps.db.First(&project, version.ProjectID).Error; err != nil {
		return err
	}

	// 3. 构建当前版本目录的绝对路径
	currentProjectDir := filepath.Join("static", "projects", project.Name)
	if config.ProjectAppConfig.NASEnabled {
		currentProjectDir = filepath.Join(config.ProjectAppConfig.NASPath, project.Name)
	}

	log.Printf("回滚: 清空当前版本目录: %s", currentProjectDir)

	// 4. 清空当前版本目录
	if err := os.RemoveAll(currentProjectDir); err != nil {
		return fmt.Errorf("清空当前版本目录失败: %v", err)
	}
	os.MkdirAll(currentProjectDir, 0755)

	// 5. 从历史版本的 zip 文件解压到当前目录
	zipPath := resolveAbsolutePath(version.FilePath)
	log.Printf("回滚: 从 zip 解压: %s -> %s", zipPath, currentProjectDir)

	if err := utils.ExtractArchive(zipPath, currentProjectDir); err != nil {
		return fmt.Errorf("解压历史版本失败: %v", err)
	}

	// 6. 更新项目当前版本
	if err := ps.db.Model(&project).Updates(map[string]interface{}{
		"current_version":   version.Version,
		"latest_version_id": versionID,
	}).Error; err != nil {
		return err
	}

	log.Printf("回滚成功: 项目 %d 已回滚到版本 %s", project.ID, version.Version)

	return nil
}

// UploadVersion 上传版本
func (ps *ProjectService) UploadVersion(projectID uint, username, description, versionType, tempDir, uploadIP string) (*models.ProjectVersion, error) {
	// 1. 获取项目信息
	var project models.Project
	if err := ps.db.First(&project, projectID).Error; err != nil {
		return nil, err
	}

	// 2. 计算新版本号
	newVersion, err := utils.CalculateNextVersion(project.CurrentVersion, versionType)
	if err != nil {
		return nil, err
	}

	// 3. 统计文件信息
	fileCount, err := utils.CountFiles(tempDir)
	if err != nil {
		return nil, err
	}

	folderSize, err := utils.GetFolderSize(tempDir)
	if err != nil {
		return nil, err
	}

	// 4. 创建存储目录
	// 当前版本目录（固定URL）
	currentProjectDir := filepath.Join("static", "projects", project.Name)
	// 历史版本目录
	historyProjectDir := filepath.Join("static", "project_histories", project.Name)
	
	if config.ProjectAppConfig.NASEnabled {
		currentProjectDir = filepath.Join(config.ProjectAppConfig.NASPath, project.Name)
		historyProjectDir = filepath.Join(config.ProjectAppConfig.NASHistoryPath, project.Name)
	}
	
	log.Printf("创建当前版本目录: %s", currentProjectDir)
	log.Printf("创建历史版本目录: %s", historyProjectDir)
	
	os.MkdirAll(currentProjectDir, 0755)
	os.MkdirAll(historyProjectDir, 0755)

	// 5. 压缩文件夹（保存到历史版本目录）
	versionDir := filepath.Join(historyProjectDir, fmt.Sprintf("v%s", newVersion))
	os.MkdirAll(versionDir, 0755)
	
	zipFileName := fmt.Sprintf("v%s.zip", newVersion)
	zipPath := filepath.Join(versionDir, zipFileName)
	
	log.Printf("压缩文件保存路径: %s", zipPath)
	
	if err := utils.CompressFolder(tempDir, zipPath); err != nil {
		return nil, err
	}
	
	log.Printf("压缩完成，文件大小: %d bytes", folderSize)

	// 6. 计算文件哈希
	fileHash, err := utils.CalculateHash(zipPath)
	if err != nil {
		return nil, err
	}

	// 7. 解压到历史版本目录
	extractedHistoryDir := filepath.Join(versionDir, "extracted")
	log.Printf("解压到历史版本目录: %s", extractedHistoryDir)
	
	if err := utils.ExtractArchive(zipPath, extractedHistoryDir); err != nil {
		return nil, err
	}
	
	log.Printf("历史版本解压完成")

	// 8. 清空当前版本目录并解压最新版本
	log.Printf("清空当前版本目录: %s", currentProjectDir)
	
	// 先删除当前目录的所有内容
	if err := os.RemoveAll(currentProjectDir); err != nil {
		log.Printf("清空当前版本目录失败: %v", err)
		return nil, fmt.Errorf("清空当前版本目录失败: %v", err)
	}
	os.MkdirAll(currentProjectDir, 0755)
	
	log.Printf("解压最新版本到当前目录: %s", currentProjectDir)
	
	// 解压到当前版本目录
	if err := utils.ExtractArchive(zipPath, currentProjectDir); err != nil {
		log.Printf("解压到当前版本目录失败: %v", err)
		return nil, err
	}
	
	log.Printf("当前版本解压完成")

	// 9. 创建版本记录（存储相对路径）
	// 将绝对路径转换为相对路径
	relativeCurrentPath := extractRelativePath(currentProjectDir)
	relativeHistoryPath := extractRelativePath(extractedHistoryDir)
	relativeZipPath := extractRelativePath(zipPath)
	
	version := &models.ProjectVersion{
		ProjectID:     projectID,
		Version:       newVersion,
		Username:      username,
		Description:   description,
		FilePath:      relativeZipPath,        // 相对路径
		FileSize:      folderSize,
		FileHash:      fileHash,
		FileCount:     fileCount,
		UploadIP:      uploadIP,
		ExtractedPath: relativeCurrentPath,    // 相对路径
		HistoryPath:   relativeHistoryPath,    // 相对路径
		CreatedAt:     time.Now(),
	}
	
	log.Printf("创建版本记录: 项目ID=%d, 版本=%s, 当前路径=%s, 历史路径=%s", 
		projectID, newVersion, relativeCurrentPath, relativeHistoryPath)

	if err := ps.db.Create(version).Error; err != nil {
		log.Printf("创建版本记录失败: %v", err)
		return nil, err
	}
	
	log.Printf("版本记录创建成功，版本ID: %d", version.ID)

	// 10. 更新项目当前版本
	if err := ps.db.Model(&project).Updates(map[string]interface{}{
		"current_version":   newVersion,
		"latest_version_id": version.ID,
	}).Error; err != nil {
		log.Printf("更新项目当前版本失败: %v", err)
		return nil, err
	}
	
	log.Printf("项目当前版本更新成功: %s", newVersion)

	// 11. 异步生成预览截图
	go ps.generateScreenshotAsync(version)

	return version, nil
}

// generateScreenshotAsync 异步生成预览截图
func (ps *ProjectService) generateScreenshotAsync(version *models.ProjectVersion) {
	// 构建预览URL
	previewURL := buildLocalPreviewURL(version.ExtractedPath)
	if previewURL == "" {
		log.Printf("版本 %d 没有预览URL，跳过截图", version.ID)
		return
	}

	// 将相对路径转换为绝对路径
	absoluteExtractedPath := resolveAbsolutePath(version.ExtractedPath)
	
	// 生成缩略图路径（绝对路径）
	thumbnailPath := filepath.Join(absoluteExtractedPath, "thumbnail.png")

	log.Printf("开始为版本 %d 生成预览截图: %s", version.ID, previewURL)
	log.Printf("截图保存路径: %s", thumbnailPath)

	// 生成缩略图（1200x800）
	if err := utils.GenerateThumbnail(previewURL, thumbnailPath, 1200, 800); err != nil {
		log.Printf("生成预览截图失败 (版本 %d): %v", version.ID, err)
		return
	}

	log.Printf("预览截图生成成功: %s", thumbnailPath)

	// 转换为相对路径
	relativeThumbnailPath := extractRelativePath(thumbnailPath)

	// 更新版本记录的缩略图路径
	if err := ps.db.Model(version).Update("thumbnail_path", relativeThumbnailPath).Error; err != nil {
		log.Printf("更新版本缩略图路径失败 (版本 %d): %v", version.ID, err)
		return
	}

	// 同时更新项目表的缩略图（使用最新版本的截图）
	if err := ps.db.Model(&models.Project{}).Where("id = ?", version.ProjectID).Update("thumbnail_path", relativeThumbnailPath).Error; err != nil {
		log.Printf("更新项目缩略图路径失败 (项目 %d): %v", version.ProjectID, err)
		return
	}

	log.Printf("版本 %d 的预览截图已保存到数据库，并更新到项目 %d", version.ID, version.ProjectID)
}

// buildLocalPreviewURL 构建本地预览URL（用于截图）
func buildLocalPreviewURL(extractedPath string) string {
	if extractedPath == "" {
		return ""
	}
	
	// 将相对路径转换为绝对路径
	absolutePath := resolveAbsolutePath(extractedPath)

	// 检查 index.html 是否存在
	indexPath := filepath.Join(absolutePath, "index.html")
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		return ""
	}

	// 统一路径分隔符
	path := filepath.ToSlash(extractedPath)
	
	// 提取相对路径（从 projects/ 之后的部分）
	projectsIdx := strings.LastIndex(path, "/projects/")
	if projectsIdx == -1 {
		projectsIdx = strings.LastIndex(path, "projects/")
		if projectsIdx != -1 {
			projectsIdx += len("projects/")
		}
	} else {
		projectsIdx += len("/projects/")
	}
	
	var relativePath string
	if projectsIdx != -1 {
		relativePath = path[projectsIdx:]
	} else {
		// 如果找不到 projects/，使用最后两段路径
		parts := strings.Split(path, "/")
		if len(parts) >= 2 {
			relativePath = parts[len(parts)-2] + "/" + parts[len(parts)-1]
		} else {
			relativePath = path
		}
	}

	// 构建完整的预览URL
	if config.ProjectAppConfig != nil && config.ProjectAppConfig.BaseURL != "" {
		baseURL := strings.TrimSuffix(config.ProjectAppConfig.BaseURL, "/")
		return fmt.Sprintf("%s/%s/index.html", baseURL, relativePath)
	}

	return ""
}

// extractRelativePath 从绝对路径提取相对路径（相对于 static 目录）
// 例如: /vol1/1003/project/editor_v2/static/projects/项目名 -> projects/项目名
//      \\192.168.3.10\project\editor_v2\static\projects\项目名 -> projects/项目名
func extractRelativePath(absolutePath string) string {
	// 统一路径分隔符
	path := filepath.ToSlash(absolutePath)
	
	// 查找 static/ 的位置
	staticIdx := strings.LastIndex(path, "/static/")
	if staticIdx == -1 {
		staticIdx = strings.LastIndex(path, "static/")
		if staticIdx != -1 {
			// 返回 static/ 之后的部分
			return path[staticIdx+7:] // 跳过 "static/"
		}
	} else {
		// 返回 static/ 之后的部分
		return path[staticIdx+8:] // 跳过 "/static/"
	}
	
	// 如果找不到 static/，返回原路径
	return path
}

// resolveAbsolutePath 将相对路径转换为绝对路径
// 例如: projects/项目名 -> /vol1/1003/project/editor_v2/static/projects/项目名
func resolveAbsolutePath(relativePath string) string {
	// 如果已经是绝对路径，直接返回
	if filepath.IsAbs(relativePath) {
		return relativePath
	}
	
	// 如果是 Windows UNC 路径
	if strings.HasPrefix(relativePath, "\\\\") {
		return relativePath
	}
	
	// 构建绝对路径
	var basePath string
	if config.ProjectAppConfig != nil && config.ProjectAppConfig.NASEnabled {
		// 使用 NAS 路径的父目录
		basePath = filepath.Dir(config.ProjectAppConfig.NASPath)
	} else {
		basePath = "static"
	}
	
	return filepath.Join(basePath, relativePath)
}

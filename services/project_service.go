package services

import (
	"fmt"
	"go_wails_project_manager/config"
	"go_wails_project_manager/models"
	"go_wails_project_manager/utils"
	"os"
	"path/filepath"
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

	// 删除所有版本文件
	for _, version := range versions {
		if version.FilePath != "" {
			os.Remove(version.FilePath)
		}
		if version.ExtractedPath != "" {
			os.RemoveAll(version.ExtractedPath)
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
	var version models.ProjectVersion
	if err := ps.db.First(&version, versionID).Error; err != nil {
		return err
	}

	return ps.db.Model(&models.Project{}).Where("id = ?", version.ProjectID).Updates(map[string]interface{}{
		"current_version":   version.Version,
		"latest_version_id": versionID,
	}).Error
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
	projectDir := filepath.Join("static", "projects", project.Name)
	if config.ProjectAppConfig.NASEnabled {
		projectDir = filepath.Join(config.ProjectAppConfig.NASPath, project.Name)
	}
	os.MkdirAll(projectDir, 0755)

	// 5. 压缩文件夹
	zipFileName := fmt.Sprintf("v%s.zip", newVersion)
	zipPath := filepath.Join(projectDir, zipFileName)
	
	if err := utils.CompressFolder(tempDir, zipPath); err != nil {
		return nil, err
	}

	// 6. 计算文件哈希
	fileHash, err := utils.CalculateHash(zipPath)
	if err != nil {
		return nil, err
	}

	// 7. 解压到预览目录（用于网页预览）
	extractedDir := filepath.Join(projectDir, fmt.Sprintf("v%s", newVersion))
	if err := utils.ExtractArchive(zipPath, extractedDir); err != nil {
		return nil, err
	}

	// 8. 创建版本记录
	version := &models.ProjectVersion{
		ProjectID:     projectID,
		Version:       newVersion,
		Username:      username,
		Description:   description,
		FilePath:      zipPath,
		FileSize:      folderSize,
		FileHash:      fileHash,
		FileCount:     fileCount,
		UploadIP:      uploadIP,
		ExtractedPath: extractedDir,
		CreatedAt:     time.Now(),
	}

	if err := ps.db.Create(version).Error; err != nil {
		return nil, err
	}

	// 9. 更新项目当前版本
	if err := ps.db.Model(&project).Updates(map[string]interface{}{
		"current_version":   newVersion,
		"latest_version_id": version.ID,
	}).Error; err != nil {
		return nil, err
	}

	return version, nil
}

package controllers

import (
	"encoding/json"
	"go_wails_project_manager/database"
	"go_wails_project_manager/logger"
	"go_wails_project_manager/models"
	"go_wails_project_manager/response"
	"go_wails_project_manager/services"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ProjectController struct {
	service *services.ProjectService
}

func NewProjectController(db *gorm.DB) *ProjectController {
	return &ProjectController{
		service: services.NewProjectService(db),
	}
}

// GetProjects 获取项目列表
func (pc *ProjectController) GetProjects(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	keyword := c.Query("keyword")

	projects, total, err := pc.service.GetProjects(page, pageSize, keyword)
	if err != nil {
		response.Error(c, response.CodeInternalServerError, "获取项目列表失败")
		return
	}

	response.Success(c, gin.H{
		"total":     total,
		"page":      page,
		"page_size": pageSize,
		"data":      projects,
	})
}

// CreateProject 创建项目
func (pc *ProjectController) CreateProject(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required,max=200"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.CodeBadRequest, "参数错误")
		return
	}

	project, err := pc.service.CreateProject(req.Name, req.Description)
	if err != nil {
		response.Error(c, response.CodeInternalServerError, "创建项目失败")
		return
	}

	response.Success(c, project)
}

// GetProject 获取项目详情
func (pc *ProjectController) GetProject(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	
	project, err := pc.service.GetProject(uint(id))
	if err != nil {
		response.Error(c, response.CodeNotFound, "项目不存在")
		return
	}

	response.Success(c, project)
}

// DeleteProject 删除项目
func (pc *ProjectController) DeleteProject(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	if err := pc.service.DeleteProject(uint(id)); err != nil {
		response.Error(c, response.CodeInternalServerError, "删除项目失败")
		return
	}

	response.Success(c, nil)
}

// UploadVersion 上传版本
func (pc *ProjectController) UploadVersion(c *gin.Context) {
	projectID, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	username := c.PostForm("username")
	description := c.PostForm("description")
	versionType := c.PostForm("version_type")
	filePathsJSON := c.PostForm("file_paths")

	if username == "" || versionType == "" {
		response.Error(c, response.CodeBadRequest, "username和version_type不能为空")
		return
	}

	// 解析文件路径列表
	var filePaths []string
	if err := json.Unmarshal([]byte(filePathsJSON), &filePaths); err != nil {
		response.Error(c, response.CodeBadRequest, "文件路径解析失败")
		return
	}

	// 获取上传的文件
	form, err := c.MultipartForm()
	if err != nil {
		response.Error(c, response.CodeBadRequest, "获取文件失败")
		return
	}

	files := form.File["files"]
	if len(files) == 0 {
		response.Error(c, response.CodeBadRequest, "没有上传文件")
		return
	}

	if len(files) != len(filePaths) {
		response.Error(c, response.CodeBadRequest, "文件数量与路径数量不匹配")
		return
	}

	// 创建临时目录
	tempDir := filepath.Join(os.TempDir(), uuid.New().String())
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		response.Error(c, response.CodeInternalServerError, "创建临时目录失败")
		return
	}
	defer os.RemoveAll(tempDir)

	// 保存文件到临时目录（使用前端传来的路径信息）
	// 需要去掉第一层目录（文件夹容器名）
	var rootFolder string
	for i, file := range files {
		// 使用前端传来的相对路径
		originalPath := filePaths[i]
		
		// 提取根文件夹名（第一次循环时）
		if i == 0 {
			parts := strings.Split(strings.ReplaceAll(originalPath, "\\", "/"), "/")
			if len(parts) > 1 {
				rootFolder = parts[0]
				logger.Log.Infof("检测到根文件夹: %s，将自动去除", rootFolder)
			}
		}
		
		// 去掉根文件夹前缀
		relativePath := originalPath
		if rootFolder != "" {
			// 去掉 "local_dist/" 前缀
			prefix := rootFolder + "/"
			if strings.HasPrefix(strings.ReplaceAll(originalPath, "\\", "/"), prefix) {
				relativePath = strings.TrimPrefix(strings.ReplaceAll(originalPath, "\\", "/"), prefix)
			}
		}
		
		destPath := filepath.Join(tempDir, relativePath)

		// 打印调试信息
		logger.Log.Infof("上传文件: %s -> %s (原路径: %s)", relativePath, destPath, originalPath)

		// 创建目录
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			response.Error(c, response.CodeInternalServerError, "创建目录失败")
			return
		}

		// 保存文件
		if err := c.SaveUploadedFile(file, destPath); err != nil {
			response.Error(c, response.CodeInternalServerError, "保存文件失败")
			return
		}
	}

	// 上传版本
	uploadIP := c.ClientIP()
	version, err := pc.service.UploadVersion(uint(projectID), username, description, versionType, tempDir, uploadIP)
	if err != nil {
		response.Error(c, response.CodeInternalServerError, "上传版本失败")
		return
	}

	response.Success(c, version)
}

// GetVersionHistory 获取版本历史
func (pc *ProjectController) GetVersionHistory(c *gin.Context) {
	projectID, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	versions, err := pc.service.GetVersionHistory(uint(projectID))
	if err != nil {
		response.Error(c, response.CodeInternalServerError, "获取版本历史失败")
		return
	}

	response.Success(c, versions)
}

// DownloadVersion 下载版本
func (pc *ProjectController) DownloadVersion(c *gin.Context) {
	versionID, _ := strconv.ParseUint(c.Param("versionId"), 10, 32)

	var version models.ProjectVersion
	
	db, _ := database.GetDB()
	if err := db.First(&version, versionID).Error; err != nil {
		response.Error(c, response.CodeNotFound, "版本不存在")
		return
	}

	c.FileAttachment(version.FilePath, filepath.Base(version.FilePath))
}

// RollbackVersion 回滚版本
func (pc *ProjectController) RollbackVersion(c *gin.Context) {
	versionID, _ := strconv.ParseUint(c.Param("versionId"), 10, 32)

	if err := pc.service.RollbackVersion(uint(versionID)); err != nil {
		response.Error(c, response.CodeInternalServerError, "回滚版本失败")
		return
	}

	response.Success(c, gin.H{"message": "回滚成功"})
}

package requirement

import (
	"go_wails_project_manager/response"
	"go_wails_project_manager/services/requirement_service"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// ProjectController 项目管理控制器
type ProjectController struct {
	service *requirement_service.ProjectService
}

// NewProjectController 创建项目控制器
func NewProjectController() *ProjectController {
	return &ProjectController{
		service: requirement_service.NewProjectService(),
	}
}

// CreateProjectRequest 创建项目请求
type CreateProjectRequest struct {
	CompanyID   uint       `json:"company_id" binding:"required"`
	Name        string     `json:"name" binding:"required,min=2,max=100"`
	Key         string     `json:"key" binding:"required,min=2,max=20"`
	Description string     `json:"description"`
	StartDate   *time.Time `json:"start_date"`
	EndDate     *time.Time `json:"end_date"`
}

// UpdateProjectRequest 更新项目请求
type UpdateProjectRequest struct {
	Name        string     `json:"name" binding:"omitempty,min=2,max=100"`
	Description string     `json:"description"`
	OwnerID     uint       `json:"owner_id"`
	Status      string     `json:"status" binding:"omitempty,oneof=active archived"`
	StartDate   *time.Time `json:"start_date"`
	EndDate     *time.Time `json:"end_date"`
}

// AddProjectMemberRequest 添加项目成员请求
type AddProjectMemberRequest struct {
	UserID uint   `json:"user_id" binding:"required"`
	Role   string `json:"role" binding:"required,oneof=project_admin developer viewer"`
}

// Create 创建项目
func (pc *ProjectController) Create(c *gin.Context) {
	var req CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	userID := c.GetUint("user_id")

	project, err := pc.service.CreateProject(userID, req.CompanyID, req.Name, req.Key, req.Description, req.StartDate, req.EndDate)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.Success(c, project)
}

// List 获取项目列表
func (pc *ProjectController) List(c *gin.Context) {
	companyID, _ := strconv.ParseUint(c.Query("company_id"), 10, 32)
	userID := c.GetUint("user_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	status := c.Query("status")

	projects, total, err := pc.service.ListUserProjects(userID, uint(companyID), page, pageSize, status)
	if err != nil {
		response.InternalServerError(c, "查询失败")
		return
	}

	response.Success(c, gin.H{
		"items":     projects,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetDetail 获取项目详情
func (pc *ProjectController) GetDetail(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	userID := c.GetUint("user_id")

	// 检查访问权限
	if !pc.service.HasAccess(uint(id), userID) {
		response.Forbidden(c, "无权访问该项目")
		return
	}

	project, err := pc.service.GetProjectDetail(uint(id))
	if err != nil {
		response.NotFound(c, "项目不存在")
		return
	}

	response.Success(c, project)
}

// Update 更新项目
func (pc *ProjectController) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	userID := c.GetUint("user_id")

	var req UpdateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	// 检查是否是项目管理员
	if !pc.service.IsAdmin(uint(id), userID) {
		response.Forbidden(c, "只有项目管理员可以更新项目")
		return
	}

	project, err := pc.service.UpdateProject(uint(id), req.Name, req.Description, req.OwnerID, req.Status, req.StartDate, req.EndDate)
	if err != nil {
		response.InternalServerError(c, "更新失败")
		return
	}

	response.Success(c, project)
}

// AddMember 添加项目成员
func (pc *ProjectController) AddMember(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	userID := c.GetUint("user_id")

	var req AddProjectMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	// 检查是否是项目管理员
	if !pc.service.IsAdmin(uint(id), userID) {
		response.Forbidden(c, "只有项目管理员可以添加成员")
		return
	}

	member, err := pc.service.AddMember(uint(id), req.UserID, req.Role)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.Success(c, member)
}

// RemoveMember 移除项目成员
func (pc *ProjectController) RemoveMember(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	memberUserID, _ := strconv.ParseUint(c.Param("user_id"), 10, 32)
	userID := c.GetUint("user_id")

	// 检查是否是项目管理员
	if !pc.service.IsAdmin(uint(id), userID) {
		response.Forbidden(c, "只有项目管理员可以移除成员")
		return
	}

	if err := pc.service.RemoveMember(uint(id), uint(memberUserID)); err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.SuccessWithMsg(c, "移除成功", nil)
}

// GetMembers 获取项目成员列表
func (pc *ProjectController) GetMembers(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	userID := c.GetUint("user_id")

	// 检查访问权限
	if !pc.service.HasAccess(uint(id), userID) {
		response.Forbidden(c, "无权访问该项目")
		return
	}

	members, err := pc.service.GetMembers(uint(id))
	if err != nil {
		response.InternalServerError(c, "查询失败")
		return
	}

	response.Success(c, members)
}

// GetStatistics 获取项目统计
func (pc *ProjectController) GetStatistics(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	userID := c.GetUint("user_id")

	// 检查访问权限
	if !pc.service.HasAccess(uint(id), userID) {
		response.Forbidden(c, "无权访问该项目")
		return
	}

	stats, err := pc.service.GetStatistics(uint(id))
	if err != nil {
		response.InternalServerError(c, "查询失败")
		return
	}

	response.Success(c, stats)
}

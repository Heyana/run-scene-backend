package requirement

import (
	"go_wails_project_manager/response"
	"go_wails_project_manager/services/requirement_service"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// MissionListController 任务列表控制器
type MissionListController struct {
	service        *requirement_service.MissionListService
	projectService *requirement_service.ProjectService
}

// NewMissionListController 创建任务列表控制器
func NewMissionListController() *MissionListController {
	return &MissionListController{
		service:        requirement_service.NewMissionListService(),
		projectService: requirement_service.NewProjectService(),
	}
}

// CreateMissionListRequest 创建任务列表请求
type CreateMissionListRequest struct {
	ProjectID   uint       `json:"project_id" binding:"required"`
	Name        string     `json:"name" binding:"required,min=2,max=100"`
	Type        string     `json:"type" binding:"required,oneof=sprint version module"`
	Description string     `json:"description"`
	StartDate   *time.Time `json:"start_date"`
	EndDate     *time.Time `json:"end_date"`
}

// UpdateMissionListRequest 更新任务列表请求
type UpdateMissionListRequest struct {
	Name        string     `json:"name" binding:"omitempty,min=2,max=100"`
	Description string     `json:"description"`
	Status      string     `json:"status" binding:"omitempty,oneof=planning active completed"`
	StartDate   *time.Time `json:"start_date"`
	EndDate     *time.Time `json:"end_date"`
}

// Create 创建任务列表
func (mlc *MissionListController) Create(c *gin.Context) {
	var req CreateMissionListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	userID := c.GetUint("user_id")

	// 检查项目访问权限
	if !mlc.projectService.HasAccess(req.ProjectID, userID) {
		response.Forbidden(c, "无权访问该项目")
		return
	}

	missionList, err := mlc.service.CreateMissionList(req.ProjectID, req.Name, req.Type, req.Description, req.StartDate, req.EndDate)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.Success(c, missionList)
}

// List 获取任务列表
func (mlc *MissionListController) List(c *gin.Context) {
	projectID, _ := strconv.ParseUint(c.Query("project_id"), 10, 32)
	userID := c.GetUint("user_id")
	status := c.Query("status")

	// 检查项目访问权限
	if !mlc.projectService.HasAccess(uint(projectID), userID) {
		response.Forbidden(c, "无权访问该项目")
		return
	}

	lists, err := mlc.service.ListMissionLists(uint(projectID), status)
	if err != nil {
		response.InternalServerError(c, "查询失败")
		return
	}

	response.Success(c, lists)
}

// GetDetail 获取任务列表详情
func (mlc *MissionListController) GetDetail(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	userID := c.GetUint("user_id")

	missionList, err := mlc.service.GetMissionListDetail(uint(id))
	if err != nil {
		response.NotFound(c, "任务列表不存在")
		return
	}

	// 检查项目访问权限
	if !mlc.projectService.HasAccess(missionList.ProjectID, userID) {
		response.Forbidden(c, "无权访问该项目")
		return
	}

	response.Success(c, missionList)
}

// Update 更新任务列表
func (mlc *MissionListController) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	userID := c.GetUint("user_id")

	var req UpdateMissionListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	missionList, err := mlc.service.GetMissionListDetail(uint(id))
	if err != nil {
		response.NotFound(c, "任务列表不存在")
		return
	}

	// 检查是否是项目管理员
	if !mlc.projectService.IsAdmin(missionList.ProjectID, userID) {
		response.Forbidden(c, "只有项目管理员可以更新任务列表")
		return
	}

	updatedList, err := mlc.service.UpdateMissionList(uint(id), req.Name, req.Description, req.Status, req.StartDate, req.EndDate)
	if err != nil {
		response.InternalServerError(c, "更新失败")
		return
	}

	response.Success(c, updatedList)
}

// Delete 删除任务列表
func (mlc *MissionListController) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	userID := c.GetUint("user_id")

	missionList, err := mlc.service.GetMissionListDetail(uint(id))
	if err != nil {
		response.NotFound(c, "任务列表不存在")
		return
	}

	// 检查是否是项目管理员
	if !mlc.projectService.IsAdmin(missionList.ProjectID, userID) {
		response.Forbidden(c, "只有项目管理员可以删除任务列表")
		return
	}

	if err := mlc.service.DeleteMissionList(uint(id)); err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.SuccessWithMsg(c, "删除成功", nil)
}

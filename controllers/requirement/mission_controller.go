package requirement

import (
	"go_wails_project_manager/response"
	"go_wails_project_manager/services/requirement_service"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// MissionController 任务控制器
type MissionController struct {
	service        *requirement_service.MissionService
	projectService *requirement_service.ProjectService
}

// NewMissionController 创建任务控制器
func NewMissionController() *MissionController {
	return &MissionController{
		service:        requirement_service.NewMissionService(),
		projectService: requirement_service.NewProjectService(),
	}
}

// CreateMissionRequestBody 创建任务请求
type CreateMissionRequestBody struct {
	MissionListID  uint       `json:"mission_list_id" binding:"required"`
	Title          string     `json:"title" binding:"required,min=2,max=200"`
	Description    string     `json:"description"`
	Type           string     `json:"type" binding:"required,oneof=feature enhancement bug"`
	Priority       string     `json:"priority" binding:"omitempty,oneof=P0 P1 P2 P3"`
	AssigneeID     *uint      `json:"assignee_id"`
	EstimatedHours float64    `json:"estimated_hours"`
	StartDate      *time.Time `json:"start_date"`
	DueDate        *time.Time `json:"due_date"`
}

// UpdateMissionRequestBody 更新任务请求
type UpdateMissionRequestBody struct {
	Title          string     `json:"title" binding:"omitempty,min=2,max=200"`
	Description    string     `json:"description"`
	Type           string     `json:"type" binding:"omitempty,oneof=feature enhancement bug"`
	Priority       string     `json:"priority" binding:"omitempty,oneof=P0 P1 P2 P3"`
	Status         string     `json:"status" binding:"omitempty,oneof=todo in_progress done closed"`
	AssigneeID     *uint      `json:"assignee_id"`
	EstimatedHours float64    `json:"estimated_hours"`
	ActualHours    float64    `json:"actual_hours"`
	StartDate      *time.Time `json:"start_date"`
	DueDate        *time.Time `json:"due_date"`
}

// AddCommentRequest 添加评论请求
type AddCommentRequest struct {
	Content  string `json:"content" binding:"required"`
	ParentID *uint  `json:"parent_id"`
}

// Create 创建任务
func (mc *MissionController) Create(c *gin.Context) {
	var req CreateMissionRequestBody
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	userID := c.GetUint("user_id")

	createReq := requirement_service.CreateMissionRequest{
		MissionListID:  req.MissionListID,
		Title:          req.Title,
		Description:    req.Description,
		Type:           req.Type,
		Priority:       req.Priority,
		AssigneeID:     req.AssigneeID,
		EstimatedHours: req.EstimatedHours,
		StartDate:      req.StartDate,
		DueDate:        req.DueDate,
	}

	mission, err := mc.service.CreateMission(userID, createReq)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.Success(c, mission)
}

// List 获取任务列表
func (mc *MissionController) List(c *gin.Context) {
	missionListID, _ := strconv.ParseUint(c.Query("mission_list_id"), 10, 32)
	projectID, _ := strconv.ParseUint(c.Query("project_id"), 10, 32)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	// 构建筛选条件
	filters := make(map[string]interface{})
	if status := c.Query("status"); status != "" {
		filters["status"] = status
	}
	if priority := c.Query("priority"); priority != "" {
		filters["priority"] = priority
	}
	if assigneeID := c.Query("assignee_id"); assigneeID != "" {
		id, _ := strconv.ParseUint(assigneeID, 10, 32)
		filters["assignee_id"] = uint(id)
	}
	if missionType := c.Query("type"); missionType != "" {
		filters["type"] = missionType
	}
	if keyword := c.Query("keyword"); keyword != "" {
		filters["keyword"] = keyword
	}

	var missions interface{}
	var total int64
	var err error

	// 支持按 project_id 或 mission_list_id 查询
	if projectID > 0 {
		missions, total, err = mc.service.ListMissionsByProject(uint(projectID), page, pageSize, filters)
	} else if missionListID > 0 {
		missions, total, err = mc.service.ListMissions(uint(missionListID), page, pageSize, filters)
	} else {
		response.BadRequest(c, "请提供 project_id 或 mission_list_id")
		return
	}

	if err != nil {
		response.InternalServerError(c, "查询失败")
		return
	}

	response.Success(c, gin.H{
		"items":     missions,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetDetail 获取任务详情
func (mc *MissionController) GetDetail(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	mission, err := mc.service.GetMissionDetail(uint(id))
	if err != nil {
		response.NotFound(c, "任务不存在")
		return
	}

	response.Success(c, mission)
}

// Update 更新任务
func (mc *MissionController) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	userID := c.GetUint("user_id")

	var req UpdateMissionRequestBody
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	// 构建更新字段
	updates := make(map[string]interface{})
	if req.Title != "" {
		updates["title"] = req.Title
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.Type != "" {
		updates["type"] = req.Type
	}
	if req.Priority != "" {
		updates["priority"] = req.Priority
	}
	if req.Status != "" {
		updates["status"] = req.Status
	}
	if req.AssigneeID != nil {
		updates["assignee_id"] = req.AssigneeID
	}
	if req.EstimatedHours > 0 {
		updates["estimated_hours"] = req.EstimatedHours
	}
	if req.ActualHours > 0 {
		updates["actual_hours"] = req.ActualHours
	}
	if req.StartDate != nil {
		updates["start_date"] = req.StartDate
	}
	if req.DueDate != nil {
		updates["due_date"] = req.DueDate
	}

	mission, err := mc.service.UpdateMission(uint(id), userID, updates)
	if err != nil {
		response.InternalServerError(c, "更新失败")
		return
	}

	response.Success(c, mission)
}

// UpdateStatus 更新任务状态
func (mc *MissionController) UpdateStatus(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	userID := c.GetUint("user_id")

	var req struct {
		Status string `json:"status" binding:"required,oneof=todo in_progress done closed"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	if err := mc.service.UpdateStatus(uint(id), userID, req.Status); err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.SuccessWithMsg(c, "状态更新成功", nil)
}

// Delete 删除任务
func (mc *MissionController) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	userID := c.GetUint("user_id")

	if err := mc.service.DeleteMission(uint(id), userID); err != nil {
		response.InternalServerError(c, "删除失败")
		return
	}

	response.SuccessWithMsg(c, "删除成功", nil)
}

// AddComment 添加评论
func (mc *MissionController) AddComment(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	userID := c.GetUint("user_id")

	var req AddCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	comment, err := mc.service.AddComment(uint(id), userID, req.Content, req.ParentID)
	if err != nil {
		response.InternalServerError(c, "添加评论失败")
		return
	}

	response.Success(c, comment)
}

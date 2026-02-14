package requirement

import (
	"go_wails_project_manager/response"
	"go_wails_project_manager/models/requirement"
	"go_wails_project_manager/services/requirement_service"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// MissionColumnController 任务列控制器
type MissionColumnController struct {
	db *gorm.DB
}

// NewMissionColumnController 创建任务列控制器
func NewMissionColumnController(db *gorm.DB) *MissionColumnController {
	return &MissionColumnController{db: db}
}

// CreateMissionColumn 创建任务列
func (mcc *MissionColumnController) CreateMissionColumn(c *gin.Context) {
	var req struct {
		MissionListID uint   `json:"mission_list_id" binding:"required"`
		Name          string `json:"name" binding:"required"`
		Color         string `json:"color"`
		SortOrder     int    `json:"sort_order"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	column := &requirement.MissionColumn{
		MissionListID: req.MissionListID,
		Name:          req.Name,
		Color:         req.Color,
		SortOrder:     req.SortOrder,
	}

	if column.Color == "" {
		column.Color = "#1890ff"
	}

	if err := requirement_service.CreateMissionColumn(mcc.db, column); err != nil {
		response.InternalServerError(c, "创建失败: "+err.Error())
		return
	}

	response.Success(c, column)
}

// GetMissionColumnList 获取任务列列表
func (mcc *MissionColumnController) GetMissionColumnList(c *gin.Context) {
	missionListIDStr := c.Query("mission_list_id")
	if missionListIDStr == "" {
		response.BadRequest(c, "缺少 mission_list_id 参数")
		return
	}

	missionListID, err := strconv.ParseUint(missionListIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "mission_list_id 参数无效")
		return
	}

	columns, err := requirement_service.GetMissionColumnList(mcc.db, uint(missionListID))
	if err != nil {
		response.InternalServerError(c, "查询失败: "+err.Error())
		return
	}

	response.Success(c, columns)
}

// GetMissionColumnDetail 获取任务列详情
func (mcc *MissionColumnController) GetMissionColumnDetail(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "ID参数无效")
		return
	}

	column, err := requirement_service.GetMissionColumnByID(mcc.db, uint(id))
	if err != nil {
		response.NotFound(c, "任务列不存在")
		return
	}

	response.Success(c, column)
}

// UpdateMissionColumn 更新任务列
func (mcc *MissionColumnController) UpdateMissionColumn(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "ID参数无效")
		return
	}

	column, err := requirement_service.GetMissionColumnByID(mcc.db, uint(id))
	if err != nil {
		response.NotFound(c, "任务列不存在")
		return
	}

	var req struct {
		Name      string `json:"name"`
		Color     string `json:"color"`
		SortOrder *int   `json:"sort_order"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	if req.Name != "" {
		column.Name = req.Name
	}
	if req.Color != "" {
		column.Color = req.Color
	}
	if req.SortOrder != nil {
		column.SortOrder = *req.SortOrder
	}

	if err := requirement_service.UpdateMissionColumn(mcc.db, column); err != nil {
		response.InternalServerError(c, "更新失败: "+err.Error())
		return
	}

	response.Success(c, column)
}

// DeleteMissionColumn 删除任务列
func (mcc *MissionColumnController) DeleteMissionColumn(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "ID参数无效")
		return
	}

	if err := requirement_service.DeleteMissionColumn(mcc.db, uint(id)); err != nil {
		response.InternalServerError(c, "删除失败: "+err.Error())
		return
	}

	response.Success(c, nil)
}

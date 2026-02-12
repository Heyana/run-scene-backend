package controllers

import (
	"go_wails_project_manager/response"
	"go_wails_project_manager/services/audit"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// AuditController 审计控制器
type AuditController struct {
	queryService   *audit.QueryService
	archiveService *audit.ArchiveService
}

// NewAuditController 创建审计控制器
func NewAuditController(queryService *audit.QueryService, archiveService *audit.ArchiveService) *AuditController {
	return &AuditController{
		queryService:   queryService,
		archiveService: archiveService,
	}
}

// ListLogs 查询审计日志列表
// @Summary 查询审计日志列表
// @Tags 审计
// @Param user_id query int false "用户ID"
// @Param username query string false "用户名"
// @Param user_ip query string false "用户IP"
// @Param action query string false "操作类型"
// @Param resource query string false "资源类型"
// @Param resource_id query int false "资源ID"
// @Param start_time query string false "开始时间"
// @Param end_time query string false "结束时间"
// @Param status_code query int false "状态码"
// @Param page query int false "页码"
// @Param page_size query int false "每页数量"
// @Success 200 {object} response.Response
// @Router /api/audit/logs [get]
func (ctrl *AuditController) ListLogs(c *gin.Context) {
	filter := audit.AuditFilter{
		Username: c.Query("username"),
		UserIP:   c.Query("user_ip"),
		Action:   c.Query("action"),
		Resource: c.Query("resource"),
	}

	// 解析可选参数
	if userID := c.Query("user_id"); userID != "" {
		if id, err := strconv.ParseUint(userID, 10, 32); err == nil {
			uid := uint(id)
			filter.UserID = &uid
		}
	}

	if resourceID := c.Query("resource_id"); resourceID != "" {
		if id, err := strconv.ParseUint(resourceID, 10, 32); err == nil {
			rid := uint(id)
			filter.ResourceID = &rid
		}
	}

	if statusCode := c.Query("status_code"); statusCode != "" {
		if code, err := strconv.Atoi(statusCode); err == nil {
			filter.StatusCode = &code
		}
	}

	if startTime := c.Query("start_time"); startTime != "" {
		if t, err := time.Parse(time.RFC3339, startTime); err == nil {
			filter.StartTime = &t
		}
	}

	if endTime := c.Query("end_time"); endTime != "" {
		if t, err := time.Parse(time.RFC3339, endTime); err == nil {
			filter.EndTime = &t
		}
	}

	if page := c.Query("page"); page != "" {
		if p, err := strconv.Atoi(page); err == nil {
			filter.Page = p
		}
	}

	if pageSize := c.Query("page_size"); pageSize != "" {
		if ps, err := strconv.Atoi(pageSize); err == nil {
			filter.PageSize = ps
		}
	}

	// 查询日志
	logs, total, err := ctrl.queryService.List(filter)
	if err != nil {
		response.InternalServerError(c, "查询审计日志失败")
		return
	}

	response.Success(c, gin.H{
		"logs":  logs,
		"total": total,
		"page":  filter.Page,
		"page_size": filter.PageSize,
	})
}

// GetLog 获取单条审计日志
// @Summary 获取单条审计日志
// @Tags 审计
// @Param id path int true "日志ID"
// @Success 200 {object} response.Response
// @Router /api/audit/logs/{id} [get]
func (ctrl *AuditController) GetLog(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的日志ID")
		return
	}

	log, err := ctrl.queryService.GetByID(uint(id))
	if err != nil {
		response.NotFound(c, "审计日志不存在")
		return
	}

	response.Success(c, log)
}

// GetUserLogs 获取用户的审计日志
// @Summary 获取用户的审计日志
// @Tags 审计
// @Param user_id path int true "用户ID"
// @Param limit query int false "限制数量"
// @Success 200 {object} response.Response
// @Router /api/audit/users/{user_id}/logs [get]
func (ctrl *AuditController) GetUserLogs(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("user_id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的用户ID")
		return
	}

	limit := 100
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	logs, err := ctrl.queryService.GetByUser(uint(userID), limit)
	if err != nil {
		response.InternalServerError(c, "查询用户审计日志失败")
		return
	}

	response.Success(c, logs)
}

// GetResourceLogs 获取资源的审计日志
// @Summary 获取资源的审计日志
// @Tags 审计
// @Param resource path string true "资源类型"
// @Param resource_id path int true "资源ID"
// @Param limit query int false "限制数量"
// @Success 200 {object} response.Response
// @Router /api/audit/resources/{resource}/{resource_id}/logs [get]
func (ctrl *AuditController) GetResourceLogs(c *gin.Context) {
	resource := c.Param("resource")
	resourceID, err := strconv.ParseUint(c.Param("resource_id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的资源ID")
		return
	}

	limit := 100
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	logs, err := ctrl.queryService.GetByResource(resource, uint(resourceID), limit)
	if err != nil {
		response.InternalServerError(c, "查询资源审计日志失败")
		return
	}

	response.Success(c, logs)
}

// GetStatistics 获取统计信息
// @Summary 获取审计统计信息
// @Tags 审计
// @Param start_time query string false "开始时间"
// @Param end_time query string false "结束时间"
// @Success 200 {object} response.Response
// @Router /api/audit/statistics [get]
func (ctrl *AuditController) GetStatistics(c *gin.Context) {
	// 默认统计最近7天
	startTime := time.Now().AddDate(0, 0, -7)
	endTime := time.Now()

	if startTimeStr := c.Query("start_time"); startTimeStr != "" {
		if t, err := time.Parse(time.RFC3339, startTimeStr); err == nil {
			startTime = t
		}
	}

	if endTimeStr := c.Query("end_time"); endTimeStr != "" {
		if t, err := time.Parse(time.RFC3339, endTimeStr); err == nil {
			endTime = t
		}
	}

	stats, err := ctrl.queryService.GetStatistics(startTime, endTime)
	if err != nil {
		response.InternalServerError(c, "获取统计信息失败")
		return
	}

	response.Success(c, stats)
}

// TriggerArchive 手动触发归档
// @Summary 手动触发归档
// @Tags 审计
// @Success 200 {object} response.Response
// @Router /api/audit/archive [post]
func (ctrl *AuditController) TriggerArchive(c *gin.Context) {
	count, err := ctrl.archiveService.ArchiveOldLogs()
	if err != nil {
		response.InternalServerError(c, "归档失败: "+err.Error())
		return
	}

	response.SuccessWithMsg(c, "归档成功", gin.H{
		"archived_count": count,
	})
}

// GetArchiveStatistics 获取归档统计信息
// @Summary 获取归档统计信息
// @Tags 审计
// @Success 200 {object} response.Response
// @Router /api/audit/archive/statistics [get]
func (ctrl *AuditController) GetArchiveStatistics(c *gin.Context) {
	stats, err := ctrl.archiveService.GetArchiveStatistics()
	if err != nil {
		response.InternalServerError(c, "获取归档统计失败")
		return
	}

	response.Success(c, stats)
}

// ListArchiveFiles 列出归档文件
// @Summary 列出归档文件
// @Tags 审计
// @Param start_date query string false "开始日期"
// @Param end_date query string false "结束日期"
// @Success 200 {object} response.Response
// @Router /api/audit/archive/files [get]
func (ctrl *AuditController) ListArchiveFiles(c *gin.Context) {
	// 默认列出最近30天的归档
	startDate := time.Now().AddDate(0, 0, -30)
	endDate := time.Now()

	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if t, err := time.Parse("2006-01-02", startDateStr); err == nil {
			startDate = t
		}
	}

	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if t, err := time.Parse("2006-01-02", endDateStr); err == nil {
			endDate = t
		}
	}

	files, err := ctrl.archiveService.ListArchiveFiles(startDate, endDate)
	if err != nil {
		response.InternalServerError(c, "列出归档文件失败")
		return
	}

	response.Success(c, gin.H{
		"files": files,
		"count": len(files),
	})
}

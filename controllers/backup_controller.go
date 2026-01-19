// Package controllers 备份管理控制器
package controllers

import (
	"strconv"

	"go_wails_project_manager/response"
	"go_wails_project_manager/services"

	"github.com/gin-gonic/gin"
)

// BackupStatus 备份状态响应
type BackupStatus struct {
	Running           bool  `json:"running" example:"true"`
	NextBackupTime    int64 `json:"next_backup_time" example:"1640995200"`
	LastBackupTime    int64 `json:"last_backup_time" example:"1640991600"`
	TotalBackups      int64 `json:"total_backups" example:"25"`
	SuccessfulBackups int64 `json:"successful_backups" example:"23"`
	FailedBackups     int64 `json:"failed_backups" example:"2"`
}

// BackupRecord 备份记录结构
type BackupRecord struct {
	ID          uint   `json:"id" example:"1"`
	Type        string `json:"type" example:"database"`
	Status      string `json:"status" example:"success"`
	StartTime   int64  `json:"start_time" example:"1640995200"`
	EndTime     int64  `json:"end_time" example:"1640995260"`
	FileSize    int64  `json:"file_size" example:"1048576"`
	FilePath    string `json:"file_path" example:"/backups/db_20240101_120000.sql"`
	Description string `json:"description" example:"数据库定时备份"`
	Error       string `json:"error,omitempty" example:""`
}

// BackupHistoryResponse 备份历史响应
type BackupHistoryResponse struct {
	Records []BackupRecord `json:"records"`
	Total   int64          `json:"total" example:"50"`
	Page    int            `json:"page" example:"1"`
	Limit   int            `json:"limit" example:"20"`
}

// BackupController 备份控制器
type BackupController struct{}

// NewBackupController 创建备份控制器
func NewBackupController() *BackupController {
	return &BackupController{}
}

// GetStatus 获取备份状态
// @Summary 获取备份系统状态
// @Description 获取当前备份调度器和备份服务的运行状态信息
// @Tags 备份管理
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=BackupStatus} "获取成功"
// @Failure 500 {object} response.Response "备份调度器未初始化"
// @Router /api/backup/status [get]
func (bc *BackupController) GetStatus(c *gin.Context) {
	scheduler := services.GetGlobalBackupScheduler()
	if scheduler == nil {
		response.InternalServerError(c, "备份调度器未初始化")
		return
	}

	response.Success(c, gin.H{
		"running":          scheduler.IsRunning(),
		"next_backup_time": scheduler.GetNextBackupTime().Unix(),
	})
}

// TriggerManualBackup 手动触发全量备份
// @Summary 手动触发全量备份
// @Description 立即执行一次完整的系统备份，包括数据库和CDN文件
// @Tags 备份管理
// @Accept json
// @Produce json
// @Success 200 {object} response.Response "备份任务已触发"
// @Failure 500 {object} response.Response "触发失败或调度器未初始化"
// @Router /api/backup/trigger [post]
func (bc *BackupController) TriggerManualBackup(c *gin.Context) {
	scheduler := services.GetGlobalBackupScheduler()
	if scheduler == nil {
		response.InternalServerError(c, "备份调度器未初始化")
		return
	}

	if err := scheduler.TriggerManualBackup(); err != nil {
		response.InternalServerError(c, "触发备份失败: "+err.Error())
		return
	}

	response.SuccessWithMsg(c, "备份任务已触发", nil)
}

// TriggerDatabaseBackup 手动触发数据库备份
// @Summary 手动触发数据库备份
// @Description 单独执行数据库备份操作，不包括CDN文件
// @Tags 备份管理
// @Accept json
// @Produce json
// @Success 200 {object} response.Response "数据库备份任务已触发"
// @Failure 500 {object} response.Response "触发失败或调度器未初始化"
// @Router /api/backup/database [post]
func (bc *BackupController) TriggerDatabaseBackup(c *gin.Context) {
	scheduler := services.GetGlobalBackupScheduler()
	if scheduler == nil {
		response.InternalServerError(c, "备份调度器未初始化")
		return
	}

	if err := scheduler.TriggerDatabaseBackup(); err != nil {
		response.InternalServerError(c, "触发数据库备份失败: "+err.Error())
		return
	}

	response.SuccessWithMsg(c, "数据库备份任务已触发", nil)
}

// TriggerCDNBackup 手动触发CDN备份
// @Summary 手动触发CDN增量备份
// @Description 单独执行CDN文件的增量备份操作
// @Tags 备份管理
// @Accept json
// @Produce json
// @Success 200 {object} response.Response "CDN备份任务已触发"
// @Failure 500 {object} response.Response "触发失败或调度器未初始化"
// @Router /api/backup/cdn [post]
func (bc *BackupController) TriggerCDNBackup(c *gin.Context) {
	scheduler := services.GetGlobalBackupScheduler()
	if scheduler == nil {
		response.InternalServerError(c, "备份调度器未初始化")
		return
	}

	if err := scheduler.TriggerCDNBackup(); err != nil {
		response.InternalServerError(c, "触发CDN备份失败: "+err.Error())
		return
	}

	response.SuccessWithMsg(c, "CDN备份任务已触发", nil)
}

// GetBackupHistory 获取备份历史
// @Summary 获取备份历史记录
// @Description 获取系统所有备份操作的历史记录，包括成功和失败的备份
// @Tags 备份管理
// @Accept json
// @Produce json
// @Param page query int false "页码，默认1" default(1)
// @Param limit query int false "每页数量，默认20，最大100" default(20)
// @Param type query string false "备份类型筛选：database | cdn | full"
// @Success 200 {object} response.Response{data=BackupHistoryResponse} "获取成功"
// @Failure 500 {object} response.Response "获取失败或调度器未初始化"
// @Router /api/backup/history [get]
func (bc *BackupController) GetBackupHistory(c *gin.Context) {
	scheduler := services.GetGlobalBackupScheduler()
	if scheduler == nil {
		response.InternalServerError(c, "备份调度器未初始化")
		return
	}

	page := 1
	limit := 20
	if p := c.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	backupService := scheduler.GetBackupService()
	records, total, err := backupService.GetBackupHistory(page, limit)
	if err != nil {
		response.InternalServerError(c, "获取备份历史失败: "+err.Error())
		return
	}

	response.Success(c, gin.H{
		"records": records,
		"total":   total,
		"page":    page,
		"limit":   limit,
	})
}

// RestoreCDNFromBackup CDN文件恢复
// @Summary 从备份恢复CDN文件
// @Description 使用指定的备份记录恢复CDN文件到指定状态（谨慎操作）
// @Tags 备份管理
// @Accept json
// @Produce json
// @Param backup_id path int true "备份记录ID（示例: 1）"
// @Success 200 {object} response.Response "CDN文件恢复成功"
// @Failure 400 {object} response.Response "无效的备份 ID"
// @Failure 404 {object} response.Response "备份记录不存在"
// @Failure 500 {object} response.Response "恢复失败或调度器未初始化"
// @Router /api/backup/restore/cdn/{backup_id} [post]
func (bc *BackupController) RestoreCDNFromBackup(c *gin.Context) {
	backupIDStr := c.Param("backup_id")
	backupID, err := strconv.ParseUint(backupIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的备份ID")
		return
	}

	scheduler := services.GetGlobalBackupScheduler()
	if scheduler == nil {
		response.InternalServerError(c, "备份调度器未初始化")
		return
	}

	cdnBackupService := scheduler.GetCDNBackupService()
	if err := cdnBackupService.RestoreCDNFromBackup(uint(backupID)); err != nil {
		response.InternalServerError(c, "恢复CDN文件失败: "+err.Error())
		return
	}

	response.SuccessWithMsg(c, "CDN文件恢复成功", nil)
}

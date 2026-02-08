package controllers

import (
	"go_wails_project_manager/response"
	"go_wails_project_manager/services"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type StatisticsController struct {
	service *services.StatisticsService
}

func NewStatisticsController(db *gorm.DB) *StatisticsController {
	return &StatisticsController{
		service: services.NewStatisticsService(db),
	}
}

// GetOverview 获取资源统计概览
// @Summary 获取资源统计概览
// @Description 获取贴图、项目、模型、资产的统计信息
// @Tags 统计
// @Accept json
// @Produce json
// @Success 200 {object} response.Response
// @Router /api/statistics/overview [get]
func (sc *StatisticsController) GetOverview(c *gin.Context) {
	stats, err := sc.service.GetOverview()
	if err != nil {
		response.Error(c, 500, "获取统计信息失败: "+err.Error())
		return
	}

	response.Success(c, stats)
}

// GetRecentActivities 获取最近活动
// @Summary 获取最近活动
// @Description 获取最近的资源操作记录
// @Tags 统计
// @Accept json
// @Produce json
// @Param limit query int false "返回数量" default(10)
// @Success 200 {object} response.Response
// @Router /api/statistics/recent-activities [get]
func (sc *StatisticsController) GetRecentActivities(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100 // 最多返回100条
	}

	activities, err := sc.service.GetRecentActivities(limit)
	if err != nil {
		response.Error(c, 500, "获取活动记录失败: "+err.Error())
		return
	}

	response.Success(c, gin.H{
		"activities": activities,
	})
}

// GetSystemStatus 获取系统状态
// @Summary 获取系统状态
// @Description 获取服务、数据库、存储、同步等状态信息
// @Tags 统计
// @Accept json
// @Produce json
// @Success 200 {object} response.Response
// @Router /api/statistics/system-status [get]
func (sc *StatisticsController) GetSystemStatus(c *gin.Context) {
	status, err := sc.service.GetSystemStatus()
	if err != nil {
		response.Error(c, 500, "获取系统状态失败: "+err.Error())
		return
	}

	response.Success(c, status)
}

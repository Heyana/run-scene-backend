package controllers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"go_wails_project_manager/models"
	"go_wails_project_manager/response"
	"go_wails_project_manager/services/blueprint"
)

// BlueprintController 蓝图控制器
type BlueprintController struct {
	service *blueprint.BlueprintService
}

// NewBlueprintController 创建蓝图控制器
func NewBlueprintController(db *gorm.DB) *BlueprintController {
	return &BlueprintController{
		service: blueprint.NewBlueprintService(db),
	}
}

// Generate 生成蓝图
// @Summary 生成蓝图
// @Description 根据用户需求和节点元数据生成蓝图
// @Tags 蓝图
// @Accept json
// @Produce json
// @Param request body models.GenerateRequest true "生成请求"
// @Success 200 {object} response.Response{data=models.GenerateResponse}
// @Router /api/blueprint/generate [post]
func (c *BlueprintController) Generate(ctx *gin.Context) {
	var req models.GenerateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	// 创建超时上下文
	genCtx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// 获取用户IP
	userIP := ctx.ClientIP()

	// 生成蓝图
	result, err := c.service.Generate(genCtx, &req, userIP)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(ctx, result)
}

// GetHistory 获取生成历史
// @Summary 获取生成历史
// @Description 分页获取蓝图生成历史记录
// @Tags 蓝图
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(20)
// @Success 200 {object} response.Response
// @Router /api/blueprint/history [get]
func (c *BlueprintController) GetHistory(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("pageSize", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	histories, total, err := c.service.GetHistory(page, pageSize)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, "查询失败")
		return
	}

	response.Success(ctx, gin.H{
		"list":     histories,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

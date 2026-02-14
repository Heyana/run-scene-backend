package controllers

import (
	"go_wails_project_manager/database"
	"go_wails_project_manager/models"
	"go_wails_project_manager/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

// PermissionController 权限管理控制器
type PermissionController struct{}

// NewPermissionController 创建权限控制器
func NewPermissionController() *PermissionController {
	return &PermissionController{}
}

// CreatePermissionRequest 创建权限请求
type CreatePermissionRequest struct {
	Code        string `json:"code" binding:"required,min=2,max=100"`
	Name        string `json:"name" binding:"required,min=2,max=100"`
	Resource    string `json:"resource" binding:"required"`
	Action      string `json:"action" binding:"required"`
	Description string `json:"description"`
}

// UpdatePermissionRequest 更新权限请求
type UpdatePermissionRequest struct {
	Name        string `json:"name" binding:"omitempty,min=2,max=100"`
	Description string `json:"description"`
}

// CreatePermissionGroupRequest 创建权限组请求
type CreatePermissionGroupRequest struct {
	Code          string `json:"code" binding:"required,min=2,max=50"`
	Name          string `json:"name" binding:"required,min=2,max=100"`
	Description   string `json:"description"`
	PermissionIDs []uint `json:"permission_ids"`
}

// UpdatePermissionGroupRequest 更新权限组请求
type UpdatePermissionGroupRequest struct {
	Name        string `json:"name" binding:"omitempty,min=2,max=100"`
	Description string `json:"description"`
}

// List 获取权限列表
func (pc *PermissionController) List(c *gin.Context) {
	db, err := database.GetDB()
	if err != nil {
		response.InternalServerError(c, "数据库连接失败")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	keyword := c.Query("keyword")
	resource := c.Query("resource")
	action := c.Query("action")

	query := db.Model(&models.Permission{})

	// 关键词搜索
	if keyword != "" {
		query = query.Where("code LIKE ? OR name LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	// 资源筛选
	if resource != "" {
		query = query.Where("resource = ?", resource)
	}

	// 操作筛选
	if action != "" {
		query = query.Where("action = ?", action)
	}

	// 分页
	var total int64
	query.Count(&total)

	var permissions []models.Permission
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Find(&permissions).Error; err != nil {
		response.InternalServerError(c, "查询失败")
		return
	}

	response.Success(c, gin.H{
		"items":     permissions,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// Create 创建权限
func (pc *PermissionController) Create(c *gin.Context) {
	var req CreatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	db, err := database.GetDB()
	if err != nil {
		response.InternalServerError(c, "数据库连接失败")
		return
	}

	// 检查代码是否存在
	var existingPerm models.Permission
	if err := db.Where("code = ?", req.Code).First(&existingPerm).Error; err == nil {
		response.Error(c, 409, "权限代码已存在")
		return
	}

	// 创建权限
	permission := models.Permission{
		Code:        req.Code,
		Name:        req.Name,
		Resource:    req.Resource,
		Action:      req.Action,
		Description: req.Description,
		IsSystem:    false,
	}

	if err := db.Create(&permission).Error; err != nil {
		response.InternalServerError(c, "创建权限失败")
		return
	}

	response.Success(c, permission)
}

// GetDetail 获取权限详情
func (pc *PermissionController) GetDetail(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	db, err := database.GetDB()
	if err != nil {
		response.InternalServerError(c, "数据库连接失败")
		return
	}

	var permission models.Permission
	if err := db.First(&permission, id).Error; err != nil {
		response.NotFound(c, "权限不存在")
		return
	}

	response.Success(c, permission)
}

// Update 更新权限
func (pc *PermissionController) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	var req UpdatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	db, err := database.GetDB()
	if err != nil {
		response.InternalServerError(c, "数据库连接失败")
		return
	}

	var permission models.Permission
	if err := db.First(&permission, id).Error; err != nil {
		response.NotFound(c, "权限不存在")
		return
	}

	// 系统权限不能修改
	if permission.IsSystem {
		response.BadRequest(c, "系统权限不能修改")
		return
	}

	// 更新字段
	if req.Name != "" {
		permission.Name = req.Name
	}
	permission.Description = req.Description

	if err := db.Save(&permission).Error; err != nil {
		response.InternalServerError(c, "更新失败")
		return
	}

	response.Success(c, permission)
}

// Delete 删除权限
func (pc *PermissionController) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	db, err := database.GetDB()
	if err != nil {
		response.InternalServerError(c, "数据库连接失败")
		return
	}

	var permission models.Permission
	if err := db.First(&permission, id).Error; err != nil {
		response.NotFound(c, "权限不存在")
		return
	}

	// 系统权限不能删除
	if permission.IsSystem {
		response.BadRequest(c, "系统权限不能删除")
		return
	}

	if err := db.Delete(&permission).Error; err != nil {
		response.InternalServerError(c, "删除失败")
		return
	}

	response.SuccessWithMsg(c, "删除成功", nil)
}

// GetResources 获取所有资源类型
func (pc *PermissionController) GetResources(c *gin.Context) {
	db, err := database.GetDB()
	if err != nil {
		response.InternalServerError(c, "数据库连接失败")
		return
	}

	var resources []string
	if err := db.Model(&models.Permission{}).Distinct("resource").Pluck("resource", &resources).Error; err != nil {
		response.InternalServerError(c, "查询失败")
		return
	}

	response.Success(c, resources)
}

// GetActions 获取所有操作类型
func (pc *PermissionController) GetActions(c *gin.Context) {
	db, err := database.GetDB()
	if err != nil {
		response.InternalServerError(c, "数据库连接失败")
		return
	}

	var actions []string
	if err := db.Model(&models.Permission{}).Distinct("action").Pluck("action", &actions).Error; err != nil {
		response.InternalServerError(c, "查询失败")
		return
	}

	response.Success(c, actions)
}

// ==================== 权限组管理 ====================

// ListGroups 获取权限组列表
func (pc *PermissionController) ListGroups(c *gin.Context) {
	db, err := database.GetDB()
	if err != nil {
		response.InternalServerError(c, "数据库连接失败")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	keyword := c.Query("keyword")

	query := db.Model(&models.PermissionGroup{})

	// 关键词搜索
	if keyword != "" {
		query = query.Where("code LIKE ? OR name LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	// 分页
	var total int64
	query.Count(&total)

	var groups []models.PermissionGroup
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Find(&groups).Error; err != nil {
		response.InternalServerError(c, "查询失败")
		return
	}

	response.Success(c, gin.H{
		"items":     groups,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// CreateGroup 创建权限组
func (pc *PermissionController) CreateGroup(c *gin.Context) {
	var req CreatePermissionGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	db, err := database.GetDB()
	if err != nil {
		response.InternalServerError(c, "数据库连接失败")
		return
	}

	// 检查代码是否存在
	var existingGroup models.PermissionGroup
	if err := db.Where("code = ?", req.Code).First(&existingGroup).Error; err == nil {
		response.Error(c, 409, "权限组代码已存在")
		return
	}

	// 创建权限组
	group := models.PermissionGroup{
		Code:        req.Code,
		Name:        req.Name,
		Description: req.Description,
		IsSystem:    false,
	}

	if err := db.Create(&group).Error; err != nil {
		response.InternalServerError(c, "创建权限组失败")
		return
	}

	// 添加权限
	if len(req.PermissionIDs) > 0 {
		var permissions []models.Permission
		db.Where("id IN ?", req.PermissionIDs).Find(&permissions)
		if len(permissions) > 0 {
			db.Model(&group).Association("Permissions").Append(permissions)
		}
	}

	response.Success(c, group)
}

// GetGroupDetail 获取权限组详情
func (pc *PermissionController) GetGroupDetail(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	db, err := database.GetDB()
	if err != nil {
		response.InternalServerError(c, "数据库连接失败")
		return
	}

	var group models.PermissionGroup
	if err := db.Preload("Permissions").First(&group, id).Error; err != nil {
		response.NotFound(c, "权限组不存在")
		return
	}

	response.Success(c, group)
}

// UpdateGroup 更新权限组
func (pc *PermissionController) UpdateGroup(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	var req UpdatePermissionGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	db, err := database.GetDB()
	if err != nil {
		response.InternalServerError(c, "数据库连接失败")
		return
	}

	var group models.PermissionGroup
	if err := db.First(&group, id).Error; err != nil {
		response.NotFound(c, "权限组不存在")
		return
	}

	// 系统权限组不能修改
	if group.IsSystem {
		response.BadRequest(c, "系统权限组不能修改")
		return
	}

	// 更新字段
	if req.Name != "" {
		group.Name = req.Name
	}
	group.Description = req.Description

	if err := db.Save(&group).Error; err != nil {
		response.InternalServerError(c, "更新失败")
		return
	}

	response.Success(c, group)
}

// DeleteGroup 删除权限组
func (pc *PermissionController) DeleteGroup(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	db, err := database.GetDB()
	if err != nil {
		response.InternalServerError(c, "数据库连接失败")
		return
	}

	var group models.PermissionGroup
	if err := db.First(&group, id).Error; err != nil {
		response.NotFound(c, "权限组不存在")
		return
	}

	// 系统权限组不能删除
	if group.IsSystem {
		response.BadRequest(c, "系统权限组不能删除")
		return
	}

	if err := db.Delete(&group).Error; err != nil {
		response.InternalServerError(c, "删除失败")
		return
	}

	response.SuccessWithMsg(c, "删除成功", nil)
}

// AddPermissionsToGroup 添加权限到权限组
func (pc *PermissionController) AddPermissionsToGroup(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	type AddRequest struct {
		PermissionIDs []uint `json:"permission_ids" binding:"required"`
	}

	var req AddRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	db, err := database.GetDB()
	if err != nil {
		response.InternalServerError(c, "数据库连接失败")
		return
	}

	var group models.PermissionGroup
	if err := db.First(&group, id).Error; err != nil {
		response.NotFound(c, "权限组不存在")
		return
	}

	// 查找权限
	var permissions []models.Permission
	if err := db.Where("id IN ?", req.PermissionIDs).Find(&permissions).Error; err != nil {
		response.InternalServerError(c, "查询权限失败")
		return
	}

	// 添加权限
	if err := db.Model(&group).Association("Permissions").Append(permissions); err != nil {
		response.InternalServerError(c, "添加权限失败")
		return
	}

	response.SuccessWithMsg(c, "权限添加成功", nil)
}

// RemovePermissionFromGroup 从权限组移除权限
func (pc *PermissionController) RemovePermissionFromGroup(c *gin.Context) {
	groupID, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	permissionID, _ := strconv.ParseUint(c.Param("permission_id"), 10, 32)

	db, err := database.GetDB()
	if err != nil {
		response.InternalServerError(c, "数据库连接失败")
		return
	}

	var group models.PermissionGroup
	if err := db.First(&group, groupID).Error; err != nil {
		response.NotFound(c, "权限组不存在")
		return
	}

	var permission models.Permission
	if err := db.First(&permission, permissionID).Error; err != nil {
		response.NotFound(c, "权限不存在")
		return
	}

	// 移除权限
	if err := db.Model(&group).Association("Permissions").Delete(&permission); err != nil {
		response.InternalServerError(c, "移除权限失败")
		return
	}

	response.SuccessWithMsg(c, "权限移除成功", nil)
}

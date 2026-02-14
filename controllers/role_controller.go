package controllers

import (
	"go_wails_project_manager/database"
	"go_wails_project_manager/middleware"
	"go_wails_project_manager/models"
	"go_wails_project_manager/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

// RoleController 角色管理控制器
type RoleController struct{}

// NewRoleController 创建角色控制器
func NewRoleController() *RoleController {
	return &RoleController{}
}

// CreateRoleRequest 创建角色请求
type CreateRoleRequest struct {
	Code                string `json:"code" binding:"required,min=2,max=50"`
	Name                string `json:"name" binding:"required,min=2,max=100"`
	Description         string `json:"description"`
	PermissionIDs       []uint `json:"permission_ids"`
	PermissionGroupIDs  []uint `json:"permission_group_ids"`
}

// UpdateRoleRequest 更新角色请求
type UpdateRoleRequest struct {
	Name        string `json:"name" binding:"omitempty,min=2,max=100"`
	Description string `json:"description"`
}

// AssignPermissionsRequest 分配权限请求
type AssignPermissionsRequest struct {
	PermissionIDs      []uint `json:"permission_ids"`
	PermissionGroupIDs []uint `json:"permission_group_ids"`
}

// List 获取角色列表
func (rc *RoleController) List(c *gin.Context) {
	db, err := database.GetDB()
	if err != nil {
		response.InternalServerError(c, "数据库连接失败")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	keyword := c.Query("keyword")

	query := db.Model(&models.Role{})

	// 关键词搜索
	if keyword != "" {
		query = query.Where("code LIKE ? OR name LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	// 分页
	var total int64
	query.Count(&total)

	var roles []models.Role
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Find(&roles).Error; err != nil {
		response.InternalServerError(c, "查询失败")
		return
	}

	response.Success(c, gin.H{
		"items":     roles,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// Create 创建角色
func (rc *RoleController) Create(c *gin.Context) {
	var req CreateRoleRequest
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
	var existingRole models.Role
	if err := db.Where("code = ?", req.Code).First(&existingRole).Error; err == nil {
		response.Error(c, 409, "角色代码已存在")
		return
	}

	// 创建角色
	role := models.Role{
		Code:        req.Code,
		Name:        req.Name,
		Description: req.Description,
		IsSystem:    false,
	}

	if err := db.Create(&role).Error; err != nil {
		response.InternalServerError(c, "创建角色失败")
		return
	}

	// 分配权限
	if len(req.PermissionIDs) > 0 {
		var permissions []models.Permission
		db.Where("id IN ?", req.PermissionIDs).Find(&permissions)
		if len(permissions) > 0 {
			db.Model(&role).Association("Permissions").Append(permissions)
		}
	}

	// 分配权限组
	if len(req.PermissionGroupIDs) > 0 {
		var groups []models.PermissionGroup
		db.Where("id IN ?", req.PermissionGroupIDs).Find(&groups)
		if len(groups) > 0 {
			db.Model(&role).Association("PermissionGroups").Append(groups)
		}
	}

	response.Success(c, role)
}

// GetDetail 获取角色详情
func (rc *RoleController) GetDetail(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	db, err := database.GetDB()
	if err != nil {
		response.InternalServerError(c, "数据库连接失败")
		return
	}

	var role models.Role
	if err := db.Preload("Permissions").Preload("PermissionGroups").
		First(&role, id).Error; err != nil {
		response.NotFound(c, "角色不存在")
		return
	}

	response.Success(c, role)
}

// Update 更新角色
func (rc *RoleController) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	var req UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	db, err := database.GetDB()
	if err != nil {
		response.InternalServerError(c, "数据库连接失败")
		return
	}

	var role models.Role
	if err := db.First(&role, id).Error; err != nil {
		response.NotFound(c, "角色不存在")
		return
	}

	// 系统角色不能修改
	if role.IsSystem {
		response.BadRequest(c, "系统角色不能修改")
		return
	}

	// 更新字段
	if req.Name != "" {
		role.Name = req.Name
	}
	role.Description = req.Description

	if err := db.Save(&role).Error; err != nil {
		response.InternalServerError(c, "更新失败")
		return
	}

	response.Success(c, role)
}

// Delete 删除角色
func (rc *RoleController) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	db, err := database.GetDB()
	if err != nil {
		response.InternalServerError(c, "数据库连接失败")
		return
	}

	var role models.Role
	if err := db.First(&role, id).Error; err != nil {
		response.NotFound(c, "角色不存在")
		return
	}

	// 系统角色不能删除
	if role.IsSystem {
		response.BadRequest(c, "系统角色不能删除")
		return
	}

	// 检查是否有用户使用该角色
	var count int64
	db.Model(&models.User{}).Joins("JOIN user_roles ON users.id = user_roles.user_id").
		Where("user_roles.role_id = ?", id).Count(&count)
	if count > 0 {
		response.BadRequest(c, "该角色正在被使用，无法删除")
		return
	}

	if err := db.Delete(&role).Error; err != nil {
		response.InternalServerError(c, "删除失败")
		return
	}

	response.SuccessWithMsg(c, "删除成功", nil)
}

// AssignPermissions 分配权限
func (rc *RoleController) AssignPermissions(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	var req AssignPermissionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	db, err := database.GetDB()
	if err != nil {
		response.InternalServerError(c, "数据库连接失败")
		return
	}

	var role models.Role
	if err := db.First(&role, id).Error; err != nil {
		response.NotFound(c, "角色不存在")
		return
	}

	// 查找权限
	var permissions []models.Permission
	if len(req.PermissionIDs) > 0 {
		if err := db.Where("id IN ?", req.PermissionIDs).Find(&permissions).Error; err != nil {
			response.InternalServerError(c, "查询权限失败")
			return
		}
	}

	// 查找权限组
	var groups []models.PermissionGroup
	if len(req.PermissionGroupIDs) > 0 {
		if err := db.Where("id IN ?", req.PermissionGroupIDs).Find(&groups).Error; err != nil {
			response.InternalServerError(c, "查询权限组失败")
			return
		}
	}

	// 替换权限
	if err := db.Model(&role).Association("Permissions").Replace(permissions); err != nil {
		response.InternalServerError(c, "分配权限失败")
		return
	}

	// 替换权限组
	if err := db.Model(&role).Association("PermissionGroups").Replace(groups); err != nil {
		response.InternalServerError(c, "分配权限组失败")
		return
	}

	// 清除所有使用该角色的用户的权限缓存
	var users []models.User
	db.Joins("JOIN user_roles ON users.id = user_roles.user_id").
		Where("user_roles.role_id = ?", id).Find(&users)
	for _, user := range users {
		middleware.InvalidatePermissionCache(user.ID)
	}

	response.SuccessWithMsg(c, "权限分配成功", nil)
}

// GetPermissions 获取角色的权限
func (rc *RoleController) GetPermissions(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	db, err := database.GetDB()
	if err != nil {
		response.InternalServerError(c, "数据库连接失败")
		return
	}

	var role models.Role
	if err := db.Preload("Permissions").Preload("PermissionGroups").
		First(&role, id).Error; err != nil {
		response.NotFound(c, "角色不存在")
		return
	}

	// 提取ID列表
	permissionIDs := make([]uint, len(role.Permissions))
	for i, perm := range role.Permissions {
		permissionIDs[i] = perm.ID
	}

	groupIDs := make([]uint, len(role.PermissionGroups))
	for i, group := range role.PermissionGroups {
		groupIDs[i] = group.ID
	}

	response.Success(c, gin.H{
		"permission_ids":       permissionIDs,
		"permission_group_ids": groupIDs,
	})
}

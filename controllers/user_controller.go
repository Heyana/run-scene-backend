package controllers

import (
	"go_wails_project_manager/database"
	"go_wails_project_manager/middleware"
	"go_wails_project_manager/models"
	"go_wails_project_manager/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

// UserController 用户管理控制器
type UserController struct{}

// NewUserController 创建用户控制器
func NewUserController() *UserController {
	return &UserController{}
}

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=6"`
	Email    string `json:"email" binding:"required,email"`
	Phone    string `json:"phone"`
	RealName string `json:"real_name"`
	RoleIDs  []uint `json:"role_ids"`
}

// UpdateUserRequest 更新用户请求
type UpdateUserRequest struct {
	Email    string `json:"email" binding:"omitempty,email"`
	Phone    string `json:"phone"`
	RealName string `json:"real_name"`
	Avatar   string `json:"avatar"`
}

// List 获取用户列表
func (uc *UserController) List(c *gin.Context) {
	db, err := database.GetDB()
	if err != nil {
		response.InternalServerError(c, "数据库连接失败")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	status := c.Query("status")
	keyword := c.Query("keyword")

	query := db.Model(&models.User{}).Preload("Roles")

	// 状态筛选
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// 关键词搜索
	if keyword != "" {
		query = query.Where("username LIKE ? OR email LIKE ? OR real_name LIKE ?",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}

	// 分页
	var total int64
	query.Count(&total)

	var users []models.User
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Find(&users).Error; err != nil {
		response.InternalServerError(c, "查询失败")
		return
	}

	// 转换为响应格式
	userResponses := make([]*models.UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = user.ToResponse()
	}

	response.Success(c, gin.H{
		"items":     userResponses,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// Create 创建用户
func (uc *UserController) Create(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	db, err := database.GetDB()
	if err != nil {
		response.InternalServerError(c, "数据库连接失败")
		return
	}

	// 检查用户名是否存在
	var existingUser models.User
	if err := db.Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
		response.Error(c, 409, "用户名已存在")
		return
	}

	// 检查邮箱是否存在
	if err := db.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		response.Error(c, 409, "邮箱已被使用")
		return
	}

	// 加密密码
	hashedPassword, err := middleware.HashPassword(req.Password)
	if err != nil {
		response.InternalServerError(c, "密码加密失败")
		return
	}

	// 创建用户
	user := models.User{
		Username: req.Username,
		Password: hashedPassword,
		Email:    req.Email,
		Phone:    req.Phone,
		RealName: req.RealName,
		Status:   models.UserStatusActive,
	}

	if err := db.Create(&user).Error; err != nil {
		response.InternalServerError(c, "创建用户失败")
		return
	}

	// 分配角色
	if len(req.RoleIDs) > 0 {
		var roles []models.Role
		db.Where("id IN ?", req.RoleIDs).Find(&roles)
		if len(roles) > 0 {
			db.Model(&user).Association("Roles").Append(roles)
		}
	}

	response.Success(c, user.ToResponse())
}

// GetDetail 获取用户详情
func (uc *UserController) GetDetail(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	
	db, err := database.GetDB()
	if err != nil {
		response.InternalServerError(c, "数据库连接失败")
		return
	}

	var user models.User
	if err := db.Preload("Roles").Preload("Permissions").Preload("PermissionGroups").
		First(&user, id).Error; err != nil {
		response.NotFound(c, "用户不存在")
		return
	}

	// 获取用户所有权限
	perms, _ := middleware.GetUserPermissions(uint(id))

	response.Success(c, gin.H{
		"user":              user.ToResponse(),
		"roles":             user.Roles,
		"permissions":       user.Permissions,
		"permission_groups": user.PermissionGroups,
		"all_permissions":   perms,
	})
}

// Update 更新用户
func (uc *UserController) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	db, err := database.GetDB()
	if err != nil {
		response.InternalServerError(c, "数据库连接失败")
		return
	}

	var user models.User
	if err := db.First(&user, id).Error; err != nil {
		response.NotFound(c, "用户不存在")
		return
	}

	// 更新字段
	if req.Email != "" {
		user.Email = req.Email
	}
	if req.Phone != "" {
		user.Phone = req.Phone
	}
	if req.RealName != "" {
		user.RealName = req.RealName
	}
	if req.Avatar != "" {
		user.Avatar = req.Avatar
	}

	if err := db.Save(&user).Error; err != nil {
		response.InternalServerError(c, "更新失败")
		return
	}

	response.Success(c, user.ToResponse())
}

// Delete 删除用户
func (uc *UserController) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	db, err := database.GetDB()
	if err != nil {
		response.InternalServerError(c, "数据库连接失败")
		return
	}

	// 不能删除自己
	currentUserID := middleware.GetUserID(c)
	if currentUserID == uint(id) {
		response.BadRequest(c, "不能删除自己")
		return
	}

	// 检查是否是超级管理员
	var user models.User
	if err := db.Preload("Roles").First(&user, id).Error; err != nil {
		response.NotFound(c, "用户不存在")
		return
	}

	for _, role := range user.Roles {
		if role.Code == models.RoleSuperAdmin {
			response.BadRequest(c, "不能删除超级管理员")
			return
		}
	}

	if err := db.Delete(&user).Error; err != nil {
		response.InternalServerError(c, "删除失败")
		return
	}

	// 清除权限缓存
	middleware.InvalidatePermissionCache(uint(id))

	response.SuccessWithMsg(c, "删除成功", nil)
}

// Disable 禁用用户
func (uc *UserController) Disable(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	db, err := database.GetDB()
	if err != nil {
		response.InternalServerError(c, "数据库连接失败")
		return
	}

	var user models.User
	if err := db.First(&user, id).Error; err != nil {
		response.NotFound(c, "用户不存在")
		return
	}

	user.Status = models.UserStatusDisabled
	if err := db.Save(&user).Error; err != nil {
		response.InternalServerError(c, "操作失败")
		return
	}

	response.SuccessWithMsg(c, "用户已禁用", nil)
}

// Enable 启用用户
func (uc *UserController) Enable(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	db, err := database.GetDB()
	if err != nil {
		response.InternalServerError(c, "数据库连接失败")
		return
	}

	var user models.User
	if err := db.First(&user, id).Error; err != nil {
		response.NotFound(c, "用户不存在")
		return
	}

	user.Status = models.UserStatusActive
	user.LockedUntil = nil
	user.LoginFailCount = 0
	if err := db.Save(&user).Error; err != nil {
		response.InternalServerError(c, "操作失败")
		return
	}

	response.SuccessWithMsg(c, "用户已启用", nil)
}

// ResetPassword 重置密码
func (uc *UserController) ResetPassword(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	type ResetRequest struct {
		NewPassword string `json:"new_password" binding:"required,min=6"`
	}

	var req ResetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	db, err := database.GetDB()
	if err != nil {
		response.InternalServerError(c, "数据库连接失败")
		return
	}

	var user models.User
	if err := db.First(&user, id).Error; err != nil {
		response.NotFound(c, "用户不存在")
		return
	}

	// 加密新密码
	hashedPassword, err := middleware.HashPassword(req.NewPassword)
	if err != nil {
		response.InternalServerError(c, "密码加密失败")
		return
	}

	user.Password = hashedPassword
	if err := db.Save(&user).Error; err != nil {
		response.InternalServerError(c, "重置密码失败")
		return
	}

	response.SuccessWithMsg(c, "密码重置成功", nil)
}

// AssignRoles 分配角色
func (uc *UserController) AssignRoles(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	type AssignRequest struct {
		RoleIDs []uint `json:"role_ids" binding:"required"`
	}

	var req AssignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	db, err := database.GetDB()
	if err != nil {
		response.InternalServerError(c, "数据库连接失败")
		return
	}

	var user models.User
	if err := db.First(&user, id).Error; err != nil {
		response.NotFound(c, "用户不存在")
		return
	}

	// 查找角色
	var roles []models.Role
	if err := db.Where("id IN ?", req.RoleIDs).Find(&roles).Error; err != nil {
		response.InternalServerError(c, "查询角色失败")
		return
	}

	// 替换角色
	if err := db.Model(&user).Association("Roles").Replace(roles); err != nil {
		response.InternalServerError(c, "分配角色失败")
		return
	}

	// 清除权限缓存
	middleware.InvalidatePermissionCache(uint(id))

	response.SuccessWithMsg(c, "角色分配成功", nil)
}

// GetPermissions 获取用户所有权限
func (uc *UserController) GetPermissions(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	perms, err := middleware.GetUserPermissions(uint(id))
	if err != nil {
		response.InternalServerError(c, "获取权限失败")
		return
	}

	response.Success(c, gin.H{
		"permissions": perms,
	})
}

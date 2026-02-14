package controllers

import (
	"go_wails_project_manager/database"
	"go_wails_project_manager/middleware"
	"go_wails_project_manager/models"
	"go_wails_project_manager/response"
	"time"

	"github.com/gin-gonic/gin"
)

// AuthController 认证控制器
type AuthController struct {
	jwtAuth *middleware.JWTAuth
}

// NewAuthController 创建认证控制器
func NewAuthController(jwtAuth *middleware.JWTAuth) *AuthController {
	return &AuthController{
		jwtAuth: jwtAuth,
	}
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=6"`
	Email    string `json:"email" binding:"required,email"`
	Phone    string `json:"phone"`
	RealName string `json:"real_name"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

// Register 注册
func (ac *AuthController) Register(c *gin.Context) {
	var req RegisterRequest
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

	// 分配默认角色（查看者）
	var viewerRole models.Role
	if err := db.Where("code = ?", models.RoleViewer).First(&viewerRole).Error; err == nil {
		db.Model(&user).Association("Roles").Append(&viewerRole)
	}

	// 生成 Token
	token, err := ac.jwtAuth.GenerateToken(user.ID, user.Username, models.RoleViewer)
	if err != nil {
		response.InternalServerError(c, "生成Token失败")
		return
	}

	response.Success(c, gin.H{
		"user":  user.ToResponse(),
		"token": token,
	})
}

// Login 登录
func (ac *AuthController) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	db, err := database.GetDB()
	if err != nil {
		response.InternalServerError(c, "数据库连接失败")
		return
	}

	// 查找用户
	var user models.User
	if err := db.Preload("Roles").Where("username = ?", req.Username).First(&user).Error; err != nil {
		response.Unauthorized(c, "用户名或密码错误")
		return
	}

	// 检查用户状态
	if user.Status == models.UserStatusDisabled {
		response.Forbidden(c, "账号已被禁用")
		return
	}

	if user.IsLocked() {
		response.Forbidden(c, "账号已被锁定，请稍后再试")
		return
	}

	// 验证密码
	if !middleware.CheckPassword(req.Password, user.Password) {
		// 记录登录失败
		user.LoginFailCount++
		if user.LoginFailCount >= 5 {
			// 锁定30分钟
			lockUntil := time.Now().Add(30 * time.Minute)
			user.LockedUntil = &lockUntil
			user.Status = models.UserStatusLocked
		}
		db.Save(&user)

		response.Unauthorized(c, "用户名或密码错误")
		return
	}

	// 登录成功，重置失败次数
	now := time.Now()
	user.LoginFailCount = 0
	user.LockedUntil = nil
	user.LastLoginAt = &now
	user.LastLoginIP = c.ClientIP()
	if user.Status == models.UserStatusLocked {
		user.Status = models.UserStatusActive
	}
	db.Save(&user)

	// 获取用户主要角色
	roleCode := models.RoleViewer
	if len(user.Roles) > 0 {
		roleCode = user.Roles[0].Code
	}

	// 生成 Token
	token, err := ac.jwtAuth.GenerateToken(user.ID, user.Username, roleCode)
	if err != nil {
		response.InternalServerError(c, "生成Token失败")
		return
	}

	// 生成刷新 Token
	refreshToken, err := ac.jwtAuth.GenerateRefreshToken(user.ID)
	if err != nil {
		response.InternalServerError(c, "生成刷新Token失败")
		return
	}

	response.Success(c, gin.H{
		"access_token":  token,
		"refresh_token": refreshToken,
		"expires_in":    24 * 3600, // 24小时
		"token_type":    "Bearer",
		"user":          user.ToResponse(),
	})
}

// Logout 登出
func (ac *AuthController) Logout(c *gin.Context) {
	userID := middleware.GetUserID(c)
	
	// 清除权限缓存
	middleware.InvalidatePermissionCache(userID)
	
	// 清除cookie
	c.SetCookie("token", "", -1, "/", "", false, true)
	
	response.SuccessWithMsg(c, "登出成功", nil)
}

// RefreshToken 刷新Token
func (ac *AuthController) RefreshToken(c *gin.Context) {
	tokenString := ac.jwtAuth.ExtractToken(c)
	if tokenString == "" {
		response.Unauthorized(c, "缺少token")
		return
	}

	newToken, err := ac.jwtAuth.RefreshToken(tokenString)
	if err != nil {
		response.Unauthorized(c, "刷新token失败")
		return
	}

	response.Success(c, gin.H{
		"access_token": newToken,
		"expires_in":   24 * 3600,
		"token_type":   "Bearer",
	})
}

// ChangePassword 修改密码
func (ac *AuthController) ChangePassword(c *gin.Context) {
	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	userID := middleware.GetUserID(c)
	db, err := database.GetDB()
	if err != nil {
		response.InternalServerError(c, "数据库连接失败")
		return
	}

	var user models.User
	if err := db.First(&user, userID).Error; err != nil {
		response.NotFound(c, "用户不存在")
		return
	}

	// 验证旧密码
	if !middleware.CheckPassword(req.OldPassword, user.Password) {
		response.BadRequest(c, "原密码错误")
		return
	}

	// 加密新密码
	hashedPassword, err := middleware.HashPassword(req.NewPassword)
	if err != nil {
		response.InternalServerError(c, "密码加密失败")
		return
	}

	// 更新密码
	user.Password = hashedPassword
	if err := db.Save(&user).Error; err != nil {
		response.InternalServerError(c, "修改密码失败")
		return
	}

	response.SuccessWithMsg(c, "密码修改成功", nil)
}

// GetProfile 获取当前用户信息
func (ac *AuthController) GetProfile(c *gin.Context) {
	userID := middleware.GetUserID(c)
	db, err := database.GetDB()
	if err != nil {
		response.InternalServerError(c, "数据库连接失败")
		return
	}

	var user models.User
	if err := db.Preload("Roles").First(&user, userID).Error; err != nil {
		response.NotFound(c, "用户不存在")
		return
	}

	// 获取用户权限
	perms, _ := middleware.GetUserPermissions(userID)

	response.Success(c, gin.H{
		"user":        user.ToResponse(),
		"roles":       user.Roles,
		"permissions": perms,
	})
}

// CheckPermission 检查权限
func (ac *AuthController) CheckPermission(c *gin.Context) {
	type CheckRequest struct {
		Resource string `json:"resource" binding:"required"`
		Action   string `json:"action" binding:"required"`
	}

	var req CheckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	userID := middleware.GetUserID(c)
	perms, err := middleware.GetUserPermissions(userID)
	if err != nil {
		response.InternalServerError(c, "获取权限失败")
		return
	}

	required := req.Resource + ":" + req.Action
	hasPermission := middleware.MatchPermission(perms, required)

	response.Success(c, gin.H{
		"has_permission": hasPermission,
		"permission":     required,
	})
}

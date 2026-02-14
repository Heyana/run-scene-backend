package database

import (
	"fmt"
	"go_wails_project_manager/logger"
	"go_wails_project_manager/middleware"
	"go_wails_project_manager/models"

	"gorm.io/gorm"
)

// InitAuthSystem 初始化权限系统
func InitAuthSystem(db *gorm.DB) error {
	logger.Log.Info("开始初始化权限系统...")

	// 1. 初始化权限
	if err := initSystemPermissions(db); err != nil {
		return fmt.Errorf("初始化权限失败: %w", err)
	}

	// 2. 初始化权限组
	if err := initSystemPermissionGroups(db); err != nil {
		return fmt.Errorf("初始化权限组失败: %w", err)
	}

	// 3. 初始化角色
	if err := initSystemRoles(db); err != nil {
		return fmt.Errorf("初始化角色失败: %w", err)
	}

	// 4. 创建默认管理员
	if err := initDefaultAdmin(db); err != nil {
		return fmt.Errorf("创建默认管理员失败: %w", err)
	}

	logger.Log.Info("权限系统初始化完成")
	return nil
}

// initSystemPermissions 初始化系统权限
func initSystemPermissions(db *gorm.DB) error {
	permissions := []models.Permission{
		// 文档权限
		{Code: "documents:read", Name: "查看文档", Resource: "documents", Action: "read", IsSystem: true},
		{Code: "documents:create", Name: "创建文档", Resource: "documents", Action: "create", IsSystem: true},
		{Code: "documents:update", Name: "更新文档", Resource: "documents", Action: "update", IsSystem: true},
		{Code: "documents:delete", Name: "删除文档", Resource: "documents", Action: "delete", IsSystem: true},
		{Code: "documents:download", Name: "下载文档", Resource: "documents", Action: "download", IsSystem: true},
		{Code: "documents:upload", Name: "上传文档", Resource: "documents", Action: "upload", IsSystem: true},
		{Code: "documents:share", Name: "分享文档", Resource: "documents", Action: "share", IsSystem: true},
		{Code: "documents:admin", Name: "文档管理", Resource: "documents", Action: "admin", IsSystem: true},

		// 模型权限
		{Code: "models:read", Name: "查看模型", Resource: "models", Action: "read", IsSystem: true},
		{Code: "models:create", Name: "创建模型", Resource: "models", Action: "create", IsSystem: true},
		{Code: "models:update", Name: "更新模型", Resource: "models", Action: "update", IsSystem: true},
		{Code: "models:delete", Name: "删除模型", Resource: "models", Action: "delete", IsSystem: true},
		{Code: "models:download", Name: "下载模型", Resource: "models", Action: "download", IsSystem: true},
		{Code: "models:upload", Name: "上传模型", Resource: "models", Action: "upload", IsSystem: true},
		{Code: "models:admin", Name: "模型管理", Resource: "models", Action: "admin", IsSystem: true},

		// 资产权限
		{Code: "assets:read", Name: "查看资产", Resource: "assets", Action: "read", IsSystem: true},
		{Code: "assets:create", Name: "创建资产", Resource: "assets", Action: "create", IsSystem: true},
		{Code: "assets:update", Name: "更新资产", Resource: "assets", Action: "update", IsSystem: true},
		{Code: "assets:delete", Name: "删除资产", Resource: "assets", Action: "delete", IsSystem: true},
		{Code: "assets:download", Name: "下载资产", Resource: "assets", Action: "download", IsSystem: true},
		{Code: "assets:upload", Name: "上传资产", Resource: "assets", Action: "upload", IsSystem: true},
		{Code: "assets:admin", Name: "资产管理", Resource: "assets", Action: "admin", IsSystem: true},

		// 贴图权限
		{Code: "textures:read", Name: "查看贴图", Resource: "textures", Action: "read", IsSystem: true},
		{Code: "textures:download", Name: "下载贴图", Resource: "textures", Action: "download", IsSystem: true},
		{Code: "textures:sync", Name: "同步贴图", Resource: "textures", Action: "sync", IsSystem: true},
		{Code: "textures:admin", Name: "贴图管理", Resource: "textures", Action: "admin", IsSystem: true},

		// 项目权限
		{Code: "projects:read", Name: "查看项目", Resource: "projects", Action: "read", IsSystem: true},
		{Code: "projects:create", Name: "创建项目", Resource: "projects", Action: "create", IsSystem: true},
		{Code: "projects:update", Name: "更新项目", Resource: "projects", Action: "update", IsSystem: true},
		{Code: "projects:delete", Name: "删除项目", Resource: "projects", Action: "delete", IsSystem: true},
		{Code: "projects:admin", Name: "项目管理", Resource: "projects", Action: "admin", IsSystem: true},

		// AI3D权限
		{Code: "ai3d:read", Name: "查看AI3D任务", Resource: "ai3d", Action: "read", IsSystem: true},
		{Code: "ai3d:create", Name: "创建AI3D任务", Resource: "ai3d", Action: "create", IsSystem: true},
		{Code: "ai3d:delete", Name: "删除AI3D任务", Resource: "ai3d", Action: "delete", IsSystem: true},
		{Code: "ai3d:admin", Name: "AI3D管理", Resource: "ai3d", Action: "admin", IsSystem: true},

		// 用户管理权限
		{Code: "users:read", Name: "查看用户", Resource: "users", Action: "read", IsSystem: true},
		{Code: "users:create", Name: "创建用户", Resource: "users", Action: "create", IsSystem: true},
		{Code: "users:update", Name: "更新用户", Resource: "users", Action: "update", IsSystem: true},
		{Code: "users:delete", Name: "删除用户", Resource: "users", Action: "delete", IsSystem: true},
		{Code: "users:admin", Name: "用户管理", Resource: "users", Action: "admin", IsSystem: true},

		// 角色管理权限
		{Code: "roles:read", Name: "查看角色", Resource: "roles", Action: "read", IsSystem: true},
		{Code: "roles:create", Name: "创建角色", Resource: "roles", Action: "create", IsSystem: true},
		{Code: "roles:update", Name: "更新角色", Resource: "roles", Action: "update", IsSystem: true},
		{Code: "roles:delete", Name: "删除角色", Resource: "roles", Action: "delete", IsSystem: true},
		{Code: "roles:admin", Name: "角色管理", Resource: "roles", Action: "admin", IsSystem: true},

		// 权限管理权限
		{Code: "permissions:read", Name: "查看权限", Resource: "permissions", Action: "read", IsSystem: true},
		{Code: "permissions:create", Name: "创建权限", Resource: "permissions", Action: "create", IsSystem: true},
		{Code: "permissions:update", Name: "更新权限", Resource: "permissions", Action: "update", IsSystem: true},
		{Code: "permissions:delete", Name: "删除权限", Resource: "permissions", Action: "delete", IsSystem: true},
		{Code: "permissions:admin", Name: "权限管理", Resource: "permissions", Action: "admin", IsSystem: true},

		// 通配符权限 - 全局操作权限
		{Code: "*:read", Name: "全局读权限", Resource: "*", Action: "read", Description: "所有资源的查看权限", IsSystem: true},
		{Code: "*:create", Name: "全局创建权限", Resource: "*", Action: "create", Description: "所有资源的创建权限", IsSystem: true},
		{Code: "*:update", Name: "全局更新权限", Resource: "*", Action: "update", Description: "所有资源的更新权限", IsSystem: true},
		{Code: "*:delete", Name: "全局删除权限", Resource: "*", Action: "delete", Description: "所有资源的删除权限", IsSystem: true},
		{Code: "*:download", Name: "全局下载权限", Resource: "*", Action: "download", Description: "所有资源的下载权限", IsSystem: true},
		{Code: "*:upload", Name: "全局上传权限", Resource: "*", Action: "upload", Description: "所有资源的上传权限", IsSystem: true},
		{Code: "*:admin", Name: "全局管理权限", Resource: "*", Action: "admin", Description: "所有资源的管理权限", IsSystem: true},
		{Code: "*:*", Name: "超级权限", Resource: "*", Action: "*", Description: "所有资源的所有权限", IsSystem: true},
	}

	for _, perm := range permissions {
		var existing models.Permission
		if err := db.Where("code = ?", perm.Code).First(&existing).Error; err != nil {
			if err := db.Create(&perm).Error; err != nil {
				return err
			}
			logger.Log.Debugf("创建权限: %s", perm.Code)
		}
	}

	return nil
}

// initSystemPermissionGroups 初始化系统权限组
func initSystemPermissionGroups(db *gorm.DB) error {
	groups := []struct {
		Group       models.PermissionGroup
		Permissions []string
	}{
		{
			Group: models.PermissionGroup{
				Code:        models.PermGroupDocumentViewer,
				Name:        "文档查看者",
				Description: "可以查看和下载文档",
				IsSystem:    true,
			},
			Permissions: []string{"documents:read", "documents:download"},
		},
		{
			Group: models.PermissionGroup{
				Code:        models.PermGroupDocumentManager,
				Name:        "文档管理员",
				Description: "文档的完整管理权限",
				IsSystem:    true,
			},
			Permissions: []string{"documents:read", "documents:create", "documents:update", "documents:delete", "documents:download", "documents:upload", "documents:share"},
		},
		{
			Group: models.PermissionGroup{
				Code:        models.PermGroupModelManager,
				Name:        "模型管理员",
				Description: "模型的完整管理权限",
				IsSystem:    true,
			},
			Permissions: []string{"models:read", "models:create", "models:update", "models:delete", "models:download", "models:upload"},
		},
		{
			Group: models.PermissionGroup{
				Code:        models.PermGroupAssetManager,
				Name:        "资产管理员",
				Description: "资产的完整管理权限",
				IsSystem:    true,
			},
			Permissions: []string{"assets:read", "assets:create", "assets:update", "assets:delete", "assets:download", "assets:upload"},
		},
		{
			Group: models.PermissionGroup{
				Code:        models.PermGroupResourceViewer,
				Name:        "资源查看者",
				Description: "所有资源的查看和下载权限",
				IsSystem:    true,
			},
			Permissions: []string{
				"documents:read", "documents:download",
				"models:read", "models:download",
				"assets:read", "assets:download",
				"textures:read", "textures:download",
				"projects:read",
			},
		},
		{
			Group: models.PermissionGroup{
				Code:        models.PermGroupResourceEditor,
				Name:        "资源编辑者",
				Description: "所有资源的编辑权限",
				IsSystem:    true,
			},
			Permissions: []string{
				"documents:read", "documents:create", "documents:update", "documents:download", "documents:upload",
				"models:read", "models:create", "models:update", "models:download", "models:upload",
				"assets:read", "assets:create", "assets:update", "assets:download", "assets:upload",
				"textures:read", "textures:download",
				"projects:read", "projects:create", "projects:update",
				"ai3d:read", "ai3d:create",
			},
		},
		{
			Group: models.PermissionGroup{
				Code:        models.PermGroupSystemAdmin,
				Name:        "系统管理员",
				Description: "用户、角色、权限管理",
				IsSystem:    true,
			},
			Permissions: []string{
				"users:read", "users:create", "users:update", "users:delete",
				"roles:read", "roles:create", "roles:update", "roles:delete",
				"permissions:read", "permissions:create", "permissions:update", "permissions:delete",
			},
		},
		{
			Group: models.PermissionGroup{
				Code:        "global_viewer",
				Name:        "全局查看者",
				Description: "所有资源的查看权限（使用通配符）",
				IsSystem:    true,
			},
			Permissions: []string{"*:read"},
		},
		{
			Group: models.PermissionGroup{
				Code:        "global_downloader",
				Name:        "全局下载者",
				Description: "所有资源的查看和下载权限",
				IsSystem:    true,
			},
			Permissions: []string{"*:read", "*:download"},
		},
		{
			Group: models.PermissionGroup{
				Code:        "global_editor",
				Name:        "全局编辑者",
				Description: "所有资源的编辑权限",
				IsSystem:    true,
			},
			Permissions: []string{"*:read", "*:create", "*:update", "*:download", "*:upload"},
		},
	}

	for _, item := range groups {
		var existing models.PermissionGroup
		if err := db.Where("code = ?", item.Group.Code).First(&existing).Error; err != nil {
			// 创建权限组
			if err := db.Create(&item.Group).Error; err != nil {
				return err
			}

			// 关联权限
			var permissions []models.Permission
			db.Where("code IN ?", item.Permissions).Find(&permissions)
			if len(permissions) > 0 {
				db.Model(&item.Group).Association("Permissions").Append(permissions)
			}
			logger.Log.Debugf("创建权限组: %s", item.Group.Code)
		}
	}

	return nil
}

// initSystemRoles 初始化系统角色
func initSystemRoles(db *gorm.DB) error {
	roles := []struct {
		Role            models.Role
		PermissionCodes []string
		GroupCodes      []string
	}{
		{
			Role: models.Role{
				Code:        models.RoleSuperAdmin,
				Name:        "超级管理员",
				Description: "系统最高权限",
				IsSystem:    true,
			},
			PermissionCodes: []string{"*:*"},
		},
		{
			Role: models.Role{
				Code:        models.RoleAdmin,
				Name:        "管理员",
				Description: "系统管理员，可管理用户和资源",
				IsSystem:    true,
			},
			GroupCodes: []string{models.PermGroupSystemAdmin, models.PermGroupResourceEditor},
		},
		{
			Role: models.Role{
				Code:        models.RoleEditor,
				Name:        "编辑者",
				Description: "可创建和编辑资源",
				IsSystem:    true,
			},
			GroupCodes: []string{models.PermGroupResourceEditor},
		},
		{
			Role: models.Role{
				Code:        models.RoleViewer,
				Name:        "查看者",
				Description: "只能查看和下载资源",
				IsSystem:    true,
			},
			GroupCodes: []string{models.PermGroupResourceViewer},
		},
	}

	for _, item := range roles {
		var existing models.Role
		if err := db.Where("code = ?", item.Role.Code).First(&existing).Error; err != nil {
			// 创建角色
			if err := db.Create(&item.Role).Error; err != nil {
				return err
			}

			// 关联权限
			if len(item.PermissionCodes) > 0 {
				var permissions []models.Permission
				db.Where("code IN ?", item.PermissionCodes).Find(&permissions)
				if len(permissions) > 0 {
					db.Model(&item.Role).Association("Permissions").Append(permissions)
				}
			}

			// 关联权限组
			if len(item.GroupCodes) > 0 {
				var groups []models.PermissionGroup
				db.Where("code IN ?", item.GroupCodes).Find(&groups)
				if len(groups) > 0 {
					db.Model(&item.Role).Association("PermissionGroups").Append(groups)
				}
			}
			logger.Log.Debugf("创建角色: %s", item.Role.Code)
		}
	}

	return nil
}

// initDefaultAdmin 创建默认管理员账号
func initDefaultAdmin(db *gorm.DB) error {
	var count int64
	db.Model(&models.User{}).Count(&count)

	// 如果已有用户，跳过
	if count > 0 {
		logger.Log.Info("已存在用户，跳过创建默认管理员")
		return nil
	}

	// 创建默认管理员
	hashedPassword, err := middleware.HashPassword("admin123456")
	if err != nil {
		return err
	}

	admin := models.User{
		Username: "admin",
		Password: hashedPassword,
		Email:    "admin@example.com",
		RealName: "系统管理员",
		Status:   models.UserStatusActive,
	}

	if err := db.Create(&admin).Error; err != nil {
		return err
	}

	// 分配超级管理员角色
	var superAdminRole models.Role
	if err := db.Where("code = ?", models.RoleSuperAdmin).First(&superAdminRole).Error; err != nil {
		return err
	}

	db.Model(&admin).Association("Roles").Append(&superAdminRole)

	logger.Log.Info("✅ 默认管理员账号已创建")
	logger.Log.Info("   用户名: admin")
	logger.Log.Info("   密码: admin123456")
	logger.Log.Warn("⚠️  生产环境请立即修改默认密码！")

	return nil
}

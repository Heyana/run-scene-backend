package database

import (
	"go_wails_project_manager/logger"
	"go_wails_project_manager/models/requirement"

	"gorm.io/gorm"
)

// InitRequirementTables 初始化需求管理平台表
func InitRequirementTables(db *gorm.DB) error {
	logger.Log.Info("开始初始化需求管理平台数据表...")

	// 自动迁移表结构
	err := db.AutoMigrate(
		&requirement.Company{},
		&requirement.CompanyMember{},
		&requirement.Project{},
		&requirement.ProjectMember{},
		&requirement.MissionList{},
		&requirement.MissionColumn{}, // 新增：任务列
		&requirement.Mission{},
		&requirement.MissionComment{},
		&requirement.MissionAttachment{},
		&requirement.MissionRelation{},
		&requirement.MissionLog{},
		&requirement.MissionTag{},
	)

	if err != nil {
		logger.Log.Errorf("需求管理平台表初始化失败: %v", err)
		return err
	}

	logger.Log.Info("需求管理平台数据表初始化完成")

	// 初始化权限数据
	if err := initRequirementPermissions(db); err != nil {
		logger.Log.Errorf("需求管理权限初始化失败: %v", err)
		return err
	}

	return nil
}

// initRequirementPermissions 初始化需求管理权限
func initRequirementPermissions(db *gorm.DB) error {
	logger.Log.Info("开始初始化需求管理权限...")

	// 权限定义
	permissions := []struct {
		Code        string
		Name        string
		Resource    string
		Action      string
		Description string
	}{
		// 公司管理权限
		{Code: "requirement:company:read", Name: "查看公司", Resource: "requirement_company", Action: "read", Description: "查看需求管理公司信息"},
		{Code: "requirement:company:create", Name: "创建公司", Resource: "requirement_company", Action: "create", Description: "创建需求管理公司"},
		{Code: "requirement:company:update", Name: "更新公司", Resource: "requirement_company", Action: "update", Description: "更新需求管理公司信息"},
		{Code: "requirement:company:delete", Name: "删除公司", Resource: "requirement_company", Action: "delete", Description: "删除需求管理公司"},
		{Code: "requirement:company:admin", Name: "管理公司", Resource: "requirement_company", Action: "admin", Description: "管理需求管理公司成员"},

		// 项目管理权限
		{Code: "requirement:project:read", Name: "查看项目", Resource: "requirement_project", Action: "read", Description: "查看需求管理项目"},
		{Code: "requirement:project:create", Name: "创建项目", Resource: "requirement_project", Action: "create", Description: "创建需求管理项目"},
		{Code: "requirement:project:update", Name: "更新项目", Resource: "requirement_project", Action: "update", Description: "更新需求管理项目"},
		{Code: "requirement:project:delete", Name: "删除项目", Resource: "requirement_project", Action: "delete", Description: "删除需求管理项目"},
		{Code: "requirement:project:admin", Name: "管理项目", Resource: "requirement_project", Action: "admin", Description: "管理需求管理项目成员"},

		// 任务管理权限
		{Code: "requirement:mission:read", Name: "查看任务", Resource: "requirement_mission", Action: "read", Description: "查看需求管理任务"},
		{Code: "requirement:mission:create", Name: "创建任务", Resource: "requirement_mission", Action: "create", Description: "创建需求管理任务"},
		{Code: "requirement:mission:update", Name: "更新任务", Resource: "requirement_mission", Action: "update", Description: "更新需求管理任务"},
		{Code: "requirement:mission:delete", Name: "删除任务", Resource: "requirement_mission", Action: "delete", Description: "删除需求管理任务"},
	}

	// 检查并创建权限
	for _, perm := range permissions {
		var count int64
		db.Table("permissions").Where("code = ?", perm.Code).Count(&count)
		
		if count == 0 {
			// 权限不存在，使用 GORM 创建
			result := db.Table("permissions").Create(map[string]interface{}{
				"code":        perm.Code,
				"name":        perm.Name,
				"resource":    perm.Resource,
				"action":      perm.Action,
				"description": perm.Description,
			})
			
			if result.Error != nil {
				logger.Log.Warnf("创建权限失败 %s: %v", perm.Code, result.Error)
			} else {
				logger.Log.Debugf("创建权限: %s", perm.Code)
			}
		}
	}

	// 创建权限组
	groups := []struct {
		Code        string
		Name        string
		Description string
		Permissions []string
	}{
		{
			Code:        "requirement_admin",
			Name:        "需求管理管理员",
			Description: "需求管理平台的完整管理权限",
			Permissions: []string{
				"requirement:company:read", "requirement:company:create", "requirement:company:update", "requirement:company:delete", "requirement:company:admin",
				"requirement:project:read", "requirement:project:create", "requirement:project:update", "requirement:project:delete", "requirement:project:admin",
				"requirement:mission:read", "requirement:mission:create", "requirement:mission:update", "requirement:mission:delete",
			},
		},
		{
			Code:        "requirement_member",
			Name:        "需求管理成员",
			Description: "可以创建和管理自己的任务",
			Permissions: []string{
				"requirement:company:read",
				"requirement:project:read",
				"requirement:mission:read", "requirement:mission:create", "requirement:mission:update",
			},
		},
		{
			Code:        "requirement_viewer",
			Name:        "需求管理观察者",
			Description: "只能查看需求管理内容",
			Permissions: []string{
				"requirement:company:read",
				"requirement:project:read",
				"requirement:mission:read",
			},
		},
	}

	for _, group := range groups {
		var count int64
		db.Table("permission_groups").Where("code = ?", group.Code).Count(&count)
		
		if count == 0 {
			// 权限组不存在，使用 GORM 创建
			result := db.Table("permission_groups").Create(map[string]interface{}{
				"code":        group.Code,
				"name":        group.Name,
				"description": group.Description,
				"is_system":   true,
			})
			
			if result.Error != nil {
				logger.Log.Warnf("创建权限组失败 %s: %v", group.Code, result.Error)
				continue
			}

			// 获取新创建的权限组ID
			var groupID uint
			db.Table("permission_groups").Where("code = ?", group.Code).Select("id").Scan(&groupID)

			// 关联权限
			for _, permCode := range group.Permissions {
				var permID uint
				if err := db.Table("permissions").Where("code = ?", permCode).Select("id").Scan(&permID).Error; err == nil && permID > 0 {
					db.Table("permission_group_permissions").Create(map[string]interface{}{
						"permission_group_id": groupID,
						"permission_id":       permID,
					})
				}
			}

			logger.Log.Debugf("创建权限组: %s", group.Code)
		}
	}

	logger.Log.Info("需求管理权限初始化完成")
	return nil
}

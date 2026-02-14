package auth

import (
	"go_wails_project_manager/models"

	"gorm.io/gorm"
)

// PermissionCalculatorService 权限计算服务
type PermissionCalculatorService struct {
	db *gorm.DB
}

// NewPermissionCalculatorService 创建权限计算服务
func NewPermissionCalculatorService(db *gorm.DB) *PermissionCalculatorService {
	return &PermissionCalculatorService{db: db}
}

// CalculateUserPermissions 计算用户最终权限
func (pcs *PermissionCalculatorService) CalculateUserPermissions(userID uint) ([]string, error) {
	var user models.User
	if err := pcs.db.Preload("Roles.Permissions").
		Preload("Roles.PermissionGroups.Permissions").
		Preload("Permissions").
		Preload("PermissionGroups.Permissions").
		First(&user, userID).Error; err != nil {
		return nil, err
	}

	permMap := make(map[string]bool)

	// 1. 从角色获取权限
	for _, role := range user.Roles {
		for _, perm := range role.Permissions {
			permMap[perm.Code] = true
		}
		// 从角色的权限组获取权限
		for _, group := range role.PermissionGroups {
			for _, perm := range group.Permissions {
				permMap[perm.Code] = true
			}
		}
	}

	// 2. 从用户直接授权获取权限
	for _, perm := range user.Permissions {
		permMap[perm.Code] = true
	}

	// 3. 从用户直接授权的权限组获取权限
	for _, group := range user.PermissionGroups {
		for _, perm := range group.Permissions {
			permMap[perm.Code] = true
		}
	}

	// 转换为数组
	perms := make([]string, 0, len(permMap))
	for perm := range permMap {
		perms = append(perms, perm)
	}

	return perms, nil
}

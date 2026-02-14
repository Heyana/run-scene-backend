package middleware

import (
	"fmt"
	"go_wails_project_manager/response"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// 权限缓存
var (
	permissionCache      = make(map[uint][]string)
	permissionCacheMutex sync.RWMutex
	cacheExpire          = 5 * time.Minute
	cacheTimestamps      = make(map[uint]time.Time)
)

// PermissionCalculator 权限计算器接口
type PermissionCalculator interface {
	CalculateUserPermissions(userID uint) ([]string, error)
}

var permCalculator PermissionCalculator

// SetPermissionCalculator 设置权限计算器
func SetPermissionCalculator(calculator PermissionCalculator) {
	permCalculator = calculator
}

// RequirePermission 权限验证中间件
func RequirePermission(resource, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := GetUserID(c)
		if userID == 0 {
			response.Unauthorized(c, "请先登录")
			c.Abort()
			return
		}

		// 获取用户权限
		perms, err := GetUserPermissions(userID)
		if err != nil {
			response.InternalServerError(c, "获取权限失败")
			c.Abort()
			return
		}

		// 检查权限
		required := fmt.Sprintf("%s:%s", resource, action)
		if !MatchPermission(perms, required) {
			response.Forbidden(c, "权限不足")
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAnyPermission 任一权限验证
func RequireAnyPermission(permissions ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := GetUserID(c)
		if userID == 0 {
			response.Unauthorized(c, "请先登录")
			c.Abort()
			return
		}

		userPerms, err := GetUserPermissions(userID)
		if err != nil {
			response.InternalServerError(c, "获取权限失败")
			c.Abort()
			return
		}

		for _, perm := range permissions {
			if MatchPermission(userPerms, perm) {
				c.Next()
				return
			}
		}

		response.Forbidden(c, "权限不足")
		c.Abort()
	}
}

// RequireAllPermissions 全部权限验证
func RequireAllPermissions(permissions ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := GetUserID(c)
		if userID == 0 {
			response.Unauthorized(c, "请先登录")
			c.Abort()
			return
		}

		userPerms, err := GetUserPermissions(userID)
		if err != nil {
			response.InternalServerError(c, "获取权限失败")
			c.Abort()
			return
		}

		for _, perm := range permissions {
			if !MatchPermission(userPerms, perm) {
				response.Forbidden(c, "权限不足")
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// GetUserPermissions 获取用户所有权限
func GetUserPermissions(userID uint) ([]string, error) {
	// 检查缓存
	permissionCacheMutex.RLock()
	if perms, ok := permissionCache[userID]; ok {
		if timestamp, exists := cacheTimestamps[userID]; exists {
			if time.Since(timestamp) < cacheExpire {
				permissionCacheMutex.RUnlock()
				return perms, nil
			}
		}
	}
	permissionCacheMutex.RUnlock()

	// 使用权限计算器
	if permCalculator == nil {
		return nil, fmt.Errorf("权限计算器未初始化")
	}

	perms, err := permCalculator.CalculateUserPermissions(userID)
	if err != nil {
		return nil, err
	}

	// 更新缓存
	permissionCacheMutex.Lock()
	permissionCache[userID] = perms
	cacheTimestamps[userID] = time.Now()
	permissionCacheMutex.Unlock()

	return perms, nil
}

// MatchPermission 权限匹配
func MatchPermission(userPerms []string, required string) bool {
	// 精确匹配
	for _, perm := range userPerms {
		if perm == required {
			return true
		}
	}

	// 通配符匹配
	parts := strings.Split(required, ":")
	if len(parts) == 2 {
		resource, action := parts[0], parts[1]

		for _, perm := range userPerms {
			// *:* 超级权限
			if perm == "*:*" {
				return true
			}

			permParts := strings.Split(perm, ":")
			if len(permParts) == 2 {
				permResource, permAction := permParts[0], permParts[1]

				// resource:*
				if permResource == resource && permAction == "*" {
					return true
				}

				// *:action
				if permResource == "*" && permAction == action {
					return true
				}
			}
		}
	}

	return false
}

// InvalidatePermissionCache 清除用户权限缓存
func InvalidatePermissionCache(userID uint) {
	permissionCacheMutex.Lock()
	delete(permissionCache, userID)
	delete(cacheTimestamps, userID)
	permissionCacheMutex.Unlock()
}

// ClearAllPermissionCache 清除所有权限缓存
func ClearAllPermissionCache() {
	permissionCacheMutex.Lock()
	permissionCache = make(map[uint][]string)
	cacheTimestamps = make(map[uint]time.Time)
	permissionCacheMutex.Unlock()
}

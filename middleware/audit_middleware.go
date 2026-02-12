package middleware

import (
	"go_wails_project_manager/services/audit"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// AuditMiddleware 审计中间件
func AuditMiddleware(auditService *audit.AuditService) gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		// 处理请求
		c.Next()

		// 推断操作类型和资源类型
		action := audit.InferActionFromMethod(c.Request.Method, c.Request.URL.Path)
		resource := audit.InferResourceFromPath(c.Request.URL.Path)

		// 尝试从上下文获取资源ID
		var resourceID *uint
		if idStr := c.Param("id"); idStr != "" {
			if id, err := strconv.ParseUint(idStr, 10, 32); err == nil {
				uid := uint(id)
				resourceID = &uid
			}
		}

		// 记录审计日志
		auditService.LogFromContext(c, startTime, action, resource, resourceID)
	}
}

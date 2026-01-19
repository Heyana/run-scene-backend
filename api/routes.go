// Package api 提供REST API实现
package api

import (
	"time"

	"go_wails_project_manager/config"
	"go_wails_project_manager/controllers"
	"go_wails_project_manager/middleware"
	"go_wails_project_manager/response"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// Ping 健康检查
// @Summary 健康检查
// @Description 检查API服务是否正常运行
// @Tags 系统
// @Accept json
// @Produce json
// @Success 200 {object} response.Response "服务正常"
// @Router /api/ping [get]
func Ping(c *gin.Context) {
	response.SuccessWithMsg(c, "服务正常运行", gin.H{
		"timestamp": time.Now().Unix(),
		"status":    "ok",
	})
}

// RegisterRoutes 注册API路由
func RegisterRoutes(router *gin.Engine, log *logrus.Logger) {
	// 初始化有用的控制器
	backupController := controllers.NewBackupController()
	securityController := controllers.NewSecurityController()

	// 设置API文档（仅开发环境）
	if config.IsDev() {
		SetupAPIDocs(router)
	}

	// ==================== 全局安全中间件（按优先级顺序）====================
	// 0. Panic恢复中间件（最先添加，确保捕获所有panic）
	router.Use(middleware.RecoveryWithLog(config.IsDev()))

	// 1. 请求ID追踪（便于日志追踪）
	router.Use(requestid.New())

	// 1. CORS中间件（支持白名单）
	router.Use(CorsMiddleware())

	// 2. 安全响应头
	router.Use(SecurityHeadersMiddleware())

	// 3. IP黑名单过滤（最高优先级 - 直接拒绝已知恶意IP）
	router.Use(IPFilterMiddleware())

	// 4. DDoS防护（并发连接限制 + 请求频率检测）
	router.Use(DDoSProtectionMiddleware())

	// 5. 连接频率限制
	router.Use(ConnectionRateLimitMiddleware())

	// 6. 速率限制（令牌桶算法）
	router.Use(RateLimitMiddleware())

	// 7. 可疑活动检测（SQL注入/XSS/路径遍历）
	router.Use(DetectSuspiciousActivityMiddleware())

	// 8. 请求大小限制（全局100MB）
	router.Use(RequestSizeLimitMiddleware(100 * 1024 * 1024))

	// 9. 敏感路径保护
	router.Use(ProtectSensitivePathsMiddleware())

	// 根级别健康检查端点（供 Electron 检测）
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":    "ok",
			"timestamp": time.Now().Unix(),
			"service":   "media-manager-server",
		})
	})

	// 设置API路由组
	api := router.Group("/api")
	{
		api.GET("/ping", Ping)

		// 备份管理API
		backup := api.Group("/backup")
		{
			backup.GET("/status", backupController.GetStatus)                             // 获取备份状态
			backup.POST("/trigger", backupController.TriggerManualBackup)                 // 手动触发全量备份
			backup.POST("/database", backupController.TriggerDatabaseBackup)              // 手动触发数据库备份
			backup.POST("/cdn", backupController.TriggerCDNBackup)                        // 手动触发CDN备份
			backup.GET("/history", backupController.GetBackupHistory)                     // 获取备份历史
			backup.POST("/restore/cdn/:backup_id", backupController.RestoreCDNFromBackup) // CDN文件恢复
		}

		// 安全管理API
		security := api.Group("/security")
		{
			security.GET("/status", securityController.GetStatus)                     // 获取安全状态
			security.GET("/blocked-ips", securityController.GetBlockedIPs)            // 获取被封禁IP列表
			security.POST("/unblock/:ip", securityController.UnblockIP)               // 解封IP地址
			security.GET("/ip-stats", securityController.GetIPStats)                  // 获取IP统计信息
			security.POST("/block/:ip", securityController.BlockIP)                   // 封禁IP地址
			security.POST("/whitelist/:ip", securityController.AddToWhitelist)        // 添加IP到白名单
			security.DELETE("/whitelist/:ip", securityController.RemoveFromWhitelist) // 从白名单移除IP
			security.GET("/connections", securityController.GetConnections)           // 获取连接统计
		}

		// TODO: 添加其他业务控制器和路由

		// ==================== 认证路由（示例）====================
		// 初始化JWT认证器
		jwtAuth := middleware.NewJWTAuth()

		auth := api.Group("/auth")
		{
			// 登录（需要实现validateUser函数）
			// auth.POST("/login", jwtAuth.LoginHandler(validateUser))
			auth.POST("/refresh", jwtAuth.RefreshHandler()) // 刷新token
			auth.POST("/logout", jwtAuth.LogoutHandler())   // 登出
		}

		// 需要认证的路由示例
		_ = jwtAuth // 使用jwtAuth.AuthMiddleware()保护需要认证的路由
		// protectedGroup := api.Group("/protected")
		// protectedGroup.Use(jwtAuth.AuthMiddleware())
		// {
		//     protectedGroup.GET("/profile", getProfile)
		// }

	}
}

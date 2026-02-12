// Package api 提供REST API实现
package api

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"go_wails_project_manager/config"
	"go_wails_project_manager/logger"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// ==================== 统一安全响应 ====================

// securityResponse 统一的安全响应函数
func securityResponse(c *gin.Context) {
	c.Status(http.StatusNotFound)
	c.Abort()
}

// ==================== 请求ID日志中间件 ====================

// getLoggerWithRequestID 获取带请求ID的日志记录器
func getLoggerWithRequestID(c *gin.Context) *logrus.Entry {
	requestID := requestid.Get(c)
	if requestID != "" {
		return logger.Log.WithField("request_id", requestID)
	}
	return logger.Log.WithField("request_id", "unknown")
}

// ==================== 速率限制中间件 ====================

// RateLimiter 速率限制器
type RateLimiter struct {
	visitors map[string]*Visitor
	mu       sync.RWMutex
	rate     time.Duration
	burst    int
}

// Visitor 访问者信息
type Visitor struct {
	lastSeen   time.Time
	tokens     int
	blockUntil *time.Time // 封禁截止时间
}

var (
	rateLimiter *RateLimiter
	once        sync.Once
)

// GetRateLimiter 获取全局速率限制器（单例）
func GetRateLimiter() *RateLimiter {
	once.Do(func() {
		rateLimiter = &RateLimiter{
			visitors: make(map[string]*Visitor),
			rate:     time.Second, // 每秒补充一个token
			burst:    10000,       // 桶容量
		}
		// 定期清理过期访客（每10分钟）
		go rateLimiter.cleanupVisitors()
	})
	return rateLimiter
}

// Allow 检查是否允许请求
func (rl *RateLimiter) Allow(ip string) (allowed bool, resetTime time.Duration) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	visitor, exists := rl.visitors[ip]

	if !exists {
		visitor = &Visitor{
			lastSeen: now,
			tokens:   rl.burst - 1,
		}
		rl.visitors[ip] = visitor
		return true, 0
	}

	// 检查是否在封禁期
	if visitor.blockUntil != nil && now.Before(*visitor.blockUntil) {
		resetTime := visitor.blockUntil.Sub(now)
		return false, resetTime
	}

	// 清除过期的封禁
	if visitor.blockUntil != nil && now.After(*visitor.blockUntil) {
		visitor.blockUntil = nil
		visitor.tokens = rl.burst
	}

	// 计算应该补充的token数量
	elapsed := now.Sub(visitor.lastSeen)
	tokensToAdd := int(elapsed / rl.rate)

	if tokensToAdd > 0 {
		visitor.tokens += tokensToAdd
		if visitor.tokens > rl.burst {
			visitor.tokens = rl.burst
		}
		visitor.lastSeen = now
	}

	// 检查是否有可用token
	if visitor.tokens > 0 {
		visitor.tokens--
		return true, 0
	}

	// Token耗尽，计算下次可用时间
	resetTime = rl.rate - elapsed%rl.rate
	return false, resetTime
}

// BlockIP 临时封禁IP（用于检测到恶意行为时）
func (rl *RateLimiter) BlockIP(ip string, duration time.Duration) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	blockUntil := time.Now().Add(duration)
	visitor, exists := rl.visitors[ip]
	if !exists {
		visitor = &Visitor{
			lastSeen: time.Now(),
			tokens:   0,
		}
		rl.visitors[ip] = visitor
	}
	visitor.blockUntil = &blockUntil

	logger.Log.Warnf("🚫 已封禁IP: %s, 封禁时长: %v, 截止: %v", ip, duration, blockUntil)
}

// cleanupVisitors 定期清理过期访客
func (rl *RateLimiter) cleanupVisitors() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for ip, visitor := range rl.visitors {
			// 清理10分钟未活动且未封禁的访客
			if visitor.blockUntil == nil && now.Sub(visitor.lastSeen) > 10*time.Minute {
				delete(rl.visitors, ip)
			}
			// 清理已解封的访客
			if visitor.blockUntil != nil && now.After(*visitor.blockUntil) {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// RateLimitMiddleware 速率限制中间件
func RateLimitMiddleware() gin.HandlerFunc {
	limiter := GetRateLimiter()
	securityCfg := config.GetSecurityConfig()

	return func(c *gin.Context) {
		path := c.Request.URL.Path

		// 排除静态资源
		if isStaticResource(path) {
			c.Next()
			return
		}

		ip := getClientIP(c)
		allowed, resetTime := limiter.Allow(ip)
		if !allowed {
			getLoggerWithRequestID(c).Warnf("⚠️ IP超过速率限制: %s", ip)

			c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", securityCfg.RateLimitPerSecond))
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", int(resetTime.Seconds())))
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "请求过于频繁，请稍后再试",
				"retry_after": int(resetTime.Seconds()),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// isStaticResource 判断是否为静态资源
func isStaticResource(path string) bool {
	// 排除 /website/ 下的所有静态资源
	if strings.HasPrefix(path, "/website/") {
		return true
	}

	// 排除常见的静态文件扩展名
	staticExtensions := []string{
		".css", ".js", ".map",
		".png", ".jpg", ".jpeg", ".gif", ".webp", ".svg", ".ico",
		".woff", ".woff2", ".ttf", ".eot",
		".mp4", ".webm", ".mp3", ".wav",
		".pdf", ".zip",
	}

	for _, ext := range staticExtensions {
		if strings.HasSuffix(path, ext) {
			return true
		}
	}

	return false
}

// ==================== 安全响应头中间件 ====================

// SecurityHeadersMiddleware 添加安全响应头
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path

		// 防止MIME类型嗅探
		c.Header("X-Content-Type-Options", "nosniff")

		// 防止点击劫持 - 对于 HTML 文档，不设置 X-Frame-Options 以允许 iframe 预览
		if strings.HasSuffix(path, ".html") || strings.HasSuffix(path, ".htm") {
			// HTML 文档不设置 X-Frame-Options，允许 iframe 嵌入
		} else {
			// 其他资源使用 SAMEORIGIN
			c.Header("X-Frame-Options", "SAMEORIGIN")
		}

		// XSS保护
		c.Header("X-XSS-Protection", "1; mode=block")

		// HTTPS强制（生产环境建议启用）
		if config.IsProd() {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		// 内容安全策略 - 对于 HTML 文档放宽限制
		if strings.HasSuffix(path, ".html") || strings.HasSuffix(path, ".htm") {
			// HTML 文档预览需要更宽松的 CSP
			c.Header("Content-Security-Policy", "default-src 'self' 'unsafe-inline' 'unsafe-eval' data: blob:; img-src 'self' data: https: blob:; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline';")
		} else {
			c.Header("Content-Security-Policy", "default-src 'self'; img-src 'self' data: https:; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline';")
		}

		// 引用来源策略
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// 权限策略
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		c.Next()
	}
}

// ==================== CORS中间件 ====================

// CorsMiddleware 返回一个处理CORS的中间件（支持动态白名单）
func CorsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")

		// 检查是否在白名单中
		if origin != "" && isOriginAllowed(origin) {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		} else if origin == "" {
			// 没有Origin头（非浏览器请求），允许访问
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		} else {
			// 不在白名单中，记录并拒绝
			getLoggerWithRequestID(c).Warnf("⚠️ CORS拒绝来源: %s, IP: %s", origin, c.ClientIP())
		}

		// 设置允许的方法
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		// 设置允许的头
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		// 设置暴露的头
		c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Length")
		// 设置预检请求的有效期
		c.Writer.Header().Set("Access-Control-Max-Age", "86400") // 24小时

		// 如果是预检请求，直接返回200
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(200)
			return
		}

		// 继续处理请求
		c.Next()
	}
}

// isOriginAllowed 检查来源是否允许
func isOriginAllowed(origin string) bool {
	// 开发环境允许localhost
	if strings.Contains(origin, "localhost") || strings.Contains(origin, "127.0.0.1") {
		return true
	}

	// 从配置中获取允许的源
	cfg := config.GetSecurityConfig()
	for _, allowedOrigin := range cfg.AllowedOrigins {
		if allowedOrigin == "*" || allowedOrigin == origin {
			return true
		}
	}

	return false
}

// ==================== IP黑名单/白名单中间件 ====================

var (
	ipBlacklist = make(map[string]time.Time) // IP -> 封禁截止时间
	ipWhitelist = make(map[string]bool)      // 白名单IP（永不限制）
	ipMutex     sync.RWMutex
)

// AddToBlacklist 添加IP到黑名单
func AddToBlacklist(ip string, duration time.Duration) {
	ipMutex.Lock()
	defer ipMutex.Unlock()

	expireTime := time.Now().Add(duration)
	ipBlacklist[ip] = expireTime

	// 同时封禁速率限制器
	GetRateLimiter().BlockIP(ip, duration)

	logger.Log.Warnf("🚫 IP已加入黑名单: %s, 有效期至: %v", ip, expireTime)
}

// RemoveFromBlacklist 从黑名单移除IP
func RemoveFromBlacklist(ip string) {
	ipMutex.Lock()
	defer ipMutex.Unlock()
	delete(ipBlacklist, ip)

	logger.Log.Infof("✅ IP已从黑名单移除: %s", ip)
}

// AddToWhitelist 添加IP到白名单
func AddToWhitelist(ip string) {
	ipMutex.Lock()
	defer ipMutex.Unlock()
	ipWhitelist[ip] = true
	logger.Log.Infof("✅ IP已加入白名单: %s", ip)
}

// RemoveFromWhitelist 从白名单移除IP
func RemoveFromWhitelist(ip string) {
	ipMutex.Lock()
	defer ipMutex.Unlock()
	delete(ipWhitelist, ip)
	logger.Log.Infof("❌ IP已从白名单移除: %s", ip)
}

// IsBlacklisted 检查IP是否在黑名单
func IsBlacklisted(ip string) (bool, time.Time) {
	ipMutex.RLock()
	defer ipMutex.RUnlock()

	expireTime, exists := ipBlacklist[ip]
	if !exists {
		return false, time.Time{}
	}

	// 检查是否已过期
	if time.Now().After(expireTime) {
		// 异步删除过期记录
		go func() {
			ipMutex.Lock()
			delete(ipBlacklist, ip)
			ipMutex.Unlock()
		}()
		return false, time.Time{}
	}

	return true, expireTime
}

// IsWhitelisted 检查IP是否在白名单
func IsWhitelisted(ip string) bool {
	ipMutex.RLock()
	defer ipMutex.RUnlock()
	return ipWhitelist[ip]
}

// IPFilterMiddleware IP过滤中间件
func IPFilterMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := getClientIP(c)

		// 白名单IP直接放行
		if IsWhitelisted(ip) {
			c.Next()
			return
		}

		// 检查黑名单
		if blocked, expireTime := IsBlacklisted(ip); blocked {
			getLoggerWithRequestID(c).Warnf("🚫 拦截黑名单IP: %s, 解封时间: %v", ip, expireTime)
			c.JSON(http.StatusForbidden, gin.H{
				"error":      "您的IP已被封禁",
				"unblock_at": expireTime.Unix(),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// ==================== 恶意行为检测中间件 ====================

var (
	suspiciousActivityCount = make(map[string]int) // IP -> 可疑活动次数
	suspiciousMutex         sync.RWMutex
)

// DetectSuspiciousActivityMiddleware 检测可疑活动
func DetectSuspiciousActivityMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := getClientIP(c)

		// 白名单IP不检测
		if IsWhitelisted(ip) {
			c.Next()
			return
		}

		// 检测SQL注入特征
		if detectSQLInjection(c) {
			incrementSuspiciousActivity(ip, "sql_injection", c)
		}

		// 检测路径遍历攻击（严重威胁，立即阻止）
		if detectPathTraversal(c) {
			incrementSuspiciousActivity(ip, "path_traversal", c)
			// 路径遍历攻击是严重威胁，立即阻止请求
			getLoggerWithRequestID(c).Errorf("🚫 检测到路径遍历攻击，立即阻止: IP=%s, Path=%s", ip, c.Request.URL.Path)
			securityResponse(c)
			return
		}

		// 检测XSS攻击
		if detectXSS(c) {
			incrementSuspiciousActivity(ip, "xss_attempt", c)
		}

		c.Next()
	}
}

// incrementSuspiciousActivity 增加可疑活动计数
func incrementSuspiciousActivity(ip, reason string, c ...*gin.Context) {
	suspiciousMutex.Lock()
	defer suspiciousMutex.Unlock()

	count := suspiciousActivityCount[ip]
	count++
	suspiciousActivityCount[ip] = count

	// 使用带请求ID的日志记录器（如果有context）
	var logEntry *logrus.Entry
	if len(c) > 0 && c[0] != nil {
		logEntry = getLoggerWithRequestID(c[0])
	} else {
		logEntry = logger.Log.WithField("request_id", "unknown")
	}
	logEntry.Warnf("⚠️ 检测到可疑活动: IP=%s, 原因=%s, 累计次数=%d", ip, reason, count)

	// 达到阈值则自动封禁（从配置读取）
	securityCfg := config.GetSecurityConfig()
	if count >= securityCfg.AutoBlockThreshold {
		if len(c) > 0 && c[0] != nil {
			getLoggerWithRequestID(c[0]).Errorf("🚫 检测到恶意行为，自动封禁IP: %s, 原因=%s", ip, reason)
		} else {
			logger.Log.Errorf("🚫 检测到恶意行为，自动封禁IP: %s, 原因=%s", ip, reason)
		}
		blockDuration := time.Duration(securityCfg.AutoBlockDuration) * time.Second

		AddToBlacklist(ip, blockDuration)
		suspiciousActivityCount[ip] = 0 // 重置计数
	}
}

// detectSQLInjection 检测SQL注入
func detectSQLInjection(c *gin.Context) bool {
	sqlPatterns := []string{
		"'", "\"", "--", ";", "/*", "*/", "xp_", "sp_",
		"union", "select", "insert", "update", "delete", "drop",
		"exec", "execute", "script", "javascript", "eval",
	}

	// 检查URL参数
	for _, value := range c.Request.URL.Query() {
		for _, v := range value {
			lowerV := strings.ToLower(v)
			for _, pattern := range sqlPatterns {
				if strings.Contains(lowerV, pattern) {
					getLoggerWithRequestID(c).Warnf("⚠️ 疑似SQL注入: %s", v)
					return true
				}
			}
		}
	}

	return false
}

// detectPathTraversal 检测路径遍历攻击
func detectPathTraversal(c *gin.Context) bool {
	// 检查URL路径
	path := c.Request.URL.Path

	dangerousPatterns := []string{
		"../", "..\\", "..",
		"./..", ".\\..",
		"%2e%2e", "%252e", "..%2f", "..%5c",
		"%2e%2e%2f", "%2e%2e%5c",
		"/etc/", "/proc/", "/sys/", "c:\\", "c:/",
		"/@fs/", "/@fs", "/fs/",
		"/etc/passwd", "/etc/shadow", "/etc/hosts",
		"/etc/group", "/etc/sudoers",
	}

	// 检查URL路径
	lowerPath := strings.ToLower(path)
	for _, pattern := range dangerousPatterns {
		if strings.Contains(lowerPath, pattern) {
			getLoggerWithRequestID(c).Warnf("⚠️ 疑似路径遍历攻击（URL路径）: %s", path)
			return true
		}
	}

	// 检查URL参数（Query参数）
	for key, values := range c.Request.URL.Query() {
		for _, value := range values {
			lowerValue := strings.ToLower(value)
			for _, pattern := range dangerousPatterns {
				if strings.Contains(lowerValue, pattern) {
					getLoggerWithRequestID(c).Warnf("⚠️ 疑似路径遍历攻击（URL参数）: %s=%s", key, value)
					return true
				}
			}
		}
	}

	// 检查路径参数（如 /api/cdn/:filepath）
	for _, param := range c.Params {
		lowerParam := strings.ToLower(param.Value)
		for _, pattern := range dangerousPatterns {
			if strings.Contains(lowerParam, pattern) {
				getLoggerWithRequestID(c).Warnf("⚠️ 疑似路径遍历攻击（路径参数）: %s=%s", param.Key, param.Value)
				return true
			}
		}
	}

	return false
}

// detectXSS 检测XSS攻击
func detectXSS(c *gin.Context) bool {
	xssPatterns := []string{
		"<script", "javascript:", "onerror=", "onload=",
		"<iframe", "<object", "<embed", "eval(", "alert(",
	}

	// 检查URL参数
	for _, value := range c.Request.URL.Query() {
		for _, v := range value {
			lowerV := strings.ToLower(v)
			for _, pattern := range xssPatterns {
				if strings.Contains(lowerV, pattern) {
					getLoggerWithRequestID(c).Warnf("⚠️ 疑似XSS攻击: %s", v)
					return true
				}
			}
		}
	}

	return false
}

// ==================== 请求大小限制中间件 ====================

// RequestSizeLimitMiddleware 限制请求体大小
func RequestSizeLimitMiddleware(maxSize int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 文件上传路由跳过大小限制（在应用层验证）
		if strings.HasPrefix(c.Request.URL.Path, "/api/documents/upload") ||
		   strings.HasPrefix(c.Request.URL.Path, "/api/models/upload") ||
		   strings.HasPrefix(c.Request.URL.Path, "/api/assets/upload") {
			c.Next()
			return
		}
		
		// 读取Content-Length头
		if c.Request.ContentLength > maxSize {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{
				"error": fmt.Sprintf("请求体过大，最大允许 %d MB", maxSize/(1024*1024)),
			})
			c.Abort()
			return
		}

		// 限制实际读取大小
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxSize)

		c.Next()
	}
}

// ==================== 敏感路径保护中间件 ====================

// ProtectSensitivePathsMiddleware 保护敏感路径
func ProtectSensitivePathsMiddleware() gin.HandlerFunc {
	protectedPaths := []string{
		"/.env",        // 环境变量文件
		"/.git",        // Git目录
		"/data/",       // 数据库文件
		"/config/",     // 配置文件
		// 注意：不再保护 /static/，因为 /textures/ 需要公开访问
		// "/static/",     // 静态资源目录（应通过 /website 访问）
		"/bootstrap/",  // 启动代码
		"/build/",      // 构建文件
		"/core/",       // 核心代码
		"/database/",   // 数据库代码
		"/dev/",        // 开发文件
		"/docs/",       // 文档
		"/frontend/",   // 前端代码
		"/logger/",     // 日志代码
		// 注意：不再保护 /models/，因为模型文件需要公开访问
		// "/models/",     // 数据模型代码（已改为模型文件静态访问）
		"/middleware/", // 中间件代码
		"/scripts/",    // 脚本
		"/server/",     // 服务器代码
		"/services/",   // 服务代码
		"/temp/",       // 临时文件
		"/tmp/",        // 临时文件
		"/utils/",      // 工具代码
		"/@fs/",        // 文件系统访问路径
		"/@fs",         // 文件系统访问路径（简化版）
	}
	
	// 允许访问的路径（白名单）
	allowedPaths := []string{
		"/textures/",   // 材质文件公开访问
		"/models/",     // 模型文件公开访问
	}

	return func(c *gin.Context) {
		path := c.Request.URL.Path
		
		// 检查是否在白名单中
		for _, allowedPath := range allowedPaths {
			if strings.HasPrefix(path, allowedPath) {
				c.Next()
				return
			}
		}

		// 检查是否访问敏感路径
		for _, protectedPath := range protectedPaths {
			if strings.HasPrefix(path, protectedPath) {
				ip := getClientIP(c)
				getLoggerWithRequestID(c).Warnf("🚫 尝试访问敏感路径: IP=%s, Path=%s", ip, path)

				// 记录可疑活动
				incrementSuspiciousActivity(ip, "sensitive_path_access", c)

				c.Status(http.StatusNotFound)
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// ==================== DDoS防护中间件 ====================

var (
	connectionCount = make(map[string]int) // IP -> 并发连接数
	connMutex       sync.RWMutex
)

// DDoSProtectionMiddleware DDoS防护中间件
func DDoSProtectionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := getClientIP(c)

		// 白名单IP不限制
		if IsWhitelisted(ip) {
			c.Next()
			return
		}

		cfg := config.GetSecurityConfig()

		// 检查并发连接数
		connMutex.Lock()
		count := connectionCount[ip]
		if count >= cfg.MaxConcurrentConnections {
			connMutex.Unlock()
			getLoggerWithRequestID(c).Warnf("🚫 IP %s 超过最大并发连接数: %d", ip, count)
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "连接数过多",
				"code":  "TOO_MANY_CONNECTIONS",
			})
			c.Abort()
			return
		}

		// 增加连接计数
		connectionCount[ip] = count + 1
		connMutex.Unlock()

		// 请求完成后减少计数
		defer func() {
			connMutex.Lock()
			if connectionCount[ip] > 0 {
				connectionCount[ip]--
			}
			connMutex.Unlock()
		}()

		c.Next()
	}
}

// ConnectionRateLimitMiddleware 连接频率限制中间件
func ConnectionRateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 这里可以实现更复杂的连接频率限制
		// 暂时使用基本的速率限制
		c.Next()
	}
}

// ==================== 工具函数 ====================

// getClientIP 获取客户端真实IP
func getClientIP(c *gin.Context) string {
	// 优先级：
	// 1. X-Real-IP（Nginx代理）
	// 2. X-Forwarded-For（标准代理头，取第一个）
	// 3. RemoteAddr（直连）

	if ip := c.GetHeader("X-Real-IP"); ip != "" {
		return ip
	}

	if xff := c.GetHeader("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	ip, _, err := net.SplitHostPort(c.Request.RemoteAddr)
	if err != nil {
		return c.Request.RemoteAddr
	}

	return ip
}

// ==================== 初始化 ====================

func init() {
	// 定期清理可疑活动记录（每小时）
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()

		for range ticker.C {
			suspiciousMutex.Lock()
			// 清空所有可疑活动计数（防止内存泄漏）
			suspiciousActivityCount = make(map[string]int)
			suspiciousMutex.Unlock()
			logger.Log.Info("✅ 已清理可疑活动记录缓存")
		}
	}()

	// 定期清理过期的黑名单记录（每30分钟）
	go func() {
		ticker := time.NewTicker(30 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			ipMutex.Lock()
			now := time.Now()
			cleanedCount := 0
			for ip, expireTime := range ipBlacklist {
				if now.After(expireTime) {
					delete(ipBlacklist, ip)
					cleanedCount++
				}
			}
			ipMutex.Unlock()
			if cleanedCount > 0 {
				logger.Log.Infof("✅ 已清理 %d 条过期黑名单记录", cleanedCount)
			}
		}
	}()
}

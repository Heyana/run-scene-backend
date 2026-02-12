package audit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"go_wails_project_manager/config"
	"go_wails_project_manager/logger"
	"go_wails_project_manager/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// AuditService 审计服务
type AuditService struct {
	db          *gorm.DB
	config      *config.AuditConfig
	logBuffer   chan *models.AuditLog
	stopChan    chan struct{}
	wg          sync.WaitGroup
	mu          sync.Mutex
}

// NewAuditService 创建审计服务
func NewAuditService(db *gorm.DB, cfg *config.AuditConfig) *AuditService {
	service := &AuditService{
		db:        db,
		config:    cfg,
		logBuffer: make(chan *models.AuditLog, cfg.BufferSize),
		stopChan:  make(chan struct{}),
	}

	// 启动异步写入协程
	if cfg.AsyncWrite {
		service.wg.Add(1)
		go service.asyncWriter()
	}

	return service
}

// Log 记录审计日志
func (s *AuditService) Log(log *models.AuditLog) {
	if !s.config.Enabled {
		return
	}

	// 异步写入
	if s.config.AsyncWrite {
		select {
		case s.logBuffer <- log:
		default:
			logger.Log.Warn("审计日志缓冲区已满，丢弃日志")
		}
	} else {
		// 同步写入
		if err := s.db.Create(log).Error; err != nil {
			logger.Log.Errorf("写入审计日志失败: %v", err)
		}
	}
}

// asyncWriter 异步写入协程
func (s *AuditService) asyncWriter() {
	defer s.wg.Done()

	ticker := time.NewTicker(time.Duration(s.config.FlushInterval) * time.Second)
	defer ticker.Stop()

	batch := make([]*models.AuditLog, 0, s.config.BatchSize)

	flush := func() {
		if len(batch) == 0 {
			return
		}

		if err := s.db.Create(&batch).Error; err != nil {
			logger.Log.Errorf("批量写入审计日志失败: %v", err)
		} else {
			logger.Log.Debugf("批量写入审计日志成功: %d 条", len(batch))
		}

		batch = batch[:0]
	}

	for {
		select {
		case log := <-s.logBuffer:
			batch = append(batch, log)
			if len(batch) >= s.config.BatchSize {
				flush()
			}

		case <-ticker.C:
			flush()

		case <-s.stopChan:
			// 停止前刷新剩余日志
			flush()
			return
		}
	}
}

// Stop 停止审计服务
func (s *AuditService) Stop() {
	close(s.stopChan)
	s.wg.Wait()
	close(s.logBuffer)
}

// LogFromContext 从 Gin Context 记录审计日志
func (s *AuditService) LogFromContext(c *gin.Context, startTime time.Time, action, resource string, resourceID *uint) {
	if !s.config.Enabled {
		return
	}

	// 检查是否排除该路径
	if s.shouldExcludePath(c.Request.URL.Path) {
		return
	}

	// 提取用户信息
	userID, username := s.extractUserInfo(c)

	// 读取请求体
	requestBody := ""
	if s.config.RecordRequestBody && c.Request.Body != nil {
		bodyBytes, _ := io.ReadAll(c.Request.Body)
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes)) // 恢复 body
		
		if len(bodyBytes) > 0 {
			requestBody = string(bodyBytes)
			if len(requestBody) > s.config.MaxRequestBodySize {
				requestBody = requestBody[:s.config.MaxRequestBodySize] + "...[truncated]"
			}
			// 脱敏处理
			requestBody = s.sanitizeBody(requestBody)
		}
	}

	// 计算耗时
	duration := time.Since(startTime).Milliseconds()

	// 创建审计日志
	log := &models.AuditLog{
		UserID:      userID,
		Username:    username,
		UserIP:      s.getClientIP(c),
		Action:      action,
		Resource:    resource,
		ResourceID:  resourceID,
		Method:      c.Request.Method,
		Path:        c.Request.URL.Path,
		StatusCode:  c.Writer.Status(),
		Duration:    duration,
		RequestBody: requestBody,
		UserAgent:   c.Request.UserAgent(),
		CreatedAt:   time.Now(),
	}

	// 记录错误信息
	if c.Writer.Status() >= 400 {
		if err, exists := c.Get("error"); exists {
			log.ErrorMsg = fmt.Sprintf("%v", err)
		}
	}

	s.Log(log)
}

// shouldExcludePath 检查是否排除该路径
func (s *AuditService) shouldExcludePath(path string) bool {
	for _, excludePath := range s.config.LogExcludePaths {
		// 支持通配符匹配
		if strings.HasSuffix(excludePath, "*") {
			prefix := strings.TrimSuffix(excludePath, "*")
			if strings.HasPrefix(path, prefix) {
				return true
			}
		} else if path == excludePath {
			return true
		}
	}
	return false
}

// extractUserInfo 提取用户信息
func (s *AuditService) extractUserInfo(c *gin.Context) (*uint, string) {
	var userID *uint
	username := "anonymous"

	if id, exists := c.Get("user_id"); exists {
		if uid, ok := id.(uint); ok {
			userID = &uid
		}
	}

	if name, exists := c.Get("username"); exists {
		if uname, ok := name.(string); ok && uname != "" {
			username = uname
		}
	}

	return userID, username
}

// getClientIP 获取客户端IP
func (s *AuditService) getClientIP(c *gin.Context) string {
	if ip := c.GetHeader("X-Real-IP"); ip != "" {
		return ip
	}
	if xff := c.GetHeader("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}
	return c.ClientIP()
}

// sanitizeBody 脱敏处理
func (s *AuditService) sanitizeBody(body string) string {
	// 尝试解析为 JSON
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(body), &data); err != nil {
		// 不是 JSON，直接返回
		return body
	}

	// 脱敏敏感字段
	for _, field := range s.config.LogSensitiveFields {
		if _, exists := data[field]; exists {
			data[field] = "***REDACTED***"
		}
	}

	// 转回 JSON
	sanitized, err := json.Marshal(data)
	if err != nil {
		return body
	}

	return string(sanitized)
}

// InferActionFromMethod 从 HTTP 方法推断操作类型
func InferActionFromMethod(method, path string) string {
	switch method {
	case "POST":
		if strings.Contains(path, "upload") {
			return models.ActionUpload
		}
		if strings.Contains(path, "login") {
			return models.ActionLogin
		}
		if strings.Contains(path, "export") {
			return models.ActionExport
		}
		return models.ActionCreate
	case "PUT", "PATCH":
		if strings.Contains(path, "move") {
			return models.ActionMove
		}
		if strings.Contains(path, "rename") {
			return models.ActionRename
		}
		return models.ActionUpdate
	case "DELETE":
		return models.ActionDelete
	case "GET":
		if strings.Contains(path, "download") {
			return models.ActionDownload
		}
		return models.ActionView
	default:
		return "unknown"
	}
}

// InferResourceFromPath 从路径推断资源类型
func InferResourceFromPath(path string) string {
	if strings.Contains(path, "/documents") {
		return models.ResourceDocument
	}
	if strings.Contains(path, "/folders") {
		return models.ResourceFolder
	}
	if strings.Contains(path, "/users") {
		return models.ResourceUser
	}
	if strings.Contains(path, "/projects") {
		return models.ResourceProject
	}
	if strings.Contains(path, "/models") {
		return models.ResourceModel
	}
	if strings.Contains(path, "/textures") {
		return models.ResourceTexture
	}
	return "unknown"
}

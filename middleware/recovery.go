// Package middleware 提供HTTP中间件
package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"go_wails_project_manager/logger"

	"github.com/gin-gonic/gin"
)

// RecoveryMiddleware Panic恢复中间件
func RecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 记录堆栈信息
				stack := debug.Stack()
				logger.Log.Errorf("🔥 Panic recovered: %v\n%s", err, string(stack))

				// 返回500错误，不暴露内部细节
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"code":      500,
					"msg":       "系统错误，请稍后重试",
					"timestamp": c.GetInt64("request_time"),
				})
			}
		}()
		c.Next()
	}
}

// RecoveryWithCallback 带回调的Panic恢复中间件
func RecoveryWithCallback(callback func(c *gin.Context, err interface{}, stack []byte)) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				stack := debug.Stack()

				// 记录日志
				logger.Log.Errorf("🔥 Panic recovered: %v\n%s", err, string(stack))

				// 执行回调
				if callback != nil {
					callback(c, err, stack)
				}

				// 返回错误
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"code": 500,
					"msg":  "系统错误，请稍后重试",
				})
			}
		}()
		c.Next()
	}
}

// PanicError Panic错误信息
type PanicError struct {
	Error   string `json:"error"`
	Stack   string `json:"stack,omitempty"`
	Request string `json:"request"`
}

// RecoveryWithLog 带详细日志的Panic恢复（开发环境）
func RecoveryWithLog(isDev bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				stack := debug.Stack()

				panicErr := PanicError{
					Error:   fmt.Sprintf("%v", err),
					Request: fmt.Sprintf("%s %s", c.Request.Method, c.Request.URL.Path),
				}

				if isDev {
					panicErr.Stack = string(stack)
				}

				logger.Log.WithFields(map[string]interface{}{
					"method": c.Request.Method,
					"path":   c.Request.URL.Path,
					"ip":     c.ClientIP(),
				}).Errorf("🔥 Panic: %v\n%s", err, string(stack))

				// 开发环境返回详细错误
				if isDev {
					c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
						"code":  500,
						"msg":   fmt.Sprintf("Panic: %v", err),
						"stack": string(stack),
					})
				} else {
					c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
						"code": 500,
						"msg":  "系统错误，请稍后重试",
					})
				}
			}
		}()
		c.Next()
	}
}

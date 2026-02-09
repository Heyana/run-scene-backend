package middleware

import (
	"github.com/gin-gonic/gin"
)

// LargeFileUpload 大文件上传中间件
// 移除请求体大小限制，支持大文件上传
func LargeFileUpload() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Gin 默认没有请求体大小限制
		// 这个中间件主要用于确保没有其他中间件添加了限制
		// 并且可以在这里添加自定义的验证逻辑
		
		c.Next()
	}
}

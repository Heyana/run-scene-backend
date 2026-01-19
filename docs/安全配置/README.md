# 安全配置模块

基于 art-manager 项目安全实践，提供 Go Web 应用通用安全配置方案。

## 📋 模块概览

| 模块                                 | 优先级 | 状态 | 说明                         |
| ------------------------------------ | ------ | ---- | ---------------------------- |
| [通用性分析](./01_通用性分析.md)     | -      | ✅   | Art-manager 安全配置分析     |
| [核心安全配置](./02_核心安全配置.md) | P0     | 📋   | 速率限制、文件验证、登录安全 |
| [安全中间件](./03_安全中间件.md)     | P1     | 📋   | IP 封禁、请求限制中间件      |
| [权限系统模板](./04_权限系统模板.md) | P2     | 📋   | RBAC 权限框架                |
| [部署安全清单](./05_部署安全清单.md) | P1     | 📋   | 生产环境安全检查             |

## 🚀 快速开始

### 1. 核心安全配置 (P0)

```go
// config/security.go
type SecurityConfig struct {
    RateLimitPerSecond int
    MaxUploadSize      int64
    MaxLoginAttempts   int
}
```

### 2. 文件验证器 (P0)

```go
// utils/validator.go
func ValidateFile(file *multipart.FileHeader) error
func CheckMagicNumber(file io.Reader, mimeType string) error
```

### 3. 安全中间件 (P1)

```go
// middleware/security.go
func RateLimitMiddleware() gin.HandlerFunc
func IPBanMiddleware() gin.HandlerFunc
```

## 🔧 环境变量配置

添加到 `.env` 文件:

```env
# 安全配置
SECURITY_RATE_LIMIT=1000
SECURITY_MAX_UPLOAD_SIZE=52428800
SECURITY_MAX_LOGIN_ATTEMPTS=5
SECURITY_LOGIN_LOCK_DURATION=1800
```

## 📖 详细文档

- **分析阶段**: [通用性分析](./01_通用性分析.md)
- **实现阶段**: [核心配置](./02_核心安全配置.md) → [中间件](./03_安全中间件.md) → [权限系统](./04_权限系统模板.md)
- **部署阶段**: [安全清单](./05_部署安全清单.md)

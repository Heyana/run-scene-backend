# 账号和权限系统文档

## 文档目录

1. [01-概述.md](./01-概述.md) - 功能定位、权限模型、预设角色
2. [02-数据模型.md](./02-数据模型.md) - 数据库表结构设计
3. [03-API接口.md](./03-API接口.md) - REST API 接口定义
4. [04-Service层设计.md](./04-Service层设计.md) - 业务逻辑层设计
5. [05-中间件设计.md](./05-中间件设计.md) - 权限验证中间件
6. [06-初始化数据.md](./06-初始化数据.md) - 系统初始化脚本
7. [07-实施计划.md](./07-实施计划.md) - 开发计划和时间估算
8. [08-配置设计.md](./08-配置设计.md) - 配置文件和参数

## 快速开始

### 系统特性

- **RBAC + 资源级权限**：灵活的权限控制模型
- **自定义权限**：支持创建自定义权限和权限组
- **JWT 认证**：基于 Token 的无状态认证
- **权限缓存**：高性能权限计算
- **审计日志**：完整的操作记录
- **细粒度控制**：支持到具体资源实例的权限

### 默认账号

系统初始化后会创建默认管理员账号：

```
用户名: admin
密码: admin123456
```

**重要：生产环境请立即修改默认密码！**

### 预设角色

- **super_admin** - 超级管理员（所有权限）
- **admin** - 管理员（用户和资源管理）
- **editor** - 编辑者（资源读写）
- **viewer** - 查看者（资源只读）

### 权限格式

```
资源:操作

示例：
documents:read          # 查看文档
documents:create        # 创建文档
documents:*             # 文档所有操作
*:read                  # 所有资源的查看权限
*:*                     # 超级管理员权限
```

### 支持的资源类型

- `documents` - 文档库
- `models` - 模型库
- `assets` - 资产库
- `textures` - 贴图库
- `projects` - 项目管理
- `ai3d` - AI 3D生成
- `users` - 用户管理
- `roles` - 角色管理
- `permissions` - 权限管理

### 支持的操作类型

- `read` - 查看
- `create` - 创建
- `update` - 更新
- `delete` - 删除
- `download` - 下载
- `upload` - 上传
- `share` - 分享
- `admin` - 管理

## 开发指南

### 添加新资源的权限保护

```go
// 1. 在路由中添加权限验证
documents.PUT("/:id",
    jwtAuth.AuthMiddleware(),                    // 认证
    RequirePermission("documents", "update"),    // 权限验证
    RequireOwnership("documents"),               // 所有权验证
    documentController.Update,
)

// 2. 在 Controller 中获取用户信息
userID := middleware.GetUserID(c)
username := middleware.GetUsername(c)

// 3. 记录审计日志
auditService.Log(userID, "update", "documents", documentID)
```

### 创建自定义权限

```go
// 通过 API 创建
POST /api/permissions
{
    "code": "documents:export",
    "name": "导出文档",
    "resource": "documents",
    "action": "export",
    "description": "允许导出文档为PDF"
}
```

### 创建自定义角色

```go
// 通过 API 创建
POST /api/roles
{
    "code": "project_manager",
    "name": "项目经理",
    "description": "项目管理相关权限",
    "permission_ids": [1, 2, 3],
    "permission_group_ids": [1]
}
```

## 安全建议

### 生产环境配置

1. **修改 JWT 密钥**

   ```yaml
   jwt:
     secret_key: "使用强随机字符串"
   ```

2. **启用密码强度验证**

   ```yaml
   password:
     min_length: 8
     require_number: true
     require_uppercase: true
   ```

3. **启用登录保护**

   ```yaml
   login:
     max_fail_count: 5
     lock_duration: 30
     enable_captcha: true
   ```

4. **使用 Redis 缓存**
   ```yaml
   permission_cache:
     type: "redis"
     redis:
       host: "redis.internal"
   ```

### 权限设计原则

1. **最小权限原则**：只授予必要的权限
2. **职责分离**：不同角色有明确的职责边界
3. **定期审查**：定期检查用户权限是否合理
4. **审计日志**：记录所有敏感操作
5. **临时权限**：使用过期时间限制临时授权

## 常见问题

### Q: 如何禁用权限系统？

A: 在配置文件中设置：

```yaml
enable_auth: false
```

### Q: 如何重置管理员密码？

A: 使用数据库工具直接修改：

```sql
UPDATE users SET password = '$2a$10$...' WHERE username = 'admin';
```

### Q: 权限缓存何时失效？

A:

- 用户角色变更时自动失效
- 用户权限变更时自动失效
- 配置的过期时间到达时失效
- 可手动调用 API 清除缓存

### Q: 如何实现部门隔离？

A: 使用资源权限功能，为用户授予特定资源的访问权限。

## 相关文档

- [安全配置文档](../安全配置/README.md)
- [审计信息文档](../审计信息/README.md)
- [API 开发指南](../../开发/api-guide.md)

## 更新日志

- 2024-XX-XX: 初始版本设计完成

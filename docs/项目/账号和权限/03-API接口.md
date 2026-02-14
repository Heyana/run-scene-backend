# API 接口设计

## 认证接口

### POST /api/auth/register

注册新用户

**请求体：**

```go
type RegisterRequest struct {
    Username string `json:"username" binding:"required,min=3,max=50"`
    Password string `json:"password" binding:"required,min=6"`
    Email    string `json:"email" binding:"required,email"`
    Phone    string `json:"phone"`
    RealName string `json:"real_name"`
}
```

**响应：**

```go
type UserResponse struct {
    ID       uint   `json:"id"`
    Username string `json:"username"`
    Email    string `json:"email"`
    Token    string `json:"token"`
}
```

### POST /api/auth/login

用户登录

**请求体：**

```go
type LoginRequest struct {
    Username string `json:"username" binding:"required"`
    Password string `json:"password" binding:"required"`
}
```

**响应：**

```go
type TokenResponse struct {
    AccessToken  string `json:"access_token"`
    RefreshToken string `json:"refresh_token"`
    ExpiresIn    int64  `json:"expires_in"`
    TokenType    string `json:"token_type"`
    User         User   `json:"user"`
}
```

### POST /api/auth/refresh

刷新 Token

### POST /api/auth/logout

登出

### POST /api/auth/change-password

修改密码

**请求体：**

```go
type ChangePasswordRequest struct {
    OldPassword string `json:"old_password" binding:"required"`
    NewPassword string `json:"new_password" binding:"required,min=6"`
}
```

## 用户管理接口

### GET /api/users

获取用户列表

**查询参数：**

- `page` - 页码
- `page_size` - 每页数量
- `status` - 状态筛选
- `role` - 角色筛选
- `keyword` - 关键词搜索

### POST /api/users

创建用户（管理员）

### GET /api/users/:id

获取用户详情

### PUT /api/users/:id

更新用户信息

### DELETE /api/users/:id

删除用户

### POST /api/users/:id/disable

禁用用户

### POST /api/users/:id/enable

启用用户

### POST /api/users/:id/reset-password

重置用户密码

### GET /api/users/:id/permissions

获取用户所有权限

**响应：**

```go
type UserPermissionsResponse struct {
    Permissions      []string `json:"permissions"` // 权限代码列表
    Roles            []Role   `json:"roles"`
    PermissionGroups []PermissionGroup `json:"permission_groups"`
}
```

### POST /api/users/:id/roles

分配角色

**请求体：**

```go
type AssignRolesRequest struct {
    RoleIDs []uint `json:"role_ids" binding:"required"`
}
```

### POST /api/users/:id/permissions

直接授予权限

**请求体：**

```go
type GrantPermissionsRequest struct {
    PermissionIDs []uint `json:"permission_ids" binding:"required"`
}
```

### POST /api/users/:id/permission-groups

授予权限组

## 角色管理接口

### GET /api/roles

获取角色列表

### POST /api/roles

创建角色

**请求体：**

```go
type CreateRoleRequest struct {
    Code             string `json:"code" binding:"required"`
    Name             string `json:"name" binding:"required"`
    Description      string `json:"description"`
    PermissionIDs    []uint `json:"permission_ids"`
    PermissionGroupIDs []uint `json:"permission_group_ids"`
}
```

### GET /api/roles/:id

获取角色详情

### PUT /api/roles/:id

更新角色

### DELETE /api/roles/:id

删除角色（非系统角色）

### POST /api/roles/:id/permissions

分配权限给角色

### DELETE /api/roles/:id/permissions/:permissionId

移除角色权限

### POST /api/roles/:id/permission-groups

分配权限组给角色

## 权限管理接口

### GET /api/permissions

获取权限列表

**查询参数：**

- `resource` - 资源类型筛选
- `action` - 操作类型筛选
- `is_system` - 是否系统权限

### POST /api/permissions

创建自定义权限

**请求体：**

```go
type CreatePermissionRequest struct {
    Code        string `json:"code" binding:"required"` // documents:custom_action
    Name        string `json:"name" binding:"required"`
    Resource    string `json:"resource" binding:"required"`
    Action      string `json:"action" binding:"required"`
    Description string `json:"description"`
}
```

### GET /api/permissions/:id

获取权限详情

### PUT /api/permissions/:id

更新权限

### DELETE /api/permissions/:id

删除权限（非系统权限）

### GET /api/permissions/resources

获取所有资源类型

**响应：**

```go
[]string // ["documents", "models", "assets", ...]
```

### GET /api/permissions/actions

获取所有操作类型

## 权限组管理接口

### GET /api/permission-groups

获取权限组列表

### POST /api/permission-groups

创建权限组

**请求体：**

```go
type CreatePermissionGroupRequest struct {
    Code          string `json:"code" binding:"required"`
    Name          string `json:"name" binding:"required"`
    Description   string `json:"description"`
    PermissionIDs []uint `json:"permission_ids"`
}
```

### GET /api/permission-groups/:id

获取权限组详情

### PUT /api/permission-groups/:id

更新权限组

### DELETE /api/permission-groups/:id

删除权限组（非系统）

### POST /api/permission-groups/:id/permissions

添加权限到权限组

**请求体：**

```go
type AddPermissionsRequest struct {
    PermissionIDs []uint `json:"permission_ids" binding:"required"`
}
```

### DELETE /api/permission-groups/:id/permissions/:permissionId

从权限组移除权限

## 资源权限接口

### POST /api/resource-permissions

授予资源权限

**请求体：**

```go
type GrantResourcePermissionRequest struct {
    UserID       uint       `json:"user_id" binding:"required"`
    ResourceType string     `json:"resource_type" binding:"required"`
    ResourceID   uint       `json:"resource_id" binding:"required"`
    Permission   string     `json:"permission" binding:"required"`
    ExpiresAt    *time.Time `json:"expires_at"`
}
```

### DELETE /api/resource-permissions/:id

撤销资源权限

### GET /api/resource-permissions/user/:userId

获取用户的资源权限列表

### GET /api/resource-permissions/resource/:type/:id

获取资源的权限列表

## 权限验证接口

### POST /api/auth/check-permission

检查用户是否有某个权限

**请求体：**

```go
type CheckPermissionRequest struct {
    Resource string `json:"resource" binding:"required"`
    Action   string `json:"action" binding:"required"`
}
```

**响应：**

```go
type CheckPermissionResponse struct {
    HasPermission bool   `json:"has_permission"`
    Reason        string `json:"reason,omitempty"`
}
```

### POST /api/auth/check-resource-permission

检查用户是否有资源访问权限

**请求体：**

```go
type CheckResourcePermissionRequest struct {
    ResourceType string `json:"resource_type" binding:"required"`
    ResourceID   uint   `json:"resource_id" binding:"required"`
    Permission   string `json:"permission" binding:"required"`
}
```

## 响应格式

### 成功响应

```go
{
    "code": 200,
    "message": "success",
    "data": { ... }
}
```

### 错误响应

```go
{
    "code": 400,
    "message": "错误信息",
    "data": null
}
```

### 分页响应

```go
{
    "code": 200,
    "message": "success",
    "data": {
        "items": [...],
        "total": 100,
        "page": 1,
        "page_size": 20
    }
}
```

## 错误码

- `401` - 未认证
- `403` - 权限不足
- `404` - 资源不存在
- `409` - 资源冲突（如用户名已存在）
- `422` - 参数验证失败
- `500` - 服务器错误

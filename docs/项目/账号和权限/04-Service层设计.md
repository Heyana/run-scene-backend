# Service 层设计

## UserService 用户服务

```go
type UserService struct {
    db *gorm.DB
}

func NewUserService(db *gorm.DB) *UserService

// 用户管理
Create(req *CreateUserRequest) (*User, error)
GetByID(id uint) (*User, error)
GetByUsername(username string) (*User, error)
GetByEmail(email string) (*User, error)
Update(id uint, req *UpdateUserRequest) error
Delete(id uint) error
List(filter *UserFilter) ([]User, int64, error)

// 状态管理
Enable(id uint) error
Disable(id uint) error
Lock(id uint, duration time.Duration) error
Unlock(id uint) error

// 密码管理
ChangePassword(id uint, oldPassword, newPassword string) error
ResetPassword(id uint, newPassword string) error
ValidatePassword(user *User, password string) error

// 登录管理
RecordLogin(id uint, ip string) error
RecordLoginFail(username string) error
IsLocked(user *User) bool

// 角色和权限
AssignRoles(userID uint, roleIDs []uint) error
GrantPermissions(userID uint, permissionIDs []uint) error
GrantPermissionGroups(userID uint, groupIDs []uint) error
GetUserPermissions(userID uint) ([]string, error)
```

## RoleService 角色服务

```go
type RoleService struct {
    db *gorm.DB
}

func NewRoleService(db *gorm.DB) *RoleService

// 角色管理
Create(req *CreateRoleRequest) (*Role, error)
GetByID(id uint) (*Role, error)
GetByCode(code string) (*Role, error)
Update(id uint, req *UpdateRoleRequest) error
Delete(id uint) error
List(filter *RoleFilter) ([]Role, int64, error)

// 权限管理
AssignPermissions(roleID uint, permissionIDs []uint) error
RemovePermission(roleID uint, permissionID uint) error
AssignPermissionGroups(roleID uint, groupIDs []uint) error
GetRolePermissions(roleID uint) ([]string, error)

// 系统角色
IsSystemRole(roleID uint) bool
GetSystemRoles() ([]Role, error)
```

## PermissionService 权限服务

```go
type PermissionService struct {
    db *gorm.DB
}

func NewPermissionService(db *gorm.DB) *PermissionService

// 权限管理
Create(req *CreatePermissionRequest) (*Permission, error)
GetByID(id uint) (*Permission, error)
GetByCode(code string) (*Permission, error)
Update(id uint, req *UpdatePermissionRequest) error
Delete(id uint) error
List(filter *PermissionFilter) ([]Permission, int64, error)

// 权限查询
GetByResource(resource string) ([]Permission, error)
GetByAction(action string) ([]Permission, error)
GetResources() ([]string, error)
GetActions() ([]string, error)

// 系统权限
IsSystemPermission(id uint) bool
GetSystemPermissions() ([]Permission, error)
```

## PermissionGroupService 权限组服务

```go
type PermissionGroupService struct {
    db *gorm.DB
}

func NewPermissionGroupService(db *gorm.DB) *PermissionGroupService

// 权限组管理
Create(req *CreatePermissionGroupRequest) (*PermissionGroup, error)
GetByID(id uint) (*PermissionGroup, error)
GetByCode(code string) (*PermissionGroup, error)
Update(id uint, req *UpdatePermissionGroupRequest) error
Delete(id uint) error
List(filter *PermissionGroupFilter) ([]PermissionGroup, int64, error)

// 权限管理
AddPermissions(groupID uint, permissionIDs []uint) error
RemovePermission(groupID uint, permissionID uint) error
GetGroupPermissions(groupID uint) ([]Permission, error)

// 系统权限组
IsSystemGroup(id uint) bool
GetSystemGroups() ([]PermissionGroup, error)
```

## ResourcePermissionService 资源权限服务

```go
type ResourcePermissionService struct {
    db *gorm.DB
}

func NewResourcePermissionService(db *gorm.DB) *ResourcePermissionService

// 资源权限管理
Grant(req *GrantResourcePermissionRequest) error
Revoke(id uint) error
GetByUser(userID uint) ([]ResourcePermission, error)
GetByResource(resourceType string, resourceID uint) ([]ResourcePermission, error)

// 权限检查
HasResourcePermission(userID uint, resourceType string, resourceID uint, permission string) bool
CleanExpired() error
```

## AuthService 认证服务

```go
type AuthService struct {
    db          *gorm.DB
    jwtAuth     *middleware.JWTAuth
    userService *UserService
}

func NewAuthService(db *gorm.DB, jwtAuth *middleware.JWTAuth) *AuthService

// 认证
Register(req *RegisterRequest) (*User, string, error)
Login(username, password string) (*TokenResponse, error)
Logout(userID uint) error
RefreshToken(token string) (string, error)

// 权限验证
HasPermission(userID uint, resource, action string) bool
HasResourcePermission(userID uint, resourceType string, resourceID uint, permission string) bool
IsOwner(userID uint, resourceType string, resourceID uint) bool

// Token 管理
GenerateToken(user *User) (*TokenResponse, error)
ValidateToken(token string) (*middleware.Claims, error)
```

## PermissionCalculator 权限计算器

```go
type PermissionCalculator struct {
    db *gorm.DB
}

func NewPermissionCalculator(db *gorm.DB) *PermissionCalculator

// 权限计算
CalculateUserPermissions(userID uint) ([]string, error)
MatchPermission(userPerms []string, required string) bool
ExpandWildcard(perm string) []string

// 缓存管理
CacheUserPermissions(userID uint, perms []string) error
GetCachedPermissions(userID uint) ([]string, bool)
InvalidateCache(userID uint) error
```

## 辅助函数

```go
// 权限匹配
func MatchPermission(userPerms []string, required string) bool {
    // 精确匹配
    // 通配符匹配: documents:* 或 *:read
}

// 权限展开
func ExpandWildcard(perm string) []string {
    // documents:* -> documents:read, documents:create, ...
}

// 密码验证
func ValidatePasswordStrength(password string) error {
    // 长度、复杂度检查
}

// 用户名验证
func ValidateUsername(username string) error {
    // 格式、长度检查
}
```

## 初始化服务

```go
type ServiceContainer struct {
    UserService               *UserService
    RoleService               *RoleService
    PermissionService         *PermissionService
    PermissionGroupService    *PermissionGroupService
    ResourcePermissionService *ResourcePermissionService
    AuthService               *AuthService
    PermissionCalculator      *PermissionCalculator
}

func NewServiceContainer(db *gorm.DB, jwtAuth *middleware.JWTAuth) *ServiceContainer {
    return &ServiceContainer{
        UserService:               NewUserService(db),
        RoleService:               NewRoleService(db),
        PermissionService:         NewPermissionService(db),
        PermissionGroupService:    NewPermissionGroupService(db),
        ResourcePermissionService: NewResourcePermissionService(db),
        AuthService:               NewAuthService(db, jwtAuth),
        PermissionCalculator:      NewPermissionCalculator(db),
    }
}
```

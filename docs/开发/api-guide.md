# API 详细指南

> 返回：[核心开发规范](./规范.md)

## 1. 响应结构详解

### 基础响应结构

```go
type Response struct {
    Code      ResponseCode `json:"code"`      // 业务状态码
    Msg       string       `json:"msg"`       // 消息
    Data      interface{}  `json:"data"`      // 数据
    Timestamp int64        `json:"timestamp"` // 时间戳
}
```

### 分页响应示例

```go
func (c *Controller) GetList(ctx *gin.Context) {
    page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
    pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "20"))

    list, total, err := service.GetList(page, pageSize)
    if err != nil {
        response.InternalServerError(ctx, "查询失败")
        return
    }

    response.SuccessWithPagination(ctx, list, total, page, pageSize)
}
```

响应格式：

```json
{
    "code": 200,
    "msg": "success",
    "data": {
        "list": [...],
        "total": 100,
        "page": 1,
        "page_size": 20,
        "pages": 5
    },
    "timestamp": 1702713600
}
```

### 验证错误详情示例

```go
// 表单验证失败时返回字段错误
details := map[string]string{
    "username": "用户名已存在",
    "email":    "邮箱格式不正确",
}
response.ValidationErrorWithDetails(c, details)
```

响应格式：

```json
{
  "code": 422,
  "msg": "数据验证失败",
  "details": {
    "username": "用户名已存在",
    "email": "邮箱格式不正确"
  },
  "timestamp": 1702713600
}
```

## 2. 完整响应函数列表

| HTTP 码 | 函数                                                         | 用途         | 示例场景      |
| ------- | ------------------------------------------------------------ | ------------ | ------------- |
| 200     | `response.Success(c, data)`                                  | 成功返回数据 | 查询成功      |
| 200     | `response.SuccessWithMsg(c, msg, data)`                      | 成功带消息   | 创建/更新成功 |
| 200     | `response.SuccessWithPagination(c, list, total, page, size)` | 分页列表     | 列表查询      |
| 400     | `response.BadRequest(c, msg)`                                | 参数错误     | 参数校验失败  |
| 401     | `response.Unauthorized(c, msg)`                              | 未认证       | 未登录        |
| 403     | `response.Forbidden(c, msg)`                                 | 无权限       | 权限不足      |
| 404     | `response.NotFound(c, msg)`                                  | 资源不存在   | 记录不存在    |
| 409     | `response.Conflict(c, msg)`                                  | 数据冲突     | 重复数据      |
| 422     | `response.ValidationError(c, msg)`                           | 验证失败     | 数据验证失败  |
| 422     | `response.ValidationErrorWithDetails(c, details)`            | 验证失败详情 | 表单字段错误  |
| 429     | `response.TooManyRequests(c, msg)`                           | 请求频繁     | 限流          |
| 500     | `response.InternalServerError(c, msg)`                       | 服务器错误   | 系统异常      |

## 3. 类型安全最佳实践

### 正确的结构体定义

```go
// 请求结构体
type CreateUserRequest struct {
    Username string `json:"username" binding:"required"`
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required,min=6"`
}

// 响应结构体
type UserResponse struct {
    ID        uint   `json:"id"`
    Username  string `json:"username"`
    Email     string `json:"email"`
    CreatedAt int64  `json:"created_at"`
}

// 列表响应结构体
type UserListResponse struct {
    Users []UserResponse `json:"users"`
    Total int64          `json:"total"`
}
```

### 控制器实现模式

```go
func (uc *UserController) CreateUser(c *gin.Context) {
    var req CreateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.BadRequest(c, "参数错误: "+err.Error())
        return
    }

    user, err := uc.userService.Create(req)
    if err != nil {
        response.InternalServerError(c, "创建失败: "+err.Error())
        return
    }

    response.SuccessWithMsg(c, "创建成功", UserResponse{
        ID:        user.ID,
        Username:  user.Username,
        Email:     user.Email,
        CreatedAt: user.CreatedAt.Unix(),
    })
}
```

## 4. 错误处理模式

### 标准错误处理流程

```go
func (c *Controller) HandleRequest(ctx *gin.Context) {
    // 1. 参数绑定和验证
    var req RequestStruct
    if err := ctx.ShouldBindJSON(&req); err != nil {
        response.BadRequest(ctx, "参数错误: "+err.Error())
        return
    }

    // 2. 业务逻辑处理
    result, err := service.ProcessRequest(req)
    if err != nil {
        // 根据错误类型返回不同状态码
        switch {
        case errors.Is(err, service.ErrNotFound):
            response.NotFound(ctx, "资源不存在")
        case errors.Is(err, service.ErrDuplicate):
            response.Conflict(ctx, "数据已存在")
        case errors.Is(err, service.ErrValidation):
            response.ValidationError(ctx, err.Error())
        default:
            response.InternalServerError(ctx, "处理失败: "+err.Error())
        }
        return
    }

    // 3. 成功响应
    response.Success(ctx, result)
}
```

### 参数验证最佳实践

```go
// 使用 gin 的 binding 标签进行基础验证
type CreateItemRequest struct {
    Name        string  `json:"name" binding:"required,min=1,max=100"`
    Description string  `json:"description" binding:"max=500"`
    Price       float64 `json:"price" binding:"required,gt=0"`
    CategoryID  uint    `json:"category_id" binding:"required"`
}

// 自定义验证逻辑
func (req *CreateItemRequest) Validate() error {
    if strings.TrimSpace(req.Name) == "" {
        return fmt.Errorf("name cannot be empty")
    }

    if req.Price > 999999 {
        return fmt.Errorf("price too high")
    }

    return nil
}

// 在控制器中使用
func (c *ItemController) Create(ctx *gin.Context) {
    var req CreateItemRequest
    if err := ctx.ShouldBindJSON(&req); err != nil {
        response.BadRequest(ctx, "参数格式错误: "+err.Error())
        return
    }

    if err := req.Validate(); err != nil {
        response.ValidationError(ctx, err.Error())
        return
    }

    // 继续处理...
}
```

## 5. 日志集成模式

### 控制器日志使用

```go
type UserController struct {
    logger      *logrus.Logger
    userService *service.UserService
}

func (uc *UserController) CreateUser(c *gin.Context) {
    requestID := c.GetHeader("X-Request-ID")

    uc.logger.WithFields(logrus.Fields{
        "request_id": requestID,
        "action":     "create_user",
        "ip":         c.ClientIP(),
    }).Info("开始创建用户")

    var req CreateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        uc.logger.WithFields(logrus.Fields{
            "request_id": requestID,
            "error":      err.Error(),
        }).Warn("参数绑定失败")

        response.BadRequest(c, "参数错误: "+err.Error())
        return
    }

    user, err := uc.userService.Create(req)
    if err != nil {
        uc.logger.WithFields(logrus.Fields{
            "request_id": requestID,
            "error":      err.Error(),
            "username":   req.Username,
        }).Error("用户创建失败")

        response.InternalServerError(c, "创建失败")
        return
    }

    uc.logger.WithFields(logrus.Fields{
        "request_id": requestID,
        "user_id":    user.ID,
        "username":   user.Username,
    }).Info("用户创建成功")

    response.SuccessWithMsg(c, "创建成功", UserResponse{...})
}
```

## 6. 性能优化建议

### 分页查询优化

```go
func (c *ItemController) GetList(ctx *gin.Context) {
    // 限制分页大小，防止大量数据查询
    page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
    pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "20"))

    if pageSize > 100 {
        pageSize = 100 // 限制最大页面大小
    }
    if page < 1 {
        page = 1
    }

    // 使用缓存（如果适用）
    cacheKey := fmt.Sprintf("items:list:%d:%d", page, pageSize)
    if cached, exists := cache.Get(cacheKey); exists {
        response.Success(ctx, cached)
        return
    }

    items, total, err := c.itemService.GetPaginatedList(page, pageSize)
    if err != nil {
        response.InternalServerError(ctx, "查询失败")
        return
    }

    result := gin.H{
        "list":      items,
        "total":     total,
        "page":      page,
        "page_size": pageSize,
        "pages":     (total + int64(pageSize) - 1) / int64(pageSize),
    }

    // 设置缓存
    cache.Set(cacheKey, result, 5*time.Minute)

    response.Success(ctx, result)
}
```

### 并发安全注意事项

```go
// 使用 sync.Pool 复用对象
var requestPool = sync.Pool{
    New: func() interface{} {
        return &CreateUserRequest{}
    },
}

func (uc *UserController) CreateUser(c *gin.Context) {
    req := requestPool.Get().(*CreateUserRequest)
    defer func() {
        // 重置对象后放回池中
        *req = CreateUserRequest{}
        requestPool.Put(req)
    }()

    if err := c.ShouldBindJSON(req); err != nil {
        response.BadRequest(c, "参数错误: "+err.Error())
        return
    }

    // 继续处理...
}
```

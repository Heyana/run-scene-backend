# Swagger 文档指南

> 返回：[核心开发规范](./规范.md)

## 1. 注释位置要求

### 独立函数 vs 匿名函数

```go
// ✓ 正确：独立函数（swag 可解析）
// @Summary 获取列表
// @Router /api/items [get]
func GetList(c *gin.Context) {}

// ✗ 错误：匿名函数（swag 无法解析）
api.GET("/items", func(c *gin.Context) {})
```

### 文件结构建议

```
controllers/
├── user_controller.go      // 所有用户相关 API
├── item_controller.go      // 所有商品相关 API
└── auth_controller.go      // 所有认证相关 API
```

## 2. 类型引用规范

### 包路径要求

```go
// ✓ 正确：使用完整包路径
// @Success 200 {object} response.Response
// @Success 200 {object} response.Response{data=models.User}
// @Param req body controllers.CreateUserRequest true "用户信息"

// ✗ 错误：缺少包路径（swag 找不到类型）
// @Success 200 {object} Response
// @Success 200 {object} User
```

### 泛型类型指定

```go
// ✓ 正确：指定具体的 data 类型
// @Success 200 {object} response.Response{data=[]models.User}
// @Success 200 {object} response.Response{data=models.PaginationResult}

// ✗ 错误：未指定 data 类型
// @Success 200 {object} response.Response
```

## 3. 参数注释规范

### 路径参数

```go
// @Param id path int true "用户ID（示例: 1）"
// @Param uuid path string true "资源UUID（示例: abc-123-def）"
```

### 查询参数

```go
// @Param page query int false "页码，默认1" default(1)
// @Param limit query int false "每页数量，默认20，最大100" default(20)
// @Param keyword query string false "搜索关键词（示例: golang）"
```

### 请求体参数

```go
// @Param req body CreateUserRequest true "用户创建请求"
// @Param data body UpdateItemRequest true "商品更新数据"
```

## 4. 响应注释规范

### 成功响应

```go
// 单个对象
// @Success 200 {object} response.Response{data=UserResponse} "获取成功"

// 数组
// @Success 200 {object} response.Response{data=[]UserResponse} "列表获取成功"

// 分页
// @Success 200 {object} response.Response{data=PaginationResponse} "分页数据"

// 无数据
// @Success 200 {object} response.Response "操作成功"
```

### 错误响应

```go
// @Failure 400 {object} response.Response "参数错误"
// @Failure 401 {object} response.Response "未认证"
// @Failure 403 {object} response.Response "权限不足"
// @Failure 404 {object} response.Response "资源不存在"
// @Failure 422 {object} response.ValidationErrorResponse "验证失败"
// @Failure 500 {object} response.Response "服务器错误"
```

## 5. 完整注释模板

### 基础 CRUD 操作

```go
// CreateUser 创建用户
// @Summary 创建新用户
// @Description 创建一个新的用户账户，用户名和邮箱必须唯一
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param req body CreateUserRequest true "用户创建信息"
// @Success 200 {object} response.Response{data=UserResponse} "创建成功"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 409 {object} response.Response "用户名或邮箱已存在"
// @Failure 500 {object} response.Response "创建失败"
// @Router /api/users [post]
func CreateUser(c *gin.Context) {}

// GetUser 获取用户详情
// @Summary 根据ID获取用户信息
// @Tags 用户管理
// @Produce json
// @Param id path int true "用户ID（示例: 1）"
// @Success 200 {object} response.Response{data=UserResponse} "获取成功"
// @Failure 404 {object} response.Response "用户不存在"
// @Router /api/users/{id} [get]
func GetUser(c *gin.Context) {}

// UpdateUser 更新用户信息
// @Summary 更新指定用户的信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param id path int true "用户ID（示例: 1）"
// @Param req body UpdateUserRequest true "更新信息"
// @Success 200 {object} response.Response{data=UserResponse} "更新成功"
// @Failure 400 {object} response.Response "参数错误"
// @Failure 404 {object} response.Response "用户不存在"
// @Router /api/users/{id} [put]
func UpdateUser(c *gin.Context) {}

// DeleteUser 删除用户
// @Summary 删除指定用户
// @Tags 用户管理
// @Param id path int true "用户ID（示例: 1）"
// @Success 200 {object} response.Response "删除成功"
// @Failure 404 {object} response.Response "用户不存在"
// @Router /api/users/{id} [delete]
func DeleteUser(c *gin.Context) {}

// GetUserList 获取用户列表
// @Summary 分页获取用户列表
// @Tags 用户管理
// @Produce json
// @Param page query int false "页码，默认1" default(1)
// @Param limit query int false "每页数量，默认20" default(20)
// @Param keyword query string false "搜索关键词（示例: john）"
// @Success 200 {object} response.Response{data=UserListResponse} "获取成功"
// @Router /api/users [get]
func GetUserList(c *gin.Context) {}
```

### 文件上传操作

```go
// UploadFile 文件上传
// @Summary 上传文件到服务器
// @Tags 文件管理
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "上传的文件"
// @Param category formData string false "文件分类（示例: image）"
// @Success 200 {object} response.Response{data=FileUploadResponse} "上传成功"
// @Failure 400 {object} response.Response "文件格式不支持"
// @Failure 413 {object} response.Response "文件过大"
// @Router /api/files/upload [post]
func UploadFile(c *gin.Context) {}
```

## 6. Tag 管理

### 在 swagger.go 中定义 Tag

```go
// @title 项目管理系统 API
// @version 1.0
// @description API 文档

// @tag.name 系统
// @tag.description 系统相关操作，健康检查等

// @tag.name 用户管理
// @tag.description 用户账户管理，认证授权

// @tag.name 文件管理
// @tag.description 文件上传下载管理

// @tag.name 备份管理
// @tag.description 系统备份和恢复操作

// @tag.name 安全管理
// @tag.description 安全设置和监控
```

### Tag 使用规范

- Tag 名称使用中文，便于理解
- 每个控制器的所有方法使用统一 Tag
- Tag 按功能模块分组

## 7. 生成和验证

### 生成命令

```bash
# 生成 swagger 文档
swag init -g docs/swagger.go --output docs

# 指定输出格式
swag init -g docs/swagger.go --output docs --outputTypes go,json,yaml
```

### 常见错误解决

#### 1. LeftDelim/RightDelim 错误

**错误信息：**

```
unknown field LeftDelim in struct literal
```

**解决方案：**
删除 `docs/docs.go` 中的这两行：

```go
// 删除这两行
LeftDelim:  "{{",
RightDelim: "}}",
```

#### 2. 类型找不到错误

**错误信息：**

```
cannot find type definition: User
```

**解决方案：**
使用完整包路径：

```go
// 错误
// @Success 200 {object} User

// 正确
// @Success 200 {object} models.User
```

#### 3. 匿名函数注释无效

**错误信息：**
注释没有生成到文档中

**解决方案：**
将匿名函数改为独立函数：

```go
// 错误
api.GET("/users", func(c *gin.Context) {})

// 正确
api.GET("/users", controller.GetUsers)
```

## 8. 最佳实践

### 注释编写原则

1. **简洁明了**：Summary 一行说清楚功能
2. **详细描述**：Description 补充必要的业务说明
3. **示例完整**：参数描述包含示例值
4. **错误全面**：列出所有可能的错误状态码

### 维护建议

1. **统一风格**：团队使用统一的注释模板
2. **及时更新**：代码变更后立即更新注释
3. **定期检查**：定期运行 `swag init` 检查注释正确性
4. **版本管理**：将生成的文档纳入版本控制

### 性能考虑

1. **按需生成**：开发环境才启用 swagger 文档
2. **缓存设置**：生产环境可考虑缓存生成的文档
3. **文档大小**：避免在注释中包含大量示例数据

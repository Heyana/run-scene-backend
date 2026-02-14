# 需求管理平台测试文档

## 测试结构

```
tests/
├── setup_test.go                  # 测试环境初始化
├── helpers.go                     # 测试辅助函数（完全类型安全）
├── requirement_company_test.go    # 公司管理测试
├── requirement_project_test.go    # 项目管理测试
├── requirement_mission_list_test.go # 任务列表测试
├── requirement_mission_test.go    # 任务管理测试
├── results/                       # 测试结果输出目录
└── README.md                      # 本文档
```

## 测试环境

- 使用独立的测试数据库：`data/test.db`
- 测试完成后自动清理数据库
- 每个测试用例独立运行，互不影响

## 运行测试

### 使用 npm 脚本（推荐）

```bash
# 运行所有测试
npm test

# 运行特定模块测试
npm run test:company
npm run test:project
npm run test:mission

# 查看测试覆盖率
npm run test:coverage

# 生成HTML覆盖率报告
npm run test:coverage-html
```

### 使用 go test 命令

```bash
# 运行所有测试
go test ./tests/... -v

# 运行特定测试文件
go test ./tests -run TestCompanyCreate -v

# 查看测试覆盖率
go test ./tests/... -cover

# 生成覆盖率报告
go test ./tests/... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## 测试特点

1. **完全类型安全** - 直接使用真实的 models 和 controllers 类型，无 `interface{}`
2. **泛型辅助函数** - 提供类型推断，避免类型断言
3. **独立测试数据库** - 使用 `data/test.db`，不影响开发数据
4. **清晰的测试结构** - 每个测试用例都有明确的目的和断言
5. **自动结果输出** - 测试结果自动保存到 `results/` 目录

## 测试覆盖

### 公司管理测试 (requirement_company_test.go)

- ✅ 创建公司（成功、缺少必填字段、未认证访问）
- ✅ 获取公司列表（分页、搜索）
- ✅ 更新公司信息
- ✅ 成员管理（添加、列表、移除）

### 项目管理测试 (requirement_project_test.go)

- ✅ 创建项目（成功、缺少必填字段）
- ✅ 获取项目列表（分页、按公司筛选）
- ✅ 更新项目信息
- ✅ 成员管理（添加、列表、移除）
- ✅ 项目统计

### 任务列表测试 (requirement_mission_list_test.go)

- ✅ 创建任务列表（成功、缺少必填字段）
- ✅ 获取任务列表
- ✅ 更新任务列表
- ✅ 删除任务列表

### 任务管理测试 (requirement_mission_test.go)

- ✅ 创建任务（成功、缺少必填字段）
- ✅ 获取任务列表（分页、筛选）
- ✅ 更新任务信息
- ✅ 更新任务状态
- ✅ 添加评论
- ✅ 删除任务

## 常见问题与解决方案

### 1. 业务状态码 vs HTTP 状态码

**问题**: 系统使用业务状态码（`response.code`）而不是 HTTP 状态码。所有响应统一返回 HTTP 200。

**表现**:

- 测试期望 HTTP 401，但实际返回 HTTP 200
- 错误信息在响应体的 `response.code` 字段中

**解决方案**:

```go
// ❌ 错误做法
assert.Equal(t, http.StatusUnauthorized, w.Code)

// ✅ 正确做法
AssertError(t, w, http.StatusOK, 401) // HTTP 200，业务状态码 401
```

### 2. 配置未正确加载

**问题**: 测试环境中 `config.RequirementCfg` 配置未完全初始化，导致默认值为空。

**表现**:

- 任务状态流转失败："不允许的状态流转"
- 任务创建时状态为空字符串
- 优先级为空

**原因**: `config.LoadRequirementConfig()` 只加载了配置文件，但嵌套结构的默认值可能为空。

**解决方案**: 在 `setup_test.go` 中确保配置有默认值：

```go
// 确保任务配置有默认值
if config.RequirementCfg.Requirement.Mission.DefaultStatus == "" {
    config.RequirementCfg.Requirement.Mission.DefaultStatus = "todo"
}
if config.RequirementCfg.Requirement.Mission.DefaultPriority == "" {
    config.RequirementCfg.Requirement.Mission.DefaultPriority = "P2"
}
```

### 3. 项目成员必须先是公司成员

**问题**: 添加项目成员时报错 "record not found"。

**原因**: 业务逻辑要求用户必须先是公司成员才能成为项目成员。这是为了确保权限层级的一致性。

**解决方案**: 测试中先添加用户到公司，再添加到项目：

```go
// 1. 先添加到公司
addCompanyMemberReq := requirementControllers.AddMemberRequest{
    UserID: member.ID,
    Role:   "member",
}
companyMemberURL := fmt.Sprintf("/api/requirement/companies/%d/members", company.ID)
w := MakeRequestWithBody(t, "POST", companyMemberURL, addCompanyMemberReq, ownerToken)
AssertSuccess(t, w, http.StatusOK)

// 2. 再添加到项目
addProjectMemberReq := requirementControllers.AddProjectMemberRequest{
    UserID: member.ID,
    Role:   "developer",
}
projectMemberURL := fmt.Sprintf("/api/requirement/projects/%d/members", project.ID)
w = MakeRequestWithBody(t, "POST", projectMemberURL, addProjectMemberReq, ownerToken)
AssertSuccess(t, w, http.StatusOK)
```

### 4. 类型比较错误

**问题**: `assert.Equal(t, 200, resp.Code)` 失败，提示类型不匹配。

**原因**: `resp.Code` 是 `response.ResponseCode` 类型（自定义类型），而 `200` 是 `int` 类型。

**解决方案**: 使用正确的类型常量：

```go
// ❌ 错误做法
assert.Equal(t, 200, resp.Code)

// ✅ 正确做法
assert.Equal(t, response.CodeSuccess, resp.Code)
```

### 5. JWT 中间件未生效

**问题**: 未认证请求返回 200 而不是 401。

**原因**: 测试路由器没有正确配置 JWT 中间件。

**解决方案**: 确保测试路由器正确传递 `TestJWT`：

```go
func setupCompanyRouter() *gin.Engine {
    router := gin.New()
    router.Use(gin.Recovery())
    api.SetupRequirementRoutes(router, TestJWT) // 传递 TestJWT
    return router
}
```

## 测试结果输出

测试结果会自动保存到 `results/` 目录，包含：

- 测试时间戳
- 测试状态（PASS/FAIL）
- 失败原因
- 详细的测试输出

文件命名格式：`test_result_YYYYMMDD_HHMMSS.json`

查看最新测试结果：

```bash
# Windows
type tests\results\test_result_*.json | Select-Object -Last 1

# Linux/Mac
cat tests/results/test_result_*.json | tail -1
```

## 测试最佳实践

1. **独立性**：每个测试用例独立运行，使用 `CleanupTestData()` 清理数据
2. **可读性**：使用描述性的测试名称和子测试
3. **完整性**：测试正常流程和异常流程
4. **断言**：使用 `testify/assert` 进行清晰的断言
5. **辅助函数**：复用 `helpers.go` 中的辅助函数
6. **类型安全**：直接使用真实类型，避免 `interface{}`

## 添加新测试

1. 创建新的测试文件：`requirement_xxx_test.go`
2. 使用 `setupXxxRouter()` 初始化路由
3. 使用 `CleanupTestData()` 清理数据
4. 使用辅助函数简化测试代码
5. 添加充分的测试用例覆盖

## 依赖

```bash
go get github.com/stretchr/testify/assert
go get github.com/gin-gonic/gin
```

## 注意事项

- 测试数据库会在测试结束后自动删除
- 确保测试环境配置正确（configs/requirement.yaml）
- 测试用户密码统一使用 `password123`
- 所有响应统一返回 HTTP 200，错误通过业务状态码区分
- 项目成员必须先是公司成员

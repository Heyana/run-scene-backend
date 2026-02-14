# 需求管理平台测试文档

## 测试结构

```
tests/
├── setup_test.go              # 测试环境初始化
├── helpers.go                 # 测试辅助函数
├── requirement_company_test.go    # 公司管理测试
├── requirement_project_test.go    # 项目管理测试（待创建）
├── requirement_mission_test.go    # 任务管理测试（待创建）
└── README.md                  # 本文档
```

## 测试环境

- 使用独立的测试数据库：`data/test.db`
- 测试完成后自动清理数据库
- 每个测试用例独立运行，互不影响

## 运行测试

### 运行所有测试

```bash
go test ./tests/... -v
```

### 运行特定测试文件

```bash
go test ./tests/requirement_company_test.go -v
```

### 运行特定测试用例

```bash
go test ./tests -run TestCompanyCreate -v
```

### 查看测试覆盖率

```bash
go test ./tests/... -cover
```

### 生成覆盖率报告

```bash
go test ./tests/... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## 测试用例说明

### 公司管理测试 (requirement_company_test.go)

#### TestCompanyCreate

- ✅ 成功创建公司
- ✅ 缺少必填字段
- ✅ 未认证访问

#### TestCompanyList

- ✅ 获取公司列表
- ✅ 搜索公司

#### TestCompanyUpdate

- ✅ 成功更新公司

#### TestCompanyMembers

- ✅ 添加成员
- ✅ 获取成员列表
- ✅ 移除成员

## 测试最佳实践

1. **独立性**：每个测试用例独立运行，使用 `CleanupTestData()` 清理数据
2. **可读性**：使用描述性的测试名称和子测试
3. **完整性**：测试正常流程和异常流程
4. **断言**：使用 `testify/assert` 进行清晰的断言
5. **辅助函数**：复用 `helpers.go` 中的辅助函数

## 添加新测试

1. 创建新的测试文件：`requirement_xxx_test.go`
2. 使用 `setupXxxRouter()` 初始化路由
3. 使用 `CleanupTestData()` 清理数据
4. 使用辅助函数简化测试代码
5. 添加充分的测试用例覆盖

## 依赖

```bash
go get github.com/stretchr/testify/assert
```

## 注意事项

- 测试数据库会在测试结束后自动删除
- 确保测试环境配置正确（config.yaml）
- 测试用户密码统一使用 `password123`
- 测试用户邮箱格式：`{username}@test.com`

# TypeScript 类型生成指南

> 返回：[核心开发规范](./规范.md)

## 1. OpenAPI Generator 推荐方案

### 安装和基础使用

```bash
# 全局安装
npm install @openapitools/openapi-generator-cli -g

# 基础生成命令
openapi-generator-cli generate \
  -i docs/swagger.json \
  -g typescript-axios \
  -o frontend/src/api \
  --additional-properties="supportsES6=true,npmName=@yourproject/api-client"
```

### 生成配置选项

```bash
# 完整配置示例
openapi-generator-cli generate \
  -i docs/swagger.json \
  -g typescript-axios \
  -o frontend/src/api \
  --additional-properties="supportsES6=true,npmName=@yourproject/api-client,withSeparateModelsAndApi=true,apiPackage=api,modelPackage=models"
```

**配置参数说明：**

- `supportsES6=true`: 使用 ES6 语法
- `npmName`: 生成的包名
- `withSeparateModelsAndApi=true`: 分离模型和 API
- `apiPackage=api`: API 文件夹名
- `modelPackage=models`: 模型文件夹名

## 2. 生成的文件结构

### 标准输出结构

```
frontend/src/api/
├── models/
│   ├── Response.ts              # 通用响应类型
│   ├── UserResponse.ts          # 用户响应类型
│   ├── CreateUserRequest.ts     # 创建用户请求类型
│   └── index.ts                 # 模型导出
├── api/
│   ├── UserApi.ts               # 用户 API 类
│   ├── BackupApi.ts             # 备份 API 类
│   └── index.ts                 # API 导出
├── base.ts                      # 基础配置
├── common.ts                    # 通用类型
├── configuration.ts             # 配置类型
└── index.ts                     # 总导出
```

### 生成的类型示例

```typescript
// models/Response.ts
export interface Response {
  code: number;
  msg: string;
  data?: any;
  timestamp: number;
}

// models/UserResponse.ts
export interface UserResponse {
  id: number;
  username: string;
  email: string;
  created_at: number;
}

// models/CreateUserRequest.ts
export interface CreateUserRequest {
  username: string;
  email: string;
  password: string;
}
```

## 3. API 客户端使用

### 基础使用方式

```typescript
import { UserApi, Configuration, CreateUserRequest } from "@/api";

// 配置 API 客户端
const configuration = new Configuration({
  basePath: "http://localhost:8080",
  // 可选：添加认证
  accessToken: "your-token-here",
});

const userApi = new UserApi(configuration);

// 使用示例
async function createUser() {
  try {
    const request: CreateUserRequest = {
      username: "john_doe",
      email: "john@example.com",
      password: "password123",
    };

    const response = await userApi.createUser(request);
    console.log("创建成功:", response.data);
  } catch (error) {
    console.error("创建失败:", error);
  }
}
```

### 拦截器配置

```typescript
import axios, { AxiosResponse } from "axios";
import { Configuration } from "@/api";

// 创建 axios 实例
const axiosInstance = axios.create({
  baseURL: "http://localhost:8080",
  timeout: 10000,
});

// 请求拦截器
axiosInstance.interceptors.request.use(
  (config) => {
    // 添加认证 token
    const token = localStorage.getItem("access_token");
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => Promise.reject(error)
);

// 响应拦截器
axiosInstance.interceptors.response.use(
  (response: AxiosResponse) => {
    // 统一处理响应数据
    if (response.data.code !== 200) {
      throw new Error(response.data.msg);
    }
    return response;
  },
  (error) => {
    // 统一错误处理
    if (error.response?.status === 401) {
      // 处理未认证
      localStorage.removeItem("access_token");
      window.location.href = "/login";
    }
    return Promise.reject(error);
  }
);

// 配置 API 客户端使用自定义 axios 实例
const configuration = new Configuration({
  basePath: "http://localhost:8080",
});
```

## 4. 自动化脚本

### 生成脚本 `scripts/generate-types.ts`

```typescript
import { execSync } from "child_process";
import * as fs from "fs";
import * as path from "path";

interface GenerateConfig {
  swaggerPath: string;
  outputPath: string;
  packageName: string;
}

const config: GenerateConfig = {
  swaggerPath: "docs/swagger.json",
  outputPath: "frontend/src/api",
  packageName: "@yourproject/api-client",
};

async function generateTypes() {
  try {
    console.log("🔄 检查 swagger.json 是否存在...");
    if (!fs.existsSync(config.swaggerPath)) {
      console.log("📝 生成 Swagger 文档...");
      execSync("swag init -g docs/swagger.go", { stdio: "inherit" });
    }

    console.log("🧹 清理旧的类型文件...");
    if (fs.existsSync(config.outputPath)) {
      fs.rmSync(config.outputPath, { recursive: true });
    }

    console.log("⚡ 生成 TypeScript 类型...");
    const generateCommand = `
            openapi-generator-cli generate \\
            -i ${config.swaggerPath} \\
            -g typescript-axios \\
            -o ${config.outputPath} \\
            --additional-properties="supportsES6=true,npmName=${config.packageName},withSeparateModelsAndApi=true"
        `
      .replace(/\s+/g, " ")
      .trim();

    execSync(generateCommand, { stdio: "inherit" });

    console.log("✨ 类型生成完成！");
    console.log(`📁 输出路径: ${config.outputPath}`);

    // 生成使用示例
    generateUsageExample();
  } catch (error) {
    console.error("❌ 生成失败:", error);
    process.exit(1);
  }
}

function generateUsageExample() {
  const examplePath = path.join(config.outputPath, "example.ts");
  const exampleContent = `
// API 使用示例
import { UserApi, BackupApi, Configuration } from './index';

// 配置
const config = new Configuration({
    basePath: process.env.REACT_APP_API_BASE_URL || 'http://localhost:8080'
});

// 创建 API 实例
export const userApi = new UserApi(config);
export const backupApi = new BackupApi(config);

// 使用示例
async function example() {
    try {
        // 获取用户列表
        const users = await userApi.getUserList();
        console.log(users.data);
        
        // 获取备份状态
        const backupStatus = await backupApi.getBackupStatus();
        console.log(backupStatus.data);
    } catch (error) {
        console.error('API 调用失败:', error);
    }
}
`;

  fs.writeFileSync(examplePath, exampleContent);
  console.log(`📄 使用示例已生成: ${examplePath}`);
}

// 运行生成
if (require.main === module) {
  generateTypes();
}
```

### package.json 脚本配置

```json
{
  "scripts": {
    "generate:types": "ts-node scripts/generate-types.ts",
    "generate:swagger": "cd ../backend && swag init -g docs/swagger.go",
    "dev:with-types": "npm run generate:types && npm run dev",
    "build:with-types": "npm run generate:types && npm run build"
  },
  "devDependencies": {
    "@openapitools/openapi-generator-cli": "^2.7.0",
    "ts-node": "^10.9.1"
  }
}
```

## 5. 其他工具选择

### 工具对比表

| 工具                             | 优点                             | 缺点                             | 推荐度     |
| -------------------------------- | -------------------------------- | -------------------------------- | ---------- |
| **OpenAPI Generator**            | 功能完整，社区活跃，支持多种语言 | 配置复杂，生成文件较多           | ⭐⭐⭐⭐⭐ |
| **swagger-codegen**              | 官方支持，稳定性好               | 更新较慢，自定义有限             | ⭐⭐⭐⭐   |
| **tygo**                         | 直接从 Go 生成，简单快速         | 仅支持 struct，不支持 API 客户端 | ⭐⭐⭐     |
| **typescriptify-golang-structs** | 轻量级                           | 功能有限，不够灵活               | ⭐⭐       |

### tygo 使用示例（备选方案）

```bash
# 安装 tygo
go install github.com/gzuidhof/tygo@latest

# 配置文件 tygo.yaml
packages:
  - path: "./models"
    type_mappings:
      time.Time: "string"
      null.String: "string | null"

# 生成命令
tygo generate
```

## 6. 集成到开发流程

### Git Hooks 集成

```bash
# .husky/pre-commit
#!/bin/sh
. "$(dirname "$0")/_/husky.sh"

# 检查 swagger.json 是否更新
if git diff --cached --name-only | grep -q "docs/swagger.json"; then
  echo "🔄 检测到 swagger.json 变更，重新生成 TypeScript 类型..."
  npm run generate:types
  git add frontend/src/api/
fi
```

### CI/CD 集成

```yaml
# .github/workflows/generate-types.yml
name: Generate TypeScript Types
on:
  push:
    paths:
      - "docs/swagger.json"
      - "controllers/**/*.go"

jobs:
  generate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Setup Go
        uses: actions/setup-go@v3
        with:
          go-version: "1.21"

      - name: Generate Swagger
        run: |
          go install github.com/swaggo/swag/cmd/swag@latest
          swag init -g docs/swagger.go

      - name: Setup Node.js
        uses: actions/setup-node@v3
        with:
          node-version: "18"

      - name: Generate TypeScript Types
        run: |
          npm install -g @openapitools/openapi-generator-cli
          npm run generate:types

      - name: Commit Generated Files
        run: |
          git config --local user.email "action@github.com"
          git config --local user.name "GitHub Action"
          git add frontend/src/api/
          git diff --staged --quiet || git commit -m "chore: update generated TypeScript types"
          git push
```

## 7. 最佳实践

### 版本管理

1. **生成文件纳入版本控制**：确保团队使用相同的类型定义
2. **自动化生成**：在 CI/CD 中自动生成和更新
3. **变更检测**：当 API 变更时自动重新生成

### 错误处理

```typescript
// 统一错误处理类型
export interface ApiError {
  code: number;
  message: string;
  details?: Record<string, string>;
}

// 错误处理函数
export function handleApiError(error: any): ApiError {
  if (error.response?.data) {
    return {
      code: error.response.data.code || error.response.status,
      message: error.response.data.msg || error.message,
      details: error.response.data.details,
    };
  }

  return {
    code: 500,
    message: error.message || "未知错误",
  };
}
```

### 类型安全使用

```typescript
// 使用生成的类型确保类型安全
import { CreateUserRequest, UserResponse } from "@/api/models";

interface UserForm {
  username: string;
  email: string;
  password: string;
  confirmPassword: string; // 前端特有字段
}

// 转换函数
function formToRequest(form: UserForm): CreateUserRequest {
  return {
    username: form.username,
    email: form.email,
    password: form.password,
    // 自动排除 confirmPassword
  };
}

// 使用
async function submitForm(form: UserForm) {
  const request = formToRequest(form);
  const response = await userApi.createUser(request);
  return response.data; // 类型为 UserResponse
}
```

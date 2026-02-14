# 需求管理平台 - 项目文档

## 文档目录

1. [需求分析](./01-需求分析.md) - 业务目标、核心功能、用户角色
2. [数据模型](./02-数据模型.md) - 数据库表设计、关系说明、索引建议
3. [API设计](./03-API设计.md) - RESTful API接口定义、请求响应格式
4. [后端架构](./04-后端架构.md) - 目录结构、核心类设计、权限控制
5. [前端架构](./05-前端架构.md) - 组件设计、路由配置、状态管理
6. [实施计划](./06-实施计划.md) - 开发阶段、时间安排、里程碑
7. [UI设计说明](./07-UI设计说明.md) - 设计原则、页面布局、交互设计

## 项目概述

需求管理平台是一个多租户的协作工具，支持团队进行项目需求的全流程管理，从需求创建、分配、跟踪到交付。

## 核心特性

- 🏢 多租户架构：公司/团队独立管理
- 📋 需求管理：创建、编辑、状态流转
- 📊 看板视图：可视化需求状态
- 💬 协作功能：评论、@提及、附件
- 📈 统计报表：项目进度、燃尽图
- 🔐 权限控制：细粒度权限管理
- 🔔 实时通知：WebSocket推送

## 技术栈

### 后端

- Go 1.21+
- Gin Web Framework
- GORM
- PostgreSQL / MySQL
- Redis
- WebSocket

### 前端

- Vue 3
- TypeScript
- Ant Design Vue
- Pinia
- Vue Router
- ECharts

## 快速开始

### 环境要求

- Go 1.21+
- Node.js 18+
- PostgreSQL 14+ / MySQL 8+
- Redis 6+

### 后端启动

```bash
cd run-scene-backend
go mod download
go run main.go
```

### 前端启动

```bash
cd run-scene-backend/frontend
npm install
npm run dev
```

## 开发规范

### 代码风格

- Go: 遵循 Go 官方规范
- TypeScript: 使用 ESLint + Prettier
- 提交信息: 遵循 Conventional Commits

### 分支管理

- main: 生产环境
- develop: 开发环境
- feature/\*: 功能分支
- bugfix/\*: 修复分支

### 测试要求

- 单元测试覆盖率 > 80%
- 集成测试覆盖核心流程
- E2E测试覆盖关键场景

## 部署说明

### Docker部署

```bash
docker-compose up -d
```

### 手动部署

1. 编译后端：`go build -o app`
2. 构建前端：`npm run build`
3. 配置Nginx反向代理
4. 启动服务

## 维护与支持

### 日志

- 应用日志：`logs/app.log`
- 错误日志：`logs/error.log`
- 访问日志：`logs/access.log`

### 监控

- 健康检查：`/api/health`
- 性能指标：Prometheus + Grafana

### 备份

- 数据库：每日自动备份
- 附件文件：定期同步到对象存储

## 路线图

### v1.0（当前版本）

- ✅ 基础需求管理
- ✅ 看板视图
- ✅ 评论与附件
- ✅ 基础统计

### v1.1（计划中）

- ⏳ 甘特图
- ⏳ 自定义字段
- ⏳ 工作流配置
- ⏳ 邮件通知

### v2.0（未来）

- 📅 时间跟踪
- 📅 自动化规则
- 📅 API开放平台
- 📅 移动端App

## 贡献指南

欢迎提交Issue和Pull Request！

1. Fork本仓库
2. 创建特性分支
3. 提交代码
4. 创建Pull Request

## 许可证

MIT License

## 联系方式

- 项目地址：http://192.168.3.8:3010
- 文档地址：./docs/项目/需求管理平台/
- 问题反馈：提交Issue

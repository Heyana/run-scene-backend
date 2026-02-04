# 混元3D接入 - 项目文档

## 文档导航

本文档集描述了腾讯混元3D API接入的完整方案，包括架构设计、实施步骤和注意事项。

### 文档列表

1. **[01\_概述.md](./01_概述.md)**
   - 项目目标和功能范围
   - 技术方案概览
   - 数据流程说明
   - API限制说明

2. **[02\_数据模型.md](./02_数据模型.md)**
   - 数据库表结构设计
   - 配置结构定义
   - 索引和关联关系

3. **[03\_架构设计.md](./03_架构设计.md)**
   - 目录结构规划
   - 服务层设计
   - 控制器设计
   - 核心流程说明

4. **[04_API接口.md](./04_API接口.md)**
   - 路由定义
   - 接口详细说明
   - 请求响应示例
   - 错误码说明

5. **[05\_配置管理.md](./05_配置管理.md)**
   - 配置文件结构
   - 环境变量支持
   - 配置加载和验证
   - 安全性考虑

6. **[06\_实施步骤.md](./06_实施步骤.md)**
   - 阶段划分
   - 详细任务清单
   - 依赖关系
   - 验收标准

7. **[07\_注意事项.md](./07_注意事项.md)**
   - API限制
   - 安全性要点
   - 性能优化
   - 常见问题

## 快速开始

### 1. 配置API密钥

在 `config.yaml` 中添加：

```yaml
hunyuan:
  secret_id: "你的SecretId"
  secret_key: "你的SecretKey"
  region: "ap-guangzhou"
```

### 2. 运行数据库迁移

```bash
# GORM会自动创建表
go run main.go
```

### 3. 测试API

```bash
# 提交文生3D任务
curl -X POST http://localhost:23357/api/hunyuan/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "inputType": "text",
    "prompt": "一只可爱的小猫"
  }'

# 查询任务状态
curl http://localhost:23357/api/hunyuan/tasks/1
```

## 核心功能

### 文生3D

根据文本描述生成3D模型

### 图生3D

根据单张或多视角图片生成3D模型

### 任务管理

- 提交任务
- 查询状态
- 轮询直到完成
- 取消任务
- 重试失败任务

### 自动保存

生成的模型自动保存到模型库，支持本地和NAS存储

## 技术栈

- **后端框架**：Gin
- **ORM**：GORM
- **数据库**：SQLite/MySQL
- **API客户端**：腾讯云API v3
- **文件存储**：本地/NAS

## 项目结构

### 后端

```
models/hunyuan/
├── task.go            # 任务模型
└── config.go          # 配置模型

services/hunyuan/
├── client.go          # 腾讯云API客户端
├── task_service.go    # 任务管理
├── config_service.go  # 配置管理
└── storage_service.go # 文件存储（支持本地/NAS）

controllers/
└── hunyuan_controller.go  # API控制器
```

### 前端

```
src/views/editor/components/NetTextureLibs/
├── NetHunyuanLibs.tsx     # 混元3D主组件
└── NetHunyuan/            # 工具和子组件
    ├── types.ts           # 类型定义
    ├── api.ts             # API调用
    ├── HunyuanTaskList.tsx
    ├── HunyuanSubmitForm.tsx
    ├── HunyuanTaskCard.tsx
    ├── HunyuanConfig.tsx
    └── utils.ts
```

## 开发进度

- [ ] 阶段1：基础设施
- [ ] 阶段2：核心服务
- [ ] 阶段3：API接口
- [ ] 阶段4：后台任务
- [ ] 阶段5：测试与优化

详见 [06\_实施步骤.md](./06_实施步骤.md)

## 相关链接

- [腾讯混元3D官方文档](https://cloud.tencent.com/document/api/1804/123447)
- [腾讯云API签名v3](https://cloud.tencent.com/document/api/1804/120833)
- [模型库文档](../模型库/)

## 联系方式

如有问题，请查看 [07\_注意事项.md](./07_注意事项.md) 中的常见问题部分。

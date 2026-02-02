# 模型库管理系统设计文档

## 概述

3D模型资源管理系统，支持本地上传模型文件和预览图，提供查询、分类和使用统计接口。

**核心流程**：前端处理 → 上传文件 → 存储元数据 → 查询使用

**支持格式**：GLB、GLT（私有格式）

---

## 文档结构

- [数据库设计](./数据库设计.md) - 表结构和字段定义
- [服务架构](./服务架构.md) - 服务层设计和核心函数
- [API接口](./API接口.md) - RESTful接口定义
- [实现要点](./实现要点.md) - 关键流程和优化策略

---

## 核心特性

### 与贴图库的差异

| 特性     | 贴图库       | 模型库     |
| -------- | ------------ | ---------- |
| 数据来源 | API同步      | 本地上传   |
| 文件类型 | 2D图片       | 3D模型文件 |
| 格式支持 | JPG/PNG/WebP | GLB/GLT    |
| 处理方式 | 后端转码     | 前端处理   |
| 预览图   | 后端生成     | 前端上传   |
| 元数据   | 后端提取     | 前端提供   |

### 模型库核心功能

- **文件上传**：模型文件 + 预览图一起上传
- **前端处理**：所有数据在前端处理完成
- **简单存储**：后端只负责存储和查询
- **分类管理**：支持自定义分类和标签
- **使用统计**：记录使用次数

---

## 快速开始

### 初始化

```go
// 1. 数据库迁移
db.AutoMigrate(&Model{}, &ModelTag{}, &ModelMetrics{})

// 2. 创建存储目录
os.MkdirAll("static/models", 0755)

// 3. 初始化上传服务
uploadService := NewModelUploadService(db, config)
```
### 上传模型

```bash
# 单文件上传（模型 + 预览图）
POST /api/models/upload
Content-Type: multipart/form-data
{
  model: <model.glb>,
  thumbnail: <preview.webp>,
  name: "Wooden Chair",
  type: "glb",
  category: "furniture",
  tags: "chair,wood,modern"
}
```

### 查询模型

```bash
# 分页查询
GET /api/models?page=1&pageSize=20&category=furniture

# 获取详情
GET /api/models/:id
```

---

## 配置示例

```yaml
model:
  storage_dir: "static/models"
  max_file_size: 104857600 # 100MB
  max_thumbnail_size: 5242880 # 5MB
  allowed_types: ["glb", "lt", "ltc"]
```

---

## 技术栈

- **数据库**：GORM + MySQL/PostgreSQL
- **文件存储**：本地文件系统
- **并发控制**：Goroutine Pool + Channel

---

## 后续规划

- [ ] 在线预览（Three.js）
- [ ] 智能推荐算法（基于使用统计）
- [ ] 模型版本管理
- [ ] 批量编辑工具
- [ ] CDN加速

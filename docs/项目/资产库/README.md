# 资产库管理系统设计文档

## 概述

通用资源管理系统，支持贴图、环境图、视频等多种资源类型的上传、管理和查询。

**核心流程**：前端处理 → 上传文件 → 存储元数据 → 查询使用

**支持类型**：图片（JPG/PNG/WebP）、视频（MP4/WebM）

---

## 文档结构

- [数据库设计](./数据库设计.md) - 表结构和字段定义
- [服务架构](./服务架构.md) - 服务层设计和核心函数
- [API接口](./API接口.md) - RESTful接口定义
- [实现要点](./实现要点.md) - 关键流程和优化策略

---

## 核心特性

### 与模型库的差异

| 特性     | 模型库   | 资产库                |
| -------- | -------- | --------------------- |
| 资源类型 | 3D模型   | 图片/视频             |
| 格式支持 | GLB/GLT  | JPG/PNG/WebP/MP4/WebM |
| 预览方式 | 3D预览   | 图片/视频播放器       |
| 元数据   | 前端提供 | 前端提供 + 后端补充   |
| 特殊处理 | 无       | 视频截帧生成预览图    |

### 资产库核心功能

- **多类型支持**：统一接口管理图片和视频
- **智能预览**：根据类型自动生成预览图
- **元数据提取**：图片尺寸、视频时长等
- **分类管理**：支持自定义分类和标签
- **使用统计**：记录使用次数和热度

---

## 支持的资源类型

### 1. 图片 (Image)

- **格式**：JPG, PNG, WebP
- **用途**：贴图、环境图、UI素材
- **元数据**：尺寸、色彩模式

### 2. 视频 (Video)

- **格式**：MP4, WebM
- **用途**：视频贴图、背景视频
- **元数据**：时长、分辨率、帧率、编码格式

---

## 快速开始

### 初始化

```go
// 1. 数据库迁移
db.AutoMigrate(&Asset{}, &AssetTag{}, &AssetMetrics{})

// 2. 创建存储目录
os.MkdirAll("static/assets", 0755)

// 3. 初始化上传服务
uploadService := NewAssetUploadService(db, config)
```

### 上传资产

```bash
# 上传图片
POST /api/assets/upload
Content-Type: multipart/form-data
{
  file: <texture.png>,
  name: "Wood Texture",
  type: "image",
  category: "material",
  tags: "wood,pbr,seamless"
}

# 上传视频
POST /api/assets/upload
Content-Type: multipart/form-data
{
  file: <video.mp4>,
  name: "Background Video",
  type: "video",
  category: "background",
  tags: "loop,hd"
}
```

### 查询资产

```bash
# 按类型查询
GET /api/assets?type=image&page=1&pageSize=20

# 获取详情
GET /api/assets/:id
```

---

## 配置示例

```yaml
asset:
  storage_dir: "static/assets"
  max_file_size:
    image: 52428800 # 50MB
    video: 524288000 # 500MB
  allowed_formats:
    image: ["jpg", "jpeg", "png", "webp"]
    video: ["mp4", "webm"]
  thumbnail:
    width: 512
    height: 512
    quality: 85
```

---

## 技术栈

- **数据库**：GORM + MySQL/PostgreSQL
- **文件存储**：本地文件系统（统一存储在 static/assets）
- **图片处理**：imaging库（缩略图生成）
- **视频处理**：FFmpeg（截图）
- **并发控制**：Goroutine Pool + Channel

---

## 后续规划

- [ ] 在线预览（图片查看器、视频播放器）
- [ ] 智能推荐算法（基于使用统计）
- [ ] 资产版本管理
- [ ] 批量编辑工具
- [ ] CDN加速
- [ ] 云存储支持（OSS/S3）

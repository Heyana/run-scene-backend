# 文件库迁移记录

## 文件路径规范

### 存储路径结构

```
static/documents/
├── 1/                    # 文档ID
│   ├── file.pdf         # 原始文件
│   ├── thumbnail.jpg    # 缩略图
│   └── preview/         # 预览文件夹（PDF多页预览）
│       ├── page_1.jpg
│       ├── page_2.jpg
│       └── ...
├── 2/
│   ├── file.docx
│   └── thumbnail.jpg
└── ...
```

### 数据库路径记录

所有路径以 `static` 为基准记录**相对路径**：

```go
// Document 模型
type Document struct {
    FilePath      string  // 相对路径：documents/1/file.pdf
    ThumbnailPath string  // 相对路径：documents/1/thumbnail.jpg
    PreviewPath   string  // 相对路径：documents/1/preview/
}
```

### URL 拼接规则

在 `AfterFind` 钩子中自动拼接完整 URL：

```go
func (d *Document) AfterFind(tx *gorm.DB) error {
    if d.FilePath != "" {
        d.FileURL = buildDocumentURL(d.FilePath)
    }
    // ...
}

func buildDocumentURL(path string) string {
    // 1. 如果已经是完整 URL，直接返回
    if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
        return path
    }

    // 2. 统一转换为正斜杠
    path = strings.ReplaceAll(path, "\\", "/")

    // 3. 获取配置的 base_url
    docConfig, _ := config.LoadDocumentConfig()
    if docConfig != nil && docConfig.BaseURL != "" {
        baseURL := strings.TrimSuffix(docConfig.BaseURL, "/")
        filePath := strings.TrimPrefix(path, "/")
        // 移除 "static/documents/" 或 "documents/" 前缀
        filePath = strings.TrimPrefix(filePath, "static/documents/")
        filePath = strings.TrimPrefix(filePath, "documents/")
        return baseURL + "/" + filePath
    }

    // 4. 默认使用相对路径
    filePath := strings.TrimPrefix(path, "/")
    filePath = strings.TrimPrefix(filePath, "static/documents/")
    filePath = strings.TrimPrefix(filePath, "documents/")
    return "/documents/" + filePath
}
```

## 配置示例

### configs/document.yaml

```yaml
document:
  # 本地存储
  local_storage_enabled: false
  storage_dir: "static/documents"

  # 网络访问
  base_url: "http://192.168.3.39:23359/documents"

  # NAS 存储
  nas_enabled: true
  nas_path: "\\\\192.168.3.10\\project\\editor_v2\\static\\documents"
```

### 路径示例

#### 数据库存储

```
documents/1/file.pdf
documents/1/thumbnail.jpg
documents/1/preview/
```

#### 拼接后的 URL

```
http://192.168.3.39:23359/documents/1/file.pdf
http://192.168.3.39:23359/documents/1/thumbnail.jpg
http://192.168.3.39:23359/documents/1/preview/
```

## 已创建文件清单

### 后端文件

#### 配置

- [x] `configs/document.yaml` - 独立配置文件
- [x] `config/document_config.go` - 配置加载和管理

#### 模型

- [x] `models/document.go` - 数据模型定义
  - Document - 文档主表
  - DocumentMetadata - 元数据表
  - DocumentAccessLog - 访问日志表
  - DocumentMetrics - 统计指标表

#### 服务

- [x] `services/document/upload_service.go` - 上传服务
- [x] `services/document/query_service.go` - 查询服务

#### 控制器

- [x] `controllers/document_controller.go` - API 控制器

#### 数据库

- [x] `database/database.go` - 添加表迁移
  - Document
  - DocumentMetadata
  - DocumentAccessLog
  - DocumentMetrics

#### 路由

- [x] `api/routes.go` - 注册路由和静态文件服务
  - POST /api/documents/upload
  - GET /api/documents
  - GET /api/documents/:id
  - PUT /api/documents/:id
  - DELETE /api/documents/:id
  - GET /api/documents/:id/download
  - GET /api/documents/:id/versions
  - GET /api/documents/:id/logs
  - GET /api/documents/statistics
  - GET /api/documents/popular
  - GET /documents/\* (静态文件服务)

### 文档

- [x] `docs/项目/文件库/01-概述.md`
- [x] `docs/项目/文件库/02-数据模型.md`
- [x] `docs/项目/文件库/03-配置设计.md`
- [x] `docs/项目/文件库/04-API接口.md`
- [x] `docs/项目/文件库/05-Service层设计.md`
- [x] `docs/项目/文件库/06-实现计划.md`
- [x] `docs/项目/文件库/MIGRATION.md`

## 待实现功能

### 阶段二：文件处理

- [ ] PDF 处理器（缩略图、预览、元数据）
- [ ] 视频处理器（缩略图、元数据）
- [ ] Office 文档处理器（可选）
- [ ] 压缩包处理器（可选）

### 阶段三：版本管理

- [ ] 版本服务实现
- [ ] 上传新版本接口
- [ ] 版本切换接口

### 阶段四：权限和日志

- [ ] 权限服务实现
- [ ] 访问日志清理

### 阶段五：统计和优化

- [ ] 批量操作
- [ ] 预览缓存
- [ ] 异步处理大文件

## 测试清单

### 基础功能测试

- [ ] 上传 PDF 文件
- [ ] 上传 Word 文件
- [ ] 上传视频文件
- [ ] 上传压缩包
- [ ] 文件列表查询
- [ ] 文件详情查询
- [ ] 文件下载
- [ ] 文件删除

### 路径测试

- [ ] 验证数据库存储的是相对路径
- [ ] 验证 AfterFind 正确拼接 URL
- [ ] 验证静态文件服务正常访问
- [ ] 验证 NAS 路径访问

### 配置测试

- [ ] 本地存储模式
- [ ] NAS 存储模式
- [ ] 环境变量覆盖

## 注意事项

1. **路径一致性**：所有路径以 `static` 为基准，数据库只存储相对路径
2. **URL 拼接**：在 `AfterFind` 中自动拼接，前端无需处理
3. **配置独立**：文件库配置完全独立，不影响其他模块
4. **存储复用**：复用 `storage.FileStorageService`，支持本地和 NAS
5. **参考实现**：参考资产库的实现模式，保持代码风格一致

## 迁移步骤

### 1. 启动服务

```bash
go run main.go
```

### 2. 验证数据库表

检查以下表是否创建成功：

- document
- document_metadata
- document_access_log
- document_metrics

### 3. 测试上传

```bash
curl -X POST http://localhost:23359/api/documents/upload \
  -F "file=@test.pdf" \
  -F "name=测试文档" \
  -F "category=技术文档"
```

### 4. 验证路径

检查数据库中的路径格式：

```sql
SELECT id, name, file_path, thumbnail_path FROM document;
```

应该看到类似：

```
1 | 测试文档 | documents/1/file.pdf | documents/1/thumbnail.jpg
```

### 5. 验证 URL

查询文档详情，检查返回的 URL：

```bash
curl http://localhost:23359/api/documents/1
```

应该看到：

```json
{
  "file_url": "http://192.168.3.39:23359/documents/1/file.pdf",
  "thumbnail_url": "http://192.168.3.39:23359/documents/1/thumbnail.jpg"
}
```

## 前端集成

前端可以直接使用返回的 `file_url` 和 `thumbnail_url`，无需手动拼接路径。

```typescript
// API 调用
const response = await fetch('/api/documents/1');
const data = await response.json();

// 直接使用 URL
<img src={data.thumbnail_url} />
<a href={data.file_url} download>下载</a>
```

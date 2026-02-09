# Service 层设计

## 服务结构

### UploadService 上传服务

```go
type UploadService struct {
    db             *gorm.DB
    config         *config.DocumentConfig
    processors     map[string]DocumentProcessor
    storageService *storage.FileStorageService
}

// 主要方法
Upload(file, metadata) (*Document, error)
UploadVersion(documentID, file, versionInfo) (*Document, error)
```

**职责**：

- 文件上传处理
- 文件类型检测和验证
- 文件哈希计算
- 重复文件检查
- 缩略图生成
- 元数据提取
- 版本管理

### QueryService 查询服务

```go
type QueryService struct {
    db *gorm.DB
}

// 主要方法
List(page, pageSize, filters) ([]*Document, int64, error)
GetDetail(id) (*Document, *DocumentMetadata, error)
Search(keyword, filters) ([]*Document, error)
GetStatistics(filters) (map[string]interface{}, error)
GetPopular(limit, filters) ([]*Document, error)
Delete(id) error
Update(id, updates) error
```

**职责**：

- 文件列表查询
- 文件详情获取
- 文件搜索
- 统计信息
- 文件更新和删除

### VersionService 版本服务

```go
type VersionService struct {
    db *gorm.DB
}

// 主要方法
CreateVersion(documentID, file, versionInfo) (*Document, error)
GetVersions(documentID) ([]*Document, error)
ActivateVersion(documentID, versionID) error
DeleteVersion(versionID) error
GetVersionHistory(documentID) ([]VersionInfo, error)
```

**职责**：

- 版本创建
- 版本列表
- 版本切换
- 版本删除
- 版本历史

### PreviewService 预览服务

```go
type PreviewService struct {
    db     *gorm.DB
    config *config.DocumentConfig
}

// 主要方法
GeneratePDFPreview(filePath, outputPath, page) error
GenerateVideoThumbnail(filePath, outputPath, time) error
GetPreview(documentID, page) (string, error)
```

**职责**：

- PDF 转图片预览
- 视频缩略图生成
- 预览图获取

### AccessLogService 访问日志服务

```go
type AccessLogService struct {
    db *gorm.DB
}

// 主要方法
LogAccess(documentID, action, userName, userIP) error
GetAccessLogs(documentID, filters) ([]*DocumentAccessLog, int64, error)
CleanOldLogs(retentionDays) error
```

**职责**：

- 记录访问日志
- 查询访问日志
- 清理过期日志

### PermissionService 权限服务

```go
type PermissionService struct {
    db     *gorm.DB
    config *config.DocumentConfig
}

// 主要方法
CheckAccess(documentID, userName, department) (bool, error)
SetPermission(documentID, permission) error
GetUserDocuments(userName, department, project) ([]*Document, error)
```

**职责**：

- 权限检查
- 权限设置
- 用户文档查询

## Processor 处理器接口

```go
type DocumentProcessor interface {
    Validate(file) error
    GenerateThumbnail(filePath, outputPath) error
    ExtractMetadata(filePath) (*DocumentMetadata, error)
    GeneratePreview(filePath, outputPath) error
}
```

### 实现类

#### PDFProcessor

- 验证 PDF 文件
- 生成首页缩略图
- 提取页数、作者等元数据
- 生成多页预览图

#### VideoProcessor

- 验证视频文件
- 生成视频缩略图
- 提取时长、分辨率等元数据

#### DocumentProcessor（Office文档）

- 验证 Office 文档
- 生成预览图（需要 LibreOffice）
- 提取文档属性

#### ArchiveProcessor

- 验证压缩包
- 提取文件列表
- 生成文件树预览

## 服务依赖关系

```
Controller
    ├── UploadService
    │   ├── StorageService
    │   └── Processors
    │       ├── PDFProcessor
    │       ├── VideoProcessor
    │       ├── DocumentProcessor
    │       └── ArchiveProcessor
    ├── QueryService
    ├── VersionService
    │   └── UploadService
    ├── PreviewService
    ├── AccessLogService
    └── PermissionService
```

## 关键流程

### 上传流程

1. 接收文件和元数据
2. 检测文件类型
3. 验证文件格式和大小
4. 计算文件哈希
5. 检查重复文件
6. 保存文件到存储
7. 生成缩略图
8. 提取元数据
9. 创建数据库记录
10. 记录访问日志

### 版本管理流程

1. 上传新版本文件
2. 将旧版本标记为非最新
3. 创建新版本记录
4. 关联父版本ID
5. 更新版本号
6. 记录版本历史

### 预览生成流程

1. 检查预览缓存
2. 如果不存在，生成预览
3. PDF: 使用 ImageMagick 或 Poppler
4. 视频: 使用 FFmpeg
5. Office: 使用 LibreOffice
6. 保存预览图
7. 返回预览URL

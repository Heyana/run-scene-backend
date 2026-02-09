# 文件库 API 接口

## Controller 结构

```go
type DocumentController struct {
    uploadService   *document.UploadService
    queryService    *document.QueryService
    versionService  *document.VersionService
    previewService  *document.PreviewService
}
```

## 接口列表

### 1. 上传文件

```
POST /api/documents/upload
Content-Type: multipart/form-data

参数:
- file: 文件（必填）
- name: 文件名（必填）
- description: 描述
- category: 分类
- tags: 标签（逗号分隔）
- department: 部门
- project: 项目
- is_public: 是否公开
- version: 版本号（可选，自动生成）

返回:
{
    "code": 200,
    "message": "上传成功",
    "data": {
        "id": 1,
        "name": "项目文档.pdf",
        "file_url": "http://...",
        "thumbnail_url": "http://..."
    }
}
```

### 2. 文件列表

```
GET /api/documents

参数:
- page: 页码（默认1）
- pageSize: 每页数量（默认20）
- type: 文件类型（document/video/archive/other）
- category: 分类
- format: 格式
- department: 部门
- project: 项目
- keyword: 关键词
- sortBy: 排序字段（created_at/download_count/file_size）
- sortOrder: 排序方向（asc/desc）

返回:
{
    "code": 200,
    "data": {
        "list": [...],
        "total": 100,
        "page": 1,
        "pageSize": 20
    }
}
```

### 3. 文件详情

```
GET /api/documents/:id

返回:
{
    "code": 200,
    "data": {
        "document": {...},
        "metadata": {...},
        "versions": [...]  // 历史版本列表
    }
}
```

### 4. 下载文件

```
GET /api/documents/:id/download

返回: 文件流
```

### 5. 预览文件

```
GET /api/documents/:id/preview

参数:
- page: 预览页码（PDF多页预览）

返回: 预览图片或预览数据
```

### 6. 更新文件信息

```
PUT /api/documents/:id

参数:
{
    "name": "新文件名",
    "description": "新描述",
    "category": "新分类",
    "tags": "标签1,标签2"
}

返回:
{
    "code": 200,
    "message": "更新成功"
}
```

### 7. 删除文件

```
DELETE /api/documents/:id

返回:
{
    "code": 200,
    "message": "删除成功"
}
```

### 8. 上传新版本

```
POST /api/documents/:id/version

参数:
- file: 新版本文件
- version: 版本号（可选）
- description: 版本说明

返回:
{
    "code": 200,
    "message": "版本上传成功",
    "data": {
        "id": 2,
        "version": "v1.1"
    }
}
```

### 9. 版本列表

```
GET /api/documents/:id/versions

返回:
{
    "code": 200,
    "data": [
        {
            "id": 2,
            "version": "v1.1",
            "is_latest": true,
            "created_at": "2024-01-01"
        },
        {
            "id": 1,
            "version": "v1.0",
            "is_latest": false,
            "created_at": "2023-12-01"
        }
    ]
}
```

### 10. 切换版本

```
POST /api/documents/:id/versions/:versionId/activate

返回:
{
    "code": 200,
    "message": "版本切换成功"
}
```

### 11. 统计信息

```
GET /api/documents/statistics

参数:
- type: 文件类型
- department: 部门
- project: 项目

返回:
{
    "code": 200,
    "data": {
        "total_documents": 1000,
        "total_size": 10737418240,
        "type_distribution": {...},
        "format_distribution": {...},
        "recent_uploads": 50
    }
}
```

### 12. 热门文件

```
GET /api/documents/popular

参数:
- limit: 数量限制（默认10）
- type: 文件类型

返回:
{
    "code": 200,
    "data": [...]
}
```

### 13. 访问日志

```
GET /api/documents/:id/logs

参数:
- page: 页码
- pageSize: 每页数量
- action: 操作类型（view/download/upload/delete）

返回:
{
    "code": 200,
    "data": {
        "list": [...],
        "total": 100
    }
}
```

### 14. 批量操作

```
POST /api/documents/batch

参数:
{
    "action": "delete",  // delete/move/tag
    "ids": [1, 2, 3],
    "params": {...}      // 操作参数
}

返回:
{
    "code": 200,
    "message": "批量操作成功",
    "data": {
        "success": 3,
        "failed": 0
    }
}
```

## 路由注册

```go
func RegisterDocumentRoutes(router *gin.Engine, db *gorm.DB) {
    controller := NewDocumentController(db)

    api := router.Group("/api/documents")
    {
        // 基础操作
        api.POST("/upload", controller.Upload)
        api.GET("", controller.List)
        api.GET("/:id", controller.GetDetail)
        api.PUT("/:id", controller.Update)
        api.DELETE("/:id", controller.Delete)

        // 文件操作
        api.GET("/:id/download", controller.Download)
        api.GET("/:id/preview", controller.Preview)

        // 版本管理
        api.POST("/:id/version", controller.UploadVersion)
        api.GET("/:id/versions", controller.GetVersions)
        api.POST("/:id/versions/:versionId/activate", controller.ActivateVersion)

        // 统计和日志
        api.GET("/statistics", controller.GetStatistics)
        api.GET("/popular", controller.GetPopular)
        api.GET("/:id/logs", controller.GetAccessLogs)

        // 批量操作
        api.POST("/batch", controller.BatchOperation)
    }
}
```

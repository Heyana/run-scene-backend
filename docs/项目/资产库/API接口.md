# API接口

## 1. 资产上传

### 1.1 单文件上传

```
POST /api/assets/upload
Content-Type: multipart/form-data
```

**请求参数**：

```
file: File (必填)
name: string (必填)
type: string (必填，image/video)
description: string (选填)
category: string (选填)
tags: string (选填，逗号分隔)
```

**响应**：

```json
{
  "code": 200,
  "message": "上传成功",
  "data": {
    "id": 1,
    "name": "Wood Texture",
    "type": "image",
    "format": "png",
    "file_size": 2048576,
    "file_path": "/static/assets/1/file.png",
    "thumbnail_path": "/static/assets/1/thumbnail.webp",
    "category": "material",
    "tags": "wood,pbr,seamless",
    "metadata": {
      "width": 2048,
      "height": 2048,
      "color_mode": "RGB"
    },
    "created_at": "2024-01-20T10:00:00Z"
  }
}
```

---

## 2. 资产查询

### 2.1 分页列表

```
GET /api/assets
```

**查询参数**：

```
page: int (默认1)
pageSize: int (默认20)
type: string (选填：image, video)
category: string (选填)
tags: string (选填，逗号分隔)
format: string (选填：jpg, png, mp4, webm等)
keyword: string (选填，搜索名称)
sortBy: string (选填：name, created_at, use_count, file_size)
sortOrder: string (选填：asc, desc，默认desc)
```

**响应**：

```json
{
  "code": 200,
  "data": {
    "list": [
      {
        "id": 1,
        "name": "Wood Texture",
        "type": "image",
        "format": "png",
        "category": "material",
        "file_size": 2048576,
        "thumbnail_path": "/static/assets/1/thumbnail.webp",
        "use_count": 5,
        "created_at": "2024-01-20T10:00:00Z"
      }
    ],
    "total": 100,
    "page": 1,
    "pageSize": 20
  }
}
```

---

### 2.2 资产详情

```
GET /api/assets/:id
```

**响应**：

```json
{
  "code": 200,
  "data": {
    "asset": {
      "id": 1,
      "name": "Wood Texture",
      "description": "Seamless wood texture",
      "type": "image",
      "format": "png",
      "category": "material",
      "file_size": 2048576,
      "file_path": "/static/assets/1/file.png",
      "thumbnail_path": "/static/assets/1/thumbnail.webp",
      "use_count": 5,
      "uploaded_by": "admin",
      "created_at": "2024-01-20T10:00:00Z"
    },
    "metadata": {
      "width": 2048,
      "height": 2048,
      "color_mode": "RGB"
    },
    "tags": [
      { "id": 1, "name": "wood", "type": "tag" },
      { "id": 2, "name": "pbr", "type": "tag" }
    ]
  }
}
```

---

### 2.3 按类型查询

```
GET /api/assets/type/:type
```

**路径参数**：

- type: texture, environment, video, audio

**查询参数**：同 2.1

**响应**：同 2.1

---

### 2.4 按分类查询

```
GET /api/assets/category/:category
```

**查询参数**：同 2.1

**响应**：同 2.1

---

### 2.5 搜索

```
GET /api/assets/search
```

**查询参数**：

```
keyword: string (必填)
type: string (选填)
page: int
pageSize: int
```

**响应**：同 2.1

---

## 3. 资产管理

### 3.1 更新资产信息

```
PUT /api/assets/:id
Content-Type: application/json
```

**请求体**：

```json
{
  "name": "Updated Name",
  "description": "Updated description",
  "category": "new-category",
  "tags": ["tag1", "tag2"]
}
```

**响应**：

```json
{
  "code": 200,
  "message": "更新成功",
  "data": {...}
}
```

---

### 3.2 删除资产

```
DELETE /api/assets/:id
```

**响应**：

```json
{
  "code": 200,
  "message": "删除成功"
}
```

**说明**：删除资产会同时删除磁盘上的文件和缩略图

---

### 3.3 记录使用

```
POST /api/assets/:id/use
```

**响应**：

```json
{
  "code": 200,
  "message": "记录成功",
  "data": {
    "use_count": 6
  }
}
```

---

### 3.4 批量删除

```
POST /api/assets/batch-delete
Content-Type: application/json
```

**请求体**：

```json
{
  "ids": [1, 2, 3, 4, 5]
}
```

**响应**：

```json
{
  "code": 200,
  "message": "批量删除成功",
  "data": {
    "deleted_count": 5
  }
}
```

---

## 4. 标签管理

### 4.1 获取所有标签

```
GET /api/assets/tags
```

**查询参数**：

```
type: string (选填：tag, category)
asset_type: string (选填：image, video)
```

**响应**：

```json
{
  "code": 200,
  "data": [
    {
      "id": 1,
      "name": "wood",
      "type": "tag",
      "use_count": 10
    }
  ]
}
```

---

### 4.2 按标签查询资产

```
GET /api/assets/tags/:tagId/assets
```

**查询参数**：

```
page: int
pageSize: int
```

**响应**：同 2.1

---

## 5. 统计信息

### 5.1 获取统计数据

```
GET /api/assets/statistics
```

**查询参数**：

```
type: string (选填：image, video)
```

**响应**：

```json
{
  "code": 200,
  "data": {
    "total_assets": 500,
    "total_size": 5368709120,
    "type_distribution": {
      "image": 400,
      "video": 100
    },
    "format_distribution": {
      "png": 200,
      "jpg": 150,
      "webp": 50,
      "mp4": 80,
      "webm": 20
    },
    "category_distribution": {
      "material": 200,
      "environment": 100,
      "background": 50
    },
    "recent_uploads": 25
  }
}
```

---

### 5.2 获取热门资产

```
GET /api/assets/popular
```

**查询参数**：

```
limit: int (默认10)
type: string (选填)
```

**响应**：同 2.1

---

### 5.3 按类型统计

```
GET /api/assets/statistics/by-type
```

**响应**：

```json
{
  "code": 200,
  "data": {
    "image": {
      "count": 400,
      "total_size": 4294967296,
      "avg_size": 10737418
    },
    "video": {
      "count": 100,
      "total_size": 1073741824,
      "avg_size": 10737418
    }
  }
}
```

---

## 6. 批量操作

### 6.1 批量上传

```
POST /api/assets/batch-upload
Content-Type: multipart/form-data
```

**请求参数**：

```
files: File[] (必填，多个文件)
type: string (必填)
category: string (选填)
tags: string (选填)
```

**响应**：

```json
{
  "code": 200,
  "message": "批量上传完成",
  "data": {
    "success_count": 8,
    "failed_count": 2,
    "results": [
      {
        "filename": "texture1.png",
        "status": "success",
        "asset_id": 1
      },
      {
        "filename": "texture2.png",
        "status": "failed",
        "error": "文件格式不支持"
      }
    ]
  }
}
```

---

### 6.2 批量更新标签

```
POST /api/assets/batch-update-tags
Content-Type: application/json
```

**请求体**：

```json
{
  "asset_ids": [1, 2, 3],
  "tags": ["tag1", "tag2"],
  "operation": "add" // add, remove, replace
}
```

**响应**：

```json
{
  "code": 200,
  "message": "批量更新成功",
  "data": {
    "updated_count": 3
  }
}
```

---

## 错误码

| 错误码 | 说明             |
| ------ | ---------------- |
| 200    | 成功             |
| 400    | 请求参数错误     |
| 404    | 资源不存在       |
| 413    | 文件过大         |
| 415    | 不支持的文件类型 |
| 500    | 服务器错误       |

**错误响应格式**：

```json
{
  "code": 400,
  "message": "文件类型不支持",
  "error": "invalid_type"
}
```

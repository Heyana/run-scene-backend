# API接口

## 1. 模型上传

### 1.1 单文件上传

```
POST /api/models/upload
Content-Type: multipart/form-data
```

**请求参数**：

```
model: File (必填，glb或glt文件)
thumbnail: File (必填，预览图)
name: string (必填)
type: string (必填，glb或glt)
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
    "name": "Wooden Chair",
    "type": "glb",
    "file_size": 2048576,
    "file_path": "/static/models/1/model.glb",
    "thumbnail_path": "/static/models/1/thumbnail.webp",
    "category": "furniture",
    "tags": "chair,wood,modern",
    "created_at": "2024-01-20T10:00:00Z"
  }
}
```

---

## 2. 模型查询

### 2.1 分页列表

```
GET /api/models
```

**查询参数**：

```
page: int (默认1)
pageSize: int (默认20)
category: string (选填)
tags: string (选填，逗号分隔)
type: string (选填：glb, glt)
keyword: string (选填，搜索名称)
sortBy: string (选填：name, created_at, use_count)
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
        "name": "Wooden Chair",
        "category": "furniture",
        "type": "glb",
        "file_size": 2048576,
        "thumbnail_path": "/static/models/1/thumbnail.webp",
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

### 2.2 模型详情

```
GET /api/models/:id
```

**响应**：

```json
{
  "code": 200,
  "data": {
    "model": {
      "id": 1,
      "name": "Wooden Chair",
      "description": "Modern wooden chair",
      "category": "furniture",
      "type": "glb",
      "file_size": 2048576,
      "file_path": "/static/models/1/model.glb",
      "thumbnail_path": "/static/models/1/thumbnail.webp",
      "use_count": 5,
      "uploaded_by": "admin",
      "created_at": "2024-01-20T10:00:00Z"
    },
    "tags": [
      {
        "id": 1,
        "name": "chair",
        "type": "tag"
      },
      {
        "id": 2,
        "name": "furniture",
        "type": "category"
      }
    ]
  }
}
```

---

### 2.3 按分类查询

```
GET /api/models/category/:category
```

**查询参数**：

```
page: int
pageSize: int
```

**响应**：同 2.1

---

### 2.4 搜索

```
GET /api/models/search
```

**查询参数**：

```
keyword: string (必填)
page: int
pageSize: int
```

**响应**：同 2.1

---

## 3. 模型管理

### 3.1 更新模型信息

```
PUT /api/models/:id
Content-Type: application/json
```

**请求体**：

```json
{
  "name": "Updated Name",
  "description": "Updated description",
  "category": "furniture",
  "tags": ["chair", "wood", "modern"]
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

### 3.2 删除模型

```
DELETE /api/models/:id
```

**响应**：

```json
{
  "code": 200,
  "message": "删除成功"
}
```

**说明**：删除模型会同时删除磁盘上的文件

---

### 3.3 记录使用

```
POST /api/models/:id/use
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

## 4. 标签管理

### 4.1 获取所有标签

```
GET /api/models/tags
```

**查询参数**：

```
type: string (选填：tag, category)
```

**响应**：

```json
{
  "code": 200,
  "data": [
    {
      "id": 1,
      "name": "chair",
      "type": "tag",
      "use_count": 10
    }
  ]
}
```

---

### 4.2 按标签查询模型

```
GET /api/models/tags/:tagId/models
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
GET /api/models/statistics
```

**响应**：

```json
{
  "code": 200,
  "data": {
    "total_models": 100,
    "total_size": 1073741824,
    "type_distribution": {
      "glb": 80,
      "glt": 20
    },
    "category_distribution": {
      "furniture": 40,
      "architecture": 30,
      "nature": 30
    },
    "recent_uploads": 15
  }
}
```

---

### 5.2 获取热门模型

```
GET /api/models/popular
```

**查询参数**：

```
limit: int (默认10)
```

**响应**：同 2.1

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

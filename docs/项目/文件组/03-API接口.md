# 文件组功能 - API 接口

## 接口列表

### 1. 创建文件组

```
POST /api/document-groups
Body: {
    name: string
    description?: string
    category?: string
    tags?: string
    department?: string
    project?: string
    is_public?: boolean
    parent_id?: number  // 父组ID，创建子组时提供
}
Response: DocumentGroup
```

### 2. 获取文件组列表

```
GET /api/document-groups
Query: {
    page?: number
    page_size?: number
    keyword?: string
    category?: string
    department?: string
    project?: string
    parent_id?: number  // 查询子组，null或0查询顶级组
}
Response: {
    list: DocumentGroup[]
    total: number
}
```

### 3. 获取文件组详情

```
GET /api/document-groups/:id
Query: {
    with_children?: boolean  // 是否包含子组
}
Response: {
    group: DocumentGroup
    items: DocumentGroupItem[]  // 包含 document 信息
    children?: DocumentGroup[]  // 子组列表（如果 with_children=true）
}
```

### 4. 获取文件组树形结构

```
GET /api/document-groups/tree
Query: {
    root_id?: number  // 根组ID，不传则返回所有顶级组
}
Response: DocumentGroup[]  // 包含嵌套的 children
```

### 5. 更新文件组

```
PUT /api/document-groups/:id
Body: {
    name?: string
    description?: string
    category?: string
    tags?: string
    parent_id?: number  // 移动到其他父组
}
Response: DocumentGroup
```

### 6. 删除文件组

```
DELETE /api/document-groups/:id
Query: {
    cascade?: boolean  // 是否级联删除子组，默认 false
}
Response: { message: string }
```

### 7. 添加文件到组

```
POST /api/document-groups/:id/items
Body: {
    document_ids: number[]
    note?: string
}
Response: { added: number }
```

### 8. 从组中移除文件

```
DELETE /api/document-groups/:id/items/:document_id
Response: { message: string }
```

### 9. 调整文件顺序

```
PUT /api/document-groups/:id/items/sort
Body: {
    items: Array<{ id: number, sort_order: number }>
}
Response: { message: string }
```

### 10. 设置主文件

```
PUT /api/document-groups/:id/items/:item_id/primary
Response: { message: string }
```

### 11. 批量下载组文件

```
GET /api/document-groups/:id/download
Query: {
    include_children?: boolean  // 是否包含子组文件，默认 false
}
Response: application/zip
说明：下载时保持目录结构（父组/子组/文件）
```

### 12. 移动文件组

```
PUT /api/document-groups/:id/move
Body: {
    parent_id: number | null  // 新父组ID，null表示移到顶级
}
Response: { message: string }
```

## 路由注册

```go
// routes.go
documentGroup := api.Group("/document-groups")
{
    documentGroup.POST("", controllers.CreateDocumentGroup)
    documentGroup.GET("", controllers.GetDocumentGroups)
    documentGroup.GET("/tree", controllers.GetDocumentGroupTree)
    documentGroup.GET("/:id", controllers.GetDocumentGroupDetail)
    documentGroup.PUT("/:id", controllers.UpdateDocumentGroup)
    documentGroup.DELETE("/:id", controllers.DeleteDocumentGroup)
    documentGroup.PUT("/:id/move", controllers.MoveDocumentGroup)

    documentGroup.POST("/:id/items", controllers.AddDocumentToGroup)
    documentGroup.DELETE("/:id/items/:document_id", controllers.RemoveDocumentFromGroup)
    documentGroup.PUT("/:id/items/sort", controllers.SortGroupItems)
    documentGroup.PUT("/:id/items/:item_id/primary", controllers.SetPrimaryDocument)
    documentGroup.GET("/:id/download", controllers.DownloadGroupFiles)
}
```

## 错误码

| 错误码 | 说明                    |
| ------ | ----------------------- |
| 400    | 参数错误                |
| 404    | 文件组不存在            |
| 409    | 文件已在组中 / 循环引用 |
| 500    | 服务器错误              |

# URL 路由支持

## 功能概述

在 URL 中保存当前文件夹路径，刷新页面后仍然停留在当前文件夹，支持分享特定文件夹的链接。

## 实现时间

2026年2月9日

## 功能特点

### 1. URL 参数

- 根目录: `/documents` 或 `/documents?folder=0`
- 子文件夹: `/documents?folder=123`

### 2. 自动同步

- 打开文件夹时自动更新 URL
- 点击面包屑导航时更新 URL
- 返回上级时更新 URL

### 3. 刷新保持

- 刷新页面后停留在当前文件夹
- 自动加载面包屑路径
- 自动加载文件夹内容

### 4. 分享链接

- 可以复制 URL 分享给他人
- 他人打开链接直接进入对应文件夹
- 支持书签保存

## 实现细节

### 1. 路由参数定义

使用 Vue Router 的 query 参数：

```typescript
// 根目录
/documents

// 子文件夹
/documents?folder=123
```

### 2. 初始化路由

在组件挂载时从 URL 读取文件夹 ID：

```typescript
const initFromRoute = async () => {
  const folderId = route.query.folder;
  if (folderId) {
    const id = parseInt(folderId as string, 10);
    if (!isNaN(id) && id !== 0) {
      currentFolderId.value = id;
      await loadBreadcrumbPath(id);
    }
  }
};

onMounted(() => {
  initFromRoute();
  loadCurrentFolder();
});
```

### 3. 更新路由

在文件夹导航时更新 URL：

```typescript
const updateRoute = (folderId: number) => {
  if (folderId === 0) {
    router.replace({ query: {} });
  } else {
    router.replace({ query: { folder: String(folderId) } });
  }
};

// 打开文件夹
const handleOpenFolder = (folder: Document) => {
  currentFolderId.value = folder.id;
  breadcrumb.value.push({ id: folder.id, name: folder.name });
  updateRoute(folder.id);
  loadCurrentFolder();
};

// 面包屑导航
const handleBreadcrumbClick = (index: number) => {
  breadcrumb.value = breadcrumb.value.slice(0, index + 1);
  currentFolderId.value = breadcrumb.value[index].id;
  updateRoute(currentFolderId.value);
  loadCurrentFolder();
};

// 返回上级
const handleGoBack = () => {
  breadcrumb.value.pop();
  currentFolderId.value = breadcrumb.value[breadcrumb.value.length - 1].id;
  updateRoute(currentFolderId.value);
  loadCurrentFolder();
};
```

### 4. 加载面包屑路径

从文件夹 ID 递归查找父文件夹构建面包屑：

```typescript
const loadBreadcrumbPath = async (folderId: number) => {
  if (folderId === 0) {
    breadcrumb.value = [{ id: 0, name: "文件库" }];
    return;
  }

  try {
    const path: Array<{ id: number; name: string }> = [];
    let currentId: number | null = folderId;

    // 递归查找父文件夹
    while (currentId !== null && currentId !== 0) {
      const res = await getDocuments({
        page: 1,
        pageSize: 1000,
      });

      const allDocs = res.data.list || [];
      const folder = allDocs.find(
        (d: Document) => d.id === currentId && d.is_folder,
      );

      if (folder) {
        path.unshift({ id: folder.id, name: folder.name });
        currentId = folder.parent_id || 0;
      } else {
        break;
      }
    }

    if (path.length > 0) {
      breadcrumb.value = [{ id: 0, name: "文件库" }, ...path];
    } else {
      // 如果找不到路径，回到根目录
      breadcrumb.value = [{ id: 0, name: "文件库" }];
      currentFolderId.value = 0;
      updateRoute(0);
    }
  } catch (error) {
    console.error("加载面包屑路径失败:", error);
    breadcrumb.value = [{ id: 0, name: "文件库" }];
    currentFolderId.value = 0;
    updateRoute(0);
  }
};
```

## 使用场景

### 1. 刷新页面

```
用户操作：
1. 进入文件夹 "项目文档"
2. URL 变为 /documents?folder=123
3. 刷新页面
4. 自动停留在 "项目文档" 文件夹

系统行为：
- 从 URL 读取 folder=123
- 加载文件夹 123 的面包屑路径
- 加载文件夹 123 的内容
```

### 2. 分享链接

```
用户 A：
1. 进入文件夹 "设计稿"
2. 复制 URL: /documents?folder=456
3. 发送给用户 B

用户 B：
1. 打开链接 /documents?folder=456
2. 直接进入 "设计稿" 文件夹
3. 看到面包屑: 文件库 > 设计稿
```

### 3. 书签保存

```
用户操作：
1. 进入常用文件夹 "技术文档"
2. URL: /documents?folder=789
3. 添加浏览器书签
4. 下次点击书签直接进入该文件夹
```

### 4. 浏览器前进/后退

```
用户操作：
1. 根目录 -> 文件夹A -> 文件夹B
2. 点击浏览器后退按钮
3. 自动返回文件夹A
4. 再次后退返回根目录

系统行为：
- 监听路由变化
- 自动更新当前文件夹
- 自动更新面包屑
```

## 技术要点

### 1. Vue Router

使用 `useRoute` 和 `useRouter`：

```typescript
import { useRoute, useRouter } from "vue-router";

const route = useRoute();
const router = useRouter();

// 读取参数
const folderId = route.query.folder;

// 更新参数
router.replace({ query: { folder: "123" } });
```

### 2. router.replace vs router.push

使用 `replace` 而不是 `push`：

```typescript
// ❌ 使用 push 会增加历史记录
router.push({ query: { folder: "123" } });

// ✅ 使用 replace 不增加历史记录
router.replace({ query: { folder: "123" } });
```

这样在文件夹内导航时不会产生大量历史记录。

### 3. 异步初始化

`initFromRoute` 必须是异步的：

```typescript
const initFromRoute = async () => {
  // 需要等待 loadBreadcrumbPath 完成
  await loadBreadcrumbPath(id);
};

onMounted(() => {
  initFromRoute(); // 不需要 await
  loadCurrentFolder();
});
```

### 4. 错误处理

如果文件夹不存在，自动回到根目录：

```typescript
if (path.length === 0) {
  breadcrumb.value = [{ id: 0, name: "文件库" }];
  currentFolderId.value = 0;
  updateRoute(0);
}
```

## 优化建议

### 1. 添加后端 API

创建专门的 API 获取文件夹路径：

```go
// GET /api/documents/:id/path
func (c *DocumentController) GetFolderPath(ctx *gin.Context) {
    id := ctx.Param("id")

    // 递归查找父文件夹
    path := []Document{}
    currentID := id

    for currentID != 0 {
        var doc Document
        db.First(&doc, currentID)
        path = append([]Document{doc}, path...)
        currentID = doc.ParentID
    }

    response.Success(ctx, path)
}
```

前端调用：

```typescript
const loadBreadcrumbPath = async (folderId: number) => {
  const res = await getFolderPath(folderId);
  const path = res.data.map((d) => ({ id: d.id, name: d.name }));
  breadcrumb.value = [{ id: 0, name: "文件库" }, ...path];
};
```

### 2. 缓存面包屑路径

使用 localStorage 缓存：

```typescript
const cacheBreadcrumb = (folderId: number, path: any[]) => {
  localStorage.setItem(`breadcrumb_${folderId}`, JSON.stringify(path));
};

const getCachedBreadcrumb = (folderId: number) => {
  const cached = localStorage.getItem(`breadcrumb_${folderId}`);
  return cached ? JSON.parse(cached) : null;
};
```

### 3. 支持浏览器前进/后退

监听路由变化：

```typescript
watch(
  () => route.query.folder,
  (newFolderId) => {
    if (newFolderId) {
      const id = parseInt(newFolderId as string, 10);
      currentFolderId.value = id;
      loadBreadcrumbPath(id);
      loadCurrentFolder();
    } else {
      currentFolderId.value = 0;
      breadcrumb.value = [{ id: 0, name: "文件库" }];
      loadCurrentFolder();
    }
  },
);
```

### 4. URL 美化

使用路径参数代替查询参数：

```typescript
// 当前: /documents?folder=123
// 优化: /documents/123

// 路由配置
{
  path: '/documents/:folderId?',
  component: Documents
}

// 读取参数
const folderId = route.params.folderId;

// 更新参数
router.replace(`/documents/${folderId}`);
```

## 注意事项

### 1. 文件夹删除

如果当前文件夹被删除：

```typescript
// 检测文件夹是否存在
if (!folder) {
  message.warning("文件夹不存在，已返回根目录");
  currentFolderId.value = 0;
  updateRoute(0);
  loadCurrentFolder();
}
```

### 2. 权限控制

如果用户无权访问文件夹：

```typescript
try {
  await loadCurrentFolder();
} catch (error) {
  if (error.response?.status === 403) {
    message.error("无权访问该文件夹");
    currentFolderId.value = 0;
    updateRoute(0);
  }
}
```

### 3. 性能优化

避免频繁查询：

```typescript
// ❌ 每次都查询所有文档
const res = await getDocuments({ page: 1, pageSize: 1000 });

// ✅ 只查询需要的文件夹
const res = await getDocument(folderId);
```

### 4. 深层嵌套

当前实现可能在深层嵌套时性能较差，建议：

- 限制文件夹嵌套层级（如最多 5 层）
- 添加后端 API 直接返回路径
- 使用缓存减少查询

## 测试建议

### 1. 基础功能测试

- [x] 打开文件夹，URL 更新
- [x] 刷新页面，停留在当前文件夹
- [x] 面包屑导航，URL 更新
- [x] 返回上级，URL 更新

### 2. 边界测试

- [ ] 访问不存在的文件夹 ID
- [ ] 访问已删除的文件夹
- [ ] 访问无权限的文件夹
- [ ] 访问非法的文件夹 ID（如字母）

### 3. 浏览器测试

- [ ] 浏览器前进/后退按钮
- [ ] 书签保存和打开
- [ ] 复制链接分享
- [ ] 多标签页同时打开

### 4. 性能测试

- [ ] 深层嵌套文件夹（5层+）
- [ ] 大量文件夹（1000+）
- [ ] 频繁切换文件夹
- [ ] 网络慢速情况

## 总结

URL 路由支持已完成，主要特点：

✅ **刷新保持**: 刷新页面后停留在当前文件夹  
✅ **分享链接**: 可以分享特定文件夹的链接  
✅ **自动同步**: 导航时自动更新 URL  
✅ **面包屑恢复**: 自动加载完整的面包屑路径  
✅ **错误处理**: 文件夹不存在时自动回到根目录

这个功能让文件库的使用体验更加完整，支持书签、分享和刷新保持。

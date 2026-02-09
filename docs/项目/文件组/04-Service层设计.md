# 文件组功能 - Service 层设计

## 服务类设计

### DocumentGroupService - 文件组服务

```go
type DocumentGroupService struct {
    db *gorm.DB
}

// 创建文件组
Create(group *DocumentGroup) error

// 获取文件组列表
GetList(params QueryParams) ([]DocumentGroup, int64, error)

// 获取文件组树形结构
GetTree(rootID *uint) ([]DocumentGroup, error)

// 获取文件组详情（包含文件列表和子组）
GetDetail(groupID uint, withChildren bool) (*DocumentGroup, []DocumentGroupItem, []DocumentGroup, error)

// 更新文件组
Update(groupID uint, updates map[string]interface{}) error

// 删除文件组
Delete(groupID uint, cascade bool) error

// 移动文件组到其他父组
Move(groupID uint, newParentID *uint) error

// 添加文件到组
AddDocuments(groupID uint, documentIDs []uint, note string) error

// 从组中移除文件
RemoveDocument(groupID uint, documentID uint) error

// 调整文件顺序
SortItems(groupID uint, items []SortItem) error

// 设置主文件
SetPrimary(groupID uint, itemID uint) error

// 更新组统计（内部方法）
updateGroupStats(groupID uint) error

// 更新父组统计（内部方法，递归向上）
updateParentStats(groupID uint) error

// 检测循环引用（内部方法）
checkCircularReference(groupID uint, newParentID uint) error

// 更新路径和层级（内部方法）
updatePathAndLevel(groupID uint) error
```

### DownloadService - 批量下载服务

```go
type DownloadService struct {
    db *gorm.DB
}

// 打包下载文件组（支持嵌套）
PackGroupFiles(groupID uint, includeChildren bool) (zipPath string, err error)

// 递归添加子组文件到 ZIP（内部方法）
addGroupToZip(zipWriter *zip.Writer, group *DocumentGroup, basePath string) error
```

## 关键流程

### 1. 创建文件组（支持嵌套）

```
1. 验证父组存在（如果提供了 parent_id）
2. 检测循环引用
3. 计算 level 和 path
   - level = parent.level + 1
   - path = parent.path + "/" + id
4. 继承父组权限（department, project）
5. 创建文件组记录
6. 更新父组的 child_count
```

### 2. 移动文件组

```
1. 验证目标父组存在
2. 检测循环引用（不能移到自己的子孙组）
3. 更新 parent_id
4. 递归更新自己和所有子孙组的 path 和 level
5. 更新旧父组和新父组的统计
```

### 3. 删除文件组

```
1. 验证文件组存在
2. 如果 cascade=true
   - 递归删除所有子组
3. 如果 cascade=false 且有子组
   - 返回错误
4. 删除所有文件关联记录
5. 删除文件组记录
6. 更新父组统计
```

### 4. 添加文件到组

```
1. 验证文件组存在
2. 验证文件存在
3. 检查文件是否已在组中（去重）
4. 批量插入关联记录
5. 更新组统计（file_count, total_files, total_size）
6. 递归更新父组统计
```

### 5. 批量下载（支持嵌套）

```
1. 查询组信息
2. 创建临时 ZIP 文件
3. 添加当前组的文件到 ZIP
   - 路径：组名/文件名
4. 如果 include_children=true
   - 递归添加子组文件
   - 路径：组名/子组名/文件名
5. 返回 ZIP 文件路径
6. 定时清理临时文件
```

### 6. 更新统计（递归向上）

```
1. 统计当前组的直接文件
   - file_count
   - 直接文件的 total_size
2. 统计所有子组
   - child_count
   - 累加子组的 total_files 和 total_size
3. 计算 total_files = file_count + 所有子组的 total_files
4. 更新当前组记录
5. 如果有父组，递归更新父组统计
```

### 7. 获取树形结构

```
1. 查询顶级组（parent_id IS NULL）或指定根组
2. 递归查询每个组的子组
3. 构建树形结构（Children 字段）
4. 返回树形数据
```

## 事务处理

需要事务的操作：

- 创建文件组（插入 + 更新父组统计）
- 移动文件组（更新 + 递归更新子组 + 更新父组统计）
- 删除文件组（删除关联 + 删除子组 + 删除组 + 更新父组统计）
- 添加文件到组（插入 + 更新统计 + 更新父组统计）

## 错误处理

```go
var (
    ErrGroupNotFound      = errors.New("文件组不存在")
    ErrDocumentNotFound   = errors.New("文件不存在")
    ErrDocumentInGroup    = errors.New("文件已在组中")
    ErrEmptyGroup         = errors.New("文件组为空")
    ErrCircularReference  = errors.New("不能移动到子组（循环引用）")
    ErrHasChildren        = errors.New("文件组包含子组，请先删除子组或使用级联删除")
)
```

## 性能优化

1. **Path 字段** - 使用路径字段快速查询所有子孙组

   ```sql
   WHERE path LIKE '/1/3/%'  -- 查询 /1/3 的所有子孙组
   ```

2. **统计缓存** - 统计字段冗余存储，避免实时计算

3. **批量更新** - 统计更新使用批量 SQL

4. **延迟更新** - 统计更新可以异步处理（可选）

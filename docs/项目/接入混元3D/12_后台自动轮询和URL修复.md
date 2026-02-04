# 后台自动轮询和URL修复

## 完成时间

2026-02-04

## 更新内容

### 1. 添加后台自动轮询机制

#### 新增文件

- `services/hunyuan/poller.go` - 任务轮询器

#### 功能特性

- **自动轮询**: 后台定期自动轮询所有等待中(WAIT)和运行中(RUN)的任务
- **并发控制**: 最多同时轮询5个任务，避免API调用过于频繁
- **智能日志**: 记录任务状态变化，完成和失败时输出日志
- **可配置间隔**: 通过 `config.yaml` 中的 `poll_interval` 配置轮询间隔

#### 工作流程

1. 服务启动时自动启动轮询器
2. 每隔N秒（默认5秒）查询所有未完成的任务
3. 并发调用腾讯云API查询任务状态
4. 自动更新数据库中的任务状态
5. 任务完成时自动下载文件到本地/NAS

#### 配置

```yaml
hunyuan:
  poll_interval: 5 # 轮询间隔（秒）
```

#### API端点

- `GET /api/hunyuan/poller/status` - 获取轮询器状态

响应示例：

```json
{
  "code": 0,
  "data": {
    "running": true,
    "interval": "5s",
    "pendingTasks": 3
  }
}
```

### 2. 修复文件URL生成

#### 问题

- 任务完成后，`localPath` 和 `nasPath` 字段有值，但前端无法直接访问
- 缺少 `fileUrl` 和 `thumbnailUrl` 字段

#### 解决方案

在 `models/hunyuan/task.go` 中添加 `AfterFind` 钩子：

```go
// AfterFind GORM钩子：查询后自动生成URL
func (t *HunyuanTask) AfterFind(tx *gorm.DB) error {
    // 生成文件URL
    if t.LocalPath != nil && *t.LocalPath != "" {
        t.FileURL = buildHunyuanURL(*t.LocalPath)
    } else if t.NASPath != nil && *t.NASPath != "" {
        t.FileURL = buildHunyuanURL(*t.NASPath)
    }

    // 生成缩略图URL
    if t.ThumbnailPath != nil && *t.ThumbnailPath != "" {
        t.ThumbnailURL = buildHunyuanURL(*t.ThumbnailPath)
    }

    return nil
}
```

#### URL生成逻辑

1. 优先使用 `localPath`，如果为空则使用 `nasPath`
2. 自动处理Windows路径分隔符（`\` -> `/`）
3. 移除路径前缀（`static/hunyuan/`）
4. 处理NAS路径格式（`\\192.168.3.10\...`）
5. 拼接完整URL：`http://192.168.3.39:23359/hunyuan/2026/02/file.glb`

#### 新增字段

```go
// 动态生成的URL字段（不存储到数据库）
FileURL      string `gorm:"-" json:"fileUrl"`
ThumbnailURL string `gorm:"-" json:"thumbnailUrl"`
```

### 3. 修改配置服务

#### 问题

- 原来从数据库读取配置，但数据库中可能没有配置记录
- 导致提示"未配置API密钥"

#### 解决方案

直接从 `config.AppConfig.Hunyuan` 读取配置，不再使用数据库：

```go
// GetConfig 获取配置（从 config.yaml 读取）
func (s *ConfigService) GetConfig() (*hunyuan.HunyuanConfig, error) {
    cfg := &config.AppConfig.Hunyuan

    if cfg.SecretID == "" || cfg.SecretKey == "" {
        return nil, errors.New("未配置API密钥，请在 config.yaml 中配置")
    }

    // 转换为模型配置
    return modelConfig, nil
}
```

#### 禁用配置更新

```go
// UpdateConfig 更新配置（已禁用）
func (s *ConfigService) UpdateConfig(config *hunyuan.HunyuanConfig) error {
    return errors.New("配置更新功能已禁用，请直接修改 config.yaml 文件")
}
```

## 使用效果

### 1. 自动轮询

- 用户提交任务后，无需手动点击"轮询任务"
- 后台自动每5秒检查一次任务状态
- 任务完成后自动下载文件
- 前端刷新列表即可看到最新状态

### 2. 文件访问

任务完成后，API返回：

```json
{
  "id": 1,
  "status": "DONE",
  "localPath": "",
  "nasPath": "\\\\192.168.3.10\\project\\editor_v2\\static\\hunyuan\\2026\\02\\file.glb",
  "thumbnailPath": "\\\\192.168.3.10\\project\\editor_v2\\static\\hunyuan\\2026\\02\\file.png",
  "fileUrl": "http://192.168.3.39:23359/hunyuan/2026/02/file.glb",
  "thumbnailUrl": "http://192.168.3.39:23359/hunyuan/2026/02/file.png"
}
```

前端可以直接使用 `fileUrl` 和 `thumbnailUrl` 访问文件。

### 3. 配置管理

- 所有配置在 `config.yaml` 中管理
- 修改配置后重启服务即可生效
- 不再需要通过API更新配置

## 日志示例

```
[INFO] 混元3D任务轮询器已启动
[INFO] 开始轮询 3 个待处理任务
[INFO] 任务 1 (JobID: job-xxx) 已完成
[WARN] 任务 2 (JobID: job-yyy) 失败: 资源不足
[INFO] 轮询完成，处理了 3 个任务
```

## 注意事项

1. **轮询间隔**: 不要设置太短，避免API调用过于频繁
2. **并发限制**: 轮询器最多同时处理5个任务
3. **服务重启**: 修改配置后需要重启服务
4. **文件路径**: 确保NAS路径可访问
5. **静态文件服务**: 确保 `/hunyuan` 路由已配置静态文件服务

## 后续优化建议

1. 添加轮询器的启动/停止API
2. 支持动态调整轮询间隔
3. 添加任务优先级队列
4. 支持Webhook通知
5. 添加任务重试机制
6. 支持批量轮询优化

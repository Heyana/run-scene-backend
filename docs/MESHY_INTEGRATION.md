# Meshy AI 3D 集成文档

## 概述

已成功集成 Meshy AI 3D 生成平台，用户可以在混元和 Meshy 之间选择使用。

## 配置

### 1. 配置文件 (config.yaml)

```yaml
# AI 3D生成平台配置
ai3d:
  default_provider: "hunyuan" # 默认平台: hunyuan | meshy

# Meshy配置
meshy:
  api_key: "" # 从 https://www.meshy.ai 获取
  base_url: "https://api.meshy.ai"

  # 默认参数
  default_enable_pbr: true
  default_should_remesh: true
  default_should_texture: true
  default_save_pre_remeshed: true
  default_result_format: "GLB"

  # 任务控制
  max_concurrent: 3
  poll_interval: 5
  task_timeout: 86400

  # 存储配置
  local_storage_enabled: false
  storage_dir: "static/meshy"
  base_url_cdn: "http://192.168.3.39:23359/meshy"

  # NAS配置
  nas_enabled: true
  nas_path: "\\\\192.168.3.10\\project\\editor_v2\\static\\meshy"

  # 其他
  default_category: "AI生成"
  max_retry_times: 3
  retry_interval: 10
```

### 2. 环境变量

```bash
# Meshy API Key
MESHY_API_KEY=your_api_key_here

# 默认平台
AI3D_DEFAULT_PROVIDER=meshy
```

## API 使用

### 统一接口 (推荐)

使用 `/api/ai3d/*` 接口，通过 `provider` 参数选择平台：

#### 提交任务

```bash
POST /api/ai3d/tasks
Content-Type: application/json

{
  "provider": "meshy",  # hunyuan | meshy
  "inputType": "image",
  "imageUrl": "https://example.com/image.jpg",
  "enablePbr": true,
  "shouldRemesh": true,
  "shouldTexture": true,
  "name": "我的模型"
}
```

#### 查询任务

```bash
GET /api/ai3d/tasks/:id?provider=meshy
```

#### 任务列表

```bash
GET /api/ai3d/tasks?provider=meshy&page=1&pageSize=20
```

#### 轮询任务

```bash
POST /api/ai3d/tasks/:id/poll?provider=meshy
```

#### 删除任务

```bash
DELETE /api/ai3d/tasks/:id?provider=meshy
```

#### 获取配置

```bash
GET /api/ai3d/config
```

返回：

```json
{
  "defaultProvider": "hunyuan",
  "providers": {
    "hunyuan": {
      "enabled": true
    },
    "meshy": {
      "enabled": true
    }
  }
}
```

## 前端集成示例

```typescript
// 提交任务
const submitTask = async (
  imageUrl: string,
  provider: "hunyuan" | "meshy" = "meshy",
) => {
  const response = await fetch("/api/ai3d/tasks", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      provider,
      inputType: "image",
      imageUrl,
      enablePbr: true,
      shouldRemesh: true,
      shouldTexture: true,
      name: "我的3D模型",
    }),
  });

  return await response.json();
};

// 轮询任务状态
const pollTask = async (taskId: number, provider: string) => {
  const response = await fetch(
    `/api/ai3d/tasks/${taskId}/poll?provider=${provider}`,
    {
      method: "POST",
    },
  );

  return await response.json();
};

// 获取任务列表
const getTasks = async (provider: string, page = 1) => {
  const response = await fetch(
    `/api/ai3d/tasks?provider=${provider}&page=${page}&pageSize=20`,
  );
  return await response.json();
};
```

## 数据库

Meshy 任务存储在 `meshy_tasks` 表中，包含以下字段：

- `id`: 主键
- `task_id`: Meshy 任务ID
- `status`: 任务状态 (PENDING, IN_PROGRESS, SUCCEEDED, FAILED)
- `progress`: 进度 (0-100)
- `image_url`: 输入图片URL
- `enable_pbr`: 是否启用PBR
- `should_remesh`: 是否重网格化
- `should_texture`: 是否生成纹理
- `model_url`: 生成的模型URL
- `local_path`: 本地存储路径
- `nas_path`: NAS存储路径
- `file_size`: 文件大小
- `file_hash`: 文件哈希
- `name`: 任务名称
- `category`: 分类
- `created_by`: 创建者
- `created_at`: 创建时间

## 特性对比

| 特性       | 混元3D  | Meshy            |
| ---------- | ------- | ---------------- |
| 文生3D     | ✅      | ❌               |
| 图生3D     | ✅      | ✅               |
| PBR材质    | ✅      | ✅               |
| 自定义面数 | ✅      | ❌               |
| 重网格化   | ❌      | ✅               |
| 多格式导出 | GLB/OBJ | GLB/FBX/USDZ/OBJ |

## 注意事项

1. Meshy 需要有效的 API Key
2. Meshy 目前仅支持图生3D
3. 任务轮询间隔建议设置为 5 秒
4. 文件会自动下载到配置的存储位置
5. 支持本地存储和 NAS 存储

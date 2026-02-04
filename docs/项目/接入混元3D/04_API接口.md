# 混元3D接入 - API接口

## 路由定义

```
POST   /api/hunyuan/tasks              提交任务
GET    /api/hunyuan/tasks              任务列表
GET    /api/hunyuan/tasks/:id          查询任务
POST   /api/hunyuan/tasks/:id/poll     轮询任务
POST   /api/hunyuan/tasks/:id/cancel   取消任务
POST   /api/hunyuan/tasks/:id/retry    重试任务
GET    /api/hunyuan/statistics         统计信息

GET    /api/hunyuan/config             获取配置
PUT    /api/hunyuan/config             更新配置
POST   /api/hunyuan/config/validate    验证配置
```

## 接口详情

### 1. 提交任务

**请求**

```
POST /api/hunyuan/tasks
Content-Type: application/json

{
  "inputType": "text",           // text/image/multi_view
  "prompt": "一只可爱的小猫",
  "imageUrl": "",                // 图生3D时使用
  "imageBase64": "",             // 或使用base64
  "multiViewImages": [],         // 多视角图片

  "model": "3.0",                // 可选，默认3.0
  "faceCount": 500000,           // 可选，默认500000
  "generateType": "Normal",      // 可选，Normal/LowPoly/Geometry/Sketch
  "enablePbr": false,            // 可选，默认false
  "resultFormat": "STL"          // 可选，额外格式
}
```

**响应**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "jobId": "1357237233311637504",
    "status": "WAIT",
    "createdAt": "2026-02-04T10:00:00Z"
  }
}
```

### 2. 查询任务

**请求**

```
GET /api/hunyuan/tasks/:id
```

**响应**

```json
{
  "code": 0,
  "data": {
    "id": 1,
    "jobId": "1357237233311637504",
    "status": "DONE",
    "inputType": "text",
    "prompt": "一只可爱的小猫",
    "model": "3.0",
    "faceCount": 500000,
    "resultFiles": [
      {
        "type": "GLB",
        "url": "https://cos.ap-guangzhou.tencentcos.cn/xxx.glb",
        "previewImageUrl": "https://cos.ap-guangzhou.tencentcos.cn/xxx.png"
      }
    ],
    "modelId": 123,
    "createdAt": "2026-02-04T10:00:00Z",
    "updatedAt": "2026-02-04T10:05:00Z"
  }
}
```

### 3. 任务列表

**请求**

```
GET /api/hunyuan/tasks?page=1&pageSize=20&status=DONE&inputType=text
```

**响应**

```json
{
  "code": 0,
  "data": {
    "list": [...],
    "total": 100,
    "page": 1,
    "pageSize": 20
  }
}
```

### 4. 轮询任务

**请求**

```
POST /api/hunyuan/tasks/:id/poll
```

**响应**

```json
{
  "code": 0,
  "message": "任务已完成",
  "data": {
    "status": "DONE",
    "modelId": 123
  }
}
```

### 5. 取消任务

**请求**

```
POST /api/hunyuan/tasks/:id/cancel
```

**响应**

```json
{
  "code": 0,
  "message": "任务已取消"
}
```

### 6. 重试任务

**请求**

```
POST /api/hunyuan/tasks/:id/retry
```

**响应**

```json
{
  "code": 0,
  "data": {
    "jobId": "1357237233311637505",
    "status": "WAIT"
  }
}
```

### 7. 统计信息

**请求**

```
GET /api/hunyuan/statistics
```

**响应**

```json
{
  "code": 0,
  "data": {
    "total": 100,
    "waiting": 2,
    "running": 1,
    "done": 90,
    "failed": 7,
    "todayCount": 15,
    "successRate": 0.93
  }
}
```

### 8. 获取配置

**请求**

```
GET /api/hunyuan/config
```

**响应**

```json
{
  "code": 0,
  "data": {
    "secretId": "AKIDxxxxx",
    "secretKey": "******",
    "region": "ap-guangzhou",
    "defaultModel": "3.0",
    "defaultFaceCount": 500000,
    "maxConcurrent": 3,
    "pollInterval": 5,
    "autoSaveToLibrary": true
  }
}
```

### 9. 更新配置

**请求**

```
PUT /api/hunyuan/config
Content-Type: application/json

{
  "secretId": "AKIDxxxxx",
  "secretKey": "xxxxxxxx",
  "region": "ap-guangzhou",
  "defaultModel": "3.1",
  "maxConcurrent": 5
}
```

**响应**

```json
{
  "code": 0,
  "message": "配置已更新"
}
```

### 10. 验证配置

**请求**

```
POST /api/hunyuan/config/validate
Content-Type: application/json

{
  "secretId": "AKIDxxxxx",
  "secretKey": "xxxxxxxx",
  "region": "ap-guangzhou"
}
```

**响应**

```json
{
  "code": 0,
  "message": "配置有效",
  "data": {
    "valid": true
  }
}
```

## 错误码

| 错误码 | 说明              |
| ------ | ----------------- |
| 400    | 参数错误          |
| 401    | 未配置API密钥     |
| 403    | API密钥无效       |
| 404    | 任务不存在        |
| 409    | 并发限制已满      |
| 500    | 服务器错误        |
| 502    | 腾讯云API调用失败 |

## 请求限制

- 单次请求大小：≤10MB
- 图片大小：≤6MB（base64）或 ≤8MB（URL）
- 图片分辨率：128-5000像素
- 提示词长度：≤1024字符

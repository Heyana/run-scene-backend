# API 使用文档

## API 结构

项目使用模块化的 API 结构，所有 API 调用通过 `apiManager` 统一管理。

### 目录结构

```
src/api/
├── http.ts              # Axios 实例和拦截器配置
├── api.ts               # API 模块导出
├── const.ts             # 常量配置
└── models/              # API 模型定义
    ├── texture.ts       # 材质相关 API
    ├── tag.ts           # 标签相关 API
    ├── backup.ts        # 备份相关 API
    ├── security.ts      # 安全相关 API
    └── system.ts        # 系统相关 API
```

## 使用方式

### 1. 在组件中使用

```typescript
import { apiManager } from "@/api/http";

// 获取材质列表
const response = await apiManager.api.texture.getTextureList({
  page: 1,
  page_size: 10,
  keyword: "搜索关键词",
});
```

### 2. 在 Store 中使用

```typescript
import { defineStore } from "pinia";
import { apiManager } from "@/api/http";

export const useTextureStore = defineStore("texture", () => {
  const fetchTextures = async (params) => {
    const response = await apiManager.api.texture.getTextureList(params);
    return response.data;
  };

  return { fetchTextures };
});
```

## API 模块

### Texture API (材质)

```typescript
// 获取材质列表
apiManager.api.texture.getTextureList(params);

// 获取材质详情
apiManager.api.texture.getTextureDetail(assetId);

// 记录材质使用
apiManager.api.texture.recordTextureUse(assetId);

// 触发同步
apiManager.api.texture.triggerSync(data);

// 获取同步进度
apiManager.api.texture.getSyncProgress();

// 获取同步状态
apiManager.api.texture.getSyncStatus(logId);

// 获取同步日志
apiManager.api.texture.getSyncLogs(params);
```

### Tag API (标签)

```typescript
// 获取标签列表
apiManager.api.tag.getTagList(params);

// 根据标签获取材质
apiManager.api.tag.getTexturesByTag(tagId, params);
```

### Backup API (备份)

```typescript
// 获取备份状态
apiManager.api.backup.getBackupStatus();

// 触发手动备份
apiManager.api.backup.triggerManualBackup();

// 触发数据库备份
apiManager.api.backup.triggerDatabaseBackup();

// 触发 CDN 备份
apiManager.api.backup.triggerCDNBackup();

// 获取备份历史
apiManager.api.backup.getBackupHistory(params);

// 从备份恢复
apiManager.api.backup.restoreCDNFromBackup(backupId);
```

### Security API (安全)

```typescript
// 获取安全状态
apiManager.api.security.getSecurityStatus();

// 获取被封禁 IP 列表
apiManager.api.security.getBlockedIPs();

// 解封 IP
apiManager.api.security.unblockIP(ip);

// 获取 IP 统计
apiManager.api.security.getIPStats(params);

// 封禁 IP
apiManager.api.security.blockIP(ip, data);

// 添加到白名单
apiManager.api.security.addToWhitelist(ip);

// 从白名单移除
apiManager.api.security.removeFromWhitelist(ip);

// 获取连接统计
apiManager.api.security.getConnections();
```

### System API (系统)

```typescript
// 健康检查
apiManager.api.system.ping();

// 服务健康检查
apiManager.api.system.health();
```

## 响应格式

所有 API 响应遵循统一格式：

```typescript
{
  code: number,      // 状态码 (0 表示成功)
  message: string,   // 响应消息
  data: any          // 响应数据
}
```

## 错误处理

HTTP 拦截器会自动处理错误：

- 401: 自动清除 token 并提示重新登录
- 其他错误: 显示错误消息

在代码中使用 try-catch 捕获错误：

```typescript
try {
  const response = await apiManager.api.texture.getTextureList();
  // 处理成功响应
} catch (error) {
  // 处理错误
  console.error("请求失败:", error);
}
```

## 配置

### API 基础地址

在 `src/api/const.ts` 中配置：

```typescript
const getBaseUrl = () => {
  return isDev ? "http://192.168.3.39:23356/" : pubUrl;
};
```

### Token 管理

Token 存储在 localStorage 中：

```typescript
localStorage.getItem(constant.clientToken);
localStorage.setItem(constant.clientToken, token);
localStorage.removeItem(constant.clientToken);
```

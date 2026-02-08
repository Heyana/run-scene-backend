# 资源统计 API 设计

## 概述

为首页提供统一的资源统计接口，包括贴图、项目、模型、资产的数量统计和趋势分析。

## API 端点

### 1. 获取资源统计概览

**接口**: `GET /api/statistics/overview`

**响应**:

```json
{
  "textures": {
    "total": 1234,
    "trend": 12.5,
    "recent_count": 45
  },
  "projects": {
    "total": 56,
    "trend": 8.3,
    "recent_count": 3
  },
  "models": {
    "total": 789,
    "trend": 15.2,
    "recent_count": 28
  },
  "assets": {
    "total": 432,
    "trend": -3.1,
    "recent_count": 12
  }
}
```

**字段说明**:

- `total`: 资源总数
- `trend`: 本月增长率（百分比，负数表示下降）
- `recent_count`: 最近7天新增数量

---

### 2. 获取最近活动

**接口**: `GET /api/statistics/recent-activities`

**参数**:

- `limit`: 返回数量，默认 10

**响应**:

```json
{
  "activities": [
    {
      "id": 1,
      "type": "texture",
      "name": "wood_floor_01",
      "action": "upload",
      "user": "admin",
      "created_at": "2026-02-08T10:30:00Z"
    },
    {
      "id": 2,
      "type": "project",
      "name": "3d-editor",
      "action": "version_upload",
      "version": "1.0.7",
      "user": "hxy",
      "created_at": "2026-02-08T09:15:00Z"
    }
  ]
}
```

**type 类型**:

- `texture`: 贴图
- `project`: 项目
- `model`: 模型
- `asset`: 资产

**action 类型**:

- `upload`: 上传
- `update`: 更新
- `delete`: 删除
- `version_upload`: 版本上传（项目专用）

---

### 3. 获取系统状态

**接口**: `GET /api/statistics/system-status`

**响应**:

```json
{
  "service": {
    "status": "running",
    "uptime": 86400
  },
  "database": {
    "status": "healthy",
    "size": 524288000
  },
  "storage": {
    "total": 1099511627776,
    "used": 824633720832,
    "usage_percent": 75.0
  },
  "sync": {
    "last_sync_at": "2026-02-08T13:00:00Z",
    "status": "success"
  }
}
```

**字段说明**:

- `service.uptime`: 运行时间（秒）
- `database.size`: 数据库大小（字节）
- `storage.*`: 存储空间信息（字节）
- `sync.last_sync_at`: 最后同步时间

---

## 后端实现

### 控制器

**文件**: `controllers/statistics_controller.go`

```go
type StatisticsController struct {
    db *gorm.DB
}

// GetOverview 获取资源统计概览
func (sc *StatisticsController) GetOverview(c *gin.Context)

// GetRecentActivities 获取最近活动
func (sc *StatisticsController) GetRecentActivities(c *gin.Context)

// GetSystemStatus 获取系统状态
func (sc *StatisticsController) GetSystemStatus(c *gin.Context)
```

---

### 服务层

**文件**: `services/statistics_service.go`

```go
type StatisticsService struct {
    db *gorm.DB
}

// GetResourceCounts 获取各类资源总数
func (ss *StatisticsService) GetResourceCounts() (map[string]int64, error)

// CalculateTrend 计算本月增长趋势
func (ss *StatisticsService) CalculateTrend(resourceType string) (float64, error)

// GetRecentCount 获取最近7天新增数量
func (ss *StatisticsService) GetRecentCount(resourceType string) (int64, error)

// GetRecentActivities 获取最近活动记录
func (ss *StatisticsService) GetRecentActivities(limit int) ([]Activity, error)

// GetStorageInfo 获取存储空间信息
func (ss *StatisticsService) GetStorageInfo() (StorageInfo, error)
```

---

### 数据模型

**文件**: `models/activity.go`

```go
type Activity struct {
    ID        uint      `json:"id"`
    Type      string    `json:"type"`      // texture/project/model/asset
    Name      string    `json:"name"`
    Action    string    `json:"action"`    // upload/update/delete
    User      string    `json:"user"`
    Version   string    `json:"version"`   // 项目版本号（可选）
    CreatedAt time.Time `json:"created_at"`
}
```

**说明**:

- 新增活动记录表，用于追踪所有资源操作
- 在各资源上传/更新/删除时插入记录

---

### 路由注册

**文件**: `api/routes.go`

```go
statistics := api.Group("/statistics")
{
    statistics.GET("/overview", statisticsController.GetOverview)
    statistics.GET("/recent-activities", statisticsController.GetRecentActivities)
    statistics.GET("/system-status", statisticsController.GetSystemStatus)
}
```

---

## 实现要点

### 1. 趋势计算

```go
// 计算本月增长率
currentMonth := time.Now().Month()
lastMonth := time.Now().AddDate(0, -1, 0).Month()

currentCount := // 本月数量
lastCount := // 上月数量

trend := ((currentCount - lastCount) / lastCount) * 100
```

### 2. 存储空间获取

- **本地存储**: 使用 `syscall.Statfs` (Linux) 或 `GetDiskFreeSpaceEx` (Windows)
- **NAS 存储**: 读取 NAS 配置路径，统计目录大小

### 3. 活动记录触发点

在以下操作时插入活动记录：

- 贴图同步完成
- 项目版本上传
- 模型上传
- 资产上传/更新/删除

### 4. 缓存优化

统计数据可缓存 5 分钟，减少数据库查询：

```go
var statsCache *StatisticsCache
var cacheExpiry time.Time

if time.Now().Before(cacheExpiry) {
    return statsCache, nil
}
```

---

## 前端集成

**文件**: `frontend/src/api/statistics.ts`

```typescript
export interface ResourceStats {
  total: number;
  trend: number;
  recent_count: number;
}

export interface OverviewResponse {
  textures: ResourceStats;
  projects: ResourceStats;
  models: ResourceStats;
  assets: ResourceStats;
}

export const getOverview = () => {
  return http.get<OverviewResponse>("/statistics/overview");
};

export const getRecentActivities = (limit?: number) => {
  return http.get("/statistics/recent-activities", { params: { limit } });
};

export const getSystemStatus = () => {
  return http.get("/statistics/system-status");
};
```

---

## 数据库变更

### 新增表: `activity`

```sql
CREATE TABLE activity (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    type VARCHAR(20) NOT NULL,
    name VARCHAR(200) NOT NULL,
    action VARCHAR(20) NOT NULL,
    user VARCHAR(100),
    version VARCHAR(20),
    created_at DATETIME NOT NULL
);

CREATE INDEX idx_activity_type ON activity(type);
CREATE INDEX idx_activity_created_at ON activity(created_at);
```

---

## 测试要点

1. **统计准确性**: 验证各资源数量统计正确
2. **趋势计算**: 验证增长率计算逻辑
3. **活动记录**: 验证操作触发活动记录插入
4. **性能**: 验证缓存机制生效
5. **存储空间**: 验证不同存储类型（本地/NAS）的空间统计

---

## 后续扩展

1. **图表数据**: 提供时间序列数据用于趋势图
2. **分类统计**: 按类型、标签等维度统计
3. **用户统计**: 各用户的上传/使用统计
4. **热门资源**: 统计使用频率最高的资源
5. **存储分析**: 各类资源的存储占用分析

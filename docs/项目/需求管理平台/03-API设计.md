# 需求管理平台 - API设计

## 公司管理

### 创建公司

```
POST /api/companies
Body: { name, logo?, description? }
Response: Company
```

### 获取公司列表

```
GET /api/companies?page=1&page_size=20
Response: { items: Company[], total, page, page_size }
```

### 获取公司详情

```
GET /api/companies/:id
Response: Company
```

### 更新公司

```
PUT /api/companies/:id
Body: { name?, logo?, description? }
Response: Company
```

### 添加成员

```
POST /api/companies/:id/members
Body: { user_id, role }
Response: CompanyMember
```

### 移除成员

```
DELETE /api/companies/:id/members/:user_id
Response: { success: true }
```

## 项目管理

### 创建项目

```
POST /api/companies/:company_id/projects
Body: { name, key, description?, start_date?, end_date? }
Response: Project
```

### 获取项目列表

```
GET /api/companies/:company_id/projects?status=active
Response: { items: Project[], total }
```

### 获取项目详情

```
GET /api/projects/:id
Response: Project (包含统计信息)
```

### 更新项目

```
PUT /api/projects/:id
Body: { name?, description?, owner_id?, status?, start_date?, end_date? }
Response: Project
```

### 项目成员管理

```
POST /api/projects/:id/members
Body: { user_id, role }

DELETE /api/projects/:id/members/:user_id

GET /api/projects/:id/members
Response: { items: ProjectMember[] }
```

## 需求列表管理

### 创建需求列表

```
POST /api/projects/:project_id/mission-lists
Body: { name, type, description?, start_date?, end_date? }
Response: MissionList
```

### 获取需求列表

```
GET /api/projects/:project_id/mission-lists?status=active
Response: { items: MissionList[] }
```

### 更新需求列表

```
PUT /api/mission-lists/:id
Body: { name?, description?, status?, start_date?, end_date? }
Response: MissionList
```

### 删除需求列表

```
DELETE /api/mission-lists/:id
Response: { success: true }
```

## 需求管理

### 创建需求

```
POST /api/mission-lists/:list_id/missions
Body: {
  title, description?, type, priority,
  assignee_id?, estimated_hours?,
  start_date?, due_date?
}
Response: Mission
```

### 获取需求列表

```
GET /api/mission-lists/:list_id/missions?status=&priority=&assignee_id=
Response: { items: Mission[], total }
```

### 获取需求详情

```
GET /api/missions/:id
Response: Mission (包含评论、附件、关联需求)
```

### 更新需求

```
PUT /api/missions/:id
Body: {
  title?, description?, type?, priority?,
  status?, assignee_id?, estimated_hours?,
  actual_hours?, start_date?, due_date?
}
Response: Mission
```

### 删除需求

```
DELETE /api/missions/:id
Response: { success: true }
```

### 批量更新状态

```
PATCH /api/missions/batch-update-status
Body: { mission_ids: uint[], status: string }
Response: { success: true, updated_count: int }
```

## 需求评论

### 添加评论

```
POST /api/missions/:id/comments
Body: { content, parent_id? }
Response: Comment
```

### 获取评论列表

```
GET /api/missions/:id/comments
Response: { items: Comment[] }
```

### 删除评论

```
DELETE /api/comments/:id
Response: { success: true }
```

## 需求附件

### 上传附件

```
POST /api/missions/:id/attachments
Body: FormData (file)
Response: Attachment
```

### 删除附件

```
DELETE /api/attachments/:id
Response: { success: true }
```

## 需求关联

### 创建关联

```
POST /api/missions/:id/relations
Body: { target_mission_id, relation_type }
Response: MissionRelation
```

### 删除关联

```
DELETE /api/mission-relations/:id
Response: { success: true }
```

## 统计报表

### 项目统计

```
GET /api/projects/:id/statistics
Response: {
  total_missions: int,
  completed_missions: int,
  in_progress_missions: int,
  completion_rate: float,
  by_priority: { P0: int, P1: int, ... },
  by_type: { feature: int, bug: int, ... }
}
```

### 成员工作量

```
GET /api/projects/:id/workload?start_date=&end_date=
Response: {
  items: [{
    user_id, user_name,
    assigned_count, completed_count,
    total_hours
  }]
}
```

### 燃尽图数据

```
GET /api/mission-lists/:id/burndown
Response: {
  dates: string[],
  ideal: int[],
  actual: int[]
}
```

## 权限说明

- 公司管理员: 公司内所有操作
- 项目管理员: 项目内所有操作
- 开发人员: 创建/更新自己的需求，评论
- 观察者: 只读权限

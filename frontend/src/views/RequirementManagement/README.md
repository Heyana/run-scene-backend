# 需求管理平台 - 前端

## 功能概览

需求管理平台是一个类似飞书任务管理的系统，支持多公司、多项目的需求管理。

## 已实现功能

### 1. 公司管理

- **路由**: `/requirement-management/companies`
- **功能**:
  - 查看公司列表
  - 创建新公司
  - 查看公司详情
  - 管理公司成员
  - 查看公司下的项目

### 2. 公司详情（成员管理）

- **路由**: `/requirement-management/companies/:companyId`
- **功能**:
  - 查看公司成员列表
  - 添加成员到公司
  - 设置成员角色（管理员/成员/访客）
  - 删除成员

### 3. 项目管理

- **路由**:
  - 所有项目: `/requirement-management/projects`
  - 公司项目: `/requirement-management/companies/:companyId/projects`
- **功能**:
  - 查看项目列表（卡片式布局）
  - 创建新项目
  - 查看项目统计（成员数、任务数、完成度）
  - 进入任务看板
  - 查看统计报表

### 4. 任务看板

- **路由**: `/requirement-management/projects/:projectId/board`
- **功能**:
  - 看板式任务展示（待处理/进行中/已完成/已关闭）
  - 创建任务
  - 查看任务详情
  - 切换任务列表（Sprint/版本/模块）
  - 刷新任务列表

### 5. 任务详情

- **组件**: 右侧抽屉式面板
- **功能**:
  - 编辑任务信息（标题、描述、类型、优先级、状态）
  - 设置截止日期
  - 添加评论
  - 上传附件
  - 查看任务历史

### 6. 统计报表

- **路由**: `/requirement-management/projects/:projectId/statistics`
- **功能**:
  - 任务总览统计（总数、已完成、进行中、逾期）
  - 完成率展示
  - 按类型分布（功能/优化/缺陷）
  - 按优先级分布（P0/P1/P2/P3）
  - 成员任务分布表格

## 组件结构

```
views/RequirementManagement/
├── index.tsx                 # 主入口（左侧导航）
├── CompanyList.tsx          # 公司列表
├── CompanyDetail.tsx        # 公司详情（成员管理）
├── ProjectList.tsx          # 项目列表
├── MissionBoard.tsx         # 任务看板
├── MissionDetail.tsx        # 任务详情
└── Statistics.tsx           # 统计报表

components/RequirementManagement/
├── StatusTag.tsx            # 状态标签
├── PriorityTag.tsx          # 优先级标签
└── MissionCard.tsx          # 任务卡片

api/
└── requirement.ts           # API 接口封装

types/
└── requirement.ts           # TypeScript 类型定义
```

## 数据流

1. **公司** → **项目** → **任务列表** → **任务**
2. 每个层级都有独立的成员管理
3. 权限继承：公司权限 → 项目权限

## 待实现功能

### 高优先级

- [ ] 任务看板拖拽功能（使用 vue-draggable-plus）
- [ ] 项目成员管理页面
- [ ] 任务关联功能（阻塞/被阻塞/关联）
- [ ] 实时更新（WebSocket）

### 中优先级

- [ ] 任务标签管理
- [ ] 高级筛选和搜索
- [ ] 批量操作任务
- [ ] 导出功能（Excel/PDF）
- [ ] 燃尽图

### 低优先级

- [ ] 任务模板
- [ ] 自定义字段
- [ ] 工作流自定义
- [ ] 邮件通知
- [ ] 移动端适配

## API 集成

所有 API 调用都在 `api/requirement.ts` 中定义，目前使用模拟数据。

需要替换的地方：

```typescript
// TODO: 调用API
// const res = await companyApi.getList();
// companies.value = res.data.items;

// 模拟数据
companies.value = [...];
```

## 样式设计

参考飞书任务管理风格：

- 简洁的卡片式布局
- 柔和的色彩系统
- 清晰的层级结构
- 流畅的交互动画

## 使用方式

1. 访问 `/requirement-management` 自动跳转到公司列表
2. 创建公司 → 添加成员 → 创建项目 → 创建任务
3. 在任务看板中管理任务状态
4. 在统计报表中查看项目进度

## 技术栈

- Vue 3 + TypeScript
- Ant Design Vue
- Vue Router
- Pinia（可选，用于状态管理）

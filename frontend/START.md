# 快速启动指南

## 项目已完成改造 ✅

- ✅ Vue 3 + TypeScript + TSX
- ✅ Ant Design Vue 组件库
- ✅ Vue Router 路由管理
- ✅ Pinia 状态管理
- ✅ 完整的 API 结构

## 启动项目

```bash
# 进入前端目录
cd frontend

# 启动开发服务器
yarn dev
```

访问: http://localhost:3000

## 项目结构

```
frontend/
├── src/
│   ├── api/              # API 接口
│   │   ├── http.ts       # Axios 配置
│   │   ├── api.ts        # API 导出
│   │   ├── const.ts      # 常量配置
│   │   └── models/       # API 模型
│   │       ├── texture.ts    # 材质 API
│   │       ├── tag.ts        # 标签 API
│   │       ├── backup.ts     # 备份 API
│   │       ├── security.ts   # 安全 API
│   │       └── system.ts     # 系统 API
│   ├── components/       # 组件
│   │   ├── Layout.tsx    # 布局组件
│   │   ├── Button.tsx    # 按钮组件
│   │   └── TextureCard.tsx # 材质卡片
│   ├── views/            # 页面
│   │   ├── Home.tsx      # 首页
│   │   ├── Textures.tsx  # 材质列表
│   │   └── About.tsx     # 关于页
│   ├── stores/           # 状态管理
│   │   ├── index.ts      # Pinia 实例
│   │   ├── app.ts        # 应用状态
│   │   └── texture.ts    # 材质状态
│   ├── router/           # 路由
│   │   └── index.ts
│   ├── types/            # 类型定义
│   ├── utils/            # 工具函数
│   ├── App.tsx           # 根组件
│   ├── main.ts           # 入口文件
│   └── style.css         # 全局样式
```

## 页面路由

- `/` - 首页
- `/textures` - 材质库列表
- `/about` - 关于系统

## API 使用示例

### 在组件中使用

```tsx
import { apiManager } from "@/api/http";

// 获取材质列表
const fetchData = async () => {
  const response = await apiManager.api.texture.getTextureList({
    page: 1,
    page_size: 10,
  });
  console.log(response.data);
};
```

### 在 Store 中使用

```typescript
import { defineStore } from "pinia";
import { apiManager } from "@/api/http";

export const useMyStore = defineStore("my", () => {
  const fetchData = async () => {
    const response = await apiManager.api.texture.getTextureList();
    return response.data;
  };

  return { fetchData };
});
```

## Ant Design Vue 组件使用

### TSX 语法

```tsx
import { Button, Card, Table } from "ant-design-vue";
import { PlusOutlined } from "@ant-design/icons-vue";

export default defineComponent({
  setup() {
    return () => (
      <Card title="标题">
        <Button type="primary">
          {{
            icon: () => <PlusOutlined />,
            default: () => "添加",
          }}
        </Button>
      </Card>
    );
  },
});
```

## 开发提示

1. **API 基础地址配置**: 在 `src/api/const.ts` 中修改
2. **主题配置**: 在 `src/stores/app.ts` 中管理
3. **路由配置**: 在 `src/router/index.ts` 中添加新路由
4. **全局样式**: 在 `src/style.css` 中修改

## 构建生产版本

```bash
yarn build
```

构建产物在 `dist/` 目录

## 常见问题

### 1. API 请求失败

检查 `src/api/const.ts` 中的 API 基础地址是否正确

### 2. 组件样式问题

确保已导入 `ant-design-vue/dist/reset.css`

### 3. 路由不工作

检查 `src/router/index.ts` 配置和 `App.tsx` 中的 RouterView

## 更多文档

- [API 使用文档](./API_USAGE.md)
- [项目说明](./README.md)
- [快速开始](./QUICKSTART.md)

# 快速开始

## 项目已完成改造

✅ Vue 3 + TSX 架构
✅ Vue Router 路由管理
✅ Pinia 状态管理
✅ TypeScript 类型支持

## 启动项目

```bash
cd frontend
yarn dev
```

访问: http://localhost:3000

## 项目结构说明

### 路由 (Router)

- `/` - 首页
- `/textures` - 材质库列表
- `/about` - 关于页面

### 状态管理 (Pinia Stores)

- `app.ts` - 应用全局状态（主题、侧边栏等）
- `texture.ts` - 材质数据管理

### 组件 (TSX Components)

- `Button.tsx` - 按钮组件
- `TextureCard.tsx` - 材质卡片组件
- `Layout.tsx` - 布局组件

### 视图 (Views)

- `Home.tsx` - 首页
- `Textures.tsx` - 材质列表页
- `About.tsx` - 关于页

## TSX 语法示例

```tsx
import { defineComponent } from "vue";

export default defineComponent({
  name: "MyComponent",
  setup() {
    const handleClick = () => {
      console.log("clicked");
    };

    return () => (
      <div class="my-component">
        <h1>Hello TSX</h1>
        <button onClick={handleClick}>Click me</button>
      </div>
    );
  },
});
```

## API 请求

使用封装的 `api` 工具：

```typescript
import { api } from "@/utils/request";

// GET 请求
const data = await api.get("/textures");

// POST 请求
await api.post("/textures", { name: "texture1" });
```

## 下一步

1. 根据后端 API 调整 `stores/texture.ts` 中的数据结构
2. 在 `views/Textures.tsx` 中完善材质列表功能
3. 添加更多页面和功能
4. 自定义样式主题

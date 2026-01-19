# Texture Library Frontend

Vue 3 + TypeScript + TSX 项目

## 技术栈

- **Vue 3** - 渐进式 JavaScript 框架
- **TypeScript** - 类型安全
- **TSX** - TypeScript + JSX 语法
- **Vue Router** - 路由管理
- **Pinia** - 状态管理
- **Vite** - 构建工具

## 项目结构

```
src/
├── components/       # TSX 组件
│   ├── Button.tsx
│   ├── Layout.tsx
│   └── TextureCard.tsx
├── views/           # 页面组件
│   ├── Home.tsx
│   ├── Textures.tsx
│   └── About.tsx
├── stores/          # Pinia 状态管理
│   ├── index.ts
│   ├── app.ts
│   └── texture.ts
├── router/          # 路由配置
│   └── index.ts
├── utils/           # 工具函数
│   └── request.ts
├── types/           # TypeScript 类型定义
│   └── index.ts
├── App.tsx          # 根组件
├── main.ts          # 入口文件
└── style.css        # 全局样式
```

## 开发命令

```bash
# 安装依赖
yarn

# 启动开发服务器
yarn dev

# 构建生产版本
yarn build

# 预览生产构建
yarn preview
```

## 特性

- ✅ Vue 3 Composition API
- ✅ TSX 语法支持
- ✅ Vue Router 路由管理
- ✅ Pinia 状态管理
- ✅ TypeScript 类型检查
- ✅ 响应式布局
- ✅ API 请求封装

## API 代理

开发环境下，所有 `/api` 请求会被代理到 `http://localhost:8080`

## 浏览器支持

现代浏览器和 IE11+

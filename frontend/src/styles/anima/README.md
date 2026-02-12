# 动画系统使用指南

参考 Ant Design 动画系统设计，提供统一的动画变量、关键帧和工具类。

## 文件结构

```
anima/
├── index.css          # 入口文件，导入所有动画模块
├── variables.css      # 动画变量定义（时长、缓动函数）
├── keyframes.css      # 动画关键帧定义
├── utilities.css      # 动画工具类
└── README.md          # 使用文档
```

## 快速开始

### 1. 导入动画系统

```typescript
import "@/styles/anima/index.css";
```

### 2. 使用工具类

```tsx
// 淡入效果
<div class="fade-in">内容</div>

// 缩放进入
<div class="zoom-in">内容</div>

// 滑动进入
<div class="slide-up-in">内容</div>

// 弹出效果
<div class="popup-in">内容</div>
```

## 动画变量

### 时长变量

```css
--motion-duration-fast: 0.1s; /* 快速 */
--motion-duration-mid: 0.2s; /* 中等 */
--motion-duration-base: 0.3s; /* 基础 */
--motion-duration-slow: 0.4s; /* 慢速 */
```

### 缓动函数

```css
/* 标准缓动 */
--motion-ease-in-out: cubic-bezier(0.645, 0.045, 0.355, 1);
--motion-ease-out: cubic-bezier(0.215, 0.61, 0.355, 1);
--motion-ease-in: cubic-bezier(0.55, 0.055, 0.675, 0.19);

/* 快速缓动 */
--motion-ease-in-out-circ: cubic-bezier(0.78, 0.14, 0.15, 0.86);
--motion-ease-out-circ: cubic-bezier(0.08, 0.82, 0.17, 1);

/* 回弹效果 */
--motion-ease-out-back: cubic-bezier(0.12, 0.4, 0.29, 1.46);
```

### 使用示例

```css
.my-element {
  transition: all var(--motion-duration-mid) var(--motion-ease-in-out);
}

.my-animation {
  animation: fadeIn var(--fade-duration) var(--fade-easing);
}
```

## 动画工具类

### 淡入淡出

- `.fade-in` - 淡入
- `.fade-out` - 淡出

### 缩放

- `.zoom-in` - 缩放进入
- `.zoom-out` - 缩放退出
- `.zoom-big-in` - 大幅缩放进入

### 滑动

- `.slide-up-in` / `.slide-up-out` - 向上滑动
- `.slide-down-in` / `.slide-down-out` - 向下滑动
- `.slide-left-in` / `.slide-left-out` - 向左滑动
- `.slide-right-in` / `.slide-right-out` - 向右滑动

### 弹出

- `.popup-in` - 弹出进入
- `.popup-out` - 弹出退出

### 特殊效果

- `.rotate` - 旋转（循环）
- `.shake` - 摇晃
- `.pulse` - 脉冲
- `.bounce` - 弹跳
- `.breathe` - 呼吸灯
- `.blink` - 闪烁

### 过渡工具类

- `.transition-all` - 所有属性过渡
- `.transition-fast` - 快速过渡
- `.transition-slow` - 慢速过渡
- `.transition-transform` - 变换过渡
- `.transition-opacity` - 透明度过渡
- `.transition-colors` - 颜色过渡

### 延迟工具类

- `.delay-100` - 延迟 0.1s
- `.delay-200` - 延迟 0.2s
- `.delay-300` - 延迟 0.3s
- `.delay-400` - 延迟 0.4s
- `.delay-500` - 延迟 0.5s

## 自定义动画

### 使用关键帧

```css
.my-custom-animation {
  animation: fadeIn 0.3s cubic-bezier(0.645, 0.045, 0.355, 1);
}
```

### 组合动画

```tsx
<div class="zoom-in delay-200">延迟 0.2s 后缩放进入</div>
```

### 在 Vue 组件中使用

```tsx
import { defineComponent } from "vue";
import "@/styles/anima/index.css";

export default defineComponent({
  setup() {
    return () => (
      <div class="fade-in">
        <h1>标题</h1>
        <p class="slide-up-in delay-100">内容</p>
      </div>
    );
  },
});
```

## 最佳实践

1. **选择合适的时长**
   - 快速交互：使用 `--motion-duration-fast`
   - 普通动画：使用 `--motion-duration-mid`
   - 复杂动画：使用 `--motion-duration-base`

2. **选择合适的缓动**
   - 进入动画：使用 `--motion-ease-out`
   - 退出动画：使用 `--motion-ease-in`
   - 双向动画：使用 `--motion-ease-in-out`

3. **避免过度动画**
   - 不要在同一元素上使用过多动画
   - 保持动画简洁流畅
   - 考虑性能影响

4. **保持一致性**
   - 在整个应用中使用统一的动画风格
   - 相似的交互使用相似的动画

## 示例：右键菜单动画

```tsx
// 菜单容器使用缩放进入
<div class="context-menu zoom-in">
  {/* 菜单项 hover 时平移 */}
  <div class="context-menu-item transition-transform">菜单项</div>

  {/* 子菜单使用滑动进入 */}
  <div class="context-menu-submenu slide-right-in">子菜单</div>
</div>
```

## 性能优化

1. 使用 `transform` 和 `opacity` 属性（GPU 加速）
2. 避免动画 `width`、`height`、`margin` 等会触发重排的属性
3. 使用 `will-change` 提示浏览器优化（谨慎使用）

```css
.optimized-animation {
  will-change: transform, opacity;
  transform: translateZ(0); /* 开启 GPU 加速 */
}
```

# 文件预览组件

基于适配器模式的文件预览系统，支持多种文件类型的在线预览。

## 文件结构

```
FilePreview/
├── adapters/              # 预览适配器
│   ├── ImageAdapter.tsx   # 图片预览
│   ├── VideoAdapter.tsx   # 视频预览
│   ├── DocumentAdapter.tsx # 文档预览
│   ├── DefaultAdapter.tsx # 默认预览（不支持的类型）
│   └── index.ts           # 适配器注册中心
├── FilePreviewModal.tsx   # 预览模态框
├── types.ts               # 类型定义
├── style.css              # 样式
├── index.ts               # 导出
└── README.md              # 文档
```

## 支持的文件类型

### 图片 (ImageAdapter)

- jpg, jpeg, png, gif, webp, bmp, svg, ico

### 视频 (VideoAdapter)

- mp4, webm, ogg, avi, mov

### 文档 (DocumentAdapter)

- pdf - PDF 文档
- txt - 文本文件
- md - Markdown 文件

### 其他 (DefaultAdapter)

- 不支持预览的文件类型，显示下载按钮

## 使用方法

### 基础用法

```tsx
import { FilePreviewModal } from "@/components/FilePreview";

const MyComponent = () => {
  const [visible, setVisible] = useState(false);
  const [file, setFile] = useState(null);

  const handlePreview = (fileInfo) => {
    setFile(fileInfo);
    setVisible(true);
  };

  return (
    <>
      <Button onClick={() => handlePreview(someFile)}>预览文件</Button>

      <FilePreviewModal
        file={file}
        visible={visible}
        onClose={() => setVisible(false)}
      />
    </>
  );
};
```

## 扩展新的适配器

### 1. 创建适配器类

```tsx
// adapters/AudioAdapter.tsx
import type { IPreviewAdapter, PreviewAdapterProps } from "../types";

class AudioPreviewAdapter implements IPreviewAdapter {
  name = "AudioPreviewAdapter";

  canPreview(format: string): boolean {
    return ["mp3", "wav", "ogg"].includes(format.toLowerCase());
  }

  render(props: PreviewAdapterProps) {
    return <AudioPreview {...props} />;
  }
}

export default new AudioPreviewAdapter();
```

### 2. 注册适配器

```tsx
// adapters/index.ts
import AudioAdapter from "./AudioAdapter";

const adapters: IPreviewAdapter[] = [
  ImageAdapter,
  VideoAdapter,
  AudioAdapter, // 添加新适配器
  DocumentAdapter,
  DefaultAdapter,
];
```

## 设计模式

### 适配器模式

每个文件类型都有独立的适配器，负责判断是否支持和渲染预览内容。

### 优点

- 易于扩展：添加新文件类型只需创建新适配器
- 职责分离：每个适配器只负责一种文件类型
- 灵活配置：可以动态注册/注销适配器

# 文件处理器系统

## 概述

统一的文件处理系统，支持多种文件类型的元数据提取、预览图生成、格式转换等功能。

## 核心特性

- ✅ **多格式支持**: 图片、视频、文档、3D模型、压缩包
- ✅ **元数据提取**: 自动提取文件元信息
- ✅ **预览生成**: 自动生成预览图和缩略图
- ✅ **任务管理**: 支持长时间任务的管理和监控
- ✅ **进度追踪**: 实时进度更新和回调
- ✅ **并发控制**: 资源限制和优先级调度
- ✅ **错误重试**: 自动重试和任务恢复
- ✅ **通知回调**: 任务完成通知和 Webhook

## 架构设计

### 四层架构

```
┌─────────────────────────────────────────────────┐
│           业务层 (Business Layer)                │
│  services/fileprocessor/                        │
│  - 文件处理器接口和实现                          │
│  - 自动路由到对应处理器                          │
└─────────────────────────────────────────────────┘
                      ↓
┌─────────────────────────────────────────────────┐
│           任务层 (Task Layer)                    │
│  services/task/                                 │
│  - 任务管理（创建、启动、取消、查询）            │
│  - 任务队列（优先级调度）                        │
│  - 任务执行器（并发控制）                        │
└─────────────────────────────────────────────────┘
                      ↓
┌─────────────────────────────────────────────────┐
│           工具层 (Utils Layer)                   │
│  utils/                                         │
│  - FFmpeg 封装（视频处理）                       │
│  - ImageMagick 封装（图片处理）                  │
│  - PDF 工具封装（文档处理）                      │
│  - Assimp 封装（3D模型处理）                     │
└─────────────────────────────────────────────────┘
                      ↓
┌─────────────────────────────────────────────────┐
│           系统层 (System Layer)                  │
│  - FFmpeg (需要系统安装)                         │
│  - ImageMagick (需要系统安装)                    │
│  - pdftoppm (需要系统安装)                       │
└─────────────────────────────────────────────────┘
```

## 文档导航

### 核心文档

- [01-架构设计.md](./01-架构设计.md) - 整体架构和设计思路
- [02-任务管理.md](./02-任务管理.md) - 任务系统详细设计
- [03-工具封装.md](./03-工具封装.md) - 底层工具封装实现
- [04-处理器实现.md](./04-处理器实现.md) - 各类型文件处理器
- [05-高级特性.md](./05-高级特性.md) - 重试、通知、工作流等

### 使用文档

- [使用示例.md](./使用示例.md) - 代码使用示例
- [API文档.md](./API文档.md) - HTTP API 接口文档
- [配置说明.md](./配置说明.md) - 配置文件说明

### 部署文档

- [部署指南.md](./部署指南.md) - 系统部署和依赖安装
- [性能优化.md](./性能优化.md) - 性能调优建议

## 快速开始

### 1. 安装依赖

```bash
# Ubuntu/Debian
sudo apt-get install ffmpeg imagemagick poppler-utils

# macOS
brew install ffmpeg imagemagick poppler

# Windows
# 下载并安装 FFmpeg 和 ImageMagick
```

### 2. 配置

```yaml
# configs/fileprocessor.yaml
fileprocessor:
  ffmpeg:
    bin_path: "ffmpeg"
    timeout: 300
  imagemagick:
    bin_path: "convert"
    timeout: 60
```

### 3. 初始化服务

```go
// 初始化文件处理器服务
config := fileprocessor.LoadConfig()
fpService := fileprocessor.NewFileProcessorService(config)

// 初始化任务服务
taskService := task.NewTaskService(db)
```

### 4. 使用

```go
// 同步使用（简单场景）
metadata, _ := fpService.ExtractMetadata("/path/to/video.mp4", "mp4")
fmt.Printf("时长: %d秒\n", metadata.Duration)

// 异步使用（长时间任务）
task := &models.Task{
    Type:       models.TaskTypeVideoPreview,
    InputFile:  "/path/to/video.mp4",
    OutputFile: "/path/to/preview.jpg",
}
taskService.CreateTask(task)
```

## 支持的文件类型

### 图片

- jpg, jpeg, png, gif, webp, bmp, tiff, svg

### 视频

- mp4, avi, mov, webm, mkv, flv, wmv

### 文档

- pdf, doc, docx, ppt, pptx, xls, xlsx

### 3D模型

- fbx, glb, gltf, obj, stl, dae

### 压缩包

- zip, rar, 7z, tar, gz

## 实现进度

### 第一阶段（核心功能）✅

- [x] 架构设计
- [x] 核心接口定义
- [x] 任务管理系统
- [x] FFmpeg 封装
- [x] ImageMagick 封装
- [x] 视频处理器
- [x] 图片处理器

### 第二阶段（扩展功能）🚧

- [ ] 文档处理器
- [ ] 3D模型处理器
- [ ] 压缩包处理器
- [ ] 任务重试机制
- [ ] 通知系统

### 第三阶段（高级特性）📋

- [ ] 工作流支持
- [ ] 断点续传
- [ ] 分布式任务
- [ ] 性能监控
- [ ] 缓存优化

## 贡献指南

1. Fork 项目
2. 创建特性分支
3. 提交代码
4. 创建 Pull Request

## 许可证

MIT License

## 联系方式

- 问题反馈: [GitHub Issues](https://github.com/...)
- 技术支持: support@example.com

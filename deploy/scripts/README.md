# 3D 模型预览图渲染脚本

使用 Blender 命令行渲染 3D 模型的预览图。

## 安装依赖

### Debian/Ubuntu

```bash
# 安装 Blender
sudo apt update
sudo apt install -y blender

# 可选：安装 ImageMagick（用于格式转换）
sudo apt install -y imagemagick
```

### 验证安装

```bash
blender --version
```

## 使用方法

### 方法 1：使用 Shell 脚本（推荐）

```bash
# 给脚本添加执行权限
chmod +x render_model.sh

# 渲染预览图
./render_model.sh model.fbx preview.png

# 指定分辨率
./render_model.sh model.fbx preview.png 1920 1080

# 输出为 WebP 格式
./render_model.sh model.fbx preview.webp 1280 720
```

### 方法 2：直接使用 Python 脚本

```bash
# 渲染 FBX 模型
blender -b -P render_fbx.py -- model.fbx output.png

# 指定分辨率
blender -b -P render_fbx.py -- model.fbx output.png 1920 1080
```

## 支持的格式

### 输入格式

- `.fbx` - Autodesk FBX
- `.obj` - Wavefront OBJ
- `.gltf` - GL Transmission Format
- `.glb` - GL Transmission Format (Binary)

### 输出格式

- `.png` - PNG 图片（推荐，支持透明背景）
- `.jpg` - JPEG 图片
- `.webp` - WebP 图片（需要 ImageMagick）

## 渲染参数

- **默认分辨率**: 1280x720 (720p)
- **渲染引擎**: Blender EEVEE（快速）
- **背景**: 透明背景
- **采样**: 64 samples
- **相机**: 自动根据模型大小调整位置
- **灯光**: 太阳光 + 环境光

## 示例

### 基础使用

```bash
# 渲染 FBX 模型
./render_model.sh /path/to/model.fbx /path/to/preview.png

# 渲染 OBJ 模型
./render_model.sh /path/to/model.obj /path/to/preview.png

# 渲染 GLTF 模型
./render_model.sh /path/to/model.gltf /path/to/preview.png
```

### 高分辨率渲染

```bash
# 1080p
./render_model.sh model.fbx preview.png 1920 1080

# 4K
./render_model.sh model.fbx preview.png 3840 2160
```

### 批量渲染

```bash
# 渲染目录下所有 FBX 文件
for file in *.fbx; do
    ./render_model.sh "$file" "${file%.fbx}.png"
done
```

## 集成到 Go 项目

在 Go 代码中调用渲染脚本：

```go
package main

import (
    "fmt"
    "os/exec"
)

func RenderFBXPreview(inputPath, outputPath string, width, height int) error {
    cmd := exec.Command(
        "/path/to/render_model.sh",
        inputPath,
        outputPath,
        fmt.Sprintf("%d", width),
        fmt.Sprintf("%d", height),
    )

    output, err := cmd.CombinedOutput()
    if err != nil {
        return fmt.Errorf("渲染失败: %w\n%s", err, output)
    }

    return nil
}
```

## 性能优化

### 1. 使用 EEVEE 引擎（默认）

- 速度快，适合预览图
- 渲染时间：1-5 秒

### 2. 使用 Cycles 引擎（高质量）

修改 `render_fbx.py` 中的渲染引擎：

```python
scene.render.engine = 'CYCLES'
scene.cycles.samples = 128
```

### 3. 降低采样数

```python
scene.eevee.taa_render_samples = 32  # 默认 64
```

### 4. 并行渲染

```bash
# 使用 GNU Parallel 批量渲染
parallel -j 4 ./render_model.sh {} {.}.png ::: *.fbx
```

## 故障排查

### 问题 1: Blender 未找到

```bash
# 检查 Blender 是否安装
which blender

# 如果未安装
sudo apt install -y blender
```

### 问题 2: 渲染失败

```bash
# 查看详细错误信息
blender -b -P render_fbx.py -- model.fbx output.png 2>&1 | less
```

### 问题 3: 内存不足

```bash
# 降低分辨率
./render_model.sh model.fbx preview.png 640 360

# 或者降低采样数（修改 render_fbx.py）
```

### 问题 4: 模型显示不完整

检查模型文件是否完整，或者调整相机距离（修改 `render_fbx.py` 中的 `distance = size * 2.5`）。

## 高级配置

### 自定义相机角度

编辑 `render_fbx.py`，修改 `setup_camera` 函数：

```python
# 俯视 45 度
angle = math.radians(45)

# 改为俯视 30 度
angle = math.radians(30)
```

### 自定义灯光

编辑 `render_fbx.py`，修改 `setup_lighting` 函数：

```python
# 增加灯光强度
sun.data.energy = 3.0  # 默认 2.0

# 增加环境光
bg_node.inputs[1].default_value = 1.0  # 默认 0.5
```

### 添加背景颜色

编辑 `render_fbx.py`，修改 `setup_render` 函数：

```python
# 关闭透明背景
scene.render.film_transparent = False

# 设置背景颜色
scene.world.use_nodes = True
bg_node = scene.world.node_tree.nodes['Background']
bg_node.inputs[0].default_value = (0.5, 0.5, 0.5, 1.0)  # 灰色背景
```

## 许可证

MIT License

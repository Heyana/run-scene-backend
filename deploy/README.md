# 3D 编辑器后端服务 - 部署说明

## 目录结构

```
deploy/
├── app-linux           # Linux 可执行文件
├── start.sh            # 启动脚本
├── config.example.yaml # 配置文件示例
├── configs/            # 贴图映射配置
├── static/             # 静态资源目录
├── data/               # 数据库目录
├── temp/               # 临时文件目录
└── logs/               # 日志目录
```

## 部署步骤

### 1. 上传文件到 NAS

将整个 `deploy` 目录上传到飞牛 NAS，例如：

```
/volume1/docker/3d-editor-backend/
```

### 2. 配置服务

复制配置文件：

```bash
cp config.example.yaml config.yaml
```

编辑 `config.yaml`，修改以下配置：

#### 基础配置

```yaml
server_port: 23359 # 服务端口
environment: prod # 生产环境
```

#### 数据库配置

```yaml
database:
  type: sqlite
  path: data/app.db # 数据库文件路径
```

#### 贴图服务配置

```yaml
texture:
  # API 配置
  api_base_url: https://api.polyhaven.com/
  api_timeout: 30

  # 代理配置（如果需要）
  proxy_enabled: true
  proxy_url: http://127.0.0.1:7890

  # 存储配置
  local_storage_enabled: false # 禁用本地存储
  storage_dir: static/textures

  # NAS SMB 配置
  nas_enabled: true
  nas_path: /volume1/project/editor_v2/static/textures # NAS 路径

  # 下载配置
  download_concurrency: 10
  download_thumbnail: true

  # 同步配置
  sync_interval: 6h
```

### 3. 启动服务

#### 方式一：直接启动

```bash
chmod +x start.sh
./start.sh
```

#### 方式二：后台运行

```bash
chmod +x app-linux
nohup ./app-linux > logs/app.log 2>&1 &
```

#### 方式三：使用 systemd（推荐）

创建服务文件 `/etc/systemd/system/3d-editor-backend.service`：

```ini
[Unit]
Description=3D Editor Backend Service
After=network.target

[Service]
Type=simple
User=your-user
WorkingDirectory=/volume1/docker/3d-editor-backend
ExecStart=/volume1/docker/3d-editor-backend/app-linux
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

启动服务：

```bash
sudo systemctl daemon-reload
sudo systemctl enable 3d-editor-backend
sudo systemctl start 3d-editor-backend
sudo systemctl status 3d-editor-backend
```

### 4. 查看日志

```bash
# 实时查看日志
tail -f logs/app.log

# 或使用 systemd
sudo journalctl -u 3d-editor-backend -f
```

### 5. 停止服务

```bash
# 如果使用 systemd
sudo systemctl stop 3d-editor-backend

# 如果使用 nohup
ps aux | grep app-linux
kill <PID>
```

## 功能说明

### 材质库同步

服务启动后会自动：

1. 执行 PolyHaven 增量同步
2. 执行 AmbientCG 增量同步
3. 每 6 小时自动同步一次

### 按需下载

- 同步时只下载元数据和预览图
- 用户点击使用时才下载实际贴图文件
- 支持 PolyHaven 和 AmbientCG 两个数据源

### API 端点

- `GET /api/textures` - 获取材质列表
- `POST /api/textures/download/:assetId` - 触发下载
- `GET /api/textures/download-status/:assetId` - 查询下载状态
- `POST /api/textures/sync` - 触发同步
- `GET /api/docs` - API 文档

## 注意事项

1. **端口配置**：确保端口 23359 未被占用
2. **NAS 路径**：确保 NAS 路径存在且有写入权限
3. **代理配置**：如果访问国外 API 较慢，建议启用代理
4. **磁盘空间**：材质文件较大，确保有足够的存储空间
5. **数据库备份**：定期备份 `data/app.db` 文件

## 故障排查

### 服务无法启动

- 检查端口是否被占用：`netstat -tlnp | grep 23359`
- 检查配置文件是否正确：`cat config.yaml`
- 查看日志：`tail -f logs/app.log`

### 材质下载失败

- 检查网络连接
- 检查代理配置
- 检查 NAS 路径权限

### 数据库错误

- 检查 `data` 目录权限
- 删除 `data/app.db` 重新初始化（会丢失数据）

## 更新服务

1. 停止服务
2. 备份数据库：`cp data/app.db data/app.db.backup`
3. 替换 `app-linux` 文件
4. 启动服务

## 技术支持

如有问题，请查看日志文件或联系技术支持。

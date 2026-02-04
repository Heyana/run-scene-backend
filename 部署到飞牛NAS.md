# 部署到飞牛 NAS 指南

## 📦 打包文件

已生成：`3d-editor-backend-linux.tar.gz` (约 18 MB)

## 🚀 快速部署

### 1. 上传文件到 NAS

通过 SSH 或文件管理器上传 `3d-editor-backend-linux.tar.gz` 到 NAS，例如：

```
/volume1/docker/3d-editor-backend/
```

### 2. 解压文件

SSH 连接到 NAS：

```bash
ssh admin@your-nas-ip
```

解压文件：

```bash
cd /volume1/docker/3d-editor-backend
tar -xzf 3d-editor-backend-linux.tar.gz
```

### 3. 运行安装脚本

```bash
chmod +x install.sh
./install.sh
```

安装脚本会自动：

- 创建配置文件
- 设置执行权限
- 创建必要目录
- 询问是否创建 systemd 服务

### 4. 配置服务

编辑 `config.yaml`：

```bash
nano config.yaml
```

**重要配置项：**

```yaml
# 服务端口
server_port: 23359

# NAS 存储路径（根据实际情况修改）
texture:
  nas_enabled: true
  nas_path: /volume1/project/editor_v2/static/textures

  # 如果需要代理访问国外 API
  proxy_enabled: true
  proxy_url: http://127.0.0.1:7890
```

### 5. 启动服务

#### 方式一：直接启动（测试用）

```bash
./start.sh
```

#### 方式二：后台运行

```bash
nohup ./app-linux > logs/app.log 2>&1 &
```

#### 方式三：systemd 服务（推荐）

```bash
sudo systemctl start 3d-editor-backend
sudo systemctl status 3d-editor-backend
```

### 6. 验证服务

访问 API 文档：

```
http://your-nas-ip:23359/api/docs
```

查看日志：

```bash
tail -f logs/app.log
# 或
sudo journalctl -u 3d-editor-backend -f
```

## 📋 目录结构

```
/volume1/docker/3d-editor-backend/
├── app-linux              # 可执行文件
├── start.sh               # 启动脚本
├── install.sh             # 安装脚本
├── config.yaml            # 配置文件
├── config.example.yaml    # 配置示例
├── configs/               # 贴图映射配置
│   └── texture_mapping.yaml
├── static/                # 静态资源
│   ├── cdn/
│   └── textures/
├── data/                  # 数据库
│   └── app.db
├── temp/                  # 临时文件
└── logs/                  # 日志文件
    └── app.log
```

## 🔧 常用命令

### 服务管理

```bash
# 启动
sudo systemctl start 3d-editor-backend

# 停止
sudo systemctl stop 3d-editor-backend

# 重启
sudo systemctl restart 3d-editor-backend

# 状态
sudo systemctl status 3d-editor-backend

# 开机自启
sudo systemctl enable 3d-editor-backend
```

### 日志查看

```bash
# 实时日志
tail -f logs/app.log

# systemd 日志
sudo journalctl -u 3d-editor-backend -f

# 最近 100 行
sudo journalctl -u 3d-editor-backend -n 100
```

### 进程管理

```bash
# 查看进程
ps aux | grep app-linux

# 杀死进程
kill <PID>

# 强制杀死
kill -9 <PID>
```

## 🎯 功能特性

### 材质库同步

- ✅ PolyHaven：733 个材质
- ✅ AmbientCG：1957 个材质
- ✅ 自动增量同步（每 6 小时）
- ✅ 按需下载（节省存储空间）

### API 端点

- `GET /api/textures` - 材质列表
- `POST /api/textures/download/:assetId` - 触发下载
- `GET /api/textures/download-status/:assetId` - 下载状态
- `POST /api/textures/sync` - 触发同步
- `GET /api/docs` - API 文档

## ⚠️ 注意事项

1. **端口占用**：确保端口 23359 未被占用
2. **NAS 路径**：确保 NAS 路径存在且有写入权限
3. **磁盘空间**：材质文件较大，建议预留 50GB+ 空间
4. **网络代理**：访问国外 API 建议配置代理
5. **数据备份**：定期备份 `data/app.db` 数据库文件

## 🐛 故障排查

### 服务无法启动

```bash
# 检查端口
netstat -tlnp | grep 23359

# 检查日志
tail -f logs/app.log

# 检查配置
cat config.yaml
```

### 材质下载失败

- 检查网络连接
- 检查代理配置
- 检查 NAS 路径权限：`ls -la /volume1/project/editor_v2/static/textures`

### 数据库错误

```bash
# 检查权限
ls -la data/

# 重新初始化（会丢失数据）
rm data/app.db
./app-linux
```

## 🔄 更新服务

1. 停止服务：

```bash
sudo systemctl stop 3d-editor-backend
```

2. 备份数据库：

```bash
cp data/app.db data/app.db.backup.$(date +%Y%m%d)
```

3. 替换可执行文件：

```bash
# 上传新的 tar.gz 并解压
tar -xzf 3d-editor-backend-linux-new.tar.gz
```

4. 启动服务：

```bash
sudo systemctl start 3d-editor-backend
```

## 📞 技术支持

如有问题，请查看：

1. 日志文件：`logs/app.log`
2. API 文档：`http://your-nas-ip:23359/api/docs`
3. README.md 文件

---

**部署完成后，记得在前端配置中修改 API 地址为 NAS 的 IP！**

cd E:\hxy\project-2026\dify-full

# 恢复备份

Copy-Item docker-compose.yaml.backup docker-compose.yaml -Force

# 启动服务

docker-compose up -d

# 等待

Start-Sleep -Seconds 30

# 检查端口

netstat -ano | findstr ":3001"

# 测试容器 IP 访问

curl http://172.18.0.5:3000

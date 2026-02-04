# Dify 局域网访问解决方案

## 问题描述

Dify 在 Windows Docker Desktop 上部署后，只能通过 `localhost` 访问，无法从局域网其他设备访问。

**根本原因**：Docker Desktop for Windows 运行在 WSL2/Hyper-V 虚拟机中，容器端口即使绑定到 `0.0.0.0`，实际上也只监听在 `127.0.0.1`，这是 Docker Desktop 的架构限制。

## 解决方案对比

| 方案               | 难度 | 推荐度     | 说明                   |
| ------------------ | ---- | ---------- | ---------------------- |
| 1. 迁移到 Linux    | ⭐⭐ | ⭐⭐⭐⭐⭐ | 最佳方案，彻底解决问题 |
| 2. PowerShell 代理 | ⭐   | ⭐⭐⭐     | 简单但需要一直运行     |
| 3. Nginx 反向代理  | ⭐⭐ | ⭐⭐⭐⭐   | 稳定可靠，适合长期使用 |
| 4. SSH 隧道        | ⭐   | ⭐⭐       | 临时方案，不适合多用户 |

---

## 方案 1：迁移到 Linux（强烈推荐）✅

### 优势

- ✅ 原生 Docker 支持，无网络隔离问题
- ✅ 性能更好，资源占用更少
- ✅ 可以 24/7 运行
- ✅ 局域网访问完全正常

### 部署步骤

#### 1. 在飞牛 NAS 上部署

```bash
# SSH 登录到 NAS
ssh admin@192.168.3.39

# 下载部署脚本
curl -fsSL https://raw.githubusercontent.com/langgenius/dify/main/docker/docker-compose.yaml -o /volume1/docker/dify/docker-compose.yaml

# 创建 .env 文件
cd /volume1/docker/dify
cat > .env << 'EOF'
# Dify 配置
CONSOLE_API_URL=http://192.168.3.39:5001
CONSOLE_WEB_URL=http://192.168.3.39:3001
APP_API_URL=http://192.168.3.39:5001
APP_WEB_URL=http://192.168.3.39:3001

# 其他配置保持默认...
EOF

# 启动服务
docker-compose up -d

# 查看日志
docker-compose logs -f
```

#### 2. 访问地址

- 局域网: `http://192.168.3.39:3001`
- 本机: `http://localhost:3001`

---

## 方案 2：PowerShell HTTP 代理（Windows 临时方案）

### 使用方法

1. 以管理员身份运行 PowerShell
2. 执行代理脚本：

```powershell
cd E:\hxy\my-project\run-scene-backend
powershell -ExecutionPolicy Bypass -File dify-proxy.ps1
```

3. 访问地址：`http://192.168.3.39:8080`

### 优缺点

- ✅ 简单快速，无需安装额外软件
- ✅ 可以随时启动/停止
- ❌ 需要一直运行 PowerShell 窗口
- ❌ 性能略低于 Nginx

---

## 方案 3：Nginx 反向代理（Windows 长期方案）

### 安装步骤

1. 以管理员身份运行 PowerShell
2. 执行安装脚本：

```powershell
cd E:\hxy\my-project\run-scene-backend
powershell -ExecutionPolicy Bypass -File setup-nginx-proxy.ps1
```

3. 访问地址：`http://192.168.3.39:8080`

### 管理命令

```powershell
# 停止 Nginx
taskkill /f /im nginx.exe

# 启动 Nginx
cd C:\nginx
start nginx.exe

# 重新加载配置
nginx.exe -s reload

# 查看日志
Get-Content C:\nginx\logs\error.log -Tail 50
```

### 优缺点

- ✅ 稳定可靠，适合长期使用
- ✅ 性能好，支持负载均衡
- ✅ 可以配置 HTTPS
- ❌ 需要安装额外软件
- ❌ 配置相对复杂

---

## 方案 4：SSH 隧道（临时访问）

### 使用方法

在需要访问的设备上执行：

```bash
# 建立 SSH 隧道
ssh -L 3001:localhost:3001 user@192.168.3.39

# 然后在浏览器访问
http://localhost:3001
```

### 优缺点

- ✅ 无需修改服务器配置
- ✅ 安全性高（通过 SSH 加密）
- ❌ 每个设备都需要单独配置
- ❌ 需要 SSH 访问权限

---

## 当前状态

### Windows 部署（E:\hxy\project-2026\dify-full）

- ✅ 服务已启动
- ✅ 本机访问正常：`http://localhost:3001`
- ❌ 局域网访问不通：`http://192.168.3.39:3001`
- ⚠️ 部分容器重启（nginx、sandbox、ssrf_proxy）

### 端口映射

- Web: `0.0.0.0:3001 -> 3000`
- API: `0.0.0.0:5001 -> 5001`
- Plugin: `0.0.0.0:5003 -> 5003`

### 已配置的端口转发（无效）

```
192.168.3.39:3001 -> 127.0.0.1:3001
192.168.3.39:5001 -> 127.0.0.1:5001
```

---

## 推荐行动方案

### 短期（今天就能用）

1. 使用 **PowerShell 代理**（方案 2）
2. 运行 `dify-proxy.ps1` 脚本
3. 通过 `http://192.168.3.39:8080` 访问

### 长期（最佳实践）

1. 在**飞牛 NAS**上重新部署（方案 1）
2. 使用 `deploy-dify-to-nas.sh` 脚本
3. 享受原生 Docker 的完美体验

---

## 下一步：配置 DeepSeek 模型

部署完成后，配置 DeepSeek：

1. 访问 Dify 界面
2. 进入 **设置 → 模型供应商**
3. 选择 **OpenAI API-Compatible**
4. 填入配置：
   - Base URL: `https://api.deepseek.com/v1`
   - API Key: `你的 DeepSeek API Key`
   - 模型名称: `deepseek-chat`

**注意**：不需要安装插件，直接使用 OpenAI 兼容接口即可。

---

## 常见问题

### Q: 为什么端口转发不工作？

A: Docker Desktop for Windows 的容器运行在虚拟机中，端口只绑定到虚拟机的 127.0.0.1，无法通过 netsh 转发。

### Q: 可以修改 Docker 配置解决吗？

A: 不行，这是 Docker Desktop 的架构限制，无法通过配置解决。

### Q: 使用 Docker Toolbox 可以吗？

A: Docker Toolbox 已经停止维护，不推荐使用。

### Q: 有没有不需要额外软件的方案？

A: 没有。要么使用代理/反向代理，要么迁移到 Linux。

---

## 参考资料

- [Docker Desktop for Windows 网络限制](https://docs.docker.com/desktop/networking/)
- [Dify 官方文档](https://docs.dify.ai/)
- [DeepSeek API 文档](https://platform.deepseek.com/api-docs/)

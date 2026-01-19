# 局域网访问配置

## 配置说明

前端已配置为支持局域网访问。

### Vite 配置

```typescript
server: {
  host: '0.0.0.0',  // 监听所有网络接口
  port: 3000,
  proxy: {
    "/api": {
      target: "http://192.168.3.39:23359",  // 后端 API 地址
      changeOrigin: true,
    },
    "/textures": {
      target: "http://192.168.3.39:23359",  // 材质文件代理
      changeOrigin: true,
    },
  },
}
```

## 访问方式

### 本机访问

```
http://localhost:3000
http://127.0.0.1:3000
```

### 局域网访问

```
http://192.168.3.39:3000
```

其他局域网设备可以通过服务器的 IP 地址访问前端。

## 启动服务

```bash
cd frontend
yarn dev
```

启动后会显示：

```
VITE v7.x.x  ready in xxx ms

➜  Local:   http://localhost:3000/
➜  Network: http://192.168.3.39:3000/
➜  press h + enter to show help
```

## 防火墙配置

如果局域网无法访问，需要检查防火墙设置：

### Windows 防火墙

1. 打开 Windows Defender 防火墙
2. 点击"高级设置"
3. 入站规则 -> 新建规则
4. 端口 -> TCP -> 特定本地端口 -> 3000
5. 允许连接
6. 应用到所有配置文件
7. 命名为 "Vite Dev Server"

### 或使用命令行（管理员权限）

```powershell
# 允许端口 3000
netsh advfirewall firewall add rule name="Vite Dev Server" dir=in action=allow protocol=TCP localport=3000
```

## 后端配置

确保后端也配置了 CORS 允许局域网访问：

```go
// api/routes.go
router.Use(CorsMiddleware())
```

后端地址：`http://192.168.3.39:23359`

## 测试连接

### 测试前端

```bash
# 从其他设备访问
curl http://192.168.3.39:3000
```

### 测试后端 API

```bash
curl http://192.168.3.39:23359/api/ping
```

### 测试材质文件

```bash
curl http://192.168.3.39:23359/textures/test.jpg
```

## 常见问题

### 1. 局域网无法访问

**检查项**：

- ✅ Vite 配置 `host: '0.0.0.0'`
- ✅ 防火墙允许端口 3000
- ✅ 确认 IP 地址正确
- ✅ 确认在同一局域网

### 2. API 请求失败

**检查项**：

- ✅ 后端服务正在运行
- ✅ 后端地址配置正确
- ✅ CORS 配置正确
- ✅ 防火墙允许后端端口

### 3. 图片无法加载

**检查项**：

- ✅ `/textures` 代理配置正确
- ✅ NAS 路径可访问
- ✅ 文件权限正确

## 生产环境部署

生产环境建议使用 Nginx 反向代理：

```nginx
server {
    listen 80;
    server_name 192.168.3.39;

    # 前端静态文件
    location / {
        root /path/to/frontend/dist;
        try_files $uri $uri/ /index.html;
    }

    # API 代理
    location /api {
        proxy_pass http://localhost:23359;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }

    # 材质文件代理
    location /textures {
        proxy_pass http://localhost:23359;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

## 安全建议

1. **生产环境**：不要使用 `0.0.0.0`，使用具体 IP
2. **防火墙**：只允许信任的 IP 访问
3. **HTTPS**：生产环境使用 HTTPS
4. **认证**：添加用户认证机制

# Nginx 配置

## 前提条件

- Nginx 已安装 (`apt install nginx` 或 `yum install nginx`)
- 服务器开放 80/443 端口

---

## 配置文件

### 1. 创建上游配置

```bash
sudo vim /etc/nginx/conf.d/online_show.conf
```

```nginx
# 上游服务器组
upstream online_show_backend {
    # 蓝绿切换：注释/取消注释对应行
    server 127.0.0.1:8001;  # 蓝 (当前活跃)
    # server 127.0.0.1:8002;  # 绿 (备用)

    # 健康检查（可选，需要 nginx-plus 或第三方模块）
    # health_check interval=5s fails=3 passes=2;
}

server {
    listen 80;
    server_name your-domain.com;  # 替换为你的域名或IP

    # 日志
    access_log /var/log/nginx/online_show_access.log;
    error_log /var/log/nginx/online_show_error.log;

    # 代理设置
    location / {
        proxy_pass http://online_show_backend;
        proxy_http_version 1.1;

        # 请求头转发
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # 超时设置
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
    }

    # 健康检查端点（直接暴露，不经过上游）
    location /nginx-health {
        return 200 'OK';
        add_header Content-Type text/plain;
    }
}
```

### 2. 测试并重载

```bash
# 测试配置语法
sudo nginx -t

# 重载配置（不中断连接）
sudo nginx -s reload
```

---

## 切换方式

### 方式 A：修改配置文件

```bash
# 编辑配置，切换注释
sudo vim /etc/nginx/conf.d/online_show.conf

# 修改前
server 127.0.0.1:8001;  # 蓝
# server 127.0.0.1:8002;  # 绿

# 修改后
# server 127.0.0.1:8001;  # 蓝
server 127.0.0.1:8002;  # 绿

# 热重载
sudo nginx -s reload
```

### 方式 B：使用 sed 一键切换

```bash
# 切换到 8002
sudo sed -i 's/server 127.0.0.1:8001;/# server 127.0.0.1:8001;/' /etc/nginx/conf.d/online_show.conf
sudo sed -i 's/# server 127.0.0.1:8002;/server 127.0.0.1:8002;/' /etc/nginx/conf.d/online_show.conf
sudo nginx -s reload

# 切换回 8001
sudo sed -i 's/# server 127.0.0.1:8001;/server 127.0.0.1:8001;/' /etc/nginx/conf.d/online_show.conf
sudo sed -i 's/server 127.0.0.1:8002;/# server 127.0.0.1:8002;/' /etc/nginx/conf.d/online_show.conf
sudo nginx -s reload
```

---

## HTTPS 配置（可选）

```nginx
server {
    listen 443 ssl http2;
    server_name your-domain.com;

    ssl_certificate /etc/nginx/ssl/cert.pem;
    ssl_certificate_key /etc/nginx/ssl/key.pem;

    # 其余配置同上...
    location / {
        proxy_pass http://online_show_backend;
        # ...
    }
}

# HTTP 重定向到 HTTPS
server {
    listen 80;
    server_name your-domain.com;
    return 301 https://$server_name$request_uri;
}
```

---

## 验证

```bash
# 查看 Nginx 状态
sudo systemctl status nginx

# 测试代理
curl http://localhost/health

# 查看连接状态
sudo netstat -tlnp | grep nginx
```

## 相关文档

- [概述](./01_概述.md)
- [部署脚本](./03_部署脚本.md)

# Dify 局域网访问 - Nginx 反向代理设置脚本

Write-Host "开始配置 Nginx 反向代理..." -ForegroundColor Green

# 下载 Nginx for Windows
$nginxUrl = "http://nginx.org/download/nginx-1.24.0.zip"
$output = "$env:TEMP\nginx.zip"

Write-Host "正在下载 Nginx..." -ForegroundColor Yellow
Invoke-WebRequest -Uri $nginxUrl -OutFile $output

# 解压到 C:\nginx
Write-Host "正在解压 Nginx..." -ForegroundColor Yellow
Expand-Archive -Path $output -DestinationPath "C:\" -Force
if (Test-Path "C:\nginx") {
    Remove-Item "C:\nginx" -Recurse -Force
}
Rename-Item -Path "C:\nginx-1.24.0" -NewName "nginx" -Force

# 创建配置文件
Write-Host "正在创建配置文件..." -ForegroundColor Yellow
$config = @'
worker_processes  1;
events {
    worker_connections  1024;
}
http {
    server {
        listen       0.0.0.0:8080;
        server_name  _;
        
        location / {
            proxy_pass http://127.0.0.1:3001;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
        }
        
        location /api {
            proxy_pass http://127.0.0.1:5001;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
        }
    }
}
'@
Set-Content -Path "C:\nginx\conf\nginx.conf" -Value $config -Encoding UTF8

# 添加防火墙规则
Write-Host "正在添加防火墙规则..." -ForegroundColor Yellow
netsh advfirewall firewall delete rule name="Nginx Port 8080" | Out-Null
netsh advfirewall firewall add rule name="Nginx Port 8080" dir=in action=allow protocol=TCP localport=8080

# 启动 Nginx
Write-Host "正在启动 Nginx..." -ForegroundColor Yellow
Start-Process -FilePath "C:\nginx\nginx.exe" -WorkingDirectory "C:\nginx"

Write-Host ""
Write-Host "✅ Nginx 已启动！" -ForegroundColor Green
Write-Host "局域网访问地址: http://192.168.3.39:8080" -ForegroundColor Cyan
Write-Host "本机访问地址: http://localhost:8080" -ForegroundColor Cyan
Write-Host ""
Write-Host "停止 Nginx: taskkill /f /im nginx.exe" -ForegroundColor Yellow

#!/bin/bash

# 3D 编辑器后端服务 - 快速安装脚本

echo "================================"
echo "3D 编辑器后端服务 - 安装向导"
echo "================================"
echo ""

# 检查是否为 root
if [ "$EUID" -ne 0 ]; then 
    echo "提示：某些操作可能需要 root 权限"
fi

# 1. 创建配置文件
if [ ! -f "config.yaml" ]; then
    echo "1. 创建配置文件..."
    cp config.example.yaml config.yaml
    echo "   ✓ 配置文件已创建"
else
    echo "1. 配置文件已存在，跳过"
fi

# 2. 设置执行权限
echo "2. 设置执行权限..."
chmod +x app-linux
chmod +x start.sh
echo "   ✓ 权限设置完成"

# 3. 创建必要目录
echo "3. 创建必要目录..."
mkdir -p data temp logs
echo "   ✓ 目录创建完成"

# 4. 询问是否创建 systemd 服务
echo ""
read -p "是否创建 systemd 服务？(y/n) " -n 1 -r
echo ""
if [[ $REPLY =~ ^[Yy]$ ]]; then
    CURRENT_DIR=$(pwd)
    CURRENT_USER=$(whoami)
    
    SERVICE_FILE="/etc/systemd/system/3d-editor-backend.service"
    
    echo "[Unit]
Description=3D Editor Backend Service
After=network.target

[Service]
Type=simple
User=$CURRENT_USER
WorkingDirectory=$CURRENT_DIR
ExecStart=$CURRENT_DIR/app-linux
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target" | sudo tee $SERVICE_FILE > /dev/null

    sudo systemctl daemon-reload
    sudo systemctl enable 3d-editor-backend
    
    echo "   ✓ systemd 服务已创建"
    echo ""
    echo "使用以下命令管理服务："
    echo "  启动: sudo systemctl start 3d-editor-backend"
    echo "  停止: sudo systemctl stop 3d-editor-backend"
    echo "  状态: sudo systemctl status 3d-editor-backend"
    echo "  日志: sudo journalctl -u 3d-editor-backend -f"
fi

echo ""
echo "================================"
echo "安装完成！"
echo "================================"
echo ""
echo "下一步："
echo "1. 编辑 config.yaml 配置文件"
echo "2. 启动服务："
echo "   - 直接启动: ./start.sh"
echo "   - 后台运行: nohup ./app-linux > logs/app.log 2>&1 &"
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo "   - systemd: sudo systemctl start 3d-editor-backend"
fi
echo ""
echo "访问 API 文档: http://your-ip:23359/api/docs"
echo ""

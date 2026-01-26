#!/bin/bash

# ========================================
# 安装 systemd 服务脚本
# ========================================

echo "========================================"
echo "安装 3D 编辑器后端 systemd 服务"
echo "========================================"
echo ""

# 检查是否以 root 运行
if [ "$EUID" -ne 0 ]; then 
    echo "❌ 请使用 sudo 运行此脚本"
    echo "   sudo ./安装systemd服务.sh"
    exit 1
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# 1. 停止现有进程
echo "1. 停止现有进程..."
pkill -f mihomo
pkill -f app-linux
sleep 2
echo "   ✓ 已停止现有进程"
echo ""

# 2. 复制 mihomo 服务文件
echo "2. 安装 Mihomo 服务..."
cp "$SCRIPT_DIR/mihomo.service" /etc/systemd/system/
chmod 644 /etc/systemd/system/mihomo.service
echo "   ✓ 服务文件已复制到 /etc/systemd/system/mihomo.service"
echo ""

# 3. 复制后端服务文件
echo "3. 安装后端服务..."
cp "$SCRIPT_DIR/3d-editor-backend.service" /etc/systemd/system/
chmod 644 /etc/systemd/system/3d-editor-backend.service
echo "   ✓ 服务文件已复制到 /etc/systemd/system/3d-editor-backend.service"
echo ""

# 4. 重新加载 systemd
echo "4. 重新加载 systemd..."
systemctl daemon-reload
echo "   ✓ systemd 已重新加载"
echo ""

# 5. 启用服务（开机自启）
echo "5. 启用服务（开机自启）..."
systemctl enable mihomo.service
systemctl enable 3d-editor-backend.service
echo "   ✓ 服务已设置为开机自启"
echo ""

# 6. 启动服务
echo "6. 启动服务..."
systemctl start mihomo.service
sleep 3
systemctl start 3d-editor-backend.service
sleep 2
echo "   ✓ 服务已启动"
echo ""

# 7. 检查服务状态
echo "========================================"
echo "服务状态"
echo "========================================"
echo ""

echo "Mihomo 代理服务:"
systemctl status mihomo.service --no-pager -l | head -n 10
echo ""

echo "后端服务:"
systemctl status 3d-editor-backend.service --no-pager -l | head -n 10
echo ""

# 8. 显示管理命令
echo "========================================"
echo "服务管理命令"
echo "========================================"
echo ""
echo "查看状态:"
echo "  sudo systemctl status mihomo"
echo "  sudo systemctl status 3d-editor-backend"
echo ""
echo "启动服务:"
echo "  sudo systemctl start mihomo"
echo "  sudo systemctl start 3d-editor-backend"
echo ""
echo "停止服务:"
echo "  sudo systemctl stop mihomo"
echo "  sudo systemctl stop 3d-editor-backend"
echo ""
echo "重启服务:"
echo "  sudo systemctl restart mihomo"
echo "  sudo systemctl restart 3d-editor-backend"
echo ""
echo "查看日志:"
echo "  sudo journalctl -u mihomo -f"
echo "  sudo journalctl -u 3d-editor-backend -f"
echo "  tail -f /vol1/1003/app/mihomo.log"
echo "  tail -f /vol1/1003/project/editor_v2/deploy/logs/app.log"
echo ""
echo "禁用开机自启:"
echo "  sudo systemctl disable mihomo"
echo "  sudo systemctl disable 3d-editor-backend"
echo ""

echo "✓ 安装完成！"
echo ""
echo "现在你可以关闭 SSH 窗口，服务会继续在后台运行。"

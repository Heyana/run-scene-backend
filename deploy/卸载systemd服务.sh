#!/bin/bash

# ========================================
# 卸载 systemd 服务脚本
# ========================================

echo "========================================"
echo "卸载 3D 编辑器后端 systemd 服务"
echo "========================================"
echo ""

# 检查是否以 root 运行
if [ "$EUID" -ne 0 ]; then 
    echo "❌ 请使用 sudo 运行此脚本"
    echo "   sudo ./卸载systemd服务.sh"
    exit 1
fi

# 1. 停止服务
echo "1. 停止服务..."
systemctl stop 3d-editor-backend.service
systemctl stop mihomo.service
echo "   ✓ 服务已停止"
echo ""

# 2. 禁用服务
echo "2. 禁用开机自启..."
systemctl disable 3d-editor-backend.service
systemctl disable mihomo.service
echo "   ✓ 已禁用开机自启"
echo ""

# 3. 删除服务文件
echo "3. 删除服务文件..."
rm -f /etc/systemd/system/mihomo.service
rm -f /etc/systemd/system/3d-editor-backend.service
echo "   ✓ 服务文件已删除"
echo ""

# 4. 重新加载 systemd
echo "4. 重新加载 systemd..."
systemctl daemon-reload
systemctl reset-failed
echo "   ✓ systemd 已重新加载"
echo ""

echo "✓ 卸载完成！"
echo ""
echo "如需手动启动服务，可以使用:"
echo "  cd /vol1/1003/app && nohup ./mihomo -d ./mihomo-config > mihomo.log 2>&1 &"
echo "  cd /vol1/1003/project/editor_v2/deploy && nohup ./app-linux > logs/app.log 2>&1 &"

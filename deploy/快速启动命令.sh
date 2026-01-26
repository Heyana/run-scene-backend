#!/bin/bash

# ========================================
# 快速启动所有服务
# ========================================

echo "========================================"
echo "启动 3D 编辑器后端服务"
echo "========================================"
echo ""

# 1. 启动 Mihomo 代理
echo "1. 启动 Mihomo 代理..."
cd /vol1/1003/app
pkill -f mihomo  # 先停止旧进程
sleep 1
nohup ./mihomo -d ./mihomo-config > mihomo.log 2>&1 &
MIHOMO_PID=$!
echo "   ✓ Mihomo 已启动 (PID: $MIHOMO_PID)"
echo "   - HTTP:  http://127.0.0.1:7890"
echo "   - SOCKS: socks5://127.0.0.1:7891"
echo ""

# 2. 等待代理启动
echo "2. 等待代理启动..."
sleep 3
echo "   ✓ 代理已就绪"
echo ""

# 3. 测试代理
echo "3. 测试代理连接..."
if curl -x http://127.0.0.1:7890 -s --connect-timeout 5 https://www.google.com > /dev/null 2>&1; then
    echo "   ✓ 代理连接正常"
else
    echo "   ✗ 代理连接失败，但继续启动后端"
fi
echo ""

# 4. 启动后端服务
echo "4. 启动后端服务..."
cd /vol1/1003/project/editor_v2/deploy
pkill -f app-linux  # 先停止旧进程
sleep 1
nohup ./app-linux > logs/app.log 2>&1 &
BACKEND_PID=$!
echo "   ✓ 后端已启动 (PID: $BACKEND_PID)"
echo "   - API: http://192.168.3.10:23359"
echo ""

# 5. 等待后端启动
echo "5. 等待后端启动..."
sleep 3
echo ""

# 6. 显示服务状态
echo "========================================"
echo "服务状态"
echo "========================================"
echo ""
echo "Mihomo 代理:"
ps aux | grep "[m]ihomo" | awk '{print "  PID: "$2"  CMD: "$11" "$12" "$13}'
echo ""
echo "后端服务:"
ps aux | grep "[a]pp-linux" | awk '{print "  PID: "$2"  CMD: "$11}'
echo ""

# 7. 显示日志位置
echo "========================================"
echo "日志文件"
echo "========================================"
echo ""
echo "Mihomo:  /vol1/1003/app/mihomo.log"
echo "后端:    /vol1/1003/project/editor_v2/deploy/logs/app.log"
echo ""
echo "查看日志命令:"
echo "  tail -f /vol1/1003/app/mihomo.log"
echo "  tail -f /vol1/1003/project/editor_v2/deploy/logs/app.log"
echo ""

# 8. 显示访问地址
echo "========================================"
echo "访问地址"
echo "========================================"
echo ""
echo "API 文档:  http://192.168.3.10:23359/api/docs"
echo "前端地址:  http://192.168.3.10:23357"
echo ""

echo "✓ 所有服务启动完成！"

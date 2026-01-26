#!/bin/bash

# Mihomo 启动脚本（解决权限问题）
# 使用当前目录作为配置目录，避免 /home 权限问题

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CONFIG_DIR="$SCRIPT_DIR/mihomo-config"
CONFIG_FILE="$CONFIG_DIR/config.yaml"

# 创建配置目录
mkdir -p "$CONFIG_DIR"

# 检查配置文件是否存在
if [ ! -f "$CONFIG_FILE" ]; then
    echo "错误: 配置文件不存在: $CONFIG_FILE"
    echo "请先下载订阅配置或手动创建配置文件"
    echo ""
    echo "下载订阅配置示例:"
    echo "wget -O $CONFIG_FILE 'YOUR_SUBSCRIPTION_URL'"
    exit 1
fi

# 启动 mihomo，指定配置目录
echo "启动 Mihomo 代理..."
echo "配置目录: $CONFIG_DIR"
echo "配置文件: $CONFIG_FILE"
echo ""

cd "$SCRIPT_DIR"
./mihomo -d "$CONFIG_DIR" > mihomo.log 2>&1 &

MIHOMO_PID=$!
echo "Mihomo 已启动，PID: $MIHOMO_PID"
echo "日志文件: $SCRIPT_DIR/mihomo.log"
echo ""
echo "代理地址:"
echo "  HTTP:  http://127.0.0.1:7890"
echo "  SOCKS: socks5://127.0.0.1:7891"
echo ""
echo "测试代理:"
echo "  curl -x http://127.0.0.1:7890 https://www.google.com"

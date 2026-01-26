#!/bin/bash

# Mihomo 配置下载脚本

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CONFIG_DIR="$SCRIPT_DIR/mihomo-config"
CONFIG_FILE="$CONFIG_DIR/config.yaml"

# 创建配置目录
mkdir -p "$CONFIG_DIR"

# 检查是否提供了订阅链接
if [ -z "$1" ]; then
    echo "用法: $0 <订阅链接>"
    echo ""
    echo "示例:"
    echo "  $0 'https://example.com/subscription'"
    exit 1
fi

SUBSCRIPTION_URL="$1"

echo "下载订阅配置..."
echo "订阅链接: $SUBSCRIPTION_URL"
echo "保存位置: $CONFIG_FILE"
echo ""

# 下载配置
wget -O "$CONFIG_FILE" "$SUBSCRIPTION_URL"

if [ $? -eq 0 ]; then
    echo ""
    echo "✓ 配置下载成功！"
    echo ""
    echo "下一步: 启动 Mihomo"
    echo "  ./mihomo-start.sh"
else
    echo ""
    echo "✗ 配置下载失败"
    exit 1
fi

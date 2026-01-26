#!/bin/bash

# Mihomo 停止脚本

echo "停止 Mihomo 代理..."

# 查找并杀死 mihomo 进程
pkill -f mihomo

if [ $? -eq 0 ]; then
    echo "Mihomo 已停止"
else
    echo "未找到运行中的 Mihomo 进程"
fi

#!/bin/bash

# 3D 编辑器后端服务启动脚本

# 设置工作目录
cd "$(dirname "$0")"

# 检查配置文件
if [ ! -f "config.yaml" ]; then
    echo "配置文件不存在，从示例文件复制..."
    cp config.example.yaml config.yaml
    echo "请编辑 config.yaml 配置文件后再启动服务"
    exit 1
fi

# 赋予执行权限
chmod +x app-linux

# 启动服务
echo "启动 3D 编辑器后端服务..."
./app-linux

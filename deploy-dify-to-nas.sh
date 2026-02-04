#!/bin/bash
# Dify 部署到飞牛 NAS 脚本

set -e

echo "========================================="
echo "  Dify 部署到飞牛 NAS"
echo "========================================="
echo ""

# 配置
INSTALL_DIR="/volume1/docker/dify"
DIFY_VERSION="1.11.4"
NAS_IP="192.168.3.39"  # 修改为你的 NAS IP

# 创建安装目录
echo "📁 创建安装目录..."
mkdir -p $INSTALL_DIR
cd $INSTALL_DIR

# 下载 docker-compose.yaml
echo "📥 下载 docker-compose.yaml..."
curl -fsSL https://raw.githubusercontent.com/langgenius/dify/main/docker/docker-compose.yaml -o docker-compose.yaml

# 下载 .env 模板
echo "📥 下载 .env 配置文件..."
curl -fsSL https://raw.githubusercontent.com/langgenius/dify/main/docker/.env.example -o .env

# 修改配置
echo "⚙️  配置环境变量..."
sed -i "s|CONSOLE_API_URL=.*|CONSOLE_API_URL=http://$NAS_IP:5001|g" .env
sed -i "s|CONSOLE_WEB_URL=.*|CONSOLE_WEB_URL=http://$NAS_IP:3001|g" .env
sed -i "s|APP_API_URL=.*|APP_API_URL=http://$NAS_IP:5001|g" .env
sed -i "s|APP_WEB_URL=.*|APP_WEB_URL=http://$NAS_IP:3001|g" .env

# 修改端口映射（避免与其他服务冲突）
echo "⚙️  修改端口映射..."
sed -i 's|"80:80"|"3001:80"|g' docker-compose.yaml
sed -i 's|"443:443"|"3443:443"|g' docker-compose.yaml

# 启动服务
echo "🚀 启动 Dify 服务..."
docker-compose up -d

echo ""
echo "========================================="
echo "  ✅ Dify 部署完成！"
echo "========================================="
echo ""
echo "访问地址:"
echo "  本地: http://localhost:3001"
echo "  局域网: http://$NAS_IP:3001"
echo ""
echo "查看日志: docker-compose logs -f"
echo "停止服务: docker-compose down"
echo "重启服务: docker-compose restart"
echo ""

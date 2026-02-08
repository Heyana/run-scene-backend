#!/bin/bash

# 安装 Chromium 浏览器（用于生成项目预览截图）

echo "正在安装 Chromium 浏览器..."

# 检测操作系统
if [ -f /etc/os-release ]; then
    . /etc/os-release
    OS=$ID
    VERSION_ID=$VERSION_ID
else
    echo "无法检测操作系统"
    exit 1
fi

# 根据不同的操作系统安装 Chromium
case $OS in
    ubuntu|debian)
        echo "检测到 Ubuntu/Debian 系统 (版本: $VERSION_ID)"
        
        # Ubuntu 20.04+ 使用 snap 安装
        if command -v snap &> /dev/null; then
            echo "使用 snap 安装 Chromium..."
            sudo snap install chromium
        else
            # 尝试使用 apt 安装
            echo "使用 apt 安装 Chromium..."
            sudo apt-get update
            sudo apt-get install -y chromium || sudo apt-get install -y chromium-browser
        fi
        ;;
    centos|rhel|fedora)
        echo "检测到 CentOS/RHEL/Fedora 系统"
        sudo yum install -y chromium
        ;;
    arch)
        echo "检测到 Arch Linux 系统"
        sudo pacman -S --noconfirm chromium
        ;;
    *)
        echo "不支持的操作系统: $OS"
        echo "请手动安装 Chromium 浏览器"
        exit 1
        ;;
esac

# 验证安装
echo ""
echo "验证 Chromium 安装..."
if command -v chromium &> /dev/null; then
    echo "✓ Chromium 安装成功: $(chromium --version)"
    CHROMIUM_PATH=$(which chromium)
elif command -v chromium-browser &> /dev/null; then
    echo "✓ Chromium 安装成功: $(chromium-browser --version)"
    CHROMIUM_PATH=$(which chromium-browser)
elif [ -f /snap/bin/chromium ]; then
    echo "✓ Chromium 安装成功 (snap): $(/snap/bin/chromium --version)"
    CHROMIUM_PATH="/snap/bin/chromium"
else
    echo "✗ Chromium 安装失败，请手动安装"
    echo ""
    echo "手动安装方法："
    echo "  Ubuntu 20.04+: sudo snap install chromium"
    echo "  Ubuntu 18.04-: sudo apt install chromium-browser"
    exit 1
fi

echo ""
echo "Chromium 路径: $CHROMIUM_PATH"
echo ""
echo "安装完成！项目预览截图功能现在可以正常工作了。"

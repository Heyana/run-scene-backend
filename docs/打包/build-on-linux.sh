#!/bin/bash

# Linux 服务器本地打包脚本
# 用途：在服务器上直接编译打包，不依赖 GitHub Actions

set -e  # 遇到错误立即退出

# ============================================
# 【配置区域】- 修改这里的项目名称即可
# ============================================
PROJECT_NAME="test_template"          # 项目名称（用于二进制文件、服务名、目录名）
DEPLOY_DIR="/root/project"            # 部署目录
BACKEND_SUBDIR="后台"                  # 后台代码子目录名称
# ============================================

# 设置 Go 代理环境变量（国内加速）
export GO111MODULE=on
export GOPROXY=https://goproxy.cn,https://goproxy.io,direct
export GOSUMDB=sum.golang.google.cn
export GOTOOLCHAIN=local

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  ${PROJECT_NAME} Linux 本地打包脚本${NC}"
echo -e "${GREEN}========================================${NC}"

# 获取脚本所在目录的上上级目录（项目根目录）
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$( cd "$SCRIPT_DIR/../.." && pwd )"
BACKEND_DIR="$PROJECT_ROOT/${BACKEND_SUBDIR}"

echo -e "${BLUE}项目目录: ${GREEN}${PROJECT_ROOT}${NC}"
echo -e "${BLUE}后台目录: ${GREEN}${BACKEND_DIR}${NC}\n"

# 检查目录是否存在
if [ ! -d "$BACKEND_DIR" ]; then
    echo -e "${RED}错误: 找不到后台目录: $BACKEND_DIR${NC}"
    exit 1
fi

cd "$BACKEND_DIR"

echo -e "${BLUE}当前Go代理: ${GREEN}${GOPROXY}${NC}\n"

# 1. 检查并安装依赖
echo -e "${BLUE}[1/5] 检查依赖...${NC}"

# 自动安装 Go
if ! command -v go &> /dev/null; then
    echo -e "${YELLOW}未检测到 Go，正在自动安装...${NC}"
    
    GO_VERSION="1.23.3"
    GO_TAR="go${GO_VERSION}.linux-amd64.tar.gz"
    GO_URL="https://go.dev/dl/${GO_TAR}"
    GO_CN_URL="https://golang.google.cn/dl/${GO_TAR}"
    
    cd /tmp
    rm -f "${GO_TAR}"
    
    echo -e "${BLUE}正在下载 Go ${GO_VERSION}...${NC}"
    echo -e "${YELLOW}如果下载太慢，请按 Ctrl+C 取消，然后手动安装${NC}"
    
    # 先尝试国内镜像（通常更快）
    if wget --timeout=10 --tries=2 --progress=bar:force "${GO_CN_URL}" 2>&1 | grep -v "^--" | grep -v "^$"; then
        echo -e "${GREEN}✓ 从国内镜像下载成功${NC}"
    else
        echo -e "${YELLOW}国内镜像失败，尝试官方源...${NC}"
        rm -f "${GO_TAR}"
        if wget --timeout=30 --tries=2 --progress=bar:force "${GO_URL}" 2>&1 | grep -v "^--" | grep -v "^$"; then
            echo -e "${GREEN}✓ 从官方源下载成功${NC}"
        else
            echo -e "${RED}✗ 下载失败，请检查网络或手动安装${NC}"
            echo -e "${YELLOW}手动安装命令:${NC}"
            echo -e "  wget ${GO_CN_URL}"
            echo -e "  sudo tar -C /usr/local -xzf ${GO_TAR}"
            echo -e "  export PATH=\$PATH:/usr/local/go/bin"
            exit 1
        fi
    fi
    
    if [ ! -f "${GO_TAR}" ]; then
        echo -e "${RED}✗ 下载的文件不存在${NC}"
        exit 1
    fi
    
    echo -e "${BLUE}正在解压安装...${NC}"
    sudo rm -rf /usr/local/go
    sudo tar -C /usr/local -xzf "${GO_TAR}"
    
    # 添加到环境变量
    if ! grep -q "/usr/local/go/bin" ~/.bashrc; then
        echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
    fi
    export PATH=$PATH:/usr/local/go/bin
    
    rm -f "${GO_TAR}"
    cd - > /dev/null
    
    if command -v go &> /dev/null; then
        echo -e "${GREEN}✓ Go 安装成功: $(go version)${NC}"
    else
        echo -e "${RED}✗ Go 安装失败，请手动安装${NC}"
        exit 1
    fi
else
    GO_VERSION=$(go version)
    echo -e "${GREEN}✓ Go 已安装: ${GO_VERSION}${NC}"
fi

# 2. 准备静态文件目录
echo -e "\n${BLUE}[2/5] 准备静态文件目录...${NC}"

# 确保 static 目录存在
mkdir -p "$BACKEND_DIR/static"
echo -e "${GREEN}✓ static 目录已就绪${NC}"

# 确保 website 目录存在
mkdir -p "$BACKEND_DIR/website"
echo -e "${GREEN}✓ website 目录已就绪${NC}"

# 3. 编译 Go 程序
echo -e "\n${BLUE}[3/5] 编译 Go 程序...${NC}"
cd "$BACKEND_DIR"

echo -e "${YELLOW}开始编译...${NC}"

# 配置编译环境
export CGO_ENABLED=0
export GOOS=linux
export GOARCH=amd64

echo -e "${BLUE}编译目标: Linux AMD64 (CGO_ENABLED=0)${NC}"

# 先下载依赖
echo -e "${YELLOW}下载 Go 模块依赖...${NC}"
go mod download || {
    echo -e "${RED}下载依赖失败，尝试清理缓存...${NC}"
    go clean -modcache
    go mod download
}

# 编译
echo -e "${YELLOW}正在编译二进制文件...${NC}"
go build -tags dev -trimpath -ldflags="-s -w" -o "${PROJECT_NAME}" .

if [ ! -f "${PROJECT_NAME}" ]; then
    echo -e "${RED}错误: 编译失败，未找到 ${PROJECT_NAME} 文件${NC}"
    exit 1
fi

# 获取文件大小
FILE_SIZE=$(ls -lh "${PROJECT_NAME}" | awk '{print $5}')
echo -e "${GREEN}✓ 编译完成，文件大小: ${FILE_SIZE}${NC}"

# 4. 准备发布文件
echo -e "\n${BLUE}[4/5] 准备发布文件...${NC}"

RELEASE_DIR="$BACKEND_DIR/release/${PROJECT_NAME}"
rm -rf "$BACKEND_DIR/release"
mkdir -p "$RELEASE_DIR"

# 复制文件
cp "${PROJECT_NAME}" "$RELEASE_DIR/"
cp linux-online-show.service "$RELEASE_DIR/" 2>/dev/null || cp manager.service "$RELEASE_DIR/${PROJECT_NAME}.service" 2>/dev/null || true
[ -f "deploy-linux.sh" ] && cp deploy-linux.sh "$RELEASE_DIR/"
[ -f "hot-deploy.sh" ] && cp hot-deploy.sh "$RELEASE_DIR/"

# 复制文档
[ -f "docs/优雅启停.md" ] && cp docs/优雅启停.md "$RELEASE_DIR/"

echo -e "${GREEN}✓ 发布文件准备完成${NC}"

# 5. 打包
echo -e "\n${BLUE}[5/5] 打包...${NC}"
cd "$BACKEND_DIR/release"

TIMESTAMP=$(date +%Y%m%d_%H%M%S)
TAR_NAME="${PROJECT_NAME}-linux-amd64-${TIMESTAMP}.tar.gz"

tar -czf "$TAR_NAME" "${PROJECT_NAME}/"

TAR_SIZE=$(ls -lh "$TAR_NAME" | awk '{print $5}')
TAR_PATH="$BACKEND_DIR/release/$TAR_NAME"

echo -e "${GREEN}✓ 打包完成${NC}"

# 完成
echo -e "\n${GREEN}========================================${NC}"
echo -e "${GREEN}  打包成功！${NC}"
echo -e "${GREEN}========================================${NC}"

echo -e "\n${BLUE}打包文件信息:${NC}"
echo -e "  文件名: ${GREEN}${TAR_NAME}${NC}"
echo -e "  文件大小: ${GREEN}${TAR_SIZE}${NC}"
echo -e "  完整路径: ${GREEN}${TAR_PATH}${NC}"

echo -e "\n${BLUE}包含文件:${NC}"
tar -tzf "$TAR_NAME" | sed 's/^/  /'

echo -e "\n${BLUE}下一步操作:${NC}"
echo -e "  1. 使用 FileZilla 上传: ${YELLOW}${TAR_PATH}${NC}"
echo -e "  2. 在服务器解压: ${YELLOW}tar -xzf ${TAR_NAME}${NC}"
echo -e "  3. 替换旧文件并重启: ${YELLOW}systemctl restart ${PROJECT_NAME}.service${NC}"

echo -e "\n或使用热更新脚本:"
echo -e "  ${YELLOW}sudo bash hot-deploy.sh /path/to/online_show${NC}"

echo -e "\n${BLUE}========================================${NC}"
echo -e "${BLUE}  服务器快速部署流程${NC}"
echo -e "${BLUE}========================================${NC}"

echo -e "\n${YELLOW}步骤1: 上传文件到服务器${NC}"
echo -e "  使用 FileZilla 或 scp 上传到: ${GREEN}/root/project/${NC}"

echo -e "\n${YELLOW}步骤2: SSH登录服务器后执行以下命令${NC}"
cat << EOF

# 进入项目目录
cd ${DEPLOY_DIR}

# 优雅停止服务（等待现有请求完成）
sudo systemctl stop ${PROJECT_NAME}

# 备份旧版本（可选，建议）
if [ -d "${PROJECT_NAME}_backup" ]; then
    rm -rf ${PROJECT_NAME}_backup_old
    mv ${PROJECT_NAME}_backup ${PROJECT_NAME}_backup_old
fi
cp -r ${PROJECT_NAME} ${PROJECT_NAME}_backup

# 解压新版本（会覆盖旧文件）
tar -xzf FILENAME.tar.gz

# 启动服务
sudo systemctl start ${PROJECT_NAME}

# 查看服务状态
systemctl status ${PROJECT_NAME}

# 查看实时日志
# tail -f ${PROJECT_NAME}/app_log.txt

EOF

echo -e "\n${YELLOW}一键部署命令（复制粘贴）:${NC}"
echo -e "${GREEN}cd ${DEPLOY_DIR} && sudo systemctl stop ${PROJECT_NAME} && tar -xzf ${TAR_NAME} && sudo systemctl start ${PROJECT_NAME} && systemctl status ${PROJECT_NAME}${NC}"

echo -e "\n${BLUE}注意事项:${NC}"
echo -e "  • ${YELLOW}停止服务到启动服务之间尽量快速操作，减少停机时间${NC}"
echo -e "  • ${YELLOW}解压会覆盖现有文件，数据库和日志不会受影响${NC}"
echo -e "  • ${YELLOW}如遇问题可用备份快速回滚: rm -rf ${PROJECT_NAME} && mv ${PROJECT_NAME}_backup ${PROJECT_NAME}${NC}"

echo -e "\n${GREEN}打包完成！${NC}"

# 自动部署流程
echo -e "\n${GREEN}========================================${NC}"
echo -e "${GREEN}  开始自动部署${NC}"
echo -e "${GREEN}========================================${NC}"

# 移动到部署目录
echo -e "\n${BLUE}[1/5] 移动安装包到部署目录...${NC}"
echo -e "${YELLOW}目标目录: ${DEPLOY_DIR}${NC}"

if [ ! -d "$DEPLOY_DIR" ]; then
    echo -e "${YELLOW}目录不存在，正在创建...${NC}"
    sudo mkdir -p "$DEPLOY_DIR"
fi

sudo mv "$TAR_PATH" "$DEPLOY_DIR/"
echo -e "${GREEN}✓ 安装包已移动到: ${DEPLOY_DIR}/${TAR_NAME}${NC}"

# 进入部署目录
echo -e "\n${BLUE}[2/5] 进入部署目录...${NC}"
cd "$DEPLOY_DIR"
echo -e "${GREEN}✓ 当前目录: $(pwd)${NC}"

# 优雅停止服务
echo -e "\n${BLUE}[3/5] 优雅停止服务...${NC}"
if sudo systemctl is-active --quiet "${PROJECT_NAME}"; then
    echo -e "${YELLOW}等待现有请求完成（最多35秒）...${NC}"
    sudo systemctl stop "${PROJECT_NAME}"
    echo -e "${GREEN}✓ 服务已优雅停止${NC}"
else
    echo -e "${YELLOW}⚠ 服务未运行${NC}"
fi

# 快速解压
echo -e "\n${BLUE}[4/5] 快速解压安装包...${NC}"
sudo tar -xzf "$TAR_NAME"
echo -e "${GREEN}✓ 解压完成${NC}"

# 启动服务
echo -e "\n${BLUE}[5/5] 启动服务...${NC}"
sudo systemctl start "${PROJECT_NAME}"

# 等待服务启动
sleep 2

# 查看服务状态
echo -e "\n${GREEN}========================================${NC}"
echo -e "${GREEN}  服务状态${NC}"
echo -e "${GREEN}========================================${NC}\n"
sudo systemctl status "${PROJECT_NAME}" --no-pager

echo -e "\n${GREEN}========================================${NC}"
echo -e "${GREEN}  部署完成！${NC}"
echo -e "${GREEN}========================================${NC}"

echo -e "\n${BLUE}有用的命令:${NC}"
echo -e "  查看实时日志: ${YELLOW}tail -f ${DEPLOY_DIR}/${PROJECT_NAME}/app_log.txt${NC}"
echo -e "  重启服务: ${YELLOW}sudo systemctl restart ${PROJECT_NAME}${NC}"
echo -e "  停止服务: ${YELLOW}sudo systemctl stop ${PROJECT_NAME}${NC}"
echo -e "  查看状态: ${YELLOW}sudo systemctl status ${PROJECT_NAME}${NC}"

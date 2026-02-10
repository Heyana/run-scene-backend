#!/bin/bash
# 3D 模型预览图渲染脚本
# 支持 FBX, OBJ, GLTF 等格式

set -e

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 打印信息
info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

# 检查 Blender 是否安装
check_blender() {
    if ! command -v blender &> /dev/null; then
        error "Blender 未安装"
        echo "请运行以下命令安装 Blender:"
        echo "  sudo apt install -y blender"
        exit 1
    fi
    
    info "Blender 已安装: $(blender --version | head -n 1)"
}

# 使用说明
usage() {
    cat << EOF
使用方法: $0 <input_file> <output_file> [width] [height]

参数:
    input_file   - 输入的 3D 模型文件 (支持 .fbx, .obj, .gltf, .glb)
    output_file  - 输出的预览图文件 (支持 .png, .jpg, .webp)
    width        - 可选，图片宽度，默认 1280
    height       - 可选，图片高度，默认 720

示例:
    $0 model.fbx preview.png
    $0 model.obj preview.jpg 1920 1080
    $0 model.gltf preview.webp 1280 720

EOF
    exit 1
}

# 主函数
main() {
    # 检查参数
    if [ $# -lt 2 ]; then
        usage
    fi
    
    INPUT_FILE="$1"
    OUTPUT_FILE="$2"
    WIDTH="${3:-1280}"
    HEIGHT="${4:-720}"
    
    # 检查输入文件
    if [ ! -f "$INPUT_FILE" ]; then
        error "输入文件不存在: $INPUT_FILE"
        exit 1
    fi
    
    # 获取文件扩展名
    EXT="${INPUT_FILE##*.}"
    EXT_LOWER=$(echo "$EXT" | tr '[:upper:]' '[:lower:]')
    
    # 检查文件格式
    case "$EXT_LOWER" in
        fbx|obj|gltf|glb)
            info "检测到 3D 模型格式: $EXT_LOWER"
            ;;
        *)
            error "不支持的文件格式: $EXT"
            echo "支持的格式: fbx, obj, gltf, glb"
            exit 1
            ;;
    esac
    
    # 检查 Blender
    check_blender
    
    # 获取脚本目录
    SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    RENDER_SCRIPT="$SCRIPT_DIR/render_fbx.py"
    
    if [ ! -f "$RENDER_SCRIPT" ]; then
        error "渲染脚本不存在: $RENDER_SCRIPT"
        exit 1
    fi
    
    # 创建输出目录
    OUTPUT_DIR="$(dirname "$OUTPUT_FILE")"
    if [ ! -d "$OUTPUT_DIR" ]; then
        mkdir -p "$OUTPUT_DIR"
        info "创建输出目录: $OUTPUT_DIR"
    fi
    
    # 执行渲染
    info "开始渲染预览图..."
    info "输入: $INPUT_FILE"
    info "输出: $OUTPUT_FILE"
    info "分辨率: ${WIDTH}x${HEIGHT}"
    
    # 临时 PNG 文件（如果输出格式不是 PNG）
    TEMP_PNG=""
    FINAL_OUTPUT="$OUTPUT_FILE"
    
    OUTPUT_EXT="${OUTPUT_FILE##*.}"
    OUTPUT_EXT_LOWER=$(echo "$OUTPUT_EXT" | tr '[:upper:]' '[:lower:]')
    
    if [ "$OUTPUT_EXT_LOWER" != "png" ]; then
        TEMP_PNG="${OUTPUT_FILE%.*}.temp.png"
        OUTPUT_FILE="$TEMP_PNG"
    fi
    
    # 运行 Blender 渲染
    if blender -b -P "$RENDER_SCRIPT" -- "$INPUT_FILE" "$OUTPUT_FILE" "$WIDTH" "$HEIGHT" 2>&1 | grep -v "^Info:"; then
        info "Blender 渲染完成"
        
        # 如果需要转换格式
        if [ -n "$TEMP_PNG" ]; then
            info "转换图片格式: PNG -> $OUTPUT_EXT_LOWER"
            
            if command -v convert &> /dev/null; then
                # 使用 ImageMagick 转换
                convert "$TEMP_PNG" "$FINAL_OUTPUT"
                rm -f "$TEMP_PNG"
                info "格式转换完成"
            elif command -v magick &> /dev/null; then
                # 使用 ImageMagick 7
                magick "$TEMP_PNG" "$FINAL_OUTPUT"
                rm -f "$TEMP_PNG"
                info "格式转换完成"
            else
                warn "ImageMagick 未安装，保留 PNG 格式"
                mv "$TEMP_PNG" "${FINAL_OUTPUT%.*}.png"
                FINAL_OUTPUT="${FINAL_OUTPUT%.*}.png"
            fi
        fi
        
        info "预览图已保存: $FINAL_OUTPUT"
        
        # 显示文件信息
        if [ -f "$FINAL_OUTPUT" ]; then
            FILE_SIZE=$(du -h "$FINAL_OUTPUT" | cut -f1)
            info "文件大小: $FILE_SIZE"
        fi
        
        exit 0
    else
        error "渲染失败"
        
        # 清理临时文件
        [ -n "$TEMP_PNG" ] && rm -f "$TEMP_PNG"
        
        exit 1
    fi
}

# 执行主函数
main "$@"

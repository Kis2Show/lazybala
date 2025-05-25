#!/bin/bash

# LazyBala 安装脚本
# 自动检测系统架构并下载对应的二进制文件

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 项目信息
REPO="kis2show/lazybala"
BINARY_NAME="lazybala"
INSTALL_DIR="/usr/local/bin"

# 函数：打印彩色消息
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 函数：检测系统架构
detect_platform() {
    local os=$(uname -s | tr '[:upper:]' '[:lower:]')
    local arch=$(uname -m)
    
    case $os in
        linux*)
            OS="linux"
            ;;
        darwin*)
            OS="darwin"
            ;;
        freebsd*)
            OS="freebsd"
            ;;
        *)
            print_error "不支持的操作系统: $os"
            exit 1
            ;;
    esac
    
    case $arch in
        x86_64|amd64)
            ARCH="amd64"
            ;;
        arm64|aarch64)
            ARCH="arm64"
            ;;
        armv7l)
            ARCH="armv7"
            ;;
        *)
            print_error "不支持的架构: $arch"
            exit 1
            ;;
    esac
    
    PLATFORM="${OS}-${ARCH}"
    print_info "检测到平台: $PLATFORM"
}

# 函数：获取最新版本
get_latest_version() {
    print_info "获取最新版本信息..."
    
    if command -v curl >/dev/null 2>&1; then
        VERSION=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    elif command -v wget >/dev/null 2>&1; then
        VERSION=$(wget -qO- "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    else
        print_error "需要 curl 或 wget 来下载文件"
        exit 1
    fi
    
    if [ -z "$VERSION" ]; then
        print_error "无法获取最新版本信息"
        exit 1
    fi
    
    print_info "最新版本: $VERSION"
}

# 函数：下载二进制文件
download_binary() {
    local filename="${BINARY_NAME}-${PLATFORM}.tar.gz"
    local url="https://github.com/$REPO/releases/download/$VERSION/$filename"
    local temp_dir=$(mktemp -d)
    
    print_info "下载 $filename..."
    
    cd "$temp_dir"
    
    if command -v curl >/dev/null 2>&1; then
        curl -L "$url" -o "$filename"
    elif command -v wget >/dev/null 2>&1; then
        wget "$url" -O "$filename"
    else
        print_error "需要 curl 或 wget 来下载文件"
        exit 1
    fi
    
    if [ ! -f "$filename" ]; then
        print_error "下载失败: $filename"
        exit 1
    fi
    
    print_info "解压文件..."
    tar -xzf "$filename"
    
    # 查找二进制文件
    BINARY_FILE=$(find . -name "${BINARY_NAME}-${PLATFORM}" -type f)
    
    if [ -z "$BINARY_FILE" ]; then
        print_error "未找到二进制文件"
        exit 1
    fi
    
    chmod +x "$BINARY_FILE"
    
    # 移动到临时位置
    mv "$BINARY_FILE" "/tmp/$BINARY_NAME"
    
    cd - >/dev/null
    rm -rf "$temp_dir"
}

# 函数：安装二进制文件
install_binary() {
    print_info "安装 $BINARY_NAME 到 $INSTALL_DIR..."
    
    # 检查是否需要 sudo
    if [ -w "$INSTALL_DIR" ]; then
        mv "/tmp/$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"
    else
        print_warning "需要管理员权限安装到 $INSTALL_DIR"
        sudo mv "/tmp/$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"
    fi
    
    # 验证安装
    if [ -f "$INSTALL_DIR/$BINARY_NAME" ]; then
        print_success "$BINARY_NAME 安装成功!"
    else
        print_error "安装失败"
        exit 1
    fi
}

# 函数：验证安装
verify_installation() {
    print_info "验证安装..."
    
    if command -v "$BINARY_NAME" >/dev/null 2>&1; then
        local version_output=$($BINARY_NAME --version 2>/dev/null || echo "LazyBala $VERSION")
        print_success "安装验证成功: $version_output"
    else
        print_warning "二进制文件已安装，但不在 PATH 中"
        print_info "请将 $INSTALL_DIR 添加到您的 PATH 环境变量中"
    fi
}

# 函数：显示使用说明
show_usage() {
    echo ""
    print_success "🎉 LazyBala 安装完成!"
    echo ""
    echo "使用方法:"
    echo "  $BINARY_NAME                    # 启动服务器"
    echo "  $BINARY_NAME --help             # 显示帮助信息"
    echo "  $BINARY_NAME --version          # 显示版本信息"
    echo ""
    echo "Web 界面:"
    echo "  启动后访问: http://localhost:8080"
    echo ""
    echo "文档链接:"
    echo "  GitHub: https://github.com/$REPO"
    echo "  Issues: https://github.com/$REPO/issues"
    echo ""
}

# 函数：清理
cleanup() {
    if [ -f "/tmp/$BINARY_NAME" ]; then
        rm -f "/tmp/$BINARY_NAME"
    fi
}

# 主函数
main() {
    echo "🚀 LazyBala 安装脚本"
    echo "===================="
    echo ""
    
    # 设置清理陷阱
    trap cleanup EXIT
    
    # 检查依赖
    if ! command -v tar >/dev/null 2>&1; then
        print_error "需要 tar 命令来解压文件"
        exit 1
    fi
    
    # 执行安装步骤
    detect_platform
    get_latest_version
    download_binary
    install_binary
    verify_installation
    show_usage
}

# 处理命令行参数
while [[ $# -gt 0 ]]; do
    case $1 in
        --version)
            VERSION="$2"
            shift 2
            ;;
        --install-dir)
            INSTALL_DIR="$2"
            shift 2
            ;;
        --help)
            echo "LazyBala 安装脚本"
            echo ""
            echo "用法: $0 [选项]"
            echo ""
            echo "选项:"
            echo "  --version VERSION     指定要安装的版本"
            echo "  --install-dir DIR     指定安装目录 (默认: /usr/local/bin)"
            echo "  --help               显示此帮助信息"
            echo ""
            exit 0
            ;;
        *)
            print_error "未知选项: $1"
            echo "使用 --help 查看帮助信息"
            exit 1
            ;;
    esac
done

# 执行主函数
main

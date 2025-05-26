#!/bin/bash
# yt-dlp 升级脚本

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 配置
BIN_DIR="./data/bin"
YT_DLP_PATH="$BIN_DIR/yt-dlp"
BACKUP_PATH="$BIN_DIR/yt-dlp.backup"

# 函数：打印带颜色的消息
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

# 函数：检测架构
detect_arch() {
    local arch=$(uname -m)
    case $arch in
        x86_64)
            echo "yt-dlp_linux"
            ;;
        aarch64)
            echo "yt-dlp_linux_aarch64"
            ;;
        armv7l)
            echo "yt-dlp_linux_armv7l"
            ;;
        *)
            echo "yt-dlp_linux"
            ;;
    esac
}

# 函数：获取当前版本
get_current_version() {
    if [ -f "$YT_DLP_PATH" ] && [ -x "$YT_DLP_PATH" ]; then
        $YT_DLP_PATH --version 2>/dev/null || echo "unknown"
    else
        echo "not installed"
    fi
}

# 函数：获取最新版本
get_latest_version() {
    curl -s https://api.github.com/repos/yt-dlp/yt-dlp/releases/latest | \
        grep '"tag_name":' | \
        sed -E 's/.*"([^"]+)".*/\1/'
}

# 函数：下载 yt-dlp
download_ytdlp() {
    local filename=$1
    local url="https://github.com/yt-dlp/yt-dlp/releases/latest/download/$filename"
    
    print_info "下载 $filename..."
    
    # 创建临时文件
    local temp_file=$(mktemp)
    
    # 下载文件
    if curl -L -o "$temp_file" "$url"; then
        # 验证下载的文件
        if [ -s "$temp_file" ]; then
            # 备份现有文件
            if [ -f "$YT_DLP_PATH" ]; then
                print_info "备份现有版本到 $BACKUP_PATH"
                cp "$YT_DLP_PATH" "$BACKUP_PATH"
            fi
            
            # 移动新文件
            mv "$temp_file" "$YT_DLP_PATH"
            chmod +x "$YT_DLP_PATH"
            
            print_success "yt-dlp 下载完成"
            return 0
        else
            print_error "下载的文件为空"
            rm -f "$temp_file"
            return 1
        fi
    else
        print_error "下载失败"
        rm -f "$temp_file"
        return 1
    fi
}

# 函数：验证安装
verify_installation() {
    if [ -f "$YT_DLP_PATH" ] && [ -x "$YT_DLP_PATH" ]; then
        local version=$($YT_DLP_PATH --version 2>/dev/null)
        if [ $? -eq 0 ]; then
            print_success "yt-dlp 验证成功，版本: $version"
            return 0
        else
            print_error "yt-dlp 验证失败"
            return 1
        fi
    else
        print_error "yt-dlp 文件不存在或不可执行"
        return 1
    fi
}

# 函数：恢复备份
restore_backup() {
    if [ -f "$BACKUP_PATH" ]; then
        print_warning "恢复备份版本..."
        cp "$BACKUP_PATH" "$YT_DLP_PATH"
        chmod +x "$YT_DLP_PATH"
        print_success "已恢复备份版本"
    else
        print_error "没有找到备份文件"
    fi
}

# 主函数
main() {
    print_info "LazyBala yt-dlp 升级工具"
    echo "================================"
    
    # 创建 bin 目录
    mkdir -p "$BIN_DIR"
    
    # 检测架构
    local filename=$(detect_arch)
    print_info "检测到架构: $filename"
    
    # 获取版本信息
    local current_version=$(get_current_version)
    print_info "当前版本: $current_version"
    
    local latest_version=$(get_latest_version)
    if [ -z "$latest_version" ]; then
        print_error "无法获取最新版本信息"
        exit 1
    fi
    print_info "最新版本: $latest_version"
    
    # 检查是否需要更新
    if [ "$current_version" = "$latest_version" ]; then
        print_success "yt-dlp 已是最新版本"
        exit 0
    fi
    
    # 下载新版本
    if download_ytdlp "$filename"; then
        # 验证安装
        if verify_installation; then
            print_success "yt-dlp 升级成功！"
            print_info "从 $current_version 升级到 $latest_version"
            
            # 清理备份（可选）
            # rm -f "$BACKUP_PATH"
        else
            print_error "新版本验证失败，恢复备份"
            restore_backup
            exit 1
        fi
    else
        print_error "下载失败"
        exit 1
    fi
}

# 处理命令行参数
case "${1:-}" in
    --help|-h)
        echo "用法: $0 [选项]"
        echo ""
        echo "选项:"
        echo "  --help, -h     显示帮助信息"
        echo "  --force, -f    强制重新下载"
        echo "  --restore, -r  恢复备份版本"
        echo "  --version, -v  显示当前版本"
        echo ""
        echo "示例:"
        echo "  $0              # 检查并升级到最新版本"
        echo "  $0 --force      # 强制重新下载最新版本"
        echo "  $0 --restore    # 恢复备份版本"
        exit 0
        ;;
    --force|-f)
        print_info "强制重新下载模式"
        # 删除当前版本以强制下载
        rm -f "$YT_DLP_PATH"
        main
        ;;
    --restore|-r)
        restore_backup
        verify_installation
        ;;
    --version|-v)
        echo "当前版本: $(get_current_version)"
        ;;
    "")
        main
        ;;
    *)
        print_error "未知选项: $1"
        echo "使用 --help 查看帮助信息"
        exit 1
        ;;
esac

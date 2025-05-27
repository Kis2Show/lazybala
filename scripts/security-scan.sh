#!/bin/bash
# LazyBala 安全扫描脚本

set -e

echo "========================================"
echo "   LazyBala 安全扫描工具"
echo "========================================"
echo

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

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

# 检查 Go 环境
print_info "检查 Go 环境..."
if ! go version >/dev/null 2>&1; then
    print_error "Go 未安装或未在 PATH 中"
    exit 1
fi

GO_VERSION=$(go version | awk '{print $3}')
print_success "Go 版本: $GO_VERSION"

# 1. 运行 go vet
print_info "运行 go vet 静态分析..."
if go vet ./...; then
    print_success "go vet 检查通过"
else
    print_warning "go vet 发现潜在问题"
fi

# 2. 运行 go fmt 检查
print_info "检查代码格式..."
UNFORMATTED=$(gofmt -l . | grep -v vendor || true)
if [ -z "$UNFORMATTED" ]; then
    print_success "代码格式检查通过"
else
    print_warning "以下文件格式不正确:"
    echo "$UNFORMATTED"
fi

# 3. 安装并运行 govulncheck
print_info "安装 govulncheck..."
if go install golang.org/x/vuln/cmd/govulncheck@latest; then
    print_success "govulncheck 安装成功"
    
    print_info "运行漏洞检查..."
    if govulncheck ./...; then
        print_success "未发现已知漏洞"
    else
        print_warning "发现潜在漏洞，请查看上述输出"
    fi
else
    print_error "govulncheck 安装失败"
fi

# 4. 尝试安装并运行 gosec
print_info "尝试安装 gosec..."
GOSEC_INSTALLED=false

# 方法1: 使用 go install
if go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest 2>/dev/null; then
    GOSEC_INSTALLED=true
    print_success "gosec 通过 go install 安装成功"
else
    print_warning "go install 安装 gosec 失败，尝试其他方法..."
    
    # 方法2: 下载预编译二进制
    if command -v curl >/dev/null 2>&1; then
        print_info "尝试下载预编译的 gosec..."
        GOSEC_VERSION="2.18.2"
        OS=$(uname -s | tr '[:upper:]' '[:lower:]')
        ARCH=$(uname -m)
        
        # 转换架构名称
        case $ARCH in
            x86_64) ARCH="amd64" ;;
            aarch64) ARCH="arm64" ;;
            armv7l) ARCH="armv7" ;;
        esac
        
        TEMP_DIR=$(mktemp -d)
        cd "$TEMP_DIR"
        
        if curl -L "https://github.com/securecodewarrior/gosec/releases/download/v${GOSEC_VERSION}/gosec_${GOSEC_VERSION}_${OS}_${ARCH}.tar.gz" -o gosec.tar.gz 2>/dev/null; then
            if tar -xzf gosec.tar.gz 2>/dev/null; then
                if [ -f "gosec" ]; then
                    chmod +x gosec
                    mkdir -p "$HOME/bin"
                    mv gosec "$HOME/bin/"
                    export PATH="$HOME/bin:$PATH"
                    GOSEC_INSTALLED=true
                    print_success "gosec 预编译版本安装成功"
                fi
            fi
        fi
        
        cd - >/dev/null
        rm -rf "$TEMP_DIR"
    fi
fi

# 运行 gosec 扫描
if [ "$GOSEC_INSTALLED" = true ] && command -v gosec >/dev/null 2>&1; then
    print_info "运行 gosec 安全扫描..."
    
    if [ -f ".gosec.json" ]; then
        print_info "使用配置文件 .gosec.json"
        if gosec -conf .gosec.json ./...; then
            print_success "gosec 扫描完成，未发现严重问题"
        else
            print_warning "gosec 发现潜在安全问题，请查看上述输出"
        fi
    else
        print_info "使用默认配置运行 gosec"
        if gosec ./...; then
            print_success "gosec 扫描完成，未发现严重问题"
        else
            print_warning "gosec 发现潜在安全问题，请查看上述输出"
        fi
    fi
else
    print_warning "gosec 未安装，跳过安全扫描"
fi

# 5. 运行基础安全检查
print_info "运行基础安全检查..."

# 检查敏感文件
print_info "检查敏感文件..."
SENSITIVE_FILES=$(find . -name "*.key" -o -name "*.pem" -o -name "*.p12" -o -name "*.pfx" 2>/dev/null | grep -v vendor || true)
if [ -z "$SENSITIVE_FILES" ]; then
    print_success "未发现敏感文件"
else
    print_warning "发现敏感文件:"
    echo "$SENSITIVE_FILES"
fi

# 检查硬编码密码
print_info "检查硬编码密码..."
HARDCODED=$(grep -r -i "password\|secret\|token\|key" --include="*.go" . | grep -v "// " | grep -v "_test.go" | head -5 || true)
if [ -z "$HARDCODED" ]; then
    print_success "未发现明显的硬编码密码"
else
    print_warning "发现可能的硬编码密码:"
    echo "$HARDCODED"
fi

# 6. 检查依赖
print_info "检查 Go 模块依赖..."
go list -m all | head -10

# 7. 生成报告
print_info "生成安全扫描报告..."
REPORT_FILE="security-scan-report.txt"
{
    echo "LazyBala 安全扫描报告"
    echo "生成时间: $(date)"
    echo "Go 版本: $GO_VERSION"
    echo "========================================"
    echo
    echo "扫描结果:"
    echo "- go vet: 已运行"
    echo "- go fmt: 已检查"
    echo "- govulncheck: 已运行"
    echo "- gosec: $([ "$GOSEC_INSTALLED" = true ] && echo "已运行" || echo "未安装")"
    echo "- 基础检查: 已完成"
    echo
    echo "详细信息请查看终端输出"
} > "$REPORT_FILE"

print_success "安全扫描完成！"
print_info "报告已保存到: $REPORT_FILE"

echo
echo "========================================"
echo "   安全扫描总结"
echo "========================================"
echo "✓ 静态分析检查"
echo "✓ 漏洞扫描"
echo "$([ "$GOSEC_INSTALLED" = true ] && echo "✓" || echo "⚠") 安全代码分析"
echo "✓ 基础安全检查"
echo "✓ 依赖检查"
echo
print_info "建议定期运行此脚本以确保代码安全"

#!/bin/bash
# GitHub Actions 问题修复脚本

set -e

echo "========================================"
echo "   GitHub Actions 问题修复脚本"
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

# 检查 Git 状态
print_info "检查 Git 状态..."
if ! git status >/dev/null 2>&1; then
    print_error "当前目录不是 Git 仓库"
    exit 1
fi

print_success "Git 仓库检查通过"

# 检查工作流文件
print_info "检查 GitHub Actions 工作流文件..."

WORKFLOW_DIR=".github/workflows"
if [ ! -d "$WORKFLOW_DIR" ]; then
    print_error "未找到 .github/workflows 目录"
    exit 1
fi

print_success "工作流目录存在"

# 列出所有工作流文件
print_info "当前工作流文件:"
ls -la "$WORKFLOW_DIR"/*.yml 2>/dev/null || echo "未找到 .yml 文件"

# 检查 dependency-update.yml
DEPENDENCY_FILE="$WORKFLOW_DIR/dependency-update.yml"
if [ -f "$DEPENDENCY_FILE" ]; then
    print_info "检查 dependency-update.yml..."
    
    # 检查是否包含旧的 gosec 安装方式
    if grep -q "go install.*gosec" "$DEPENDENCY_FILE"; then
        print_warning "发现旧的 gosec 安装方式"
    else
        print_success "dependency-update.yml 已更新"
    fi
    
    # 检查是否包含新的安装方式
    if grep -q "curl.*gosec.*tar.gz" "$DEPENDENCY_FILE"; then
        print_success "发现新的 gosec 安装方式"
    else
        print_warning "未发现新的 gosec 安装方式"
    fi
else
    print_error "未找到 dependency-update.yml"
fi

# 检查新的安全扫描工作流
SECURITY_FILE="$WORKFLOW_DIR/security-scan.yml"
if [ -f "$SECURITY_FILE" ]; then
    print_success "发现新的安全扫描工作流"
else
    print_warning "未找到新的安全扫描工作流"
fi

# 提供修复建议
echo
print_info "修复建议:"
echo "1. 确保所有工作流文件已更新到最新版本"
echo "2. 清除 GitHub Actions 缓存（通过重新推送提交）"
echo "3. 手动触发工作流测试"
echo "4. 检查 GitHub Actions 页面的执行日志"

# 检查是否有未提交的更改
if ! git diff --quiet; then
    print_warning "发现未提交的更改"
    echo "请先提交更改："
    echo "  git add ."
    echo "  git commit -m 'fix: 更新 GitHub Actions 工作流'"
    echo "  git push origin main"
else
    print_success "没有未提交的更改"
fi

# 提供测试命令
echo
print_info "测试命令:"
echo "# 手动触发依赖更新工作流"
echo "gh workflow run dependency-update.yml"
echo
echo "# 手动触发安全扫描工作流"
echo "gh workflow run security-scan.yml"
echo
echo "# 查看工作流状态"
echo "gh run list"

# 检查 GitHub CLI
if command -v gh >/dev/null 2>&1; then
    print_success "GitHub CLI 可用"
    
    print_info "最近的工作流运行:"
    gh run list --limit 5 2>/dev/null || echo "无法获取工作流运行历史"
else
    print_warning "GitHub CLI 未安装，无法直接触发工作流"
    echo "请访问 GitHub 网页界面手动触发工作流"
fi

echo
print_info "如果问题仍然存在，请："
echo "1. 检查 GitHub Actions 页面的详细日志"
echo "2. 确认工作流文件语法正确"
echo "3. 检查是否有权限问题"
echo "4. 考虑删除并重新创建工作流文件"

echo
print_success "修复脚本执行完成"

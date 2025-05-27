#!/bin/bash
# 测试 Docker yt-dlp 修复脚本

set -e

echo "========================================"
echo "   LazyBala Docker yt-dlp 修复测试"
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

# 检查 Docker 是否运行
print_info "检查 Docker 环境..."
if ! docker --version >/dev/null 2>&1; then
    print_error "Docker 未安装或未运行"
    exit 1
fi

if ! docker-compose --version >/dev/null 2>&1; then
    print_error "docker-compose 未安装"
    exit 1
fi

print_success "Docker 环境检查通过"

# 停止现有容器
print_info "停止现有容器..."
docker-compose down 2>/dev/null || true

# 清理旧镜像（可选）
if [ "$1" = "--clean" ]; then
    print_info "清理旧镜像..."
    docker rmi lazybala:latest 2>/dev/null || true
    docker system prune -f
fi

# 构建新镜像
print_info "构建新镜像..."
if docker-compose build --no-cache; then
    print_success "镜像构建成功"
else
    print_error "镜像构建失败"
    exit 1
fi

# 启动容器
print_info "启动容器..."
if docker-compose up -d; then
    print_success "容器启动成功"
else
    print_error "容器启动失败"
    exit 1
fi

# 等待容器完全启动
print_info "等待容器启动..."
sleep 10

# 检查容器状态
print_info "检查容器状态..."
if docker-compose ps | grep -q "Up"; then
    print_success "容器运行正常"
else
    print_error "容器未正常运行"
    docker-compose logs
    exit 1
fi

# 检查 yt-dlp 是否存在
print_info "检查 yt-dlp 文件..."
if docker exec lazybala test -f /app/bin/yt-dlp; then
    print_success "yt-dlp 文件存在"
else
    print_error "yt-dlp 文件不存在"
    docker exec lazybala ls -la /app/bin/
    exit 1
fi

# 检查 yt-dlp 权限
print_info "检查 yt-dlp 权限..."
if docker exec lazybala test -x /app/bin/yt-dlp; then
    print_success "yt-dlp 具有执行权限"
else
    print_error "yt-dlp 缺少执行权限"
    docker exec lazybala ls -la /app/bin/yt-dlp
    exit 1
fi

# 测试 yt-dlp 版本
print_info "测试 yt-dlp 版本..."
YT_VERSION=$(docker exec lazybala /app/bin/yt-dlp --version 2>/dev/null || echo "failed")
if [ "$YT_VERSION" != "failed" ]; then
    print_success "yt-dlp 版本: $YT_VERSION"
else
    print_error "yt-dlp 无法执行"
    docker exec lazybala /app/bin/yt-dlp --version
    exit 1
fi

# 检查应用是否响应
print_info "检查应用响应..."
sleep 5

# 测试健康检查端点
if curl -s http://localhost:8080/health >/dev/null; then
    print_success "健康检查端点响应正常"
else
    print_warning "健康检查端点无响应，检查主页..."
    if curl -s http://localhost:8080/ >/dev/null; then
        print_success "主页响应正常"
    else
        print_warning "应用可能还在启动中，请稍后手动检查"
    fi
fi

# 检查 Docker 健康状态
print_info "检查 Docker 健康状态..."
HEALTH_STATUS=$(docker inspect --format='{{.State.Health.Status}}' lazybala 2>/dev/null || echo "unknown")
if [ "$HEALTH_STATUS" = "healthy" ]; then
    print_success "Docker 健康检查: $HEALTH_STATUS"
elif [ "$HEALTH_STATUS" = "starting" ]; then
    print_info "Docker 健康检查: $HEALTH_STATUS (正在启动)"
else
    print_warning "Docker 健康检查: $HEALTH_STATUS"
fi

# 显示日志
print_info "显示启动日志..."
docker-compose logs --tail=20

echo
print_success "所有测试通过！"
echo
print_info "访问地址: http://localhost:8080"
print_info "查看日志: docker-compose logs -f"
print_info "停止服务: docker-compose down"
echo

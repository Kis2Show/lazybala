#!/bin/bash
# 群晖 NAS 部署脚本

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

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

# 配置
DOCKER_DIR="/volume1/docker/lazybala"
COMPOSE_FILE="docker-compose.synology.yml"

print_info "LazyBala 群晖 NAS 部署脚本"
echo "================================"

# 检查是否在群晖系统上
if [ ! -d "/volume1" ]; then
    print_warning "警告：未检测到群晖系统，但继续执行部署"
fi

# 创建必要的目录
print_info "创建数据目录..."
sudo mkdir -p "$DOCKER_DIR"/{audiobooks,config,cookies,bin}

# 设置目录权限和所有权
print_info "设置目录权限和所有权..."
sudo chmod -R 755 "$DOCKER_DIR"
sudo chmod -R 777 "$DOCKER_DIR"/{audiobooks,config,cookies,bin}
# 设置为群晖标准用户组 (PUID=100, PGID=100)
sudo chown -R 100:100 "$DOCKER_DIR"

# 检查 Docker 和 Docker Compose
print_info "检查 Docker 环境..."
if ! command -v docker &> /dev/null; then
    print_error "Docker 未安装或不在 PATH 中"
    exit 1
fi

if ! command -v docker-compose &> /dev/null; then
    print_error "Docker Compose 未安装或不在 PATH 中"
    exit 1
fi

# 停止现有容器（如果存在）
print_info "停止现有容器..."
docker-compose -f "$COMPOSE_FILE" down 2>/dev/null || true

# 构建并启动容器
print_info "构建并启动 LazyBala..."
export VERSION=$(git describe --tags --always 2>/dev/null || echo "latest")
export BUILDTIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
export PUID=100
export PGID=100

docker-compose -f "$COMPOSE_FILE" up -d --build

# 等待容器启动
print_info "等待容器启动..."
sleep 10

# 检查容器状态
if docker-compose -f "$COMPOSE_FILE" ps | grep -q "Up"; then
    print_success "LazyBala 部署成功！"
    print_info "访问地址: http://$(hostname -I | awk '{print $1}'):8080"
    print_info "或者: http://localhost:8080"
else
    print_error "容器启动失败，请检查日志："
    docker-compose -f "$COMPOSE_FILE" logs
    exit 1
fi

# 显示容器信息
print_info "容器状态："
docker-compose -f "$COMPOSE_FILE" ps

print_info "查看日志："
echo "docker-compose -f $COMPOSE_FILE logs -f"

print_info "停止服务："
echo "docker-compose -f $COMPOSE_FILE down"

print_success "部署完成！"

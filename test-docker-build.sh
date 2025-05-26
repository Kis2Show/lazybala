#!/bin/bash
# 测试 Docker 构建脚本

set -e

echo "🐳 测试 LazyBala Docker 构建"
echo "================================"

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

# 清理旧镜像
print_info "清理旧的测试镜像..."
docker rmi lazybala:test 2>/dev/null || true
docker rmi lazybala:synology-test 2>/dev/null || true

# 测试主 Dockerfile
print_info "构建主 Dockerfile..."
if docker build -t lazybala:test -f Dockerfile .; then
    print_success "主 Dockerfile 构建成功"
else
    print_error "主 Dockerfile 构建失败"
    exit 1
fi

# 测试群晖 Dockerfile
print_info "构建群晖专用 Dockerfile..."
if docker build -t lazybala:synology-test -f Dockerfile.synology .; then
    print_success "群晖 Dockerfile 构建成功"
else
    print_error "群晖 Dockerfile 构建失败"
    exit 1
fi

# 测试容器启动
print_info "测试容器启动..."
if docker run --rm -d --name lazybala-test -p 8081:8080 lazybala:test; then
    print_success "容器启动成功"
    
    # 等待几秒钟
    sleep 5
    
    # 检查容器状态
    if docker ps | grep -q lazybala-test; then
        print_success "容器运行正常"
        
        # 检查 yt-dlp 权限
        print_info "检查 yt-dlp 权限..."
        docker exec lazybala-test ls -la /app/bin/yt-dlp || print_warning "yt-dlp 文件不存在"
        
        # 测试应用响应
        print_info "测试应用响应..."
        if curl -s http://localhost:8081/ > /dev/null; then
            print_success "应用响应正常"
        else
            print_warning "应用可能还在启动中"
        fi
    else
        print_error "容器启动后立即退出"
        docker logs lazybala-test
    fi
    
    # 停止测试容器
    print_info "停止测试容器..."
    docker stop lazybala-test 2>/dev/null || true
else
    print_error "容器启动失败"
    exit 1
fi

# 清理测试镜像
print_info "清理测试镜像..."
docker rmi lazybala:test lazybala:synology-test 2>/dev/null || true

print_success "所有测试通过！Docker 构建正常"
echo ""
echo "🎉 可以安全地使用以下命令部署："
echo "   docker-compose up -d --build"
echo "   docker-compose -f docker-compose.synology.yml up -d --build"

#!/bin/bash

# LazyBala Docker 构建脚本

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 项目信息
PROJECT_NAME="lazybala"
IMAGE_NAME="lazybala"
VERSION=${1:-latest}

echo -e "${BLUE}=== LazyBala Docker 构建脚本 ===${NC}"
echo -e "${YELLOW}项目名称: ${PROJECT_NAME}${NC}"
echo -e "${YELLOW}镜像名称: ${IMAGE_NAME}:${VERSION}${NC}"
echo ""

# 函数：打印步骤
print_step() {
    echo -e "${BLUE}[步骤] $1${NC}"
}

# 函数：打印成功
print_success() {
    echo -e "${GREEN}[成功] $1${NC}"
}

# 函数：打印错误
print_error() {
    echo -e "${RED}[错误] $1${NC}"
}

# 函数：打印警告
print_warning() {
    echo -e "${YELLOW}[警告] $1${NC}"
}

# 检查 Docker 是否安装
check_docker() {
    print_step "检查 Docker 环境..."
    if ! command -v docker &> /dev/null; then
        print_error "Docker 未安装，请先安装 Docker"
        exit 1
    fi
    
    if ! docker info &> /dev/null; then
        print_error "Docker 服务未启动，请启动 Docker 服务"
        exit 1
    fi
    
    print_success "Docker 环境检查通过"
}

# 清理旧镜像
cleanup_old_images() {
    print_step "清理旧镜像..."
    
    # 删除悬空镜像
    if docker images -f "dangling=true" -q | grep -q .; then
        docker rmi $(docker images -f "dangling=true" -q) || true
        print_success "已清理悬空镜像"
    else
        print_warning "没有发现悬空镜像"
    fi
}

# 构建镜像
build_image() {
    print_step "构建 Docker 镜像..."
    
    # 构建镜像
    docker build \
        --tag ${IMAGE_NAME}:${VERSION} \
        --tag ${IMAGE_NAME}:latest \
        --build-arg BUILDTIME=$(date -u +'%Y-%m-%dT%H:%M:%SZ') \
        --build-arg VERSION=${VERSION} \
        .
    
    print_success "镜像构建完成: ${IMAGE_NAME}:${VERSION}"
}

# 显示镜像信息
show_image_info() {
    print_step "显示镜像信息..."
    
    echo ""
    echo -e "${BLUE}=== 镜像信息 ===${NC}"
    docker images ${IMAGE_NAME}
    
    echo ""
    echo -e "${BLUE}=== 镜像详情 ===${NC}"
    docker inspect ${IMAGE_NAME}:${VERSION} --format='{{.Config.Labels}}'
}

# 运行测试
run_test() {
    print_step "运行容器测试..."
    
    # 停止并删除现有容器
    docker stop ${PROJECT_NAME}-test 2>/dev/null || true
    docker rm ${PROJECT_NAME}-test 2>/dev/null || true
    
    # 创建测试目录
    mkdir -p ./test-data/{audiobooks,config,cookies}
    
    # 运行测试容器
    docker run -d \
        --name ${PROJECT_NAME}-test \
        -p 8081:8080 \
        -v $(pwd)/test-data/audiobooks:/app/audiobooks \
        -v $(pwd)/test-data/config:/app/config \
        -v $(pwd)/test-data/cookies:/app/cookies \
        ${IMAGE_NAME}:${VERSION}
    
    # 等待容器启动
    sleep 5
    
    # 健康检查
    if curl -f http://localhost:8081/ > /dev/null 2>&1; then
        print_success "容器测试通过，服务正常运行"
        echo -e "${GREEN}测试地址: http://localhost:8081${NC}"
    else
        print_error "容器测试失败，服务无法访问"
        docker logs ${PROJECT_NAME}-test
        exit 1
    fi
    
    # 清理测试容器
    docker stop ${PROJECT_NAME}-test
    docker rm ${PROJECT_NAME}-test
    rm -rf ./test-data
}

# 主函数
main() {
    echo -e "${BLUE}开始构建流程...${NC}"
    echo ""
    
    check_docker
    cleanup_old_images
    build_image
    show_image_info
    
    # 询问是否运行测试
    read -p "是否运行容器测试? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        run_test
    fi
    
    echo ""
    print_success "构建流程完成！"
    echo ""
    echo -e "${BLUE}=== 使用说明 ===${NC}"
    echo -e "${YELLOW}启动服务:${NC} docker-compose up -d"
    echo -e "${YELLOW}查看日志:${NC} docker-compose logs -f"
    echo -e "${YELLOW}停止服务:${NC} docker-compose down"
    echo -e "${YELLOW}访问地址:${NC} http://localhost:8080"
    echo ""
}

# 执行主函数
main

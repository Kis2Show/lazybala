# LazyBala Docker 部署指南

## 📋 概述

LazyBala 是一个基于 Go 和纯 JavaScript 的媒体下载应用，支持 Docker 容器化部署。

## 🚀 快速开始

### 方法一：使用 Docker Compose（推荐）

```bash
# 克隆项目
git clone <repository-url>
cd lazybala

# 启动服务
docker-compose up -d

# 查看日志
docker-compose logs -f

# 访问应用
open http://localhost:8080
```

### 方法二：使用构建脚本

#### Linux/macOS
```bash
# 给脚本执行权限
chmod +x build.sh

# 构建并测试
./build.sh latest
```

#### Windows (PowerShell)
```powershell
# 执行构建脚本
.\build.ps1 -Version latest
```

### 方法三：手动构建

```bash
# 构建镜像
docker build -t lazybala:latest .

# 运行容器
docker run -d \
  --name lazybala \
  -p 8080:8080 \
  -v ./data/audiobooks:/app/audiobooks \
  -v ./data/config:/app/config \
  -v ./data/cookies:/app/cookies \
  lazybala:latest
```

## 📁 目录结构

```
lazybala/
├── Dockerfile              # Docker 构建文件
├── docker-compose.yml      # Docker Compose 配置
├── .dockerignore           # Docker 忽略文件
├── build.sh               # Linux/macOS 构建脚本
├── build.ps1              # Windows 构建脚本
├── data/                  # 数据目录（挂载点）
│   ├── audiobooks/        # 下载的音频文件
│   ├── config/           # 配置文件
│   └── cookies/          # 认证 cookies
└── ...
```

## ⚙️ 配置说明

### 环境变量

| 变量名 | 默认值 | 说明 |
|--------|--------|------|
| `PORT` | `8080` | 服务端口 |
| `GIN_MODE` | `release` | Gin 运行模式 |
| `TZ` | `Asia/Shanghai` | 时区设置 |

### 数据卷

| 容器路径 | 宿主机路径 | 说明 |
|----------|------------|------|
| `/app/audiobooks` | `./data/audiobooks` | 下载的音频文件存储 |
| `/app/config` | `./data/config` | 应用配置文件 |
| `/app/cookies` | `./data/cookies` | 认证 cookies 存储 |

### 端口映射

| 容器端口 | 宿主机端口 | 说明 |
|----------|------------|------|
| `8080` | `8080` | Web 服务端口 |

## 🔧 高级配置

### 自定义 Docker Compose

```yaml
version: '3.8'

services:
  lazybala:
    build: .
    container_name: lazybala
    restart: unless-stopped
    ports:
      - "8080:8080"
    volumes:
      - ./data/audiobooks:/app/audiobooks
      - ./data/config:/app/config
      - ./data/cookies:/app/cookies
    environment:
      - PORT=8080
      - GIN_MODE=release
      - TZ=Asia/Shanghai
    networks:
      - lazybala-network

networks:
  lazybala-network:
    driver: bridge
```

### 使用外部二进制工具

如果您有预下载的 yt-dlp 和 ffmpeg 二进制文件：

```yaml
volumes:
  - ./bin:/app/bin:ro  # 只读挂载二进制工具
```

### 反向代理配置

#### Nginx
```nginx
server {
    listen 80;
    server_name lazybala.example.com;
    
    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

#### Traefik
```yaml
labels:
  - "traefik.enable=true"
  - "traefik.http.routers.lazybala.rule=Host(`lazybala.example.com`)"
  - "traefik.http.services.lazybala.loadbalancer.server.port=8080"
```

## 🛠️ 维护操作

### 查看日志
```bash
# 实时日志
docker-compose logs -f

# 查看最近100行日志
docker-compose logs --tail=100
```

### 更新应用
```bash
# 停止服务
docker-compose down

# 拉取最新代码
git pull

# 重新构建并启动
docker-compose up -d --build
```

### 备份数据
```bash
# 备份下载文件
tar -czf audiobooks-backup-$(date +%Y%m%d).tar.gz data/audiobooks/

# 备份配置
tar -czf config-backup-$(date +%Y%m%d).tar.gz data/config/ data/cookies/
```

### 清理资源
```bash
# 停止并删除容器
docker-compose down

# 清理未使用的镜像
docker image prune -f

# 清理未使用的卷
docker volume prune -f
```

## 🔍 故障排除

### 常见问题

1. **容器无法启动**
   ```bash
   # 查看容器日志
   docker-compose logs lazybala
   
   # 检查端口占用
   netstat -tlnp | grep 8080
   ```

2. **下载失败**
   ```bash
   # 检查 yt-dlp 是否正常
   docker exec lazybala /app/bin/yt-dlp --version
   
   # 检查网络连接
   docker exec lazybala ping -c 3 www.bilibili.com
   ```

3. **权限问题**
   ```bash
   # 检查目录权限
   ls -la data/
   
   # 修复权限
   sudo chown -R 1001:1001 data/
   ```

### 健康检查

容器内置健康检查，可以通过以下命令查看状态：

```bash
# 查看健康状态
docker inspect lazybala --format='{{.State.Health.Status}}'

# 查看健康检查日志
docker inspect lazybala --format='{{range .State.Health.Log}}{{.Output}}{{end}}'
```

## 📊 监控

### 资源使用情况
```bash
# 查看容器资源使用
docker stats lazybala

# 查看磁盘使用
du -sh data/audiobooks/
```

### 性能优化

1. **限制资源使用**
   ```yaml
   deploy:
     resources:
       limits:
         cpus: '1.0'
         memory: 512M
       reservations:
         cpus: '0.5'
         memory: 256M
   ```

2. **使用多阶段构建**
   - Dockerfile 已经使用多阶段构建来减小镜像大小

3. **缓存优化**
   - 使用 `.dockerignore` 减少构建上下文
   - 合理安排 Dockerfile 层级顺序

## 🔐 安全建议

1. **使用非 root 用户运行**
   - 容器已配置为使用 `lazybala` 用户运行

2. **网络安全**
   - 使用自定义网络隔离容器
   - 配置防火墙规则

3. **数据安全**
   - 定期备份重要数据
   - 使用加密存储敏感信息

## 📞 支持

如果您遇到问题，请：

1. 查看本文档的故障排除部分
2. 检查 GitHub Issues
3. 提交新的 Issue 并附上详细的错误信息和日志

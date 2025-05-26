# 群晖 NAS 部署指南

## 📋 概述

本指南专门针对群晖 NAS 系统的 Docker 部署，解决常见的权限问题和配置优化。

## 🚨 权限问题解决

### 问题描述
在群晖 NAS 上运行 LazyBala 时，可能遇到以下权限错误：
- `open audiobooks: permission denied`
- `chmod config: operation not permitted`
- `chmod audiobooks: operation not permitted`

### 解决方案

#### 方案一：使用群晖专用镜像（推荐）

使用 `Dockerfile.synology` 构建的镜像，该镜像使用 root 用户运行，避免权限问题。

```bash
# 使用群晖专用 compose 文件
docker-compose -f docker-compose.synology.yml up -d
```

#### 方案二：手动设置权限

```bash
# 在群晖 SSH 中执行
sudo mkdir -p /volume1/docker/lazybala/{audiobooks,config,cookies,bin}
sudo chmod -R 777 /volume1/docker/lazybala/
```

## 🚀 快速部署

### 1. 准备工作

#### 启用 SSH（可选）
1. 打开群晖控制面板
2. 进入 "终端机和 SNMP"
3. 启用 SSH 服务

#### 安装 Docker
1. 打开套件中心
2. 搜索并安装 "Docker"

### 2. 部署方法

#### 方法一：使用自动部署脚本

```bash
# SSH 连接到群晖
ssh admin@your-synology-ip

# 克隆项目
git clone https://github.com/kis2show/lazybala.git
cd lazybala

# 运行部署脚本
chmod +x scripts/deploy-synology.sh
./scripts/deploy-synology.sh
```

#### 方法二：手动部署

```bash
# 创建数据目录
sudo mkdir -p /volume1/docker/lazybala/{audiobooks,config,cookies,bin}
sudo chmod -R 777 /volume1/docker/lazybala/

# 部署应用
docker-compose -f docker-compose.synology.yml up -d --build
```

#### 方法三：群晖 Docker GUI

1. 打开 Docker 套件
2. 在 "注册表" 中搜索 `kis2show/lazybala`
3. 下载镜像
4. 创建容器时设置：
   - **端口映射**: `8080:8080`
   - **卷映射**:
     - `/volume1/docker/lazybala/audiobooks` → `/app/audiobooks`
     - `/volume1/docker/lazybala/config` → `/app/config`
     - `/volume1/docker/lazybala/cookies` → `/app/cookies`
     - `/volume1/docker/lazybala/bin` → `/app/bin`
   - **环境变量**:
     - `PORT=8080`
     - `GIN_MODE=release`
     - `TZ=Asia/Shanghai`

## 📁 目录结构

```
/volume1/docker/lazybala/
├── audiobooks/          # 下载的音频文件
├── config/             # 配置文件
├── cookies/            # 认证 cookies
└── bin/                # 二进制工具
    └── yt-dlp          # yt-dlp 可执行文件
```

## ⚙️ 配置优化

### 群晖专用配置

```yaml
# docker-compose.synology.yml
services:
  lazybala:
    user: "0:0"  # 使用 root 用户
    environment:
      - PUID=0
      - PGID=0
    volumes:
      - /volume1/docker/lazybala/audiobooks:/app/audiobooks
      - /volume1/docker/lazybala/config:/app/config
      - /volume1/docker/lazybala/cookies:/app/cookies
      - /volume1/docker/lazybala/bin:/app/bin
```

### 性能优化

```yaml
deploy:
  resources:
    limits:
      cpus: '2.0'      # 限制 CPU 使用
      memory: 1G       # 限制内存使用
    reservations:
      cpus: '0.5'
      memory: 256M
```

## 🔧 常见问题

### 1. 权限被拒绝
```bash
# 解决方案：设置正确权限
sudo chmod -R 777 /volume1/docker/lazybala/
```

### 2. 容器无法启动
```bash
# 查看日志
docker-compose -f docker-compose.synology.yml logs

# 检查端口占用
netstat -tlnp | grep 8080
```

### 3. 下载失败
```bash
# 检查 yt-dlp 是否正常
docker exec lazybala /app/bin/yt-dlp --version

# 手动更新 yt-dlp
docker exec lazybala wget -O /app/bin/yt-dlp https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp_linux
docker exec lazybala chmod +x /app/bin/yt-dlp
```

### 4. 网络问题
```bash
# 检查网络连接
docker exec lazybala ping -c 3 www.bilibili.com

# 检查 DNS
docker exec lazybala nslookup www.bilibili.com
```

## 🔄 升级和维护

### 升级应用
```bash
# 停止容器
docker-compose -f docker-compose.synology.yml down

# 拉取最新代码
git pull

# 重新构建并启动
docker-compose -f docker-compose.synology.yml up -d --build
```

### 升级 yt-dlp
```bash
# 使用升级脚本
./scripts/update-ytdlp.sh

# 或手动升级
docker exec lazybala wget -O /app/bin/yt-dlp https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp_linux
docker exec lazybala chmod +x /app/bin/yt-dlp
```

### 备份数据
```bash
# 备份下载文件
tar -czf audiobooks-backup-$(date +%Y%m%d).tar.gz /volume1/docker/lazybala/audiobooks/

# 备份配置
tar -czf config-backup-$(date +%Y%m%d).tar.gz /volume1/docker/lazybala/config/ /volume1/docker/lazybala/cookies/
```

## 📊 监控

### 查看资源使用
```bash
# 查看容器状态
docker stats lazybala

# 查看磁盘使用
du -sh /volume1/docker/lazybala/audiobooks/
```

### 日志管理
```bash
# 查看实时日志
docker-compose -f docker-compose.synology.yml logs -f

# 查看最近日志
docker-compose -f docker-compose.synology.yml logs --tail=100
```

## 🔐 安全建议

1. **网络安全**
   - 配置群晖防火墙规则
   - 使用 HTTPS 反向代理

2. **访问控制**
   - 限制访问 IP 范围
   - 配置强密码策略

3. **数据安全**
   - 定期备份重要数据
   - 启用群晖快照功能

## 📞 支持

如果遇到问题：
1. 查看本文档的常见问题部分
2. 检查容器日志
3. 在 GitHub Issues 中报告问题并附上群晖系统信息

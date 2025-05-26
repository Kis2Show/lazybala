# yt-dlp 升级指南

## 📋 概述

LazyBala 现在支持将 yt-dlp 二进制文件作为可挂载的卷，允许用户手动升级 yt-dlp 而无需重新构建 Docker 镜像。

## 🏗️ 目录结构

```
lazybala/
├── data/
│   ├── audiobooks/     # 下载的音频文件
│   ├── config/         # 配置文件
│   ├── cookies/        # 认证 cookies
│   └── bin/            # 二进制工具目录 (新增)
│       └── yt-dlp      # yt-dlp 可执行文件
├── scripts/
│   ├── update-ytdlp.sh    # Linux/macOS 升级脚本
│   └── update-ytdlp.ps1   # Windows 升级脚本
└── docker-compose.yml
```

## 🚀 使用方法

### 方法一：使用自动升级脚本（推荐）

#### Linux/macOS
```bash
# 给脚本执行权限
chmod +x scripts/update-ytdlp.sh

# 检查并升级到最新版本
./scripts/update-ytdlp.sh

# 强制重新下载
./scripts/update-ytdlp.sh --force

# 恢复备份版本
./scripts/update-ytdlp.sh --restore

# 查看当前版本
./scripts/update-ytdlp.sh --version
```

#### Windows (PowerShell)
```powershell
# 检查并升级到最新版本
.\scripts\update-ytdlp.ps1

# 强制重新下载
.\scripts\update-ytdlp.ps1 -Force

# 恢复备份版本
.\scripts\update-ytdlp.ps1 -Restore

# 查看当前版本
.\scripts\update-ytdlp.ps1 -Version
```

### 方法二：手动升级

#### 1. 停止容器
```bash
docker-compose down
```

#### 2. 下载最新的 yt-dlp

**Linux AMD64:**
```bash
wget -O data/bin/yt-dlp https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp_linux
chmod +x data/bin/yt-dlp
```

**Linux ARM64:**
```bash
wget -O data/bin/yt-dlp https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp_linux_aarch64
chmod +x data/bin/yt-dlp
```

**Windows:**
```powershell
Invoke-WebRequest -Uri "https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp.exe" -OutFile "data\bin\yt-dlp.exe"
```

#### 3. 重启容器
```bash
docker-compose up -d
```

## 🔧 Docker 配置

### docker-compose.yml 配置

```yaml
services:
  lazybala:
    volumes:
      - ./data/audiobooks:/app/audiobooks
      - ./data/config:/app/config
      - ./data/cookies:/app/cookies
      - ./data/bin:/app/bin  # 二进制工具目录
```

### Dockerfile 配置

```dockerfile
# 数据卷
VOLUME ["/app/audiobooks", "/app/config", "/app/cookies", "/app/bin"]
```

## 📦 版本管理

### 查看当前版本
```bash
# 在容器内
docker exec lazybala /app/bin/yt-dlp --version

# 在宿主机（如果已挂载）
./data/bin/yt-dlp --version
```

### 备份和恢复

升级脚本会自动创建备份：
- Linux: `data/bin/yt-dlp.backup`
- Windows: `data/bin/yt-dlp.backup.exe`

如果新版本有问题，可以使用恢复功能：
```bash
# Linux/macOS
./scripts/update-ytdlp.sh --restore

# Windows
.\scripts\update-ytdlp.ps1 -Restore
```

## 🎯 架构支持

脚本会自动检测系统架构并下载对应版本：

| 架构 | Linux 文件名 | Windows 文件名 |
|------|-------------|---------------|
| AMD64 (x86_64) | `yt-dlp_linux` | `yt-dlp.exe` |
| ARM64 (aarch64) | `yt-dlp_linux_aarch64` | `yt-dlp_arm64.exe` |
| ARM (armv7l) | `yt-dlp_linux_armv7l` | - |

## 🔍 故障排除

### 权限问题

#### 容器内权限问题
```bash
# 检查容器内 yt-dlp 权限
docker exec lazybala ls -la /app/bin/yt-dlp

# 手动修复容器内权限
docker exec lazybala chmod +x /app/bin/yt-dlp

# 重启容器（会自动修复权限）
docker-compose restart lazybala
```

#### 宿主机权限问题
```bash
# 确保 bin 目录有正确权限
chmod 755 data/bin/
chmod +x data/bin/yt-dlp

# 群晖系统权限修复
sudo chown -R 100:100 /volume1/docker/lazybala/bin/
sudo chmod +x /volume1/docker/lazybala/bin/yt-dlp
```

### 下载失败
1. 检查网络连接
2. 确认 GitHub 可访问
3. 尝试手动下载

### 容器内无法执行
```bash
# 检查文件是否存在
docker exec lazybala ls -la /app/bin/

# 检查权限
docker exec lazybala ls -la /app/bin/yt-dlp

# 测试执行
docker exec lazybala /app/bin/yt-dlp --version
```

## 📝 注意事项

1. **首次运行**：如果 `data/bin/` 目录不存在，Docker 会自动创建并从镜像复制 yt-dlp
2. **权限**：确保 yt-dlp 文件有执行权限
3. **备份**：升级前会自动备份当前版本
4. **验证**：升级后会自动验证新版本是否正常工作
5. **回滚**：如果新版本有问题，可以快速恢复到备份版本

## 🚀 自动化升级

可以将升级脚本加入到 cron 任务中实现自动升级：

```bash
# 每周检查一次更新（周日凌晨2点）
0 2 * * 0 /path/to/lazybala/scripts/update-ytdlp.sh >/dev/null 2>&1
```

## 📞 支持

如果遇到问题：
1. 查看脚本输出的错误信息
2. 检查 Docker 容器日志：`docker-compose logs lazybala`
3. 在 GitHub Issues 中报告问题

# LazyBala 发布指南

## 📋 概述

本文档说明如何使用 GitHub Actions 自动发布 LazyBala 的新版本，包括二进制文件和 Docker 镜像的自动构建和发布。

## 🚀 发布流程

### 1. 准备发布

在发布新版本之前，请确保：

- [ ] 所有功能已完成并测试
- [ ] 代码已合并到 `main` 分支
- [ ] 更新了版本相关文档
- [ ] 运行了本地测试

### 2. 创建发布标签

```bash
# 1. 确保在 main 分支
git checkout main
git pull origin main

# 2. 创建并推送标签
git tag v1.0.0
git push origin v1.0.0
```

### 3. 自动化流程

推送标签后，GitHub Actions 会自动：

1. **创建 Release** - 生成 Release 页面和变更日志
2. **构建二进制文件** - 8个平台的静态可执行文件
3. **构建 Docker 镜像** - 多平台容器镜像
4. **上传资产** - 将所有文件上传到 Release

## 📦 构建产物

### 二进制文件

每次发布会自动构建以下平台的二进制文件：

| 平台 | 架构 | 文件名 | 大小 (约) |
|------|------|--------|-----------|
| 🐧 Linux | x64 | `lazybala-linux-amd64.tar.gz` | ~15MB |
| 🐧 Linux | ARM64 | `lazybala-linux-arm64.tar.gz` | ~14MB |
| 🐧 Linux | ARMv7 | `lazybala-linux-armv7.tar.gz` | ~13MB |
| 🪟 Windows | x64 | `lazybala-windows-amd64.exe.zip` | ~16MB |
| 🪟 Windows | ARM64 | `lazybala-windows-arm64.exe.zip` | ~15MB |
| 🍎 macOS | x64 | `lazybala-darwin-amd64.tar.gz` | ~15MB |
| 🍎 macOS | Apple Silicon | `lazybala-darwin-arm64.tar.gz` | ~14MB |
| 🔧 FreeBSD | x64 | `lazybala-freebsd-amd64.tar.gz` | ~15MB |

### Docker 镜像

同时会构建多平台 Docker 镜像：

- `ghcr.io/kis2show/lazybala:v1.0.0`
- `ghcr.io/kis2show/lazybala:1.0`
- `ghcr.io/kis2show/lazybala:1`
- `ghcr.io/kis2show/lazybala:latest`

## 🔐 安全特性

### 文件校验

每个二进制文件都提供 SHA256 校验和：

```bash
# 下载文件和校验和
wget https://github.com/kis2show/lazybala/releases/download/v1.0.0/lazybala-linux-amd64.tar.gz
wget https://github.com/kis2show/lazybala/releases/download/v1.0.0/lazybala-linux-amd64.tar.gz.sha256

# 验证文件完整性
sha256sum -c lazybala-linux-amd64.tar.gz.sha256
```

### 构建证明

所有构建都包含 GitHub 的构建证明，可以验证构建的真实性和完整性。

## 📥 安装方法

### 自动安装脚本

**Linux/macOS:**
```bash
curl -fsSL https://raw.githubusercontent.com/kis2show/lazybala/main/install.sh | bash
```

**Windows (PowerShell):**
```powershell
iwr -useb https://raw.githubusercontent.com/kis2show/lazybala/main/install.ps1 | iex
```

### 手动安装

1. 从 [Releases 页面](https://github.com/kis2show/lazybala/releases) 下载对应平台的文件
2. 解压文件
3. 将可执行文件移动到 PATH 目录

**Linux/macOS 示例:**
```bash
# 下载并解压
tar -xzf lazybala-linux-amd64.tar.gz

# 移动到系统目录
sudo mv lazybala-linux-amd64 /usr/local/bin/lazybala

# 设置执行权限
sudo chmod +x /usr/local/bin/lazybala
```

**Windows 示例:**
```powershell
# 解压文件
Expand-Archive lazybala-windows-amd64.exe.zip

# 移动到程序目录
Move-Item lazybala-windows-amd64.exe $env:LOCALAPPDATA\Programs\LazyBala\lazybala.exe
```

### Docker 安装

```bash
# 拉取镜像
docker pull ghcr.io/kis2show/lazybala:latest

# 运行容器
docker run -d \
  --name lazybala \
  -p 8080:8080 \
  -v ./data/audiobooks:/app/audiobooks \
  -v ./data/config:/app/config \
  -v ./data/cookies:/app/cookies \
  ghcr.io/kis2show/lazybala:latest
```

## 🔄 版本管理

### 版本号规范

使用语义化版本控制 (SemVer)：

- `v1.0.0` - 主版本.次版本.修订版本
- `v1.0.0-rc.1` - 候选版本
- `v1.0.0-beta.1` - 测试版本

### 分支策略

- `main` - 稳定版本，用于发布
- `develop` - 开发版本，用于集成
- `feature/*` - 功能分支
- `hotfix/*` - 紧急修复

### 发布类型

| 类型 | 触发条件 | 示例标签 | 说明 |
|------|----------|----------|------|
| 正式版本 | `v*.*.*` | `v1.0.0` | 稳定发布版本 |
| 候选版本 | `v*.*.*-rc.*` | `v1.0.0-rc.1` | 发布候选版本 |
| 测试版本 | `v*.*.*-beta.*` | `v1.0.0-beta.1` | 测试版本 |
| 预发布版本 | `v*.*.*-alpha.*` | `v1.0.0-alpha.1` | 早期预览版本 |

## 🛠️ 故障排除

### 常见问题

1. **构建失败**
   - 检查 Go 代码是否有语法错误
   - 确认所有依赖都可用
   - 查看 Actions 日志获取详细错误信息

2. **上传失败**
   - 检查 GitHub Token 权限
   - 确认 Release 已正确创建
   - 验证文件路径和名称

3. **Docker 推送失败**
   - 检查容器注册表权限
   - 验证镜像标签格式
   - 确认网络连接正常

### 调试技巧

```bash
# 本地测试构建
go build -ldflags="-w -s -X main.version=test" -o lazybala-test .

# 本地测试 Docker 构建
docker build -t lazybala:test .

# 检查二进制文件信息
file lazybala-test
ldd lazybala-test  # Linux 上检查依赖
```

## 📊 发布统计

### 构建时间

- 二进制构建：约 5-8 分钟
- Docker 构建：约 10-15 分钟
- 总发布时间：约 15-25 分钟

### 文件大小

- 单个二进制文件：13-16MB
- Docker 镜像：约 50-80MB
- 总 Release 大小：约 150-200MB

## 📞 支持

如果在发布过程中遇到问题：

1. 查看 [GitHub Actions 日志](https://github.com/kis2show/lazybala/actions)
2. 检查 [Issues 页面](https://github.com/kis2show/lazybala/issues)
3. 提交新的 Issue 并附上详细信息

## 📚 相关文档

- [GitHub Actions 配置](GITHUB_ACTIONS.md)
- [Docker 部署指南](DOCKER.md)
- [贡献指南](CONTRIBUTING.md)

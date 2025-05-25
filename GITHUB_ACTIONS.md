# GitHub Actions CI/CD 配置指南

## 📋 概述

LazyBala 项目配置了完整的 GitHub Actions CI/CD 流水线，支持自动化构建、测试、安全扫描和部署。

## 🚀 工作流概览

### 1. CI/CD 主流水线 (`ci-cd.yml`)

**触发条件：**
- 推送到 `main` 或 `develop` 分支
- 创建 Pull Request 到 `main` 分支
- 手动触发

**流程步骤：**
1. **测试阶段** - Go 代码测试、格式检查
2. **构建阶段** - 多平台 Docker 镜像构建
3. **安全扫描** - Trivy 漏洞扫描
4. **部署阶段** - 自动部署到测试/生产环境
5. **通知阶段** - 构建结果通知

### 2. 发布流水线 (`release.yml`)

**触发条件：**
- 推送 `v*` 标签
- 手动触发

**流程步骤：**
1. **创建发布** - 自动生成 Release Notes
2. **构建二进制** - 8个平台的静态可执行文件
3. **构建镜像** - 发布 Docker 镜像到多个仓库

**支持的平台：**
- Linux (x64, ARM64, ARMv7)
- Windows (x64, ARM64)
- macOS (x64, Apple Silicon)
- FreeBSD (x64)

### 3. Docker 构建 (`docker-build.yml`)

**触发条件：**
- 推送到主要分支
- 创建标签
- Pull Request

**功能特性：**
- 多平台构建 (linux/amd64, linux/arm64)
- 镜像缓存优化
- 安全签名验证

### 4. 依赖更新 (`dependency-update.yml`)

**触发条件：**
- 每周一定时执行
- 手动触发

**功能特性：**
- 自动更新 Go 依赖
- 自动更新 Docker 基础镜像
- 安全漏洞扫描

## ⚙️ 配置要求

### 必需的 Secrets

在 GitHub 仓库设置中配置以下 Secrets：

| Secret 名称 | 描述 | 必需 |
|-------------|------|------|
| `GITHUB_TOKEN` | GitHub 自动生成 | ✅ |
| `DOCKERHUB_USERNAME` | Docker Hub 用户名 | ❌ |
| `DOCKERHUB_TOKEN` | Docker Hub 访问令牌 | ❌ |

### 可选的 Secrets

| Secret 名称 | 描述 | 用途 |
|-------------|------|------|
| `SLACK_WEBHOOK` | Slack 通知 Webhook | 构建通知 |
| `DEPLOY_KEY` | 部署密钥 | 自动部署 |

## 📦 二进制文件发布

### 自动构建的平台

每次发布都会自动构建以下平台的静态可执行文件：

| 平台 | 架构 | 文件名 | 压缩格式 |
|------|------|--------|----------|
| Linux | x64 | `lazybala-linux-amd64` | tar.gz |
| Linux | ARM64 | `lazybala-linux-arm64` | tar.gz |
| Linux | ARMv7 | `lazybala-linux-armv7` | tar.gz |
| Windows | x64 | `lazybala-windows-amd64.exe` | zip |
| Windows | ARM64 | `lazybala-windows-arm64.exe` | zip |
| macOS | x64 | `lazybala-darwin-amd64` | tar.gz |
| macOS | Apple Silicon | `lazybala-darwin-arm64` | tar.gz |
| FreeBSD | x64 | `lazybala-freebsd-amd64` | tar.gz |

### 安装脚本

提供了自动安装脚本：

**Linux/macOS:**
```bash
curl -fsSL https://raw.githubusercontent.com/kis2show/lazybala/main/install.sh | bash
```

**Windows (PowerShell):**
```powershell
iwr -useb https://raw.githubusercontent.com/kis2show/lazybala/main/install.ps1 | iex
```

### 手动下载

从 [Releases 页面](https://github.com/kis2show/lazybala/releases) 下载对应平台的文件。

## 🐳 Docker 镜像发布

### 支持的镜像仓库

1. **GitHub Container Registry (默认)**
   ```bash
   ghcr.io/kis2show/lazybala:latest
   ```

2. **Docker Hub (可选)**
   ```bash
   kis2show/lazybala:latest
   ```

### 镜像标签策略

| 触发条件 | 生成的标签 |
|----------|------------|
| `main` 分支推送 | `latest` |
| `develop` 分支推送 | `develop` |
| `v1.2.3` 标签 | `v1.2.3`, `1.2`, `1`, `latest` |
| Pull Request | `pr-123` |
| 提交 SHA | `main-abc1234` |

## 🔧 本地测试

### 测试 Docker 构建

```bash
# 使用构建脚本
./build.sh latest

# 或手动构建
docker build -t lazybala:test .
```

### 测试 GitHub Actions (使用 act)

```bash
# 安装 act
curl https://raw.githubusercontent.com/nektos/act/master/install.sh | sudo bash

# 测试 CI 流水线
act push

# 测试发布流水线
act push -e .github/workflows/test-events/release.json
```

## 📊 构建状态徽章

在 README.md 中添加构建状态徽章：

```markdown
![CI/CD](https://github.com/kis2show/lazybala/workflows/CI/CD%20Pipeline/badge.svg)
![Docker](https://github.com/kis2show/lazybala/workflows/Build%20and%20Push%20Docker%20Image/badge.svg)
![Release](https://github.com/kis2show/lazybala/workflows/Release/badge.svg)
```

## 🚀 部署配置

### 自动部署到 Kubernetes

在 `ci-cd.yml` 中配置部署步骤：

```yaml
- name: Deploy to Kubernetes
  run: |
    kubectl set image deployment/lazybala \
      lazybala=${{ needs.build.outputs.image }} \
      --namespace=production
```

### 自动部署到 Docker Compose

```yaml
- name: Deploy with Docker Compose
  run: |
    ssh user@server "cd /path/to/lazybala && \
      docker-compose pull && \
      docker-compose up -d"
```

## 🔍 监控和日志

### 构建监控

- **GitHub Actions 页面** - 查看构建历史和日志
- **Dependabot** - 自动依赖更新提醒
- **Security 页面** - 安全漏洞报告

### 镜像监控

- **GitHub Packages** - 镜像存储和版本管理
- **Docker Hub** - 镜像下载统计
- **Trivy 扫描** - 安全漏洞报告

## 🛠️ 自定义配置

### 修改构建平台

在 `docker-build.yml` 中修改：

```yaml
platforms: linux/amd64,linux/arm64,linux/arm/v7
```

### 添加新的部署环境

```yaml
deploy-staging:
  name: Deploy to Staging
  needs: [build, security-scan]
  runs-on: ubuntu-latest
  if: github.ref == 'refs/heads/develop'
  environment:
    name: staging
    url: https://staging.lazybala.example.com
```

### 自定义通知

```yaml
- name: Notify Slack
  if: always()
  uses: 8398a7/action-slack@v3
  with:
    status: ${{ job.status }}
    webhook_url: ${{ secrets.SLACK_WEBHOOK }}
```

## 📋 最佳实践

### 1. 分支策略

- `main` - 生产环境代码
- `develop` - 开发环境代码
- `feature/*` - 功能分支
- `hotfix/*` - 紧急修复分支

### 2. 标签规范

- `v1.2.3` - 正式版本
- `v1.2.3-rc.1` - 候选版本
- `v1.2.3-beta.1` - 测试版本

### 3. 提交信息规范

```
feat: 添加新功能
fix: 修复bug
docs: 更新文档
style: 代码格式调整
refactor: 代码重构
test: 添加测试
chore: 构建过程或辅助工具的变动
```

### 4. 安全考虑

- 定期更新依赖
- 启用 Dependabot
- 配置安全扫描
- 使用最小权限原则

## 🔧 故障排除

### 常见问题

1. **构建失败**
   - 检查 Go 版本兼容性
   - 验证依赖是否可用
   - 查看详细错误日志

2. **Docker 推送失败**
   - 检查仓库权限
   - 验证认证信息
   - 确认镜像标签格式

3. **部署失败**
   - 检查部署环境配置
   - 验证网络连接
   - 确认服务状态

### 调试技巧

```yaml
- name: Debug information
  run: |
    echo "Event: ${{ github.event_name }}"
    echo "Ref: ${{ github.ref }}"
    echo "SHA: ${{ github.sha }}"
    env
```

## 📞 支持

如果您在配置 GitHub Actions 时遇到问题：

1. 查看 [GitHub Actions 文档](https://docs.github.com/en/actions)
2. 检查工作流日志中的错误信息
3. 在项目 Issues 中搜索相关问题
4. 提交新的 Issue 并附上详细信息

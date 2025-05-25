# LazyBala 项目设置指南

## 📋 概述

本文档说明如何将 LazyBala 项目发布到 GitHub 仓库 `kis2show/lazybala`。

## 🚀 发布步骤

### 1. 创建 GitHub 仓库

首先需要在 GitHub 上创建仓库：

1. 访问 [GitHub](https://github.com)
2. 点击右上角的 "+" 按钮，选择 "New repository"
3. 填写仓库信息：
   - **Repository name**: `lazybala`
   - **Description**: `LazyBala - 基于 Go 和纯 JavaScript 的媒体下载应用，支持 Bilibili 等平台的音频下载`
   - **Visibility**: Public (推荐) 或 Private
   - **Initialize this repository with**: 不要勾选任何选项（因为我们已有代码）
4. 点击 "Create repository"

### 2. 推送代码到远程仓库

创建仓库后，在本地执行以下命令：

```bash
# 确保在项目目录中
cd /path/to/lazybala

# 检查当前状态
git status

# 如果还没有提交，先提交代码
git add .
git commit -m "feat: 初始化 LazyBala 项目 - 支持 Bilibili 音频下载的多平台应用"

# 添加远程仓库（如果还没有添加）
git remote add origin https://github.com/kis2show/lazybala.git

# 推送代码到远程仓库
git push -u origin master

# 或者如果使用 main 分支
git branch -M main
git push -u origin main
```

### 3. 配置 GitHub Actions

推送代码后，GitHub Actions 会自动激活。需要配置以下设置：

#### 3.1 启用 Actions

1. 进入仓库页面
2. 点击 "Actions" 标签页
3. 如果看到提示，点击 "I understand my workflows, go ahead and enable them"

#### 3.2 配置 Secrets（可选）

如果要发布到 Docker Hub，需要配置以下 Secrets：

1. 进入仓库的 "Settings" > "Secrets and variables" > "Actions"
2. 点击 "New repository secret"
3. 添加以下 Secrets：
   - `DOCKERHUB_USERNAME`: Docker Hub 用户名
   - `DOCKERHUB_TOKEN`: Docker Hub 访问令牌

#### 3.3 配置权限

确保 GitHub Actions 有足够的权限：

1. 进入仓库的 "Settings" > "Actions" > "General"
2. 在 "Workflow permissions" 部分：
   - 选择 "Read and write permissions"
   - 勾选 "Allow GitHub Actions to create and approve pull requests"
3. 点击 "Save"

### 4. 创建第一个发布

推送代码后，可以创建第一个发布版本：

```bash
# 创建并推送标签
git tag v1.0.0
git push origin v1.0.0
```

这会触发 GitHub Actions 自动：
- 构建 8 个平台的二进制文件
- 构建多平台 Docker 镜像
- 创建 GitHub Release
- 上传所有构建产物

### 5. 验证部署

发布完成后，检查以下内容：

#### 5.1 GitHub Release
- 访问 `https://github.com/kis2show/lazybala/releases`
- 确认 Release 页面包含所有二进制文件
- 验证下载链接正常工作

#### 5.2 Docker 镜像
- 检查 GitHub Container Registry: `ghcr.io/kis2show/lazybala`
- 如果配置了 Docker Hub，检查: `kis2show/lazybala`

#### 5.3 GitHub Actions
- 访问 `https://github.com/kis2show/lazybala/actions`
- 确认所有工作流运行成功

## 🔧 故障排除

### 常见问题

1. **推送失败 - Repository not found**
   - 确认 GitHub 仓库已创建
   - 检查仓库名称是否正确
   - 确认有推送权限

2. **GitHub Actions 失败**
   - 检查工作流文件语法
   - 确认权限配置正确
   - 查看 Actions 日志获取详细错误

3. **Docker 推送失败**
   - 检查 DOCKERHUB_USERNAME 和 DOCKERHUB_TOKEN
   - 确认 Docker Hub 仓库存在
   - 验证令牌权限

### 调试命令

```bash
# 检查远程仓库配置
git remote -v

# 检查当前分支
git branch

# 检查提交历史
git log --oneline

# 强制推送（谨慎使用）
git push -f origin master
```

## 📚 相关文档

创建仓库后，用户可以参考以下文档：

- [README.md](README.md) - 项目介绍和使用说明
- [DOCKER.md](DOCKER.md) - Docker 部署指南
- [GITHUB_ACTIONS.md](GITHUB_ACTIONS.md) - GitHub Actions 配置说明
- [RELEASE.md](RELEASE.md) - 发布流程指南

## 🎉 完成

完成以上步骤后，LazyBala 项目就成功发布到 `kis2show/lazybala` 仓库了！

用户可以通过以下方式使用：

### 快速安装
```bash
# Linux/macOS
curl -fsSL https://raw.githubusercontent.com/kis2show/lazybala/main/install.sh | bash

# Windows (PowerShell)
iwr -useb https://raw.githubusercontent.com/kis2show/lazybala/main/install.ps1 | iex
```

### Docker 使用
```bash
docker run -d --name lazybala -p 8080:8080 ghcr.io/kis2show/lazybala:latest
```

### 手动下载
从 [Releases 页面](https://github.com/kis2show/lazybala/releases) 下载对应平台的二进制文件。

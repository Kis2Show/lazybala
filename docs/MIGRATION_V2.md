# LazyBala v2.0 迁移指南

## 🚨 重要变更

LazyBala v2.0 修复了一个关键的 Docker 卷挂载问题，该问题导致 yt-dlp 文件在容器启动时被外部空目录覆盖。

### 主要变更

1. **移除 bin 目录卷挂载** - yt-dlp 现在保持在容器内
2. **改进的启动检查** - 容器启动时会验证 yt-dlp 可用性
3. **容器内更新机制** - 通过应用内置功能更新 yt-dlp

## 🔧 迁移步骤

### 1. 停止现有容器

```bash
docker-compose down
```

### 2. 备份数据（可选）

```bash
# 备份下载文件
cp -r ./data/audiobooks ./backup-audiobooks

# 备份配置
cp -r ./data/config ./backup-config
cp -r ./data/cookies ./backup-cookies
```

### 3. 更新配置文件

拉取最新代码或手动更新 `docker-compose.yml`：

```bash
git pull
```

或手动移除 bin 目录挂载：

```yaml
volumes:
  - ./data/audiobooks:/app/audiobooks
  - ./data/config:/app/config
  - ./data/cookies:/app/cookies
  # 移除这行: - ./data/bin:/app/bin
```

### 4. 重新构建和启动

```bash
# 重新构建镜像
docker-compose build --no-cache

# 启动服务
docker-compose up -d
```

### 5. 验证部署

```bash
# 检查容器状态
docker-compose ps

# 查看启动日志
docker-compose logs -f

# 验证 yt-dlp
docker exec lazybala /app/bin/yt-dlp --version
```

## 🔍 故障排除

### 问题：容器启动失败，提示 "yt-dlp 文件不存在"

**原因**: 仍然挂载了 bin 目录，覆盖了容器内的 yt-dlp

**解决方案**:
1. 检查 `docker-compose.yml` 确保没有 `./data/bin:/app/bin` 挂载
2. 重新构建镜像: `docker-compose build --no-cache`

### 问题：如何更新 yt-dlp？

**解决方案**:
1. 通过 Web 界面的设置页面更新（推荐）
2. 或使用 API: `curl -X POST http://localhost:8080/api/update-ytdlp`

### 问题：需要手动升级 yt-dlp

**解决方案**:
```bash
# 进入容器
docker exec -it lazybala sh

# 手动下载最新版本
wget -O /app/bin/yt-dlp https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp_linux
chmod +x /app/bin/yt-dlp

# 验证
/app/bin/yt-dlp --version
```

## 📋 检查清单

- [ ] 停止旧容器
- [ ] 备份重要数据
- [ ] 更新 docker-compose.yml（移除 bin 目录挂载）
- [ ] 重新构建镜像
- [ ] 启动新容器
- [ ] 验证 yt-dlp 可用性
- [ ] 测试下载功能

## 🎯 优势

v2.0 的变更带来以下优势：

1. **解决权限问题** - 不再有外部卷覆盖容器内文件的问题
2. **简化部署** - 减少了卷挂载配置的复杂性
3. **更可靠的启动** - 容器启动时会验证 yt-dlp 可用性
4. **内置更新** - 通过应用界面直接更新 yt-dlp

## 📞 支持

如果在迁移过程中遇到问题，请：

1. 查看容器日志: `docker-compose logs -f`
2. 检查 GitHub Issues
3. 提交新的 Issue 并附上详细的错误信息

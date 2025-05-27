# Docker 健康检查修复指南

## 🚨 问题描述

用户报告 Docker 容器健康检查失败，显示错误：
```
http://localhost:8080/: Remote file does not exist -- broken link!!
```

虽然应用功能正常，下载工作正常，但 Docker 健康检查持续失败。

## 🔍 问题分析

### 根本原因
1. **健康检查端点不合适** - 原来使用根路径 `/` 进行健康检查
2. **启动时间不足** - 应用需要更长时间完全启动
3. **检查方式不当** - 使用 `wget --spider` 可能不适合检查动态内容

### 技术细节
- 应用使用纯 JavaScript + HTML，不依赖 Vue.js 构建
- 根路径 `/` 返回完整的 HTML 页面，可能包含动态内容
- 健康检查应该使用简单的 API 端点

## 🔧 修复方案

### 1. 添加专用健康检查端点

在 `main.go` 中添加了新的 `/health` 端点：

```go
// 健康检查端点 - 简单的 JSON 响应
r.GET("/health", func(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{
        "status":  "ok",
        "service": "lazybala",
    })
})
```

### 2. 更新 Dockerfile 健康检查

```dockerfile
# 健康检查 - 使用专用的健康检查端点
HEALTHCHECK --interval=30s --timeout=10s --start-period=60s --retries=3 \
  CMD curl -f http://localhost:8080/health || wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1
```

**关键改进**：
- 使用 `/health` 端点替代根路径 `/`
- 增加启动等待时间：`start-period=60s`（从 30s 增加到 60s）
- 保留 curl 和 wget 双重检查机制

### 3. 更新 docker-compose.yml

```yaml
healthcheck:
  test: ["CMD", "sh", "-c", "curl -f http://localhost:8080/health || wget --no-verbose --tries=1 --spider http://localhost:8080/health"]
  interval: 30s
  timeout: 10s
  retries: 3
  start_period: 60s
```

## 📋 验证步骤

### 1. 手动测试健康检查端点

```bash
# 测试新的健康检查端点
curl http://localhost:8080/health

# 预期响应
{"service":"lazybala","status":"ok"}
```

### 2. 检查 Docker 健康状态

```bash
# 查看容器健康状态
docker ps

# 查看详细健康检查信息
docker inspect --format='{{.State.Health}}' lazybala

# 查看健康检查日志
docker inspect --format='{{range .State.Health.Log}}{{.Output}}{{end}}' lazybala
```

### 3. 使用测试脚本

```bash
# 运行完整测试
./scripts/test-docker-fix.sh

# 或手动测试
docker-compose down
docker-compose build --no-cache
docker-compose up -d
```

## 🎯 预期结果

修复后，用户应该看到：

1. **健康检查通过** - Docker 容器状态显示为 `healthy`
2. **应用正常工作** - 所有功能包括下载继续正常工作
3. **无错误日志** - 不再有 "broken link" 错误信息

## 🔍 故障排除

### 如果健康检查仍然失败

1. **检查应用启动日志**：
   ```bash
   docker-compose logs -f lazybala
   ```

2. **手动测试端点**：
   ```bash
   # 进入容器测试
   docker exec -it lazybala sh
   curl http://localhost:8080/health
   wget -O- http://localhost:8080/health
   ```

3. **检查端口绑定**：
   ```bash
   # 确认端口正确映射
   docker port lazybala
   netstat -tlnp | grep 8080
   ```

4. **增加启动等待时间**：
   ```yaml
   healthcheck:
     start_period: 120s  # 增加到 2 分钟
   ```

### 常见问题

**Q: 为什么不直接使用根路径 `/` 进行健康检查？**
A: 根路径返回完整的 HTML 页面，可能包含动态内容和外部资源引用，不适合简单的健康检查。

**Q: 健康检查失败会影响应用功能吗？**
A: 不会。健康检查只是 Docker 的监控机制，失败不会影响应用本身的功能。

**Q: 可以禁用健康检查吗？**
A: 可以，但不推荐。健康检查有助于监控和自动恢复。

## 📞 支持

如果问题仍然存在：

1. 查看 [GitHub Issues](https://github.com/kis2show/lazybala/issues)
2. 提交新的 Issue 并附上：
   - Docker 版本信息
   - 容器日志
   - 健康检查详细信息
   - 系统环境信息

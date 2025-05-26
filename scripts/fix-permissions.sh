#!/bin/sh
# 容器启动时修复权限脚本

set -e

echo "正在检查和修复文件权限..."

# 修复 yt-dlp 权限
if [ -f "/app/bin/yt-dlp" ]; then
    echo "修复 yt-dlp 执行权限..."
    chmod +x /app/bin/yt-dlp
    echo "yt-dlp 权限已修复"
else
    echo "警告: yt-dlp 文件不存在于 /app/bin/yt-dlp"
fi

# 修复目录权限
echo "修复目录权限..."
chmod -R 755 /app/audiobooks /app/config /app/cookies /app/bin 2>/dev/null || true

echo "权限修复完成"

# 启动应用
echo "启动 LazyBala..."
exec "$@"

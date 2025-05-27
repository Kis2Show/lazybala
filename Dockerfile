# LazyBala 多阶段构建 Dockerfile
# Go 构建阶段
FROM golang:1.23-alpine AS backend-builder

# 安装构建依赖
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# 复制 Go 模块文件
COPY go.mod go.sum ./

# 设置 Go 代理和环境变量
ENV GOPROXY=https://proxy.golang.org,direct
ENV GOSUMDB=sum.golang.org
ENV GO111MODULE=on

# 下载依赖
RUN go mod download

# 复制源代码
COPY *.go ./
COPY *.html ./

# 构建参数
ARG BUILDTIME
ARG VERSION
ARG TARGETARCH
ARG TARGETOS

# 构建应用（支持多架构）
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build \
    -ldflags="-w -s -extldflags '-static' -X main.version=${VERSION} -X main.buildTime=${BUILDTIME}" \
    -a -installsuffix cgo \
    -o lazybala .

# 最终运行镜像
FROM alpine:latest

# 安装必要的依赖
RUN apk --no-cache add \
    ca-certificates \
    tzdata \
    wget \
    curl \
    python3 \
    py3-pip \
    ffmpeg

# 设置时区
ENV TZ=Asia/Shanghai

WORKDIR /app

# 创建必要的目录
RUN mkdir -p /app/bin /app/config /app/cookies /app/audiobooks

# 复制构建的二进制文件
COPY --from=backend-builder /app/lazybala .

# 复制 HTML 文件
COPY --from=backend-builder /app/*.html ./

# 创建启动脚本
RUN echo '#!/bin/sh\n\
echo "======================================"\n\
echo "   LazyBala - 费劲巴拉下载器"\n\
echo "======================================"\n\
echo "正在启动 LazyBala..."\n\
\n\
# 检查 yt-dlp\n\
echo "检查 yt-dlp 状态..."\n\
if [ -f "/app/bin/yt-dlp" ]; then\n\
    chmod +x /app/bin/yt-dlp 2>/dev/null || true\n\
    if /app/bin/yt-dlp --version >/dev/null 2>&1; then\n\
        YT_VERSION=$(/app/bin/yt-dlp --version 2>/dev/null || echo "unknown")\n\
        echo "✓ yt-dlp 可用: $YT_VERSION"\n\
    else\n\
        echo "✗ yt-dlp 文件存在但无法执行"\n\
        exit 1\n\
    fi\n\
else\n\
    echo "✗ yt-dlp 文件不存在: /app/bin/yt-dlp"\n\
    echo "这可能是由于 Docker 卷挂载覆盖了容器内的文件"\n\
    exit 1\n\
fi\n\
\n\
echo "启动应用..."\n\
exec ./lazybala' > /usr/local/bin/start.sh && \
    chmod +x /usr/local/bin/start.sh

# 下载最新的 yt-dlp（支持多架构）
ARG TARGETARCH
RUN case "${TARGETARCH}" in \
        "amd64") YT_DLP_FILE="yt-dlp_linux" ;; \
        "arm64") YT_DLP_FILE="yt-dlp_linux_aarch64" ;; \
        "arm") YT_DLP_FILE="yt-dlp_linux_armv7l" ;; \
        *) YT_DLP_FILE="yt-dlp_linux" ;; \
    esac && \
    echo "Downloading yt-dlp for ${TARGETARCH}: ${YT_DLP_FILE}" && \
    wget -O /app/bin/yt-dlp "https://github.com/yt-dlp/yt-dlp/releases/latest/download/${YT_DLP_FILE}" && \
    chmod +x /app/bin/yt-dlp

# 验证 yt-dlp 安装
RUN /app/bin/yt-dlp --version || echo "yt-dlp installation verification failed"

# 创建非 root 用户 (支持 PUID/PGID)
ARG PUID=1001
ARG PGID=1001
RUN addgroup -g ${PGID} -S lazybala && \
    adduser -S lazybala -u ${PUID} -G lazybala

# 更改目录所有权和权限
RUN chown -R ${PUID}:${PGID} /app && \
    chmod -R 755 /app && \
    chmod -R 777 /app/audiobooks /app/config /app/cookies /app/bin

# 切换到非 root 用户
USER lazybala

# 暴露端口
EXPOSE 8080

# 设置环境变量
ENV PORT=8080
ENV GIN_MODE=release
ENV DOCKER_ENV=true

# 健康检查 - 使用专用的健康检查端点
HEALTHCHECK --interval=30s --timeout=10s --start-period=60s --retries=3 \
  CMD curl -f http://localhost:8080/health || wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# 数据卷 (移除 /app/bin，保持 yt-dlp 在容器内)
VOLUME ["/app/audiobooks", "/app/config", "/app/cookies"]

# 添加标签
LABEL org.opencontainers.image.title="LazyBala" \
      org.opencontainers.image.description="A web-based media downloader with bilibili support" \
      org.opencontainers.image.vendor="Kis2Show" \
      org.opencontainers.image.licenses="MIT" \
      org.opencontainers.image.source="https://github.com/kis2show/lazybala"

# 启动命令
CMD ["/usr/local/bin/start.sh"]

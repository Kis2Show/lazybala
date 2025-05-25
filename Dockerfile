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

# 验证 Go 版本和模块文件
RUN go version
RUN cat go.mod
RUN go mod download -x

# 复制源代码
COPY *.go ./
COPY *.html ./

# 构建参数
ARG BUILDTIME
ARG VERSION

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
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

# 下载最新的 yt-dlp
RUN wget -O /app/bin/yt-dlp https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp_linux && \
    chmod +x /app/bin/yt-dlp

# 或者复制预下载的二进制工具（如果存在）
# ARG TARGETARCH
# COPY bin/yt-dlp_linux${TARGETARCH:+_$TARGETARCH} /app/bin/yt-dlp_linux
# COPY bin/ffmpeg_linux${TARGETARCH:+_$TARGETARCH} /app/bin/ffmpeg_linux
# COPY bin/ffprobe_linux${TARGETARCH:+_$TARGETARCH} /app/bin/ffprobe_linux
# RUN chmod +x /app/bin/*

# 创建非 root 用户
RUN addgroup -g 1001 -S lazybala && \
    adduser -S lazybala -u 1001 -G lazybala

# 更改目录所有权
RUN chown -R lazybala:lazybala /app

# 切换到非 root 用户
USER lazybala

# 暴露端口
EXPOSE 8080

# 设置环境变量
ENV PORT=8080
ENV GIN_MODE=release

# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=10s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/ || exit 1

# 数据卷
VOLUME ["/app/audiobooks", "/app/config", "/app/cookies"]

# 启动命令
CMD ["./lazybala"]

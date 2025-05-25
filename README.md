# LazyBala - 费劲巴拉下载器

![LazyBala Logo](https://img.shields.io/badge/LazyBala-费劲巴拉下载器-orange?style=for-the-badge)
![Go Version](https://img.shields.io/badge/Go-1.21+-blue?style=flat-square)
![Vue Version](https://img.shields.io/badge/Vue-3.0+-green?style=flat-square)
![License](https://img.shields.io/badge/License-MIT-yellow?style=flat-square)

> 一个简洁、高效的哔哩哔哩音频下载器，支持扫码登录、实时进度显示、断点续传等功能。

## ✨ 特性

- 🎵 **音频下载**: 支持哔哩哔哩视频音频提取
- 📱 **扫码登录**: 使用哔哩哔哩 APP 扫码登录获取 cookies
- 📊 **实时进度**: WebSocket 实时显示下载进度和状态
- 🔄 **断点续传**: 支持下载中断后恢复
- 📁 **批量下载**: 支持合集和播放列表下载
- ⚙️ **配置管理**: 持久化配置，支持自定义文件名格式
- 🔄 **自动更新**: 自动检测并更新 yt-dlp 到最新版本
- 🐳 **Docker 支持**: 提供多架构 Docker 镜像
- 📱 **响应式设计**: 适配桌面和移动设备

## 🚀 快速开始

### Docker 运行（推荐）

```bash
# 使用 docker-compose
git clone https://github.com/your-username/lazybala.git
cd lazybala
docker-compose up -d

# 或直接运行 Docker 镜像
docker run -d \
  --name lazybala \
  -p 8080:8080 \
  -v $(pwd)/audiobooks:/app/audiobooks \
  -v $(pwd)/config:/app/config \
  ghcr.io/your-username/lazybala:latest
```

### 本地开发

#### 前置要求

- Go 1.21+
- Node.js 18+
- Git

#### 安装步骤

1. **克隆仓库**
   ```bash
   git clone https://github.com/your-username/lazybala.git
   cd lazybala
   ```

2. **构建前端**
   ```bash
   cd frontend
   npm install
   npm run build
   cd ..
   ```

3. **安装 Go 依赖**
   ```bash
   go mod download
   ```

4. **运行应用**
   ```bash
   go run .
   ```

5. **访问应用**
   
   打开浏览器访问 `http://localhost:8080`

## 📖 使用说明

### 1. 扫码登录

首次使用需要扫码登录哔哩哔哩账号：

1. 点击"生成二维码"按钮
2. 使用哔哩哔哩 APP 扫描二维码
3. 在手机上确认登录
4. 登录成功后 cookies 会自动保存

### 2. 下载音频

1. **输入链接**: 在输入框中粘贴哔哩哔哩视频或合集链接
   - 普通视频: `https://www.bilibili.com/video/BV1KmzCYMEaq/`
   - 合集: `https://space.bilibili.com/2589478/lists/4279030?type=season`

2. **设置保存路径**: 可选择保存到 audiobooks 下的子文件夹

3. **高级设置**: 可自定义文件名格式
   - `%(title)s.%(ext)s` - 标题.扩展名
   - `%(uploader)s - %(title)s.%(ext)s` - 上传者 - 标题.扩展名

4. **开始下载**: 点击"开始下载"按钮

### 3. 监控进度

- 实时显示当前下载文件名和进度
- 显示下载速度和预计完成时间
- 合集下载时显示总体进度（n/总数）
- 支持缩略图预览

### 4. 配置管理

在设置页面可以：

- 设置默认保存路径
- 选择音频质量
- 自定义文件名格式
- 设置重试次数
- 检查和更新 yt-dlp 版本

## 🔧 配置说明

### 环境变量

| 变量名 | 默认值 | 说明 |
|--------|--------|------|
| `PORT` | `8080` | 服务端口 |
| `TZ` | `UTC` | 时区设置 |

### 目录结构

```
lazybala/
├── audiobooks/          # 下载的音频文件
├── config/             # 配置文件
│   ├── config.json     # 应用配置
│   └── yt.config       # yt-dlp 配置
├── cookies/            # 登录 cookies
│   └── cookies.txt     # 哔哩哔哩 cookies
├── bin/                # 二进制工具
│   ├── yt-dlp_linux    # yt-dlp (amd64)
│   ├── yt-dlp_linux_aarch64  # yt-dlp (arm64)
│   ├── ffmpeg_linux    # ffmpeg (amd64)
│   └── ffmpeg_linux_aarch64  # ffmpeg (arm64)
└── frontend/           # 前端文件
```

### 支持的链接格式

1. **普通视频**
   ```
   https://www.bilibili.com/video/BV1KmzCYMEaq/
   https://www.bilibili.com/video/BV1KmzCYMEaq?p=2
   ```

2. **合集/播放列表**
   ```
   https://space.bilibili.com/2589478/lists/4279030?type=season
   ```

3. **分享链接**
   ```
   分享地址：https://www.bilibili.com/video/BV1KmzCYMEaq?p=2/type=playlist
   ```

## 🛠️ 开发

### 技术栈

- **后端**: Go + Gin + WebSocket
- **前端**: Vue 3 + Vite + Pinia
- **工具**: yt-dlp + ffmpeg
- **容器**: Docker + Docker Compose

### API 接口

#### 登录相关
- `POST /api/qrcode/generate` - 生成登录二维码
- `POST /api/qrcode/scan` - 检查登录状态

#### 下载相关
- `POST /api/download` - 开始下载
- `GET /api/download/progress` - 获取下载进度
- `POST /api/download/stop` - 停止下载

#### 配置相关
- `GET /api/config` - 获取配置
- `POST /api/config` - 保存配置

#### 版本管理
- `GET /api/version/check` - 检查版本更新
- `POST /api/version/update` - 更新版本

### 构建

```bash
# 构建前端
cd frontend && npm run build

# 构建后端
go build -o lazybala .

# 构建 Docker 镜像
docker build -t lazybala .
```

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启 Pull Request

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 🙏 致谢

- [yt-dlp](https://github.com/yt-dlp/yt-dlp) - 强大的视频下载工具
- [FFmpeg](https://ffmpeg.org/) - 多媒体处理框架
- [Vue.js](https://vuejs.org/) - 渐进式 JavaScript 框架
- [Gin](https://gin-gonic.com/) - Go Web 框架

## 📞 支持

如果你觉得这个项目有用，请给它一个 ⭐️！

如有问题或建议，请提交 [Issue](https://github.com/your-username/lazybala/issues)。

---

**作者**: Kis2Show  
**项目**: LazyBala - 费劲巴拉下载器

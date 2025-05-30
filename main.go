package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// 构建时注入的版本信息
var (
	version   = "dev"
	buildTime = "unknown"
)

func init() {
	// 创建必要的目录
	createDirectories()

	// 检查并修复 yt-dlp 权限
	if err := ensureYtDlpExecutable(); err != nil {
		log.Printf("yt-dlp 权限检查失败: %v", err)
	}
}

func main() {
	// 初始化历史记录
	initializeHistory()

	// 设置 Gin 模式
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// 配置 CORS - 更安全的配置
	corsOrigins := []string{
		"http://localhost:8080",
		"http://127.0.0.1:8080",
		"http://localhost:3000", // 开发环境
		"http://127.0.0.1:3000", // 开发环境
	}

	// 从环境变量获取额外的允许源
	if extraOrigins := os.Getenv("CORS_ORIGINS"); extraOrigins != "" {
		corsOrigins = append(corsOrigins, extraOrigins)
	}

	r.Use(cors.New(cors.Config{
		AllowOrigins: corsOrigins,
		AllowMethods: []string{"GET", "POST", "OPTIONS"},
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Accept",
			"Authorization",
			"X-Requested-With",
		},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * 3600, // 12 小时
	}))

	// 添加安全头部中间件
	r.Use(func(c *gin.Context) {
		// 防止点击劫持
		c.Header("X-Frame-Options", "DENY")
		// 防止 MIME 类型嗅探
		c.Header("X-Content-Type-Options", "nosniff")
		// XSS 保护
		c.Header("X-XSS-Protection", "1; mode=block")
		// 引用者策略
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		// 内容安全策略
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; connect-src 'self' ws: wss:")
		c.Next()
	})

	// 添加 favicon 路由
	r.GET("/favicon.svg", func(c *gin.Context) {
		c.Status(http.StatusNotFound)
	})

	// 静态HTML文件路由
	r.GET("/test_websocket.html", func(c *gin.Context) {
		c.File("test_websocket.html")
	})

	r.GET("/simple.html", func(c *gin.Context) {
		c.File("simple.html")
	})

	r.GET("/test.html", func(c *gin.Context) {
		c.File("test.html")
	})

	// 首页路由 - 使用生产版本
	r.GET("/", func(c *gin.Context) {
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.File("index.html")
	})

	// 直接访问index.html
	r.GET("/index.html", func(c *gin.Context) {
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.File("index.html")
	})

	// 状态检查页面
	r.GET("/status", func(c *gin.Context) {
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.File("status.html")
	})

	// 健康检查端点 - 简单的 JSON 响应
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "lazybala",
		})
	})

	// API 路由
	api := r.Group("/api")
	{
		// 登录相关
		api.POST("/qrcode/generate", generateQRCode)
		api.POST("/qrcode/force-generate", forceGenerateQRCode)
		api.POST("/qrcode/scan", scanLogin)

		// 下载相关
		api.POST("/download/precheck", preCheckVideo)
		api.POST("/download", startDownloadHandler)
		api.GET("/download/progress", getDownloadProgress)
		api.GET("/download/status", getDownloadStatus)
		api.POST("/download/stop", stopDownload)
		api.POST("/download/pause", pauseDownload)
		api.POST("/download/resume", resumeDownload)
		api.GET("/download/history", getDownloadHistory)
		api.POST("/download/resume-background", resumeBackgroundDownload)

		// 配置相关
		api.GET("/config", getConfig)
		api.POST("/config", saveConfig)

		// 版本信息
		api.GET("/version", getVersionInfo)
		api.GET("/version/check", checkVersion)
		api.POST("/version/update", updateVersion)

		// 用户信息和认证
		api.GET("/user/info", getUserInfo)
		api.GET("/auth/export-cookies", exportCookies)
		api.POST("/auth/logout", logoutUser)
	}

	// WebSocket 路由
	r.GET("/ws", handleWebSocket)

	// 静态文件服务
	r.Static("/static", "./static")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("LazyBala %s (构建时间: %s) 服务启动在端口 %s", version, buildTime, port)
	log.Fatal(r.Run(":" + port))
}

func createDirectories() {
	dirs := []string{"audiobooks", "config", "cookies"}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0777); err != nil {
			log.Printf("创建目录 %s 失败: %v", dir, err)
		} else {
			// 尝试设置目录权限（Docker环境中可能失败，但不影响功能）
			if err := os.Chmod(dir, 0777); err != nil {
				// 在Docker环境中，挂载卷的权限修改可能失败，但不影响实际功能
				// 只在非Docker环境或确实有权限问题时记录警告
				if _, dockerEnv := os.LookupEnv("DOCKER_ENV"); !dockerEnv {
					log.Printf("警告: 设置目录 %s 权限失败: %v (功能不受影响)", dir, err)
				}
			}
		}
	}
}

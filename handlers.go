package main

import (
	"fmt"
	"net/http"
	"path/filepath"
	"runtime"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/skip2/go-qrcode"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// QR码生成响应
type QRCodeResponse struct {
	QRCodeKey string `json:"qrcode_key"`
	URL       string `json:"url"`
	QRCode    string `json:"qrcode"` // base64 编码的二维码图片
}

// 登录状态响应
type LoginStatusResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Cookies string `json:"cookies,omitempty"`
}

// 下载请求
type DownloadRequest struct {
	URL            string `json:"url"`
	SavePath       string `json:"save_path"`
	TitleRegex     string `json:"title_regex,omitempty"`
	Quality        string `json:"quality,omitempty"`
	RetryCount     int    `json:"retry_count,omitempty"`
	WriteThumbnail bool   `json:"write_thumbnail,omitempty"`
}

// 预检查请求
type PreCheckRequest struct {
	URL string `json:"url"`
}

// 预检查响应
type PreCheckResponse struct {
	Title      string      `json:"title"`
	Uploader   string      `json:"uploader"`
	Duration   string      `json:"duration"`
	Thumbnail  string      `json:"thumbnail"`
	AudioCount int         `json:"audio_count"`
	IsPlaylist bool        `json:"is_playlist"`
	Entries    []VideoInfo `json:"entries,omitempty"`
}

// 视频信息
type VideoInfo struct {
	Title     string `json:"title"`
	Duration  string `json:"duration"`
	Thumbnail string `json:"thumbnail"`
}

// DownloadProgress 已在 download.go 中定义

// 配置结构
type Config struct {
	SavePath       string `json:"save_path"`
	TitleRegex     string `json:"title_regex"`
	Quality        string `json:"quality"`
	RetryCount     int    `json:"retry_count"`
	WriteThumbnail bool   `json:"write_thumbnail"`
}

// 生成二维码
func generateQRCode(c *gin.Context) {
	// 检查是否已有有效的cookies
	if hasCookies() {
		fmt.Println("检测到cookies文件，正在验证有效性...")
		valid, err := checkCookiesValid()
		if err != nil {
			fmt.Printf("验证cookies时出错: %v\n", err)
		} else if valid {
			// cookies有效，返回已登录状态
			c.JSON(http.StatusOK, gin.H{
				"already_logged_in": true,
				"message":           "已成功登录，不需要扫码",
			})
			return
		} else {
			fmt.Println("cookies已失效，将生成新的二维码")
		}
	}

	// 调用哔哩哔哩 API 生成登录二维码
	qrData, err := generateBilibiliQRCode()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 生成二维码图片
	qrCodePNG, err := qrcode.Encode(qrData.URL, qrcode.Medium, 160)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成二维码失败"})
		return
	}

	response := QRCodeResponse{
		QRCodeKey: qrData.QRCodeKey,
		URL:       qrData.URL,
		QRCode:    encodeBase64(qrCodePNG),
	}

	c.JSON(http.StatusOK, response)
}

// 强制生成二维码（忽略现有cookies）
func forceGenerateQRCode(c *gin.Context) {
	fmt.Println("强制生成新的二维码，将覆盖现有cookies")

	// 调用哔哩哔哩 API 生成登录二维码
	qrData, err := generateBilibiliQRCode()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 生成二维码图片
	qrCodePNG, err := qrcode.Encode(qrData.URL, qrcode.Medium, 160)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成二维码失败"})
		return
	}

	response := QRCodeResponse{
		QRCodeKey: qrData.QRCodeKey,
		URL:       qrData.URL,
		QRCode:    encodeBase64(qrCodePNG),
	}

	c.JSON(http.StatusOK, response)
}

// 扫码登录
func scanLogin(c *gin.Context) {
	var req struct {
		QRCodeKey string `json:"qrcode_key"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	// 轮询检查登录状态
	status, cookies, err := checkLoginStatus(req.QRCodeKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 只有成功获取到cookies才认为真正登录成功
	actualStatus := status
	if status == 0 && cookies == "" {
		// B站API返回登录成功，但没有获取到有效cookies，认为登录未完成
		actualStatus = 86090 // 设置为"已扫码未确认"状态
		fmt.Println("B站API返回登录成功，但未获取到有效cookies")
	}

	response := LoginStatusResponse{
		Code:    actualStatus,
		Message: getLoginMessage(actualStatus),
	}

	if status == 0 && cookies != "" {
		// 真正登录成功，cookies已在checkLoginStatus中保存
		fmt.Printf("登录成功，获取到cookies，长度: %d\n", len(cookies))
		response.Cookies = cookies
		response.Code = 0 // 确保返回成功状态
		response.Message = "登录成功"
	}

	c.JSON(http.StatusOK, response)
}

// 预检查视频信息
func preCheckVideo(c *gin.Context) {
	var req PreCheckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	// 解析链接
	parsedURL, err := parseURL(req.URL)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "链接解析失败: " + err.Error()})
		return
	}

	// 获取视频信息
	info, err := getVideoInfo(parsedURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取视频信息失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, info)
}

// 开始下载 (HTTP处理器)
func startDownloadHandler(c *gin.Context) {
	// 检查是否已有正在进行的下载
	if hasActiveDownload() {
		c.JSON(http.StatusConflict, gin.H{"error": "已有下载任务正在进行中，请等待完成或停止当前任务"})
		return
	}

	var req DownloadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	// 解析链接
	parsedURL, err := parseURL(req.URL)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "链接解析失败: " + err.Error()})
		return
	}

	// 创建下载任务
	task := &DownloadTask{
		URL:            parsedURL,
		SavePath:       req.SavePath,
		TitleRegex:     req.TitleRegex,
		Quality:        req.Quality,
		RetryCount:     req.RetryCount,
		WriteThumbnail: req.WriteThumbnail,
	}

	// 启动下载任务
	go func() {
		// 添加任务开始记录
		addTaskToHistory(task, "downloading", "")

		if err := startDownload(task); err != nil {
			fmt.Printf("下载失败: %v\n", err)
			// 添加失败记录
			addTaskToHistory(task, "failed", err.Error())
		} else {
			fmt.Println("下载成功完成")
			// 添加成功记录
			addTaskToHistory(task, "completed", "")
		}
	}()

	c.JSON(http.StatusOK, gin.H{"message": "下载任务已启动"})
}

// 获取下载进度
func getDownloadProgress(c *gin.Context) {
	progress := getCurrentProgress()
	c.JSON(http.StatusOK, progress)
}

// 获取下载状态
func getDownloadStatus(c *gin.Context) {
	progress := getCurrentProgress()
	c.JSON(http.StatusOK, progress)
}

// 停止下载
func stopDownload(c *gin.Context) {
	stopCurrentDownload()
	c.JSON(http.StatusOK, gin.H{"message": "下载已停止"})
}

// 暂停下载
func pauseDownload(c *gin.Context) {
	if pauseCurrentDownload() {
		c.JSON(http.StatusOK, gin.H{"message": "下载已暂停"})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "没有正在进行的下载任务"})
	}
}

// 继续下载
func resumeDownload(c *gin.Context) {
	if resumeCurrentDownload() {
		c.JSON(http.StatusOK, gin.H{"message": "下载已继续"})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "没有暂停的下载任务"})
	}
}

// 获取下载历史
func getDownloadHistory(c *gin.Context) {
	history := getTaskHistory()
	hasActive := hasActiveDownload()

	c.JSON(http.StatusOK, gin.H{
		"tasks":         history,
		"hasActiveTask": hasActive,
	})
}

// 恢复后台下载
func resumeBackgroundDownload(c *gin.Context) {
	if hasActiveDownload() {
		c.JSON(http.StatusOK, gin.H{"message": "后台下载任务已恢复"})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "没有后台下载任务"})
	}
}

// 获取配置
func getConfig(c *gin.Context) {
	config, err := loadConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "加载配置失败"})
		return
	}

	// 检查cookies状态
	hasCookiesFile := hasCookies()
	cookiesValid := false

	if hasCookiesFile {
		fmt.Println("检测到已保存的cookies文件，正在验证有效性...")
		valid, err := checkCookiesValid()
		if err != nil {
			fmt.Printf("验证cookies时出错: %v\n", err)
		} else {
			cookiesValid = valid
			if valid {
				fmt.Println("cookies验证通过，用户已登录")
			} else {
				fmt.Println("cookies已失效，需要重新登录")
			}
		}
	}

	// 添加cookies状态信息
	response := map[string]interface{}{
		"save_path":       config.SavePath,
		"title_regex":     config.TitleRegex,
		"quality":         config.Quality,
		"retry_count":     config.RetryCount,
		"write_thumbnail": config.WriteThumbnail,
		"has_cookies":     hasCookiesFile,
		"cookies_valid":   cookiesValid,
	}

	c.JSON(http.StatusOK, response)
}

// 保存配置
func saveConfig(c *gin.Context) {
	var config Config
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	if err := saveConfigToFile(config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存配置失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "配置已保存"})
}

// 获取版本信息
func getVersionInfo(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"version":    version,
		"build_time": buildTime,
		"go_version": runtime.Version(),
		"platform":   runtime.GOOS + "/" + runtime.GOARCH,
	})
}

// 检查版本更新
func checkVersion(c *gin.Context) {
	latestVersion, downloadURL, err := checkLatestVersion()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	currentVersion, _ := getCurrentYtDlpVersion()

	c.JSON(http.StatusOK, gin.H{
		"current_version": currentVersion,
		"latest_version":  latestVersion,
		"has_update":      latestVersion != currentVersion,
		"download_url":    downloadURL,
	})
}

// 更新版本
func updateVersion(c *gin.Context) {
	go func() {
		if err := updateYtDlp(); err != nil {
			fmt.Printf("更新失败: %v\n", err)
		}
	}()

	c.JSON(http.StatusOK, gin.H{"message": "更新任务已启动"})
}

// WebSocket 处理
func handleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	// 添加到连接池
	addWebSocketConnection(conn)
	defer removeWebSocketConnection(conn)

	// 保持连接
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

// 辅助函数
func getLoginMessage(code int) string {
	switch code {
	case 86101:
		return "未扫码"
	case 86090:
		return "已扫码未确认"
	case 0:
		return "登录成功"
	default:
		return "登录失败"
	}
}

func getBinaryPath(name string) string {
	os := runtime.GOOS
	arch := runtime.GOARCH

	if os == "windows" {
		return filepath.Join("bin", name+"_win64.exe")
	} else if os == "linux" {
		if arch == "amd64" {
			return filepath.Join("bin", name+"_linux")
		} else if arch == "arm64" {
			return filepath.Join("bin", name+"_linux_aarch64")
		}
		return filepath.Join("bin", name+"_linux")
	}

	// 默认返回Linux版本
	return filepath.Join("bin", name+"_linux")
}

// 获取用户信息
func getUserInfo(c *gin.Context) {
	// 检查是否有有效的cookies
	if !hasCookies() {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	valid, err := checkCookiesValid()
	if err != nil || !valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "登录已过期"})
		return
	}

	// 获取用户信息
	userInfo, err := getBilibiliUserInfo()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户信息失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, userInfo)
}

// 导出Cookies
func exportCookies(c *gin.Context) {
	// 检查是否有cookies文件
	if !hasCookies() {
		c.JSON(http.StatusNotFound, gin.H{"error": "没有找到cookies文件"})
		return
	}

	// 读取cookies文件内容
	cookies, err := readCookiesFile()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取cookies文件失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"cookies": cookies,
		"message": "cookies导出成功",
	})
}

// 用户登出
func logoutUser(c *gin.Context) {
	// 删除cookies文件
	err := deleteCookiesFile()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "清除登录信息失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "登录信息已清除"})
}

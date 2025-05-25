package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// 哔哩哔哩 API 响应结构
type BilibiliQRResponse struct {
	Code int `json:"code"`
	Data struct {
		URL       string `json:"url"`
		QRCodeKey string `json:"qrcode_key"`
	} `json:"data"`
}

type BilibiliLoginResponse struct {
	Code int `json:"code"`
	Data struct {
		URL          string `json:"url"`
		RefreshToken string `json:"refresh_token"`
		Timestamp    int64  `json:"timestamp"`
		Code         int    `json:"code"`    // 内层状态码
		Message      string `json:"message"` // 内层消息
	} `json:"data"`
}

type QRData struct {
	URL       string
	QRCodeKey string
}

// 生成哔哩哔哩登录二维码
func generateBilibiliQRCode() (*QRData, error) {
	resp, err := http.Get("https://passport.bilibili.com/x/passport-login/web/qrcode/generate")
	if err != nil {
		return nil, fmt.Errorf("请求二维码生成接口失败: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	fmt.Printf("QR生成API响应: %s\n", string(body))

	var qrResp BilibiliQRResponse
	if err := json.Unmarshal(body, &qrResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	if qrResp.Code != 0 {
		return nil, fmt.Errorf("生成二维码失败，错误码: %d", qrResp.Code)
	}

	fmt.Printf("生成QR码成功: URL=%s, Key=%s\n", qrResp.Data.URL, qrResp.Data.QRCodeKey)

	return &QRData{
		URL:       qrResp.Data.URL,
		QRCodeKey: qrResp.Data.QRCodeKey,
	}, nil
}

// 检查登录状态
func checkLoginStatus(qrcodeKey string) (int, string, error) {
	apiURL := fmt.Sprintf("https://passport.bilibili.com/x/passport-login/web/qrcode/poll?qrcode_key=%s", qrcodeKey)

	fmt.Printf("检查登录状态: qrcode_key=%s\n", qrcodeKey)

	resp, err := http.Get(apiURL)
	if err != nil {
		return -1, "", fmt.Errorf("请求登录状态接口失败: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return -1, "", fmt.Errorf("读取响应失败: %v", err)
	}

	fmt.Printf("登录状态API响应: %s\n", string(body))

	var loginResp BilibiliLoginResponse
	if err := json.Unmarshal(body, &loginResp); err != nil {
		return -1, "", fmt.Errorf("解析响应失败: %v", err)
	}

	fmt.Printf("B站API响应: 外层Code=%d, 内层Code=%d, Message=%s, URL=%s\n",
		loginResp.Code, loginResp.Data.Code, loginResp.Data.Message, loginResp.Data.URL)

	// 使用内层的状态码作为真正的登录状态
	actualCode := loginResp.Data.Code

	// 如果登录成功，提取并保存 cookies
	var cookies string
	if actualCode == 0 && loginResp.Data.URL != "" {
		fmt.Printf("开始提取cookies，URL: %s\n", loginResp.Data.URL)
		cookies = extractCookiesFromURL(loginResp.Data.URL)
		if cookies != "" {
			fmt.Printf("成功提取cookies，长度: %d\n", len(cookies))
			// 立即保存cookies到文件
			if err := saveCookies(cookies); err != nil {
				fmt.Printf("保存cookies失败: %v\n", err)
			} else {
				fmt.Println("cookies已保存到文件")
			}
		} else {
			fmt.Println("未能提取到有效cookies")
		}
	} else if actualCode == 0 {
		fmt.Println("登录成功但URL为空")
	}

	return actualCode, cookies, nil
}

// 从 URL 中提取 cookies
func extractCookiesFromURL(urlStr string) string {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		fmt.Printf("解析URL失败: %v\n", err)
		return ""
	}

	// 从URL参数中提取cookies信息
	query := parsedURL.Query()

	// 构建yt-dlp可用的cookies格式
	var cookieParts []string

	// 提取常见的B站认证cookies
	if sessData := query.Get("SESSDATA"); sessData != "" {
		cookieParts = append(cookieParts, "# Netscape HTTP Cookie File")
		cookieParts = append(cookieParts, fmt.Sprintf(".bilibili.com\tTRUE\t/\tTRUE\t0\tSESSDATA\t%s", sessData))
	}

	if biliJct := query.Get("bili_jct"); biliJct != "" {
		cookieParts = append(cookieParts, fmt.Sprintf(".bilibili.com\tTRUE\t/\tTRUE\t0\tbili_jct\t%s", biliJct))
	}

	if dedeUserId := query.Get("DedeUserID"); dedeUserId != "" {
		cookieParts = append(cookieParts, fmt.Sprintf(".bilibili.com\tTRUE\t/\tTRUE\t0\tDedeUserID\t%s", dedeUserId))
	}

	if dedeUserIdCkMd5 := query.Get("DedeUserID__ckMd5"); dedeUserIdCkMd5 != "" {
		cookieParts = append(cookieParts, fmt.Sprintf(".bilibili.com\tTRUE\t/\tTRUE\t0\tDedeUserID__ckMd5\t%s", dedeUserIdCkMd5))
	}

	if sid := query.Get("sid"); sid != "" {
		cookieParts = append(cookieParts, fmt.Sprintf(".bilibili.com\tTRUE\t/\tTRUE\t0\tsid\t%s", sid))
	}

	if len(cookieParts) > 1 { // 至少有header和一个cookie
		return strings.Join(cookieParts, "\n")
	}

	fmt.Printf("未找到有效的cookies信息，URL: %s\n", urlStr)
	return ""
}

// 保存 cookies 到文件
func saveCookies(cookies string) error {
	cookiesDir := "cookies"
	if err := os.MkdirAll(cookiesDir, 0755); err != nil {
		return err
	}

	cookiesFile := filepath.Join(cookiesDir, "cookies.txt")
	return os.WriteFile(cookiesFile, []byte(cookies), 0644)
}

// 解析哔哩哔哩链接
func parseURL(inputURL string) (string, error) {
	// 正则表达式匹配 BV 号
	bvRegex := regexp.MustCompile(`BV[a-zA-Z0-9]+`)
	bvMatch := bvRegex.FindString(inputURL)

	if bvMatch != "" {
		// 普通视频链接
		return fmt.Sprintf("https://www.bilibili.com/video/%s/", bvMatch), nil
	}

	// 检查是否为合集链接
	// 匹配格式: https://space.bilibili.com/{uid}/lists/{sid}?type=season
	spaceRegex := regexp.MustCompile(`space\.bilibili\.com/(\d+)/lists/(\d+)`)
	spaceMatch := spaceRegex.FindStringSubmatch(inputURL)

	if len(spaceMatch) == 3 {
		uid := spaceMatch[1]
		sid := spaceMatch[2]
		return fmt.Sprintf("https://space.bilibili.com/%s/lists/%s?type=season", uid, sid), nil
	}

	// 如果都不匹配，返回原链接
	if strings.Contains(inputURL, "bilibili.com") {
		return inputURL, nil
	}

	return "", fmt.Errorf("无法识别的哔哩哔哩链接格式")
}

// Base64 编码
func encodeBase64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

// 获取 cookies 文件路径
func getCookiesPath() string {
	return filepath.Join("cookies", "cookies.txt")
}

// 检查 cookies 是否存在
func hasCookies() bool {
	cookiesPath := getCookiesPath()
	if _, err := os.Stat(cookiesPath); os.IsNotExist(err) {
		return false
	}
	return true
}

// 检查cookies是否有效
func checkCookiesValid() (bool, error) {
	if !hasCookies() {
		return false, nil
	}

	// 读取cookies内容
	cookies, err := readCookies()
	if err != nil {
		return false, err
	}

	// 检查cookies是否包含必要的字段
	if !strings.Contains(cookies, "SESSDATA") {
		fmt.Println("cookies缺少SESSDATA字段")
		return false, nil
	}

	// 通过访问B站API来验证cookies有效性
	return validateCookiesWithAPI()
}

// 通过API验证cookies有效性
func validateCookiesWithAPI() (bool, error) {
	// 使用cookies访问B站用户信息API
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("GET", "https://api.bilibili.com/x/web-interface/nav", nil)
	if err != nil {
		return false, err
	}

	// 从cookies文件中提取cookies并设置到请求中
	if err := setCookiesFromFile(req); err != nil {
		return false, err
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("验证cookies失败: %v\n", err)
		return false, nil // 网络错误不算cookies无效
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}

	// 解析响应
	var navResp struct {
		Code int `json:"code"`
		Data struct {
			IsLogin bool `json:"isLogin"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &navResp); err != nil {
		return false, err
	}

	isValid := navResp.Code == 0 && navResp.Data.IsLogin
	if isValid {
		fmt.Println("cookies验证成功，用户已登录")
	} else {
		fmt.Printf("cookies验证失败: code=%d, isLogin=%v\n", navResp.Code, navResp.Data.IsLogin)
	}

	return isValid, nil
}

// 从cookies文件中提取cookies并设置到HTTP请求中
func setCookiesFromFile(req *http.Request) error {
	cookiesContent, err := readCookies()
	if err != nil {
		return err
	}

	// 解析Netscape格式的cookies
	lines := strings.Split(cookiesContent, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Netscape格式: domain	flag	path	secure	expiration	name	value
		parts := strings.Split(line, "\t")
		if len(parts) >= 7 {
			name := parts[5]
			value := parts[6]
			if name != "" && value != "" {
				cookie := &http.Cookie{
					Name:  name,
					Value: value,
				}
				req.AddCookie(cookie)
			}
		}
	}

	return nil
}

// 读取 cookies
func readCookies() (string, error) {
	cookiesPath := getCookiesPath()
	data, err := os.ReadFile(cookiesPath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// 获取视频信息
func getVideoInfo(url string) (*PreCheckResponse, error) {
	ytDlpPath := getBinaryPath("yt-dlp")

	// 首先获取播放列表基本信息
	args := []string{
		"--dump-json",
		"--no-download",
		"--flat-playlist",
	}

	// 添加 cookies 如果存在
	if hasCookies() {
		args = append(args, "--cookies", getCookiesPath())
	}

	args = append(args, url)

	cmd := exec.Command(ytDlpPath, args...)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("执行yt-dlp失败: %v", err)
	}

	// 解析JSON输出
	lines := strings.Split(string(output), "\n")
	var entries []VideoInfo
	var mainInfo map[string]interface{}
	var playlistTitle string
	var playlistUploader string
	var totalCount int

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		var info map[string]interface{}
		if err := json.Unmarshal([]byte(line), &info); err != nil {
			continue
		}

		if mainInfo == nil {
			mainInfo = info
			playlistTitle = getString(info, "playlist")
			if playlistTitle == "" {
				playlistTitle = getString(info, "playlist_title")
			}
			playlistUploader = getString(info, "playlist_uploader")
			totalCount = int(getFloat64(info, "n_entries"))
		}

		// 对于播放列表，我们只显示前几个条目的基本信息
		if len(entries) < 5 {
			entries = append(entries, VideoInfo{
				Title:     fmt.Sprintf("第%d集", len(entries)+1),
				Duration:  "未知",
				Thumbnail: "",
			})
		}
	}

	if mainInfo == nil {
		return nil, fmt.Errorf("无法获取视频信息")
	}

	// 如果是单个视频，获取详细信息
	if totalCount <= 1 {
		return getSingleVideoInfo(url)
	}

	// 构建播放列表响应
	response := &PreCheckResponse{
		Title:      playlistTitle,
		Uploader:   playlistUploader,
		Duration:   fmt.Sprintf("共%d集", totalCount),
		Thumbnail:  "",
		AudioCount: totalCount,
		IsPlaylist: totalCount > 1,
		Entries:    entries,
	}

	if response.AudioCount == 0 {
		response.AudioCount = 1
	}

	return response, nil
}

// 获取单个视频详细信息
func getSingleVideoInfo(url string) (*PreCheckResponse, error) {
	ytDlpPath := getBinaryPath("yt-dlp")

	args := []string{
		"--dump-json",
		"--no-download",
	}

	// 添加 cookies 如果存在
	if hasCookies() {
		args = append(args, "--cookies", getCookiesPath())
	}

	args = append(args, url)

	cmd := exec.Command(ytDlpPath, args...)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("执行yt-dlp失败: %v", err)
	}

	var info map[string]interface{}
	if err := json.Unmarshal(output, &info); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	response := &PreCheckResponse{
		Title:      getString(info, "title"),
		Uploader:   getString(info, "uploader"),
		Duration:   formatDuration(getFloat64(info, "duration")),
		Thumbnail:  getString(info, "thumbnail"),
		AudioCount: 1,
		IsPlaylist: false,
		Entries:    nil,
	}

	return response, nil
}

// 辅助函数：从map中安全获取字符串
func getString(m map[string]interface{}, key string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return ""
}

// 辅助函数：从map中安全获取float64
func getFloat64(m map[string]interface{}, key string) float64 {
	if val, ok := m[key].(float64); ok {
		return val
	}
	return 0
}

// 辅助函数：格式化时长
func formatDuration(seconds float64) string {
	if seconds <= 0 {
		return "未知"
	}

	hours := int(seconds) / 3600
	minutes := (int(seconds) % 3600) / 60
	secs := int(seconds) % 60

	if hours > 0 {
		return fmt.Sprintf("%d:%02d:%02d", hours, minutes, secs)
	}
	return fmt.Sprintf("%d:%02d", minutes, secs)
}

// 用户信息结构
type BilibiliUserInfo struct {
	Name       string `json:"name"`
	Level      int    `json:"level"`
	VipType    int    `json:"vip_type"`
	VipStatus  int    `json:"vip_status"`
	VipDueDate int64  `json:"vip_due_date"`
	Following  int    `json:"following"`
	Follower   int    `json:"follower"`
}

// 获取哔哩哔哩用户信息
func getBilibiliUserInfo() (*BilibiliUserInfo, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// 获取用户基本信息
	req, err := http.NewRequest("GET", "https://api.bilibili.com/x/web-interface/nav", nil)
	if err != nil {
		return nil, err
	}

	// 设置cookies
	if err := setCookiesFromFile(req); err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// 解析用户信息
	var navResp struct {
		Code int `json:"code"`
		Data struct {
			IsLogin   bool   `json:"isLogin"`
			Uname     string `json:"uname"`
			Level     int    `json:"level_info.current_level"`
			VipType   int    `json:"vipType"`
			VipStatus int    `json:"vipStatus"`
			VipDue    int64  `json:"vipDueDate"`
			Mid       int64  `json:"mid"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &navResp); err != nil {
		return nil, err
	}

	if navResp.Code != 0 || !navResp.Data.IsLogin {
		return nil, fmt.Errorf("用户未登录或cookies无效")
	}

	userInfo := &BilibiliUserInfo{
		Name:       navResp.Data.Uname,
		Level:      navResp.Data.Level,
		VipType:    navResp.Data.VipType,
		VipStatus:  navResp.Data.VipStatus,
		VipDueDate: navResp.Data.VipDue,
	}

	// 获取关注和粉丝数（可选，如果API允许）
	if navResp.Data.Mid > 0 {
		following, follower := getUserStats(navResp.Data.Mid)
		userInfo.Following = following
		userInfo.Follower = follower
	}

	return userInfo, nil
}

// 获取用户统计信息（关注数、粉丝数）
func getUserStats(mid int64) (int, int) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	url := fmt.Sprintf("https://api.bilibili.com/x/relation/stat?vmid=%d", mid)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, 0
	}

	// 设置cookies
	setCookiesFromFile(req)

	resp, err := client.Do(req)
	if err != nil {
		return 0, 0
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, 0
	}

	var statResp struct {
		Code int `json:"code"`
		Data struct {
			Following int `json:"following"`
			Follower  int `json:"follower"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &statResp); err != nil {
		return 0, 0
	}

	if statResp.Code != 0 {
		return 0, 0
	}

	return statResp.Data.Following, statResp.Data.Follower
}

// 读取cookies文件内容
func readCookiesFile() (string, error) {
	return readCookies()
}

// 删除cookies文件
func deleteCookiesFile() error {
	cookiesPath := getCookiesPath()
	if _, err := os.Stat(cookiesPath); os.IsNotExist(err) {
		return nil // 文件不存在，认为删除成功
	}
	return os.Remove(cookiesPath)
}

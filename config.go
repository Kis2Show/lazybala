package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

// 默认配置
var defaultConfig = Config{
	SavePath:       "",
	TitleRegex:     "%(title)s.%(ext)s",
	Quality:        "bestaudio/best",
	RetryCount:     5,
	WriteThumbnail: true,
}

// 加载配置
func loadConfig() (*Config, error) {
	configPath := filepath.Join("config", "config.json")

	// 如果配置文件不存在，返回默认配置
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return &defaultConfig, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %v", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %v", err)
	}

	return &config, nil
}

// 保存配置到文件
func saveConfigToFile(config Config) error {
	configDir := "config"
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("创建配置目录失败: %v", err)
	}

	configPath := filepath.Join(configDir, "config.json")
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %v", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %v", err)
	}

	// 同时生成 yt-dlp 配置文件
	return generateYtDlpConfig(config)
}

// 生成 yt-dlp 配置文件
func generateYtDlpConfig(config Config) error {
	configDir := "config"
	ytConfigPath := filepath.Join(configDir, "yt.config")

	var lines []string

	// 音频质量
	lines = append(lines, fmt.Sprintf("-f %s", config.Quality))

	// 保存路径
	if config.SavePath != "" {
		lines = append(lines, fmt.Sprintf("-P ./audiobooks/%s", config.SavePath))
	} else {
		lines = append(lines, "-P ./audiobooks")
	}

	// 输出格式
	lines = append(lines, fmt.Sprintf("-o \"%s\"", config.TitleRegex))

	// 重试次数
	lines = append(lines, fmt.Sprintf("--extractor-retries %d", config.RetryCount))

	// 缩略图
	if config.WriteThumbnail {
		lines = append(lines, "--write-thumbnail")
	}

	// 其他固定选项
	lines = append(lines, "--newline")
	lines = append(lines, "--progress")

	// cookies
	if hasCookies() {
		lines = append(lines, fmt.Sprintf("--cookies %s", getCookiesPath()))
	}

	content := strings.Join(lines, "\n")
	return os.WriteFile(ytConfigPath, []byte(content), 0644)
}

// GitHub Release 结构
type GitHubRelease struct {
	TagName string `json:"tag_name"`
	HTMLURL string `json:"html_url"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

// 检查LazyBala应用最新版本
func checkLatestAppVersion() (string, string, error) {
	resp, err := http.Get("https://api.github.com/repos/kis2show/lazybala/releases/latest")
	if err != nil {
		return "", "", fmt.Errorf("获取LazyBala版本信息失败: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", fmt.Errorf("读取响应失败: %v", err)
	}

	var release GitHubRelease
	if err := json.Unmarshal(body, &release); err != nil {
		return "", "", fmt.Errorf("解析响应失败: %v", err)
	}

	// 获取发布页面链接
	downloadURL := release.HTMLURL
	if downloadURL == "" {
		downloadURL = "https://github.com/kis2show/lazybala/releases/latest"
	}

	return release.TagName, downloadURL, nil
}

// 检查yt-dlp最新版本
func checkLatestYtDlpVersion() (string, string, error) {
	resp, err := http.Get("https://api.github.com/repos/yt-dlp/yt-dlp/releases/latest")
	if err != nil {
		return "", "", fmt.Errorf("获取yt-dlp版本信息失败: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", fmt.Errorf("读取响应失败: %v", err)
	}

	var release GitHubRelease
	if err := json.Unmarshal(body, &release); err != nil {
		return "", "", fmt.Errorf("解析响应失败: %v", err)
	}

	// 根据系统架构选择下载链接
	arch := runtime.GOARCH
	var targetName string
	if arch == "amd64" {
		targetName = "yt-dlp_linux"
	} else if arch == "arm64" {
		targetName = "yt-dlp_linux_aarch64"
	} else {
		targetName = "yt-dlp_linux"
	}

	var downloadURL string
	for _, asset := range release.Assets {
		if asset.Name == targetName {
			downloadURL = asset.BrowserDownloadURL
			break
		}
	}

	if downloadURL == "" {
		return "", "", fmt.Errorf("未找到适合的下载链接")
	}

	return release.TagName, downloadURL, nil
}

// 保持向后兼容的函数
func checkLatestVersion() (string, string, error) {
	return checkLatestYtDlpVersion()
}

// 获取当前 yt-dlp 版本
func getCurrentYtDlpVersion() (string, error) {
	ytDlpPath := getYtDlpPath()

	if _, err := os.Stat(ytDlpPath); os.IsNotExist(err) {
		return "未安装", nil
	}

	// 尝试执行 yt-dlp --version 来获取真实版本
	cmd := exec.Command(ytDlpPath, "--version")
	output, err := cmd.Output()
	if err != nil {
		// 如果执行失败，返回文件修改时间作为版本标识
		info, statErr := os.Stat(ytDlpPath)
		if statErr != nil {
			return "未知", statErr
		}
		return "文件版本: " + info.ModTime().Format("2006-01-02"), nil
	}

	version := strings.TrimSpace(string(output))
	if version == "" {
		return "未知版本", nil
	}

	return version, nil
}

// 版本比较函数
func compareVersions(current, latest string) int {
	// 处理特殊情况
	if current == "dev" || current == "unknown" {
		return -1 // 开发版本总是认为需要更新
	}

	// 移除版本号前的 'v' 前缀
	current = strings.TrimPrefix(current, "v")
	latest = strings.TrimPrefix(latest, "v")

	// 简单的字符串比较（对于语义版本号）
	if current == latest {
		return 0
	}

	// 分割版本号进行比较
	currentParts := strings.Split(current, ".")
	latestParts := strings.Split(latest, ".")

	// 补齐版本号长度
	maxLen := len(currentParts)
	if len(latestParts) > maxLen {
		maxLen = len(latestParts)
	}

	for len(currentParts) < maxLen {
		currentParts = append(currentParts, "0")
	}
	for len(latestParts) < maxLen {
		latestParts = append(latestParts, "0")
	}

	// 逐个比较版本号部分
	for i := 0; i < maxLen; i++ {
		currentNum := parseVersionPart(currentParts[i])
		latestNum := parseVersionPart(latestParts[i])

		if currentNum < latestNum {
			return -1
		} else if currentNum > latestNum {
			return 1
		}
	}

	return 0
}

// 解析版本号部分
func parseVersionPart(part string) int {
	// 移除非数字字符
	numStr := ""
	for _, char := range part {
		if char >= '0' && char <= '9' {
			numStr += string(char)
		} else {
			break
		}
	}

	if numStr == "" {
		return 0
	}

	num, err := strconv.Atoi(numStr)
	if err != nil {
		return 0
	}

	return num
}

// 更新 yt-dlp
func updateYtDlp() error {
	latestVersion, downloadURL, err := checkLatestVersion()
	if err != nil {
		return err
	}

	// 下载新版本
	resp, err := http.Get(downloadURL)
	if err != nil {
		return fmt.Errorf("下载失败: %v", err)
	}
	defer resp.Body.Close()

	// 保存到临时文件
	tempPath := filepath.Join("bin", "yt-dlp_temp")
	tempFile, err := os.Create(tempPath)
	if err != nil {
		return fmt.Errorf("创建临时文件失败: %v", err)
	}
	defer tempFile.Close()

	_, err = io.Copy(tempFile, resp.Body)
	if err != nil {
		return fmt.Errorf("保存文件失败: %v", err)
	}

	// 设置执行权限
	if err := os.Chmod(tempPath, 0755); err != nil {
		return fmt.Errorf("设置权限失败: %v", err)
	}

	// 替换原文件
	ytDlpPath := getYtDlpPath()
	if err := os.Rename(tempPath, ytDlpPath); err != nil {
		return fmt.Errorf("替换文件失败: %v", err)
	}

	fmt.Printf("yt-dlp 已更新到版本 %s\n", latestVersion)
	return nil
}

// 验证标题格式（未使用，保留以备将来使用）
// func validateTitleRegex(regex string) error {
// 	// 检查是否包含必要的占位符
// 	if !strings.Contains(regex, "%(title)s") && !strings.Contains(regex, "%(ext)s") {
// 		return errors.New("正则表达式必须包含 %(title)s 和 %(ext)s")
// 	}
//
// 	// 检查正则表达式语法
// 	_, err := regexp.Compile(regex)
// 	return err
// }

// 获取可用的音频格式（未使用，保留以备将来使用）
// func getAvailableFormats() []string {
// 	return []string{
// 		"bestaudio/best",
// 		"bestaudio[ext=m4a]/best[ext=m4a]",
// 		"bestaudio[ext=mp3]/best[ext=mp3]",
// 		"worst",
// 	}
// }

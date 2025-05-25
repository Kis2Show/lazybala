package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/gorilla/websocket"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

var (
	currentDownload *DownloadTask
	downloadMutex   sync.RWMutex
	wsConnections   []*websocket.Conn
	wsConnMutex     sync.RWMutex
)

type DownloadTask struct {
	URL            string
	SavePath       string
	TitleRegex     string
	Progress       *DownloadProgress
	IsRunning      bool
	Cancel         context.CancelFunc
	Cmd            *exec.Cmd // 添加命令引用以支持暂停/继续
	Quality        string    // 添加质量设置
	RetryCount     int       // 添加重试次数
	WriteThumbnail bool      // 添加缩略图设置
}

type DownloadProgress struct {
	IsDownloading  bool     `json:"isDownloading"`
	IsPaused       bool     `json:"isPaused"`
	Progress       float64  `json:"progress"`     // 整体进度
	FileProgress   float64  `json:"fileProgress"` // 当前文件进度
	Speed          string   `json:"speed"`
	ETA            string   `json:"eta"`
	CurrentFile    string   `json:"currentFile"`
	Status         string   `json:"status"`
	LastActivity   string   `json:"lastActivity"`
	CurrentIndex   int      `json:"currentIndex"`
	TotalCount     int      `json:"totalCount"`
	CompletedFiles []string `json:"completedFiles"`
	CurrentTitle   string   `json:"currentTitle"`

	// 新增字段
	FileSize       string `json:"fileSize"`       // 文件大小
	Duration       string `json:"duration"`       // 视频时长
	Uploader       string `json:"uploader"`       // 上传者
	ViewCount      string `json:"viewCount"`      // 播放量
	Thumbnail      string `json:"thumbnail"`      // 缩略图URL
	PlaylistTitle  string `json:"playlistTitle"`  // 播放列表标题
	ErrorMessage   string `json:"errorMessage"`   // 错误信息
	WarningMessage string `json:"warningMessage"` // 警告信息
	StartTime      string `json:"startTime"`      // 开始时间
	Phase          string `json:"phase"`          // 当前阶段：extracting, downloading, merging, completed
}

// 获取yt-dlp可执行文件路径
func getYtDlpPath() string {
	var filename string
	switch runtime.GOOS {
	case "windows":
		filename = "yt-dlp_win64.exe"
	case "linux":
		filename = "yt-dlp_linux"
	case "darwin":
		filename = "yt-dlp_macos"
	default:
		filename = "yt-dlp"
	}
	return filepath.Join("bin", filename)
}

// isAudioFile 检查文件是否为音频文件
func isAudioFile(fileName string) bool {
	audioExtensions := []string{".m4a", ".mp3", ".opus", ".webm", ".aac", ".flac", ".wav", ".ogg"}

	// 清理文件名，确保UTF-8编码正确处理
	fileName = cleanUTF8String(fileName)
	fileName = strings.ToLower(fileName)

	for _, ext := range audioExtensions {
		if strings.HasSuffix(fileName, ext) {
			return true
		}
	}
	return false
}

// scanExistingAudioFiles 扫描指定目录下的所有音频文件
func scanExistingAudioFiles(dirPath string) []string {
	var audioFiles []string

	// 检查目录是否存在
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		fmt.Printf("目录不存在: %s\n", dirPath)
		return audioFiles
	}

	// 遍历目录
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("访问文件失败: %s, 错误: %v\n", path, err)
			return nil // 继续遍历其他文件
		}

		// 跳过目录
		if info.IsDir() {
			return nil
		}

		// 获取相对于目录的文件名
		relPath, err := filepath.Rel(dirPath, path)
		if err != nil {
			fmt.Printf("获取相对路径失败: %s, 错误: %v\n", path, err)
			return nil
		}

		// 只取文件名，不包含路径
		fileName := filepath.Base(relPath)

		// 检查是否为音频文件
		if isAudioFile(fileName) {
			audioFiles = append(audioFiles, fileName)
			fmt.Printf("发现音频文件: %s\n", fileName)
		}

		return nil
	})

	if err != nil {
		fmt.Printf("扫描目录失败: %s, 错误: %v\n", dirPath, err)
	}

	fmt.Printf("扫描完成，共发现 %d 个音频文件\n", len(audioFiles))
	return audioFiles
}

// 清理UTF-8字符串，移除无效字符并处理编码转换
func cleanUTF8String(s string) string {
	if utf8.ValidString(s) {
		return s
	}

	// 在Windows系统上，尝试从GBK转换到UTF-8
	if runtime.GOOS == "windows" {
		if converted, err := convertGBKToUTF8(s); err == nil && utf8.ValidString(converted) {
			return converted
		}
	}

	// 如果字符串包含无效UTF-8字符，清理它们
	var buf strings.Builder
	for _, r := range s {
		if r != utf8.RuneError {
			buf.WriteRune(r)
		}
	}
	return buf.String()
}

// 将GBK编码转换为UTF-8
func convertGBKToUTF8(s string) (string, error) {
	reader := transform.NewReader(strings.NewReader(s), simplifiedchinese.GBK.NewDecoder())
	decoded, err := io.ReadAll(reader)
	if err != nil {
		return s, err
	}
	return string(decoded), nil
}

// 开始下载
func startDownload(task *DownloadTask) error {
	// 检查是否已有正在进行的下载
	downloadMutex.Lock()
	if currentDownload != nil && currentDownload.IsRunning {
		downloadMutex.Unlock()
		return fmt.Errorf("已有下载任务正在进行中")
	}

	// 设置当前下载任务
	ctx, cancel := context.WithCancel(context.Background())
	task.Cancel = cancel
	task.IsRunning = true
	task.Progress = &DownloadProgress{
		IsDownloading:  true,
		Status:         "准备开始下载...",
		LastActivity:   "初始化下载任务",
		CompletedFiles: make([]string, 0),
		StartTime:      time.Now().Format("2006-01-02 15:04:05"),
		Phase:          "initializing",
	}
	currentDownload = task
	downloadMutex.Unlock()

	// 立即广播初始状态
	broadcastProgress()

	fmt.Printf("开始下载任务: URL=%s, SavePath=%s, TitleRegex=%s\n", task.URL, task.SavePath, task.TitleRegex)

	// 创建保存目录
	savePath := filepath.Join("audiobooks", task.SavePath)
	if err := os.MkdirAll(savePath, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %v", err)
	}
	fmt.Printf("创建目录成功: %s\n", savePath)

	// 扫描已存在的音频文件并预加载到完成列表
	fmt.Println("扫描已存在的音频文件...")
	existingFiles := scanExistingAudioFiles(savePath)
	if len(existingFiles) > 0 {
		downloadMutex.Lock()
		// 只保留最近3个文件用于显示
		if len(existingFiles) > 3 {
			currentDownload.Progress.CompletedFiles = existingFiles[len(existingFiles)-3:]
		} else {
			currentDownload.Progress.CompletedFiles = existingFiles
		}
		currentDownload.Progress.Status = fmt.Sprintf("发现 %d 个已存在的音频文件", len(existingFiles))
		downloadMutex.Unlock()

		// 广播更新状态
		broadcastProgress()
		fmt.Printf("预加载了 %d 个已存在的音频文件到完成列表\n", len(currentDownload.Progress.CompletedFiles))
	}

	// 构建yt-dlp命令
	cmd := buildYtDlpCommand(task, false) // 初始下载不使用continue
	task.Cmd = cmd                        // 保存命令引用

	// 设置环境变量确保UTF-8编码
	cmd.Env = append(os.Environ(),
		"PYTHONIOENCODING=utf-8",
		"LC_ALL=en_US.UTF-8",
		"LANG=en_US.UTF-8",
	)

	// 显示执行的命令
	fmt.Printf("执行命令: %s\n", formatCommand(cmd))

	// 创建输出管道
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("创建stdout管道失败: %v", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("创建stderr管道失败: %v", err)
	}

	// 启动命令
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("启动yt-dlp失败: %v", err)
	}

	fmt.Printf("yt-dlp已启动，PID: %d\n", cmd.Process.Pid)

	// 启动输出解析goroutine
	go parseOutput(stdout, "stdout")
	go parseOutput(stderr, "stderr")

	// 启动状态监控
	go monitorDownload(ctx)

	// 等待命令完成
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	// 设置超时
	timeout := time.After(10 * time.Minute)

	select {
	case err := <-done:
		fmt.Printf("yt-dlp进程结束: %v\n", err)
		return err
	case <-timeout:
		fmt.Println("下载超时，终止进程")
		cmd.Process.Kill()
		return fmt.Errorf("下载超时")
	case <-ctx.Done():
		fmt.Println("下载被取消，终止进程")
		cmd.Process.Kill()
		return fmt.Errorf("下载被取消")
	}
}

// 构建yt-dlp命令
func buildYtDlpCommand(task *DownloadTask, isContinue bool) *exec.Cmd {
	ytDlpPath := getYtDlpPath()

	// 基础参数
	quality := "bestaudio/best"
	if task.Quality != "" {
		quality = task.Quality
	}

	retryCount := "5"
	if task.RetryCount > 0 {
		retryCount = fmt.Sprintf("%d", task.RetryCount)
	}

	args := []string{
		"-f", quality,
		"-P", filepath.Join("audiobooks", task.SavePath),
		"--extractor-retries", retryCount,
		"--newline",
		"--progress-template", "download:%(progress._percent_str)s %(progress._speed_str)s",
		"--no-warnings",
	}

	// 添加缩略图选项
	if task.WriteThumbnail {
		args = append(args, "--write-thumbnail")
	}

	// 添加继续下载选项
	if isContinue {
		args = append(args, "--continue")
		fmt.Println("使用 --continue 选项继续下载")
	}

	// 添加cookies如果存在
	if hasCookies() {
		cookiesPath := getCookiesPath()
		args = append(args, "--cookies", cookiesPath)
		fmt.Printf("使用cookies文件: %s\n", cookiesPath)
	}

	// 添加输出格式
	outputFormat := "%(title)s.%(ext)s"
	if task.TitleRegex != "" {
		outputFormat = task.TitleRegex
	}
	args = append(args, "-o", outputFormat)

	// 添加URL
	args = append(args, task.URL)

	return exec.Command(ytDlpPath, args...)
}

// 格式化命令显示
func formatCommand(cmd *exec.Cmd) string {
	var parts []string
	parts = append(parts, cmd.Path)

	for _, arg := range cmd.Args[1:] {
		if strings.Contains(arg, " ") || strings.Contains(arg, "%") {
			parts = append(parts, fmt.Sprintf(`"%s"`, arg))
		} else {
			parts = append(parts, arg)
		}
	}

	return strings.Join(parts, " ")
}

// 解析输出
func parseOutput(pipe io.Reader, source string) {
	// 确保UTF-8编码处理
	scanner := bufio.NewScanner(pipe)
	scanner.Split(bufio.ScanLines)

	// 设置更大的缓冲区以处理长行
	const maxCapacity = 1024 * 1024 // 1MB
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)

	lineCount := 0

	// 进度解析正则表达式 - 更全面的匹配
	progressRegex := regexp.MustCompile(`download:(\d+\.?\d*)%\s+(.*)`)
	progressRegex2 := regexp.MustCompile(`\[download\]\s+(\d+\.?\d*)%\s+of\s+([^\s]+)\s+at\s+([^\s]+)\s+ETA\s+([^\s]+)`)
	progressRegex3 := regexp.MustCompile(`\[download\]\s+(\d+\.?\d*)%\s+of\s+([^\s]+)\s+at\s+([^\s]+)`)
	// 新增：匹配简单的进度格式（如 "  30.1%  604.19KiB/s"）
	progressRegex4 := regexp.MustCompile(`^\s*(\d+\.?\d*)%\s+(.+)$`)

	// 项目和播放列表信息
	itemRegex := regexp.MustCompile(`\[download\] Downloading item (\d+) of (\d+)`)
	playlistRegex := regexp.MustCompile(`\[.*\] Playlist .*: Downloading (\d+) items`)
	playlistTitleRegex := regexp.MustCompile(`\[.*\] (.+): Downloading (\d+) items`)

	// 提取和处理信息
	extractingRegex := regexp.MustCompile(`\[.*\] Extracting URL: (.*)`)
	extractorRegex := regexp.MustCompile(`\[(.+)\] (.+): (.*)`)

	// 文件信息
	fileInfoRegex := regexp.MustCompile(`\[info\] (.+): Downloading \d+ format\(s\)`)
	titleRegex := regexp.MustCompile(`\[info\] (.+): (.*)`)
	downloadingRegex := regexp.MustCompile(`\[download\] (.+\.(m4a|mp3|opus|webm)) has already been downloaded`)
	destinationRegex := regexp.MustCompile(`\[download\] Destination: (.+)`)

	// 新增：更全面的文件完成检测
	mergeCompleteRegex := regexp.MustCompile(`\[Merger\] Merging formats into "(.+)"`)
	ffmpegCompleteRegex := regexp.MustCompile(`\[ffmpeg\] Merging formats into "(.+)"`)

	// 完成状态
	completedRegex := regexp.MustCompile(`^100\.0%\s+(.+)`)
	finishedRegex := regexp.MustCompile(`\[download\] 100% of .+ in \d+:\d+`)
	mergeRegex := regexp.MustCompile(`\[Merger\] Merging formats into "(.+)"`)

	// 错误和警告
	errorRegex := regexp.MustCompile(`ERROR: (.*)`)
	warningRegex := regexp.MustCompile(`WARNING: (.*)`)

	// 元数据信息
	metadataRegex := regexp.MustCompile(`\[info\] (.+): (.+)`)
	durationRegex := regexp.MustCompile(`duration:\s*(\d+:\d+:\d+|\d+:\d+)`)
	uploaderRegex := regexp.MustCompile(`uploader:\s*(.+)`)
	viewCountRegex := regexp.MustCompile(`view_count:\s*(\d+)`)

	for scanner.Scan() {
		line := scanner.Text()
		lineCount++

		// 清理和验证UTF-8字符串
		line = cleanUTF8String(line)

		fmt.Printf("[%s] 输出[%d]: %s\n", source, lineCount, line)

		downloadMutex.Lock()
		if currentDownload == nil || !currentDownload.IsRunning {
			downloadMutex.Unlock()
			break
		}

		// 解析进度 - 支持多种格式
		if match := progressRegex.FindStringSubmatch(line); len(match) > 2 {
			if fileProgress, err := strconv.ParseFloat(match[1], 64); err == nil {
				// 计算整体进度
				var overallProgress float64
				if currentDownload.Progress.TotalCount > 0 && currentDownload.Progress.CurrentIndex > 0 {
					// 播放列表进度计算：(已完成项目数 + 当前项目进度) / 总项目数 * 100
					completedItems := float64(currentDownload.Progress.CurrentIndex - 1) // 已完成的项目数
					currentItemProgress := fileProgress / 100.0                          // 当前项目进度 (0-1)
					overallProgress = (completedItems + currentItemProgress) / float64(currentDownload.Progress.TotalCount) * 100.0
				} else {
					// 单个文件下载，直接使用文件进度
					overallProgress = fileProgress
				}

				currentDownload.Progress.Progress = overallProgress
				currentDownload.Progress.FileProgress = fileProgress
				currentDownload.Progress.Speed = strings.TrimSpace(match[2])

				if currentDownload.Progress.TotalCount > 1 {
					currentDownload.Progress.Status = fmt.Sprintf("整体进度: %.1f%% (当前文件: %.1f%%)", overallProgress, fileProgress)
				} else {
					currentDownload.Progress.Status = fmt.Sprintf("下载中: %.1f%%", fileProgress)
				}
				currentDownload.Progress.Phase = "downloading"
			}
		} else if match := progressRegex4.FindStringSubmatch(line); len(match) > 2 {
			// 匹配简单的进度格式（如 "  30.1%  604.19KiB/s"）
			if fileProgress, err := strconv.ParseFloat(match[1], 64); err == nil {
				// 检查是否是已跳过文件，避免覆盖跳过状态
				isSkippedFile := currentDownload.Progress.Phase == "skipped"

				// 计算整体进度
				var overallProgress float64
				if currentDownload.Progress.TotalCount > 0 && currentDownload.Progress.CurrentIndex > 0 {
					// 播放列表进度计算：(已完成项目数 + 当前项目进度) / 总项目数 * 100
					completedItems := float64(currentDownload.Progress.CurrentIndex - 1) // 已完成的项目数
					currentItemProgress := fileProgress / 100.0                          // 当前项目进度 (0-1)
					overallProgress = (completedItems + currentItemProgress) / float64(currentDownload.Progress.TotalCount) * 100.0
				} else {
					// 单个文件下载，直接使用文件进度
					overallProgress = fileProgress
				}

				currentDownload.Progress.Progress = overallProgress
				currentDownload.Progress.FileProgress = fileProgress

				// 只有不是跳过文件时才更新速度和状态
				if !isSkippedFile {
					currentDownload.Progress.Speed = strings.TrimSpace(match[2])
					if currentDownload.Progress.TotalCount > 1 {
						currentDownload.Progress.Status = fmt.Sprintf("整体进度: %.1f%% (当前文件: %.1f%%)", overallProgress, fileProgress)
					} else {
						currentDownload.Progress.Status = fmt.Sprintf("下载中: %.1f%%", fileProgress)
					}
					currentDownload.Progress.Phase = "downloading"
				}

				fmt.Printf("解析进度: 文件进度=%.1f%%, 整体进度=%.1f%%, 速度=%s, 跳过=%v\n", fileProgress, overallProgress, strings.TrimSpace(match[2]), isSkippedFile)
			}
		} else if match := progressRegex2.FindStringSubmatch(line); len(match) > 4 {
			// [download] 50.0% of 10.5MiB at 1.2MiB/s ETA 00:05
			if fileProgress, err := strconv.ParseFloat(match[1], 64); err == nil {
				// 计算整体进度
				var overallProgress float64
				if currentDownload.Progress.TotalCount > 0 && currentDownload.Progress.CurrentIndex > 0 {
					completedItems := float64(currentDownload.Progress.CurrentIndex - 1)
					currentItemProgress := fileProgress / 100.0
					overallProgress = (completedItems + currentItemProgress) / float64(currentDownload.Progress.TotalCount) * 100.0
				} else {
					overallProgress = fileProgress
				}

				currentDownload.Progress.Progress = overallProgress
				currentDownload.Progress.FileProgress = fileProgress
				currentDownload.Progress.Speed = strings.TrimSpace(match[3])
				currentDownload.Progress.ETA = strings.TrimSpace(match[4])
				currentDownload.Progress.FileSize = strings.TrimSpace(match[2])

				if currentDownload.Progress.TotalCount > 1 {
					currentDownload.Progress.Status = fmt.Sprintf("整体进度: %.1f%% (当前文件: %.1f%%, 大小: %s)", overallProgress, fileProgress, match[2])
				} else {
					currentDownload.Progress.Status = fmt.Sprintf("下载中: %.1f%% (大小: %s)", fileProgress, match[2])
				}
				currentDownload.Progress.Phase = "downloading"
			}
		} else if match := progressRegex3.FindStringSubmatch(line); len(match) > 3 {
			// [download] 50.0% of 10.5MiB at 1.2MiB/s
			if fileProgress, err := strconv.ParseFloat(match[1], 64); err == nil {
				// 计算整体进度
				var overallProgress float64
				if currentDownload.Progress.TotalCount > 0 && currentDownload.Progress.CurrentIndex > 0 {
					completedItems := float64(currentDownload.Progress.CurrentIndex - 1)
					currentItemProgress := fileProgress / 100.0
					overallProgress = (completedItems + currentItemProgress) / float64(currentDownload.Progress.TotalCount) * 100.0
				} else {
					overallProgress = fileProgress
				}

				currentDownload.Progress.Progress = overallProgress
				currentDownload.Progress.FileProgress = fileProgress
				currentDownload.Progress.Speed = strings.TrimSpace(match[3])
				currentDownload.Progress.FileSize = strings.TrimSpace(match[2])

				if currentDownload.Progress.TotalCount > 1 {
					currentDownload.Progress.Status = fmt.Sprintf("整体进度: %.1f%% (当前文件: %.1f%%, 大小: %s)", overallProgress, fileProgress, match[2])
				} else {
					currentDownload.Progress.Status = fmt.Sprintf("下载中: %.1f%% (大小: %s)", fileProgress, match[2])
				}
				currentDownload.Progress.Phase = "downloading"
			}
		}

		// 解析播放列表进度
		if match := itemRegex.FindStringSubmatch(line); len(match) > 2 {
			if current, err := strconv.Atoi(match[1]); err == nil {
				if total, err := strconv.Atoi(match[2]); err == nil {
					currentDownload.Progress.CurrentIndex = current
					currentDownload.Progress.TotalCount = total
					currentDownload.Progress.Status = fmt.Sprintf("下载项目: %d/%d", current, total)
				}
			}
		}

		// 解析播放列表标题和数量
		if match := playlistTitleRegex.FindStringSubmatch(line); len(match) > 2 {
			if total, err := strconv.Atoi(match[2]); err == nil {
				currentDownload.Progress.TotalCount = total
				currentDownload.Progress.PlaylistTitle = match[1]
				currentDownload.Progress.Status = fmt.Sprintf("播放列表: %s (%d项)", match[1], total)
				currentDownload.Progress.Phase = "extracting"
			}
		}

		// 解析提取信息
		if match := extractingRegex.FindStringSubmatch(line); len(match) > 1 {
			currentDownload.Progress.Status = "正在提取视频信息..."
			currentDownload.Progress.LastActivity = fmt.Sprintf("提取: %s", match[1])
			currentDownload.Progress.Phase = "extracting"
		}

		// 解析提取器信息
		if match := extractorRegex.FindStringSubmatch(line); len(match) > 3 {
			currentDownload.Progress.LastActivity = fmt.Sprintf("[%s] %s: %s", match[1], match[2], match[3])
			currentDownload.Progress.Phase = "extracting"
		}

		// 解析播放列表信息
		if match := playlistRegex.FindStringSubmatch(line); len(match) > 1 {
			if total, err := strconv.Atoi(match[1]); err == nil {
				currentDownload.Progress.TotalCount = total
				currentDownload.Progress.Status = fmt.Sprintf("发现播放列表，共%d个项目", total)
			}
		}

		// 解析已完成的文件 - 多种检测方式（避免重复）
		var completedFileName string
		var detectionMethod string

		// 方式1: 检测 "has already been downloaded" (最可靠)
		if match := downloadingRegex.FindStringSubmatch(line); len(match) > 1 {
			completedFileName = match[1]
			detectionMethod = "already_downloaded"

			// 对于已下载的文件，设置特殊的状态和速度，并标记为跳过状态
			currentDownload.Progress.Speed = "已跳过"
			currentDownload.Progress.Status = fmt.Sprintf("跳过已下载文件: %s", filepath.Base(match[1]))
			currentDownload.Progress.Phase = "skipped" // 添加跳过状态标记
		}

		// 方式2: 检测合并完成
		if completedFileName == "" {
			if match := mergeCompleteRegex.FindStringSubmatch(line); len(match) > 1 {
				completedFileName = match[1]
				detectionMethod = "merge_complete"
			}
		}

		// 方式3: 检测ffmpeg合并完成
		if completedFileName == "" {
			if match := ffmpegCompleteRegex.FindStringSubmatch(line); len(match) > 1 {
				completedFileName = match[1]
				detectionMethod = "ffmpeg_complete"
			}
		}

		// 方式4: 已移除100%完成状态检测，避免与already_downloaded重复
		// 注意：100%进度检测已移至下方的独立进度更新逻辑中

		// 方式5: 检测播放列表完成时，只处理停止逻辑，不重复添加文件
		if strings.Contains(line, "[download] Finished downloading playlist") {
			// 标记下载完成
			currentDownload.Progress.Progress = 100.0
			currentDownload.Progress.Status = "下载完成"
			currentDownload.Progress.Phase = "completed"
			currentDownload.Progress.IsDownloading = false
			currentDownload.IsRunning = false
			fmt.Println("播放列表下载完成，自动停止下载任务")
			// 不在这里添加文件，避免重复
		}

		// 处理完成的文件名
		if completedFileName != "" {
			// 提取文件名（去掉路径）
			fileName := completedFileName
			if idx := strings.LastIndex(fileName, "\\"); idx != -1 {
				fileName = fileName[idx+1:]
			}
			if idx := strings.LastIndex(fileName, "/"); idx != -1 {
				fileName = fileName[idx+1:]
			}

			// 清理文件名，确保UTF-8编码正确
			fileName = cleanUTF8String(fileName)

			// 只处理音频文件，过滤掉图片文件
			if isAudioFile(fileName) {
				// 更严格的重复检测：检查文件名是否已存在
				isDuplicate := false
				for _, existingFile := range currentDownload.Progress.CompletedFiles {
					// 清理现有文件名并进行比较
					cleanExistingFile := cleanUTF8String(existingFile)
					if strings.EqualFold(cleanExistingFile, fileName) { // 不区分大小写比较
						isDuplicate = true
						break
					}
				}

				if !isDuplicate {
					currentDownload.Progress.CompletedFiles = append(currentDownload.Progress.CompletedFiles, fileName)
					// 只保留最近3个
					if len(currentDownload.Progress.CompletedFiles) > 3 {
						currentDownload.Progress.CompletedFiles = currentDownload.Progress.CompletedFiles[len(currentDownload.Progress.CompletedFiles)-3:]
					}
					fmt.Printf("添加完成音频文件: %s (方法: %s, 总数: %d)\n", fileName, detectionMethod, len(currentDownload.Progress.CompletedFiles))
				} else {
					fmt.Printf("跳过重复文件: %s (方法: %s)\n", fileName, detectionMethod)
				}
			} else {
				fmt.Printf("跳过非音频文件: %s (方法: %s)\n", fileName, detectionMethod)
			}
		}

		// 解析完成进度 - 检测100%完成状态
		if match := completedRegex.FindStringSubmatch(line); len(match) > 1 {
			fileProgress := 100.0 // 因为正则表达式已经匹配了100.0%

			// 检查是否是已跳过文件的100%输出（避免覆盖"已跳过"状态）
			isSkippedFile := currentDownload.Progress.Phase == "skipped"

			// 计算整体进度
			var overallProgress float64
			if currentDownload.Progress.TotalCount > 0 && currentDownload.Progress.CurrentIndex > 0 {
				// 播放列表进度计算：(已完成项目数 + 当前项目进度) / 总项目数 * 100
				completedItems := float64(currentDownload.Progress.CurrentIndex - 1) // 已完成的项目数
				currentItemProgress := fileProgress / 100.0                          // 当前项目进度 (0-1)
				overallProgress = (completedItems + currentItemProgress) / float64(currentDownload.Progress.TotalCount) * 100.0
			} else {
				// 单个文件下载，直接使用文件进度
				overallProgress = fileProgress
			}

			currentDownload.Progress.Progress = overallProgress
			currentDownload.Progress.FileProgress = fileProgress // 保存单个文件进度

			// 只有不是跳过文件时才更新状态和速度
			if !isSkippedFile {
				currentDownload.Progress.Status = "文件下载完成"
				currentDownload.Progress.Phase = "downloading"
			}

			// 当文件进度达到100%时，添加当前文件到已完成列表
			if currentDownload.Progress.CurrentFile != "" {
				fileName := currentDownload.Progress.CurrentFile
				// 提取文件名（去掉路径）
				if idx := strings.LastIndex(fileName, "\\"); idx != -1 {
					fileName = fileName[idx+1:]
				}
				if idx := strings.LastIndex(fileName, "/"); idx != -1 {
					fileName = fileName[idx+1:]
				}

				// 清理文件名，确保UTF-8编码正确
				fileName = cleanUTF8String(fileName)

				// 只处理音频文件
				if isAudioFile(fileName) {
					// 更严格的重复检测：检查文件名是否已存在
					isDuplicate := false
					for _, existingFile := range currentDownload.Progress.CompletedFiles {
						// 清理现有文件名并进行比较
						cleanExistingFile := cleanUTF8String(existingFile)
						if strings.EqualFold(cleanExistingFile, fileName) { // 不区分大小写比较
							isDuplicate = true
							break
						}
					}

					if !isDuplicate {
						currentDownload.Progress.CompletedFiles = append(currentDownload.Progress.CompletedFiles, fileName)
						// 只保留最近3个
						if len(currentDownload.Progress.CompletedFiles) > 3 {
							currentDownload.Progress.CompletedFiles = currentDownload.Progress.CompletedFiles[len(currentDownload.Progress.CompletedFiles)-3:]
						}
						fmt.Printf("添加完成音频文件: %s (方法: 100%%_progress, 总数: %d)\n", fileName, len(currentDownload.Progress.CompletedFiles))
					} else {
						fmt.Printf("跳过重复文件: %s (方法: 100%%_progress)\n", fileName)
					}
				}
			}
		}

		// 解析下载完成
		if match := finishedRegex.FindStringSubmatch(line); len(match) > 0 {
			currentDownload.Progress.Progress = 100.0
			currentDownload.Progress.Status = "下载完成"
			currentDownload.Progress.Phase = "completed"
		}

		// 解析合并信息
		if match := mergeRegex.FindStringSubmatch(line); len(match) > 1 {
			currentDownload.Progress.Status = fmt.Sprintf("正在合并: %s", filepath.Base(match[1]))
			currentDownload.Progress.Phase = "merging"
		}

		// 解析目标文件路径
		if match := destinationRegex.FindStringSubmatch(line); len(match) > 1 {
			currentDownload.Progress.CurrentFile = match[1]
			fileName := filepath.Base(match[1])
			currentDownload.Progress.CurrentTitle = fileName // 更新当前标题为文件名
			currentDownload.Progress.Status = fmt.Sprintf("正在下载: %s", fileName)
			fmt.Printf("设置当前文件: %s\n", match[1])
		}

		// 解析文件信息
		if match := fileInfoRegex.FindStringSubmatch(line); len(match) > 1 {
			currentDownload.Progress.CurrentTitle = match[1]
			// 从当前项目索引构建文件名
			if currentDownload.Progress.CurrentIndex > 0 && currentDownload.Progress.PlaylistTitle != "" {
				// 构建预期的文件名格式
				expectedFileName := fmt.Sprintf("%s p%02d", currentDownload.Progress.PlaylistTitle, currentDownload.Progress.CurrentIndex)
				currentDownload.Progress.CurrentFile = expectedFileName + ".m4a"
				fmt.Printf("根据项目索引设置当前文件: %s\n", currentDownload.Progress.CurrentFile)
			} else {
				currentDownload.Progress.CurrentFile = match[1]
			}
			currentDownload.Progress.Status = fmt.Sprintf("准备下载: %s", match[1])
		}

		// 解析标题信息
		if match := titleRegex.FindStringSubmatch(line); len(match) > 2 {
			currentDownload.Progress.CurrentTitle = match[1]
			currentDownload.Progress.LastActivity = fmt.Sprintf("标题: %s - %s", match[1], match[2])
		}

		// 解析错误信息
		if match := errorRegex.FindStringSubmatch(line); len(match) > 1 {
			currentDownload.Progress.Status = fmt.Sprintf("错误: %s", match[1])
			currentDownload.Progress.ErrorMessage = match[1]
			currentDownload.Progress.LastActivity = fmt.Sprintf("ERROR: %s", match[1])
			currentDownload.Progress.Phase = "error"
		}

		// 解析警告信息
		if match := warningRegex.FindStringSubmatch(line); len(match) > 1 {
			currentDownload.Progress.WarningMessage = match[1]
			currentDownload.Progress.LastActivity = fmt.Sprintf("WARNING: %s", match[1])
		}

		// 解析元数据信息
		if match := metadataRegex.FindStringSubmatch(line); len(match) > 2 {
			key := strings.ToLower(match[1])
			value := match[2]

			switch key {
			case "duration":
				if match := durationRegex.FindStringSubmatch(value); len(match) > 1 {
					currentDownload.Progress.Duration = match[1]
					currentDownload.Progress.LastActivity = fmt.Sprintf("时长: %s", match[1])
				}
			case "uploader":
				if match := uploaderRegex.FindStringSubmatch(value); len(match) > 1 {
					currentDownload.Progress.Uploader = match[1]
					currentDownload.Progress.LastActivity = fmt.Sprintf("上传者: %s", match[1])
				}
			case "view_count":
				if match := viewCountRegex.FindStringSubmatch(value); len(match) > 1 {
					currentDownload.Progress.ViewCount = match[1]
					currentDownload.Progress.LastActivity = fmt.Sprintf("播放量: %s", match[1])
				}
			}
		}

		// 更新最后活动（如果没有其他更具体的信息）
		if currentDownload.Progress.LastActivity == "" ||
			!strings.Contains(currentDownload.Progress.LastActivity, "ERROR") &&
				!strings.Contains(currentDownload.Progress.LastActivity, "WARNING") &&
				!strings.Contains(currentDownload.Progress.LastActivity, "时长") &&
				!strings.Contains(currentDownload.Progress.LastActivity, "上传者") &&
				!strings.Contains(currentDownload.Progress.LastActivity, "播放量") {
			currentDownload.Progress.LastActivity = line
		}

		downloadMutex.Unlock()

		// 广播进度
		broadcastProgress()
	}

	fmt.Printf("[%s] 输出解析结束，共处理%d行\n", source, lineCount)
}

// 监控下载状态
func monitorDownload(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			downloadMutex.RLock()
			if currentDownload != nil && currentDownload.IsRunning {
				progress := currentDownload.Progress
				if progress != nil {
					fmt.Printf("=== 下载状态 ===\n")
					fmt.Printf("状态: %s\n", progress.Status)
					fmt.Printf("进度: %.1f%%\n", progress.Progress)
					fmt.Printf("速度: %s\n", progress.Speed)
					fmt.Printf("项目: %d/%d\n", progress.CurrentIndex, progress.TotalCount)
					fmt.Printf("===============\n")
				}
			}
			downloadMutex.RUnlock()
		}
	}
}

// 获取当前下载进度
func getCurrentProgress() *DownloadProgress {
	downloadMutex.RLock()
	defer downloadMutex.RUnlock()

	if currentDownload == nil {
		return &DownloadProgress{IsDownloading: false}
	}

	// 返回进度的副本
	progress := *currentDownload.Progress
	return &progress
}

// 停止当前下载
func stopCurrentDownload() {
	downloadMutex.Lock()
	defer downloadMutex.Unlock()

	if currentDownload != nil {
		// 如果有正在运行的任务，先取消
		if currentDownload.IsRunning {
			if currentDownload.Cancel != nil {
				currentDownload.Cancel()
			}

			// 如果有进程在运行，终止它
			if currentDownload.Cmd != nil && currentDownload.Cmd.Process != nil {
				fmt.Printf("停止下载，终止进程 PID: %d\n", currentDownload.Cmd.Process.Pid)
				currentDownload.Cmd.Process.Kill()
			}

			// 添加停止记录到历史
			addTaskToHistory(currentDownload, "stopped", "用户手动停止")
		}

		// 更新状态
		currentDownload.IsRunning = false
		if currentDownload.Progress != nil {
			currentDownload.Progress.IsDownloading = false
			currentDownload.Progress.IsPaused = false
			currentDownload.Progress.Status = "下载已停止"
			currentDownload.Progress.Phase = "stopped"
			currentDownload.Progress.LastActivity = "用户停止下载"
		}

		fmt.Println("下载任务已停止")

		// 广播状态更新
		go broadcastProgress()

		// 延迟清空当前下载任务，给前端时间接收状态更新
		go func() {
			time.Sleep(2 * time.Second)
			downloadMutex.Lock()
			currentDownload = nil
			downloadMutex.Unlock()
			fmt.Println("下载任务已清理")
		}()
	}
}

// 检查是否有正在进行的下载
func hasActiveDownload() bool {
	downloadMutex.RLock()
	defer downloadMutex.RUnlock()

	return currentDownload != nil && currentDownload.IsRunning
}

// 暂停当前下载 - 通过发送中断信号优雅地终止yt-dlp进程
func pauseCurrentDownload() bool {
	downloadMutex.Lock()
	defer downloadMutex.Unlock()

	if currentDownload == nil || !currentDownload.IsRunning {
		fmt.Println("没有正在进行的下载任务")
		return false
	}

	if currentDownload.Progress.IsPaused {
		fmt.Println("下载已经暂停")
		return false
	}

	// 发送中断信号给yt-dlp进程
	if currentDownload.Cmd != nil && currentDownload.Cmd.Process != nil {
		fmt.Printf("暂停下载，发送中断信号给进程 PID: %d\n", currentDownload.Cmd.Process.Pid)

		// 在Windows上使用Kill，在Unix系统上使用Interrupt
		if runtime.GOOS == "windows" {
			err := currentDownload.Cmd.Process.Kill()
			if err != nil {
				fmt.Printf("暂停下载失败: %v\n", err)
				return false
			}
		} else {
			err := currentDownload.Cmd.Process.Signal(os.Interrupt)
			if err != nil {
				fmt.Printf("暂停下载失败: %v\n", err)
				return false
			}
		}

		// 更新状态
		currentDownload.Progress.IsPaused = true
		currentDownload.Progress.IsDownloading = false
		currentDownload.Progress.Status = "下载已暂停"
		currentDownload.Progress.Phase = "paused"
		currentDownload.Progress.LastActivity = "用户暂停下载"
		currentDownload.IsRunning = false

		fmt.Println("下载已暂停，可以使用继续功能恢复下载")
		go broadcastProgress()
		return true
	}

	fmt.Println("无法找到下载进程")
	return false
}

// 继续当前下载 - 使用--continue选项重新启动yt-dlp
func resumeCurrentDownload() bool {
	downloadMutex.Lock()
	defer downloadMutex.Unlock()

	if currentDownload == nil {
		fmt.Println("没有下载任务")
		return false
	}

	if !currentDownload.Progress.IsPaused {
		fmt.Println("下载未暂停")
		return false
	}

	fmt.Println("继续下载，使用 --continue 选项")

	// 重新构建命令，使用--continue选项
	cmd := buildYtDlpCommand(currentDownload, true)
	currentDownload.Cmd = cmd

	// 设置环境变量
	cmd.Env = append(os.Environ(),
		"PYTHONIOENCODING=utf-8",
		"LC_ALL=en_US.UTF-8",
		"LANG=en_US.UTF-8",
	)

	fmt.Printf("继续执行命令: %s\n", formatCommand(cmd))

	// 创建输出管道
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Printf("创建stdout管道失败: %v\n", err)
		return false
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		fmt.Printf("创建stderr管道失败: %v\n", err)
		return false
	}

	// 启动命令
	if err := cmd.Start(); err != nil {
		fmt.Printf("继续下载失败: %v\n", err)
		return false
	}

	fmt.Printf("yt-dlp已重新启动，PID: %d\n", cmd.Process.Pid)

	// 更新状态
	currentDownload.Progress.IsPaused = false
	currentDownload.Progress.IsDownloading = true
	currentDownload.Progress.Status = "继续下载中..."
	currentDownload.Progress.Phase = "downloading"
	currentDownload.Progress.LastActivity = "用户继续下载"
	currentDownload.IsRunning = true

	// 重新创建context
	ctx, cancel := context.WithCancel(context.Background())
	currentDownload.Cancel = cancel

	// 启动输出解析goroutine
	go parseOutput(stdout, "stdout")
	go parseOutput(stderr, "stderr")

	// 启动状态监控
	go monitorDownload(ctx)

	// 在新的goroutine中等待命令完成
	go func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("下载goroutine panic: %v\n", r)
			}
		}()

		// 等待命令完成
		done := make(chan error, 1)
		go func() {
			defer func() {
				if r := recover(); r != nil {
					fmt.Printf("命令等待goroutine panic: %v\n", r)
					done <- fmt.Errorf("命令执行异常: %v", r)
				}
			}()
			done <- cmd.Wait()
		}()

		// 设置超时
		timeout := time.After(10 * time.Minute)

		select {
		case err := <-done:
			if err != nil {
				fmt.Printf("yt-dlp进程结束: %v\n", err)
				downloadMutex.Lock()
				if currentDownload != nil && currentDownload.Progress != nil {
					currentDownload.Progress.Status = "下载失败"
					currentDownload.Progress.Phase = "error"
					currentDownload.Progress.ErrorMessage = err.Error()
					currentDownload.IsRunning = false
				}
				downloadMutex.Unlock()
				broadcastProgress()
			} else {
				fmt.Println("yt-dlp进程正常结束")
			}
		case <-timeout:
			fmt.Println("下载超时，终止进程")
			if cmd.Process != nil {
				cmd.Process.Kill()
			}
		case <-ctx.Done():
			fmt.Println("下载被取消，终止进程")
			if cmd.Process != nil {
				cmd.Process.Kill()
			}
		}
	}()

	go broadcastProgress()
	return true
}

// 任务历史结构
type TaskHistory struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Status    string `json:"status"`
	URL       string `json:"url"`
	CreatedAt string `json:"created_at"`
	FileSize  string `json:"file_size,omitempty"`
	Duration  string `json:"duration,omitempty"`
	Progress  int    `json:"progress,omitempty"`
	Error     string `json:"error,omitempty"`
}

// 历史记录存储
var (
	taskHistoryMutex sync.RWMutex
	taskHistoryList  []TaskHistory
	nextTaskID       int = 1
)

// 初始化历史记录，扫描已存在的文件
func initializeHistory() {
	fmt.Println("初始化历史记录，扫描 audiobooks 目录...")

	// 扫描 audiobooks 目录下的所有子目录
	audiobooksDir := "audiobooks"
	if _, err := os.Stat(audiobooksDir); os.IsNotExist(err) {
		fmt.Printf("audiobooks 目录不存在: %s\n", audiobooksDir)
		return
	}

	// 遍历 audiobooks 下的所有子目录
	entries, err := os.ReadDir(audiobooksDir)
	if err != nil {
		fmt.Printf("读取 audiobooks 目录失败: %v\n", err)
		return
	}

	taskHistoryMutex.Lock()
	defer taskHistoryMutex.Unlock()

	for _, entry := range entries {
		if entry.IsDir() {
			subDir := filepath.Join(audiobooksDir, entry.Name())
			audioFiles := scanExistingAudioFiles(subDir)

			if len(audioFiles) > 0 {
				// 为每个有音频文件的目录创建一个历史记录
				historyItem := TaskHistory{
					ID:        nextTaskID,
					Title:     fmt.Sprintf("已完成：%s", entry.Name()),
					Status:    "completed",
					URL:       fmt.Sprintf("本地目录: %s", subDir),
					CreatedAt: "2025/01/20 00:00:00", // 使用固定时间，表示历史文件
					FileSize:  fmt.Sprintf("%d个文件", len(audioFiles)),
					Progress:  100,
				}

				taskHistoryList = append(taskHistoryList, historyItem)
				nextTaskID++

				fmt.Printf("添加历史记录: %s (%d个音频文件)\n", entry.Name(), len(audioFiles))
			}
		}
	}

	fmt.Printf("历史记录初始化完成，共发现 %d 个已完成的目录\n", len(taskHistoryList))
}

// 添加任务到历史记录
func addTaskToHistory(task *DownloadTask, status string, errorMsg string) {
	taskHistoryMutex.Lock()
	defer taskHistoryMutex.Unlock()

	// 获取任务标题
	title := "未知任务"
	if task.Progress != nil {
		if task.Progress.PlaylistTitle != "" {
			title = task.Progress.PlaylistTitle
		} else if task.Progress.CurrentTitle != "" {
			title = task.Progress.CurrentTitle
		}
	}

	// 计算进度
	progress := 0
	if task.Progress != nil {
		progress = int(task.Progress.Progress)
	}

	// 创建历史记录
	historyItem := TaskHistory{
		ID:        nextTaskID,
		Title:     title,
		Status:    status,
		URL:       task.URL,
		CreatedAt: time.Now().Format("2006-01-02 15:04:05"),
		Progress:  progress,
		Error:     errorMsg,
	}

	// 添加文件大小和时长信息
	if task.Progress != nil {
		historyItem.FileSize = task.Progress.FileSize
		historyItem.Duration = task.Progress.Duration
	}

	// 添加到历史列表
	taskHistoryList = append(taskHistoryList, historyItem)
	nextTaskID++

	// 只保留最近50个记录
	if len(taskHistoryList) > 50 {
		taskHistoryList = taskHistoryList[len(taskHistoryList)-50:]
	}

	fmt.Printf("添加任务到历史记录: ID=%d, Title=%s, Status=%s\n", historyItem.ID, historyItem.Title, historyItem.Status)
}

// 获取任务历史
func getTaskHistory() []TaskHistory {
	taskHistoryMutex.RLock()
	defer taskHistoryMutex.RUnlock()

	// 返回真实历史记录，按时间倒序排列
	result := make([]TaskHistory, len(taskHistoryList))
	copy(result, taskHistoryList)

	// 简单的倒序排列
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	return result
}

// WebSocket连接管理
func addWebSocketConnection(conn *websocket.Conn) {
	wsConnMutex.Lock()
	defer wsConnMutex.Unlock()
	wsConnections = append(wsConnections, conn)
}

func removeWebSocketConnection(conn *websocket.Conn) {
	wsConnMutex.Lock()
	defer wsConnMutex.Unlock()

	for i, c := range wsConnections {
		if c == conn {
			wsConnections = slices.Delete(wsConnections, i, i+1)
			break
		}
	}
}

// 广播进度更新
func broadcastProgress() {
	wsConnMutex.Lock()
	defer wsConnMutex.Unlock()

	progress := getCurrentProgress()
	message, err := json.Marshal(map[string]any{
		"type": "progress",
		"data": progress,
	})
	if err != nil {
		fmt.Printf("序列化进度数据失败: %v\n", err)
		return
	}

	fmt.Printf("广播进度更新: 连接数=%d, 状态=%+v\n", len(wsConnections), progress)

	// 使用倒序遍历，方便删除失效连接
	for i := len(wsConnections) - 1; i >= 0; i-- {
		conn := wsConnections[i]
		err := conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			fmt.Printf("WebSocket连接[%d]发送失败: %v，移除连接\n", i, err)
			// 移除失效连接
			wsConnections = slices.Delete(wsConnections, i, i+1)
		} else {
			fmt.Printf("WebSocket连接[%d]发送成功\n", i)
		}
	}
}

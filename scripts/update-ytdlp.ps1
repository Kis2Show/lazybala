# yt-dlp 升级脚本 (PowerShell)

param(
    [switch]$Force = $false,
    [switch]$Restore = $false,
    [switch]$Version = $false,
    [switch]$Help = $false
)

# 配置
$BinDir = ".\data\bin"
$YtDlpPath = "$BinDir\yt-dlp.exe"
$BackupPath = "$BinDir\yt-dlp.backup.exe"

# 函数：打印带颜色的消息
function Write-Info {
    param([string]$Message)
    Write-Host "[INFO] $Message" -ForegroundColor Blue
}

function Write-Success {
    param([string]$Message)
    Write-Host "[SUCCESS] $Message" -ForegroundColor Green
}

function Write-Warning {
    param([string]$Message)
    Write-Host "[WARNING] $Message" -ForegroundColor Yellow
}

function Write-Error {
    param([string]$Message)
    Write-Host "[ERROR] $Message" -ForegroundColor Red
}

# 函数：检测架构
function Get-Architecture {
    $arch = $env:PROCESSOR_ARCHITECTURE
    switch ($arch) {
        "AMD64" { return "yt-dlp.exe" }
        "ARM64" { return "yt-dlp_arm64.exe" }
        default { return "yt-dlp.exe" }
    }
}

# 函数：获取当前版本
function Get-CurrentVersion {
    if (Test-Path $YtDlpPath) {
        try {
            $version = & $YtDlpPath --version 2>$null
            return $version
        }
        catch {
            return "unknown"
        }
    }
    else {
        return "not installed"
    }
}

# 函数：获取最新版本
function Get-LatestVersion {
    try {
        $response = Invoke-RestMethod -Uri "https://api.github.com/repos/yt-dlp/yt-dlp/releases/latest"
        return $response.tag_name
    }
    catch {
        Write-Error "无法获取最新版本信息: $($_.Exception.Message)"
        return $null
    }
}

# 函数：下载 yt-dlp
function Download-YtDlp {
    param([string]$Filename)
    
    $url = "https://github.com/yt-dlp/yt-dlp/releases/latest/download/$Filename"
    $tempFile = [System.IO.Path]::GetTempFileName()
    
    Write-Info "下载 $Filename..."
    
    try {
        # 下载文件
        Invoke-WebRequest -Uri $url -OutFile $tempFile -UseBasicParsing
        
        # 验证下载的文件
        if ((Get-Item $tempFile).Length -gt 0) {
            # 备份现有文件
            if (Test-Path $YtDlpPath) {
                Write-Info "备份现有版本到 $BackupPath"
                Copy-Item $YtDlpPath $BackupPath -Force
            }
            
            # 移动新文件
            Move-Item $tempFile $YtDlpPath -Force
            
            Write-Success "yt-dlp 下载完成"
            return $true
        }
        else {
            Write-Error "下载的文件为空"
            Remove-Item $tempFile -Force -ErrorAction SilentlyContinue
            return $false
        }
    }
    catch {
        Write-Error "下载失败: $($_.Exception.Message)"
        Remove-Item $tempFile -Force -ErrorAction SilentlyContinue
        return $false
    }
}

# 函数：验证安装
function Test-Installation {
    if (Test-Path $YtDlpPath) {
        try {
            $version = & $YtDlpPath --version 2>$null
            if ($LASTEXITCODE -eq 0) {
                Write-Success "yt-dlp 验证成功，版本: $version"
                return $true
            }
            else {
                Write-Error "yt-dlp 验证失败"
                return $false
            }
        }
        catch {
            Write-Error "yt-dlp 验证失败: $($_.Exception.Message)"
            return $false
        }
    }
    else {
        Write-Error "yt-dlp 文件不存在"
        return $false
    }
}

# 函数：恢复备份
function Restore-Backup {
    if (Test-Path $BackupPath) {
        Write-Warning "恢复备份版本..."
        Copy-Item $BackupPath $YtDlpPath -Force
        Write-Success "已恢复备份版本"
        return $true
    }
    else {
        Write-Error "没有找到备份文件"
        return $false
    }
}

# 函数：显示帮助
function Show-Help {
    Write-Host @"
LazyBala yt-dlp 升级工具

用法: .\update-ytdlp.ps1 [选项]

选项:
  -Force      强制重新下载
  -Restore    恢复备份版本
  -Version    显示当前版本
  -Help       显示帮助信息

示例:
  .\update-ytdlp.ps1           # 检查并升级到最新版本
  .\update-ytdlp.ps1 -Force    # 强制重新下载最新版本
  .\update-ytdlp.ps1 -Restore  # 恢复备份版本
  .\update-ytdlp.ps1 -Version  # 显示当前版本
"@ -ForegroundColor White
}

# 主函数
function Main {
    Write-Info "LazyBala yt-dlp 升级工具"
    Write-Host "================================"
    
    # 创建 bin 目录
    if (-not (Test-Path $BinDir)) {
        New-Item -ItemType Directory -Path $BinDir -Force | Out-Null
    }
    
    # 检测架构
    $filename = Get-Architecture
    Write-Info "检测到架构: $filename"
    
    # 获取版本信息
    $currentVersion = Get-CurrentVersion
    Write-Info "当前版本: $currentVersion"
    
    $latestVersion = Get-LatestVersion
    if (-not $latestVersion) {
        Write-Error "无法获取最新版本信息"
        exit 1
    }
    Write-Info "最新版本: $latestVersion"
    
    # 检查是否需要更新
    if ($currentVersion -eq $latestVersion -and -not $Force) {
        Write-Success "yt-dlp 已是最新版本"
        return
    }
    
    # 下载新版本
    if (Download-YtDlp $filename) {
        # 验证安装
        if (Test-Installation) {
            Write-Success "yt-dlp 升级成功！"
            Write-Info "从 $currentVersion 升级到 $latestVersion"
        }
        else {
            Write-Error "新版本验证失败，恢复备份"
            Restore-Backup
            exit 1
        }
    }
    else {
        Write-Error "下载失败"
        exit 1
    }
}

# 处理命令行参数
if ($Help) {
    Show-Help
    exit 0
}

if ($Version) {
    $currentVersion = Get-CurrentVersion
    Write-Host "当前版本: $currentVersion"
    exit 0
}

if ($Restore) {
    if (Restore-Backup) {
        Test-Installation
    }
    exit 0
}

if ($Force) {
    Write-Info "强制重新下载模式"
    if (Test-Path $YtDlpPath) {
        Remove-Item $YtDlpPath -Force
    }
}

# 运行主函数
Main

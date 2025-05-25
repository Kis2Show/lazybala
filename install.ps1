# LazyBala Windows 安装脚本
# 自动检测系统架构并下载对应的二进制文件

param(
    [string]$Version = "",
    [string]$InstallDir = "$env:LOCALAPPDATA\Programs\LazyBala",
    [switch]$Help
)

# 项目信息
$Repo = "kis2show/lazybala"
$BinaryName = "lazybala"

# 函数：打印彩色消息
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

# 函数：显示帮助信息
function Show-Help {
    Write-Host "LazyBala Windows 安装脚本" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "用法: .\install.ps1 [参数]" -ForegroundColor White
    Write-Host ""
    Write-Host "参数:" -ForegroundColor White
    Write-Host "  -Version <版本>      指定要安装的版本" -ForegroundColor Gray
    Write-Host "  -InstallDir <目录>   指定安装目录 (默认: $env:LOCALAPPDATA\Programs\LazyBala)" -ForegroundColor Gray
    Write-Host "  -Help               显示此帮助信息" -ForegroundColor Gray
    Write-Host ""
    Write-Host "示例:" -ForegroundColor White
    Write-Host "  .\install.ps1                           # 安装最新版本" -ForegroundColor Gray
    Write-Host "  .\install.ps1 -Version v1.0.0           # 安装指定版本" -ForegroundColor Gray
    Write-Host "  .\install.ps1 -InstallDir C:\Tools      # 安装到指定目录" -ForegroundColor Gray
    Write-Host ""
}

# 函数：检测系统架构
function Get-Platform {
    $arch = $env:PROCESSOR_ARCHITECTURE
    
    switch ($arch) {
        "AMD64" { return "windows-amd64" }
        "ARM64" { return "windows-arm64" }
        default {
            Write-Error "不支持的架构: $arch"
            exit 1
        }
    }
}

# 函数：获取最新版本
function Get-LatestVersion {
    Write-Info "获取最新版本信息..."
    
    try {
        $response = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases/latest"
        $latestVersion = $response.tag_name
        
        if (-not $latestVersion) {
            throw "无法获取版本信息"
        }
        
        Write-Info "最新版本: $latestVersion"
        return $latestVersion
    }
    catch {
        Write-Error "无法获取最新版本信息: $($_.Exception.Message)"
        exit 1
    }
}

# 函数：下载二进制文件
function Download-Binary {
    param(
        [string]$Version,
        [string]$Platform
    )
    
    $filename = "$BinaryName-$Platform.exe.zip"
    $url = "https://github.com/$Repo/releases/download/$Version/$filename"
    $tempDir = [System.IO.Path]::GetTempPath()
    $downloadPath = Join-Path $tempDir $filename
    $extractPath = Join-Path $tempDir "lazybala-extract"
    
    Write-Info "下载 $filename..."
    
    try {
        # 下载文件
        Invoke-WebRequest -Uri $url -OutFile $downloadPath -UseBasicParsing
        
        if (-not (Test-Path $downloadPath)) {
            throw "下载失败"
        }
        
        Write-Info "解压文件..."
        
        # 创建解压目录
        if (Test-Path $extractPath) {
            Remove-Item $extractPath -Recurse -Force
        }
        New-Item -ItemType Directory -Path $extractPath -Force | Out-Null
        
        # 解压文件
        Expand-Archive -Path $downloadPath -DestinationPath $extractPath -Force
        
        # 查找二进制文件
        $binaryFile = Get-ChildItem -Path $extractPath -Filter "*.exe" -Recurse | Select-Object -First 1
        
        if (-not $binaryFile) {
            throw "未找到二进制文件"
        }
        
        return $binaryFile.FullName
    }
    catch {
        Write-Error "下载失败: $($_.Exception.Message)"
        exit 1
    }
    finally {
        # 清理下载文件
        if (Test-Path $downloadPath) {
            Remove-Item $downloadPath -Force
        }
    }
}

# 函数：安装二进制文件
function Install-Binary {
    param(
        [string]$SourcePath,
        [string]$InstallDir
    )
    
    Write-Info "安装 $BinaryName 到 $InstallDir..."
    
    try {
        # 创建安装目录
        if (-not (Test-Path $InstallDir)) {
            New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
        }
        
        $targetPath = Join-Path $InstallDir "$BinaryName.exe"
        
        # 如果目标文件存在且正在运行，尝试停止
        if (Test-Path $targetPath) {
            $processes = Get-Process -Name $BinaryName -ErrorAction SilentlyContinue
            if ($processes) {
                Write-Warning "检测到 $BinaryName 正在运行，尝试停止..."
                $processes | Stop-Process -Force
                Start-Sleep -Seconds 2
            }
        }
        
        # 复制文件
        Copy-Item $SourcePath $targetPath -Force
        
        if (Test-Path $targetPath) {
            Write-Success "$BinaryName 安装成功!"
        } else {
            throw "安装失败"
        }
        
        return $targetPath
    }
    catch {
        Write-Error "安装失败: $($_.Exception.Message)"
        exit 1
    }
}

# 函数：添加到 PATH
function Add-ToPath {
    param([string]$InstallDir)
    
    Write-Info "检查 PATH 环境变量..."
    
    $currentPath = [Environment]::GetEnvironmentVariable("PATH", "User")
    
    if ($currentPath -notlike "*$InstallDir*") {
        Write-Info "添加安装目录到 PATH..."
        
        try {
            $newPath = "$currentPath;$InstallDir"
            [Environment]::SetEnvironmentVariable("PATH", $newPath, "User")
            Write-Success "已添加到 PATH 环境变量"
            Write-Warning "请重新启动 PowerShell 或命令提示符以使 PATH 生效"
        }
        catch {
            Write-Warning "无法自动添加到 PATH，请手动添加: $InstallDir"
        }
    } else {
        Write-Info "安装目录已在 PATH 中"
    }
}

# 函数：创建桌面快捷方式
function New-DesktopShortcut {
    param([string]$TargetPath)
    
    Write-Info "创建桌面快捷方式..."
    
    try {
        $shell = New-Object -ComObject WScript.Shell
        $shortcutPath = Join-Path $env:USERPROFILE "Desktop\LazyBala.lnk"
        $shortcut = $shell.CreateShortcut($shortcutPath)
        $shortcut.TargetPath = $TargetPath
        $shortcut.WorkingDirectory = Split-Path $TargetPath
        $shortcut.Description = "LazyBala 媒体下载工具"
        $shortcut.Save()
        
        Write-Success "桌面快捷方式创建成功"
    }
    catch {
        Write-Warning "无法创建桌面快捷方式: $($_.Exception.Message)"
    }
}

# 函数：验证安装
function Test-Installation {
    param([string]$TargetPath)
    
    Write-Info "验证安装..."
    
    if (Test-Path $TargetPath) {
        try {
            $versionOutput = & $TargetPath --version 2>$null
            if (-not $versionOutput) {
                $versionOutput = "LazyBala $Version"
            }
            Write-Success "安装验证成功: $versionOutput"
        }
        catch {
            Write-Success "二进制文件已安装: $TargetPath"
        }
    } else {
        Write-Error "安装验证失败"
        exit 1
    }
}

# 函数：显示使用说明
function Show-Usage {
    param([string]$TargetPath)
    
    Write-Host ""
    Write-Success "🎉 LazyBala 安装完成!"
    Write-Host ""
    Write-Host "使用方法:" -ForegroundColor White
    Write-Host "  $BinaryName                    # 启动服务器" -ForegroundColor Gray
    Write-Host "  $BinaryName --help             # 显示帮助信息" -ForegroundColor Gray
    Write-Host "  $BinaryName --version          # 显示版本信息" -ForegroundColor Gray
    Write-Host ""
    Write-Host "或者双击桌面快捷方式启动" -ForegroundColor Gray
    Write-Host ""
    Write-Host "Web 界面:" -ForegroundColor White
    Write-Host "  启动后访问: http://localhost:8080" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "文档链接:" -ForegroundColor White
    Write-Host "  GitHub: https://github.com/$Repo" -ForegroundColor Cyan
    Write-Host "  Issues: https://github.com/$Repo/issues" -ForegroundColor Cyan
    Write-Host ""
}

# 函数：清理临时文件
function Clear-TempFiles {
    $tempDir = [System.IO.Path]::GetTempPath()
    $extractPath = Join-Path $tempDir "lazybala-extract"
    
    if (Test-Path $extractPath) {
        Remove-Item $extractPath -Recurse -Force -ErrorAction SilentlyContinue
    }
}

# 主函数
function Main {
    if ($Help) {
        Show-Help
        return
    }
    
    Write-Host "🚀 LazyBala Windows 安装脚本" -ForegroundColor Cyan
    Write-Host "============================" -ForegroundColor Cyan
    Write-Host ""
    
    try {
        # 检测平台
        $platform = Get-Platform
        Write-Info "检测到平台: $platform"
        
        # 获取版本
        if (-not $Version) {
            $Version = Get-LatestVersion
        } else {
            Write-Info "使用指定版本: $Version"
        }
        
        # 下载二进制文件
        $binaryPath = Download-Binary -Version $Version -Platform $platform
        
        # 安装二进制文件
        $targetPath = Install-Binary -SourcePath $binaryPath -InstallDir $InstallDir
        
        # 添加到 PATH
        Add-ToPath -InstallDir $InstallDir
        
        # 创建桌面快捷方式
        New-DesktopShortcut -TargetPath $targetPath
        
        # 验证安装
        Test-Installation -TargetPath $targetPath
        
        # 显示使用说明
        Show-Usage -TargetPath $targetPath
    }
    catch {
        Write-Error "安装失败: $($_.Exception.Message)"
        exit 1
    }
    finally {
        # 清理临时文件
        Clear-TempFiles
    }
}

# 执行主函数
Main

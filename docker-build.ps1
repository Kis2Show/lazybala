# LazyBala Docker 构建脚本 (PowerShell)

param(
    [Parameter(Position=0)]
    [string]$Command = "help",
    
    [string]$Version = "",
    [string]$Platform = "linux/amd64,linux/arm64",
    [switch]$Push = $false,
    [switch]$NoCache = $false
)

# 设置变量
$APP_NAME = "lazybala"
$BUILDTIME = (Get-Date).ToUniversalTime().ToString("yyyy-MM-ddTHH:mm:ssZ")

# 获取 Git 版本
if ([string]::IsNullOrEmpty($Version)) {
    try {
        $Version = (git describe --tags --always --dirty 2>$null)
        if ([string]::IsNullOrEmpty($Version)) {
            $Version = "dev"
        }
    }
    catch {
        $Version = "dev"
    }
}

$DOCKER_IMAGE = "${APP_NAME}:${Version}"
$DOCKER_IMAGE_LATEST = "${APP_NAME}:latest"
$DOCKER_HUB_IMAGE = "kis2show/${APP_NAME}:${Version}"
$DOCKER_HUB_IMAGE_LATEST = "kis2show/${APP_NAME}:latest"

function Write-Header {
    Write-Host "========================================" -ForegroundColor Cyan
    Write-Host "LazyBala Docker 构建工具" -ForegroundColor Cyan
    Write-Host "========================================" -ForegroundColor Cyan
    Write-Host "应用名称: $APP_NAME" -ForegroundColor Green
    Write-Host "版本: $Version" -ForegroundColor Green
    Write-Host "构建时间: $BUILDTIME" -ForegroundColor Green
    Write-Host "Docker 镜像: $DOCKER_IMAGE" -ForegroundColor Green
    Write-Host "========================================" -ForegroundColor Cyan
}

function Build-Image {
    Write-Host "构建 Docker 镜像..." -ForegroundColor Yellow
    
    $buildArgs = @(
        "build"
        "--build-arg", "VERSION=$Version"
        "--build-arg", "BUILDTIME=$BUILDTIME"
        "-t", $DOCKER_IMAGE
        "-t", $DOCKER_IMAGE_LATEST
    )
    
    if ($NoCache) {
        $buildArgs += "--no-cache"
    }
    
    $buildArgs += "."
    
    & docker @buildArgs
    
    if ($LASTEXITCODE -ne 0) {
        Write-Error "构建失败！"
        exit 1
    }
    
    Write-Host "构建成功！" -ForegroundColor Green
}

function Build-MultiArch {
    Write-Host "构建多架构镜像..." -ForegroundColor Yellow
    
    $buildArgs = @(
        "buildx", "build"
        "--platform", $Platform
        "--build-arg", "VERSION=$Version"
        "--build-arg", "BUILDTIME=$BUILDTIME"
        "-t", $DOCKER_IMAGE
        "-t", $DOCKER_IMAGE_LATEST
    )
    
    if ($Push) {
        $buildArgs += "--push"
    }
    
    if ($NoCache) {
        $buildArgs += "--no-cache"
    }
    
    $buildArgs += "."
    
    & docker @buildArgs
    
    if ($LASTEXITCODE -ne 0) {
        Write-Error "多架构构建失败！"
        exit 1
    }
    
    Write-Host "多架构构建成功！" -ForegroundColor Green
}

function Start-Container {
    Write-Host "启动容器..." -ForegroundColor Yellow
    
    & docker-compose up -d
    
    if ($LASTEXITCODE -ne 0) {
        Write-Error "启动失败！"
        exit 1
    }
    
    Write-Host "容器已启动，访问 http://localhost:8080" -ForegroundColor Green
}

function Stop-Container {
    Write-Host "停止容器..." -ForegroundColor Yellow
    
    & docker-compose down
    
    Write-Host "容器已停止" -ForegroundColor Green
}

function Show-Logs {
    Write-Host "查看日志..." -ForegroundColor Yellow
    
    & docker-compose logs -f
}

function Clean-Resources {
    Write-Host "清理资源..." -ForegroundColor Yellow
    
    & docker-compose down --volumes --remove-orphans
    & docker system prune -f
    
    Write-Host "清理完成" -ForegroundColor Green
}

function Push-Images {
    Write-Host "推送镜像到 Docker Hub..." -ForegroundColor Yellow
    
    & docker tag $DOCKER_IMAGE $DOCKER_HUB_IMAGE
    & docker tag $DOCKER_IMAGE_LATEST $DOCKER_HUB_IMAGE_LATEST
    & docker push $DOCKER_HUB_IMAGE
    & docker push $DOCKER_HUB_IMAGE_LATEST
    
    if ($LASTEXITCODE -ne 0) {
        Write-Error "推送失败！"
        exit 1
    }
    
    Write-Host "推送完成" -ForegroundColor Green
}

function Test-Image {
    Write-Host "测试 Docker 镜像..." -ForegroundColor Yellow
    
    $containerName = "test-$APP_NAME"
    
    # 启动测试容器
    & docker run --rm -d -p 8080:8080 --name $containerName $DOCKER_IMAGE_LATEST
    
    if ($LASTEXITCODE -ne 0) {
        Write-Error "启动测试容器失败！"
        exit 1
    }
    
    # 等待容器启动
    Start-Sleep -Seconds 10
    
    # 健康检查
    try {
        $response = Invoke-WebRequest -Uri "http://localhost:8080/" -TimeoutSec 10
        if ($response.StatusCode -eq 200) {
            Write-Host "健康检查通过！" -ForegroundColor Green
        } else {
            Write-Error "健康检查失败！状态码: $($response.StatusCode)"
            & docker stop $containerName
            exit 1
        }
    }
    catch {
        Write-Error "健康检查失败！错误: $($_.Exception.Message)"
        & docker stop $containerName
        exit 1
    }
    
    # 停止测试容器
    & docker stop $containerName
    
    Write-Host "测试通过！" -ForegroundColor Green
}

function Start-DevEnvironment {
    Write-Host "启动开发环境..." -ForegroundColor Yellow
    
    & docker-compose --profile dev up --build lazybala-dev
}

function Show-Help {
    Write-Host @"
LazyBala Docker 构建工具

用法: .\docker-build.ps1 [命令] [选项]

命令:
  build         构建 Docker 镜像
  build-multi   构建多架构镜像
  run           运行生产容器
  stop          停止容器
  logs          查看日志
  clean         清理 Docker 资源
  push          推送镜像到 Docker Hub
  test          测试 Docker 镜像
  dev           启动开发环境
  help          显示此帮助信息

选项:
  -Version      指定版本标签 (默认: git describe)
  -Platform     指定构建平台 (默认: linux/amd64,linux/arm64)
  -Push         构建后推送镜像
  -NoCache      不使用缓存构建

示例:
  .\docker-build.ps1 build
  .\docker-build.ps1 build-multi -Push
  .\docker-build.ps1 run
  .\docker-build.ps1 test
  .\docker-build.ps1 push
"@ -ForegroundColor White
}

# 主逻辑
Write-Header

switch ($Command.ToLower()) {
    "build" { Build-Image }
    "build-multi" { Build-MultiArch }
    "run" { Start-Container }
    "stop" { Stop-Container }
    "logs" { Show-Logs }
    "clean" { Clean-Resources }
    "push" { Push-Images }
    "test" { Test-Image }
    "dev" { Start-DevEnvironment }
    "help" { Show-Help }
    default {
        Write-Error "未知命令: $Command"
        Show-Help
        exit 1
    }
}

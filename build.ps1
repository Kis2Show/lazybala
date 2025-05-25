# LazyBala Docker 构建脚本 (PowerShell)

param(
    [string]$Version = "latest"
)

# 项目信息
$ProjectName = "lazybala"
$ImageName = "lazybala"

Write-Host "=== LazyBala Docker 构建脚本 ===" -ForegroundColor Blue
Write-Host "项目名称: $ProjectName" -ForegroundColor Yellow
Write-Host "镜像名称: ${ImageName}:${Version}" -ForegroundColor Yellow
Write-Host ""

# 函数：打印步骤
function Write-Step {
    param([string]$Message)
    Write-Host "[步骤] $Message" -ForegroundColor Blue
}

# 函数：打印成功
function Write-Success {
    param([string]$Message)
    Write-Host "[成功] $Message" -ForegroundColor Green
}

# 函数：打印错误
function Write-Error {
    param([string]$Message)
    Write-Host "[错误] $Message" -ForegroundColor Red
}

# 函数：打印警告
function Write-Warning {
    param([string]$Message)
    Write-Host "[警告] $Message" -ForegroundColor Yellow
}

# 检查 Docker 是否安装
function Test-Docker {
    Write-Step "检查 Docker 环境..."
    
    try {
        $null = Get-Command docker -ErrorAction Stop
        $null = docker info 2>$null
        Write-Success "Docker 环境检查通过"
        return $true
    }
    catch {
        Write-Error "Docker 未安装或未启动，请检查 Docker 环境"
        return $false
    }
}

# 清理旧镜像
function Remove-OldImages {
    Write-Step "清理旧镜像..."
    
    try {
        $danglingImages = docker images -f "dangling=true" -q
        if ($danglingImages) {
            docker rmi $danglingImages
            Write-Success "已清理悬空镜像"
        }
        else {
            Write-Warning "没有发现悬空镜像"
        }
    }
    catch {
        Write-Warning "清理镜像时出现错误: $_"
    }
}

# 构建镜像
function Build-Image {
    Write-Step "构建 Docker 镜像..."
    
    try {
        $buildTime = (Get-Date).ToUniversalTime().ToString("yyyy-MM-ddTHH:mm:ssZ")
        
        docker build `
            --tag "${ImageName}:${Version}" `
            --tag "${ImageName}:latest" `
            --build-arg BUILDTIME=$buildTime `
            --build-arg VERSION=$Version `
            .
        
        if ($LASTEXITCODE -eq 0) {
            Write-Success "镜像构建完成: ${ImageName}:${Version}"
            return $true
        }
        else {
            Write-Error "镜像构建失败"
            return $false
        }
    }
    catch {
        Write-Error "构建过程中出现错误: $_"
        return $false
    }
}

# 显示镜像信息
function Show-ImageInfo {
    Write-Step "显示镜像信息..."
    
    Write-Host ""
    Write-Host "=== 镜像信息 ===" -ForegroundColor Blue
    docker images $ImageName
    
    Write-Host ""
    Write-Host "=== 镜像详情 ===" -ForegroundColor Blue
    docker inspect "${ImageName}:${Version}" --format='{{.Config.Labels}}'
}

# 运行测试
function Test-Container {
    Write-Step "运行容器测试..."
    
    try {
        # 停止并删除现有容器
        docker stop "${ProjectName}-test" 2>$null
        docker rm "${ProjectName}-test" 2>$null
        
        # 创建测试目录
        $testDataPath = Join-Path $PWD "test-data"
        New-Item -ItemType Directory -Force -Path "$testDataPath\audiobooks" | Out-Null
        New-Item -ItemType Directory -Force -Path "$testDataPath\config" | Out-Null
        New-Item -ItemType Directory -Force -Path "$testDataPath\cookies" | Out-Null
        
        # 运行测试容器
        docker run -d `
            --name "${ProjectName}-test" `
            -p 8081:8080 `
            -v "${testDataPath}\audiobooks:/app/audiobooks" `
            -v "${testDataPath}\config:/app/config" `
            -v "${testDataPath}\cookies:/app/cookies" `
            "${ImageName}:${Version}"
        
        # 等待容器启动
        Start-Sleep -Seconds 5
        
        # 健康检查
        try {
            $response = Invoke-WebRequest -Uri "http://localhost:8081/" -TimeoutSec 10
            if ($response.StatusCode -eq 200) {
                Write-Success "容器测试通过，服务正常运行"
                Write-Host "测试地址: http://localhost:8081" -ForegroundColor Green
            }
            else {
                throw "HTTP 状态码: $($response.StatusCode)"
            }
        }
        catch {
            Write-Error "容器测试失败，服务无法访问: $_"
            docker logs "${ProjectName}-test"
            return $false
        }
        finally {
            # 清理测试容器
            docker stop "${ProjectName}-test" 2>$null
            docker rm "${ProjectName}-test" 2>$null
            Remove-Item -Recurse -Force $testDataPath -ErrorAction SilentlyContinue
        }
        
        return $true
    }
    catch {
        Write-Error "测试过程中出现错误: $_"
        return $false
    }
}

# 主函数
function Main {
    Write-Host "开始构建流程..." -ForegroundColor Blue
    Write-Host ""
    
    if (-not (Test-Docker)) {
        exit 1
    }
    
    Remove-OldImages
    
    if (-not (Build-Image)) {
        exit 1
    }
    
    Show-ImageInfo
    
    # 询问是否运行测试
    $runTest = Read-Host "是否运行容器测试? (y/N)"
    if ($runTest -match "^[Yy]$") {
        if (-not (Test-Container)) {
            exit 1
        }
    }
    
    Write-Host ""
    Write-Success "构建流程完成！"
    Write-Host ""
    Write-Host "=== 使用说明 ===" -ForegroundColor Blue
    Write-Host "启动服务: docker-compose up -d" -ForegroundColor Yellow
    Write-Host "查看日志: docker-compose logs -f" -ForegroundColor Yellow
    Write-Host "停止服务: docker-compose down" -ForegroundColor Yellow
    Write-Host "访问地址: http://localhost:8080" -ForegroundColor Yellow
    Write-Host ""
}

# 执行主函数
Main

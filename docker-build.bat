@echo off
REM LazyBala Docker 构建脚本 (Windows)

setlocal enabledelayedexpansion

REM 设置变量
set APP_NAME=lazybala
set VERSION=dev
set BUILDTIME=%date:~0,4%-%date:~5,2%-%date:~8,2%T%time:~0,2%:%time:~3,2%:%time:~6,2%Z
set DOCKER_IMAGE=%APP_NAME%:%VERSION%
set DOCKER_IMAGE_LATEST=%APP_NAME%:latest
set DOCKER_HUB_IMAGE=kis2show/%APP_NAME%:%VERSION%
set DOCKER_HUB_IMAGE_LATEST=kis2show/%APP_NAME%:latest

REM 获取 Git 版本（如果可用）
for /f "tokens=*" %%i in ('git describe --tags --always --dirty 2^>nul') do set VERSION=%%i
if "%VERSION%"=="" set VERSION=dev

echo ========================================
echo LazyBala Docker 构建工具
echo ========================================
echo 应用名称: %APP_NAME%
echo 版本: %VERSION%
echo 构建时间: %BUILDTIME%
echo Docker 镜像: %DOCKER_IMAGE%
echo ========================================

if "%1"=="" (
    echo 用法: %0 [命令]
    echo.
    echo 可用命令:
    echo   build      - 构建 Docker 镜像
    echo   run        - 运行容器
    echo   stop       - 停止容器
    echo   logs       - 查看日志
    echo   clean      - 清理资源
    echo   push       - 推送到 Docker Hub
    echo   test       - 测试镜像
    echo   dev        - 启动开发环境
    echo   help       - 显示帮助
    goto :end
)

if "%1"=="build" goto :build
if "%1"=="run" goto :run
if "%1"=="stop" goto :stop
if "%1"=="logs" goto :logs
if "%1"=="clean" goto :clean
if "%1"=="push" goto :push
if "%1"=="test" goto :test
if "%1"=="dev" goto :dev
if "%1"=="help" goto :help

echo 未知命令: %1
goto :end

:build
echo 构建 Docker 镜像...
docker build ^
  --build-arg VERSION=%VERSION% ^
  --build-arg BUILDTIME=%BUILDTIME% ^
  -t %DOCKER_IMAGE% ^
  -t %DOCKER_IMAGE_LATEST% ^
  .
if %errorlevel% neq 0 (
    echo 构建失败！
    exit /b 1
)
echo 构建成功！
goto :end

:run
echo 启动容器...
docker-compose up -d
if %errorlevel% neq 0 (
    echo 启动失败！
    exit /b 1
)
echo 容器已启动，访问 http://localhost:8080
goto :end

:stop
echo 停止容器...
docker-compose down
echo 容器已停止
goto :end

:logs
echo 查看日志...
docker-compose logs -f
goto :end

:clean
echo 清理资源...
docker-compose down --volumes --remove-orphans
docker system prune -f
echo 清理完成
goto :end

:push
echo 推送镜像到 Docker Hub...
docker tag %DOCKER_IMAGE% %DOCKER_HUB_IMAGE%
docker tag %DOCKER_IMAGE_LATEST% %DOCKER_HUB_IMAGE_LATEST%
docker push %DOCKER_HUB_IMAGE%
docker push %DOCKER_HUB_IMAGE_LATEST%
echo 推送完成
goto :end

:test
echo 测试 Docker 镜像...
docker run --rm -d -p 8080:8080 --name test-%APP_NAME% %DOCKER_IMAGE_LATEST%
timeout /t 10 /nobreak >nul
curl -f http://localhost:8080/ >nul 2>&1
if %errorlevel% neq 0 (
    echo 健康检查失败！
    docker stop test-%APP_NAME%
    exit /b 1
)
docker stop test-%APP_NAME%
echo 测试通过！
goto :end

:dev
echo 启动开发环境...
docker-compose --profile dev up --build lazybala-dev
goto :end

:help
echo LazyBala Docker 构建工具
echo.
echo 用法: %0 [命令]
echo.
echo 可用命令:
echo   build      - 构建 Docker 镜像
echo   run        - 运行生产容器
echo   stop       - 停止容器
echo   logs       - 查看日志
echo   clean      - 清理 Docker 资源
echo   push       - 推送镜像到 Docker Hub
echo   test       - 测试 Docker 镜像
echo   dev        - 启动开发环境
echo   help       - 显示此帮助信息
echo.
echo 示例:
echo   %0 build     # 构建镜像
echo   %0 run       # 启动服务
echo   %0 logs      # 查看日志
goto :end

:end
endlocal

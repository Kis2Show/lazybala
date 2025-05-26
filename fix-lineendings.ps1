# 修复换行符问题的 PowerShell 脚本

Write-Host "修复 Git 换行符配置..." -ForegroundColor Yellow

# 设置 Git 配置
Write-Host "配置 Git 换行符处理..." -ForegroundColor Green
git config core.autocrlf false
git config core.eol lf

# 检查当前状态
Write-Host "检查 Git 状态..." -ForegroundColor Green
$status = git status --porcelain

if ($status) {
    Write-Host "提交当前更改..." -ForegroundColor Green
    git add .
    git commit -m "规范化换行符配置"
    
    Write-Host "推送到远程仓库..." -ForegroundColor Green
    git push origin main
} else {
    Write-Host "没有需要提交的更改" -ForegroundColor Green
}

Write-Host "换行符配置完成！" -ForegroundColor Cyan
Write-Host "现在 Git 将使用 .gitattributes 文件来管理换行符" -ForegroundColor Cyan

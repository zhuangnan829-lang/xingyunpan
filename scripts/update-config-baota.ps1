# 更新配置文件以连接宝塔 MySQL
param(
    [Parameter(Mandatory=$true)]
    [string]$Host,
    
    [Parameter(Mandatory=$true)]
    [string]$Password,
    
    [string]$Port = "3306",
    [string]$Username = "root",
    [string]$Database = "xingyunpan"
)

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "更新配置文件以连接宝塔 MySQL" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

$configFile = "configs/config.yaml"
$backupFile = "configs/config.yaml.backup.$(Get-Date -Format 'yyyyMMdd_HHmmss')"

# 检查配置文件是否存在
if (-not (Test-Path $configFile)) {
    Write-Host "❌ 配置文件不存在: $configFile" -ForegroundColor Red
    exit 1
}

# 备份原配置文件
Write-Host "正在备份原配置文件..." -ForegroundColor Yellow
Copy-Item $configFile $backupFile
Write-Host "✅ 备份完成: $backupFile" -ForegroundColor Green
Write-Host ""

# 读取配置文件
$content = Get-Content $configFile -Raw

# 更新数据库配置
Write-Host "正在更新数据库配置..." -ForegroundColor Yellow
$content = $content -replace 'host: localhost', "host: $Host"
$content = $content -replace 'port: 3306', "port: $Port"
$content = $content -replace 'username: root', "username: $Username"
$content = $content -replace 'password: your_password_here', "password: $Password"
$content = $content -replace 'database: xingyunpan', "database: $Database"

# 写入配置文件
Set-Content $configFile $content -NoNewline

Write-Host "✅ 配置文件已更新！" -ForegroundColor Green
Write-Host ""
Write-Host "新配置：" -ForegroundColor Cyan
Write-Host "  服务器: $Host:$Port" -ForegroundColor White
Write-Host "  用户名: $Username" -ForegroundColor White
Write-Host "  数据库: $Database" -ForegroundColor White
Write-Host ""
Write-Host "下一步：" -ForegroundColor Cyan
Write-Host "1. 运行测试: go run scripts/test-baota-connection.go" -ForegroundColor White
Write-Host "2. 启动应用: go run cmd/server/main.go" -ForegroundColor White
Write-Host ""
Write-Host "如需恢复原配置，运行：" -ForegroundColor Yellow
Write-Host "  Copy-Item $backupFile $configFile" -ForegroundColor White

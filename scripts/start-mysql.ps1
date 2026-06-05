# 宝塔 MySQL 启动脚本 (PowerShell)
# 编码: UTF-8

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "宝塔 MySQL 启动脚本" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# 函数：测试端口是否被占用
function Test-Port {
    param([int]$Port)
    $connection = Test-NetConnection -ComputerName localhost -Port $Port -WarningAction SilentlyContinue
    return $connection.TcpTestSucceeded
}

# 函数：查找 MySQL 服务
function Find-MySQLService {
    Write-Host "[1] 查找 MySQL 服务..." -ForegroundColor Yellow
    Write-Host ""
    
    $services = Get-Service | Where-Object { $_.Name -like "*mysql*" -or $_.DisplayName -like "*mysql*" }
    
    if ($services) {
        Write-Host "找到以下 MySQL 服务:" -ForegroundColor Green
        foreach ($service in $services) {
            Write-Host "  - 名称: $($service.Name)" -ForegroundColor White
            Write-Host "    显示名: $($service.DisplayName)" -ForegroundColor Gray
            Write-Host "    状态: $($service.Status)" -ForegroundColor $(if ($service.Status -eq 'Running') { 'Green' } else { 'Red' })
            Write-Host ""
        }
        return $services
    } else {
        Write-Host "❌ 未找到 MySQL 服务" -ForegroundColor Red
        return $null
    }
}

# 函数：启动 MySQL 服务
function Start-MySQLService {
    param($Service)
    
    Write-Host "正在启动服务: $($Service.Name)..." -ForegroundColor Yellow
    
    try {
        Start-Service -Name $Service.Name -ErrorAction Stop
        Write-Host "✓ 服务启动成功！" -ForegroundColor Green
        return $true
    } catch {
        Write-Host "❌ 启动失败: $($_.Exception.Message)" -ForegroundColor Red
        return $false
    }
}

# 函数：查找宝塔 MySQL 安装目录
function Find-BTMySQL {
    Write-Host "[2] 查找宝塔 MySQL 安装目录..." -ForegroundColor Yellow
    Write-Host ""
    
    $possiblePaths = @(
        "C:\BtSoft\mysql",
        "D:\BtSoft\mysql",
        "E:\BtSoft\mysql",
        "C:\phpstudy_pro\Extensions\MySQL5.7.26",
        "D:\phpstudy_pro\Extensions\MySQL5.7.26",
        "C:\phpstudy_pro\Extensions\MySQL8.0.12",
        "D:\phpstudy_pro\Extensions\MySQL8.0.12",
        "C:\Program Files\MySQL",
        "C:\Program Files (x86)\MySQL"
    )
    
    foreach ($path in $possiblePaths) {
        if (Test-Path "$path\bin\mysqld.exe") {
            Write-Host "✓ 找到 MySQL: $path" -ForegroundColor Green
            return $path
        }
    }
    
    Write-Host "❌ 未找到宝塔 MySQL 安装目录" -ForegroundColor Red
    return $null
}

# 主逻辑
Write-Host "开始检查 MySQL 状态..." -ForegroundColor Cyan
Write-Host ""

# 检查 3306 端口
if (Test-Port -Port 3306) {
    Write-Host "✓ 端口 3306 已被占用，MySQL 可能正在运行" -ForegroundColor Green
    Write-Host ""
    Write-Host "尝试连接数据库..." -ForegroundColor Yellow
    & go run scripts/test-db-connection.go
    exit 0
} else {
    Write-Host "⚠ 端口 3306 未被占用，MySQL 未运行" -ForegroundColor Yellow
    Write-Host ""
}

# 查找并启动 MySQL 服务
$services = Find-MySQLService

if ($services) {
    foreach ($service in $services) {
        if ($service.Status -eq 'Running') {
            Write-Host "✓ 服务 $($service.Name) 已在运行" -ForegroundColor Green
        } else {
            Write-Host "尝试启动服务 $($service.Name)..." -ForegroundColor Yellow
            
            # 需要管理员权限
            if (-NOT ([Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole] "Administrator")) {
                Write-Host "❌ 需要管理员权限才能启动服务" -ForegroundColor Red
                Write-Host ""
                Write-Host "请以管理员身份运行此脚本，或手动启动服务：" -ForegroundColor Yellow
                Write-Host "  1. 以管理员身份打开 PowerShell" -ForegroundColor White
                Write-Host "  2. 运行: Start-Service $($service.Name)" -ForegroundColor White
                Write-Host ""
            } else {
                if (Start-MySQLService -Service $service) {
                    Write-Host ""
                    Write-Host "等待服务启动..." -ForegroundColor Yellow
                    Start-Sleep -Seconds 3
                    
                    Write-Host "测试数据库连接..." -ForegroundColor Yellow
                    & go run scripts/test-db-connection.go
                    exit 0
                }
            }
        }
    }
}

# 如果服务方式失败，尝试查找安装目录
Write-Host ""
$mysqlPath = Find-BTMySQL

if ($mysqlPath) {
    Write-Host ""
    Write-Host "找到 MySQL 安装目录，但无法通过服务启动" -ForegroundColor Yellow
    Write-Host "MySQL 路径: $mysqlPath" -ForegroundColor White
    Write-Host ""
    Write-Host "建议操作：" -ForegroundColor Yellow
    Write-Host "1. 在宝塔面板中启动 MySQL" -ForegroundColor White
    Write-Host "2. 或者手动注册 MySQL 服务：" -ForegroundColor White
    Write-Host "   $mysqlPath\bin\mysqld.exe --install MySQL --defaults-file=$mysqlPath\my.ini" -ForegroundColor Gray
    Write-Host "   然后运行: net start MySQL" -ForegroundColor Gray
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "手动启动指南" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "方法 1: 使用宝塔面板（推荐）" -ForegroundColor Yellow
Write-Host "  1. 打开浏览器访问: http://127.0.0.1:888" -ForegroundColor White
Write-Host "  2. 登录宝塔面板" -ForegroundColor White
Write-Host "  3. 进入'软件商店'或'已安装'" -ForegroundColor White
Write-Host "  4. 找到 MySQL，点击'启动'按钮" -ForegroundColor White
Write-Host ""
Write-Host "方法 2: 使用服务管理器" -ForegroundColor Yellow
Write-Host "  1. 按 Win+R，输入: services.msc" -ForegroundColor White
Write-Host "  2. 找到 MySQL 服务" -ForegroundColor White
Write-Host "  3. 右键点击'启动'" -ForegroundColor White
Write-Host ""
Write-Host "方法 3: 使用命令行（需要管理员权限）" -ForegroundColor Yellow
Write-Host "  net start MySQL" -ForegroundColor White
Write-Host ""

Write-Host "按任意键退出..."
$null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")

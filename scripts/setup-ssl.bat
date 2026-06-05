@echo off
REM SSL 证书申请脚本 (Windows 版本 - 用于远程管理)
REM 此脚本用于通过 SSH 连接到 Linux 服务器并申请证书

setlocal enabledelayedexpansion

echo ========================================
echo SSL 证书申请脚本 (远程执行)
echo ========================================
echo.

REM 检查参数
if "%~1"=="" (
    echo 错误: 缺少参数
    echo.
    echo 用法: %~nx0 ^<domain^> ^<email^> [--staging]
    echo 示例: %~nx0 example.com admin@example.com
    echo       %~nx0 example.com admin@example.com --staging
    echo.
    exit /b 1
)

if "%~2"=="" (
    echo 错误: 缺少邮箱参数
    echo.
    echo 用法: %~nx0 ^<domain^> ^<email^> [--staging]
    exit /b 1
)

set DOMAIN=%~1
set EMAIL=%~2
set STAGING=%~3

REM 服务器配置
set SERVER_HOST=117.24.15.9
set SERVER_USER=root
set SERVER_PORT=22

echo 域名: %DOMAIN%
echo 邮箱: %EMAIL%
if "%STAGING%"=="--staging" (
    echo 模式: 测试模式 (staging^)
) else (
    echo 模式: 生产模式
)
echo.
echo 服务器: %SERVER_USER%@%SERVER_HOST%:%SERVER_PORT%
echo.

REM 检查 SSH 客户端
where ssh >nul 2>&1
if %errorlevel% neq 0 (
    echo 错误: 未找到 SSH 客户端
    echo 请安装 OpenSSH 客户端或使用 PuTTY
    exit /b 1
)

echo 正在连接到服务器...
echo.

REM 创建临时脚本
set TEMP_SCRIPT=%TEMP%\setup-ssl-remote.sh

(
echo #!/bin/bash
echo set -e
echo.
echo # 颜色输出
echo RED='\033[0;31m'
echo GREEN='\033[0;32m'
echo YELLOW='\033[1;33m'
echo NC='\033[0m'
echo.
echo print_info^(^) { echo -e "${GREEN}[INFO]${NC} $1"; }
echo print_warn^(^) { echo -e "${YELLOW}[WARN]${NC} $1"; }
echo print_error^(^) { echo -e "${RED}[ERROR]${NC} $1"; }
echo.
echo DOMAIN="%DOMAIN%"
echo EMAIL="%EMAIL%"
echo STAGING="%STAGING%"
echo.
echo print_info "=== SSL 证书申请 ==="
echo print_info "域名: $DOMAIN"
echo print_info "邮箱: $EMAIL"
echo.
echo # 检查 certbot
echo print_info "检查 certbot..."
echo if ! command -v certbot ^&^> /dev/null; then
echo     print_warn "certbot 未安装，正在安装..."
echo     if [ -f /etc/debian_version ]; then
echo         apt update ^&^& apt install certbot python3-certbot-nginx -y
echo     elif [ -f /etc/redhat-release ]; then
echo         yum install epel-release -y ^&^& yum install certbot python3-certbot-nginx -y
echo     fi
echo fi
echo.
echo # 申请证书
echo print_info "申请证书..."
echo if systemctl is-active --quiet nginx; then
echo     certbot --nginx -d $DOMAIN --email $EMAIL --agree-tos --no-eff-email --redirect $STAGING
echo else
echo     certbot certonly --standalone -d $DOMAIN --email $EMAIL --agree-tos --no-eff-email $STAGING
echo fi
echo.
echo # 配置自动续期
echo print_info "配置自动续期..."
echo if systemctl list-unit-files ^| grep -q certbot.timer; then
echo     systemctl enable certbot.timer
echo     systemctl start certbot.timer
echo else
echo     ^(crontab -l 2^>/dev/null ^| grep -v "certbot renew"; echo "30 2 * * * certbot renew --quiet --post-hook 'systemctl reload nginx'"^) ^| crontab -
echo fi
echo.
echo # 测试续期
echo print_info "测试自动续期..."
echo certbot renew --dry-run
echo.
echo print_info "证书申请完成！"
echo certbot certificates -d $DOMAIN
) > "%TEMP_SCRIPT%"

REM 上传并执行脚本
echo 上传脚本到服务器...
scp -P %SERVER_PORT% "%TEMP_SCRIPT%" %SERVER_USER%@%SERVER_HOST%:/tmp/setup-ssl.sh

if %errorlevel% neq 0 (
    echo 错误: 上传脚本失败
    del "%TEMP_SCRIPT%"
    exit /b 1
)

echo.
echo 执行脚本...
ssh -p %SERVER_PORT% %SERVER_USER%@%SERVER_HOST% "chmod +x /tmp/setup-ssl.sh && /tmp/setup-ssl.sh && rm /tmp/setup-ssl.sh"

if %errorlevel% neq 0 (
    echo.
    echo 错误: 证书申请失败
    del "%TEMP_SCRIPT%"
    exit /b 1
)

REM 清理临时文件
del "%TEMP_SCRIPT%"

echo.
echo ========================================
echo SSL 证书申请完成
echo ========================================
echo.
echo 下一步:
echo 1. 检查 Nginx 配置
echo 2. 访问 https://%DOMAIN% 验证证书
echo 3. 使用 SSL Labs 测试: https://www.ssllabs.com/ssltest/
echo.

endlocal

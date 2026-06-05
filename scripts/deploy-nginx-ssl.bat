@echo off
REM Nginx HTTPS 配置部署脚本 (Windows 版本 - 用于远程管理)

setlocal enabledelayedexpansion

echo ========================================
echo Nginx HTTPS 配置部署 (远程执行)
echo ========================================
echo.

REM 检查参数
if "%~1"=="" (
    echo 错误: 缺少域名参数
    echo.
    echo 用法: %~nx0 ^<domain^>
    echo 示例: %~nx0 example.com
    echo.
    exit /b 1
)

set DOMAIN=%~1

REM 服务器配置
set SERVER_HOST=117.24.15.9
set SERVER_USER=root
set SERVER_PORT=22

echo 域名: %DOMAIN%
echo 服务器: %SERVER_USER%@%SERVER_HOST%:%SERVER_PORT%
echo.

REM 检查 SSH 客户端
where ssh >nul 2>&1
if %errorlevel% neq 0 (
    echo 错误: 未找到 SSH 客户端
    exit /b 1
)

REM 检查配置文件
if not exist "configs\nginx-ssl.conf" (
    echo 错误: 配置文件不存在: configs\nginx-ssl.conf
    exit /b 1
)

echo 正在上传配置文件...
scp -P %SERVER_PORT% configs\nginx-ssl.conf %SERVER_USER%@%SERVER_HOST%:/tmp/nginx-ssl.conf

if %errorlevel% neq 0 (
    echo 错误: 上传配置文件失败
    exit /b 1
)

echo.
echo 正在部署配置...
echo.

REM 创建部署脚本
set TEMP_SCRIPT=%TEMP%\deploy-nginx-ssl.sh

(
echo #!/bin/bash
echo set -e
echo.
echo DOMAIN="%DOMAIN%"
echo CONFIG_FILE="/tmp/nginx-ssl.conf"
echo SITE_NAME="xingyunpan"
echo.
echo echo "[INFO] 检查 SSL 证书..."
echo if [ ! -f "/etc/letsencrypt/live/$DOMAIN/fullchain.pem" ]; then
echo     echo "[ERROR] SSL 证书不存在"
echo     exit 1
echo fi
echo.
echo echo "[INFO] 生成配置文件..."
echo sed "s/your-domain.com/$DOMAIN/g" "$CONFIG_FILE" ^> /etc/nginx/sites-available/$SITE_NAME
echo.
echo echo "[INFO] 启用站点..."
echo ln -sf /etc/nginx/sites-available/$SITE_NAME /etc/nginx/sites-enabled/$SITE_NAME
echo.
echo echo "[INFO] 删除默认站点..."
echo rm -f /etc/nginx/sites-enabled/default
echo.
echo echo "[INFO] 创建目录..."
echo mkdir -p /var/www/letsencrypt
echo.
echo echo "[INFO] 测试配置..."
echo nginx -t
echo.
echo echo "[INFO] 重载 Nginx..."
echo systemctl reload nginx
echo.
echo echo "[INFO] 检查状态..."
echo systemctl status nginx --no-pager ^| head -n 10
echo.
echo echo "[INFO] 配置部署完成"
echo rm -f "$CONFIG_FILE"
) > "%TEMP_SCRIPT%"

REM 上传并执行部署脚本
scp -P %SERVER_PORT% "%TEMP_SCRIPT%" %SERVER_USER%@%SERVER_HOST%:/tmp/deploy-nginx.sh

if %errorlevel% neq 0 (
    echo 错误: 上传部署脚本失败
    del "%TEMP_SCRIPT%"
    exit /b 1
)

ssh -p %SERVER_PORT% %SERVER_USER%@%SERVER_HOST% "chmod +x /tmp/deploy-nginx.sh && /tmp/deploy-nginx.sh && rm /tmp/deploy-nginx.sh"

if %errorlevel% neq 0 (
    echo.
    echo 错误: 配置部署失败
    del "%TEMP_SCRIPT%"
    exit /b 1
)

REM 清理
del "%TEMP_SCRIPT%"

echo.
echo ========================================
echo 配置部署完成
echo ========================================
echo.
echo 下一步:
echo 1. 测试 HTTP 重定向: curl -I http://%DOMAIN%
echo 2. 测试 HTTPS 访问: curl -I https://%DOMAIN%
echo 3. 在浏览器中访问: https://%DOMAIN%
echo 4. SSL Labs 测试: https://www.ssllabs.com/ssltest/analyze.html?d=%DOMAIN%
echo.

endlocal

#!/bin/bash

# Nginx HTTPS 配置部署脚本

set -e

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

print_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查是否以 root 运行
if [ "$EUID" -ne 0 ]; then 
    print_error "请使用 root 权限运行此脚本"
    echo "使用: sudo $0"
    exit 1
fi

# 检查参数
if [ $# -lt 1 ]; then
    print_error "用法: $0 <domain>"
    echo "示例: $0 example.com"
    exit 1
fi

DOMAIN=$1
CONFIG_FILE="configs/nginx-ssl.conf"
NGINX_SITES_AVAILABLE="/etc/nginx/sites-available"
NGINX_SITES_ENABLED="/etc/nginx/sites-enabled"
SITE_NAME="xingyunpan"

print_info "=== Nginx HTTPS 配置部署 ==="
print_info "域名: $DOMAIN"
echo ""

# 检查配置文件是否存在
if [ ! -f "$CONFIG_FILE" ]; then
    print_error "配置文件不存在: $CONFIG_FILE"
    exit 1
fi

# 检查 Nginx 是否已安装
print_info "检查 Nginx..."
if ! command -v nginx &> /dev/null; then
    print_error "Nginx 未安装"
    echo "请先安装 Nginx:"
    echo "  Ubuntu/Debian: sudo apt install nginx"
    echo "  CentOS/RHEL: sudo yum install nginx"
    exit 1
fi

print_info "Nginx 版本: $(nginx -v 2>&1 | cut -d'/' -f2)"

# 检查 SSL 证书是否存在
print_info "检查 SSL 证书..."
CERT_PATH="/etc/letsencrypt/live/$DOMAIN/fullchain.pem"
KEY_PATH="/etc/letsencrypt/live/$DOMAIN/privkey.pem"

if [ ! -f "$CERT_PATH" ] || [ ! -f "$KEY_PATH" ]; then
    print_error "SSL 证书不存在"
    echo "证书路径: $CERT_PATH"
    echo "私钥路径: $KEY_PATH"
    echo ""
    echo "请先申请 SSL 证书:"
    echo "  sudo ./scripts/setup-ssl.sh $DOMAIN your-email@example.com"
    exit 1
fi

print_info "SSL 证书已找到"

# 创建临时配置文件（替换域名）
print_info "生成配置文件..."
TEMP_CONFIG="/tmp/nginx-ssl-$DOMAIN.conf"
sed "s/your-domain.com/$DOMAIN/g" "$CONFIG_FILE" > "$TEMP_CONFIG"

# 验证配置文件语法
print_info "验证配置文件语法..."
nginx -t -c /etc/nginx/nginx.conf -g "include $TEMP_CONFIG;" 2>&1 | grep -q "syntax is ok" || {
    print_error "配置文件语法错误"
    nginx -t -c /etc/nginx/nginx.conf -g "include $TEMP_CONFIG;"
    rm "$TEMP_CONFIG"
    exit 1
}

print_info "配置文件语法正确"

# 备份现有配置
if [ -f "$NGINX_SITES_AVAILABLE/$SITE_NAME" ]; then
    print_info "备份现有配置..."
    cp "$NGINX_SITES_AVAILABLE/$SITE_NAME" "$NGINX_SITES_AVAILABLE/$SITE_NAME.backup.$(date +%Y%m%d%H%M%S)"
fi

# 复制配置文件
print_info "部署配置文件..."
cp "$TEMP_CONFIG" "$NGINX_SITES_AVAILABLE/$SITE_NAME"
rm "$TEMP_CONFIG"

# 创建符号链接（如果不存在）
if [ ! -L "$NGINX_SITES_ENABLED/$SITE_NAME" ]; then
    print_info "启用站点配置..."
    ln -s "$NGINX_SITES_AVAILABLE/$SITE_NAME" "$NGINX_SITES_ENABLED/$SITE_NAME"
fi

# 删除默认站点（如果存在）
if [ -L "$NGINX_SITES_ENABLED/default" ]; then
    print_warn "删除默认站点配置..."
    rm "$NGINX_SITES_ENABLED/default"
fi

# 创建日志目录
print_info "创建日志目录..."
mkdir -p /var/log/nginx

# 创建 Let's Encrypt 验证目录
print_info "创建 Let's Encrypt 验证目录..."
mkdir -p /var/www/letsencrypt

# 测试配置
print_info "测试 Nginx 配置..."
nginx -t

if [ $? -ne 0 ]; then
    print_error "Nginx 配置测试失败"
    exit 1
fi

print_info "Nginx 配置测试成功"

# 重载 Nginx
print_info "重载 Nginx..."
systemctl reload nginx

if [ $? -ne 0 ]; then
    print_error "Nginx 重载失败"
    systemctl status nginx
    exit 1
fi

print_info "Nginx 重载成功"

# 检查 Nginx 状态
print_info "检查 Nginx 状态..."
systemctl status nginx --no-pager | head -n 10

# 检查端口监听
print_info "检查端口监听..."
netstat -tlnp | grep nginx | grep -E ":(80|443) "

echo ""
print_info "=== 配置部署完成 ==="
echo ""
print_info "下一步:"
echo "  1. 测试 HTTP 重定向: curl -I http://$DOMAIN"
echo "  2. 测试 HTTPS 访问: curl -I https://$DOMAIN"
echo "  3. 在浏览器中访问: https://$DOMAIN"
echo "  4. 使用 SSL Labs 测试: https://www.ssllabs.com/ssltest/analyze.html?d=$DOMAIN"
echo ""
print_info "配置文件位置:"
echo "  $NGINX_SITES_AVAILABLE/$SITE_NAME"
echo ""
print_info "日志文件位置:"
echo "  /var/log/nginx/xingyunpan-https-access.log"
echo "  /var/log/nginx/xingyunpan-https-error.log"

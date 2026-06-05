#!/bin/bash

# SSL 证书申请脚本
# 用于自动化申请 Let's Encrypt SSL 证书

set -e

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 打印带颜色的消息
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
if [ $# -lt 2 ]; then
    print_error "用法: $0 <domain> <email> [--staging]"
    echo "示例: $0 example.com admin@example.com"
    echo "      $0 example.com admin@example.com --staging  # 测试模式"
    exit 1
fi

DOMAIN=$1
EMAIL=$2
STAGING=""

# 检查是否使用测试模式
if [ "$3" == "--staging" ]; then
    STAGING="--staging"
    print_warn "使用测试模式（staging）申请证书"
fi

print_info "=== SSL 证书申请脚本 ==="
print_info "域名: $DOMAIN"
print_info "邮箱: $EMAIL"
echo ""

# 检查 certbot 是否已安装
print_info "检查 certbot 是否已安装..."
if ! command -v certbot &> /dev/null; then
    print_warn "certbot 未安装，正在安装..."
    
    # 检测操作系统
    if [ -f /etc/debian_version ]; then
        # Debian/Ubuntu
        apt update
        apt install certbot python3-certbot-nginx -y
    elif [ -f /etc/redhat-release ]; then
        # CentOS/RHEL
        yum install epel-release -y
        yum install certbot python3-certbot-nginx -y
    else
        print_error "不支持的操作系统"
        exit 1
    fi
    
    print_info "certbot 安装完成"
else
    print_info "certbot 已安装: $(certbot --version)"
fi

# 检查域名解析
print_info "检查域名解析..."
DOMAIN_IP=$(dig +short $DOMAIN | tail -n1)
SERVER_IP=$(curl -s ifconfig.me)

if [ -z "$DOMAIN_IP" ]; then
    print_error "域名 $DOMAIN 无法解析"
    exit 1
fi

print_info "域名 IP: $DOMAIN_IP"
print_info "服务器 IP: $SERVER_IP"

if [ "$DOMAIN_IP" != "$SERVER_IP" ]; then
    print_warn "域名 IP 与服务器 IP 不匹配"
    read -p "是否继续？(y/n) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

# 检查 Nginx 是否运行
print_info "检查 Nginx 状态..."
if systemctl is-active --quiet nginx; then
    print_info "Nginx 正在运行"
    NGINX_RUNNING=true
else
    print_warn "Nginx 未运行"
    NGINX_RUNNING=false
fi

# 检查 80 端口是否可用
print_info "检查 80 端口..."
if netstat -tlnp | grep -q ":80 "; then
    print_info "80 端口已被占用（正常，Nginx 应该在监听）"
else
    print_warn "80 端口未被占用"
fi

# 申请证书
print_info "开始申请 SSL 证书..."
echo ""

if [ "$NGINX_RUNNING" = true ]; then
    # 使用 nginx 插件
    print_info "使用 Nginx 插件申请证书..."
    certbot --nginx \
        -d $DOMAIN \
        --email $EMAIL \
        --agree-tos \
        --no-eff-email \
        --redirect \
        $STAGING
else
    # 使用 standalone 模式
    print_info "使用 Standalone 模式申请证书..."
    certbot certonly --standalone \
        -d $DOMAIN \
        --email $EMAIL \
        --agree-tos \
        --no-eff-email \
        $STAGING
fi

if [ $? -eq 0 ]; then
    print_info "证书申请成功！"
    echo ""
    
    # 显示证书信息
    print_info "证书信息:"
    certbot certificates -d $DOMAIN
    echo ""
    
    # 证书文件位置
    print_info "证书文件位置:"
    echo "  证书: /etc/letsencrypt/live/$DOMAIN/fullchain.pem"
    echo "  私钥: /etc/letsencrypt/live/$DOMAIN/privkey.pem"
    echo ""
    
    # 配置自动续期
    print_info "配置自动续期..."
    
    # 检查 systemd timer
    if systemctl list-unit-files | grep -q certbot.timer; then
        systemctl enable certbot.timer
        systemctl start certbot.timer
        print_info "已启用 systemd timer 自动续期"
        systemctl status certbot.timer --no-pager
    else
        # 配置 cron job
        print_info "配置 cron job 自动续期..."
        CRON_CMD="30 2 * * * certbot renew --quiet --post-hook 'systemctl reload nginx'"
        (crontab -l 2>/dev/null | grep -v "certbot renew"; echo "$CRON_CMD") | crontab -
        print_info "已添加 cron job（每天 2:30 AM 检查续期）"
    fi
    
    # 测试自动续期
    print_info "测试自动续期配置..."
    certbot renew --dry-run
    
    if [ $? -eq 0 ]; then
        print_info "自动续期配置测试成功！"
    else
        print_warn "自动续期配置测试失败，请检查配置"
    fi
    
    # 重载 Nginx
    if [ "$NGINX_RUNNING" = true ]; then
        print_info "重载 Nginx 配置..."
        systemctl reload nginx
    fi
    
    echo ""
    print_info "=== SSL 证书配置完成 ==="
    print_info "下一步："
    print_info "1. 检查 Nginx 配置是否正确"
    print_info "2. 访问 https://$DOMAIN 验证证书"
    print_info "3. 使用 SSL Labs 测试 SSL 配置: https://www.ssllabs.com/ssltest/"
    
else
    print_error "证书申请失败"
    echo ""
    print_info "故障排查："
    print_info "1. 检查域名解析是否正确"
    print_info "2. 检查防火墙是否开放 80 和 443 端口"
    print_info "3. 检查 Nginx 配置是否正确"
    print_info "4. 查看详细日志: /var/log/letsencrypt/letsencrypt.log"
    exit 1
fi

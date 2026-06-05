#!/bin/bash
# 星云盘 MySQL 远程访问修复脚本（宝塔面板版）
# 在宝塔面板终端中执行此脚本

echo "========================================"
echo "星云盘 MySQL 远程访问修复"
echo "========================================"
echo ""

# 提示输入 MySQL root 密码
echo "请输入 MySQL root 密码（在宝塔面板 > 数据库 > root密码 中查看）:"
read -s MYSQL_ROOT_PASSWORD

echo ""
echo "开始修复 MySQL 远程访问权限..."
echo ""

# 执行修复命令
mysql -u root -p"$MYSQL_ROOT_PASSWORD" <<'MYSQL_SCRIPT'

-- 显示当前状态
SELECT '=== 当前 xingyunpan 用户 ===' as '';
SELECT user, host, plugin FROM mysql.user WHERE user='xingyunpan';

-- 删除所有现有的 xingyunpan 用户
SELECT '' as '';
SELECT '=== 删除现有用户 ===' as '';
DROP USER IF EXISTS 'xingyunpan'@'localhost';
DROP USER IF EXISTS 'xingyunpan'@'127.0.0.1';
DROP USER IF EXISTS 'xingyunpan'@'%';

-- 重新创建用户（使用 mysql_native_password 认证方式）
SELECT '' as '';
SELECT '=== 创建新用户（使用 mysql_native_password）===' as '';
CREATE USER 'xingyunpan'@'%' IDENTIFIED WITH mysql_native_password BY 'uKcgJLKyWHHJzm';

-- 授予所有权限
SELECT '' as '';
SELECT '=== 授予权限 ===' as '';
GRANT ALL PRIVILEGES ON xingyunpan.* TO 'xingyunpan'@'%';

-- 刷新权限
SELECT '' as '';
SELECT '=== 刷新权限 ===' as '';
FLUSH PRIVILEGES;

-- 验证结果
SELECT '' as '';
SELECT '=== 验证用户和权限 ===' as '';
SELECT user, host, plugin FROM mysql.user WHERE user='xingyunpan';

SELECT '' as '';
SHOW GRANTS FOR 'xingyunpan'@'%';

SELECT '' as '';
SELECT '=== 完成！===' as '';

MYSQL_SCRIPT

if [ $? -eq 0 ]; then
    echo ""
    echo "========================================"
    echo "✅ 修复成功！"
    echo "========================================"
    echo ""
    echo "现在请在你的本地电脑上测试连接："
    echo "go run scripts/test-db-connection.go"
    echo ""
else
    echo ""
    echo "========================================"
    echo "❌ 修复失败"
    echo "========================================"
    echo ""
    echo "可能的原因："
    echo "1. MySQL root 密码不正确"
    echo "2. MySQL 服务未运行"
    echo ""
    echo "请检查后重试"
    echo ""
fi

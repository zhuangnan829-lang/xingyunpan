#!/bin/bash
# 星云盘 MySQL 远程访问修复脚本
# 在宝塔面板终端中执行

echo "========================================"
echo "星云盘 MySQL 远程访问修复脚本"
echo "========================================"
echo ""

# MySQL root 密码（请在宝塔面板中查看并替换）
echo "请输入 MySQL root 密码（在宝塔面板 > 数据库 > root密码 中查看）:"
read -s MYSQL_ROOT_PASSWORD

echo ""
echo "开始修复..."
echo ""

# 连接 MySQL 并执行修复命令
mysql -u root -p"$MYSQL_ROOT_PASSWORD" <<EOF
-- 显示当前用户
SELECT '当前 xingyunpan 用户:' as '';
SELECT user, host FROM mysql.user WHERE user='xingyunpan';

-- 删除所有现有的 xingyunpan 用户
SELECT '删除现有用户...' as '';
DROP USER IF EXISTS 'xingyunpan'@'localhost';
DROP USER IF EXISTS 'xingyunpan'@'127.0.0.1';
DROP USER IF EXISTS 'xingyunpan'@'%';

-- 重新创建用户（使用 mysql_native_password）
SELECT '创建新用户...' as '';
CREATE USER 'xingyunpan'@'%' IDENTIFIED WITH mysql_native_password BY 'uKcgJLKyWHHJzm';

-- 授予权限
SELECT '授予权限...' as '';
GRANT ALL PRIVILEGES ON xingyunpan.* TO 'xingyunpan'@'%';

-- 刷新权限
SELECT '刷新权限...' as '';
FLUSH PRIVILEGES;

-- 验证
SELECT '验证用户和权限:' as '';
SELECT user, host, plugin FROM mysql.user WHERE user='xingyunpan';
SHOW GRANTS FOR 'xingyunpan'@'%';

SELECT '完成!' as '';
EOF

echo ""
echo "========================================"
echo "修复完成！"
echo "========================================"
echo ""
echo "现在请测试连接："
echo "go run scripts/test-db-connection.go"
echo ""

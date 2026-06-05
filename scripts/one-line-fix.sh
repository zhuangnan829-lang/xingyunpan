#!/bin/bash
# 一行命令修复 MySQL 远程访问
# 使用方法：将下面的命令复制到宝塔终端执行

echo "请输入 MySQL root 密码:" && read -s PASS && mysql -u root -p"$PASS" -e "DROP USER IF EXISTS 'xingyunpan'@'localhost'; DROP USER IF EXISTS 'xingyunpan'@'127.0.0.1'; DROP USER IF EXISTS 'xingyunpan'@'%'; CREATE USER 'xingyunpan'@'%' IDENTIFIED WITH mysql_native_password BY 'uKcgJLKyWHHJzm'; GRANT ALL PRIVILEGES ON xingyunpan.* TO 'xingyunpan'@'%'; FLUSH PRIVILEGES; SELECT user, host, plugin FROM mysql.user WHERE user='xingyunpan';" && echo "✅ 修复完成！现在测试连接：go run scripts/test-db-connection.go"

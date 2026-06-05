-- 星云盘远程数据库访问修复脚本
-- 在宝塔面板的 phpMyAdmin 中按顺序执行

-- ============================================
-- 步骤 1: 删除所有现有的 xingyunpan 用户
-- ============================================
DROP USER IF EXISTS 'xingyunpan'@'localhost';
DROP USER IF EXISTS 'xingyunpan'@'127.0.0.1';
DROP USER IF EXISTS 'xingyunpan'@'%';

-- ============================================
-- 步骤 2: 重新创建用户（允许所有 IP）
-- ============================================
CREATE USER 'xingyunpan'@'%' IDENTIFIED BY 'uKcgJLKyWHHJzm';

-- ============================================
-- 步骤 3: 授予所有权限
-- ============================================
GRANT ALL PRIVILEGES ON xingyunpan.* TO 'xingyunpan'@'%';

-- 如果需要授予所有数据库的权限（不推荐）
-- GRANT ALL PRIVILEGES ON *.* TO 'xingyunpan'@'%';

-- ============================================
-- 步骤 4: 刷新权限（非常重要！）
-- ============================================
FLUSH PRIVILEGES;

-- ============================================
-- 步骤 5: 验证用户和权限
-- ============================================
-- 查看用户列表
SELECT user, host, authentication_string FROM mysql.user WHERE user='xingyunpan';

-- 查看用户权限
SHOW GRANTS FOR 'xingyunpan'@'%';

-- ============================================
-- 可选：如果上面的方法不行，尝试使用旧的密码格式
-- ============================================
-- ALTER USER 'xingyunpan'@'%' IDENTIFIED WITH mysql_native_password BY 'uKcgJLKyWHHJzm';
-- FLUSH PRIVILEGES;

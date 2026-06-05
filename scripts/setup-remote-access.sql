-- 星云盘远程数据库访问配置 SQL 脚本
-- 在宝塔面板的 phpMyAdmin 或 MySQL 命令行中执行

-- ============================================
-- 方案 1: 允许所有 IP 访问（开发环境）
-- ============================================
-- 注意：这会允许任何 IP 连接，仅用于开发环境

-- 创建或更新用户，允许从任何 IP 连接
CREATE USER IF NOT EXISTS 'xingyunpan'@'%' IDENTIFIED BY 'uKcgJLKyWHHJzm';

-- 授予所有权限
GRANT ALL PRIVILEGES ON xingyunpan.* TO 'xingyunpan'@'%';

-- 刷新权限
FLUSH PRIVILEGES;

-- ============================================
-- 方案 2: 只允许特定 IP 访问（推荐）
-- ============================================
-- 替换下面的 IP 地址为你的实际 IP

-- 创建或更新用户，只允许从特定 IP 连接
-- CREATE USER IF NOT EXISTS 'xingyunpan'@'39.144.218.6' IDENTIFIED BY 'uKcgJLKyWHHJzm';

-- 授予所有权限
-- GRANT ALL PRIVILEGES ON xingyunpan.* TO 'xingyunpan'@'39.144.218.6';

-- 刷新权限
-- FLUSH PRIVILEGES;

-- ============================================
-- 方案 3: 允许 IP 段访问
-- ============================================
-- 允许 39.144.218.* 网段的所有 IP

-- CREATE USER IF NOT EXISTS 'xingyunpan'@'39.144.218.%' IDENTIFIED BY 'uKcgJLKyWHHJzm';
-- GRANT ALL PRIVILEGES ON xingyunpan.* TO 'xingyunpan'@'39.144.218.%';
-- FLUSH PRIVILEGES;

-- ============================================
-- 查看当前用户权限
-- ============================================
SELECT user, host FROM mysql.user WHERE user='xingyunpan';

-- ============================================
-- 删除旧的本地用户（可选）
-- ============================================
-- 如果你不需要本地连接，可以删除 localhost 用户
-- DROP USER IF EXISTS 'xingyunpan'@'localhost';
-- FLUSH PRIVILEGES;

-- 授予远程访问权限
-- 在宝塔面板的 phpMyAdmin 中执行此脚本

-- 方法 1：只允许你的 Windows IP（推荐）
GRANT ALL PRIVILEGES ON xingyunpan.* TO 'root'@'106.80.102.238' IDENTIFIED BY 'RMGdedeMoMnMc';

-- 方法 2：允许所有 IP（不推荐，仅用于测试）
-- GRANT ALL PRIVILEGES ON xingyunpan.* TO 'root'@'%' IDENTIFIED BY 'RMGdedeMoMnMc';

-- 刷新权限
FLUSH PRIVILEGES;

-- 查看当前权限
SELECT host, user FROM mysql.user WHERE user='root';

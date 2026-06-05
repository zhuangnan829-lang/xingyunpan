@echo off
setlocal

set "MYSQL_BASE=C:\Program Files\MySQL\MySQL Server 9.6"
set "DATA_DIR=D:\Desktop\xingyunpan\.local-mysql\data"
set "ERR_LOG=D:\Desktop\xingyunpan\logs\local-mysql-runtime.err"

start "" /min "%MYSQL_BASE%\bin\mysqld.exe" ^
  --basedir="%MYSQL_BASE%" ^
  --datadir="%DATA_DIR%" ^
  --port=3307 ^
  --bind-address=127.0.0.1 ^
  --mysqlx=0 ^
  --log-error="%ERR_LOG%"

endlocal

$ErrorActionPreference = "Stop"
$PSNativeCommandUseErrorActionPreference = $false

$serviceName = "MySQL96"
$mysqlBase = "C:\Program Files\MySQL\MySQL Server 9.6"
$mysqldExe = Join-Path $mysqlBase "bin\mysqld.exe"
$mysqlExe = Join-Path $mysqlBase "bin\mysql.exe"
$defaultsFile = "C:\ProgramData\MySQL\MySQL Server 9.6\my.ini"
$logsDir = "D:\Desktop\xingyunpan\logs"
$resetSql = Join-Path $logsDir "mysql-reset-root.sql"
$serverLog = Join-Path $logsDir "mysql-reset-server.log"
$newPassword = "XingyunpanRoot123!"

New-Item -ItemType Directory -Force -Path $logsDir | Out-Null
"ALTER USER 'root'@'localhost' IDENTIFIED BY '$newPassword'; FLUSH PRIVILEGES;" |
  Set-Content -LiteralPath $resetSql -Encoding ascii

function Test-MySqlLogin {
  param(
    [string]$Password
  )

  $output = & $mysqlExe '--protocol=tcp' '-h127.0.0.1' '-P3306' '-uroot' "-p$Password" '-e' 'SELECT VERSION();' 2>&1
  return $LASTEXITCODE -eq 0
}

Write-Host "Stopping $serviceName..."
Stop-Service $serviceName -Force
Start-Sleep -Seconds 8

Write-Host "Starting mysqld with init-file..."
$proc = Start-Process -FilePath $mysqldExe `
  -ArgumentList @(
    "--defaults-file=$defaultsFile",
    "--init-file=$resetSql",
    "--console",
    "--log-error=$serverLog"
  ) `
  -PassThru

$loginOk = $false
for ($i = 0; $i -lt 15; $i++) {
  Start-Sleep -Seconds 2
  if (Test-MySqlLogin -Password $newPassword) {
    $loginOk = $true
    break
  }
}

if (-not $loginOk) {
  if (-not $proc.HasExited) {
    Stop-Process -Id $proc.Id -Force
  }
  throw "Password reset did not verify. Check $serverLog"
}

Write-Host "Stopping temporary mysqld..."
if (-not $proc.HasExited) {
  Stop-Process -Id $proc.Id -Force
  Start-Sleep -Seconds 5
}

Write-Host "Starting $serviceName..."
Start-Service $serviceName
Start-Sleep -Seconds 10

& $mysqlExe '--protocol=tcp' '-h127.0.0.1' '-P3306' '-uroot' "-p$newPassword" '-e' 'SELECT VERSION() AS version, CURRENT_USER() AS current_user;'

Write-Host ""
Write-Host "MySQL root password has been reset to: $newPassword"

@echo off
setlocal
PowerShell -ExecutionPolicy Bypass -File "%~dp0reset-mysql-root.ps1"
pause
endlocal

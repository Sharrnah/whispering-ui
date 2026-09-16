@echo off
powershell.exe -NoProfile -ExecutionPolicy Bypass -File "%~dp0BuildTools\package-windows.ps1"
exit /b %errorlevel%

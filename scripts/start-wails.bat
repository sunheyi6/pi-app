@echo off
cd /d "%~dp0.."
powershell -ExecutionPolicy Bypass -NoProfile -File "scripts\restart-wails.ps1" -StartOnly
if errorlevel 1 (
    echo 启动失败，请查看 wails-dev.log 和 wails-dev.err.log。
    pause
)

@echo off
setlocal
chcp 65001 >nul
title Biên dịch MikroTik Automation Tool (Production)
powershell -NoProfile -ExecutionPolicy Bypass -File "%~dp0build.ps1"
if %errorlevel% neq 0 (
    echo.
    echo [LỖI] Quá trình biên dịch thất bại!
    pause
)

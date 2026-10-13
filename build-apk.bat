@echo off
chcp 65001 >nul
title MikroTik Automation Tool - Android APK Packager
powershell -NoProfile -ExecutionPolicy Bypass -File "%~dp0build-apk.ps1"
pause

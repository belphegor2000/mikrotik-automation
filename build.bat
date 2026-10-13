@echo off
setlocal
title Bien dich Mikrotik Automation Tool (Golang)

echo ========================================================
echo   BAT DAU BIEN DICH MIKROTIK AUTOMATION TOOL (GOLANG)
echo ========================================================
echo.

set "PATH=C:\Program Files\Go\bin;C:\Users\WhitePoplar\AppData\Local\Microsoft\WinGet\Packages\BrechtSanders.WinLibs.POSIX.UCRT_Microsoft.Winget.Source_8wekyb3d8bbwe\mingw64\bin;%PATH%"
set CGO_ENABLED=1

echo [1/2] Kiem tra trinh bien dich...
where go >nul 2>nul
if %errorlevel% neq 0 (
    echo [LOI] Khong tim thay Go!
    pause
    exit /b 1
)

where gcc >nul 2>nul
if %errorlevel% neq 0 (
    echo [LOI] Khong tim thay GCC! Fyne can GCC de bien dich CGo.
    pause
    exit /b 1
)

echo.
echo [2/2] Dang bien dich mikrotik-automation.exe...
echo       (Cac co toi uu: -H=windowsgui -s -w)
go build -ldflags "-H=windowsgui -s -w" -o mikrotik-automation.exe ./cmd/mikrotik-app

if %errorlevel% equ 0 (
    echo.
    echo ========================================================
    echo   [THANH CONG] Da tao file mikrotik-automation.exe!
    echo ========================================================
    echo Ban co the nhap dup chuot vao mikrotik-automation.exe de su dung.
) else (
    echo.
    echo [THAT BAI] Qua trinh bien dich gap loi!
)

echo.
pause


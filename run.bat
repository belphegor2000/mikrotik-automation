@echo off
if exist "%~dp0mikrotik-automation.exe" (
    start "" "%~dp0mikrotik-automation.exe"
    exit
)
set "PATH=C:\Program Files\Go\bin;C:\Users\WhitePoplar\AppData\Local\Microsoft\WinGet\Packages\BrechtSanders.WinLibs.POSIX.UCRT_Microsoft.Winget.Source_8wekyb3d8bbwe\mingw64\bin;%PATH%"
set CGO_ENABLED=1
echo Dang bien dich va khoi dong ung dung...
go build -ldflags "-H=windowsgui -s -w" -o "%~dp0mikrotik-automation.exe" "%~dp0"
start "" "%~dp0mikrotik-automation.exe"
exit

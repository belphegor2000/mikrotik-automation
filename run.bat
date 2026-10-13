@echo off
set "PATH=C:\Program Files\Go\bin;C:\Users\WhitePoplar\AppData\Local\Microsoft\WinGet\Packages\BrechtSanders.WinLibs.POSIX.UCRT_Microsoft.Winget.Source_8wekyb3d8bbwe\mingw64\bin;%PATH%"
set CGO_ENABLED=1
echo Dang khoi dong Mikrotik Automation Tool bang Go...
go run ./cmd/mikrotik-app

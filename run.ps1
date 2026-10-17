# Script chạy ứng dụng không hiện cửa sổ console
$ExePath = Join-Path $PSScriptRoot "mikrotik-automation.exe"
if (Test-Path $ExePath) {
    Start-Process $ExePath
    exit
}
$env:Path = "C:\Program Files\Go\bin;C:\Users\WhitePoplar\AppData\Local\Microsoft\WinGet\Packages\BrechtSanders.WinLibs.POSIX.UCRT_Microsoft.Winget.Source_8wekyb3d8bbwe\mingw64\bin;" + $env:Path
$env:CGO_ENABLED = "1"
Write-Host "Dang bien dich va khoi dong ung dung..." -ForegroundColor Cyan
go build -ldflags "-H=windowsgui -s -w" -o $ExePath $PSScriptRoot
Start-Process $ExePath

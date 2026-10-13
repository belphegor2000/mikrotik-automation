# Script chạy thử nghiệm ứng dụng trực tiếp bằng Go (không cần build)
$env:Path = "C:\Program Files\Go\bin;C:\Users\WhitePoplar\AppData\Local\Microsoft\WinGet\Packages\BrechtSanders.WinLibs.POSIX.UCRT_Microsoft.Winget.Source_8wekyb3d8bbwe\mingw64\bin;" + $env:Path
$env:CGO_ENABLED = "1"

Write-Host "Dang khoi dong Mikrotik Automation Tool bang Go..." -ForegroundColor Cyan
go run .

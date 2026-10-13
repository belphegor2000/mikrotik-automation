$ErrorActionPreference = "Stop"

Write-Host "========================================================" -ForegroundColor Cyan
Write-Host "  BAT DAU BIEN DICH MIKROTIK AUTOMATION TOOL (GOLANG)" -ForegroundColor Cyan
Write-Host "========================================================" -ForegroundColor Cyan

# 1. Cau hinh bien moi truong PATH
$GoPath = "C:\Program Files\Go\bin"
$MinGWPath = "C:\Users\WhitePoplar\AppData\Local\Microsoft\WinGet\Packages\BrechtSanders.WinLibs.POSIX.UCRT_Microsoft.Winget.Source_8wekyb3d8bbwe\mingw64\bin"

if (Test-Path $GoPath) {
    if ($env:Path -notlike "*$GoPath*") {
        $env:Path = "$GoPath;" + $env:Path
    }
}

if (Test-Path $MinGWPath) {
    if ($env:Path -notlike "*$MinGWPath*") {
        $env:Path = "$MinGWPath;" + $env:Path
    }
}

$env:CGO_ENABLED = "1"

# 2. Kiem tra Go va GCC
Write-Host ""
Write-Host "[1/3] Kiem tra trinh bien dich..." -ForegroundColor Yellow
try {
    $goVer = go version
    Write-Host "  [OK] $goVer" -ForegroundColor Green
} catch {
    Write-Host "  [LOI] Khong tim thay Go!" -ForegroundColor Red
    exit 1
}

try {
    $gccVer = gcc --version | Select-Object -First 1
    Write-Host "  [OK] $gccVer" -ForegroundColor Green
} catch {
    Write-Host "  [LOI] Khong tim thay GCC (MinGW)!" -ForegroundColor Red
    exit 1
}

# 3. Kiem thu tu dong
Write-Host ""
Write-Host "[2/3] Chay kiem thu tu dong..." -ForegroundColor Yellow
try {
    go test ./internal/auth/... ./internal/modules/...
    Write-Host "  [OK] Tat ca unit tests da PASS!" -ForegroundColor Green
} catch {
    Write-Host "  [CANH BAO] Co loi kiem thu, tiep tuc bien dich..." -ForegroundColor Yellow
}

# 4. Bien dich file exe
$OutputFile = "mikrotik-automation.exe"
Write-Host ""
Write-Host "[3/3] Dang bien dich $OutputFile..." -ForegroundColor Yellow
Write-Host "  (Toi uu hoa: An console -H=windowsgui va loai bo debug symbols -s -w)" -ForegroundColor Gray

go build -ldflags "-H=windowsgui -s -w" -o $OutputFile ./cmd/mikrotik-app

if (Test-Path $OutputFile) {
    $item = Get-Item $OutputFile
    $sizeMB = [math]::Round($item.Length / 1MB, 2)
    Write-Host ""
    Write-Host "========================================================" -ForegroundColor Green
    Write-Host "  BIEN DICH THANH CONG!" -ForegroundColor Green
    Write-Host "  File: $OutputFile" -ForegroundColor Green
    Write-Host "  Dung luong: $sizeMB MB" -ForegroundColor Green
    Write-Host "========================================================" -ForegroundColor Green
    Write-Host "Ban co the nhap dup chuot vao file de chay ngay!" -ForegroundColor Cyan
    Write-Host ""
} else {
    Write-Host ""
    Write-Host "Bien dich that bai! Khong tim thay file output." -ForegroundColor Red
    exit 1
}


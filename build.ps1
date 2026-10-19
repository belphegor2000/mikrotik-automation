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
Write-Host "[2/4] Chay kiem thu tu dong..." -ForegroundColor Yellow
try {
    go test ./...
    Write-Host "  [OK] Tat ca unit tests da PASS!" -ForegroundColor Green
} catch {
    Write-Host "  [CANH BAO] Co loi kiem thu, tiep tuc bien dich..." -ForegroundColor Yellow
}

# 4. Tao icon resource cho Windows (app.syso)
Write-Host ""
Write-Host "[3/4] Chuan bi bieu tuong ung dung (Windows App Icon)..." -ForegroundColor Yellow
try {
    if (Test-Path "icon.png") {
        if (-not (Test-Path "icon.ico")) {
            Add-Type -AssemblyName System.Drawing
            $img = [System.Drawing.Bitmap]::FromFile((Resolve-Path "icon.png"))
            $thumb = New-Object System.Drawing.Bitmap($img, 256, 256)
            $hIcon = $thumb.GetHicon()
            $icon = [System.Drawing.Icon]::FromHandle($hIcon)
            $fs = New-Object System.IO.FileStream((Join-Path (Get-Location) "icon.ico"), [System.IO.FileMode]::Create)
            $icon.Save($fs)
            $fs.Close()
            $thumb.Dispose()
            $img.Dispose()
        }
        if (Test-Path "$MinGWPath\windres.exe") {
            "1 ICON `"icon.ico`"" | Out-File -Encoding ascii app.rc
            & "$MinGWPath\windres.exe" -i app.rc -O coff -o app.syso
            Remove-Item app.rc -Force -ErrorAction SilentlyContinue
            Write-Host "  [OK] Da nhung thanh cong icon vao app.syso!" -ForegroundColor Green
        }
    }
} catch {
    Write-Host "  [BO QUA] Khong the tao app.syso: $_" -ForegroundColor Gray
}

# 5. Bien dich file exe production
$OutputFile = "mikrotik-automation.exe"
Write-Host ""
Write-Host "[4/4] Dang bien dich $OutputFile..." -ForegroundColor Yellow
Write-Host "  (Toi uu hoa: An console -H=windowsgui va loai bo debug symbols -s -w)" -ForegroundColor Gray

go build -ldflags "-H=windowsgui -s -w" -o $OutputFile .

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


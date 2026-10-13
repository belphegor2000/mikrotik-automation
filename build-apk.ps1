# ==============================================================================
# MikroTik Automation Tool - Kich Ban Dong Goi APK (Samsung Galaxy A56 / ARM64)
# ==============================================================================
param (
    [string]$TargetArch = "arm64", # Samsung Galaxy A56 su dung chip Exynos 1580 (ARM64-v8a)
    [string]$AppID = "com.mikrotik.automation",
    [string]$AppName = "MikroTik Auto",
    [string]$IconPath = "icon.png"
)

$ErrorActionPreference = "Continue"

Write-Host "=================================================================" -ForegroundColor Cyan
Write-Host "   MIKROTIK AUTOMATION TOOL - ANDROID APK PACKAGER (ARM64)       " -ForegroundColor Yellow
Write-Host "   Thiet bi muc tieu: Samsung Galaxy A56 (Android 15, One UI 7) " -ForegroundColor Green
Write-Host "=================================================================" -ForegroundColor Cyan
Write-Host ""

# 1. Kiem tra Go Compiler
Write-Host "[1/5] Kiem tra moi truong Go..." -ForegroundColor White
$goVersion = & go version 2>$null
if (-not $goVersion) {
    Write-Host "[-] Loi: Khong tim thay Go! Vui long cai dat Go 1.22+." -ForegroundColor Red
    exit 1
}
Write-Host " [+] Go: $goVersion" -ForegroundColor Green

# 2. Kiem tra Fyne CLI Tool
Write-Host "[2/5] Kiem tra cong cu Fyne CLI..." -ForegroundColor White
$fyneCmd = Get-Command fyne -ErrorAction SilentlyContinue
if (-not $fyneCmd) {
    $gopath = & go env GOPATH
    $candidate = Join-Path $gopath "bin\fyne.exe"
    if (Test-Path $candidate) {
        $env:Path += ";$gopath\bin"
        $fyneCmd = Get-Command fyne -ErrorAction SilentlyContinue
    }
}

if (-not $fyneCmd) {
    Write-Host " [*] Thu kiem tra/cai dat Fyne CLI qua go install..." -ForegroundColor Yellow
    try {
        & go install fyne.io/tools/cmd/fyne@latest 2>$null
        $gopath = & go env GOPATH
        $env:Path += ";$gopath\bin"
        $fyneCmd = Get-Command fyne -ErrorAction SilentlyContinue
    } catch {
        # Bo qua loi mang
    }
}

if ($fyneCmd) {
    Write-Host " [+] Fyne CLI da san sang: $($fyneCmd.Source)" -ForegroundColor Green
} else {
    Write-Host " [!] Chua co Fyne CLI trong PATH. Khi co internet, ban co the chay: go install fyne.io/tools/cmd/fyne@latest" -ForegroundColor Yellow
}

# 3. Kiem tra Java JDK 17
Write-Host "[3/5] Kiem tra Java JDK..." -ForegroundColor White
$javaCmd = Get-Command java -ErrorAction SilentlyContinue
if ($javaCmd) {
    Write-Host " [+] Java JDK da san sang: $($javaCmd.Source)" -ForegroundColor Green
} else {
    Write-Host " [!] Canh bao: Chua phat hien Java JDK tren may tinh." -ForegroundColor Yellow
    Write-Host "     De cai dat Java JDK 17 tu dong tren Windows, chay lenh: winget install Microsoft.OpenJDK.17" -ForegroundColor Cyan
}

# 4. Kiem tra Android SDK & NDK (gomobile backend)
Write-Host "[4/5] Kiem tra Android SDK & NDK..." -ForegroundColor White
if ($env:ANDROID_HOME -or $env:ANDROID_NDK_HOME) {
    Write-Host " [+] ANDROID_HOME: $env:ANDROID_HOME" -ForegroundColor Green
    Write-Host " [+] ANDROID_NDK_HOME: $env:ANDROID_NDK_HOME" -ForegroundColor Green
} else {
    Write-Host " [!] Chua thiet lap bien moi truong ANDROID_HOME hoac ANDROID_NDK_HOME." -ForegroundColor Yellow
    Write-Host "     Goi y:" -ForegroundColor Gray
    Write-Host "     - Cach 1: Cai dat Android Commandline Tools & NDK va set bien moi truong tren may." -ForegroundColor Gray
    Write-Host "     - Cach 2 (Khuyen dung): Su dung GitHub Actions (.github/workflows/build-apk.yml) de build tren Cloud mien phi ma khong can tai 5-8GB SDK ve may tinh." -ForegroundColor Cyan
}

# 5. Thuc hien dong goi APK
Write-Host "[5/5] Tien hanh dong goi file APK cho Samsung Galaxy A56..." -ForegroundColor White

if (-not (Test-Path $IconPath)) {
    Write-Host " [-] Khong tim thay file icon: $IconPath" -ForegroundColor Red
    exit 1
}

if (-not $fyneCmd) {
    Write-Host " [-] Chua co cong cu Fyne CLI. Vui long cai dat Fyne CLI hoac su dung GitHub Actions de build APK." -ForegroundColor Red
    exit 0
}

$buildTarget = "android/$TargetArch"
Write-Host " [*] Thuc thi lenh: fyne package -os $buildTarget -app-id $AppID -name `"$AppName`" -icon `"$IconPath`"" -ForegroundColor Cyan

try {
    & fyne package -os $buildTarget -app-id $AppID -name "$AppName" -icon "$IconPath"
    if ($LASTEXITCODE -eq 0) {
        Write-Host ""
        Write-Host "=================================================================" -ForegroundColor Green
        Write-Host "   BIEN DICH APK THANH CONG CHO SAMSUNG GALAXY A56!              " -ForegroundColor Green
        Write-Host "=================================================================" -ForegroundColor Green
        Get-ChildItem -Filter "*.apk" | ForEach-Object {
            $mb = [math]::Round($_.Length / 1048576, 2)
            Write-Host " [+] File APK: $($_.FullName) ($mb MB)" -ForegroundColor Yellow
        }
        Write-Host ""
        Write-Host "Cach cai dat len dien thoai Samsung Galaxy A56:" -ForegroundColor Cyan
        Write-Host "1. Copy file .apk vao thu muc Download tren dien thoai qua cap USB hoac Google Drive/Zalo." -ForegroundColor White
        Write-Host "2. Mo ung dung 'File cua ban' (My Files) tren Samsung A56 va bam vao file de Cai dat." -ForegroundColor White
    } else {
        Write-Host ""
        Write-Host "[-] Qua trinh build offline can co Android NDK va gomobile tren may tinh." -ForegroundColor Yellow
        Write-Host "    Ban co the dung GitHub Actions trong .github/workflows/build-apk.yml de build tu dong." -ForegroundColor Cyan
    }
} catch {
    Write-Host "[-] Loi thuc thi: $_" -ForegroundColor Red
}

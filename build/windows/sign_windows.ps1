$ErrorActionPreference = "Stop"

# assumes CI exported required env vars (bash script sources build/buildvars.env with export)

if (-not $env:SIGN_WINDOWS_CERT) { throw "SIGN_WINDOWS_CERT not set" }
if (-not $env:SIGN_WINDOWS_CERT_PASSWORD) { throw "SIGN_WINDOWS_CERT_PASSWORD not set" }
if (-not $env:COMMON_BUILD_NAME) { throw "COMMON_BUILD_NAME not set" }

$certDir = "certificate"
New-Item -ItemType Directory -Force -Path $certDir | Out-Null

$b64Path = Join-Path $certDir "certificate.txt"
$pfxPath = Join-Path $certDir "certificate.pfx"
Set-Content -Path $b64Path -Value $env:SIGN_WINDOWS_CERT -NoNewline -Encoding Ascii
certutil -decode $b64Path $pfxPath | Out-Null

# Try to locate signtool.exe in typical Windows Kits locations
$signtoolCandidates = @(
  "C:\Program Files (x86)\Windows Kits\10\bin\x64\signtool.exe",
  "C:\Program Files (x86)\Windows Kits\10\bin\x86\signtool.exe"
)

$kitsRoot = "C:\Program Files (x86)\Windows Kits\10\bin"
if (Test-Path $kitsRoot) {
  # Also scan versioned subfolders, e.g. 10.0.22621.0\x64\signtool.exe
  Get-ChildItem -Path $kitsRoot -Directory -ErrorAction SilentlyContinue |
    Sort-Object Name -Descending |
    ForEach-Object {
      foreach ($arch in @("x64","x86")) {
        $candidate = Join-Path $_.FullName "$arch\signtool.exe"
        $signtoolCandidates += $candidate
      }
    }
}

$signtool = $signtoolCandidates | Where-Object { Test-Path $_ } | Select-Object -First 1
if (-not $signtool) {
  throw "signtool.exe not found in Windows Kits paths"
}

$appExe = "build/bin/$($env:COMMON_BUILD_NAME).exe"
$installerExe = $env:WIN_INSTALLER_PATH
if (-not $installerExe) {
  $installerExe = "build/bin/$($env:COMMON_BUILD_NAME)-amd64-installer.exe"
}

if (-not (Test-Path $appExe)) { throw "App exe not found: $appExe" }
if (-not (Test-Path $installerExe)) { throw "Installer exe not found: $installerExe" }

Write-Host "Signing app exe: $appExe"
& $signtool sign /fd sha256 /tr http://ts.ssl.com /td sha256 /f $pfxPath /p $env:SIGN_WINDOWS_CERT_PASSWORD $appExe

Write-Host "Signing installer exe: $installerExe"
& $signtool sign /fd sha256 /tr http://ts.ssl.com /td sha256 /f $pfxPath /p $env:SIGN_WINDOWS_CERT_PASSWORD $installerExe

Write-Host "Signing done."

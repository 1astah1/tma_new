$ErrorActionPreference = "Stop"
$root = Split-Path $PSScriptRoot -Parent

& (Join-Path $PSScriptRoot "stop-dev.ps1")

Write-Host "Starting Postgres..."
docker start tma_postgres 2>$null
Start-Sleep -Seconds 2

$backendDir = Join-Path $root "tma-backend"
Set-Location $backendDir

Write-Host "Building backend..."
go build -o server.exe ./cmd/server
if ($LASTEXITCODE -ne 0) { throw "go build failed" }

Write-Host "Starting backend on http://localhost:8081 ..."
$env:SERVER_PORT = "8081"
Start-Process -FilePath ".\server.exe" -WorkingDirectory $backendDir -WindowStyle Hidden

$deadline = (Get-Date).AddSeconds(20)
do {
  Start-Sleep -Seconds 1
  try {
    Invoke-RestMethod "http://localhost:8081/health" -TimeoutSec 2 | Out-Null
    break
  } catch {
    if ((Get-Date) -gt $deadline) { throw "Backend did not start in 20s" }
  }
} while ($true)

Write-Host "Starting admin on http://localhost:5174 ..."
Start-Process powershell -ArgumentList @(
  '-NoExit', '-Command',
  "Set-Location '$root\admin-panel'; npm run dev"
) -WindowStyle Minimized

Write-Host "Starting TMA on http://localhost:5173 ..."
Start-Process powershell -ArgumentList @(
  '-NoExit', '-Command',
  "Set-Location '$root\tma-frontend'; npm run dev"
) -WindowStyle Minimized

$deadline = (Get-Date).AddSeconds(30)
do {
  Start-Sleep -Seconds 2
  try {
    Invoke-WebRequest "http://localhost:5173/" -UseBasicParsing -TimeoutSec 3 | Out-Null
    Invoke-WebRequest "http://localhost:5174/" -UseBasicParsing -TimeoutSec 3 | Out-Null
    break
  } catch {
    if ((Get-Date) -gt $deadline) { throw "Frontends did not start in 30s" }
  }
} while ($true)

Write-Host ""
Write-Host "Running verification..."
& (Join-Path $PSScriptRoot "verify-dev.ps1")
if ($LASTEXITCODE -ne 0) { exit 1 }

$lanIp = (Get-NetIPAddress -AddressFamily IPv4 |
  Where-Object { $_.IPAddress -notlike '127.*' -and $_.PrefixOrigin -ne 'WellKnown' } |
  Select-Object -First 1).IPAddress

Write-Host ""
Write-Host "Ready:"
Write-Host "  TMA (PC):      http://localhost:5173"
if ($lanIp) {
  Write-Host "  TMA (phone):   http://${lanIp}:5173  (for Telegram on same Wi-Fi)"
}
Write-Host "  Admin:         http://localhost:5174"
Write-Host "  API:           http://localhost:8081/health"

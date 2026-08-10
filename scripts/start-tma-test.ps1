$ErrorActionPreference = "Stop"
$root = Split-Path $PSScriptRoot -Parent
$backendDir = Join-Path $root "tma-backend"
$envFile = Join-Path $backendDir ".env"

function Read-EnvValue([string]$key) {
  if (-not (Test-Path $envFile)) { return $null }
  foreach ($line in Get-Content $envFile) {
    if ($line -match "^\s*$key\s*=\s*(.+)$") {
      return $Matches[1].Trim().Trim('"').Trim("'")
    }
  }
  return $null
}

function Set-EnvValue([string]$key, [string]$value) {
  $lines = @()
  $found = $false
  if (Test-Path $envFile) {
    foreach ($line in Get-Content $envFile) {
      if ($line -match "^\s*$key\s*=") {
        $lines += "$key=$value"
        $found = $true
      } else {
        $lines += $line
      }
    }
  }
  if (-not $found) { $lines += "$key=$value" }
  Set-Content -Path $envFile -Value $lines -Encoding UTF8
}

function Wait-HttpOk([string]$url, [int]$seconds = 30) {
  $deadline = (Get-Date).AddSeconds($seconds)
  do {
    try {
      Invoke-WebRequest $url -UseBasicParsing -TimeoutSec 3 | Out-Null
      return
    } catch {
      if ((Get-Date) -gt $deadline) { throw "Timeout waiting for $url" }
      Start-Sleep -Seconds 1
    }
  } while ($true)
}

function Stop-PortListener([int]$port) {
  Get-NetTCPConnection -LocalPort $port -State Listen -ErrorAction SilentlyContinue |
    ForEach-Object { Stop-Process -Id $_.OwningProcess -Force -ErrorAction SilentlyContinue }
}

Write-Host "=== TMA local test setup ===" -ForegroundColor Cyan

# Postgres
Write-Host "Starting Postgres..."
docker start tma_postgres 2>$null | Out-Null
$dbUrl = Read-EnvValue "DATABASE_URL"
if (-not $dbUrl) {
  $dbUrl = "postgres://postgres:postgres@localhost:5434/tma_shop?sslmode=disable"
}

# Backend on 8081 (matches Vite proxy)
$backendPort = Read-EnvValue "SERVER_PORT"
if (-not $backendPort) { $backendPort = "8081" }

if (-not (Test-Path (Join-Path $backendDir "server.exe"))) {
  Write-Host "Building backend..."
  Push-Location $backendDir
  go build -o server.exe ./cmd/server
  if ($LASTEXITCODE -ne 0) { throw "go build failed" }
  Pop-Location
}

$backendListening = $false
try {
  Invoke-RestMethod "http://localhost:$backendPort/health" -TimeoutSec 2 | Out-Null
  $backendListening = $true
} catch {}

if (-not $backendListening) {
  Write-Host "Starting backend on http://localhost:$backendPort ..."
  Get-Process -Name server -ErrorAction SilentlyContinue | Stop-Process -Force
  $env:DATABASE_URL = $dbUrl
  $env:JWT_SECRET = (Read-EnvValue "JWT_SECRET")
  if (-not $env:JWT_SECRET) { $env:JWT_SECRET = "super-secret-key-min-32-chars-long!!" }
  $env:ACCOUNT_ENCRYPTION_KEY = (Read-EnvValue "ACCOUNT_ENCRYPTION_KEY")
  if (-not $env:ACCOUNT_ENCRYPTION_KEY) { $env:ACCOUNT_ENCRYPTION_KEY = "0123456789abcdef0123456789abcdef" }
  $env:SERVER_PORT = $backendPort
  $env:BOT_TOKEN = (Read-EnvValue "BOT_TOKEN")
  $env:TMA_URL = (Read-EnvValue "TMA_URL")
  if (-not $env:TMA_URL) { $env:TMA_URL = "http://localhost:5173" }
  $env:ADMIN_PANEL_URL = (Read-EnvValue "ADMIN_PANEL_URL")
  if (-not $env:ADMIN_PANEL_URL) { $env:ADMIN_PANEL_URL = "http://localhost:5174" }
  $env:UPLOAD_DIR = (Read-EnvValue "UPLOAD_DIR")
  if (-not $env:UPLOAD_DIR) { $env:UPLOAD_DIR = "./uploads" }
  $env:ENVIRONMENT = "development"
  Start-Process -FilePath (Join-Path $backendDir "server.exe") -WorkingDirectory $backendDir -WindowStyle Hidden
  Wait-HttpOk "http://localhost:$backendPort/health"
}

# TMA frontend
$tmaListening = $false
try {
  Invoke-WebRequest "http://localhost:5173/" -UseBasicParsing -TimeoutSec 2 | Out-Null
  $tmaListening = $true
} catch {}

if (-not $tmaListening) {
  Write-Host "Starting TMA on http://localhost:5173 ..."
  Stop-PortListener 5173
  Start-Process powershell -ArgumentList @(
    '-NoExit', '-Command',
    "Set-Location '$root\tma-frontend'; npm run dev"
  ) -WindowStyle Minimized
  Wait-HttpOk "http://localhost:5173/"
}

# ngrok tunnel
Write-Host "Starting ngrok tunnel to port 5173..."
Get-Process -Name ngrok -ErrorAction SilentlyContinue | Stop-Process -Force
Start-Process -FilePath "ngrok" -ArgumentList "http", "5173", "--log=stdout" -WindowStyle Hidden
Start-Sleep -Seconds 3

$tunnels = Invoke-RestMethod "http://127.0.0.1:4040/api/tunnels" -TimeoutSec 5
$publicUrl = ($tunnels.tunnels | Where-Object { $_.proto -eq "https" } | Select-Object -First 1).public_url
if (-not $publicUrl) { throw "ngrok did not return a public URL. Check ngrok auth token." }

Write-Host ""
Write-Host "Public TMA URL: $publicUrl" -ForegroundColor Green

# Update TMA_URL in .env and restart backend with new URL
Set-EnvValue "TMA_URL" $publicUrl
Write-Host "Updated TMA_URL in .env"

Get-Process -Name server -ErrorAction SilentlyContinue | Stop-Process -Force
Start-Sleep -Seconds 1

$env:DATABASE_URL = $dbUrl
$env:JWT_SECRET = (Read-EnvValue "JWT_SECRET")
if (-not $env:JWT_SECRET) { $env:JWT_SECRET = "super-secret-key-min-32-chars-long!!" }
$env:ACCOUNT_ENCRYPTION_KEY = (Read-EnvValue "ACCOUNT_ENCRYPTION_KEY")
if (-not $env:ACCOUNT_ENCRYPTION_KEY) { $env:ACCOUNT_ENCRYPTION_KEY = "0123456789abcdef0123456789abcdef" }
$env:SERVER_PORT = $backendPort
$env:BOT_TOKEN = (Read-EnvValue "BOT_TOKEN")
$env:TMA_URL = $publicUrl
$env:ADMIN_PANEL_URL = (Read-EnvValue "ADMIN_PANEL_URL")
if (-not $env:ADMIN_PANEL_URL) { $env:ADMIN_PANEL_URL = "http://localhost:5174" }
$env:UPLOAD_DIR = (Read-EnvValue "UPLOAD_DIR")
if (-not $env:UPLOAD_DIR) { $env:UPLOAD_DIR = "./uploads" }
$env:ENVIRONMENT = "development"
Start-Process -FilePath (Join-Path $backendDir "server.exe") -WorkingDirectory $backendDir -WindowStyle Hidden
Wait-HttpOk "http://localhost:$backendPort/health"

# Set Telegram menu button if BOT_TOKEN is configured
$botToken = Read-EnvValue "BOT_TOKEN"
if ($botToken) {
  $body = @{
    menu_button = @{
      type = "web_app"
      text = "Открыть магазин"
      web_app = @{ url = $publicUrl }
    }
  } | ConvertTo-Json -Depth 5 -Compress
  try {
    Invoke-RestMethod -Method POST `
      -Uri "https://api.telegram.org/bot$botToken/setChatMenuButton" `
      -ContentType "application/json" `
      -Body $body | Out-Null
    Write-Host "Telegram menu button updated." -ForegroundColor Green
  } catch {
    Write-Host "Could not set menu button automatically. Set manually in @BotFather:" -ForegroundColor Yellow
    Write-Host "  /setmenubutton -> your bot -> Web App -> $publicUrl"
  }
} else {
  Write-Host "BOT_TOKEN not set. Configure in BotFather:" -ForegroundColor Yellow
  Write-Host "  /setmenubutton -> your bot -> Web App -> $publicUrl"
}

Write-Host ""
Write-Host "Ready for TMA testing:" -ForegroundColor Cyan
Write-Host "  Telegram:  open your bot -> menu button /start"
Write-Host "  Local:     http://localhost:5173"
Write-Host "  Public:    $publicUrl"
Write-Host "  API:       http://localhost:$backendPort/health"
Write-Host "  ngrok UI:  http://127.0.0.1:4040"

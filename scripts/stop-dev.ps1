$ErrorActionPreference = "SilentlyContinue"

Write-Host "Stopping dev processes..."

foreach ($port in @(8080, 8081, 5173, 5174)) {
  Get-NetTCPConnection -LocalPort $port -State Listen -ErrorAction SilentlyContinue |
    ForEach-Object { Stop-Process -Id $_.OwningProcess -Force -ErrorAction SilentlyContinue }
}

Get-Process -Name server,tma_server -ErrorAction SilentlyContinue | Stop-Process -Force

Start-Sleep -Seconds 1
Write-Host "Done. Postgres (docker) left running on port 5434."

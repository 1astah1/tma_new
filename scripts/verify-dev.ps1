$ErrorActionPreference = "Stop"
$failed = 0

function Test-Ok($name, $scriptBlock) {
  try {
    & $scriptBlock
    Write-Host "  OK  $name" -ForegroundColor Green
  } catch {
    Write-Host "  FAIL $name - $($_.Exception.Message)" -ForegroundColor Red
    $script:failed++
  }
}

Write-Host "Verifying dev stack..."

Test-Ok "backend health" {
  $h = Invoke-RestMethod "http://localhost:8081/health" -TimeoutSec 5
  if ($h.status -ne "ok") { throw "status=$($h.status)" }
}

Test-Ok "auth (dev bypass)" {
  $auth = Invoke-RestMethod -Method POST -Uri "http://localhost:8081/api/v1/auth/telegram" `
    -ContentType "application/json" -Body '{"initData":"test"}' -TimeoutSec 5
  if (-not $auth.token) { throw "no token" }
}

Test-Ok "preorders API" {
  $r = Invoke-RestMethod "http://localhost:8081/api/v1/products?section=preorder&type=game&limit=5" -TimeoutSec 10
  if ($r.data.Count -lt 1) { throw "empty (total=$($r.meta.total))" }
}

Test-Ok "new releases API" {
  $r = Invoke-RestMethod "http://localhost:8081/api/v1/products?section=new&type=game&limit=5" -TimeoutSec 10
  if ($r.data.Count -lt 1) { throw "empty (total=$($r.meta.total))" }
}

Test-Ok "popular API" {
  $r = Invoke-RestMethod "http://localhost:8081/api/v1/content/popular-products" -TimeoutSec 10
  if ($r.data.Count -lt 1) { throw "empty" }
}

Test-Ok "home-feed API" {
  $r = Invoke-RestMethod "http://localhost:8081/api/v1/content/home-feed" -TimeoutSec 10
  if ($r.data.preorders.Count -lt 1 -and $r.data.new_releases.Count -lt 1) { throw "empty feed" }
  if ($r.data.popular.Count -lt 1) { throw "empty popular" }
}

Test-Ok "no fake 1 RUB games" {
  $r = Invoke-RestMethod "http://localhost:8081/api/v1/products?type=game&status=active&limit=100" -TimeoutSec 10
  $cheap = @($r.data | Where-Object { $_.price -le 1.01 -and $_.game_section -ne 'preorder' })
  if ($cheap.Count -gt 0) { throw "found $($cheap.Count) active games at 1 RUB" }
}

Test-Ok "no invalid release years" {
  $r = Invoke-RestMethod "http://localhost:8081/api/v1/products?type=game&status=active&section=preorder&limit=50" -TimeoutSec 10
  foreach ($p in $r.data) {
    if ($p.release_date -and $p.release_date -like '9998*') { throw "invalid date on $($p.title)" }
  }
}

Test-Ok "TMA page" {
  Invoke-WebRequest "http://localhost:5173/" -UseBasicParsing -TimeoutSec 10 | Out-Null
}

Test-Ok "TMA proxy -> home-feed" {
  $r = Invoke-RestMethod "http://localhost:5173/api/v1/content/home-feed" -TimeoutSec 10
  if ($r.data.preorders.Count -lt 1 -and $r.data.new_releases.Count -lt 1) { throw "empty via proxy" }
}

Test-Ok "admin page" {
  Invoke-WebRequest "http://localhost:5174/" -UseBasicParsing -TimeoutSec 10 | Out-Null
}

Write-Host ""
if ($failed -gt 0) {
  Write-Host "FAILED: $failed check(s)" -ForegroundColor Red
  exit 1
}
Write-Host "All checks passed." -ForegroundColor Green

param(
    [string]$BaseUrl = "http://localhost:3001",
    [ValidateSet("all", "01", "02", "03", "04", "05")]
    [string]$Scenario = "all",
    [switch]$ReplenishStock = $true
)

$ErrorActionPreference = "Continue"

function Write-Section([string]$Title) {
    Write-Host ""
    Write-Host "=======================================================" -ForegroundColor Cyan
    Write-Host "  $Title" -ForegroundColor Cyan
    Write-Host "=======================================================" -ForegroundColor Cyan
}

function Write-Step([string]$Msg) {
    Write-Host "  >> $Msg" -ForegroundColor Yellow
}

function Write-OK([string]$Msg) {
    Write-Host "  [OK] $Msg" -ForegroundColor Green
}

function Write-Warn([string]$Msg) {
    Write-Host "  [WARN] $Msg" -ForegroundColor DarkYellow
}

function Write-Err([string]$Msg) {
    Write-Host "  [ERR] $Msg" -ForegroundColor Red
}

# Load .env values with defaults from .env.example
$envVars = @{
    MYSQL_USER     = "gouser"
    MYSQL_PASSWORD = "gosecret123"
    MYSQL_DATABASE = "go_ecommerce_db"
    REDIS_PASSWORD = "redissecret123"
}

$repoRoot = Split-Path -Parent $PSScriptRoot
$envFile = Join-Path $repoRoot ".env"
if (Test-Path $envFile) {
    Get-Content $envFile | ForEach-Object {
        if ($_ -match '^\s*([A-Za-z_][A-Za-z0-9_]*)\s*=\s*(.*)\s*$') {
            $key = $Matches[1]
            $val = $Matches[2].Trim('"', "'")
            if ($envVars.ContainsKey($key) -and $val -ne "") {
                $envVars[$key] = $val
            }
        }
    }
}

function Get-MetricsSnapshot([string]$Url) {
    try {
        $resp = Invoke-WebRequest -Uri "$Url/metrics" -UseBasicParsing -TimeoutSec 5
        $raw = $resp.Content
        $goroutines = if ($raw -match '(?m)^go_goroutines\s+([0-9.eE+-]+)') { [double]$Matches[1] } else { 0 }
        $heapBytes = if ($raw -match '(?m)^go_memstats_alloc_bytes\s+([0-9.eE+-]+)') { [double]$Matches[1] } else { 0 }
        $gcCount = if ($raw -match '(?m)^go_gc_duration_seconds_count\s+([0-9.eE+-]+)') { [double]$Matches[1] } else { 0 }

        $totalHttp = 0
        $raw -split "`n" | Where-Object { $_ -match '^http_requests_total\{' } | ForEach-Object {
            if ($_ -match '\s([0-9.eE+-]+)\s*$') {
                $totalHttp += [double]$Matches[1]
            }
        }

        return @{
            OK         = $true
            Goroutines = $goroutines
            HeapMB     = [Math]::Round($heapBytes / 1MB, 2)
            GCCount    = $gcCount
            TotalHTTP  = $totalHttp
        }
    } catch {
        return @{ OK = $false }
    }
}

function Get-DockerAppStats {
    try {
        $line = docker stats go-app-service --no-stream --format "{{.CPUPerc}}|{{.MemUsage}}|{{.MemPerc}}" 2>$null
        if ($line) {
            $parts = $line -split '\|'
            return @{
                OK      = $true
                CPU     = $parts[0]
                Mem     = $parts[1]
                MemPerc = $parts[2]
            }
        }
    } catch {}
    return @{ OK = $false }
}

function Invoke-StockReplenish {
    if (-not $ReplenishStock) {
        Write-Step "Skipping MySQL stock replenishment (-ReplenishStock:`$false)"
        return
    }
    Write-Step "Replenishing product stock in MySQL (IDs 20..45 -> 10000, ID 1 -> 100)..."
    $sql = "UPDATE products SET stock = 10000 WHERE id BETWEEN 20 AND 45; UPDATE products SET stock = 100 WHERE id = 1;"
    $out = docker exec db-mysql-service mysql "-u$($envVars.MYSQL_USER)" "-p$($envVars.MYSQL_PASSWORD)" $envVars.MYSQL_DATABASE -e $sql 2>&1
    if ($LASTEXITCODE -eq 0) {
        Write-OK "Product stock replenished"
    } else {
        Write-Warn "Could not replenish stock via docker exec (db-mysql-service): $out"
    }
}

function Invoke-RedisFlush {
    Write-Step "Flushing Redis cache to test cold-cache singleflight stampede..."
    $out = docker exec redis-cache-service redis-cli -a $envVars.REDIS_PASSWORD FLUSHDB 2>&1
    if ($LASTEXITCODE -eq 0) {
        Write-OK "Redis cache flushed"
    } else {
        Write-Warn "Could not flush Redis via docker exec (redis-cache-service): $out"
    }
}

function Get-OutboxSummary {
    try {
        $sql = "SELECT status, COUNT(*) AS cnt FROM outbox_events GROUP BY status;"
        $out = docker exec db-mysql-service mysql "-u$($envVars.MYSQL_USER)" "-p$($envVars.MYSQL_PASSWORD)" $envVars.MYSQL_DATABASE -e $sql 2>$null
        return $out
    } catch {
        return $null
    }
}

# 1. Pre-flight checks
Write-Section "PRE-FLIGHT CHECKS"

$useDockerK6 = $false
if (Get-Command k6 -ErrorAction SilentlyContinue) {
    $k6Ver = k6 version 2>&1
    Write-OK "Local k6 CLI found: $k6Ver"
} elseif (Get-Command docker -ErrorAction SilentlyContinue) {
    $useDockerK6 = $true
    Write-Warn "Local k6 CLI not found; falling back to Docker image grafana/k6"
} else {
    Write-Err "Neither 'k6' CLI nor 'docker' is available in PATH. Install k6 from https://k6.io/docs/get-started/installation/"
    exit 1
}

Write-Step "Checking Catalyst API health at $BaseUrl/metrics..."
$preCheck = Get-MetricsSnapshot -Url $BaseUrl
if (-not $preCheck.OK) {
    Write-Err "Cannot reach $BaseUrl/metrics. Make sure Catalyst is running (docker-compose up -d)."
    exit 1
}
Write-OK "Catalyst API is reachable (Goroutines=$($preCheck.Goroutines), Heap=$($preCheck.HeapMB) MB)"

# 2. Resolve scenario list
$allScripts = @(
    @{ Id = "01"; Name = "Cache Hit & Singleflight Stampede"; File = "01_cache_stampede.js"; FlushRedis = $true; Replenish = $false },
    @{ Id = "02"; Name = "Auth Bcrypt CPU Stress";           File = "02_auth_bcrypt_stress.js"; FlushRedis = $false; Replenish = $false },
    @{ Id = "03"; Name = "Transaction & Outbox Throughput";  File = "03_transaction_throughput.js"; FlushRedis = $false; Replenish = $true },
    @{ Id = "04"; Name = "Flash Sale & Idempotency Race";    File = "04_flash_sale_contention.js"; FlushRedis = $false; Replenish = $true },
    @{ Id = "05"; Name = "Mixed E-Commerce Load (75/15/10)"; File = "05_mixed_ecommerce_load.js"; FlushRedis = $false; Replenish = $true }
)

$selectedScripts = if ($Scenario -eq "all") {
    $allScripts
} else {
    $allScripts | Where-Object { $_.Id -eq $Scenario }
}

$k6Dir = Join-Path $repoRoot "test\k6"

foreach ($item in $selectedScripts) {
    Write-Section "SCENARIO $($item.Id): $($item.Name)"

    if ($item.FlushRedis) {
        Invoke-RedisFlush
    }
    if ($item.Replenish) {
        Invoke-StockReplenish
    }

    $beforeMetrics = Get-MetricsSnapshot -Url $BaseUrl
    $beforeDocker = Get-DockerAppStats

    $scriptPath = Join-Path $k6Dir $item.File
    Write-Step "Running $scriptPath..."

    if (-not $useDockerK6) {
        & k6 run -e "BASE_URL=$BaseUrl" $scriptPath
    } else {
        $dockerBaseUrl = $BaseUrl -replace 'localhost', 'host.docker.internal' -replace '127\.0\.0\.1', 'host.docker.internal'
        & docker run --rm -i `
            --add-host=host.docker.internal:host-gateway `
            -v "${k6Dir}:/scripts" `
            -e "BASE_URL=$dockerBaseUrl" `
            grafana/k6 run "/scripts/$($item.File)"
    }

    $exitCode = $LASTEXITCODE
    $afterMetrics = Get-MetricsSnapshot -Url $BaseUrl
    $afterDocker = Get-DockerAppStats

    Write-Host ""
    Write-Step "Computational Efficiency Summary (Scenario $($item.Id)):"
    if ($beforeMetrics.OK -and $afterMetrics.OK) {
        $deltaReq = [Math]::Round($afterMetrics.TotalHTTP - $beforeMetrics.TotalHTTP, 0)
        $deltaGoroutines = [Math]::Round($afterMetrics.Goroutines - $beforeMetrics.Goroutines, 0)
        $deltaGC = [Math]::Round($afterMetrics.GCCount - $beforeMetrics.GCCount, 0)
        Write-Host ("    HTTP Requests Recorded (/metrics) : +{0}" -f $deltaReq) -ForegroundColor White
        Write-Host ("    Go Goroutines (Before -> After)  : {0} -> {1} (delta: {2})" -f $beforeMetrics.Goroutines, $afterMetrics.Goroutines, $deltaGoroutines) -ForegroundColor White
        Write-Host ("    Go Heap Alloc (Before -> After)  : {0} MB -> {1} MB" -f $beforeMetrics.HeapMB, $afterMetrics.HeapMB) -ForegroundColor White
        Write-Host ("    Go GC Cycles Triggered           : +{0}" -f $deltaGC) -ForegroundColor White
    }
    if ($afterDocker.OK) {
        Write-Host ("    Container CPU / Mem (Post-Test)  : CPU={0}, Mem={1} ({2})" -f $afterDocker.CPU, $afterDocker.Mem, $afterDocker.MemPerc) -ForegroundColor White
    }

    if ($item.Replenish) {
        $outbox = Get-OutboxSummary
        if ($outbox) {
            Write-Host "    Outbox Queue Status:" -ForegroundColor White
            $outbox | ForEach-Object { Write-Host "      $_" -ForegroundColor Gray }
        }
    }

    if ($exitCode -eq 0) {
        Write-OK "Scenario $($item.Id) passed all k6 thresholds"
    } else {
        Write-Warn "Scenario $($item.Id) finished with exit code $exitCode (check threshold results above)"
    }
}

Write-Section "BENCHMARK COMPLETE"
Write-Host "  Prometheus : http://localhost:9090" -ForegroundColor White
Write-Host "  Grafana    : http://localhost:3002 (admin / admin)" -ForegroundColor White
Write-Host "  Jaeger     : http://localhost:16686" -ForegroundColor White
Write-Host ""


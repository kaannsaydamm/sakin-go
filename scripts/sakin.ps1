# SGE - Master Control Script (Windows PowerShell)
# Usage: .\sakin.ps1 [-Action] {Start|Stop|Restart|Status|Logs} [-Target] {All|Infra|Services}

param (
    [Parameter(Mandatory = $false)]
    [ValidateSet("Start", "Stop", "Restart", "Status", "Logs")]
    [string]$Action = "Start",

    [Parameter(Mandatory = $false)]
    [ValidateSet("All", "Infra", "Services")]
    [string]$Target = "All"
)

# Load Environment from .env
if (Test-Path ".env") {
    Get-Content ".env" | ForEach-Object {
        if ($_ -match "^\s*([^#=]+)=(.*)$") {
            [System.Environment]::SetEnvironmentVariable($matches[1], $matches[2], "Process")
        }
    }
}
else {
    Write-Warning ".env file not found! Using defaults."
}

# Create logs directory
New-Item -ItemType Directory -Force -Path "logs" | Out-Null

$Services = @("ingest", "enrichment", "correlation", "analytics", "soar", "panel-api")

function Start-Infra {
    Write-Host "🐳 Starting Infrastructure (Docker)..." -ForegroundColor Cyan
    docker-compose up -d
}

function Stop-Infra {
    Write-Host "🛑 Stopping Infrastructure..." -ForegroundColor Yellow
    docker-compose down
}

function Start-Services {
    Write-Host "🚀 Starting SGE Services..." -ForegroundColor Green
    foreach ($svc in $Services) {
        $logFile = "logs\$svc.log"
        Write-Host "   - Starting sge-$svc..."
        # Start-Process with -NoNewWindow would block, so we use start and redirect
        # This is a bit tricky in PS background. Using Start-Process to spawn independent windows or background jobs.
        Start-Job -Name "sge-$svc" -ScriptBlock {
            param($s, $log)
            go run cmd/sge-$s/main.go > $log 2>&1
        } -ArgumentList $svc, $logFile
    }
    Write-Host "✅ Services started as Background Jobs. Use 'Get-Job' or '.\sakin.ps1 -Action Logs' to check." -ForegroundColor Green
}

function Stop-Services {
    Write-Host "🛑 Stopping SGE Services..." -ForegroundColor Yellow
    # Stop PowerShell Jobs
    Get-Job | Where-Object { $_.Name -like "sge-*" } | Stop-Job
    Get-Job | Where-Object { $_.Name -like "sge-*" } | Remove-Job
    
    # Also kill by process name if running purely as go.exe (risky if other go apps run)
    # Stop-Process -Name "main" -ErrorAction SilentlyContinue
}

function Show-Logs {
    Write-Host "📄 Tailing logs (Ctrl+C to exit)..." -ForegroundColor Cyan
    Get-Content logs\*.log -Wait
}

# Main Logic
switch ($Action) {
    "Start" {
        if ($Target -eq "Infra") { Start-Infra }
        elseif ($Target -eq "Services") { Start-Services }
        else { Start-Infra; Start-Services }
    }
    "Stop" {
        if ($Target -eq "Services") { Stop-Services }
        elseif ($Target -eq "Infra") { Stop-Infra }
        else { Stop-Services; Stop-Infra }
    }
    "Restart" {
        Stop-Services
        Start-Services
    }
    "Status" {
        docker-compose ps
        Get-Job | Format-Table
    }
    "Logs" {
        Show-Logs
    }
}

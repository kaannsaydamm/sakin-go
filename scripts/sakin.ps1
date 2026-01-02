# SGE - Master Control Script (Windows PowerShell)
# Usage: .\sakin.ps1 [-Action] {Start|Stop|Restart|Status|Logs} [-Target] {All|Infra|Services}
# Run without arguments for Interactive Mode

param (
    [Parameter(Mandatory = $false)]
    [ValidateSet("Start", "Stop", "Restart", "Status", "Logs")]
    [string]$Action = $null,

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

function Show-Banner {
    Clear-Host
    Write-Host "   _____   ___      __ __   ____  _   __       ______         ______    ___ __  _           " -ForegroundColor Cyan
    Write-Host "  ╱ ___╱  ╱   │    ╱ ╱╱_╱  ╱  _╱ ╱ │ ╱ ╱ _    ╱ ____╱___     ╱ ____╱___╱ (_) ╱_(_)___  ____ " -ForegroundColor Cyan
    Write-Host "  ╲__ ╲  ╱ ╱│ │   ╱ ,<     ╱ ╱  ╱  │╱ ╱ (_)  ╱ ╱ __╱ __ ╲   ╱ __╱ ╱ __  ╱ ╱ __╱ ╱ __ ╲╱ __ ╲" -ForegroundColor Cyan
    Write-Host " ___╱ ╱ ╱ ___ │_ ╱ ╱│ │_ _╱ ╱_ ╱ ╱│  ╱ _    ╱ ╱_╱ ╱ ╱_╱ ╱  ╱ ╱___╱ ╱_╱ ╱ ╱ ╱_╱ ╱ ╱_╱ ╱ ╱ ╱ ╱" -ForegroundColor Cyan
    Write-Host "╱____(_)_╱  │_(_)_╱ │_(_)___(_)_╱ │_(_│_)   ╲____╱╲____╱  ╱_____╱╲__,_╱_╱╲__╱_╱╲____╱_╱ ╱_╱ " -ForegroundColor Cyan
    Write-Host "                                                                                            "
    Write-Host "    Master Control CLI " -ForegroundColor Green
    Write-Host "=======================" -ForegroundColor Yellow
}

function Start-Infra {
    Write-Host "`n🐳 Starting Infrastructure (Docker)..." -ForegroundColor Cyan
    docker-compose up -d
}

function Stop-Infra {
    Write-Host "`n🛑 Stopping Infrastructure..." -ForegroundColor Yellow
    docker-compose down
}

function Start-Services {
    Write-Host "`n🚀 Starting SGE Services..." -ForegroundColor Green
    foreach ($svc in $Services) {
        $logFile = "logs\$svc.log"
        Write-Host "   - Starting sge-$svc..."
        Start-Job -Name "sge-$svc" -ScriptBlock {
            param($s, $log)
            go run cmd/sge-$s/main.go > $log 2>&1
        } -ArgumentList $svc, $logFile
    }
    Write-Host "✅ Services started as Background Jobs. Use 'Get-Job' or '.\sakin.ps1 -Action Logs' to check." -ForegroundColor Green
}

function Stop-Services {
    Write-Host "`n🛑 Stopping SGE Services..." -ForegroundColor Yellow
    Get-Job | Where-Object { $_.Name -like "sge-*" } | Stop-Job
    Get-Job | Where-Object { $_.Name -like "sge-*" } | Remove-Job
}

function Show-Logs {
    Write-Host "`n📄 Tailing logs (Ctrl+C to exit)..." -ForegroundColor Cyan
    Get-Content logs\*.log -Wait
}

function Get-Status {
    Write-Host "`n📊 Infrastructure Status:" -ForegroundColor Cyan
    docker-compose ps
    Write-Host "`n📊 Services Status (Jobs):" -ForegroundColor Cyan
    Get-Job | Where-Object { $_.Name -like "sge-*" } | Format-Table
}

function Invoke-Action {
    param($act, $tgt)
    switch ($act) {
        "Start" {
            if ($tgt -eq "Infra") { Start-Infra }
            elseif ($tgt -eq "Services") { Start-Services }
            else { Start-Infra; Start-Services }
        }
        "Stop" {
            if ($tgt -eq "Services") { Stop-Services }
            elseif ($tgt -eq "Infra") { Stop-Infra }
            else { Stop-Services; Stop-Infra }
        }
        "Restart" {
            Stop-Services; Stop-Infra
            Start-Sleep -Seconds 2
            Start-Infra; Start-Services
        }
        "Status" { Get-Status }
        "Logs" { Show-Logs }
    }
}

# --- Main Logic ---
if ($PSBoundParameters.Count -eq 0) {
    # Interactive Mode
    Show-Banner
    Write-Host "Select an action:"
    Write-Host "1) Start All (Infra + Services)"
    Write-Host "2) Stop All"
    Write-Host "3) Restart All"
    Write-Host "4) Status Check"
    Write-Host "5) Tail Logs"
    Write-Host "6) Start Services Only"
    Write-Host "7) Stop Services Only"
    Write-Host "8) Exit"

    $choice = Read-Host "Enter choice [1-8]"

    switch ($choice) {
        "1" { Invoke-Action "Start" "All" }
        "2" { Invoke-Action "Stop" "All" }
        "3" { Invoke-Action "Restart" "All" }
        "4" { Invoke-Action "Status" "All" }
        "5" { Invoke-Action "Logs" "All" }
        "6" { Invoke-Action "Start" "Services" }
        "7" { Invoke-Action "Stop" "Services" }
        "8" { exit }
        Default { Write-Warning "Invalid option." }
    }
}
else {
    # Argument Mode
    Invoke-Action $Action $Target
}

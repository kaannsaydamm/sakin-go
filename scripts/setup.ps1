# SGE - Sakin Go Edition: Interactive Setup Script (Windows)
# Usage: .\setup.ps1

Write-Host "   _____   ___      __ __   ____  _   __       ______         ______    ___ __  _           " -ForegroundColor Blue
Write-Host "  ╱ ___╱  ╱   │    ╱ ╱╱_╱  ╱  _╱ ╱ │ ╱ ╱ _    ╱ ____╱___     ╱ ____╱___╱ (_) ╱_(_)___  ____ " -ForegroundColor Blue
Write-Host "  ╲__ ╲  ╱ ╱│ │   ╱ ,<     ╱ ╱  ╱  │╱ ╱ (_)  ╱ ╱ __╱ __ ╲   ╱ __╱ ╱ __  ╱ ╱ __╱ ╱ __ ╲╱ __ ╲" -ForegroundColor Blue
Write-Host " ___╱ ╱ ╱ ___ │_ ╱ ╱│ │_ _╱ ╱_ ╱ ╱│  ╱ _    ╱ ╱_╱ ╱ ╱_╱ ╱  ╱ ╱___╱ ╱_╱ ╱ ╱ ╱_╱ ╱ ╱_╱ ╱ ╱ ╱ ╱" -ForegroundColor Blue
Write-Host "╱____(_)_╱  │_(_)_╱ │_(_)___(_)_╱ │_(_│_)   ╲____╱╲____╱  ╱_____╱╲__,_╱_╱╲__╱_╱╲____╱_╱ ╱_╱ " -ForegroundColor Blue
Write-Host "                                                                                            " -ForegroundColor Blue
Write-Host "      Sakin Go Edition - Infrastructure Setup        " -ForegroundColor Cyan
Write-Host "=====================================================" -ForegroundColor Yellow

function Test-Dependencies {
    Write-Host "`n[1/3] Checking Dependencies..." -ForegroundColor Cyan
    
    if (Get-Command go -ErrorAction SilentlyContinue) {
        $goVersion = go version
        Write-Host "✅ Go is installed: $goVersion" -ForegroundColor Green
    }
    else {
        Write-Host "❌ Go is not installed!" -ForegroundColor Red
        exit 1
    }

    if (Get-Command docker -ErrorAction SilentlyContinue) {
        Write-Host "✅ Docker is installed." -ForegroundColor Green
    }
    else {
        Write-Host "❌ Docker is not installed!" -ForegroundColor Red
        exit 1
    }
}

function Install-GoModules {
    Write-Host "`n[2/3] Installing Go Modules..." -ForegroundColor Cyan
    go mod tidy
    go mod download
    Write-Host "✅ Modules downloaded." -ForegroundColor Green
}

function Initialize-Directories {
    Write-Host "`n[3/3] Setting up Certificates & Directories..." -ForegroundColor Cyan
    
    New-Item -ItemType Directory -Force -Path "certs" | Out-Null
    New-Item -ItemType Directory -Force -Path "logs" | Out-Null
    
    if (-not (Test-Path ".env")) {
        Write-Host "⚠️  .env file missing. Creating from example..." -ForegroundColor Yellow
        if (Test-Path ".env.example") {
            Copy-Item ".env.example" ".env"
            Write-Host "✅ Created .env from .env.example." -ForegroundColor Green
        }
    }

    Write-Host "✅ Directory structure ready." -ForegroundColor Green
}

function Invoke-FullSetup {
    Test-Dependencies
    Install-GoModules
    Initialize-Directories
    Write-Host "`n🎉 Setup Complete! You can now run '.\scripts\sakin.ps1' to start." -ForegroundColor Green
}

# Menu
Write-Host "Select an action:"
Write-Host "1) Full Setup (Dependencies + Modules + Certs)"
Write-Host "2) Install Go Modules Only"
Write-Host "3) Setup Directories & .env Only"
Write-Host "4) Exit"

$choice = Read-Host "Enter choice [1-4]"

switch ($choice) {
    "1" { Invoke-FullSetup }
    "2" { Install-GoModules }
    "3" { Initialize-Directories }
    "4" { exit }
    Default { Write-Warning "Invalid option." }
}

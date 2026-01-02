Write-Host "🚀 SGE: Setting up Development Environment (Windows)..." -ForegroundColor Cyan

# 1. Go Dependencies
Write-Host "📦 Downloading Go modules..."
go mod tidy
go mod download

# 2. Check Docker
if (-not (Get-Command docker -ErrorAction SilentlyContinue)) {
    Write-Host "❌ Docker not found! Please install Docker Desktop." -ForegroundColor Red
    exit 1
}

Write-Host "✅ Docker ready."

# 3. Create Certs directory
New-Item -ItemType Directory -Force -Path "certs" | Out-Null
Write-Host "⚠️  Please generate mTLS certificates in 'certs' folder." -ForegroundColor Yellow

Write-Host "✅ Setup Complete!" -ForegroundColor Green

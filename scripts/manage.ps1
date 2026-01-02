param (
    [string]$action
)

if ($action -eq "start") {
    Write-Host "🚀 Starting Infrastructure (Docker)..."
    docker-compose up -d
    
    Write-Host "ℹ️  Run these in separate terminals:"
    Write-Host "  go run cmd/sge-ingest/main.go"
    Write-Host "  go run cmd/sge-correlation/main.go"
    Write-Host "  go run cmd/sge-enrichment/main.go"
    Write-Host "  go run cmd/sge-analytics/main.go"
    Write-Host "  go run cmd/sge-soar/main.go"
    Write-Host "  go run cmd/sge-panel-api/main.go"
}
elseif ($action -eq "stop") {
    Write-Host "🛑 Stopping Docker..."
    docker-compose down
    Write-Host "⚠️  Please close the running Go terminals manually."
}
else {
    Write-Host "Usage: .\manage.ps1 start | stop"
}

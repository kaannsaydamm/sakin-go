#!/bin/bash

# SGE - Master Control Script (Linux/macOS)
# Usage: ./sakin.sh {start|stop|restart|status|logs} [infra|services|all]

set -e

# Load Environment
if [ -f .env ]; then
    set -a
    source .env
    set +a
else
    echo "⚠️  .env file not found! Using defaults."
fi

# Directory for logs
mkdir -p logs

SERVICES=("ingest" "enrichment" "correlation" "analytics" "soar" "panel-api")

function start_infra() {
    echo "🐳 Starting Infrastructure (Docker)..."
    docker-compose up -d
    echo "⏳ Waiting for databases to be ready..."
    sleep 5 # Simple wait, ideally use health checks
}

function stop_infra() {
    echo "🛑 Stopping Infrastructure..."
    docker-compose down
}

function start_services() {
    echo "🚀 Starting SGE Services..."
    for svc in "${SERVICES[@]}"; do
        if pgrep -f "cmd/sge-$svc/main.go" > /dev/null; then
            echo "   - sge-$svc is already running."
        else
            echo "   - Starting sge-$svc..."
            nohup go run cmd/sge-$svc/main.go > logs/$svc.log 2>&1 &
        fi
    done
    echo "✅ All services started in background. Check logs/ directory for output."
}

function stop_services() {
    echo "🛑 Stopping SGE Services..."
    # Kill all 'go run' processes associated with our services
    # Warning: This is a bit aggressive, matches path string
    pkill -f "cmd/sge-" || echo "   No services found running."
}

function logs() {
    echo "📄 Tailing logs (Ctrl+C to exit)..."
    tail -f logs/*.log
}

case "$1" in
    "start")
        if [ "$2" == "services" ]; then
            start_services
        elif [ "$2" == "infra" ]; then
            start_infra
        else
            start_infra
            start_services
        fi
        ;;
    "stop")
        if [ "$2" == "services" ]; then
            stop_services
        elif [ "$2" == "infra" ]; then
            stop_infra
        else
            stop_services
            stop_infra
        fi
        ;;
    "restart")
        $0 stop
        sleep 2
        $0 start
        ;;
    "status")
        echo "📊 Infrastructure Status:"
        docker-compose ps
        echo ""
        echo "📊 Services Status:"
        pgrep -a -f "cmd/sge-"
        ;;
    "logs")
        logs
        ;;
    *)
        echo "Usage: $0 {start|stop|restart|status|logs} [infra|services]"
        echo "  start       : Start Infra + Services"
        echo "  start infra : Start only Docker Infra"
        echo "  stop        : Stop everything"
        echo "  logs        : Tail all service logs"
        exit 1
        ;;
esac

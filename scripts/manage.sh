#!/bin/bash

ACTION=$1

case "$ACTION" in
    "start")
        echo "🚀 Starting Infrastructure (Docker)..."
        # Assuming docker-compose.yml exists (we likely need to create one, or assume user has it)
        # We haven't created docker-compose.yml yet! We MUST add that to the plan or just create it now implicitly.
        # I'll create it in a moment.
        docker-compose up -d
        
        echo "🚀 Starting Services (Local)..."
        # In real dev env, we might use modd, air, or just background jobs
        # For simplicity:
        echo "Start commands: (Run in separate terminals)"
        echo "  go run cmd/sge-ingest/main.go"
        echo "  go run cmd/sge-correlation/main.go"
        echo "  go run cmd/sge-enrichment/main.go"
        echo "  go run cmd/sge-analytics/main.go"
        echo "  go run cmd/sge-soar/main.go"
        echo "  go run cmd/sge-panel-api/main.go"
        ;;
    
    "stop")
        echo "🛑 Stopping..."
        docker-compose down
        pkill -f "go run cmd/sge"
        ;;
    
    *)
        echo "Usage: $0 {start|stop}"
        exit 1
        ;;
esac

#!/bin/bash
set -e

echo "🚀 SGE: Setting up Development Environment..."

# 1. Go Dependencies
echo "📦 Downloading Go modules..."
go mod tidy
go mod download

# 2. Check Docker
if ! command -v docker &> /dev/null; then
    echo "❌ Docker not found! Please install Docker."
    exit 1
fi

if ! command -v docker-compose &> /dev/null; then
    # try "docker compose" plugin
    if ! docker compose version &> /dev/null; then
        echo "❌ Docker Compose not found! Please install."
        exit 1
    fi
fi

echo "✅ Docker ready."

# 3. Create Certs (mTLS)
echo "🔐 Generating CA and Certificates..."
# We assume we have a compiled tool or we run the secure-comms test or we add a cert-gen tool.
# For now, let's create a placeholder directory
mkdir -p certs
# go run cmd/tools/certgen/main.go (If we made one, otherwise notify user)
echo "⚠️  Certificates must be generated manually or by running the cert manager test."

echo "✅ Setup Complete!"

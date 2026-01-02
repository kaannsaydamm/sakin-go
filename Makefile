# SGE (Sakin Go Edition) Makefile
# Build ve deployment işlemlerini kolaylaştırır

.PHONY: all build clean test fmt lint docker help

# Varsayılan hedef
all: fmt lint test build

# Build komutları
build: build-agent build-ingest build-network-sensor build-correlation build-analytics build-soar

build-agent:
	@echo "Building sge-agent..."
	@cd cmd/sge-agent && go build -o ../../bin/sge-agent

build-ingest:
	@echo "Building sge-ingest..."
	@cd cmd/sge-ingest && go build -o ../../bin/sge-ingest

build-network-sensor:
	@echo "Building sge-network-sensor..."
	@cd cmd/sge-network-sensor && go build -o ../../bin/sge-network-sensor

build-correlation:
	@echo "Building sge-correlation..."
	@cd cmd/sge-correlation && go build -o ../../bin/sge-correlation

build-analytics:
	@echo "Building sge-analytics..."
	@cd cmd/sge-analytics && go build -o ../../bin/sge-analytics

build-soar:
	@echo "Building sge-soar..."
	@cd cmd/sge-soar && go build -o ../../bin/sge-soar

# Production build (obfuscated)
build-prod:
	@echo "Building production binaries with garble..."
	@command -v garble >/dev/null 2>&1 || { echo "garble not found. Installing..."; go install mvdan.cc/garble@latest; }
	@cd cmd/sge-agent && garble -literals -tiny build -o ../../bin/sge-agent
	@cd cmd/sge-ingest && garble -literals -tiny build -o ../../bin/sge-ingest
	@cd cmd/sge-network-sensor && garble -literals -tiny build -o ../../bin/sge-network-sensor
	@cd cmd/sge-correlation && garble -literals -tiny build -o ../../bin/sge-correlation
	@cd cmd/sge-analytics && garble -literals -tiny build -o ../../bin/sge-analytics
	@cd cmd/sge-soar && garble -literals -tiny build -o ../../bin/sge-soar
	@echo "Production binaries created in bin/"

# Cross-platform build
build-linux:
	@echo "Building for Linux (amd64)..."
	@GOOS=linux GOARCH=amd64 $(MAKE) build

build-windows:
	@echo "Building for Windows (amd64)..."
	@GOOS=windows GOARCH=amd64 $(MAKE) build

build-darwin:
	@echo "Building for macOS (amd64)..."
	@GOOS=darwin GOARCH=amd64 $(MAKE) build

build-all-platforms: build-linux build-windows build-darwin

# Test komutları
test:
	@echo "Running tests..."
	@go test -v ./...

test-race:
	@echo "Running tests with race detection..."
	@go test -race ./...

test-cover:
	@echo "Running tests with coverage..."
	@go test -cover -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

bench:
	@echo "Running benchmarks..."
	@go test -bench=. -benchmem ./...

# Code quality
fmt:
	@echo "Formatting code..."
	@go fmt ./...

lint:
	@echo "Running linter..."
	@command -v golangci-lint >/dev/null 2>&1 || { echo "golangci-lint not found. Installing..."; go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; }
	@golangci-lint run ./...

vet:
	@echo "Running go vet..."
	@go vet ./...

# Dependencies
deps:
	@echo "Downloading dependencies..."
	@go mod download

deps-update:
	@echo "Updating dependencies..."
	@go get -u ./...
	@go mod tidy

# Temizlik
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@echo "Clean complete"

# Docker komutları
docker-build:
	@echo "Building Docker images..."
	@docker-compose -f deployments/docker/docker-compose.yml build

docker-up:
	@echo "Starting services with Docker Compose..."
	@docker-compose -f deployments/docker/docker-compose.yml up -d

docker-down:
	@echo "Stopping services..."
	@docker-compose -f deployments/docker/docker-compose.yml down

docker-logs:
	@docker-compose -f deployments/docker/docker-compose.yml logs -f

# Development
run-agent:
	@echo "Running sge-agent..."
	@cd cmd/sge-agent && go run main.go

run-ingest:
	@echo "Running sge-ingest..."
	@cd cmd/sge-ingest && go run main.go

run-network-sensor:
	@echo "Running sge-network-sensor..."
	@cd cmd/sge-network-sensor && go run main.go

# Binary dizini oluştur
bin:
	@mkdir -p bin

# Yardım
help:
	@echo "SGE (Sakin Go Edition) - Makefile Komutları"
	@echo ""
	@echo "Build:"
	@echo "  make build              - Tüm servisleri derle"
	@echo "  make build-prod         - Production build (obfuscated)"
	@echo "  make build-linux        - Linux için derle"
	@echo "  make build-windows      - Windows için derle"
	@echo "  make build-darwin       - macOS için derle"
	@echo ""
	@echo "Test:"
	@echo "  make test               - Testleri çalıştır"
	@echo "  make test-race          - Race detection ile test"
	@echo "  make test-cover         - Coverage raporu oluştur"
	@echo "  make bench              - Benchmark testleri"
	@echo ""
	@echo "Code Quality:"
	@echo "  make fmt                - Kodu formatla"
	@echo "  make lint               - Linter çalıştır"
	@echo "  make vet                - Go vet çalıştır"
	@echo ""
	@echo "Dependencies:"
	@echo "  make deps               - Bağımlılıkları indir"
	@echo "  make deps-update        - Bağımlılıkları güncelle"
	@echo ""
	@echo "Docker:"
	@echo "  make docker-build       - Docker image'ları oluştur"
	@echo "  make docker-up          - Servisleri başlat"
	@echo "  make docker-down        - Servisleri durdur"
	@echo "  make docker-logs        - Logları göster"
	@echo ""
	@echo "Development:"
	@echo "  make run-agent          - Agent'ı çalıştır"
	@echo "  make run-ingest         - Ingest servisini çalıştır"
	@echo "  make run-network-sensor - Network sensor'ı çalıştır"
	@echo ""
	@echo "Diğer:"
	@echo "  make clean              - Build artifact'ları temizle"
	@echo "  make help               - Bu yardım mesajını göster"

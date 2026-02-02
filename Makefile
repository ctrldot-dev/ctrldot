.PHONY: dev test migrate clean docker-up docker-down

# Database connection
DB_HOST ?= localhost
DB_PORT ?= 5432
DB_USER ?= kernel
DB_PASSWORD ?= kernel
DB_NAME ?= kernel
DB_URL = postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable

# Server port
PORT ?= 8080

# Start docker services
docker-up:
	docker-compose up -d

# Stop docker services
docker-down:
	docker-compose down

# Run migrations (init + FinLedger namespaces)
migrate:
	@echo "Running migrations..."
	@PGPASSWORD=$(DB_PASSWORD) psql -h $(DB_HOST) -p $(DB_PORT) -U $(DB_USER) -d $(DB_NAME) -f migrations/0001_init.sql
	@PGPASSWORD=$(DB_PASSWORD) psql -h $(DB_HOST) -p $(DB_PORT) -U $(DB_USER) -d $(DB_NAME) -f migrations/0002_finledger_namespaces.sql

# Register Financial Ledger policy set for FinLedger:* (run after migrate; kernel need not be running)
bootstrap-finledger:
	@echo "Registering financialledger-policyset for FinLedger:*..."
	@DB_URL=$(DB_URL) go run ./cmd/bootstrap_finledger

# Ingest FinLedger seed graphs (kernel must be running)
seed-finledger:
	@echo "Ingesting FinLedger seed graphs..."
	@KERNEL_URL=$${KERNEL_URL:-http://localhost:8080} go run ./cmd/seed_finledger kesteron-treasury-finledger-seed.yaml kesteron-stablecoinreserves-finledger-seed.yaml

# Run kernel service
dev: docker-up
	@echo "Waiting for Postgres to be ready..."
	@sleep 2
	@$(MAKE) migrate
	@echo "Starting kernel on port $(PORT)..."
	@DB_URL=$(DB_URL) PORT=$(PORT) go run cmd/kernel/main.go

# Run tests
test:
	@echo "Running tests..."
	@DB_URL=$(DB_URL) go test -v ./...

# Run dot-cli tests
test-dot:
	@echo "Running dot-cli tests..."
	@go test -v ./cmd/dot/...

# Run all tests (kernel + dot-cli)
test-all: test test-dot

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@go clean
	@rm -f kernel

# Build binary
build:
	@echo "Building kernel..."
	@go build -o kernel cmd/kernel/main.go

# Build dot CLI
build-dot:
	@echo "Building dot CLI..."
	@go build -o bin/dot ./cmd/dot
	@chmod +x bin/dot
	@echo "Built: bin/dot"

# Install dot CLI to ~/bin
install-dot: build-dot
	@mkdir -p ~/bin
	@cp bin/dot ~/bin/dot
	@chmod +x ~/bin/dot
	@echo "Installed to ~/bin/dot"
	@echo "Make sure ~/bin is in your PATH: export PATH=\$$PATH:\$$HOME/bin"

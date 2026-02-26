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

# ---------------------------------------------------------------------------
# Ctrl Dot
# ---------------------------------------------------------------------------
CTRLDOT_PORT ?= 7777

# Run Ctrl Dot migrations (requires Postgres; use docker exec if needed)
ctrldot-migrate:
	@echo "Running Ctrl Dot migrations..."
	@docker exec -i futurematic-kernel-postgres psql -U kernel -d kernel < migrations/0007_ctrldot_tables.sql 2>/dev/null || \
		echo "Tip: start Postgres with docker-compose up -d, then run: docker exec -i futurematic-kernel-postgres psql -U kernel -d kernel < migrations/0007_ctrldot_tables.sql"

# Build Ctrl Dot daemon and CLI
build-ctrldot:
	@mkdir -p bin
	@go build -o bin/ctrldotd ./cmd/ctrldotd
	@go build -o bin/ctrldot ./cmd/ctrldot
	@go build -o bin/ctrldot-mcp ./cmd/ctrldot-mcp
	@echo "Built: bin/ctrldotd, bin/ctrldot, bin/ctrldot-mcp"

# Install Ctrl Dot binaries to ~/bin (so 'ctrldot' in PATH is current)
install-ctrldot: build-ctrldot
	@mkdir -p ~/bin
	@cp bin/ctrldot bin/ctrldotd bin/ctrldot-mcp ~/bin/
	@chmod +x ~/bin/ctrldot ~/bin/ctrldotd ~/bin/ctrldot-mcp
	@echo "Installed ctrldot, ctrldotd, ctrldot-mcp to ~/bin"
	@echo "Ensure ~/bin is in your PATH: export PATH=\$$PATH:\$$HOME/bin"

# Start Ctrl Dot daemon (requires Postgres and migrations)
ctrldot-up: docker-up build-ctrldot
	@echo "Starting Ctrl Dot on port $(CTRLDOT_PORT)..."
	@DB_URL=$(DB_URL) PORT=$(CTRLDOT_PORT) ./bin/ctrldotd

# Build OpenClaw plugin
setup-openclaw-plugin:
	@echo "Building Ctrl Dot OpenClaw plugin..."
	@cd adapters/openclaw/plugin && npm install && npm run build
	@echo "Plugin built: adapters/openclaw/plugin/dist/"

# CrewAI setup (creates venv and installs; run from repo root, then activate .venv and run example)
setup-crewai:
	@echo "Setting up CrewAI adapter..."
	@python3 -m venv .venv
	@echo "Run: source .venv/bin/activate"
	@echo "Then: pip install --upgrade pip && pip install crewai && pip install -e adapters/crewai"
	@echo "Then: python adapters/crewai/examples/crew_minimal.py"

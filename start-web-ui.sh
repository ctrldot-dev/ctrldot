#!/usr/bin/env bash
# Run this in your terminal (where Docker/Colima work) to start Postgres, kernel, and Web Dot.
set -e
cd "$(dirname "$0")"

echo "Starting Postgres..."
docker-compose up -d
echo "Waiting for Postgres..."
sleep 4

echo "Running migrations (if any)..."
docker-compose exec -T postgres psql -U kernel -d kernel -f /docker-entrypoint-initdb.d/0001_init.sql 2>/dev/null || true
docker-compose exec -T postgres psql -U kernel -d kernel -f /docker-entrypoint-initdb.d/0002_finledger_namespaces.sql 2>/dev/null || true

echo "Starting kernel on :8080 (background)..."
DB_URL="postgres://kernel:kernel@localhost:5432/kernel?sslmode=disable" PORT=8080 go run ./cmd/kernel &
KERNEL_PID=$!
sleep 4
if ! curl -s http://localhost:8080/v1/healthz >/dev/null; then
  echo "Kernel failed to start. Check DB_URL and Postgres."
  kill $KERNEL_PID 2>/dev/null || true
  exit 1
fi
echo "Kernel OK"

echo "Building and starting Web Dot on :3000..."
cd web-dot && npm run build && npm run server

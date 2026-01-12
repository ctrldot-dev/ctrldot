# Futurematic Kernel - Terminal Usage Guide

The Futurematic Kernel runs as an HTTP API server. You can interact with it using `curl` or any HTTP client.

## Starting the Kernel

```bash
# Start the kernel server (starts Docker, runs migrations, and starts the server)
make dev

# Or manually:
docker-compose up -d
make migrate
go run cmd/kernel/main.go
```

The server will start on `http://localhost:8080` by default.

## API Endpoints

### 1. Health Check

```bash
curl http://localhost:8080/v1/healthz
```

### 2. Create a Plan

Create a plan to propose changes:

```bash
curl -X POST http://localhost:8080/v1/plan \
  -H "Content-Type: application/json" \
  -d '{
    "actor_id": "user:alice",
    "capabilities": ["read", "write"],
    "namespace_id": "ProductTree:/MyProject",
    "intents": [
      {
        "kind": "CreateNode",
        "namespace_id": "ProductTree:/MyProject",
        "payload": {
          "node_id": "node:goal-1",
          "title": "Launch Product",
          "meta": {}
        }
      }
    ]
  }'
```

### 3. Apply a Plan

Apply a plan to execute the changes:

```bash
curl -X POST http://localhost:8080/v1/apply \
  -H "Content-Type: application/json" \
  -d '{
    "actor_id": "user:alice",
    "capabilities": ["read", "write"],
    "plan_id": "plan:...",
    "plan_hash": "..."
  }'
```

### 4. Expand (Query Nodes)

Get nodes with their relationships:

```bash
curl "http://localhost:8080/v1/expand?ids=node:goal-1&namespace_id=ProductTree:/MyProject&depth=1"
```

### 5. History

Get operation history for a target:

```bash
curl "http://localhost:8080/v1/history?target=node:goal-1&limit=10"
```

### 6. Diff

Get differences between two sequence numbers:

```bash
curl "http://localhost:8080/v1/diff?a_seq=1&b_seq=5&target=node:goal-1"
```

## Example Workflow

Here's a complete example workflow:

```bash
# 1. Check health
curl http://localhost:8080/v1/healthz

# 2. Create a namespace (via direct DB or through a node creation)
# For now, namespaces are created automatically when needed

# 3. Create a plan to add a node
PLAN_RESPONSE=$(curl -s -X POST http://localhost:8080/v1/plan \
  -H "Content-Type: application/json" \
  -d '{
    "actor_id": "user:alice",
    "capabilities": ["read", "write"],
    "namespace_id": "ProductTree:/Test",
    "intents": [
      {
        "kind": "CreateNode",
        "namespace_id": "ProductTree:/Test",
        "payload": {
          "node_id": "node:test-1",
          "title": "Test Node",
          "meta": {}
        }
      }
    ]
  }')

# Extract plan_id and plan_hash (requires jq)
PLAN_ID=$(echo $PLAN_RESPONSE | jq -r '.id')
PLAN_HASH=$(echo $PLAN_RESPONSE | jq -r '.hash')

# 4. Apply the plan
curl -X POST http://localhost:8080/v1/apply \
  -H "Content-Type: application/json" \
  -d "{
    \"actor_id\": \"user:alice\",
    \"capabilities\": [\"read\", \"write\"],
    \"plan_id\": \"$PLAN_ID\",
    \"plan_hash\": \"$PLAN_HASH\"
  }"

# 5. Query the node
curl "http://localhost:8080/v1/expand?ids=node:test-1&namespace_id=ProductTree:/Test"
```

## Using with jq for Pretty Output

Install `jq` for better JSON formatting:

```bash
# macOS
brew install jq

# Then pipe responses through jq:
curl -s http://localhost:8080/v1/healthz | jq
```

## Environment Variables

You can customize the server:

```bash
# Change database URL
DB_URL="postgres://user:pass@host:5432/dbname?sslmode=disable" make dev

# Change port
PORT=3000 make dev
```

## Stopping the Server

Press `Ctrl+C` in the terminal where the server is running, or:

```bash
# Stop Docker
make docker-down
```

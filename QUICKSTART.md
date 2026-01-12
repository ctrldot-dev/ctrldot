# Quick Start - Using the Kernel in Terminal

## Prerequisites

1. **Docker/Colima running** (for PostgreSQL)
2. **Go installed**
3. **Database migrated**

## Step 1: Start the Database

```bash
make docker-up
```

Wait a few seconds for PostgreSQL to be ready.

## Step 2: Run Migrations

```bash
make migrate
```

## Step 3: Start the Kernel Server

```bash
# Option 1: Using make (recommended)
make dev

# Option 2: Manual start
DB_URL="postgres://kernel:kernel@localhost:5432/kernel?sslmode=disable" \
PORT=8080 \
go run cmd/kernel/main.go
```

The server will start on `http://localhost:8080` (or the port you specify).

## Step 4: Test the Server

Open a **new terminal window** and run:

```bash
# Health check
curl http://localhost:8080/v1/healthz

# Should return: {"ok":true}
```

## Step 5: Use the Kernel

### Create a Node

```bash
# 1. Create a plan
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
          "node_id": "node:my-goal",
          "title": "My First Goal",
          "meta": {}
        }
      }
    ]
  }' | jq .

# 2. Save the plan_id and plan_hash from the response, then apply:
curl -X POST http://localhost:8080/v1/apply \
  -H "Content-Type: application/json" \
  -d '{
    "actor_id": "user:alice",
    "capabilities": ["read", "write"],
    "plan_id": "plan:YOUR_PLAN_ID_HERE",
    "plan_hash": "YOUR_PLAN_HASH_HERE"
  }' | jq .

# 3. Query the node
curl "http://localhost:8080/v1/expand?ids=node:my-goal&namespace_id=ProductTree:/MyProject" | jq .
```

### Or Use the Example Script

```bash
# Make sure the server is running, then:
./examples.sh
```

## Available Endpoints

- `GET /v1/healthz` - Health check
- `POST /v1/plan` - Create a plan
- `POST /v1/apply` - Apply a plan
- `GET /v1/expand?ids=node1,node2&namespace_id=...&depth=1` - Query nodes
- `GET /v1/history?target=node:xxx&limit=10` - Get operation history
- `GET /v1/diff?a_seq=1&b_seq=5&target=node:xxx` - Get differences

## Troubleshooting

**Server won't start:**
- Check if PostgreSQL is running: `docker ps | grep postgres`
- Check database connection: `psql postgres://kernel:kernel@localhost:5432/kernel`
- Check logs for errors

**Port already in use:**
- Use a different port: `PORT=8081 make dev`
- Or kill the process using port 8080

**404 errors:**
- Make sure you're using the correct endpoint paths (they start with `/v1/`)
- Check that the server started successfully

## Stopping the Server

Press `Ctrl+C` in the terminal where the server is running.

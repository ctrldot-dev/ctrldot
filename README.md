# Futurematic Kernel v0.1

An append-only fact engine for Nodes/Links/Materials with deterministic planning, policy enforcement, and addressable history.

## Overview

The Futurematic Kernel is a Go service that serves as the system of record for a graph-based data model. It enforces strict invariants:

- **Append-only operations**: All changes are immutable audit records
- **Plan → Apply workflow**: All mutations must be planned before applying
- **Deterministic hashing**: Same inputs always produce the same plan hash
- **Policy enforcement**: YAML-based policies evaluated at plan and apply time
- **Addressable history**: Query any point in time using `asof_seq` or `asof_time`

## Architecture

The kernel is organized into clean layers:

- **`internal/domain`**: Core domain types (Node, Link, Material, Operation, Plan, etc.)
- **`internal/store`**: Database persistence layer with transaction support
- **`internal/planner`**: Intent expansion and deterministic hashing
- **`internal/policy`**: YAML policy parsing and evaluation with predicates
- **`internal/query`**: Read operations (expand, history, diff)
- **`internal/kernel`**: Plan/Apply orchestration
- **`internal/api`**: HTTP REST API handlers
- **`cmd/kernel`**: Main application entry point

## Setup

### Prerequisites

- Go 1.21+
- PostgreSQL 14+
- Docker and Docker Compose (for local development)

### Database Setup

1. Start PostgreSQL:
```bash
docker-compose up -d
```

2. Run migrations:
```bash
make migrate
```

### Running the Kernel

```bash
make dev
```

Or manually:
```bash
go run cmd/kernel/main.go
```

The server will start on port 8080 (configurable via `PORT` environment variable).

## API Endpoints

### POST /v1/plan

Creates a plan from intents.

**Request:**
```json
{
  "actor_id": "user:gareth",
  "capabilities": ["read", "write:additive"],
  "namespace_id": "ProductTree:/MachinePay",
  "asof": { "seq": null, "time": null },
  "intents": [
    { "kind": "CreateNode", "namespace_id": "ProductTree:/MachinePay", "payload": { "title": "New Goal" } }
  ]
}
```

**Response:**
```json
{
  "id": "plan:01J...",
  "created_at": "2026-01-11T00:00:00Z",
  "actor_id": "user:gareth",
  "namespace_id": "ProductTree:/MachinePay",
  "intents": [...],
  "expanded": [...],
  "class": 1,
  "policy_report": { "denies": [], "warns": [], "infos": [] },
  "hash": "sha256:..."
}
```

### POST /v1/apply

Applies a plan.

**Request:**
```json
{
  "actor_id": "user:gareth",
  "capabilities": ["read", "write:additive"],
  "plan_id": "plan:01J...",
  "plan_hash": "sha256:..."
}
```

**Response:**
```json
{
  "id": "op:01J...",
  "seq": 123,
  "occurred_at": "2026-01-11T00:00:01Z",
  "actor_id": "user:gareth",
  "capabilities": ["read","write:additive"],
  "plan_id": "plan:01J...",
  "plan_hash": "sha256:...",
  "class": 1,
  "changes": [...]
}
```

### GET /v1/expand

Expands nodes with their roles, links, and materials.

**Query Parameters:**
- `ids` (required): Comma-separated node IDs
- `namespace_id` (optional): Filter by namespace
- `depth` (optional, default: 1): Expansion depth
- `asof_seq` (optional): Query at specific sequence
- `asof_time` (optional): Query at specific time (ISO 8601)

**Response:**
```json
{
  "nodes": [...],
  "role_assignments": [...],
  "links": [...],
  "materials": [...]
}
```

**CLI Usage:**
```bash
dot show node:123
dot show node:123 --json
```

### GET /v1/history

Retrieves operations for a target.

**Query Parameters:**
- `target` (required): Node ID or namespace ID
- `limit` (optional, default: 100): Maximum number of operations

**Response:**
```json
[
  {
    "id": "op:01J...",
    "seq": 123,
    "occurred_at": "2026-01-11T00:00:01Z",
    ...
  }
]
```

### GET /v1/diff

Computes the difference between two sequence numbers.

**Query Parameters:**
- `a_seq` (required): First sequence number
- `b_seq` (required): Second sequence number
- `target` (optional): Node ID or namespace ID to filter

**Response:**
```json
{
  "changes": [...]
}
```

### GET /v1/healthz

Health check endpoint.

**Response:**
```json
{
  "ok": true
}
```

## Policy Language

Policies are namespace-scoped YAML files that define rules for operations. See `context/08-default-policyset.yaml` for an example.

### Predicates

- `acyclic(link_type)`: Ensures no cycles in hierarchy
- `role_edge_allowed(parent_role[], child_role[])`: Validates role transitions
- `child_has_only_one_parent(child_role, link_type)`: Enforces single parent constraint
- `has_capability(cap)`: Capability check (stub in v0.1)

### Effects

- `deny`: Blocks apply
- `warn`: Logs warning but allows apply
- `info`: Logs information

## Development

### Running Tests

```bash
make test
```

### Building

```bash
make build
```

### Project Structure

```
.
├── cmd/kernel/          # Main application
├── internal/
│   ├── domain/         # Domain types
│   ├── store/          # Database layer
│   ├── planner/        # Intent expansion
│   ├── policy/         # Policy engine
│   ├── query/          # Query operations
│   ├── kernel/         # Plan/Apply orchestration
│   └── api/            # HTTP handlers
├── migrations/         # Database migrations
├── context/            # Specification documents
└── docker-compose.yml  # Local Postgres setup
```

## Configuration

Environment variables:

- `DB_URL`: PostgreSQL connection string (default: `postgres://kernel:kernel@localhost:5432/kernel?sslmode=disable`)
- `PORT`: HTTP server port (default: `8080`)

## Status

✅ **All Phases Complete**: Full implementation with integration tests

- ✅ Phase 1: Foundation & Infrastructure
- ✅ Phase 2: Store Layer
- ✅ Phase 3: Planner Layer
- ✅ Phase 4: Policy Engine
- ✅ Phase 5: Query Engine
- ✅ Phase 6: Kernel Service
- ✅ Phase 7: API Layer
- ✅ Phase 8: Main Application
- ✅ Phase 9: Integration Tests (10/10 acceptance criteria covered)

## Testing

Comprehensive integration tests are available covering all 10 acceptance criteria. See [TESTING.md](TESTING.md) for details.

Run tests:
```bash
make test
```

## Next Steps

1. ✅ Integration tests (Phase 9) - Complete
2. Add policy set management endpoints
3. Improve error handling and validation
4. Add request logging middleware
5. Performance optimization and indexing
6. Add test isolation (separate DB per test)

## License

[Your License Here]

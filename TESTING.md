# Testing Guide

## Integration Tests

Comprehensive integration tests have been implemented in `internal/kernel/kernel_test.go` covering all 10 acceptance criteria from the specification.

### Test Coverage

**A. Plan/Apply + Audit:**
1. ✅ TestPlanHashingIsDeterministic - Verifies plan hashes are deterministic
2. ✅ TestApplyAppendsOneOperationWithMonotonicSeq - Verifies operations have monotonic sequence numbers
3. ✅ TestReapplySamePlanFailsWithConflict - Verifies plans are single-use

**B. As-of semantics:**
4. ✅ TestAsOfSemanticsForNodes - Verifies nodes are visible at correct sequences
5. ✅ TestAsOfSemanticsForLinks - Verifies links are visible/retired at correct sequences

**C. Policy enforcement:**
6. ✅ TestPolicyDeniesCycle - Verifies acyclic predicate blocks cycles
7. ✅ TestPolicyDeniesWorkItemUnderGoal - Verifies role_edge_allowed predicate
8. ✅ TestPolicyDeniesSecondParent - Verifies child_has_only_one_parent predicate

**D. Diff:**
9. ✅ TestDiffReturnsNetChanges - Verifies diff computation (simplified)

**E. Expand:**
10. ✅ TestExpandReturnsNodesRolesAndLinks - Verifies expand returns nodes, roles, and links

### Running Tests

#### Prerequisites

1. **Start PostgreSQL:**
```bash
docker-compose up -d
```

2. **Run migrations:**
```bash
make migrate
```

#### Run All Tests

```bash
make test
```

Or directly:
```bash
go test -v ./internal/kernel/...
```

#### Run Specific Test

```bash
go test -v ./internal/kernel/... -run TestPlanHashingIsDeterministic
```

#### With Custom Database URL

```bash
DB_URL=postgres://user:pass@host:port/db?sslmode=disable go test -v ./internal/kernel/...
```

### Test Structure

Tests use `TestMain` to set up a shared database connection and kernel service. Each test:
- Creates necessary test data (nodes, links, roles)
- Exercises the kernel service (Plan/Apply)
- Verifies expected behavior
- Cleans up automatically (via database transactions)

### Test Database

Tests use the same database as development by default. For isolated testing:
1. Create a separate test database
2. Set `DB_URL` environment variable
3. Run migrations on test database

## Manual Testing

### Start the Kernel

1. **Start PostgreSQL:**
```bash
docker-compose up -d
```

2. **Run migrations:**
```bash
make migrate
```

3. **Start kernel:**
```bash
make dev
```

Or:
```bash
go run cmd/kernel/main.go
```

The kernel will start on port 8080 (or PORT environment variable).

### Test Endpoints

#### Health Check
```bash
curl http://localhost:8080/v1/healthz
```

#### Create a Plan
```bash
curl -X POST http://localhost:8080/v1/plan \
  -H "Content-Type: application/json" \
  -d '{
    "actor_id": "user:test",
    "capabilities": ["read", "write"],
    "namespace_id": "ProductTree:/Test",
    "asof": {},
    "intents": [
      {
        "kind": "CreateNode",
        "namespace_id": "ProductTree:/Test",
        "payload": {
          "title": "My First Node"
        }
      }
    ]
  }'
```

#### Apply a Plan
```bash
curl -X POST http://localhost:8080/v1/apply \
  -H "Content-Type: application/json" \
  -d '{
    "actor_id": "user:test",
    "capabilities": ["read", "write"],
    "plan_id": "plan:...",
    "plan_hash": "sha256:..."
  }'
```

#### Expand Nodes
```bash
curl "http://localhost:8080/v1/expand?ids=node:123&namespace_id=ProductTree:/Test&depth=1"
```

#### Get History
```bash
curl "http://localhost:8080/v1/history?target=node:123&limit=10"
```

#### Get Diff
```bash
curl "http://localhost:8080/v1/diff?a_seq=10&b_seq=20&target=ProductTree:/Test"
```

## Troubleshooting

### Database Connection Issues

- Ensure PostgreSQL is running: `docker-compose ps`
- Check connection string in `DB_URL`
- Verify migrations ran: `make migrate`

### Test Failures

- Ensure database is clean (tests may interfere with each other)
- Check database logs: `docker-compose logs postgres`
- Run tests with verbose output: `go test -v`

### Port Already in Use

- Change PORT environment variable: `PORT=8081 make dev`
- Or kill existing process on port 8080

## Next Steps

- Add test isolation (separate database per test)
- Add test fixtures for common scenarios
- Add performance/load tests
- Add API integration tests using HTTP client

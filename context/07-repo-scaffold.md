# Repository Scaffold Requirements (v0.1)

## Commands
- `cmd/kernel/main.go` — starts HTTP server
- Expose endpoints in `02-api-contract.md`

## Dev environment
- `docker-compose.yml` providing Postgres
- `Makefile` targets:
  - `make dev` (run kernel)
  - `make test` (run tests)
  - `make migrate` (apply migrations)

## Migrations
- `migrations/0001_init.sql` must create:
  - namespaces
  - nodes
  - role_assignments
  - links
  - materials
  - policy_sets (active policy per namespace)
  - plans (store planned expanded changes and hash)
  - operations (append-only; monotonic seq)

## Logging and errors
- Structured JSON logs
- Stable error codes per API contract

## Testing
- Integration tests for acceptance criteria
- Run against real Postgres

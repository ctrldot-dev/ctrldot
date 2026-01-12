# Futurematic Kernel v0.1 — Cursor Build Packet (Overview)

## What you are building
A Go service (“Futurematic Kernel”) that is the system of record for Nodes/Links/Materials with:
- Append-only **Operations** log (immutable)
- Deterministic **plan → apply** for every mutation
- Namespace-scoped **PolicySets** evaluated at plan-time and rechecked at apply-time
- Addressable history via `asof_seq` / `asof_time` using `created_seq` / `retired_seq`
- Core read APIs: `expand`, `history`, `diff`

The kernel is a **fact engine**. It is small, strict, and boring.

## What you are NOT building (non-goals)
- No workflow engine / state machine
- No AI model execution
- No notification/scheduling system
- No content parsing or full-text indexing in-kernel
- No hidden inference (links are asserted facts only)
- No background daemons besides the HTTP API server

## Dot
“Dot” is a species of shells/clients. This repo builds the kernel only.
- The kernel does not know personas or embodiment.
- The kernel only sees: `ActorID` and `Capabilities`, recorded on each Operation.

## Storage
Use Postgres for v0.1.
- Maintain an append-only operations log
- Maintain projection tables with `created_seq` and nullable `retired_seq`

## Deliverables
- Go service in `cmd/kernel`
- Packages under `internal/` per `04-go-interfaces.md`
- SQL migration(s) under `migrations/` (start with `0001_init.sql`)
- `docker-compose.yml` for local Postgres
- Makefile targets: `make dev`, `make test`, `make migrate`
- Tests implementing acceptance criteria in `06-acceptance-tests.md`

## Hard constraints (do not deviate)
- Every mutating command must be `plan` then `apply`
- `apply` must append exactly one Operation with monotonic `seq`
- All facts must include `created_seq` and `retired_seq` (nullable)
- Policy denies must block apply; warnings must not block (configurable later)
- Plan hash must be deterministic across runs given same inputs

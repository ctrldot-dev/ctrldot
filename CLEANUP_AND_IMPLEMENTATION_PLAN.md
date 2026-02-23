# Ctrl Dot Implementation Plan

## Overview

Based on `context/ctrldot_spec.md`, this document outlines:
1. What to remove from the current codebase
2. What to keep (if anything)
3. Implementation plan for Ctrl Dot v0.1

## Current State Analysis

The current codebase is a **Futurematic Kernel** implementation with:
- **Kernel Engine** (reusable): Plan/Apply workflow, append-only ledger, resolution system, policy evaluation
- Postgres-based storage
- Domain model: Nodes, Links, Materials, Plans, Operations (kernel domain)
- API endpoints: `/v1/plan`, `/v1/apply`, `/v1/expand`, `/v1/history`, `/v1/diff`
- CLI: `dot` command
- Web UI: `web-dot` (React/TypeScript)
- Seed data: Kesteron product/financial ledger examples

## Ctrl Dot Requirements (from spec)

Ctrl Dot needs:
- **Kernel Engine** (reuse as-is): The Plan/Apply/Resolution infrastructure
- **Postgres** storage (keeping existing infrastructure)
- **New Domain Model**: Agents, Sessions, Actions, Decisions, Events (Ctrl Dot domain)
- API endpoints: `/v1/agents/*`, `/v1/actions/propose`, `/v1/events`, etc.
- CLI: `ctrldot` command (different from `dot`)
- Daemon: `ctrldotd` (different from current `kernel`)
- **No UI** in v0.1 (explicitly out of scope)

## Architecture Approach

**Key Insight**: The kernel is the **engine** (Plan → Apply → Operations ledger). Ctrl Dot is a **different domain** that uses this engine.

### How Ctrl Dot Uses the Kernel

```
Ctrl Dot Layer:
  - Action Proposal → Kernel Intent → Kernel Plan
  - Decision → Kernel Apply → Kernel Operation (append-only)
  - Events → Kernel Operations (query operations as events)
  - Resolution Tokens → Kernel Resolutions
  - Agents/Sessions → Stored via kernel operations (as changes)
```

### Component Reuse

**Keep As-Is**:
- `internal/kernel/` - Core engine (Plan, Apply, Resolve methods)
- `internal/planner/` - Intent expansion and deterministic hashing
- `internal/policy/` - Policy evaluation engine (can evaluate Ctrl Dot rules)
- `internal/store/` - Postgres storage layer (extend with Ctrl Dot methods)
- `internal/query/` - Query engine (reuse for event queries)

**Extend**:
- `internal/domain/intent.go` - Add Ctrl Dot intent kinds
- `internal/domain/change.go` - Add Ctrl Dot change kinds
- `internal/kernel/service.go` - Extend `applyChanges` to handle Ctrl Dot changes
- `internal/store/` - Add Ctrl Dot-specific storage methods

**New**:
- `internal/domain/agent.go`, `session.go`, `action.go`, `decision.go`, `event.go` - Ctrl Dot domain types
- `internal/ctrldot/` - Ctrl Dot service layer (wraps kernel, adds limits/rules/loop detection)
- `internal/api/ctrldot/` - Ctrl Dot API handlers
- `cmd/ctrldotd/` - Ctrl Dot daemon
- `cmd/ctrldot/` - Ctrl Dot CLI

## Files/Folders to Remove

### 1. Web UI (not needed for v0.1)
- `web-dot/` - Entire directory
- `start-web-ui.sh` - Script to start web UI

### 2. Seed Data & Bootstrap Scripts (Kesteron-specific)
- `kesteron-*.yaml` - All seed files:
  - `kesteron-assetlink-product-seed.yaml`
  - `kesteron-fieldserve-product-seed.yaml`
  - `kesteron-stablecoinreserves-finledger-seed.yaml`
  - `kesteron-treasury-finledger-seed.yaml`
- `seed-kesteron.sh` - Seed script
- `seed-kesteron-materials.sh` - Materials seed script
- `cmd/bootstrap_finledger/` - Bootstrap command
- `cmd/bootstrap_productledger/` - Bootstrap command
- `cmd/seed_finledger/` - Seed command

### 3. Materials Directory (demo content)
- `materials/` - Entire directory (demo markdown files)

### 4. Policy Sets (current kernel domain-specific)
- `productledger-policyset.yaml` - Product ledger policies
- `financialledger-policyset.yaml` - Financial ledger policies
- `context/productledger-policyset.yaml` - Duplicate
- `context/financialledger-policyset.yaml` - Duplicate
- `context/08-default-policyset.yaml` - Current kernel policy set

### 5. Documentation (current kernel)
- `README.md` - Current kernel README (will be replaced)
- `QUICKSTART.md` - Current kernel quickstart
- `USAGE.md` - Current kernel usage
- `TESTING.md` - Current kernel testing
- `IMPLEMENTATION_PLAN.md` - Old implementation plan
- `SEED_PRODUCT_LEDGER.md` - Seed documentation
- `docs/` - May contain current kernel docs (review first)

### 6. Migrations (Postgres)
- `migrations/` - **KEEP** existing Postgres migrations structure
  - Remove old kernel-specific migrations (Nodes/Links/Materials)
  - Add new Ctrl Dot migrations (Agents/Sessions/Actions/Events)

### 7. Docker Compose (Postgres)
- `docker-compose.yml` - **KEEP** Postgres setup for local development

### 8. Scripts (deployment/demo specific)
- `scripts/` - Review contents, likely remove:
  - `deploy-to-hetzner.sh` - Deployment script
  - `install-dld-systemd-units.sh` - Systemd units
  - `server-seed-materials.sh` - Materials seeding
  - `decision-ledger-auth/` - Auth demo

### 9. Context Files (review)
- `context/` - Review contents:
  - Keep: `ctrldot_spec.md` (the spec!)
  - Remove: Policy sets, seed files, demo configs

### 10. Other
- `examples.sh` - Current kernel examples
- `scaffold/` - Review if needed
- `bin/` - Build artifacts (can regenerate)
- `bootstrap_productledger` - Binary artifact

## What to Keep

### 1. Kernel Engine (Core Infrastructure)
- `internal/kernel/` - **KEEP AS-IS** - Plan/Apply/Resolution orchestration
- `internal/planner/` - **KEEP AS-IS** - Intent expansion and deterministic hashing
- `internal/policy/` - **KEEP AS-IS** - Policy evaluation engine
- `internal/store/` - **KEEP** - Postgres storage layer (extend for Ctrl Dot domain)
- `internal/query/` - **KEEP** - Query engine (may need Ctrl Dot-specific queries)

### 2. Domain Types (Extend)
- `internal/domain/` - **EXTEND** with Ctrl Dot domain types:
  - Keep: `resolution.go`, `proposal.go` (kernel concepts)
  - Add: `agent.go`, `session.go`, `action.go`, `decision.go`, `event.go`
  - Extend: `intent.go`, `change.go` (add Ctrl Dot intent/change kinds)

### 3. Infrastructure
- `go.mod` / `go.sum` - Keep (same module)
- `migrations/` - Keep structure, add Ctrl Dot migrations
- `docker-compose.yml` - Keep Postgres setup
- `Makefile` - Update for Ctrl Dot commands

### 3. Build Infrastructure
- `Makefile` - Update for Ctrl Dot
- `.gitignore` - Keep
- `.cursor/` - Keep (rules)

## Implementation Plan

### Phase 1: Cleanup (First Step)
1. Remove all files/folders listed above
2. Update `go.mod` module name to match Ctrl Dot
3. Clean up any remaining references

### Phase 2: Foundation
1. **Module Setup**
   - Keep module name: `github.com/futurematic/kernel` (or update if needed)
   - Kernel engine stays as-is

2. **Storage Layer** (Postgres)
   - **Extend** `internal/store/` to support Ctrl Dot domain:
     - Add methods: `CreateAgent`, `GetAgent`, `CreateSession`, etc.
     - Add methods: `AppendEvent`, `GetEvents`, `GetEventsByAgent`
     - Add methods: `UpdateLimitsState`, `GetLimitsState`
     - Add methods: `HaltAgent`, `ResumeAgent`
   - Create new Postgres migrations for Ctrl Dot tables:
     - `agents` table
     - `sessions` table
     - `events` table (append-only, or use kernel's `operations` table)
     - `limits_state` table
     - `halted_agents` table
   - Indexes as per spec
   - Reuse existing Postgres connection patterns

3. **Config System**
   - `internal/config/config.go`
   - YAML config loader (`~/.ctrldot/config.yaml`)
   - Default config generation
   - Postgres connection string (via config or `DB_URL` env var)
   - Default: `postgres://kernel:kernel@localhost:5432/kernel?sslmode=disable`

### Phase 3: Domain Model
1. **Core Types** (`internal/domain/`)
   - `agent.go` - Agent struct (Ctrl Dot domain)
   - `session.go` - Session struct (Ctrl Dot domain)
   - `action.go` - Action proposal types (Ctrl Dot domain)
   - `decision.go` - Decision response types (Ctrl Dot domain)
   - `event.go` - Ledger event types (Ctrl Dot domain, maps to kernel Operations)

2. **Extend Kernel Domain** (`internal/domain/`)
   - Extend `intent.go` - Add Ctrl Dot intent kinds:
     - `IntentRegisterAgent`
     - `IntentProposeAction`
     - `IntentCreateSession`
     - etc.
   - Extend `change.go` - Add Ctrl Dot change kinds:
     - `ChangeCreateAgent`
     - `ChangeAppendEvent`
     - `ChangeUpdateLimitsState`
     - etc.

### Phase 4: Kernel Integration
1. **Extend Kernel Service** (`internal/kernel/service.go`)
   - Keep existing Plan/Apply/Resolve methods
   - Extend `applyChanges` to handle Ctrl Dot change kinds:
     - `ChangeCreateAgent` → create agent record
     - `ChangeAppendEvent` → append to events table
     - `ChangeUpdateLimitsState` → update limits
     - etc.
   - Events map to kernel Operations (append-only ledger)

### Phase 5: Ctrl Dot Service Layer
1. **Ctrl Dot Service** (`internal/ctrldot/service.go`)
   - Wraps kernel service
   - Handles action proposal → decision flow
   - Coordinates limits, rules, loop detection
   - Returns decision response

2. **Limits Engine** (`internal/limits/`)
   - `budgets.go` - Budget tracking per agent
   - `thresholds.go` - Warn/throttle/stop thresholds
   - `tokenbucket.go` - Rate limiting (if needed)

3. **Rules Engine** (`internal/rules/`)
   - `rules.go` - YAML rule parsing (Ctrl Dot rules)
   - `match.go` - Action matching against rules
   - Resolution requirement checking
   - Can use kernel's policy engine or separate Ctrl Dot rules

4. **Loop Detection** (`internal/loop/`)
   - `detector.go` - Action hash tracking
   - Sliding window detection
   - Threshold-based triggers

5. **Resolution Tokens** (`internal/resolution/`)
   - `tokens.go` - HMAC-signed tokens
   - Token validation
   - TTL management
   - Maps to kernel Resolutions

### Phase 6: API Layer
1. **HTTP Server** (`internal/api/ctrldot/`)
   - `server.go` - HTTP server setup (port 7777)
   - `handlers_agents.go` - Agent endpoints
   - `handlers_actions.go` - Action proposal endpoint (uses Ctrl Dot service)
   - `handlers_events.go` - Event feed endpoints (queries kernel operations)
   - `handlers_budget.go` - Budget endpoints
   - `middleware.go` - Optional bearer token auth

### Phase 7: Daemon (`cmd/ctrldotd`)
1. **Main Entry Point**
   - Config loading
   - Postgres connection initialization
   - Initialize kernel service (reuse `kernel.NewService`)
   - Initialize Ctrl Dot service (new `ctrldot.NewService`)
   - Run migrations on startup (or separate migrate command)
   - API server startup (Ctrl Dot endpoints on port 7777)
   - Graceful shutdown

### Phase 8: CLI (`cmd/ctrldot`)
1. **CLI Commands**
   - `status` - Daemon status
   - `agents` - List agents
   - `agent show` - Show agent details
   - `halt` / `resume` - Agent control
   - `budget` - Budget info
   - `events tail` - Event feed
   - `rules show/edit` - Rule management
   - `resolve allow-once` - Generate resolution token
   - `daemon start/stop/logs` - Daemon control

### Phase 9: Testing
1. Unit tests for:
   - Threshold evaluation
   - Resolution token validation
   - Loop detection
   - Budget accumulation/reset
2. Integration tests for API endpoints
3. CLI tests

## Key Differences from Current Kernel

| Aspect | Current Kernel | Ctrl Dot |
|--------|---------------|----------|
| **Kernel Engine** | Plan/Apply/Resolution | **Reused as-is** |
| Storage | Postgres | Postgres (same) |
| Domain | Nodes/Links/Materials | Agents/Sessions/Actions/Decisions |
| API | `/v1/plan`, `/v1/apply` | `/v1/actions/propose`, `/v1/agents/*` |
| CLI | `dot` | `ctrldot` |
| Daemon | `kernel` | `ctrldotd` |
| Port | 8080 | 7777 |
| UI | Web Dot (React) | None in v0.1 |
| Workflow | Plan → Apply | Propose → Decision (uses kernel Plan/Apply) |

## Next Steps

1. **Review this plan** and confirm approach
2. **Execute Phase 1** (cleanup) - remove unnecessary files
3. **Start Phase 2** (foundation) - set up new module structure
4. **Implement incrementally** following phases above

## Notes

- **Kernel Engine is reused as-is** - The Plan/Apply/Resolution infrastructure stays unchanged
- Ctrl Dot defines a new domain model (Agents/Sessions/Actions) that uses the kernel engine
- Mapping: Action Proposals → Kernel Intents/Plans, Decisions → Kernel Operations
- Events are stored as kernel Operations (append-only ledger)
- Resolution tokens use kernel's Resolution system
- Extend `applyChanges` in kernel service to handle Ctrl Dot change kinds
- Postgres storage layer extended with Ctrl Dot-specific tables and methods
- Focus on MVP acceptance criteria from spec section 13
- Default port changed from 8080 to 7777 per spec
- Both `kernel` and `ctrldotd` can coexist (different domains, same engine)

# Futurematic Kernel v0.1 - Implementation Plan

## Overview
Build a Go service that serves as an append-only fact engine for Nodes/Links/Materials with:
- Plan → Apply workflow (deterministic planning, immutable operations)
- Policy enforcement via YAML PolicySets
- Addressable history via `asof_seq` / `asof_time`
- Core APIs: plan, apply, expand, history, diff

## Architecture Layers

### 1. Domain Layer (`internal/domain`)
**Purpose**: Core domain types and value objects

**Components**:
- `node.go` - Node type with ID, title, metadata
- `link.go` - Link type (from/to nodes, type, namespace, metadata)
- `material.go` - Material type (content ref, media type, size, hash)
- `role_assignment.go` - RoleAssignment type (node + namespace + role)
- `namespace.go` - Namespace type
- `operation.go` - Operation type (immutable audit record)
- `plan.go` - Plan type (intents, expanded changes, policy report, hash)
- `intent.go` - Intent types (CreateNode, CreateLink, CreateMaterial, etc.)
- `change.go` - Atomic Change types (expanded from intents)
- `asof.go` - AsOf type (seq or time)

**Key Requirements**:
- All types must be JSON-serializable
- Plan hash must be deterministic (canonical JSON)
- Changes must be atomic and reversible

---

### 2. Store Layer (`internal/store`)
**Purpose**: Database persistence, transactions, asof resolution

**Components**:
- `store.go` - Store interface (OpenTx, GetNextSeq, ResolveAsOf)
- `tx.go` - Transaction interface (CRUD operations, atomic commits)
- `postgres.go` - Postgres implementation
- `projections.go` - Projection table updates (nodes, links, materials, roles)
- `operations.go` - Operations log append
- `plans.go` - Plan storage and retrieval
- `policy_sets.go` - PolicySet storage (active policy per namespace)

**Key Requirements**:
- All mutations in single transaction
- Projection updates must include `created_seq` and `retired_seq`
- AsOf resolution: time → seq lookup
- Monotonic seq allocation

---

### 3. Planner Layer (`internal/planner`)
**Purpose**: Expand intents to atomic changes, compute deterministic hash

**Components**:
- `planner.go` - Planner interface (Plan method)
- `expander.go` - Intent expansion logic
- `hasher.go` - Deterministic plan hashing (canonical JSON)
- `validator.go` - Intent validation

**Key Requirements**:
- Deterministic expansion (same inputs → same changes)
- Deterministic hashing (canonical JSON, stable ordering)
- Include policy set hash in plan hash
- Expand intents to atomic Changes

**Intent Types to Support**:
- `CreateNode` → Change: CreateNode
- `CreateLink` → Change: CreateLink
- `CreateMaterial` → Change: CreateMaterial
- `AssignRole` → Change: CreateRoleAssignment
- `RetireNode` → Change: RetireNode
- `RetireLink` → Change: RetireLink
- `Move` → Change: RetireLink + CreateLink (atomic)

---

### 4. Policy Engine (`internal/policy`)
**Purpose**: Evaluate PolicySet YAML against proposed changes

**Components**:
- `engine.go` - PolicyEngine interface (Evaluate method)
- `parser.go` - YAML PolicySet parsing
- `evaluator.go` - Rule evaluation logic
- `predicates.go` - Predicate implementations:
  - `acyclic(link_type)` - Check for cycles in hierarchy
  - `role_edge_allowed(parent_role[], child_role[])` - Validate role transitions
  - `child_has_only_one_parent(child_role, link_type)` - Single parent constraint
  - `has_capability(cap)` - Capability check (stub for v0.1)

**Key Requirements**:
- Load active PolicySet for namespace
- Evaluate rules matching `when` conditions
- Return PolicyReport (denies, warns, infos)
- Denies block apply; warns/infos do not
- Re-check policy at apply time

---

### 5. Query Engine (`internal/query`)
**Purpose**: Read operations (expand, history, diff)

**Components**:
- `engine.go` - QueryEngine interface
- `expand.go` - Expand logic (nodes + roles + links + materials)
- `history.go` - History query (operations for target)
- `diff.go` - Diff logic (changes between two seqs)

**Key Requirements**:
- Support `asof_seq` and `asof_time`
- Expand with depth parameter
- Filter by namespace when provided
- History returns operations (most recent first)
- Diff returns net changes between seqs

---

### 6. Kernel Service (`internal/kernel`)
**Purpose**: Orchestrate plan/apply workflow

**Components**:
- `service.go` - KernelService (Plan, Apply methods)
- `workflow.go` - Plan → Apply orchestration

**Key Requirements**:
- Plan: expand intents, evaluate policy, compute hash, store plan
- Apply: verify plan hash, re-check policy, allocate seq, append operation, update projections (all in one transaction)
- Fail on policy denies
- Fail on plan hash mismatch
- Fail on duplicate apply (plans are single-use)

---

### 7. API Layer (`internal/api`)
**Purpose**: HTTP handlers for REST API

**Components**:
- `server.go` - HTTP server setup
- `handlers.go` - Request handlers:
  - `POST /v1/plan` - Create plan
  - `POST /v1/apply` - Apply plan
  - `GET /v1/expand` - Expand nodes
  - `GET /v1/history` - Get history
  - `GET /v1/diff` - Get diff
  - `GET /v1/healthz` - Health check
- `errors.go` - Error response formatting
- `middleware.go` - Request logging, etc.

**Key Requirements**:
- JSON request/response
- Structured error responses
- Query parameter parsing
- Request validation

---

### 8. Main Application (`cmd/kernel`)
**Purpose**: Application entry point

**Components**:
- `main.go` - Initialize services, start HTTP server

**Key Requirements**:
- Load config (DB connection, port)
- Initialize all layers
- Start HTTP server
- Graceful shutdown

---

## Implementation Phases

### Phase 1: Foundation & Infrastructure
**Goal**: Set up project structure, database, and basic types

1. **Project Setup**
   - Create Go module (`go.mod`)
   - Set up directory structure (`internal/`, `cmd/`, `migrations/`)
   - Create `docker-compose.yml` for Postgres
   - Create `Makefile` with targets: `dev`, `test`, `migrate`

2. **Database Migration**
   - Review `migrations/0001_init.sql` (already exists)
   - Ensure all tables match spec
   - Test migration runs successfully

3. **Domain Types** (`internal/domain`)
   - Implement all domain types (Node, Link, Material, RoleAssignment, Namespace, Operation, Plan, Intent, Change, AsOf)
   - Add JSON tags for serialization
   - Add validation methods

---

### Phase 2: Store Layer
**Goal**: Database persistence and transaction management

1. **Store Interface** (`internal/store`)
   - Define Store and Tx interfaces
   - Implement Postgres connection pool
   - Implement transaction management

2. **Operations Log** (`internal/store/operations.go`)
   - Implement `AppendOperation` (allocate seq, insert)
   - Implement `GetOperation` by seq/ID
   - Implement `GetOperationsForTarget` (for history)

3. **Projection Tables** (`internal/store/projections.go`)
   - Implement node CRUD with `created_seq`/`retired_seq`
   - Implement link CRUD with `created_seq`/`retired_seq`
   - Implement material CRUD with `created_seq`/`retired_seq`
   - Implement role assignment CRUD with `created_seq`/`retired_seq`
   - Implement asof filtering (WHERE created_seq <= asof AND (retired_seq IS NULL OR retired_seq > asof))

4. **AsOf Resolution** (`internal/store/asof.go`)
   - Implement `ResolveAsOf` (time → seq lookup)
   - Query operations table for latest seq at/before time

5. **Plan Storage** (`internal/store/plans.go`)
   - Implement plan storage (insert, get by ID)
   - Track applied status

6. **PolicySet Storage** (`internal/store/policy_sets.go`)
   - Implement PolicySet storage
   - Implement active policy lookup by namespace

---

### Phase 3: Planner Layer
**Goal**: Intent expansion and deterministic hashing

1. **Intent Expansion** (`internal/planner/expander.go`)
   - Implement expansion for each intent type:
     - `CreateNode` → `CreateNode` change
     - `CreateLink` → `CreateNode` change (if nodes don't exist) + `CreateLink` change
     - `CreateMaterial` → `CreateNode` change (if node doesn't exist) + `CreateMaterial` change
     - `AssignRole` → `CreateRoleAssignment` change
     - `RetireNode` → `RetireNode` change
     - `RetireLink` → `RetireLink` change
     - `Move` → `RetireLink` + `CreateLink` changes

2. **Deterministic Hashing** (`internal/planner/hasher.go`)
   - Implement canonical JSON encoding (stable key ordering)
   - Hash: expanded changes + policy set hash + namespace ID + intents (canonicalized)
   - Use SHA256

3. **Planner** (`internal/planner/planner.go`)
   - Implement Plan method:
     - Load active PolicySet for namespace
     - Expand intents to changes
     - Compute plan hash
     - Return Plan with hash

---

### Phase 4: Policy Engine
**Goal**: YAML policy evaluation

1. **Policy Parser** (`internal/policy/parser.go`)
   - Parse YAML PolicySet
   - Validate structure
   - Compute policy hash

2. **Predicate Library** (`internal/policy/predicates.go`)
   - Implement `acyclic(link_type)`:
     - Bounded traversal up PARENT_OF chain
     - Check for cycles
   - Implement `role_edge_allowed(parent_role[], child_role[])`:
     - Check if parent role is in allowed list
     - Check if child role is in allowed list
   - Implement `child_has_only_one_parent(child_role, link_type)`:
     - Query existing parents for child
     - Ensure only one parent exists
   - Implement `has_capability(cap)` (stub for v0.1)

3. **Rule Evaluator** (`internal/policy/evaluator.go`)
   - Match rules by `when` conditions (op, link_type)
   - Evaluate predicates
   - Collect denies, warns, infos

4. **Policy Engine** (`internal/policy/engine.go`)
   - Implement Evaluate method:
     - Load PolicySet for namespace
     - Evaluate all matching rules
     - Return PolicyReport

---

### Phase 5: Query Engine
**Goal**: Read operations (expand, history, diff)

1. **Expand** (`internal/query/expand.go`)
   - Load nodes by IDs (asof-aware)
   - Load roles for nodes in namespace (asof-aware)
   - Load links (adjacent, filtered by namespace, asof-aware)
   - Load materials for nodes (asof-aware)
   - Support depth parameter (recursive expansion)

2. **History** (`internal/query/history.go`)
   - Query operations for target (node ID or namespace ID)
   - Filter by target
   - Return most recent first
   - Support limit

3. **Diff** (`internal/query/diff.go`)
   - Load all facts at `a_seq` (asof-aware)
   - Load all facts at `b_seq` (asof-aware)
   - Compute net changes (created, retired)
   - Return diff

---

### Phase 6: Kernel Service
**Goal**: Orchestrate plan/apply workflow

1. **Kernel Service** (`internal/kernel/service.go`)
   - Implement Plan method:
     - Call planner.Plan
     - Store plan in database
     - Return plan
   - Implement Apply method:
     - Load plan by ID
     - Verify plan hash matches
     - Check if plan already applied (fail if so)
     - Re-evaluate policy (fail on denies)
     - Start transaction
     - Allocate next seq
     - Append operation
     - Apply all changes to projections
     - Mark plan as applied
     - Commit transaction
     - Return operation

---

### Phase 7: API Layer
**Goal**: HTTP REST API

1. **Error Handling** (`internal/api/errors.go`)
   - Define error types (POLICY_DENIED, VALIDATION, NOT_FOUND, CONFLICT, INTERNAL)
   - Implement error response formatting

2. **Handlers** (`internal/api/handlers.go`)
   - `POST /v1/plan`:
     - Parse request (actor_id, capabilities, namespace_id, asof, intents)
     - Call kernel.Plan
     - Return plan JSON
   - `POST /v1/apply`:
     - Parse request (actor_id, plan_id, plan_hash)
     - Call kernel.Apply
     - Return operation JSON
   - `GET /v1/expand`:
     - Parse query params (ids, namespace_id, depth, asof_seq, asof_time)
     - Call query.Expand
     - Return nodes/roles/links/materials JSON
   - `GET /v1/history`:
     - Parse query params (target, limit)
     - Call query.History
     - Return operations JSON
   - `GET /v1/diff`:
     - Parse query params (a_seq, b_seq, target)
     - Call query.Diff
     - Return changes JSON
   - `GET /v1/healthz`:
     - Return `{"ok": true}`

3. **Server** (`internal/api/server.go`)
   - Set up HTTP router (chi or standard library)
   - Register handlers
   - Add middleware (logging, error handling)
   - Start server on configurable port

---

### Phase 8: Main Application
**Goal**: Wire everything together

1. **Main** (`cmd/kernel/main.go`)
   - Load config (env vars or config file):
     - Database connection string
     - Server port
   - Initialize Postgres connection
   - Initialize all layers (store, planner, policy, query, kernel, api)
   - Start HTTP server
   - Implement graceful shutdown (SIGTERM/SIGINT)

2. **Configuration**
   - Use environment variables or config file
   - Defaults: port 8080, localhost Postgres

---

### Phase 9: Testing
**Goal**: Integration tests for acceptance criteria

1. **Test Setup**
   - Use testcontainers or docker-compose for Postgres
   - Test helpers for creating test data

2. **Acceptance Tests** (`internal/kernel/kernel_test.go` or `cmd/kernel/integration_test.go`)
   - Test A: Plan/Apply + Audit
     - Deterministic plan hashing
     - Apply appends one operation with monotonic seq
     - Re-apply same plan fails with CONFLICT
   - Test B: As-of semantics
     - Create node at seq N, query asof_seq=N-1 (not visible), asof_seq=N (visible)
     - Create link, retire it, verify asof behavior
   - Test C: Policy enforcement
     - Create PARENT_OF cycle → denied
     - Link WorkItem under Goal → denied
     - Give WorkItem second parent → denied
   - Test D: Diff
     - Diff between two seqs returns net changes
   - Test E: Expand
     - Expand with depth=1 returns nodes + roles + adjacent links

---

### Phase 10: Dev Environment & Documentation
**Goal**: Make it easy to run and understand

1. **Docker Compose** (`docker-compose.yml`)
   - Postgres service
   - Health check
   - Volume for data persistence

2. **Makefile**
   - `make dev` - Run kernel (with migrations)
   - `make test` - Run tests
   - `make migrate` - Apply migrations
   - `make clean` - Clean up

3. **README.md**
   - Project overview
   - Setup instructions
   - API documentation
   - Development guide

---

## Technical Decisions

### Deterministic Hashing
- Use `encoding/json` with sorted keys (custom encoder or post-processing)
- Include: expanded changes (canonical JSON) + policy set hash + namespace ID + intents (canonical JSON)
- Use SHA256

### Policy Evaluation
- Load active PolicySet for namespace at plan time
- Include policy set hash in plan hash
- Re-evaluate policy at apply time (policy may have changed)
- Denies block apply; warnings do not

### AsOf Resolution
- `asof_time` → query operations table for latest seq at/before time
- Cache seq resolution if needed (but keep it simple for v0.1)

### Transaction Management
- All apply operations in single Postgres transaction
- Append operation + update all projections atomically
- Use `BEGIN` / `COMMIT` / `ROLLBACK`

### Error Handling
- Structured errors with codes (POLICY_DENIED, VALIDATION, etc.)
- JSON error responses per API contract
- Log errors with context

---

## Dependencies

### Go Packages
- `database/sql` + `lib/pq` or `pgx` for Postgres
- `encoding/json` for JSON serialization
- `net/http` for HTTP server (or `chi` for routing)
- `gopkg.in/yaml.v3` for YAML parsing
- `crypto/sha256` for hashing
- `github.com/google/uuid` for ID generation (or custom)

### External Services
- Postgres 14+ (via docker-compose)

---

## Success Criteria

✅ All acceptance tests pass (10 tests from `06-acceptance-tests.md`)
✅ All API endpoints work per contract (`02-api-contract.md`)
✅ Plan hashing is deterministic
✅ Policy enforcement works (denies block apply)
✅ AsOf semantics work correctly
✅ Operations log is append-only
✅ Projections updated atomically with operations

---

## Estimated Complexity

- **Domain Layer**: Low (types and validation)
- **Store Layer**: Medium (transactions, asof resolution)
- **Planner Layer**: Medium (expansion logic, hashing)
- **Policy Engine**: Medium-High (graph traversal for acyclic, rule matching)
- **Query Engine**: Medium (expand with depth, diff computation)
- **Kernel Service**: Low-Medium (orchestration)
- **API Layer**: Low (HTTP handlers)
- **Testing**: Medium (integration tests with Postgres)

**Total Estimated Time**: 2-3 weeks for experienced Go developer

---

## Next Steps

1. Review this plan
2. Start with Phase 1 (Foundation & Infrastructure)
3. Iterate through phases, testing as we go
4. Run acceptance tests after each phase
5. Refine based on findings

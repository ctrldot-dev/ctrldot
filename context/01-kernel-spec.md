# Futurematic Kernel Specification v0.1 (Go)

## 0. Purpose and non-goals
**Purpose:** A Unix-like, append-only fact engine that stores Nodes/Links/Materials and enforces policies at apply time.

**Non-goals:** No workflows, no AI execution, no content parsing, no notifications, no background automation.

## 1. Kernel invariants (MUST)
1. Append-only truth: changes are new facts; no silent overwrites.
2. Plan → Apply: all mutations proposed as Plans, committed only by Apply referencing a plan hash.
3. Deterministic evaluation: same facts + same policy set + same plan => same result.
4. Explicit links only: stored links are asserted, never inferred.
5. Addressable history: all reads support `asof` using seq/time.
6. Namespace-scoped rules: hierarchy rules live in namespace policy sets.

## 2. Core domain objects
- Node: primitive identity + metadata
- RoleAssignment: contextual typing of node within a namespace
- Link: typed relationship (optionally namespace-scoped)
- Material: opaque content ref linked to a node (kernel does not interpret)
- Namespace: context (e.g. ProductTree:/MachinePay)
- Operation: immutable record of applied change (canonical audit)
- Plan: proposed set of expanded atomic changes + policy report + deterministic hash

## 3. Operations
- Each Apply appends exactly one Operation with monotonic `seq`.
- Retirements are tombstones (retired_seq), not deletion.

## 4. Plan and Apply semantics
- `plan` expands intents into atomic Changes, evaluates policy, computes a deterministic hash.
- `apply` must:
  - verify plan hash
  - re-check policy at apply time
  - allocate next seq
  - append Operation
  - apply expanded Changes to projection tables within the same transaction

## 5. Addressable history
- Projection rows include `created_seq` and `retired_seq` (nullable).
- `asof_time` is resolved to seq by selecting latest operation at/before time.
- `yesterday:` is resolved client-side; kernel only needs `asof`.

## 6. Hierarchy
- Hierarchy is represented by `Link{Type:"PARENT_OF", NamespaceID:<ns>}`.
- Tree constraints (acyclic, allowed role edges, single parent) are policy predicates.

## 7. Performance stance
- Postgres store of record for v0.1.
- Reads served from projection tables, not log reconstruction.
- API is batch-friendly (expand, query-batch later).

---

# Kernel Decisions v0.1 (Constitution)

## Invariants
- Append-only Operations; no hidden mutation.
- Plan→Apply boundary.
- Deterministic plan hashing and evaluation.
- Explicit relationships; no inferred links stored as facts.
- History first-class (`created_seq`/`retired_seq` everywhere).

## Data model decisions
- Nodes are primitive; roles are contextual within namespaces.
- Hierarchy is a typed link in a namespace.
- Materials are opaque content refs; kernel never parses.

## Performance principles (without cleverness)
- Postgres + projections first.
- Hot write path is one transaction: append op + write projections.
- Index adjacency; avoid chatty APIs.

## Explicitly deferred
- Closure tables, exotic storage, Kafka, full-text in-kernel, workflow engines, AI in kernel.

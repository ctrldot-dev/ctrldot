# Go Interface Skeleton (v0.1)

Use these interfaces to structure the codebase. Do not collapse into a monolith.

- store.Store / store.Tx: transactions, persistence, asof resolution
- planner.Planner: validate intents, expand to atomic changes, hash deterministically
- policy.Engine: evaluate PolicySet (YAML) via small predicate library
- query.Engine: expand, history, diff, query (minimal DSL ok)
- kernel.Service: orchestrates plan/apply and exposes handlers

## Required packages
- internal/domain
- internal/store
- internal/planner
- internal/policy
- internal/query
- internal/kernel
- internal/api

## Deterministic hashing requirement
- plan hash must use canonical JSON of Expanded changes + policy set hash + namespace id + intent list (canonicalised)
- stable ordering of slices
- stable ordering of map keys

## Projection requirement
- projection tables must be updated in the same transaction as operation append
- each projection row stores created_seq and retired_seq

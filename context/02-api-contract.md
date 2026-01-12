# Kernel HTTP API Contract v0.1

Base URL: `http://<host>:<port>`

## Error format
All non-2xx responses MUST be JSON:
```json
{
  "error": {
    "code": "POLICY_DENIED|VALIDATION|NOT_FOUND|CONFLICT|INTERNAL",
    "message": "Human readable message",
    "details": { }
  }
}
```

Policy denials should include a `policy_report` in `details` when available.

---

## POST /v1/plan
Request:
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

Response: Plan
```json
{
  "id": "plan:01J...",
  "created_at": "2026-01-11T00:00:00Z",
  "actor_id": "user:gareth",
  "namespace_id": "ProductTree:/MachinePay",
  "intents": [ ... ],
  "expanded": [ ... ],
  "class": 1,
  "policy_report": { "denies": [], "warns": [], "infos": [] },
  "hash": "sha256:..."
}
```

Notes:
- The server MUST load the active PolicySet for `namespace_id` (if any) and include its hash in plan hashing.

---

## POST /v1/apply
Request:
```json
{ "actor_id": "user:gareth", "plan_id": "plan:01J...", "plan_hash": "sha256:..." }
```
Response: Operation
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
  "changes": [ ... ]
}
```

Apply MUST fail with `CONFLICT` if the plan hash does not match the stored plan, or if the plan has already been applied (v0.1 rule: plans are single-use).

---

## GET /v1/expand
Query params:
- `ids` comma-separated node IDs (required)
- `namespace_id` optional
- `depth` integer (default 1)
- `asof_seq` optional OR `asof_time` optional (ISO 8601)

Response:
```json
{ "nodes": [...], "roles": [...], "links": [...], "materials": [...] }
```

---

## GET /v1/history
Query params:
- `target` string (node id or namespace id)
- `limit` integer (default 100)

Response: list of Operations (most recent first).

---

## GET /v1/diff
Query params:
- `a_seq` (required)
- `b_seq` (required)
- `target` string (namespace id or node id)

Response:
```json
{ "changes": [...] }
```

---

## GET /v1/healthz
Response:
```json
{ "ok": true }
```

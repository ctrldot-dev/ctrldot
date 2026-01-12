# Dot CLI Output Format v0.1

## Text output principles
- No ANSI colour by default (can add later)
- Stable formatting suitable for copy/paste
- Important data first: ids, seqs, denies

## Printing a Plan
- Header: plan id, hash, class
- Policy denies/warns/infos
- Expanded changes list (one per line)
Example:
```
PLAN plan:01J... hash=sha256:... class=Additive
DENIES:
  - workitem_only_under_job: Work Items must be children of Jobs.
CHANGES:
  + NodeCreated node:abc title="New Goal"
```

## Printing an Operation
```
APPLIED op:01J... seq=42 occurred_at=...
CHANGES:
  + ...
```

## JSON output (`--json`)
Return one JSON object per command on stdout:
- For plan-only commands: `{ "plan": <plan> }`
- For apply: `{ "operation": <op> }`
- For reads: `{ "result": <expand|diff|history> }`
- For status: `{ "ok": true, "server": "...", "actor_id":"...", "namespace_id":"..." }`

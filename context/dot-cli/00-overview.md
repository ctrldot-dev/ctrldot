# Dot CLI (Individual Dot) — Cursor Build Packet v0.1

## Purpose
Build the first individual Dot: a deterministic, scriptable command-line shell (`dot`) that talks to the Futurematic Kernel over HTTP.

This CLI is the primary developer/operator interface for the kernel and should feel `git`-like:
- one-shot commands
- stable output
- explicit plan → apply for every mutation

## Non-goals (v0.1)
- No AI chat mode
- No background daemons
- No interactive TUI (optional later)
- No attempt to be a full workflow tool
- No local system-of-record: kernel is authoritative

## Hard constraints
- All mutating commands MUST call `POST /v1/plan` then `POST /v1/apply`
- If policy denies exist, CLI MUST exit non-zero and MUST NOT apply
- Support `--dry-run` (plan only)
- Support `--yes` to skip confirmation
- Support `--json` to print machine-readable output

## Minimum command set (v0.1)
Connectivity & config:
- `dot status`
- `dot config get <key>`
- `dot config set <key> <value>`
- `dot use <namespace>`
- `dot whereami`

Read:
- `dot show <node-id> [--ns <namespace>] [--asof-seq N|--asof-time ISO] [--depth N]`
- `dot history <target> [--limit N]`
- `dot diff <a> <b> <target>`  (a/b are seq numbers or time aliases)
- `dot ls <node-id> [--ns <namespace>] [--asof-seq N|--asof-time ISO]` (optional convenience; can be implemented via expand + filter PARENT_OF children)

Write:
- `dot new node "<title>" [--meta k=v ...] [--ns <namespace>]`
- `dot role assign <node-id> <role> [--ns <namespace>]`
- `dot link <from-node-id> <type> <to-node-id> [--ns <namespace>]`
- `dot move <child-node-id> --to <parent-node-id> [--ns <namespace>]` (implemented as retire+create PARENT_OF via intents, or as a single Move intent if kernel supports it)

## Exit codes
- 0: success
- 1: usage/validation error (client-side)
- 2: policy denied
- 3: conflict (plan already applied / hash mismatch)
- 4: network/server error

## Config
Store config at `~/.dot/config.json`:

```json
{
  "server": "http://localhost:8080",
  "actor_id": "user:gareth",
  "capabilities": ["read","write:additive"],
  "namespace_id": "ProductTree:/MachinePay"
}
```

Env overrides:
- DOT_SERVER
- DOT_ACTOR
- DOT_NAMESPACE
- DOT_CAPABILITIES (comma-separated)

## Output conventions
- Text output is human readable and stable.
- `--json` outputs a single JSON object per command.
- Every successful `apply` prints: op_id + seq.

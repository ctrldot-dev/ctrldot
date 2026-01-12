# Dot CLI Command Specification v0.1

## Global flags
- `--server <url>` override server
- `--actor <id>` override actor_id
- `--ns <namespace>` override namespace for this command
- `--cap <capabilities>` override capabilities (comma-separated)
- `--json` output JSON
- `-n, --dry-run` plan only (mutations)
- `-y, --yes` skip confirmation (mutations)

## status
`dot status`
- Calls `GET /v1/healthz`
- Prints server URL, actor id, namespace, ok boolean.

## config
`dot config get <key>`
`dot config set <key> <value>`
Keys:
- server
- actor_id
- namespace_id
- capabilities

## use
`dot use <namespace>`
- Sets `namespace_id` in config.

## whereami
`dot whereami`
- Prints resolved config (server, actor, namespace, capabilities).

## show
`dot show <node-id> [--depth N] [--asof-seq N | --asof-time ISO] [--ns <namespace>]`
- Calls `GET /v1/expand?ids=<id>&depth=<N>&namespace_id=<ns>&asof_seq=<N>`
- Prints:
  - Node title + id
  - Roles (if namespace provided)
  - Links (grouped by type)
  - Materials summary

## ls
`dot ls <node-id> [--asof-seq N | --asof-time ISO] [--ns <namespace>]`
- Implement via expand depth=1 then list children where link.type==PARENT_OF and link.from==<node-id> in given namespace.

## history
`dot history <target> [--limit N]`
- Calls `GET /v1/history?target=<target>&limit=<N>`
- Prints: seq, occurred_at, actor_id, op_id, class, change summary.

## diff
`dot diff <a> <b> <target>`
- v0.1: accept only:
  - seq integer
  - `now` (resolve to head seq via `/v1/history?target=<target>&limit=1`)
- Calls `GET /v1/diff?a_seq=<A>&b_seq=<B>&target=<target>`
- Prints created/retired changes.

## new node
`dot new node "<title>" [--meta k=v ...] [--ns <namespace>]`
- Creates intents:
  - CreateNode with payload: title, meta
- Calls plan → prints plan → confirm → apply.

## role assign
`dot role assign <node-id> <role> [--ns <namespace>]`
- Intent AssignRole with payload: node_id, role

## link
`dot link <from> <type> <to> [--ns <namespace>]`
- Intent CreateLink with payload: from, type, to, namespace_id

## move
`dot move <child> --to <parent> [--ns <namespace>]`
- Prefer Intent Move if kernel supports it.
- Else: fetch current parent link via expand; retire link; create new link.

## Mutations: plan/apply flow
For any mutation:
1) Call POST /v1/plan
2) Print plan + policy report
3) If any denies => exit code 2
4) If dry-run => exit code 0
5) Prompt for confirmation unless --yes
6) Call POST /v1/apply
7) Print op result

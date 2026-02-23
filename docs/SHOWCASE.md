# Showcase — 60–90 second demo

Script to demonstrate Ctrl Dot: build, start daemon, register an agent, propose actions (ALLOW vs DENY), show panic mode and a signed bundle.

## Prerequisites

- Go 1.21+
- Repo cloned: `git clone https://github.com/ctrldot-dev/ctrldot.git && cd ctrldot`

## Demo script (copy/paste)

```bash
# 1. Build (≈5 s)
make build-ctrldot

# 2. Start daemon in background (SQLite, no Postgres)
./bin/ctrldot daemon start &
sleep 2
./bin/ctrldot status

# 3. Register agent
curl -s -X POST http://127.0.0.1:7777/v1/agents/register \
  -H "Content-Type: application/json" \
  -d '{"agent_id":"demo","display_name":"Demo Agent"}'

# 4. Propose a safe action → ALLOW
curl -s -X POST http://127.0.0.1:7777/v1/actions/propose \
  -H "Content-Type: application/json" \
  -d '{"agent_id":"demo","intent":{"title":"Read file"},"action":{"type":"tool.call","target":{},"inputs":{}},"cost":{"currency":"GBP","estimated_gbp":0.01,"estimated_tokens":1,"model":""},"context":{"tool":"test"}}' | jq -r '.decision'

# 5. Propose git.push without resolution → DENY + recommendation
curl -s -X POST http://127.0.0.1:7777/v1/actions/propose \
  -H "Content-Type: application/json" \
  -d '{"agent_id":"demo","intent":{"title":"Push"},"action":{"type":"git.push","target":{"repo_path":"/tmp/repo"},"inputs":{}},"cost":{"currency":"GBP","estimated_gbp":0.01,"estimated_tokens":1,"model":""},"context":{}}' | jq '{decision, reason, recommendation: .recommendation.next_steps}'

# 6. Panic on, then propose again (still DENY, stricter)
./bin/ctrldot panic on
curl -s -X POST http://127.0.0.1:7777/v1/actions/propose \
  -H "Content-Type: application/json" \
  -d '{"agent_id":"demo","intent":{"title":"Push"},"action":{"type":"git.push","target":{},"inputs":{}},"cost":{"currency":"GBP","estimated_gbp":0.01,"estimated_tokens":1,"model":""},"context":{}}' | jq -r '.decision'
./bin/ctrldot panic off

# 7. Capabilities (agent discovery)
curl -s http://127.0.0.1:7777/v1/capabilities | jq '{version, runtime_store, ledger_sink, panic}'

# 8. Stop daemon; list and verify a bundle (if autobundle wrote one)
pkill -f ctrldotd || true
./bin/ctrldot bundle ls
# If a path is listed:
# ./bin/ctrldot bundle verify <path>
```

## What to show

- **ALLOW** for low-risk actions (e.g. tool.call).
- **DENY** for high-risk actions (e.g. git.push) with `reason` and `recommendation.next_steps`.
- **Panic mode** toggled with `panic on` / `panic off`; behaviour stays strict until off.
- **GET /v1/capabilities** for agent discovery (no secrets).
- **Bundle** (after a DENY or shutdown): `bundle ls` and `bundle verify <path>`.

## Full smoke test

For a full automated run (build, daemon, register, propose, events, panic, autobundle, shutdown, bundle verify):

```bash
./scripts/smoke-test-ctrldot.sh
```

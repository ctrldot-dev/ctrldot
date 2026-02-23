# Ctrl Dot

**Runtime execution boundary for AI agents.**

Ctrl Dot is a local daemon that sits between agents and their actions. It enforces budget ceilings, loop detection, resolution gating, filesystem/network scope, and deterministic STOP conditions at runtime.

*Autonomy without boundaries is fragility.*

## Quickstart (SQLite, no Postgres)

```bash
git clone https://github.com/ctrldot-dev/ctrldot.git
cd ctrldot
make build-ctrldot
./bin/ctrldot daemon start
```

In another terminal:

```bash
./bin/ctrldot panic on
```

## Integrations

- **OpenClaw:** [adapters/openclaw/GETTING_STARTED.md](adapters/openclaw/GETTING_STARTED.md)
- **CrewAI:** [adapters/crewai/README.md](adapters/crewai/README.md)

## Agent Experience

- **Capabilities:** `GET /v1/capabilities`
- **MCP tool:** `ctrldot-mcp capabilities` (or `ctrldot_capabilities`)
- **Recommendations:** machine-readable `recommendation` object on DENY/STOP (next_steps, docs_hint)
- **Bundles:** signed run artefacts with `README.md` inside each bundle

See [docs/AGENT_EXPERIENCE.md](docs/AGENT_EXPERIENCE.md) for agents.

---

## Why boundaries

Agents that can execute arbitrary actions need a single place to enforce limits and policy. Ctrl Dot gives you:

- **Budget enforcement** — daily spend caps and warn/throttle/stop thresholds per agent
- **Loop detection** — repeated identical actions trigger STOP
- **Resolution gating** — high-risk actions (e.g. `git.push`, `filesystem.delete`) require a short-lived resolution token
- **Deterministic STOP** — DENY/STOP with stable reason codes and a recommendation object so agents (or operators) know what to do next
- **Signed run artefacts** — bundles (and auto-bundles on DENY/STOP) for audit and debugging

## Panic Mode

One switch to tighten behaviour without editing config:

```bash
./bin/ctrldot panic on --ttl 30m --reason "testing new tool"
./bin/ctrldot panic status
./bin/ctrldot panic off
```

When panic is on: budget is clamped, thresholds tightened, resolution forced for more actions, filesystem/network and loop rules use stricter overlays. State persists across daemon restarts. See [docs/SETUP_GUIDE.md](docs/SETUP_GUIDE.md#panic-mode-strict-behaviour).

## Bundles

Signed artefact per run (or per DENY/STOP when auto-bundles are enabled). Each bundle includes events, decisions, config snapshot, and a `README.md` with suggested next steps.

```bash
./bin/ctrldot bundle ls
./bin/ctrldot bundle verify /path/to/bundle
```

## Docs

| Doc | Description |
|-----|-------------|
| [docs/SETUP_GUIDE.md](docs/SETUP_GUIDE.md) | Install, run daemon, CrewAI/OpenClaw, smoke tests |
| [docs/AGENT_EXPERIENCE.md](docs/AGENT_EXPERIENCE.md) | Capabilities, recommendations, panic, resolution, bundles |
| [docs/CONFIG.md](docs/CONFIG.md) | Configuration reference |
| [docs/TROUBLESHOOTING.md](docs/TROUBLESHOOTING.md) | Common issues |
| [docs/DEPLOY.md](docs/DEPLOY.md) | Deploy ctrldot.dev, DNS, GitHub Pages |

## Licence

Apache-2.0. See [LICENSE](LICENSE).

---

## Prerequisites

- **Go 1.21+** (and `make` for `make build-ctrldot`)
- **Optional:** PostgreSQL + Docker for Kernel HTTP sink and legacy runtime store; otherwise SQLite is used by default.

## Build

```bash
make build-ctrldot
# Or:
go build -o bin/ctrldotd ./cmd/ctrldotd
go build -o bin/ctrldot ./cmd/ctrldot
go build -o bin/ctrldot-mcp ./cmd/ctrldot-mcp
```

## CLI (summary)

```bash
./bin/ctrldot status
./bin/ctrldot doctor
./bin/ctrldot agents
./bin/ctrldot panic on | off | status
./bin/ctrldot autobundle status | test
./bin/ctrldot resolve allow-once --agent <id> --action <type> --ttl 10m
./bin/ctrldot bundle ls
./bin/ctrldot bundle verify <path>
```

## API (summary)

- `GET /v1/health` — health check
- `GET /v1/capabilities` — agent discovery (no secrets)
- `POST /v1/agents/register` — register agent
- `POST /v1/actions/propose` — propose action (returns ALLOW / WARN / THROTTLE / DENY / STOP)
- `GET /v1/events` — event feed
- `GET /v1/panic`, `POST /v1/panic/on`, `POST /v1/panic/off`
- `GET /v1/autobundle`, `POST /v1/autobundle/test`

Web UI: `http://127.0.0.1:7777/ui` (when daemon is running).

## Security and contributing

- **Vulnerabilities:** Report to **security@ctrldot.dev** (see [SECURITY.md](SECURITY.md)).
- **Contributing:** [CONTRIBUTING.md](CONTRIBUTING.md) — tests, gofmt, backward compatibility, reason codes.

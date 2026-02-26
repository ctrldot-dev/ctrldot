# Configuration Reference

Ctrl Dot reads configuration from `~/.ctrldot/config.yaml` (or the path in `CTRLDOT_CONFIG`). **If the file doesn't exist, the daemon creates it with defaults on first run** — so after starting `ctrldotd` once, you'll have `~/.ctrldot/config.yaml` to edit (e.g. to change limits or rules; restart the daemon to apply). Missing or partial files are filled with defaults.

## Config file location

- **Default:** `~/.ctrldot/config.yaml`
- **Override:** set `CTRLDOT_CONFIG` to the full path

## Main sections

| Section | Description |
|--------|-------------|
| `server` | `host`, `port` (default 7777) |
| `runtime_store` | `kind`: `sqlite` (default) or `postgres`; `sqlite_path`; `db_url` for Postgres |
| `ledger_sink` | `kind`: `none` (default), `bundle`, or `kernel_http`; `kernel_http.base_url`, `bundle.output_dir`, signing |
| `agents.default` | `daily_budget_gbp`, `warn_pct`, `throttle_pct`, `hard_stop_pct`, `max_iterations_per_action` |
| `display_currency` | `gbp` (default), `usd`, or `eur` — for UI display only; amounts are stored in GBP and converted for display |
| `rules` | `require_resolution` (action types), `filesystem.allow_roots`, `network.deny_all`, `network.allow_domains` |
| `panic` | TTL, max budget, thresholds, resolution/filesystem/network/loop overlays when panic is on |
| `autobundle` | `enabled`, `output_dir`, `debounce_seconds`, `triggers` (on_deny, on_stop, etc.), `include` |

## Environment overrides

| Variable | Effect |
|----------|--------|
| `DB_URL` | Database URL (used for ledger and Postgres runtime store) |
| `PORT` | Server port (e.g. 7777) |
| `CTRLDOT_CONFIG` | Config file path |
| `CTRLDOT_RUNTIME_STORE` | `sqlite` or `postgres` |
| `CTRLDOT_SQLITE_PATH` | Path to SQLite DB file |
| `CTRLDOT_LEDGER_SINK` | `none`, `bundle`, or `kernel_http` |
| `CTRLDOT_KERNEL_URL` | Kernel base URL when sink is `kernel_http` |
| `CTRLDOT_BUNDLE_DIR` | Bundle output directory |
| `CTRLDOT_PANIC` | `1` / `true` / `on` to enable panic at startup |
| `CTRLDOT_PANIC_TTL` | Panic TTL in seconds |
| `CTRLDOT_PANIC_BUDGET_USD` | Max daily budget (USD) when panic is on |
| `CTRLDOT_AUTOBUNDLE` | `1` / `true` / `on` to enable auto-bundles |
| `CTRLDOT_AUTOBUNDLE_DIR` | Auto-bundle output directory |

## Example minimal config (YAML)

```yaml
server:
  host: 127.0.0.1
  port: 7777

runtime_store:
  kind: sqlite
  sqlite_path: ~/.ctrldot/ctrldot.sqlite

ledger_sink:
  kind: none

agents:
  default:
    daily_budget_gbp: 10.0
    warn_pct: [0.70, 0.90]
    throttle_pct: 0.95
    hard_stop_pct: 1.00

rules:
  require_resolution:
    - git.push
    - filesystem.delete
  filesystem:
    allow_roots:
      - ~/dev
  network:
    deny_all: true
    allow_domains:
      - api.openai.com
```

See [SETUP_GUIDE.md](SETUP_GUIDE.md) for run modes (SQLite, bundle sink, kernel_http, panic, autobundle).

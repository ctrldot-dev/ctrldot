# Troubleshooting

## Daemon won’t start

- **Port in use:** Change port with `PORT=7778 ./bin/ctrldotd` or set `server.port` in config.
- **Config path:** Ensure `~/.ctrldot/config.yaml` exists or set `CTRLDOT_CONFIG` to a valid path. The daemon will run with defaults if the file is missing.
- **SQLite path:** For default SQLite runtime, the daemon creates `~/.ctrldot/` if needed. If you set `CTRLDOT_SQLITE_PATH`, ensure the parent directory exists and is writable.
- **Postgres:** If using `CTRLDOT_RUNTIME_STORE=postgres`, ensure `DB_URL` is set and Postgres is reachable; run migrations (e.g. `migrations/0007_ctrldot_tables.sql`) before starting.

## Database locked (SQLite)

With SQLite (default runtime store), use WAL mode to reduce lock contention. The embedded SQLite driver typically enables WAL. If you see "database is locked":

- Ensure only one daemon process is using the same SQLite file.
- Close other tools that might hold the DB open.
- Restart the daemon once other processes release the file.

## OpenClaw tool-calls not working with provider X

- **OpenAI:** Tool invocation is supported; use an OpenAI model (e.g. gpt-4.1-mini) for smoke-testing `ctrldot_propose` / `ctrldot_exec`.
- **Other providers:** Some providers do not yet emit or parse tool_calls in the way OpenClaw expects. If the model sees tools but no tool_call reaches the Ctrl Dot plugin, try OpenAI for the integration path, or check OpenClaw/docs for provider-specific behaviour.
- **Gateway env:** Ensure API keys (e.g. `OPENAI_API_KEY`) are available to the OpenClaw gateway process (e.g. in `~/.openclaw/.env` or in the config `env` block).

## kernel_http sink failures

- When `ledger_sink` is `kernel_http`, decision records are POSTed to the Kernel. If the Kernel is down or unreachable:
  - With `required: false` (default), failures are logged but do not change the decision; the daemon keeps running.
  - With `required: true`, sink failures may affect behaviour (see config).
- Fix: ensure the Kernel is running and `CTRLDOT_KERNEL_URL` (or `ledger_sink.kernel_http.base_url`) is correct; check daemon logs for POST errors.

## Panic mode not applying

- Confirm state: `./bin/ctrldot panic status`. If it shows disabled, turn on with `./bin/ctrldot panic on`.
- If panic was enabled with a TTL, it may have expired; check `expires_at` in `GET /v1/panic` or panic status. Enable again if needed.
- Ensure the daemon was restarted after config changes to `panic:` (e.g. thresholds, overlays). Runtime panic state is persisted in the runtime store; config changes apply when panic is on according to `internal/config/effective.go`.

## Bundles or autobundles not created

- **Autobundle:** Check `./bin/ctrldot autobundle status` (or `GET /v1/autobundle`). Ensure `enabled` is true and the trigger (e.g. on_deny, on_shutdown) is enabled.
- **Output dir:** If `output_dir` is set, it must be writable; otherwise the default is under `~/.ctrldot/`.
- **Signing:** If bundle signing is enabled, the key path must exist and be readable; otherwise bundle creation may fail. See config `ledger_sink.bundle.sign`.

## Resolution token rejected

- Tokens are short-lived (default TTL, e.g. 10 minutes). Generate a new one with `./bin/ctrldot resolve allow-once --agent <id> --action <type> --ttl 10m`.
- Ensure the agent ID and action type match the proposal. Use the same resolution secret (config) on the daemon that generated the token.

# Agent Experience — Ctrl Dot

This doc is for **agents and tooling** that need to discover Ctrl Dot, interpret “why blocked” signals, and suggest enablement steps.

---

## How agents detect Ctrl Dot

**HTTP:** `GET /v1/capabilities` returns a JSON object describing the daemon and its effective configuration (no secrets).

```bash
curl http://127.0.0.1:7777/v1/capabilities
```

Response shape (v1):

- `ctrldot.version`, `ctrldot.build` (git_sha, built_at)
- `ctrldot.api.base_url`, `ctrldot.api.version`
- `ctrldot.runtime_store.kind` (e.g. `sqlite`), `ctrldot.runtime_store.sqlite_path`
- `ctrldot.ledger_sink.kind` (`none` | `bundle` | `kernel_http`), `ctrldot.ledger_sink.bundle_dir`
- `ctrldot.panic.enabled`, `ctrldot.panic.expires_at`, `ctrldot.panic.effective` (when panic is on)
- `ctrldot.features`: `resolution_tokens`, `loop_detector`, `budget_limits`, `rules_engine`, `auto_bundles`, `bundle_verify`

**MCP / CLI:** If using the Ctrl Dot MCP CLI:

```bash
ctrldot-mcp capabilities
# or
CTRLDOT_URL=http://127.0.0.1:7777 ctrldot-mcp capabilities
```

Output is the same JSON as the API.

---

## How to interpret recommendation objects

When `POST /v1/actions/propose` returns **DENY**, **STOP**, or **THROTTLE**, the response can include:

- **`reasons`** — Array of `{ "code": "...", "message": "..." }` with stable codes, e.g.:
  - `PANIC_RESOLUTION_REQUIRED`, `RESOLUTION_REQUIRED` — action requires a resolution token
  - `NETWORK_DOMAIN_DENIED` — domain not in allowlist
  - `FILESYSTEM_DENIED` — path not under allow_roots
  - `LOOP_STOP_THRESHOLD` — action repeated too many times
  - `BUDGET_STOP_THRESHOLD` — daily budget exceeded
  - `AGENT_HALTED` — agent was halted via API

- **`recommendation`** — Object with:
  - **`kind`** — `use_resolution` | `enable_panic` | `tighten_scope` | `reduce_loop` | `enable_ctrldot`
  - **`title`**, **`summary`** — Short human/agent-readable explanation
  - **`next_steps`** — List of runnable commands (e.g. `ctrldot resolve allow-once --agent <id> --ttl 120s`, `ctrldot panic off`)
  - **`docs_hint`** — Optional path into docs (e.g. `docs/SETUP_GUIDE.md#panic-mode`)
  - **`tags`** — e.g. `["panic", "resolution", "exec"]`

- **`autobundle_path`**, **`autobundle_trigger`** — When an auto-bundle was written, path and trigger (e.g. `decision_deny`, `decision_stop`). Agents can cite this path for sharing or debugging.

Use **`recommendation.next_steps`** to suggest exact commands to the user. Use **`reasons[].code`** for programmatic handling (e.g. “if PANIC_RESOLUTION_REQUIRED, show resolve instructions”).

---

## When to enable panic mode

Panic mode tightens budget, resolution, filesystem/network, and loop detection without editing config.

- **Enable:** `ctrldot panic on` (optional: `--ttl 30m`, `--reason "..."`)
- **Disable:** `ctrldot panic off`
- **Status:** `ctrldot panic status`

Use panic when running untrusted prompts, unattended agents, or new adapters. When panic is on, many actions will require resolution or be denied until the user runs `ctrldot resolve allow-once` or disables panic.

---

## How to use resolution tokens

Actions that require resolution (e.g. exec, git.push, filesystem.write) are DENY until a short-lived token is issued.

1. User runs: `ctrldot resolve allow-once --agent <agent_id> --ttl 120s` (or equivalent via API).
2. The client then sends the same proposal again with the resolution token set (or the adapter obtains the token and retries).
3. Ctrl Dot returns ALLOW for that one action within the TTL.

Recommendation objects for “resolution required” denials include a **next_steps** entry like:

`ctrldot resolve allow-once --agent <agent_id> --ttl 120s`

---

## How bundles help (sharing / debugging)

When auto-bundles are enabled, DENY/STOP (and shutdown) produce a **signed bundle** directory containing:

- `manifest.json` (trigger, timestamps, hashes)
- `decision_records.jsonl`, `events.jsonl`, optional `config_snapshot.yaml`
- **`README.md`** — Summary of what happened, reason codes, suggested next steps, and verify command
- `signature.ed25519`, `public_key.ed25519`

The propose response can include **`autobundle_path`** when a bundle was written. Agents can:

- Tell the user: “A signed bundle was created at \<path\>.”
- Suggest: “Verify with: `ctrldot bundle verify \<path\>`.”

The README inside the bundle repeats the suggested next steps so the recipient can run them without calling the API again.

---

## OpenClaw: reading recommendation in tool output

The `ctrldot_propose` tool returns the full decision JSON. When the decision is DENY or STOP, the tool **prepends** a short block so the agent sees:

- **Recommendation:** \<summary\>
- **Next steps:** \<list of commands\>
- **Bundle:** \<path\> (if present)

Then the full JSON follows. So the model can both reason over the structured fields (`reasons`, `recommendation`, `autobundle_path`) and surface the same to the user in natural language.

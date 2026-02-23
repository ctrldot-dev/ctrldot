# Ctrl Dot Integration Setup Guide

## Full Installation & Testing Steps (CrewAI + OpenClaw)

This document contains **all steps required** to:

1. Install system prerequisites
2. Start Ctrl Dot daemon
3. Install and test CrewAI integration
4. Install and test OpenClaw integration
5. Build plugins
6. Configure guardrails
7. Run smoke tests

You can run all commands below in the integrated terminal.

---

# 0️⃣ Start Ctrl Dot Daemon (Required for adapters)

Before running CrewAI or OpenClaw with Ctrl Dot, start the daemon.

## Quick start (no Postgres)

Ctrl Dot runs with **SQLite by default** — no Docker or Postgres required.

**Terminal 1 — from repo root** (e.g. `cd /Volumes/FutureBox/Code/CtrlDot`):

```bash
go build -o bin/ctrldot ./cmd/ctrldot && go build -o bin/ctrldotd ./cmd/ctrldotd
./bin/ctrldot daemon start
```

This creates `~/.ctrldot/ctrldot.sqlite` if needed and starts the daemon. Check: `./bin/ctrldot status` or `./bin/ctrldot doctor`.

## Connected mode (Kernel + Postgres)

To send decision records to the Kernel (e.g. for ledger or web UI), use Postgres and the Kernel HTTP sink:

```bash
# Start Postgres (if not already running)
docker-compose up -d

# Run Ctrl Dot migrations (first time only)
docker exec -i futurematic-kernel-postgres psql -U kernel -d kernel < migrations/0007_ctrldot_tables.sql

# Build and start daemon with Kernel sink
go build -o bin/ctrldotd ./cmd/ctrldotd
CTRLDOT_RUNTIME_STORE=postgres CTRLDOT_LEDGER_SINK=kernel_http DB_URL="postgres://kernel:kernel@localhost:5432/kernel?sslmode=disable" ./bin/ctrldotd
```

Or use the Makefile (if ctrldot targets are defined):

```bash
make ctrldot-up
```

## Bundle mode (signed artefact)

To write decision records to a signed bundle under `~/.ctrldot/bundles/`:

```bash
CTRLDOT_LEDGER_SINK=bundle ./bin/ctrldotd
```

Then: `./bin/ctrldot bundle ls` and `./bin/ctrldot bundle verify <path>`.

## Panic Mode (strict behaviour)

Panic mode tightens budget, resolution, filesystem/network rules, and loop detection without editing YAML.

- **Turn on:** `./bin/ctrldot panic on` (optional: `--ttl 30m`, `--reason "trying new tool"`)
- **Status:** `./bin/ctrldot panic status`
- **Turn off:** `./bin/ctrldot panic off`

State persists across daemon restarts. When enabled, effective limits and rules are applied automatically (see config `panic:` and env `CTRLDOT_PANIC`, `CTRLDOT_PANIC_TTL`, `CTRLDOT_PANIC_BUDGET_USD`).

## Auto-bundles

With autobundle enabled (default), the daemon writes a signed bundle on DENY/STOP, loop/budget stop, and shutdown (even when `ledger_sink=none`). Same format as the bundle sink; `./bin/ctrldot bundle verify <path>` works.

- **Status:** `./bin/ctrldot autobundle status`
- **Force one bundle:** `./bin/ctrldot autobundle test`

Configure in config under `autobundle:` or env `CTRLDOT_AUTOBUNDLE`, `CTRLDOT_AUTOBUNDLE_DIR`. Debounce (default 10s) limits one bundle per session per trigger type.

---

Leave the daemon running in one terminal. In **another terminal**, run the setup steps below. **Both terminals should start in the repo root** (the directory that contains `bin/`, `adapters/`, and `.venv`), e.g.:

```bash
cd /path/to/CtrlDot   # or: cd /Volumes/FutureBox/Code/CtrlDot
```

---

# 1️⃣ System Prerequisites

## 1. Install Python (3.10+)

Check:

```bash
python3 --version
```

If not installed:

- **macOS:** `brew install python`
- **Ubuntu/Debian:** `sudo apt update && sudo apt install python3 python3-venv python3-pip -y`

## 2. Install Node.js (18+)

Check:

```bash
node --version
npm --version
```

If not installed:

- **macOS:** `brew install node`
- **Ubuntu/Debian:** `sudo apt install nodejs npm -y`

---

# 2️⃣ CrewAI Adapter Setup

## Step 1 — Create Python Virtual Environment

Use **Python 3.10–3.12** for CrewAI (3.14 may require Rust for the `tiktoken` dependency). From repo root:

```bash
# Prefer Python 3.12 if installed via Homebrew (fewer build issues)
python3.12 -m venv .venv   # or: python3 -m venv .venv
source .venv/bin/activate   # Windows: .venv\Scripts\activate
```

## Step 2 — Install CrewAI

```bash
pip install --upgrade pip
pip install crewai
```

## Step 3 — Install Ctrl Dot CrewAI Adapter

```bash
pip install -e adapters/crewai
```

## Step 4 — Run Example

Ensure the Ctrl Dot daemon is running in the other terminal (see section 0). **Terminal 2 — from repo root:**

```bash
cd /path/to/CtrlDot
source .venv/bin/activate
python adapters/crewai/examples/crew_minimal.py
```

**One-shot setup (venv + install + run example):**

```bash
./scripts/setup-crewai.sh
```

**Expected result:**

- Ctrl Dot daemon logs `llm.call`
- Tool calls are proposed before execution
- Deny rules trigger exceptions correctly

**Troubleshooting "Cannot connect" / "Remote end closed connection":**

1. **Confirm the daemon is running** in the other terminal and is really `ctrldotd`:
   ```bash
   curl -s http://127.0.0.1:7777/v1/health
   ```
   You should see `{"ok":true,"version":"0.1.0"}`.

2. **Only one process on 7777:** If you saw "address already in use", another process is bound to 7777. Stop it (`lsof -ti :7777 | xargs kill`) then start `ctrldotd` again so the Python example talks to the real daemon.

3. **Same machine:** Run the example from the same machine where the daemon is running (both use `127.0.0.1:7777`).

---

# 3️⃣ OpenClaw Integration Setup

## Step 1 — Install OpenClaw

- **Homebrew (recommended on macOS/Linux):** `brew install openclaw-cli` — then `openclaw onboard --install-daemon` and run `openclaw gateway` (see [docs.openclaw.ai/install](https://docs.openclaw.ai/install)).
- **npm:** `npm install -g openclaw` then `openclaw onboard --install-daemon` if needed.
- **Installer script:** `curl -fsSL https://openclaw.ai/install.sh | bash`
- **From source:** `git clone https://github.com/openclaw/openclaw && cd openclaw && pnpm install && pnpm build`

## Step 2 — Build Ctrl Dot Plugin

From repo root:

```bash
cd adapters/openclaw/plugin
npm install
npm run build
cd ../../..
```

Or use the script from repo root:

```bash
./scripts/setup-openclaw-plugin.sh
```

Plugin output is in `adapters/openclaw/plugin/dist/`.

## Step 3 — Configure OpenClaw

Edit OpenClaw config (e.g. `~/.openclaw/openclaw.json`). Replace `<ABSOLUTE_PATH_TO_REPO>` with your Ctrl Dot repo path (e.g. `/Volumes/FutureBox/Code/CtrlDot`).

```json
{
  "tools": {
    "profile": "coding",
    "deny": ["group:runtime", "group:web"]
  },
  "plugins": {
    "enabled": true,
    "allow": ["ctrldot"],
    "load": { "paths": ["<ABSOLUTE_PATH_TO_REPO>/adapters/openclaw/plugin"] },
    "entries": {
      "ctrldot": {
        "enabled": true,
        "config": {
          "ctrldotUrl": "http://127.0.0.1:7777",
          "agentId": "openclaw",
          "agentName": "OpenClaw Agent"
        }
      }
    }
  }
}
```

Restart OpenClaw after saving.

## Step 4 — Install Skill

Copy the Ctrl Dot skill into OpenClaw’s skills directory:

```bash
./scripts/install-openclaw-skill.sh
```

Or manually:

```bash
mkdir -p ~/.openclaw/skills
cp -r adapters/openclaw/skills/ctrldot-guard ~/.openclaw/skills/
```

Restart OpenClaw.

---

# 4️⃣ Smoke Tests

## Test 0 — Ctrl Dot CLI + SQLite + bundle (no adapters)

From repo root, run the smoke script (builds ctrldot/ctrldotd, starts daemon with SQLite + bundle sink, register/propose/events, verifies bundle):

```bash
./scripts/smoke-test-ctrldot.sh
```

Expected: exit 0, “Smoke test passed”, and at least one bundle under `$CTRLDOT_BUNDLE_DIR` (default `/tmp/ctrldot-smoke-bundles`).

## Test 1 — Runtime Guard

Ask the OpenClaw agent: **"List files in workspace."**

**Expected:** Agent uses `ctrldot_exec`, Ctrl Dot receives the proposal, command runs only if allowed.

## Test 2 — Deny Rule

In Ctrl Dot config (e.g. `~/.ctrldot/config.yaml`), add a rule requiring resolution for `tool.call.exec`. Repeat the “List files” test.

**Expected:** Execution denied, agent receives an error.

## Test 3 — Web Fetch Guard

Ask the agent: **"Fetch example.com"**

**Expected:** `ctrldot_web_fetch` is used, response size is capped, proposal is logged.

---

# 5️⃣ Optional Advanced Tests

Simulate:

- High-frequency tool calls
- Looping agent tasks
- Budget exceed threshold

Confirm:

- WARN at 70–85%
- THROTTLE mode applied
- STOP at hard cap

---

# 6️⃣ How to deploy (ctrldot.dev)

The minimal marketing site in `site/` is deployed to **GitHub Pages** via the workflow `.github/workflows/pages.yml` (runs on push to `main`). No build step — the `site/` folder is uploaded as-is.

1. In the repo: **Settings → Pages** → set source to **GitHub Actions**.
2. Push to `main`; the workflow deploys `site/` to Pages.
3. Add custom domain `ctrldot.dev` in Pages settings and enable **Enforce HTTPS**.
4. At your DNS registrar, add the A and CNAME records from **[docs/DEPLOY.md](DEPLOY.md)** (DNS table and full steps).

After DNS propagates, `https://ctrldot.dev` will serve the static site. Optional: add `site/.well-known/security.txt` for a security contact (see DEPLOY.md).

---

# 7️⃣ Directory Summary

After setup, the repo should include:

```
adapters/
  crewai/
  openclaw/
.venv/
```

Plugin compiled to:

```
adapters/openclaw/plugin/dist/
```

---

# 8️⃣ Success Criteria

- CrewAI example runs
- OpenClaw loads plugin
- Built-in runtime tools denied
- Guarded tools enforce Ctrl Dot decisions
- Deny/Throttle paths behave correctly

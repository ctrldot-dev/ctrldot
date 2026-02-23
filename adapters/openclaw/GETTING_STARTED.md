# Getting the OpenClaw Integration Working

Checklist to get Ctrl Dot guarding OpenClaw tool execution.

---

## Prerequisite: OpenClaw

You need **OpenClaw** installed and runnable. The Ctrl Dot plugin is built to load into OpenClaw; we don’t ship OpenClaw.

- **Homebrew (macOS/Linux):** `brew install openclaw-cli` — installs the OpenClaw CLI (openclaw.ai). Then run `openclaw onboard --install-daemon` and use `openclaw gateway` (or as per [docs.openclaw.ai/install](https://docs.openclaw.ai/install)).
- **npm:** `npm install -g openclaw` then run `openclaw` (and `openclaw onboard --install-daemon` if needed).
- **Installer script:** `curl -fsSL https://openclaw.ai/install.sh | bash`
- **From source:** clone the [OpenClaw repo](https://github.com/openclaw/openclaw), `pnpm install && pnpm build`, then run per that project’s instructions.

The plugin uses **local type stubs** (no `@openclaw/core` on npm), so it builds without OpenClaw. When you run OpenClaw, it must support loading plugins from a path and the config shape below (or the equivalent in your OpenClaw version).

---

## 1. Run Ctrl Dot daemon

Ctrl Dot runs with **SQLite by default** — no Postgres required. In a terminal (from the Ctrl Dot repo root):

```bash
cd /Volumes/FutureBox/Code/CtrlDot
go build -o bin/ctrldot ./cmd/ctrldot && go build -o bin/ctrldotd ./cmd/ctrldotd
./bin/ctrldot daemon start
```

Leave it running. Check: `curl -s http://127.0.0.1:7777/v1/health` → `{"ok":true,"version":"0.1.0"}` or `./bin/ctrldot status`.

**Optional (Kernel ledger):** To send decisions to the Kernel, use Postgres and set `CTRLDOT_RUNTIME_STORE=postgres` and `CTRLDOT_LEDGER_SINK=kernel_http` with `DB_URL=...`; see `docs/SETUP_GUIDE.md`.

---

## 2. Build the Ctrl Dot plugin

In another terminal, from the **Ctrl Dot repo root**:

```bash
cd /Volumes/FutureBox/Code/CtrlDot
./scripts/setup-openclaw-plugin.sh
```

Or manually:

```bash
cd adapters/openclaw/plugin
npm install
npm run build
```

You should see `adapters/openclaw/plugin/dist/` with `index.js` (and other compiled files).

---

## 3. Configure OpenClaw to load the plugin

OpenClaw needs a config file that:

- Loads the Ctrl Dot plugin from an **absolute path**
- Denies the built-in runtime/web tools you want to replace
- Enables the Ctrl Dot tools

**Config location** (depends on how OpenClaw is run):

- Often: `~/.openclaw/openclaw.json`
- Or a path your OpenClaw docs specify

**Example config** (adjust paths and top-level keys to match your OpenClaw version):

**If your OpenClaw uses `tools` + `plugins` (like the full setup guide):**

Create or edit `~/.openclaw/openclaw.json`:

```json
{
  "tools": {
    "profile": "coding",
    "deny": ["group:runtime", "group:web"]
  },
  "plugins": {
    "enabled": true,
    "allow": ["ctrldot"],
    "load": {
      "paths": ["/Volumes/FutureBox/Code/CtrlDot/adapters/openclaw/plugin"]
    },
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

Replace `/Volumes/FutureBox/Code/CtrlDot` with your **actual Ctrl Dot repo path** if different.

**If your OpenClaw uses a different schema** (e.g. `plugins` as an array with `path` and `config`), use the structure that your OpenClaw docs or UI expect; the important parts are:

- Plugin path: `.../adapters/openclaw/plugin` (the directory that contains `dist/index.js`).
- Plugin config: `ctrldotUrl`, `agentId`, `agentName` as above.
- Built-in `exec` / `web_fetch` denied (or the groups that contain them).
- Ctrl Dot tools `ctrldot_exec`, `ctrldot_web_fetch`, `ctrldot_propose` allowed.

You can copy from:

- `adapters/openclaw/examples/openclaw-home-config.example.json` (tools + plugins object)
- `adapters/openclaw/examples/openclaw.json.example` (plugins array style)

and then fix the `<ABSOLUTE_PATH_TO_REPO>` (or equivalent) to your path.

---

## 4. Install the Ctrl Dot skill

So the agent is instructed to use the guarded tools:

```bash
cd /Volumes/FutureBox/Code/CtrlDot
./scripts/install-openclaw-skill.sh
```

This copies `adapters/openclaw/skills/ctrldot-guard` into `~/.openclaw/skills/`. If OpenClaw uses a different skills directory, copy the folder there instead:

```bash
cp -r adapters/openclaw/skills/ctrldot-guard ~/.openclaw/skills/
```

---

## 5. Restart OpenClaw

Restart OpenClaw so it:

- Loads the new config
- Loads the plugin from the path you set
- Picks up the skill

How you “restart” depends on how you run OpenClaw (e.g. quit and run `openclaw` again, or restart a service).

---

## 6. Smoke test

With the daemon still running and OpenClaw using the plugin and skill:

- Ask the agent to **list files** (e.g. “List files in the workspace” or “Run ls in /tmp”). It should use `ctrldot_exec` and you should see the proposal (and event) in Ctrl Dot.
- Ask it to **fetch a URL** (e.g. “Fetch https://example.com”). It should use `ctrldot_web_fetch`.

If the agent uses the built-in `exec` or `web_fetch` instead, check:

- Config: `deny` for the right tool groups, `allow` / plugin config so `ctrldot_exec` and `ctrldot_web_fetch` are available.
- Skill is in the right place and enabled so the agent is instructed to use the guarded tools.

---

## DeepInfra or custom provider: 401 from model

If the agent returns "401 status code (no body)" and your API key works with `curl` to the provider:

- The **gateway** process may not be resolving `$DEEPINFRA_API_KEY` (or your provider’s env var) from `~/.openclaw/.env` or the process environment.
- **Workaround:** In `~/.openclaw/openclaw.json`, either:
  - Add an `env` block at the top level: `"env": { "DEEPINFRA_API_KEY": "your-key" }`, or
  - Set the key directly in `models.providers.<name>.apiKey` (e.g. `"apiKey": "your-key"`).
- Restart the gateway after changing config. Prefer storing the key in a file that is not committed (e.g. `.env`) and, if you put it in the config, restrict permissions: `chmod 600 ~/.openclaw/openclaw.json`.

---

## If OpenClaw isn’t available yet

The plugin is built and ready to load. We don’t have a public OpenClaw package; the adapter was written to the spec in `context/ctrldot_adapters_cursor_combined.md`. When you have OpenClaw (from a repo, private npm, or future public release):

1. Use the same steps above.
2. If the config or plugin API differs, adjust the config shape and/or the plugin’s `dist` entry (e.g. `openclaw.plugin.json` and `src/index.ts`) to match OpenClaw’s expected format.

The plugin’s **runtime** dependency on OpenClaw is only that OpenClaw loads it and calls the registered tools; it talks to Ctrl Dot over HTTP, so no OpenClaw-specific package is required for the build.

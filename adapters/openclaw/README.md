# Ctrl Dot OpenClaw Integration

OpenClaw plugin and skill for guarding tool execution via Ctrl Dot.

## Overview

This integration:
- Replaces risky built-in tools (`exec`, `web_fetch`) with Ctrl Dot guarded versions
- Provides a skill that instructs agents to use guarded tools
- Denies built-in risky tools and only allows Ctrl Dot replacements

## Setup

### 1. Start Ctrl Dot Daemon

Ensure Ctrl Dot daemon is running:

```bash
DB_URL="postgres://kernel:kernel@localhost:5432/kernel?sslmode=disable" PORT=7777 ./bin/ctrldotd
```

### 2. Build Plugin

The plugin uses local type stubs (no `@openclaw/core` dependency required to build). When OpenClaw is available, you can switch to the real package.

```bash
cd adapters/openclaw/plugin
npm install
npm run build
```

### 3. Configure OpenClaw

For `~/.openclaw/openclaw.json`, use `examples/openclaw-home-config.example.json` (tools/plugins structure). For other config formats, use `examples/openclaw.json.example`. Replace `<ABSOLUTE_PATH_TO_REPO>` with your repo path.

```json
{
  "plugins": [
    {
      "id": "ctrldot",
      "path": "/absolute/path/to/adapters/openclaw/plugin",
      "config": {
        "ctrldotUrl": "http://127.0.0.1:7777",
        "agentId": "openclaw-agent",
        "agentName": "OpenClaw Agent"
      }
    }
  ],
  "toolGroups": {
    "runtime": {
      "enabled": false,
      "deny": ["exec"]
    },
    "web": {
      "enabled": false,
      "deny": ["web_fetch"]
    },
    "ctrldot": {
      "enabled": true,
      "allow": ["ctrldot_exec", "ctrldot_web_fetch", "ctrldot_propose"]
    }
  },
  "skills": [
    {
      "id": "ctrldot-guard",
      "path": "/absolute/path/to/adapters/openclaw/skills/ctrldot-guard"
    }
  ]
}
```

### 4. Copy Skill

Copy the skill to your OpenClaw skills folder:

```bash
cp -r adapters/openclaw/skills/ctrldot-guard /path/to/openclaw/skills/
```

### 5. Restart OpenClaw

Restart OpenClaw to load the plugin and skill.

## Tools

### `ctrldot_exec`

Guarded shell command execution. Replaces built-in `exec`.

**Parameters:**
- `command` (string, required): Shell command to execute
- `cwd` (string, optional): Working directory
- `timeoutMs` (number, optional): Timeout in milliseconds (default: 60000)

**Example:**
```javascript
ctrldot_exec({command: "ls -la", cwd: "/tmp"})
```

### `ctrldot_web_fetch`

Guarded HTTP GET requests. Replaces built-in `web_fetch`.

**Parameters:**
- `url` (string, required): URL to fetch
- `maxBytes` (number, optional): Maximum bytes to fetch (default: 2097152 = 2MB)

**Example:**
```javascript
ctrldot_web_fetch({url: "https://example.com"})
```

### `ctrldot_propose`

Diagnostic tool to check if an action would be allowed without executing it.

**Parameters:**
- `kind` (string, required): Action kind (e.g., "tool.call.exec")
- `name` (string, required): Tool/action name
- `args` (object, required): Action arguments
- `meta` (object, optional): Optional metadata

**Example:**
```javascript
ctrldot_propose({
  kind: "tool.call.exec",
  name: "exec",
  args: {command: "rm -rf /tmp/test"}
})
```

## Testing

### Smoke Test 1: List Directory

```javascript
ctrldot_exec({command: "ls -la /tmp"})
```

### Smoke Test 2: Fetch URL

```javascript
ctrldot_web_fetch({url: "https://httpbin.org/json"})
```

## Tool list and profile

The plugin registers `ctrldot_exec`, `ctrldot_web_fetch`, and `ctrldot_propose` via OpenClaw’s `api.registerTool()` (optional tools). For the agent to see them:

- Use **`tools.profile: "full"`** and **`tools.deny: ["exec", "web_fetch"]`** so all tools (including the plugin’s) are available except the built-in exec/web_fetch, or  
- Use **`tools.profile: "coding"`**, **`tools.deny: ["group:runtime", "group:web"]`**, and **`tools.allow`** (or **`agents.list[].tools.allow`**) that includes **`"ctrldot"`** and/or **`"ctrldot_exec"`**, **`"ctrldot_web_fetch"`**, **`"ctrldot_propose"`**.

Load the plugin from the **built output directory** so the manifest is found, e.g. `plugins.load.paths: ["/path/to/adapters/openclaw/plugin/dist"]`.

## How It Works

1. **Plugin Load**: Plugin registers agent with Ctrl Dot on load
2. **Tool Call**: When agent calls `ctrldot_exec` or `ctrldot_web_fetch`:
   - Tool builds action proposal
   - Proposes to Ctrl Dot API
   - If DENY/STOP: throws error
   - If ALLOW/WARN/THROTTLE: executes action
3. **Built-in Tools**: Built-in `exec` and `web_fetch` are denied via toolGroups config
4. **Skill**: Skill instructs agents to use guarded tools

## Decision Handling

- **ALLOW**: Proceeds normally
- **WARN**: Logs warning, proceeds
- **THROTTLE**: Proceeds with degraded constraints
- **DENY/STOP**: Throws error, action not executed

## Troubleshooting

**Plugin not loading:**
- Check plugin path is absolute
- Verify `dist/index.js` exists after build
- Check OpenClaw logs for errors

**Ctrl Dot connection errors:**
- Ensure daemon is running on configured URL
- Check `ctrldotUrl` in config
- Verify network connectivity

**Tool calls denied:**
- Check Ctrl Dot rules configuration
- Verify agent is registered
- Check Ctrl Dot logs: `ctrldot events tail`

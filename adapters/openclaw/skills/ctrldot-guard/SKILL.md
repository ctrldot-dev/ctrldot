# Ctrl Dot Guard Skill

This skill instructs agents to use Ctrl Dot guarded tools instead of risky built-in tools.

## Rules

1. **Never use `exec`** - Always use `ctrldot_exec` instead
   - The built-in `exec` tool is disabled for safety
   - Use `ctrldot_exec` for all shell command execution
   - Example: `ctrldot_exec({command: "ls -la", cwd: "/tmp"})`

2. **Never use `web_fetch`** - Always use `ctrldot_web_fetch` instead
   - The built-in `web_fetch` tool is disabled for safety
   - Use `ctrldot_web_fetch` for all HTTP requests
   - Example: `ctrldot_web_fetch({url: "https://example.com"})`

3. **Use `ctrldot_propose` when unsure**
   - If you're not sure if an action will be allowed, use `ctrldot_propose` first
   - This will return a decision without executing the action
   - Example: `ctrldot_propose({kind: "tool.call.exec", name: "exec", args: {command: "rm -rf /"}})`

## Why?

Ctrl Dot provides:
- Budget tracking and limits
- Rule enforcement (filesystem allow-lists, network domain allow-lists)
- Loop detection
- Human resolution for high-risk actions

All tool executions are logged to the Ctrl Dot ledger for auditability.

## Decision Meanings

- **ALLOW**: Action is permitted, proceed
- **WARN**: Action is permitted but with a warning
- **THROTTLE**: Action is permitted but with degraded constraints
- **DENY**: Action is denied, do not proceed
- **STOP**: Action is denied and agent should halt

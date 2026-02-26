# MCP (Model Context Protocol) Integration

Ctrl Dot can be used as an MCP server, allowing AI tools (like Claude Desktop, Cursor, etc.) to interact with it.

## MCP Servers

### 1. Ctrl Dot decisions (`ctrldot-mcp`)

The `ctrldot-mcp` binary provides an MCP-compatible interface to the Ctrl Dot daemon:

```bash
# Propose an action
./bin/ctrldot-mcp propose my-agent git.push '{"repo_path":"/tmp/repo"}' 1.2

# Output: {"decision":"DENY","reason":"Requires resolution","allowed":false}
```

### 2. Docs server (`ctrldot-docs`)

The `mcp-docs` project provides a **read‑only** MCP server over the local `docs/` folder. It exposes three tools:

- **`List Docs`** — list all available documentation pages (id, title, path)
- **`Fetch Doc`** — fetch a single page’s full Markdown and (optionally) its headings + anchors
- **`Search Docs`** — search across docs and return ranked matches with snippets and nearest headings

From the repo root:

```bash
cd mcp-docs
npx tsx src/index.ts
```

By default it reads from `../docs`. To point it elsewhere:

```bash
CTRLDOT_DOCS_ROOT=/absolute/path/to/docs npx tsx src/index.ts
```

You can also enable simple “hot reload” on every call:

```bash
CTRLDOT_DOCS_RELOAD=1 npx tsx src/index.ts
```

Typical MCP client configuration (Cursor / Claude) looks like:

```json
{
  "mcpServers": {
    "ctrldot-docs": {
      "command": "npx",
      "args": ["tsx", "src/index.ts"],
      "cwd": "/ABS/PATH/TO/CtrlDot/mcp-docs",
      "env": {
        "CTRLDOT_DOCS_ROOT": "/ABS/PATH/TO/CtrlDot/docs",
        "CTRLDOT_DOCS_RELOAD": "1"
      }
    }
  }
}
```

## MCP Tools (ctrldot-mcp)

Ctrl Dot exposes these MCP tools:

### 1. `ctrldot_propose_action`
Propose an action and get a decision.

**Input:**
- `agent_id` (string): Agent identifier
- `action_type` (string): Type of action (e.g., "git.push")
- `target` (object): Action target parameters
- `cost_gbp` (number): Estimated cost in GBP

**Output:**
- `decision` (string): ALLOW, WARN, THROTTLE, DENY, or STOP
- `reason` (string): Reason for decision
- `allowed` (boolean): Whether action is allowed

### 2. `ctrldot_register_agent`
Register a new agent.

**Input:**
- `agent_id` (string)
- `display_name` (string)

**Output:**
- Agent object

### 3. `ctrldot_get_events`
Get events feed.

**Input:**
- `agent_id` (string, optional)
- `limit` (number, optional)

**Output:**
- Array of events

## Integration with MCP Clients

### Claude Desktop

Add to `~/Library/Application Support/Claude/claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "ctrldot": {
      "command": "/path/to/ctrldot-mcp",
      "args": []
    }
  }
}
```

### Cursor / Other MCP Clients

Configure MCP server pointing to `ctrldot-mcp` binary.

## Benefits

- **Standard Protocol**: Works with any MCP-compatible tool
- **No Custom Integration**: Use existing MCP infrastructure
- **Tool Discovery**: MCP clients can discover Ctrl Dot capabilities
- **Unified Interface**: Same interface across different AI tools

## Future Enhancements

- Full MCP server implementation (not just CLI wrapper)
- MCP resource providers (agents, events as resources)
- MCP prompts (pre-built prompts for resolution requests)
- Streaming support for events feed

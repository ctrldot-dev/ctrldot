# Ctrl Dot Integration Guide

## Overview

Ctrl Dot acts as a **gatekeeper** between your AI agents and their actions. Before executing any action, agents should:

1. **Propose** the action to Ctrl Dot
2. **Check** the decision (ALLOW/WARN/THROTTLE/DENY/STOP)
3. **Execute** only if allowed, or request resolution if denied

## Quick Integration Pattern

### Before Every Action:

```python
# Python example
decision = ctrldot.propose_action(
    agent_id="my-agent",
    action_type="git.push",
    target={"repo_path": "/path/to/repo", "branch": "main"},
    cost_gbp=1.2
)

if decision["decision"] in ["ALLOW", "WARN", "THROTTLE"]:
    execute_git_push()  # Proceed with action
else:
    handle_denial(decision)  # Stop or request resolution
```

## Integration Steps

### 1. Register Your Agent

**On agent startup:**

```bash
curl -X POST http://127.0.0.1:7777/v1/agents/register \
  -H 'Content-Type: application/json' \
  -d '{
    "agent_id": "my-openclaw-agent",
    "display_name": "OpenClaw Agent"
  }'
```

### 2. Start a Session (Optional)

```bash
curl -X POST http://127.0.0.1:7777/v1/sessions/start \
  -H 'Content-Type: application/json' \
  -d '{
    "agent_id": "my-openclaw-agent",
    "metadata": {"framework": "openclaw"}
  }'
```

### 3. Before Each Tool Call

**Propose action to Ctrl Dot:**

```bash
curl -X POST http://127.0.0.1:7777/v1/actions/propose \
  -H 'Content-Type: application/json' \
  -d '{
    "agent_id": "my-openclaw-agent",
    "session_id": "sess_123",
    "intent": {
      "title": "Fix bug from issue #42"
    },
    "action": {
      "type": "git.push",
      "target": {
        "repo_path": "/Users/gareth/dev/myrepo",
        "branch": "main"
      },
      "inputs": {
        "commit_message": "Fix null pointer"
      }
    },
    "cost": {
      "currency": "GBP",
      "estimated_gbp": 1.20,
      "estimated_tokens": 12000,
      "model": "claude-opus-4.6"
    },
    "context": {
      "tool": "git",
      "tags": ["code", "deploy"]
    }
  }'
```

**Response:**
```json
{
  "decision": "DENY",
  "reason": "Requires resolution for git.push",
  "ledger_event_id": "evt_..."
}
```

### 4. Handle Decisions

- **ALLOW**: Execute the action
- **WARN**: Execute but log warning (budget threshold reached)
- **THROTTLE**: Execute with degraded constraints (cheap model, fewer parallel tasks)
- **DENY**: Stop - action requires resolution or violates rules
- **STOP**: Stop immediately - agent halted or loop detected

### 5. Request Resolution (for DENY actions)

If an action is denied due to requiring resolution:

```bash
# Generate resolution token via CLI
ctrldot resolve allow-once \
  --agent my-openclaw-agent \
  --action git.push \
  --ttl 10m

# Output: res:abc123...xyz
```

Then include token in proposal:
```json
{
  "agent_id": "my-openclaw-agent",
  "action": {...},
  "resolution_token": "res:abc123...xyz"
}
```

## OpenClaw Integration Example

### Middleware Pattern

```python
class CtrlDotMiddleware:
    def __init__(self, agent_id, ctrldot_url):
        self.agent_id = agent_id
        self.ctrldot_url = ctrldot_url
        self.session_id = None
    
    def before_tool_call(self, tool_name, tool_args, estimated_cost):
        """Called before every tool execution"""
        proposal = {
            "agent_id": self.agent_id,
            "session_id": self.session_id,
            "action": {
                "type": tool_name,
                "target": tool_args.get("target", {}),
                "inputs": tool_args.get("inputs", {})
            },
            "cost": {
                "estimated_gbp": estimated_cost,
                "estimated_tokens": int(estimated_cost * 10000),
                "model": "claude-opus"
            },
            "context": {"tool": tool_name}
        }
        
        response = requests.post(
            f"{self.ctrldot_url}/v1/actions/propose",
            json=proposal
        )
        decision = response.json()
        
        if decision["decision"] in ["DENY", "STOP"]:
            raise ActionDeniedError(decision["reason"])
        
        return decision
```

### Usage in OpenClaw

```python
# Initialize middleware
ctrl_dot = CtrlDotMiddleware(
    agent_id="openclaw-agent",
    ctrldot_url="http://127.0.0.1:7777"
)

# Wrap tool calls
def safe_tool_call(tool_name, **kwargs):
    decision = ctrl_dot.before_tool_call(
        tool_name=tool_name,
        tool_args=kwargs,
        estimated_cost=estimate_cost(tool_name, kwargs)
    )
    
    # Apply throttle constraints if needed
    if decision.get("throttle"):
        apply_throttle(decision["throttle"])
    
    # Execute tool
    return execute_tool(tool_name, **kwargs)
```

## Testing Without Real Agents

Use the provided test agent:

```bash
# Go test agent
go run examples/test-agent.go

# Python test agent (requires requests)
python3 examples/python-agent.py
```

These demonstrate:
- Agent registration
- Session management
- Action proposals
- Decision handling

## Monitoring

### Watch Events

```bash
# Tail events
ctrldot events tail --n 20

# Filter by agent
ctrldot events tail --agent my-agent --n 10
```

### Check Budget

```bash
# View agent budget (when implemented)
ctrldot budget my-agent
```

### View in UI

Open `http://127.0.0.1:7777/ui` to see:
- All agents
- Event feed
- Rules
- Status

## Common Scenarios

### Scenario 1: Safe Action (ALLOW)

```python
decision = propose_action("filesystem.read", {"path": "~/dev/file.txt"})
# Returns: {"decision": "ALLOW"}
# → Execute action
```

### Scenario 2: Requires Resolution (DENY)

```python
decision = propose_action("git.push", {...})
# Returns: {"decision": "DENY", "reason": "Requires resolution"}
# → Request resolution token via CLI
# → Retry with resolution_token
```

### Scenario 3: Budget Warning (WARN)

```python
decision = propose_action("expensive.action", {...})
# Returns: {"decision": "WARN", "warnings": [{"code": "BUDGET_70", ...}]}
# → Execute but log warning
```

### Scenario 4: Budget Throttle (THROTTLE)

```python
decision = propose_action("expensive.action", {...})
# Returns: {"decision": "THROTTLE", "throttle": {"model_policy": "cheap", ...}}
# → Execute with degraded constraints
```

### Scenario 5: Agent Halted (STOP)

```python
decision = propose_action("any.action", {...})
# Returns: {"decision": "STOP", "reason": "Agent is halted"}
# → Stop all operations, wait for resume
```

## Next Steps

1. **Build OpenClaw integration** - Create middleware/hook for OpenClaw
2. **Create client libraries** - Python, Node.js, Go SDKs
3. **Add webhooks** - Real-time notifications for decisions
4. **Mobile/watch apps** - Quick resolution token generation

## Examples

See `examples/` directory:
- `test-agent.go` - Go integration example
- `python-agent.py` - Python integration example
- `README.md` - More examples and patterns

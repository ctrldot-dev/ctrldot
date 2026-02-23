# Ctrl Dot Integration Examples

## Test Agent

A simple Go program that demonstrates how to integrate with Ctrl Dot.

### Run the test agent:

```bash
go run examples/test-agent.go
```

This will:
1. Register an agent with Ctrl Dot
2. Start a session
3. Propose various actions (safe, git.push, filesystem.delete)
4. Show how Ctrl Dot responds with decisions

## Integration Pattern

For real agents (like OpenClaw), integrate Ctrl Dot like this:

```go
// Before executing any action:
func (agent *YourAgent) ExecuteAction(action Action) error {
    // 1. Propose action to Ctrl Dot
    decision := agent.ProposeToCtrlDot(action)
    
    // 2. Check decision
    switch decision {
    case "ALLOW", "WARN", "THROTTLE":
        // Execute the action
        return agent.execute(action)
    case "DENY", "STOP":
        // Stop or request resolution
        return fmt.Errorf("action denied by Ctrl Dot")
    }
}
```

## API Integration

### Python Example

```python
import requests

def propose_action(agent_id, action_type, target, cost_gbp):
    response = requests.post(
        "http://127.0.0.1:7777/v1/actions/propose",
        json={
            "agent_id": agent_id,
            "intent": {"title": "Execute action"},
            "action": {
                "type": action_type,
                "target": target,
                "inputs": {}
            },
            "cost": {
                "currency": "GBP",
                "estimated_gbp": cost_gbp,
                "estimated_tokens": int(cost_gbp * 10000),
                "model": "your-model"
            },
            "context": {"tool": action_type}
        }
    )
    return response.json()["decision"]
```

## OpenClaw Integration

To integrate with OpenClaw or similar frameworks:

1. **Register your agent** on startup:
   ```bash
   curl -X POST http://127.0.0.1:7777/v1/agents/register \
     -H 'Content-Type: application/json' \
     -d '{"agent_id":"openclaw-agent","display_name":"OpenClaw"}'
   ```

2. **Before each tool call**, propose to Ctrl Dot:
   - Call `/v1/actions/propose` with action details
   - Check the `decision` field
   - If ALLOW/WARN/THROTTLE: proceed with execution
   - If DENY/STOP: stop and optionally request resolution token

3. **Handle resolution tokens**:
   - For actions requiring resolution, use CLI to generate token:
     ```bash
     ctrldot resolve allow-once --agent openclaw-agent --action git.push --ttl 10m
     ```
   - Include token in proposal: `"resolution_token": "..."`

4. **Monitor events**:
   - Watch `/v1/events` for decision events
   - Track budget usage via events
   - React to halt events

## Next Steps

- Build OpenClaw integration layer
- Create Python/Node.js client libraries
- Add webhook support for notifications
- Build mobile/watch apps for resolution tokens

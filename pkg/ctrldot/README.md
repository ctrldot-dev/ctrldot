# Ctrl Dot SDK

Go client library for integrating with Ctrl Dot.

## Usage

```go
import "github.com/futurematic/kernel/pkg/ctrldot"

// Create client
client := ctrldot.NewClient("http://127.0.0.1:7777")

// Register agent
agent, err := client.RegisterAgent(ctx, "my-agent", "My Agent", "")
if err != nil {
    log.Fatal(err)
}

// Start session
session, err := client.StartSession(ctx, agent.AgentID, map[string]interface{}{
    "framework": "my-framework",
})

// Propose action
proposal := domain.ActionProposal{
    AgentID: agent.AgentID,
    SessionID: session.SessionID,
    Action: domain.Action{
        Type: "git.push",
        Target: map[string]interface{}{"repo_path": "/tmp/repo"},
        Inputs: map[string]interface{}{},
    },
    Cost: domain.CostEstimate{
        EstimatedGBP: 1.2,
        EstimatedTokens: 12000,
        Model: "claude-opus",
    },
    Context: domain.ActionContext{Tool: "git"},
}

decision, err := client.ProposeAction(ctx, proposal)
if err != nil {
    log.Fatal(err)
}

if decision.Decision == "ALLOW" || decision.Decision == "WARN" {
    // Execute action
} else {
    // Handle denial
}
```

## Methods

- `RegisterAgent(ctx, agentID, displayName, defaultMode)`
- `StartSession(ctx, agentID, metadata)`
- `ProposeAction(ctx, proposal)`
- `GetEvents(ctx, agentID, sinceTS, limit)`
- `ListAgents(ctx)`
- `GetAgent(ctx, agentID)`
- `HaltAgent(ctx, agentID, reason)`
- `ResumeAgent(ctx, agentID)`

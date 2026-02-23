package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/futurematic/kernel/internal/domain"
	"github.com/futurematic/kernel/pkg/ctrldot"
)

// MCP server implementation for Ctrl Dot
// MCP (Model Context Protocol) allows AI tools to interact with Ctrl Dot

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: ctrldot-mcp <command> [args...]")
		fmt.Println("\nCommands:")
		fmt.Println("  capabilities          — GET /v1/capabilities (agent discovery)")
		fmt.Println("  propose <agent_id> <action_type> <target_json> <cost_gbp>")
		fmt.Println("  register <agent_id> <display_name>")
		fmt.Println("  events <agent_id> [limit]")
		fmt.Println("  agents")
		os.Exit(1)
	}

	baseURL := os.Getenv("CTRLDOT_URL")
	if baseURL == "" {
		baseURL = "http://127.0.0.1:7777"
	}
	client := ctrldot.NewClient(baseURL)
	ctx := context.Background()

	command := os.Args[1]

	switch command {
	case "capabilities", "ctrldot_capabilities":
		caps, err := client.GetCapabilities(ctx)
		if err != nil {
			log.Fatal(err)
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(caps); err != nil {
			log.Fatal(err)
		}
		return
	}

	switch command {
	case "propose":
		if len(os.Args) < 6 {
			log.Fatal("Usage: propose <agent_id> <action_type> <target_json> <cost_gbp>")
		}
		agentID := os.Args[2]
		actionType := os.Args[3]
		targetJSON := os.Args[4]
		var costGBP float64
		fmt.Sscanf(os.Args[5], "%f", &costGBP)

		var target map[string]interface{}
		if err := json.Unmarshal([]byte(targetJSON), &target); err != nil {
			log.Fatalf("Invalid target JSON: %v", err)
		}

		proposal := domain.ActionProposal{
			AgentID: agentID,
			Intent: domain.ActionIntent{
				Title: fmt.Sprintf("MCP action: %s", actionType),
			},
			Action: domain.Action{
				Type:   actionType,
				Target: target,
				Inputs: map[string]interface{}{},
			},
			Cost: domain.CostEstimate{
				Currency:       "GBP",
				EstimatedGBP:   costGBP,
				EstimatedTokens: int64(costGBP * 10000),
				Model:          "mcp-client",
			},
			Context: domain.ActionContext{
				Tool: actionType,
			},
		}

		decision, err := client.ProposeAction(ctx, proposal)
		if err != nil {
			log.Fatal(err)
		}

		// Output in MCP-friendly format (JSON)
		output := map[string]interface{}{
			"decision": decision.Decision,
			"reason":   decision.Reason,
			"allowed":  decision.Decision == "ALLOW" || decision.Decision == "WARN" || decision.Decision == "THROTTLE",
		}
		json.NewEncoder(os.Stdout).Encode(output)

	case "register":
		if len(os.Args) < 4 {
			log.Fatal("Usage: register <agent_id> <display_name>")
		}
		agent, err := client.RegisterAgent(ctx, os.Args[2], os.Args[3], "")
		if err != nil {
			log.Fatal(err)
		}
		json.NewEncoder(os.Stdout).Encode(agent)

	case "events":
		agentID := ""
		limit := 10
		if len(os.Args) >= 3 {
			agentID = os.Args[2]
		}
		if len(os.Args) >= 4 {
			fmt.Sscanf(os.Args[3], "%d", &limit)
		}

		var agentIDPtr *string
		if agentID != "" {
			agentIDPtr = &agentID
		}

		events, err := client.GetEvents(ctx, agentIDPtr, nil, limit)
		if err != nil {
			log.Fatal(err)
		}
		json.NewEncoder(os.Stdout).Encode(events)

	case "agents":
		agents, err := client.ListAgents(ctx)
		if err != nil {
			log.Fatal(err)
		}
		json.NewEncoder(os.Stdout).Encode(agents)

	default:
		log.Fatalf("Unknown command: %s", command)
	}
}

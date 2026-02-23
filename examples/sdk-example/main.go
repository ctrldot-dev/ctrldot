package main

import (
	"context"
	"fmt"
	"log"

	"github.com/futurematic/kernel/internal/domain"
	"github.com/futurematic/kernel/pkg/ctrldot"
)

func main() {
	fmt.Println("📦 SDK Example - Ctrl Dot Client")
	fmt.Println("===============================")
	fmt.Println()

	// Create client
	client := ctrldot.NewClient("http://127.0.0.1:7777")

	ctx := context.Background()

	// Register agent
	agent, err := client.RegisterAgent(ctx, "sdk-agent", "SDK Test Agent", "")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("✅ Registered: %s\n", agent.AgentID)

	// Start session
	session, err := client.StartSession(ctx, agent.AgentID, map[string]interface{}{
		"framework": "sdk-example",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("✅ Session: %s\n", session.SessionID)

	// Propose action
	proposal := domain.ActionProposal{
		AgentID:   agent.AgentID,
		SessionID: session.SessionID,
		Intent: domain.ActionIntent{
			Title: "Test action via SDK",
		},
		Action: domain.Action{
			Type: "git.push",
			Target: map[string]interface{}{
				"repo_path": "/tmp/repo",
				"branch":    "main",
			},
			Inputs: map[string]interface{}{
				"commit_message": "Test commit",
			},
		},
		Cost: domain.CostEstimate{
			Currency:        "GBP",
			EstimatedGBP:    1.2,
			EstimatedTokens: 12000,
			Model:           "test-model",
		},
		Context: domain.ActionContext{
			Tool: "git",
			Tags: []string{"test"},
		},
	}

	decision, err := client.ProposeAction(ctx, proposal)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("\n📋 Decision: %s\n", decision.Decision)
	if decision.Reason != "" {
		fmt.Printf("   Reason: %s\n", decision.Reason)
	}

	// List agents
	agents, err := client.ListAgents(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("\n📊 Total agents: %d\n", len(agents))

	// Get events
	events, err := client.GetEvents(ctx, &agent.AgentID, nil, 5)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("📜 Events for agent: %d\n", len(events))

	fmt.Println("\n✅ SDK example complete!")
}

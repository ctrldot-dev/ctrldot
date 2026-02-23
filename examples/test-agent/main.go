package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// TestAgent is a simple agent that demonstrates Ctrl Dot integration
type TestAgent struct {
	AgentID    string
	ServerURL  string
	SessionID  string
	HTTPClient *http.Client
}

// NewTestAgent creates a new test agent
func NewTestAgent(agentID string, serverURL string) *TestAgent {
	return &TestAgent{
		AgentID:    agentID,
		ServerURL:  serverURL,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// Register registers the agent with Ctrl Dot
func (a *TestAgent) Register(displayName string) error {
	reqBody := map[string]string{
		"agent_id":     a.AgentID,
		"display_name": displayName,
	}
	jsonBody, _ := json.Marshal(reqBody)

	resp, err := a.HTTPClient.Post(a.ServerURL+"/v1/agents/register", "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to register: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("registration failed: %s", string(body))
	}

	fmt.Printf("✅ Agent '%s' registered\n", a.AgentID)
	return nil
}

// StartSession starts a session
func (a *TestAgent) StartSession() error {
	reqBody := map[string]interface{}{
		"agent_id": a.AgentID,
		"metadata": map[string]string{
			"started_by": "test-agent",
		},
	}
	jsonBody, _ := json.Marshal(reqBody)

	resp, err := a.HTTPClient.Post(a.ServerURL+"/v1/sessions/start", "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to start session: %w", err)
	}
	defer resp.Body.Close()

	var session map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&session); err != nil {
		return err
	}

	a.SessionID = session["session_id"].(string)
	fmt.Printf("✅ Session started: %s\n", a.SessionID)
	return nil
}

// ProposeAction proposes an action and gets a decision
func (a *TestAgent) ProposeAction(actionType string, target map[string]interface{}, inputs map[string]interface{}, costGBP float64) (string, error) {
	proposal := map[string]interface{}{
		"agent_id": a.AgentID,
		"session_id": a.SessionID,
		"intent": map[string]string{
			"title": fmt.Sprintf("Test action: %s", actionType),
		},
		"action": map[string]interface{}{
			"type":   actionType,
			"target": target,
			"inputs": inputs,
		},
		"cost": map[string]interface{}{
			"currency":        "GBP",
			"estimated_gbp":   costGBP,
			"estimated_tokens": int64(costGBP * 10000), // Rough estimate
			"model":           "test-model",
		},
		"context": map[string]interface{}{
			"tool": actionType,
			"tags": []string{"test"},
		},
	}

	jsonBody, _ := json.Marshal(proposal)
	req, _ := http.NewRequest("POST", a.ServerURL+"/v1/actions/propose", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to propose action: %w", err)
	}
	defer resp.Body.Close()

	var decision map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&decision); err != nil {
		return "", err
	}

	decisionStr := decision["decision"].(string)
	reason, _ := decision["reason"].(string)

	fmt.Printf("📋 Action: %s\n", actionType)
	fmt.Printf("   Decision: %s\n", decisionStr)
	if reason != "" {
		fmt.Printf("   Reason: %s\n", reason)
	}

	return decisionStr, nil
}

// ExecuteAction executes an action if allowed
func (a *TestAgent) ExecuteAction(actionType string, target map[string]interface{}, inputs map[string]interface{}, costGBP float64) error {
	decision, err := a.ProposeAction(actionType, target, inputs, costGBP)
	if err != nil {
		return err
	}

	switch decision {
	case "ALLOW", "WARN", "THROTTLE":
		fmt.Printf("✅ Executing action: %s\n", actionType)
		// In a real agent, you would execute the action here
		// For testing, we just simulate
		time.Sleep(100 * time.Millisecond)
		fmt.Printf("✅ Action completed\n")
		return nil
	case "DENY", "STOP":
		return fmt.Errorf("action denied: %s", decision)
	default:
		return fmt.Errorf("unknown decision: %s", decision)
	}
}

func main() {
	fmt.Println("🤖 Test Agent - Ctrl Dot Integration Demo")
	fmt.Println("=======================================")
	fmt.Println()

	// Create agent
	agent := NewTestAgent("test-bot", "http://127.0.0.1:7777")

	// Register agent
	if err := agent.Register("Test Bot"); err != nil {
		fmt.Printf("❌ Registration failed: %v\n", err)
		return
	}

	// Start session
	if err := agent.StartSession(); err != nil {
		fmt.Printf("❌ Session start failed: %v\n", err)
		return
	}

	fmt.Println()
	fmt.Println("--- Testing Actions ---")
	fmt.Println()

	// Test 1: Safe action (should be ALLOWED)
	fmt.Println("Test 1: Safe read action")
	err := agent.ExecuteAction("filesystem.read", map[string]interface{}{
		"path": "~/dev/test.txt",
	}, map[string]interface{}{}, 0.1)
	if err != nil {
		fmt.Printf("❌ %v\n", err)
	}

	fmt.Println()

	// Test 2: Git push (should require resolution)
	fmt.Println("Test 2: Git push action")
	decision, err := agent.ProposeAction("git.push", map[string]interface{}{
		"repo_path": "/Users/gareth/dev/myrepo",
		"branch":    "main",
	}, map[string]interface{}{
		"commit_message": "Test commit",
	}, 1.2)
	if err != nil {
		fmt.Printf("❌ %v\n", err)
	} else {
		if decision == "DENY" {
			fmt.Println("✅ Correctly denied (requires resolution)")
		}
	}

	fmt.Println()

	// Test 3: Filesystem delete (should require resolution)
	fmt.Println("Test 3: Filesystem delete action")
	decision, err = agent.ProposeAction("filesystem.delete", map[string]interface{}{
		"path": "/tmp/test.txt",
	}, map[string]interface{}{}, 0.5)
	if err != nil {
		fmt.Printf("❌ %v\n", err)
	} else {
		if decision == "DENY" {
			fmt.Println("✅ Correctly denied (requires resolution)")
		}
	}

	fmt.Println("\n✅ Test agent demo complete!")
	fmt.Println("\nTo use with real agents, integrate Ctrl Dot's /v1/actions/propose endpoint")
	fmt.Println("before executing any actions.")
}

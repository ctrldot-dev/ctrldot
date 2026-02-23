package ctrldot

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/futurematic/kernel/internal/domain"
)

// Client is a Ctrl Dot API client
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

// NewClient creates a new Ctrl Dot client
func NewClient(baseURL string) *Client {
	if baseURL == "" {
		baseURL = "http://127.0.0.1:7777"
	}
	return &Client{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// RegisterAgent registers an agent
func (c *Client) RegisterAgent(ctx context.Context, agentID string, displayName string, defaultMode string) (*domain.Agent, error) {
	reqBody := map[string]string{
		"agent_id":     agentID,
		"display_name": displayName,
	}
	if defaultMode != "" {
		reqBody["default_mode"] = defaultMode
	}

	var agent domain.Agent
	if err := c.postJSON(ctx, "/v1/agents/register", reqBody, &agent); err != nil {
		return nil, err
	}
	return &agent, nil
}

// StartSession starts a session
func (c *Client) StartSession(ctx context.Context, agentID string, metadata map[string]interface{}) (*domain.Session, error) {
	reqBody := map[string]interface{}{
		"agent_id": agentID,
		"metadata": metadata,
	}

	var session domain.Session
	if err := c.postJSON(ctx, "/v1/sessions/start", reqBody, &session); err != nil {
		return nil, err
	}
	return &session, nil
}

// ProposeAction proposes an action and returns a decision
func (c *Client) ProposeAction(ctx context.Context, proposal domain.ActionProposal) (*domain.DecisionResponse, error) {
	var decision domain.DecisionResponse
	if err := c.postJSON(ctx, "/v1/actions/propose", proposal, &decision); err != nil {
		return nil, err
	}
	return &decision, nil
}

// GetEvents retrieves events
func (c *Client) GetEvents(ctx context.Context, agentID *string, sinceTS *int64, limit int) ([]domain.Event, error) {
	url := fmt.Sprintf("%s/v1/events?limit=%d", c.BaseURL, limit)
	if agentID != nil {
		url += "&agent_id=" + *agentID
	}
	if sinceTS != nil {
		url += fmt.Sprintf("&since_ts=%d", *sinceTS)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	var events []domain.Event
	if err := json.NewDecoder(resp.Body).Decode(&events); err != nil {
		return nil, err
	}
	return events, nil
}

// ListAgents lists all agents
func (c *Client) ListAgents(ctx context.Context) ([]domain.Agent, error) {
	var agents []domain.Agent
	if err := c.getJSON(ctx, "/v1/agents", &agents); err != nil {
		return nil, err
	}
	return agents, nil
}

// GetAgent retrieves an agent
func (c *Client) GetAgent(ctx context.Context, agentID string) (*domain.Agent, error) {
	var agent domain.Agent
	if err := c.getJSON(ctx, "/v1/agents/"+agentID, &agent); err != nil {
		return nil, err
	}
	return &agent, nil
}

// GetCapabilities returns agent-discovery capabilities (GET /v1/capabilities).
func (c *Client) GetCapabilities(ctx context.Context) (*domain.CapabilitiesResponse, error) {
	var out domain.CapabilitiesResponse
	if err := c.getJSON(ctx, "/v1/capabilities", &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// HaltAgent halts an agent
func (c *Client) HaltAgent(ctx context.Context, agentID string, reason string) error {
	reqBody := map[string]string{"reason": reason}
	return c.postJSON(ctx, "/v1/agents/"+agentID+"/halt", reqBody, nil)
}

// ResumeAgent resumes an agent
func (c *Client) ResumeAgent(ctx context.Context, agentID string) error {
	return c.postJSON(ctx, "/v1/agents/"+agentID+"/resume", map[string]interface{}{}, nil)
}

// Helper methods

func (c *Client) postJSON(ctx context.Context, path string, body interface{}, result interface{}) error {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.BaseURL+path, bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) getJSON(ctx context.Context, path string, result interface{}) error {
	req, err := http.NewRequestWithContext(ctx, "GET", c.BaseURL+path, nil)
	if err != nil {
		return err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	return json.NewDecoder(resp.Body).Decode(result)
}

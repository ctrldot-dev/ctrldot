package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Client represents an HTTP client for the kernel API
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new kernel API client
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: strings.TrimSuffix(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Healthz checks the health of the kernel server
func (c *Client) Healthz() (*HealthResponse, error) {
	var resp HealthResponse
	if err := c.get("/v1/healthz", nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Plan creates a plan from intents
func (c *Client) Plan(req PlanRequest) (*PlanResponse, error) {
	var resp PlanResponse
	if err := c.post("/v1/plan", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Apply applies a plan
func (c *Client) Apply(req ApplyRequest) (*ApplyResponse, error) {
	var resp ApplyResponse
	if err := c.post("/v1/apply", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Expand expands nodes with their relationships
func (c *Client) Expand(req ExpandRequest) (*ExpandResponse, error) {
	params := url.Values{}
	params.Add("ids", strings.Join(req.IDs, ","))
	if req.NamespaceID != nil {
		params.Add("namespace_id", *req.NamespaceID)
	}
	params.Add("depth", strconv.Itoa(req.Depth))
	if req.AsOfSeq > 0 {
		params.Add("asof_seq", strconv.FormatInt(req.AsOfSeq, 10))
	} else if req.AsOfTime != nil {
		params.Add("asof_time", req.AsOfTime.Format(time.RFC3339))
	}

	var resp ExpandResponse
	if err := c.get("/v1/expand", params, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// History gets operation history for a target
func (c *Client) History(req HistoryRequest) (HistoryResponse, error) {
	params := url.Values{}
	params.Add("target", req.Target)
	if req.Limit > 0 {
		params.Add("limit", strconv.Itoa(req.Limit))
	}

	var resp HistoryResponse
	if err := c.get("/v1/history", params, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// Diff gets differences between two sequence numbers
func (c *Client) Diff(req DiffRequest) (*DiffResponse, error) {
	params := url.Values{}
	params.Add("a_seq", strconv.FormatInt(req.ASeq, 10))
	params.Add("b_seq", strconv.FormatInt(req.BSeq, 10))
	if req.Target != "" {
		params.Add("target", req.Target)
	}

	var resp DiffResponse
	if err := c.get("/v1/diff", params, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Helper methods

func (c *Client) get(path string, params url.Values, result interface{}) error {
	fullURL := c.baseURL + path
	if params != nil && len(params) > 0 {
		fullURL += "?" + params.Encode()
	}

	httpResp, err := c.httpClient.Get(fullURL)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer httpResp.Body.Close()

	return c.handleResponse(httpResp, result)
}

func (c *Client) post(path string, body interface{}, result interface{}) error {
	jsonData, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	fullURL := c.baseURL + path
	httpResp, err := c.httpClient.Post(fullURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer httpResp.Body.Close()

	return c.handleResponse(httpResp, result)
}

func (c *Client) handleResponse(httpResp *http.Response, result interface{}) error {
	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	// Check for error response
	if httpResp.StatusCode >= 400 {
		var errResp ErrorResponse
		if err := json.Unmarshal(body, &errResp); err == nil {
			return &APIError{
				Code:    errResp.Error.Code,
				Message: errResp.Error.Message,
				Status:  httpResp.StatusCode,
			}
		}
		return fmt.Errorf("server error: %s (status %d)", string(body), httpResp.StatusCode)
	}

	// Parse successful response
	if err := json.Unmarshal(body, result); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	return nil
}

// APIError represents an API error
type APIError struct {
	Code    string
	Message string
	Status  int
}

func (e *APIError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// ExitCode returns the exit code for this error
func (e *APIError) ExitCode() int {
	switch e.Code {
	case "POLICY_DENIED":
		return 2
	case "CONFLICT":
		return 3
	case "VALIDATION", "NOT_FOUND":
		return 1
	default:
		return 4
	}
}

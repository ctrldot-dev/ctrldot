package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/futurematic/kernel/internal/domain"
)

func TestClientHealthz(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/healthz" {
			t.Errorf("Expected path /v1/healthz, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{"ok": true})
	}))
	defer server.Close()

	client := NewClient(server.URL)
	health, err := client.Healthz()
	if err != nil {
		t.Fatalf("Failed to get health: %v", err)
	}

	if !health.OK {
		t.Error("Expected health.OK to be true")
	}
}

func TestClientPlan(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/v1/plan" {
			t.Errorf("Expected path /v1/plan, got %s", r.URL.Path)
		}

		var req PlanRequest
		json.NewDecoder(r.Body).Decode(&req)

		response := PlanResponse{
			ID:   "plan:test",
			Hash: "sha256:abc123",
			Intents: req.Intents,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	req := PlanRequest{
		ActorID:     "user:test",
		Capabilities: []string{"read", "write"},
		Intents: []domain.Intent{
			{
				Kind: domain.IntentCreateNode,
				Payload: map[string]interface{}{
					"node_id": "node:test",
					"title":   "Test",
				},
			},
		},
	}

	plan, err := client.Plan(req)
	if err != nil {
		t.Fatalf("Failed to create plan: %v", err)
	}

	if plan.ID != "plan:test" {
		t.Errorf("Expected plan ID plan:test, got %s", plan.ID)
	}

	if plan.Hash != "sha256:abc123" {
		t.Errorf("Expected hash sha256:abc123, got %s", plan.Hash)
	}
}

func TestClientErrorHandling(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ErrorResponse{
			Error: struct {
				Code    string                 `json:"code"`
				Message string                 `json:"message"`
				Details map[string]interface{} `json:"details,omitempty"`
			}{
				Code:    "POLICY_DENIED",
				Message: "Policy violation",
			},
		})
	}))
	defer server.Close()

	client := NewClient(server.URL)
	req := PlanRequest{
		ActorID:     "user:test",
		Capabilities: []string{"read", "write"},
		Intents:     []domain.Intent{},
	}

	_, err := client.Plan(req)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("Expected APIError, got %T", err)
	}

	if apiErr.Code != "POLICY_DENIED" {
		t.Errorf("Expected code POLICY_DENIED, got %s", apiErr.Code)
	}

	if apiErr.ExitCode() != 2 {
		t.Errorf("Expected exit code 2, got %d", apiErr.ExitCode())
	}
}

func TestClientExpand(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/expand" {
			t.Errorf("Expected path /v1/expand, got %s", r.URL.Path)
		}

		ids := r.URL.Query().Get("ids")
		if ids != "node:test" {
			t.Errorf("Expected ids=node:test, got %s", ids)
		}

		response := ExpandResponse{
			Nodes: []domain.Node{
				{ID: "node:test", Title: "Test Node"},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	req := ExpandRequest{
		IDs:     []string{"node:test"},
		Depth:   1,
		AsOfSeq: 0,
	}

	result, err := client.Expand(req)
	if err != nil {
		t.Fatalf("Failed to expand: %v", err)
	}

	if len(result.Nodes) != 1 {
		t.Fatalf("Expected 1 node, got %d", len(result.Nodes))
	}

	if result.Nodes[0].ID != "node:test" {
		t.Errorf("Expected node ID node:test, got %s", result.Nodes[0].ID)
	}
}

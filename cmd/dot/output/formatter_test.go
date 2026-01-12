package output

import (
	"bytes"
	"testing"

	"github.com/futurematic/kernel/cmd/dot/client"
	"github.com/futurematic/kernel/internal/domain"
)

func TestTextFormatterPrintPlan(t *testing.T) {
	var buf bytes.Buffer
	formatter := NewFormatter("text", &buf)

	plan := &client.PlanResponse{
		ID:    "plan:test",
		Hash:  "sha256:abc123",
		Class: 1,
		PolicyReport: domain.PolicyReport{
			Denies: []domain.PolicyViolation{
				{RuleID: "rule1", Message: "Test deny"},
			},
		},
		Expanded: []domain.Change{
			{
				Kind: domain.ChangeCreateNode,
				Payload: map[string]interface{}{
					"node_id": "node:test",
					"title":   "Test Node",
				},
			},
		},
	}

	err := formatter.PrintPlan(plan)
	if err != nil {
		t.Fatalf("Failed to print plan: %v", err)
	}

	output := buf.String()
	if !contains(output, "PLAN plan:test") {
		t.Errorf("Expected PLAN output, got: %s", output)
	}
	if !contains(output, "DENIES") {
		t.Errorf("Expected DENIES section, got: %s", output)
	}
	if !contains(output, "CHANGES") {
		t.Errorf("Expected CHANGES section, got: %s", output)
	}
}

func TestJSONFormatterPrintPlan(t *testing.T) {
	var buf bytes.Buffer
	formatter := NewFormatter("json", &buf)

	plan := &client.PlanResponse{
		ID:   "plan:test",
		Hash: "sha256:abc123",
	}

	err := formatter.PrintPlan(plan)
	if err != nil {
		t.Fatalf("Failed to print plan: %v", err)
	}

	output := buf.String()
	if !contains(output, `"plan"`) {
		t.Errorf("Expected JSON with plan key, got: %s", output)
	}
}

func TestTextFormatterPrintOperation(t *testing.T) {
	var buf bytes.Buffer
	formatter := NewFormatter("text", &buf)

	op := &client.ApplyResponse{
		ID:   "op:test",
		Seq:  42,
		Changes: []domain.Change{
			{
				Kind: domain.ChangeCreateNode,
				Payload: map[string]interface{}{
					"node_id": "node:test",
					"title":   "Test Node",
				},
			},
		},
	}

	err := formatter.PrintOperation(op)
	if err != nil {
		t.Fatalf("Failed to print operation: %v", err)
	}

	output := buf.String()
	if !contains(output, "APPLIED op:test") {
		t.Errorf("Expected APPLIED output, got: %s", output)
	}
	if !contains(output, "seq=42") {
		t.Errorf("Expected seq=42, got: %s", output)
	}
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

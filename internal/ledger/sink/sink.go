package sink

import (
	"context"
	"strings"
	"time"

	"github.com/futurematic/kernel/internal/domain"
)

// RedactMap redacts sensitive keys in a map (values replaced with "[redacted]").
func RedactMap(m map[string]interface{}) map[string]interface{} {
	if m == nil {
		return nil
	}
	redactKeys := []string{"api_key", "token", "password", "secret", "key"}
	out := make(map[string]interface{}, len(m))
	for k, v := range m {
		if isSensitive(k, redactKeys) {
			out[k] = "[redacted]"
			continue
		}
		out[k] = redactValue(v, redactKeys)
	}
	return out
}

func isSensitive(k string, keys []string) bool {
	lower := strings.ToLower(k)
	for _, r := range keys {
		if lower == r || strings.Contains(lower, r) {
			return true
		}
	}
	return false
}

func redactValue(v interface{}, keys []string) interface{} {
	switch x := v.(type) {
	case map[string]interface{}:
		out := make(map[string]interface{}, len(x))
		for k, val := range x {
			if isSensitive(k, keys) {
				out[k] = "[redacted]"
			} else {
				out[k] = redactValue(val, keys)
			}
		}
		return out
	case []interface{}:
		arr := make([]interface{}, len(x))
		for i, e := range x {
			arr[i] = redactValue(e, keys)
		}
		return arr
	default:
		return v
	}
}

// DecisionRecord is the immutable record emitted to a ledger sink when a decision is issued.
type DecisionRecord struct {
	ID            string                 `json:"id"`
	AgentID       string                 `json:"agent_id"`
	SessionID     string                 `json:"session_id,omitempty"`
	Timestamp     time.Time              `json:"timestamp"`
	ActionType    string                 `json:"action_type"`
	ActionTarget  map[string]interface{} `json:"action_target"`  // redacted
	ActionInputs  map[string]interface{} `json:"action_inputs"`  // redacted
	Decision      domain.Decision        `json:"decision"`
	Reason        string                 `json:"reason,omitempty"`
	Warnings      []domain.Warning       `json:"warnings,omitempty"`
	Throttle      *domain.ThrottleInfo   `json:"throttle,omitempty"`
	BudgetSpent   float64                `json:"budget_spent_gbp,omitempty"`
	BudgetLimit   float64                `json:"budget_limit_gbp,omitempty"`
	ActionCount   int                    `json:"action_count,omitempty"`
	ActionHash    string                 `json:"action_hash,omitempty"`
	ExecutionTokenPresent bool           `json:"execution_token_present,omitempty"`
}

// LedgerSink emits immutable decision (and optional event) records.
// Implementations should be non-blocking or bounded-latency.
type LedgerSink interface {
	EmitDecision(ctx context.Context, d *DecisionRecord) error
	EmitEvent(ctx context.Context, e *domain.Event) error
	Close() error
}

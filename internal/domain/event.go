package domain

import "time"

// Event represents a ledger event (maps to kernel Operations)
type Event struct {
	EventID    string                 `json:"event_id"`
	TS         time.Time              `json:"ts"`
	Type       string                 `json:"type"`
	AgentID    string                 `json:"agent_id,omitempty"`
	SessionID  string                 `json:"session_id,omitempty"`
	Severity   string                 `json:"severity"` // info | warn | error
	PayloadJSON map[string]interface{} `json:"payload_json"`
	ActionHash  string                 `json:"action_hash,omitempty"`
	CostGBP     *float64               `json:"cost_gbp,omitempty"`
	CostTokens  *int64                 `json:"cost_tokens,omitempty"`
}

// Event types
const (
	EventTypeAgentRegistered    = "agent.registered"
	EventTypeSessionStarted     = "session.started"
	EventTypeActionProposed     = "action.proposed"
	EventTypeDecisionIssued     = "decision.issued"
	EventTypeLimitWarning       = "limit.warning"
	EventTypeLimitThrottleApplied = "limit.throttle_applied"
	EventTypeLimitExceeded      = "limit.exceeded"
	EventTypeAgentHalted        = "agent.halted"
	EventTypeRuleBlocked        = "rule.blocked"
	EventTypeLoopDetected       = "loop.detected"
)

// Event severity levels
const (
	EventSeverityInfo  = "info"
	EventSeverityWarn  = "warn"
	EventSeverityError = "error"
)

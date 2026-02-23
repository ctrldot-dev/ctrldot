package domain

import "time"

// LimitsState tracks budget and limits for an agent in a time window
type LimitsState struct {
	AgentID         string    `json:"agent_id"`
	WindowStart     time.Time `json:"window_start"`
	WindowType      string    `json:"window_type"` // daily, hourly, etc.
	BudgetSpentGBP  float64   `json:"budget_spent_gbp"`
	BudgetSpentTokens int64   `json:"budget_spent_tokens"`
	ActionCount     int       `json:"action_count"`
}

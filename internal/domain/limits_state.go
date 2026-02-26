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

// AgentLimitsResponse is the API response for GET /v1/agents/{id}/limits
type AgentLimitsResponse struct {
	AgentID      string    `json:"agent_id"`
	WindowStart  time.Time `json:"window_start"`
	WindowType   string    `json:"window_type"`
	SpentGBP     float64   `json:"spent_gbp"`
	LimitGBP     float64   `json:"limit_gbp"`
	Percentage   float64   `json:"percentage"`
	WarnPct      []float64 `json:"warn_pct"`
	ThrottlePct  float64   `json:"throttle_pct"`
	HardStopPct  float64   `json:"hard_stop_pct"`
	ActionCount  int       `json:"action_count"`
}

// LimitsConfigResponse is the API response for GET /v1/limits/config (default limits from config).
type LimitsConfigResponse struct {
	DailyBudgetGBP float64   `json:"daily_budget_gbp"`
	WarnPct        []float64 `json:"warn_pct"`
	ThrottlePct    float64   `json:"throttle_pct"`
	HardStopPct    float64   `json:"hard_stop_pct"`
}

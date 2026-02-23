package domain

// ActionProposal represents an action proposed by an agent
type ActionProposal struct {
	AgentID      string                 `json:"agent_id"`
	SessionID    string                 `json:"session_id,omitempty"`
	Intent       ActionIntent           `json:"intent"`
	Action       Action                 `json:"action"`
	Cost         CostEstimate           `json:"cost"`
	Context      ActionContext          `json:"context"`
	ResolutionToken string              `json:"resolution_token,omitempty"`
}

// ActionIntent describes the intent/goal of the action
type ActionIntent struct {
	Title  string `json:"title"`
	GoalID string `json:"goal_id,omitempty"`
}

// Action describes the action to be performed
type Action struct {
	Type   string                 `json:"type"`
	Target map[string]interface{} `json:"target"`
	Inputs map[string]interface{} `json:"inputs"`
}

// CostEstimate estimates the cost of the action
type CostEstimate struct {
	Currency      string  `json:"currency"`
	EstimatedGBP  float64 `json:"estimated_gbp"`
	EstimatedTokens int64  `json:"estimated_tokens"`
	Model         string  `json:"model"`
}

// ActionContext provides context about the action
type ActionContext struct {
	Tool string   `json:"tool"`
	Tags []string `json:"tags,omitempty"`
	Hash string   `json:"hash,omitempty"` // optional action hash for loop detection
}

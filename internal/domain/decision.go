package domain

// Decision represents the decision response from Ctrl Dot
type Decision string

const (
	DecisionAllow    Decision = "ALLOW"
	DecisionWarn     Decision = "WARN"
	DecisionThrottle Decision = "THROTTLE"
	DecisionDeny     Decision = "DENY"
	DecisionStop     Decision = "STOP"
)

// DecisionResponse is the response to an action proposal
type DecisionResponse struct {
	Decision          Decision        `json:"decision"`
	ExecutionToken    string          `json:"execution_token,omitempty"`
	Warnings          []Warning       `json:"warnings,omitempty"`
	Throttle          *ThrottleInfo   `json:"throttle,omitempty"`
	Reason            string          `json:"reason,omitempty"`
	Reasons           []Reason        `json:"reasons,omitempty"`
	Recommendation    *Recommendation `json:"recommendation,omitempty"`
	LedgerEventID     string          `json:"ledger_event_id,omitempty"`
	AutobundlePath    string          `json:"autobundle_path,omitempty"`
	AutobundleTrigger string          `json:"autobundle_trigger,omitempty"`
}

// Warning represents a warning message
type Warning struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ThrottleInfo describes throttling constraints
type ThrottleInfo struct {
	MaxParallelTasks int      `json:"max_parallel_tasks"`
	ModelPolicy       string   `json:"model_policy"`
	ToolRestrictions  []string `json:"tool_restrictions"`
}

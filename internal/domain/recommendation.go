package domain

// Recommendation is the machine-readable "why blocked" and what to do next (for DENY/STOP/THROTTLE).
type Recommendation struct {
	Kind      string   `json:"kind"`      // enable_ctrldot | enable_panic | use_resolution | tighten_scope | reduce_loop
	Title     string   `json:"title"`
	Summary   string   `json:"summary"`
	NextSteps []string `json:"next_steps"` // runnable commands
	DocsHint  string   `json:"docs_hint,omitempty"`
	Tags      []string `json:"tags,omitempty"`
}

// Reason is a structured reason with a stable code for agent logic.
type Reason struct {
	Code    string `json:"code"`
	Message string `json:"message,omitempty"`
}

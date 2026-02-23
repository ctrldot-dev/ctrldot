package domain

import "time"

// Resolution records a human (or delegated) choice from a ProposalSet
type Resolution struct {
	ResolutionID   string    `json:"resolution_id"`
	ProposalSetID  string    `json:"proposal_set_id"`
	OptionID       string    `json:"option_id"`
	Resolver       string    `json:"resolver"`        // identity, e.g. "Gareth" or "AI (delegated)"
	Rationale      string    `json:"rationale,omitempty"`
	Constraints    string    `json:"constraints,omitempty"`
	AssumptionsSnap string   `json:"assumptions_snap,omitempty"` // snapshot at resolution time
	CreatedAt      time.Time `json:"created_at"`
}

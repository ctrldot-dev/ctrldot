package domain

import "time"

// ProposalOption is one option in a ProposalSet
type ProposalOption struct {
	ID           string   `json:"id"`
	Summary      string   `json:"summary"`
	Tradeoffs    string   `json:"tradeoffs,omitempty"`
	Risks        string   `json:"risks,omitempty"`
	Reversibility string   `json:"reversibility,omitempty"`
	Assumptions  string   `json:"assumptions,omitempty"`
	Signals      []string `json:"signals,omitempty"`
}

// ProposalSet is a set of options proposed by AI (not yet resolved)
type ProposalSet struct {
	ProposalSetID     string          `json:"proposal_set_id"`
	Question          string          `json:"question"`
	FocusedNodeID     string          `json:"focused_node_id"`
	NamespaceID       string          `json:"namespace_id"`
	Options           []ProposalOption `json:"options"`
	RecommendedOptionID *string        `json:"recommended_option_id,omitempty"`
	Confidence        *float64         `json:"confidence,omitempty"`
	RawModelOutput    string          `json:"raw_model_output,omitempty"` // artefact for provenance
	CreatedAt         time.Time       `json:"created_at"`
}

// MinOptionsCount is the default minimum number of options (policy)
const MinOptionsCount = 3

// Validate validates the proposal set (e.g. minimum options)
func (p *ProposalSet) Validate(minOptions int) error {
	if minOptions <= 0 {
		minOptions = MinOptionsCount
	}
	if len(p.Options) < minOptions {
		return ErrProposalSetTooFewOptions
	}
	if p.Question == "" || p.NamespaceID == "" {
		return ErrProposalSetInvalid
	}
	return nil
}

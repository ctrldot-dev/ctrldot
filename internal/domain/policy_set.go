package domain

import "time"

// PolicySet represents a namespace-scoped policy set
type PolicySet struct {
	ID          string     `json:"policy_set_id"`
	NamespaceID string     `json:"namespace_id"`
	PolicyYAML  string     `json:"policy_yaml"`
	PolicyHash  string     `json:"policy_hash"`
	CreatedAt   time.Time  `json:"created_at"`
	CreatedSeq  int64      `json:"created_seq"`
	RetiredSeq  *int64     `json:"retired_seq,omitempty"`
}

// IsActive returns true if the policy set is active (not retired)
func (p PolicySet) IsActive() bool {
	return p.RetiredSeq == nil
}

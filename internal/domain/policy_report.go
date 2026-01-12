package domain

// PolicyReport contains the results of policy evaluation
type PolicyReport struct {
	Denies []PolicyViolation `json:"denies"`
	Warns  []PolicyViolation `json:"warns"`
	Infos  []PolicyViolation `json:"infos"`
}

// PolicyViolation represents a policy rule violation
type PolicyViolation struct {
	RuleID  string `json:"rule_id"`
	Message string `json:"message"`
}

// HasDenies returns true if there are any denies
func (p PolicyReport) HasDenies() bool {
	return len(p.Denies) > 0
}

// IsEmpty returns true if the report has no violations
func (p PolicyReport) IsEmpty() bool {
	return len(p.Denies) == 0 && len(p.Warns) == 0 && len(p.Infos) == 0
}

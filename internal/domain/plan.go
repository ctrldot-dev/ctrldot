package domain

import "time"

// Plan represents a proposed set of expanded atomic changes with policy evaluation
type Plan struct {
	ID           string                 `json:"id"`
	CreatedAt    time.Time              `json:"created_at"`
	ActorID      string                 `json:"actor_id"`
	NamespaceID  *string                `json:"namespace_id,omitempty"`
	AsOfSeq      *int64                 `json:"asof_seq,omitempty"`
	Intents      []Intent               `json:"intents"`
	Expanded     []Change               `json:"expanded"`
	Class        int                    `json:"class"`
	PolicyReport PolicyReport           `json:"policy_report"`
	Hash         string                 `json:"hash"`
}

// Validate checks if the plan is valid
func (p Plan) Validate() error {
	if p.ID == "" {
		return ErrInvalidPlanID
	}
	if p.ActorID == "" {
		return ErrInvalidActorID
	}
	if p.Hash == "" {
		return ErrInvalidPlanHash
	}
	if len(p.Intents) == 0 {
		return ErrEmptyIntents
	}
	for _, intent := range p.Intents {
		if err := intent.Validate(); err != nil {
			return err
		}
	}
	for _, change := range p.Expanded {
		if err := change.Validate(); err != nil {
			return err
		}
	}
	return nil
}

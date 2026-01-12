package domain

import "time"

// Operation represents an immutable record of an applied change
// Operations are append-only and form the canonical audit log
type Operation struct {
	ID          string    `json:"id"`
	Seq         int64     `json:"seq"`
	OccurredAt  time.Time `json:"occurred_at"`
	ActorID     string    `json:"actor_id"`
	Capabilities []string  `json:"capabilities"`
	PlanID      string    `json:"plan_id"`
	PlanHash    string    `json:"plan_hash"`
	Class       int       `json:"class"`
	Changes     []Change  `json:"changes"`
}

// Validate checks if the operation is valid
func (o Operation) Validate() error {
	if o.ID == "" {
		return ErrInvalidOperationID
	}
	if o.Seq <= 0 {
		return ErrInvalidSeq
	}
	if o.ActorID == "" {
		return ErrInvalidActorID
	}
	if o.PlanID == "" {
		return ErrInvalidPlanID
	}
	if o.PlanHash == "" {
		return ErrInvalidPlanHash
	}
	return nil
}

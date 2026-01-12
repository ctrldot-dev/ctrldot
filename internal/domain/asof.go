package domain

import "time"

// AsOf represents a point in time for querying historical state
// Either seq or time should be set, not both
type AsOf struct {
	Seq  *int64     `json:"seq,omitempty"`
	Time *time.Time `json:"time,omitempty"`
}

// IsSet returns true if either seq or time is set
func (a AsOf) IsSet() bool {
	return a.Seq != nil || a.Time != nil
}

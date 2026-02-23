package domain

import "time"

// PanicState is the persisted runtime state for panic mode.
type PanicState struct {
	Enabled     bool       `json:"enabled"`
	EnabledAt   time.Time  `json:"enabled_at"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	TTLSeconds  int        `json:"ttl_seconds"`
	Reason      string     `json:"reason,omitempty"`
}

package domain

import "time"

// Agent represents a registered agent
type Agent struct {
	AgentID     string    `json:"agent_id"`
	DisplayName string    `json:"display_name,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	DefaultMode string    `json:"default_mode"` // normal | cheap | throttled
}

// AgentMode represents agent operation modes
const (
	AgentModeNormal    = "normal"
	AgentModeCheap     = "cheap"
	AgentModeThrottled = "throttled"
)

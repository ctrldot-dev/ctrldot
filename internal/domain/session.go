package domain

import "time"

// Session represents an agent session
type Session struct {
	SessionID string                 `json:"session_id"`
	AgentID   string                 `json:"agent_id"`
	StartedAt time.Time               `json:"started_at"`
	EndedAt   *time.Time              `json:"ended_at,omitempty"`
	Metadata  map[string]interface{}  `json:"metadata,omitempty"`
}

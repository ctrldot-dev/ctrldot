package client

import (
	"time"

	"github.com/futurematic/kernel/internal/domain"
)

// PlanRequest represents a plan request
type PlanRequest struct {
	ActorID     string          `json:"actor_id"`
	Capabilities []string        `json:"capabilities"`
	NamespaceID *string          `json:"namespace_id,omitempty"`
	AsOf        domain.AsOf     `json:"asof"`
	Intents     []domain.Intent  `json:"intents"`
}

// PlanResponse represents a plan response
type PlanResponse struct {
	ID           string                 `json:"id"`
	CreatedAt    time.Time              `json:"created_at"`
	ActorID      string                 `json:"actor_id"`
	NamespaceID  *string                `json:"namespace_id,omitempty"`
	AsOfSeq      *int64                 `json:"asof_seq,omitempty"`
	Intents      []domain.Intent        `json:"intents"`
	Expanded     []domain.Change        `json:"expanded"`
	Class        int                    `json:"class"`
	PolicyReport domain.PolicyReport    `json:"policy_report"`
	Hash         string                 `json:"hash"`
}

// ApplyRequest represents an apply request
type ApplyRequest struct {
	ActorID     string   `json:"actor_id"`
	Capabilities []string `json:"capabilities"`
	PlanID      string   `json:"plan_id"`
	PlanHash    string   `json:"plan_hash"`
}

// ApplyResponse represents an apply response
type ApplyResponse struct {
	ID           string          `json:"id"`
	Seq          int64           `json:"seq"`
	OccurredAt   time.Time       `json:"occurred_at"`
	ActorID      string          `json:"actor_id"`
	Capabilities []string        `json:"capabilities"`
	PlanID       string          `json:"plan_id"`
	PlanHash     string          `json:"plan_hash"`
	Class        int             `json:"class"`
	Changes      []domain.Change `json:"changes"`
}

// ExpandRequest represents an expand request
type ExpandRequest struct {
	IDs         []string
	NamespaceID *string
	Depth       int
	AsOfSeq     int64
	AsOfTime    *time.Time
}

// ExpandResponse represents an expand response
type ExpandResponse struct {
	Nodes            []domain.Node            `json:"nodes"`
	Links            []domain.Link            `json:"links"`
	Materials        []domain.Material        `json:"materials"`
	RoleAssignments  []domain.RoleAssignment  `json:"role_assignments"`
}

// HistoryRequest represents a history request
type HistoryRequest struct {
	Target string
	Limit  int
}

// HistoryResponse represents a history response (array of operations)
type HistoryResponse []domain.Operation

// DiffRequest represents a diff request
type DiffRequest struct {
	ASeq   int64
	BSeq   int64
	Target string
}

// DiffResponse represents a diff response
type DiffResponse struct {
	Changes []domain.Change `json:"changes"`
}

// HealthResponse represents a health check response
type HealthResponse struct {
	OK bool `json:"ok"`
}

// ErrorResponse represents an API error response
type ErrorResponse struct {
	Error struct {
		Code    string                 `json:"code"`
		Message string                 `json:"message"`
		Details map[string]interface{} `json:"details,omitempty"`
	} `json:"error"`
}

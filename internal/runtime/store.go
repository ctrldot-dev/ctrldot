package runtime

import (
	"context"

	"github.com/futurematic/kernel/internal/domain"
)

// EventFilter filters events for ListEvents.
type EventFilter struct {
	AgentID *string
	SinceTS *int64 // unix milliseconds
	Limit   int
}

// RuntimeStore holds mutable Ctrl Dot operational state (agents, sessions, limits, events, halt).
// It does not include Kernel ledger operations (operations, plans, policy, etc.).
type RuntimeStore interface {
	// Lifecycle
	Migrate(ctx context.Context) error
	Close() error

	// Agents
	CreateAgent(ctx context.Context, a domain.Agent) error
	ListAgents(ctx context.Context) ([]domain.Agent, error)
	GetAgent(ctx context.Context, id string) (*domain.Agent, error)
	IsAgentHalted(ctx context.Context, agentID string) (bool, error)

	// Sessions
	CreateSession(ctx context.Context, s domain.Session) error
	GetSession(ctx context.Context, sessionID string) (*domain.Session, error)
	EndSession(ctx context.Context, sessionID string) error

	// Limits state (windowStart is unix milliseconds for daily window start)
	GetLimitsState(ctx context.Context, agentID string, windowStart int64, windowType string) (*domain.LimitsState, error)
	UpdateLimitsState(ctx context.Context, state domain.LimitsState) error

	// Events (append-only runtime log; no Kernel op_seq required)
	AppendEvent(ctx context.Context, e *domain.Event) error
	ListEvents(ctx context.Context, filter EventFilter) ([]domain.Event, error)
	GetEvent(ctx context.Context, eventID string) (*domain.Event, error)

	// Agent control
	HaltAgent(ctx context.Context, agentID string, reason string) error
	ResumeAgent(ctx context.Context, agentID string) error

	// Panic mode (persisted)
	GetPanicState(ctx context.Context) (*domain.PanicState, error)
	SetPanicState(ctx context.Context, state domain.PanicState) error
}

package runtime

import (
	"context"

	"github.com/futurematic/kernel/internal/domain"
	"github.com/futurematic/kernel/internal/store"
)

// PostgresStore wraps the full store.Store to implement RuntimeStore (Ctrl Dot runtime only).
// Caller must run migrations (including 0008 for nullable op_seq) before using AppendEvent.
type PostgresStore struct {
	st store.Store
}

// NewPostgresStore returns a RuntimeStore that delegates to the given store.Store.
func NewPostgresStore(st store.Store) *PostgresStore {
	return &PostgresStore{st: st}
}

// Migrate is a no-op; caller is responsible for running Postgres migrations.
func (s *PostgresStore) Migrate(ctx context.Context) error {
	return nil
}

// Close is a no-op; the owner of the underlying store.Store is responsible for closing it.
func (s *PostgresStore) Close() error {
	return nil
}

// CreateAgent delegates to store.CreateAgent.
func (s *PostgresStore) CreateAgent(ctx context.Context, a domain.Agent) error {
	return s.st.CreateAgent(ctx, a)
}

// ListAgents delegates to store.ListAgents.
func (s *PostgresStore) ListAgents(ctx context.Context) ([]domain.Agent, error) {
	return s.st.ListAgents(ctx)
}

// GetAgent delegates to store.GetAgent.
func (s *PostgresStore) GetAgent(ctx context.Context, id string) (*domain.Agent, error) {
	return s.st.GetAgent(ctx, id)
}

// IsAgentHalted delegates to store.IsAgentHalted.
func (s *PostgresStore) IsAgentHalted(ctx context.Context, agentID string) (bool, error) {
	return s.st.IsAgentHalted(ctx, agentID)
}

// CreateSession delegates to store.CreateSession.
func (s *PostgresStore) CreateSession(ctx context.Context, sess domain.Session) error {
	return s.st.CreateSession(ctx, sess)
}

// GetSession delegates to store.GetSession.
func (s *PostgresStore) GetSession(ctx context.Context, sessionID string) (*domain.Session, error) {
	return s.st.GetSession(ctx, sessionID)
}

// EndSession delegates to store.EndSession.
func (s *PostgresStore) EndSession(ctx context.Context, sessionID string) error {
	return s.st.EndSession(ctx, sessionID)
}

// GetLimitsState delegates to store.GetLimitsState (windowStart is unix ms).
func (s *PostgresStore) GetLimitsState(ctx context.Context, agentID string, windowStart int64, windowType string) (*domain.LimitsState, error) {
	return s.st.GetLimitsState(ctx, agentID, windowStart, windowType)
}

// UpdateLimitsState delegates to store.UpdateLimitsState.
func (s *PostgresStore) UpdateLimitsState(ctx context.Context, state domain.LimitsState) error {
	return s.st.UpdateLimitsState(ctx, state)
}

// AppendEvent delegates to store.AppendEvent (runtime-only; op_seq NULL).
func (s *PostgresStore) AppendEvent(ctx context.Context, e *domain.Event) error {
	return s.st.AppendEvent(ctx, *e)
}

// ListEvents delegates to store.GetEvents.
func (s *PostgresStore) ListEvents(ctx context.Context, filter EventFilter) ([]domain.Event, error) {
	return s.st.GetEvents(ctx, filter.AgentID, filter.SinceTS, filter.Limit)
}

// GetEvent delegates to store.GetEvent.
func (s *PostgresStore) GetEvent(ctx context.Context, eventID string) (*domain.Event, error) {
	return s.st.GetEvent(ctx, eventID)
}

// HaltAgent delegates to store.HaltAgent.
func (s *PostgresStore) HaltAgent(ctx context.Context, agentID string, reason string) error {
	return s.st.HaltAgent(ctx, agentID, reason)
}

// ResumeAgent delegates to store.ResumeAgent.
func (s *PostgresStore) ResumeAgent(ctx context.Context, agentID string) error {
	return s.st.ResumeAgent(ctx, agentID)
}

// GetPanicState delegates to store.GetPanicState.
func (s *PostgresStore) GetPanicState(ctx context.Context) (*domain.PanicState, error) {
	return s.st.GetPanicState(ctx)
}

// SetPanicState delegates to store.SetPanicState.
func (s *PostgresStore) SetPanicState(ctx context.Context, state domain.PanicState) error {
	return s.st.SetPanicState(ctx, state)
}

// Ensure PostgresStore implements RuntimeStore.
var _ RuntimeStore = (*PostgresStore)(nil)

package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/futurematic/kernel/internal/domain"
)

// Ctrl Dot: Agents (non-transactional)

// CreateAgent creates an agent
func (s *PostgresStore) CreateAgent(ctx context.Context, agent domain.Agent) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO ctrldot_agents (agent_id, display_name, created_at, default_mode)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (agent_id) DO NOTHING`,
		agent.AgentID, agent.DisplayName, agent.CreatedAt, agent.DefaultMode,
	)
	if err != nil {
		return fmt.Errorf("failed to create agent: %w", err)
	}
	return nil
}

// GetAgent retrieves an agent
func (s *PostgresStore) GetAgent(ctx context.Context, agentID string) (*domain.Agent, error) {
	var agent domain.Agent
	err := s.db.QueryRowContext(ctx,
		`SELECT agent_id, display_name, created_at, default_mode
		 FROM ctrldot_agents
		 WHERE agent_id = $1`,
		agentID,
	).Scan(&agent.AgentID, &agent.DisplayName, &agent.CreatedAt, &agent.DefaultMode)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get agent: %w", err)
	}
	return &agent, nil
}

// ListAgents lists all agents
func (s *PostgresStore) ListAgents(ctx context.Context) ([]domain.Agent, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT agent_id, display_name, created_at, default_mode
		 FROM ctrldot_agents
		 ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list agents: %w", err)
	}
	defer rows.Close()

	var agents []domain.Agent
	for rows.Next() {
		var agent domain.Agent
		if err := rows.Scan(&agent.AgentID, &agent.DisplayName, &agent.CreatedAt, &agent.DefaultMode); err != nil {
			return nil, fmt.Errorf("failed to scan agent: %w", err)
		}
		agents = append(agents, agent)
	}
	return agents, nil
}

// IsAgentHalted checks if an agent is halted
func (s *PostgresStore) IsAgentHalted(ctx context.Context, agentID string) (bool, error) {
	var count int
	err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM ctrldot_halted_agents WHERE agent_id = $1`,
		agentID,
	).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check halted agent: %w", err)
	}
	return count > 0, nil
}

// Ctrl Dot: Sessions (non-transactional)

// CreateSession creates a session
func (s *PostgresStore) CreateSession(ctx context.Context, session domain.Session) error {
	metadataJSON, _ := json.Marshal(session.Metadata)
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO ctrldot_sessions (session_id, agent_id, started_at, ended_at, metadata_json)
		 VALUES ($1, $2, $3, $4, $5)`,
		session.SessionID, session.AgentID, session.StartedAt, session.EndedAt, metadataJSON,
	)
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	return nil
}

// GetSession retrieves a session
func (s *PostgresStore) GetSession(ctx context.Context, sessionID string) (*domain.Session, error) {
	var session domain.Session
	var metadataJSON []byte
	var endedAt sql.NullTime
	err := s.db.QueryRowContext(ctx,
		`SELECT session_id, agent_id, started_at, ended_at, metadata_json
		 FROM ctrldot_sessions
		 WHERE session_id = $1`,
		sessionID,
	).Scan(&session.SessionID, &session.AgentID, &session.StartedAt, &endedAt, &metadataJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}
	if endedAt.Valid {
		session.EndedAt = &endedAt.Time
	}
	if len(metadataJSON) > 0 {
		json.Unmarshal(metadataJSON, &session.Metadata)
	}
	return &session, nil
}

// EndSession ends a session
func (s *PostgresStore) EndSession(ctx context.Context, sessionID string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE ctrldot_sessions SET ended_at = now() WHERE session_id = $1`,
		sessionID,
	)
	if err != nil {
		return fmt.Errorf("failed to end session: %w", err)
	}
	return nil
}

// Ctrl Dot: Events (non-transactional)

// AppendEvent appends an event without a Kernel operation (op_seq NULL).
// Use this for runtime-only event log when not using Kernel Plan/Apply.
// Requires migration 0008 (op_seq nullable).
func (s *PostgresStore) AppendEvent(ctx context.Context, event domain.Event) error {
	payloadJSON, _ := json.Marshal(event.PayloadJSON)
	var sessionID interface{} = nil
	if event.SessionID != "" {
		sessionID = event.SessionID
	}
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO ctrldot_events (event_id, op_seq, event_type, agent_id, session_id, severity, payload_json, action_hash, cost_gbp, cost_tokens, created_at)
		 VALUES ($1, NULL, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		event.EventID, event.Type, event.AgentID, sessionID, event.Severity,
		payloadJSON, event.ActionHash, event.CostGBP, event.CostTokens, event.TS,
	)
	if err != nil {
		return fmt.Errorf("failed to append event: %w", err)
	}
	return nil
}

// GetEvents retrieves events with optional filtering
func (s *PostgresStore) GetEvents(ctx context.Context, agentID *string, sinceTS *int64, limit int) ([]domain.Event, error) {
	query := `SELECT event_id, op_seq, event_type, agent_id, session_id, severity, payload_json, action_hash, cost_gbp, cost_tokens, created_at
			  FROM ctrldot_events
			  WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if agentID != nil {
		query += fmt.Sprintf(" AND agent_id = $%d", argIdx)
		args = append(args, *agentID)
		argIdx++
	}
	if sinceTS != nil {
		query += fmt.Sprintf(" AND EXTRACT(EPOCH FROM created_at) * 1000 >= $%d", argIdx)
		args = append(args, *sinceTS)
		argIdx++
	}

	query += " ORDER BY created_at DESC"
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIdx)
		args = append(args, limit)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query events: %w", err)
	}
	defer rows.Close()

	var events []domain.Event
	for rows.Next() {
		var event domain.Event
		var payloadJSON []byte
		var actionHash sql.NullString
		var costGBP sql.NullFloat64
		var costTokens sql.NullInt64
		var sessionID sql.NullString
		var opSeq sql.NullInt64
		err := rows.Scan(&event.EventID, &opSeq, &event.Type, &event.AgentID, &sessionID,
			&event.Severity, &payloadJSON, &actionHash, &costGBP, &costTokens, &event.TS)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}
		if sessionID.Valid {
			event.SessionID = sessionID.String
		}
		if len(payloadJSON) > 0 {
			json.Unmarshal(payloadJSON, &event.PayloadJSON)
		}
		if actionHash.Valid {
			event.ActionHash = actionHash.String
		}
		if costGBP.Valid {
			event.CostGBP = &costGBP.Float64
		}
		if costTokens.Valid {
			event.CostTokens = &costTokens.Int64
		}
		events = append(events, event)
	}
	return events, nil
}

// GetEvent retrieves a single event
func (s *PostgresStore) GetEvent(ctx context.Context, eventID string) (*domain.Event, error) {
	var event domain.Event
	var payloadJSON []byte
	var actionHash sql.NullString
	var costGBP sql.NullFloat64
	var costTokens sql.NullInt64
	var sessionID sql.NullString
	var opSeq sql.NullInt64
	err := s.db.QueryRowContext(ctx,
		`SELECT event_id, op_seq, event_type, agent_id, session_id, severity, payload_json, action_hash, cost_gbp, cost_tokens, created_at
		 FROM ctrldot_events
		 WHERE event_id = $1`,
		eventID,
	).Scan(&event.EventID, &opSeq, &event.Type, &event.AgentID, &sessionID,
		&event.Severity, &payloadJSON, &actionHash, &costGBP, &costTokens, &event.TS)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get event: %w", err)
	}
	if sessionID.Valid {
		event.SessionID = sessionID.String
	}
	if len(payloadJSON) > 0 {
		json.Unmarshal(payloadJSON, &event.PayloadJSON)
	}
	if actionHash.Valid {
		event.ActionHash = actionHash.String
	}
	if costGBP.Valid {
		event.CostGBP = &costGBP.Float64
	}
	if costTokens.Valid {
		event.CostTokens = &costTokens.Int64
	}
	return &event, nil
}

// Ctrl Dot: Limits State (non-transactional)

// GetLimitsState retrieves limits state for an agent
func (s *PostgresStore) GetLimitsState(ctx context.Context, agentID string, windowStart int64, windowType string) (*domain.LimitsState, error) {
	var state domain.LimitsState
	windowStartTime := time.Unix(windowStart/1000, 0)
	err := s.db.QueryRowContext(ctx,
		`SELECT agent_id, window_start, window_type, budget_spent_gbp, budget_spent_tokens, action_count
		 FROM ctrldot_limits_state
		 WHERE agent_id = $1 AND window_start = $2 AND window_type = $3`,
		agentID, windowStartTime, windowType,
	).Scan(&state.AgentID, &state.WindowStart, &state.WindowType, &state.BudgetSpentGBP, &state.BudgetSpentTokens, &state.ActionCount)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get limits state: %w", err)
	}
	return &state, nil
}

// UpdateLimitsState updates limits state
func (s *PostgresStore) UpdateLimitsState(ctx context.Context, state domain.LimitsState) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO ctrldot_limits_state (agent_id, window_start, window_type, budget_spent_gbp, budget_spent_tokens, action_count)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 ON CONFLICT (agent_id, window_start, window_type)
		 DO UPDATE SET budget_spent_gbp = EXCLUDED.budget_spent_gbp,
		               budget_spent_tokens = EXCLUDED.budget_spent_tokens,
		               action_count = EXCLUDED.action_count`,
		state.AgentID, state.WindowStart, state.WindowType, state.BudgetSpentGBP, state.BudgetSpentTokens, state.ActionCount,
	)
	if err != nil {
		return fmt.Errorf("failed to update limits state: %w", err)
	}
	return nil
}

// Ctrl Dot: Agent Control (non-transactional)

// HaltAgent halts an agent
func (s *PostgresStore) HaltAgent(ctx context.Context, agentID string, reason string) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO ctrldot_halted_agents (agent_id, reason, halted_at)
		 VALUES ($1, $2, now())
		 ON CONFLICT (agent_id) DO UPDATE SET reason = EXCLUDED.reason, halted_at = now()`,
		agentID, reason,
	)
	if err != nil {
		return fmt.Errorf("failed to halt agent: %w", err)
	}
	return nil
}

// ResumeAgent resumes an agent
func (s *PostgresStore) ResumeAgent(ctx context.Context, agentID string) error {
	_, err := s.db.ExecContext(ctx,
		`DELETE FROM ctrldot_halted_agents WHERE agent_id = $1`,
		agentID,
	)
	if err != nil {
		return fmt.Errorf("failed to resume agent: %w", err)
	}
	return nil
}

// GetPanicState returns the current panic mode state (single row id=1).
func (s *PostgresStore) GetPanicState(ctx context.Context) (*domain.PanicState, error) {
	var enabled bool
	var enabledAt, expiresAt sql.NullTime
	var reasonStr sql.NullString
	var ttlSeconds int
	err := s.db.QueryRowContext(ctx,
		`SELECT enabled, enabled_at, expires_at, ttl_seconds, reason FROM ctrldot_panic_state WHERE id = 1`).Scan(
		&enabled, &enabledAt, &expiresAt, &ttlSeconds, &reasonStr)
	if err != nil {
		if err == sql.ErrNoRows {
			return &domain.PanicState{Enabled: false}, nil
		}
		return nil, fmt.Errorf("get panic state: %w", err)
	}
	st := &domain.PanicState{Enabled: enabled, TTLSeconds: ttlSeconds}
	if enabledAt.Valid {
		st.EnabledAt = enabledAt.Time
	}
	if expiresAt.Valid {
		st.ExpiresAt = &expiresAt.Time
	}
	if reasonStr.Valid {
		st.Reason = reasonStr.String
	}
	return st, nil
}

// SetPanicState updates the panic mode state.
func (s *PostgresStore) SetPanicState(ctx context.Context, state domain.PanicState) error {
	var expiresAt interface{}
	if state.ExpiresAt != nil {
		expiresAt = state.ExpiresAt
	} else {
		expiresAt = nil
	}
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO ctrldot_panic_state (id, enabled, enabled_at, expires_at, ttl_seconds, reason)
		 VALUES (1, $1, $2, $3, $4, $5)
		 ON CONFLICT (id) DO UPDATE SET enabled = EXCLUDED.enabled, enabled_at = EXCLUDED.enabled_at, expires_at = EXCLUDED.expires_at, ttl_seconds = EXCLUDED.ttl_seconds, reason = EXCLUDED.reason`,
		state.Enabled, state.EnabledAt, expiresAt, state.TTLSeconds, state.Reason)
	if err != nil {
		return fmt.Errorf("set panic state: %w", err)
	}
	return nil
}

// Transactional methods (PostgresTx)

// CreateAgentTx creates an agent in a transaction
func (t *PostgresTx) CreateAgentTx(ctx context.Context, agent domain.Agent) error {
	_, err := t.tx.ExecContext(ctx,
		`INSERT INTO ctrldot_agents (agent_id, display_name, created_at, default_mode)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (agent_id) DO NOTHING`,
		agent.AgentID, agent.DisplayName, agent.CreatedAt, agent.DefaultMode,
	)
	if err != nil {
		return fmt.Errorf("failed to create agent: %w", err)
	}
	return nil
}

// CreateSessionTx creates a session in a transaction
func (t *PostgresTx) CreateSessionTx(ctx context.Context, session domain.Session) error {
	metadataJSON, _ := json.Marshal(session.Metadata)
	_, err := t.tx.ExecContext(ctx,
		`INSERT INTO ctrldot_sessions (session_id, agent_id, started_at, ended_at, metadata_json)
		 VALUES ($1, $2, $3, $4, $5)`,
		session.SessionID, session.AgentID, session.StartedAt, session.EndedAt, metadataJSON,
	)
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	return nil
}

// EndSessionTx ends a session in a transaction
func (t *PostgresTx) EndSessionTx(ctx context.Context, sessionID string) error {
	_, err := t.tx.ExecContext(ctx,
		`UPDATE ctrldot_sessions SET ended_at = now() WHERE session_id = $1`,
		sessionID,
	)
	if err != nil {
		return fmt.Errorf("failed to end session: %w", err)
	}
	return nil
}

// AppendEventTx appends an event in a transaction
func (t *PostgresTx) AppendEventTx(ctx context.Context, event domain.Event, opSeq int64) error {
	payloadJSON, _ := json.Marshal(event.PayloadJSON)
	
	// Use NULL for empty session_id to avoid foreign key constraint violation
	var sessionID interface{} = nil
	if event.SessionID != "" {
		sessionID = event.SessionID
	}
	
	_, err := t.tx.ExecContext(ctx,
		`INSERT INTO ctrldot_events (event_id, op_seq, event_type, agent_id, session_id, severity, payload_json, action_hash, cost_gbp, cost_tokens, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		event.EventID, opSeq, event.Type, event.AgentID, sessionID, event.Severity,
		payloadJSON, event.ActionHash, event.CostGBP, event.CostTokens, event.TS,
	)
	if err != nil {
		return fmt.Errorf("failed to append event: %w", err)
	}
	return nil
}

// UpdateLimitsStateTx updates limits state in a transaction
func (t *PostgresTx) UpdateLimitsStateTx(ctx context.Context, state domain.LimitsState) error {
	_, err := t.tx.ExecContext(ctx,
		`INSERT INTO ctrldot_limits_state (agent_id, window_start, window_type, budget_spent_gbp, budget_spent_tokens, action_count)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 ON CONFLICT (agent_id, window_start, window_type)
		 DO UPDATE SET budget_spent_gbp = EXCLUDED.budget_spent_gbp,
		               budget_spent_tokens = EXCLUDED.budget_spent_tokens,
		               action_count = EXCLUDED.action_count`,
		state.AgentID, state.WindowStart, state.WindowType, state.BudgetSpentGBP, state.BudgetSpentTokens, state.ActionCount,
	)
	if err != nil {
		return fmt.Errorf("failed to update limits state: %w", err)
	}
	return nil
}

// HaltAgentTx halts an agent in a transaction
func (t *PostgresTx) HaltAgentTx(ctx context.Context, agentID string, reason string) error {
	_, err := t.tx.ExecContext(ctx,
		`INSERT INTO ctrldot_halted_agents (agent_id, reason, halted_at)
		 VALUES ($1, $2, now())
		 ON CONFLICT (agent_id) DO UPDATE SET reason = EXCLUDED.reason, halted_at = now()`,
		agentID, reason,
	)
	if err != nil {
		return fmt.Errorf("failed to halt agent: %w", err)
	}
	return nil
}

// ResumeAgentTx resumes an agent in a transaction
func (t *PostgresTx) ResumeAgentTx(ctx context.Context, agentID string) error {
	_, err := t.tx.ExecContext(ctx,
		`DELETE FROM ctrldot_halted_agents WHERE agent_id = $1`,
		agentID,
	)
	if err != nil {
		return fmt.Errorf("failed to resume agent: %w", err)
	}
	return nil
}

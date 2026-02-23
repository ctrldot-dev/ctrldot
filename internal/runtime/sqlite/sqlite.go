package sqlite

import (
	"context"
	"database/sql"
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/futurematic/kernel/internal/domain"
	"github.com/futurematic/kernel/internal/runtime"
	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// Store implements runtime.RuntimeStore using SQLite.
type Store struct {
	db *sql.DB
}

// Open opens or creates the SQLite database at path, sets PRAGMAs, and runs migrations.
// Path is expanded: ~ is replaced by the user's home directory.
func Open(ctx context.Context, path string) (*Store, error) {
	path = expandPath(path)
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return nil, fmt.Errorf("create db dir: %w", err)
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	// PRAGMAs for single-writer daemon + readers
	for _, s := range []string{
		"PRAGMA journal_mode=WAL",
		"PRAGMA synchronous=NORMAL",
		"PRAGMA busy_timeout=5000",
	} {
		if _, err := db.Exec(s); err != nil {
			db.Close()
			return nil, fmt.Errorf("%s: %w", s, err)
		}
	}
	s := &Store{db: db}
	if err := s.Migrate(ctx); err != nil {
		db.Close()
		return nil, err
	}
	return s, nil
}

func expandPath(p string) string {
	if len(p) >= 2 && p[:2] == "~/" {
		home, _ := os.UserHomeDir()
		if home != "" {
			return filepath.Join(home, p[2:])
		}
	}
	return p
}

// Migrate runs embedded migrations.
func (s *Store) Migrate(ctx context.Context) error {
	for _, name := range []string{"migrations/0001_ctrldot_runtime.sql", "migrations/0002_panic_state.sql"} {
		sqlBytes, err := migrationsFS.ReadFile(name)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", name, err)
		}
		if _, err := s.db.ExecContext(ctx, string(sqlBytes)); err != nil {
			return fmt.Errorf("migrate %s: %w", name, err)
		}
	}
	return nil
}

// Close closes the database.
func (s *Store) Close() error {
	return s.db.Close()
}

// CreateAgent implements runtime.RuntimeStore.
func (s *Store) CreateAgent(ctx context.Context, a domain.Agent) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO ctrldot_agents (agent_id, display_name, created_at, default_mode) VALUES (?, ?, ?, ?) ON CONFLICT (agent_id) DO NOTHING`,
		a.AgentID, a.DisplayName, a.CreatedAt.Format(time.RFC3339), a.DefaultMode,
	)
	if err != nil {
		return fmt.Errorf("create agent: %w", err)
	}
	return nil
}

// ListAgents implements runtime.RuntimeStore.
func (s *Store) ListAgents(ctx context.Context) ([]domain.Agent, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT agent_id, display_name, created_at, default_mode FROM ctrldot_agents ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("list agents: %w", err)
	}
	defer rows.Close()
	var out []domain.Agent
	for rows.Next() {
		var a domain.Agent
		var createdAt string
		if err := rows.Scan(&a.AgentID, &a.DisplayName, &createdAt, &a.DefaultMode); err != nil {
			return nil, err
		}
		a.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		out = append(out, a)
	}
	return out, rows.Err()
}

// GetAgent implements runtime.RuntimeStore.
func (s *Store) GetAgent(ctx context.Context, id string) (*domain.Agent, error) {
	var a domain.Agent
	var createdAt string
	err := s.db.QueryRowContext(ctx,
		`SELECT agent_id, display_name, created_at, default_mode FROM ctrldot_agents WHERE agent_id = ?`,
		id,
	).Scan(&a.AgentID, &a.DisplayName, &createdAt, &a.DefaultMode)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get agent: %w", err)
	}
	a.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	return &a, nil
}

// IsAgentHalted implements runtime.RuntimeStore.
func (s *Store) IsAgentHalted(ctx context.Context, agentID string) (bool, error) {
	var n int
	err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM ctrldot_halted_agents WHERE agent_id = ?`, agentID).Scan(&n)
	if err != nil {
		return false, fmt.Errorf("is agent halted: %w", err)
	}
	return n > 0, nil
}

// CreateSession implements runtime.RuntimeStore.
func (s *Store) CreateSession(ctx context.Context, sess domain.Session) error {
	meta, _ := json.Marshal(sess.Metadata)
	var endedAt interface{}
	if sess.EndedAt != nil {
		endedAt = sess.EndedAt.Format(time.RFC3339)
	}
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO ctrldot_sessions (session_id, agent_id, started_at, ended_at, metadata_json) VALUES (?, ?, ?, ?, ?)`,
		sess.SessionID, sess.AgentID, sess.StartedAt.Format(time.RFC3339), endedAt, string(meta),
	)
	if err != nil {
		return fmt.Errorf("create session: %w", err)
	}
	return nil
}

// GetSession implements runtime.RuntimeStore.
func (s *Store) GetSession(ctx context.Context, sessionID string) (*domain.Session, error) {
	var sess domain.Session
	var startedAt, metadataJSON string
	var endedAt sql.NullString
	err := s.db.QueryRowContext(ctx,
		`SELECT session_id, agent_id, started_at, ended_at, metadata_json FROM ctrldot_sessions WHERE session_id = ?`,
		sessionID,
	).Scan(&sess.SessionID, &sess.AgentID, &startedAt, &endedAt, &metadataJSON)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get session: %w", err)
	}
	sess.StartedAt, _ = time.Parse(time.RFC3339, startedAt)
	if endedAt.Valid {
		t, _ := time.Parse(time.RFC3339, endedAt.String)
		sess.EndedAt = &t
	}
	if metadataJSON != "" {
		json.Unmarshal([]byte(metadataJSON), &sess.Metadata)
	}
	return &sess, nil
}

// EndSession implements runtime.RuntimeStore.
func (s *Store) EndSession(ctx context.Context, sessionID string) error {
	_, err := s.db.ExecContext(ctx, `UPDATE ctrldot_sessions SET ended_at = ? WHERE session_id = ?`, time.Now().UTC().Format(time.RFC3339), sessionID)
	if err != nil {
		return fmt.Errorf("end session: %w", err)
	}
	return nil
}

// GetLimitsState implements runtime.RuntimeStore (windowStart is unix ms).
func (s *Store) GetLimitsState(ctx context.Context, agentID string, windowStart int64, windowType string) (*domain.LimitsState, error) {
	var st domain.LimitsState
	err := s.db.QueryRowContext(ctx,
		`SELECT agent_id, window_start, window_type, budget_spent_gbp, budget_spent_tokens, action_count FROM ctrldot_limits_state WHERE agent_id = ? AND window_start = ? AND window_type = ?`,
		agentID, windowStart, windowType,
	).Scan(&st.AgentID, &windowStart, &st.WindowType, &st.BudgetSpentGBP, &st.BudgetSpentTokens, &st.ActionCount)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get limits state: %w", err)
	}
	st.WindowStart = time.Unix(windowStart/1000, 0)
	return &st, nil
}

// UpdateLimitsState implements runtime.RuntimeStore.
func (s *Store) UpdateLimitsState(ctx context.Context, state domain.LimitsState) error {
	windowStart := state.WindowStart.Unix() * 1000
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO ctrldot_limits_state (agent_id, window_start, window_type, budget_spent_gbp, budget_spent_tokens, action_count) VALUES (?, ?, ?, ?, ?, ?)
		 ON CONFLICT (agent_id, window_start, window_type) DO UPDATE SET budget_spent_gbp = excluded.budget_spent_gbp, budget_spent_tokens = excluded.budget_spent_tokens, action_count = excluded.action_count`,
		state.AgentID, windowStart, state.WindowType, state.BudgetSpentGBP, state.BudgetSpentTokens, state.ActionCount,
	)
	if err != nil {
		return fmt.Errorf("update limits state: %w", err)
	}
	return nil
}

// AppendEvent implements runtime.RuntimeStore.
func (s *Store) AppendEvent(ctx context.Context, e *domain.Event) error {
	payload, _ := json.Marshal(e.PayloadJSON)
	var sessionID interface{}
	if e.SessionID != "" {
		sessionID = e.SessionID
	}
	var costGBP, costTokens interface{}
	if e.CostGBP != nil {
		costGBP = *e.CostGBP
	}
	if e.CostTokens != nil {
		costTokens = *e.CostTokens
	}
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO ctrldot_events (event_id, event_type, agent_id, session_id, severity, payload_json, action_hash, cost_gbp, cost_tokens, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		e.EventID, e.Type, e.AgentID, sessionID, e.Severity, string(payload), e.ActionHash, costGBP, costTokens, e.TS.Format(time.RFC3339),
	)
	if err != nil {
		return fmt.Errorf("append event: %w", err)
	}
	return nil
}

// ListEvents implements runtime.RuntimeStore.
func (s *Store) ListEvents(ctx context.Context, filter runtime.EventFilter) ([]domain.Event, error) {
	query := `SELECT event_id, event_type, agent_id, session_id, severity, payload_json, action_hash, cost_gbp, cost_tokens, created_at FROM ctrldot_events WHERE 1=1`
	args := []interface{}{}
	argIdx := 1
	if filter.AgentID != nil {
		query += fmt.Sprintf(" AND agent_id = ?")
		args = append(args, *filter.AgentID)
		argIdx++
	}
	if filter.SinceTS != nil {
		// created_at is ISO8601 text; compare with datetime from unix ms
		query += " AND created_at >= datetime(?/1000, 'unixepoch')"
		args = append(args, *filter.SinceTS)
		argIdx++
	}
	query += " ORDER BY created_at DESC"
	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT ?")
		args = append(args, filter.Limit)
	}
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list events: %w", err)
	}
	defer rows.Close()
	var out []domain.Event
	for rows.Next() {
		var e domain.Event
		var payloadJSON, sessionID sql.NullString
		var costGBP sql.NullFloat64
		var costTokens sql.NullInt64
		var createdAt string
		if err := rows.Scan(&e.EventID, &e.Type, &e.AgentID, &sessionID, &e.Severity, &payloadJSON, &e.ActionHash, &costGBP, &costTokens, &createdAt); err != nil {
			return nil, err
		}
		e.TS, _ = time.Parse(time.RFC3339, createdAt)
		if sessionID.Valid {
			e.SessionID = sessionID.String
		}
		if payloadJSON.Valid && payloadJSON.String != "" {
			json.Unmarshal([]byte(payloadJSON.String), &e.PayloadJSON)
		}
		if costGBP.Valid {
			e.CostGBP = &costGBP.Float64
		}
		if costTokens.Valid {
			e.CostTokens = &costTokens.Int64
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// GetEvent implements runtime.RuntimeStore.
func (s *Store) GetEvent(ctx context.Context, eventID string) (*domain.Event, error) {
	var e domain.Event
	var payloadJSON, sessionID sql.NullString
	var costGBP sql.NullFloat64
	var costTokens sql.NullInt64
	var createdAt string
	err := s.db.QueryRowContext(ctx,
		`SELECT event_id, event_type, agent_id, session_id, severity, payload_json, action_hash, cost_gbp, cost_tokens, created_at FROM ctrldot_events WHERE event_id = ?`,
		eventID,
	).Scan(&e.EventID, &e.Type, &e.AgentID, &sessionID, &e.Severity, &payloadJSON, &e.ActionHash, &costGBP, &costTokens, &createdAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get event: %w", err)
	}
	e.TS, _ = time.Parse(time.RFC3339, createdAt)
	if sessionID.Valid {
		e.SessionID = sessionID.String
	}
	if payloadJSON.Valid && payloadJSON.String != "" {
		json.Unmarshal([]byte(payloadJSON.String), &e.PayloadJSON)
	}
	if costGBP.Valid {
		e.CostGBP = &costGBP.Float64
	}
	if costTokens.Valid {
		e.CostTokens = &costTokens.Int64
	}
	return &e, nil
}

// HaltAgent implements runtime.RuntimeStore.
func (s *Store) HaltAgent(ctx context.Context, agentID string, reason string) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO ctrldot_halted_agents (agent_id, reason, halted_at) VALUES (?, ?, ?) ON CONFLICT (agent_id) DO UPDATE SET reason = excluded.reason, halted_at = excluded.halted_at`,
		agentID, reason, time.Now().UTC().Format(time.RFC3339),
	)
	if err != nil {
		return fmt.Errorf("halt agent: %w", err)
	}
	return nil
}

// ResumeAgent implements runtime.RuntimeStore.
func (s *Store) ResumeAgent(ctx context.Context, agentID string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM ctrldot_halted_agents WHERE agent_id = ?`, agentID)
	if err != nil {
		return fmt.Errorf("resume agent: %w", err)
	}
	return nil
}

// GetPanicState implements runtime.RuntimeStore.
func (s *Store) GetPanicState(ctx context.Context) (*domain.PanicState, error) {
	var enabled int
	var enabledAt, expiresAt, reason sql.NullString
	var ttlSeconds sql.NullInt64
	err := s.db.QueryRowContext(ctx,
		`SELECT enabled, enabled_at, expires_at, ttl_seconds, reason FROM ctrldot_panic_state WHERE id = 1`).Scan(
		&enabled, &enabledAt, &expiresAt, &ttlSeconds, &reason)
	if err != nil {
		if err == sql.ErrNoRows {
			return &domain.PanicState{Enabled: false}, nil
		}
		return nil, fmt.Errorf("get panic state: %w", err)
	}
	st := &domain.PanicState{Enabled: enabled != 0, TTLSeconds: int(ttlSeconds.Int64)}
	if enabledAt.Valid {
		t, _ := time.Parse(time.RFC3339, enabledAt.String)
		st.EnabledAt = t
	}
	if expiresAt.Valid {
		t, _ := time.Parse(time.RFC3339, expiresAt.String)
		st.ExpiresAt = &t
	}
	if reason.Valid {
		st.Reason = reason.String
	}
	return st, nil
}

// SetPanicState implements runtime.RuntimeStore.
func (s *Store) SetPanicState(ctx context.Context, state domain.PanicState) error {
	enabled := 0
	if state.Enabled {
		enabled = 1
	}
	enabledAt := state.EnabledAt.Format(time.RFC3339)
	var expiresAt string
	if state.ExpiresAt != nil {
		expiresAt = state.ExpiresAt.Format(time.RFC3339)
	}
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO ctrldot_panic_state (id, enabled, enabled_at, expires_at, ttl_seconds, reason) VALUES (1, ?, ?, ?, ?, ?)
		 ON CONFLICT(id) DO UPDATE SET enabled = excluded.enabled, enabled_at = excluded.enabled_at, expires_at = excluded.expires_at, ttl_seconds = excluded.ttl_seconds, reason = excluded.reason`,
		enabled, enabledAt, nullString(expiresAt), state.TTLSeconds, state.Reason)
	if err != nil {
		return fmt.Errorf("set panic state: %w", err)
	}
	return nil
}

func nullString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

var _ runtime.RuntimeStore = (*Store)(nil)

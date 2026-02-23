-- Ctrl Dot runtime schema for SQLite (no Kernel dependency)
-- Use PRAGMA journal_mode=WAL; and busy_timeout in application on open.

CREATE TABLE IF NOT EXISTS schema_version (
  version INTEGER NOT NULL PRIMARY KEY
);
INSERT OR IGNORE INTO schema_version (version) VALUES (1);

-- Agents
CREATE TABLE IF NOT EXISTS ctrldot_agents (
  agent_id TEXT PRIMARY KEY,
  display_name TEXT,
  created_at TEXT NOT NULL,
  default_mode TEXT NOT NULL DEFAULT 'normal'
);
CREATE INDEX IF NOT EXISTS idx_ctrldot_agents_created_at ON ctrldot_agents(created_at);

-- Sessions
CREATE TABLE IF NOT EXISTS ctrldot_sessions (
  session_id TEXT PRIMARY KEY,
  agent_id TEXT NOT NULL,
  started_at TEXT NOT NULL,
  ended_at TEXT,
  metadata_json TEXT,
  FOREIGN KEY (agent_id) REFERENCES ctrldot_agents(agent_id)
);
CREATE INDEX IF NOT EXISTS idx_ctrldot_sessions_agent ON ctrldot_sessions(agent_id);
CREATE INDEX IF NOT EXISTS idx_ctrldot_sessions_started_at ON ctrldot_sessions(started_at);

-- Events (append-only; no op_seq)
CREATE TABLE IF NOT EXISTS ctrldot_events (
  event_id TEXT PRIMARY KEY,
  event_type TEXT NOT NULL,
  agent_id TEXT,
  session_id TEXT,
  severity TEXT NOT NULL DEFAULT 'info',
  payload_json TEXT,
  action_hash TEXT,
  cost_gbp REAL,
  cost_tokens INTEGER,
  created_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_ctrldot_events_created_at ON ctrldot_events(created_at);
CREATE INDEX IF NOT EXISTS idx_ctrldot_events_agent_created ON ctrldot_events(agent_id, created_at);
CREATE INDEX IF NOT EXISTS idx_ctrldot_events_action_hash ON ctrldot_events(action_hash, created_at);
CREATE INDEX IF NOT EXISTS idx_ctrldot_events_type ON ctrldot_events(event_type);

-- Limits state (window_start stored as unix ms integer for daily window)
CREATE TABLE IF NOT EXISTS ctrldot_limits_state (
  agent_id TEXT NOT NULL,
  window_start INTEGER NOT NULL,
  window_type TEXT NOT NULL DEFAULT 'daily',
  budget_spent_gbp REAL NOT NULL DEFAULT 0,
  budget_spent_tokens INTEGER NOT NULL DEFAULT 0,
  action_count INTEGER NOT NULL DEFAULT 0,
  PRIMARY KEY (agent_id, window_start, window_type),
  FOREIGN KEY (agent_id) REFERENCES ctrldot_agents(agent_id)
);
CREATE INDEX IF NOT EXISTS idx_ctrldot_limits_state_agent_window ON ctrldot_limits_state(agent_id, window_start);

-- Halted agents
CREATE TABLE IF NOT EXISTS ctrldot_halted_agents (
  agent_id TEXT PRIMARY KEY,
  reason TEXT NOT NULL,
  halted_at TEXT NOT NULL,
  FOREIGN KEY (agent_id) REFERENCES ctrldot_agents(agent_id)
);
CREATE INDEX IF NOT EXISTS idx_ctrldot_halted_agents_halted_at ON ctrldot_halted_agents(halted_at);

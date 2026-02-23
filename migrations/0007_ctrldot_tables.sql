-- Ctrl Dot v0.1 - Initial schema
-- Extends kernel with Ctrl Dot domain tables

BEGIN;

-- Agents table
CREATE TABLE IF NOT EXISTS ctrldot_agents (
  agent_id TEXT PRIMARY KEY,
  display_name TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  default_mode TEXT NOT NULL DEFAULT 'normal'
);
CREATE INDEX IF NOT EXISTS idx_ctrldot_agents_created_at ON ctrldot_agents(created_at);

-- Sessions table
CREATE TABLE IF NOT EXISTS ctrldot_sessions (
  session_id TEXT PRIMARY KEY,
  agent_id TEXT NOT NULL REFERENCES ctrldot_agents(agent_id),
  started_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  ended_at TIMESTAMPTZ,
  metadata_json JSONB
);
CREATE INDEX IF NOT EXISTS idx_ctrldot_sessions_agent ON ctrldot_sessions(agent_id);
CREATE INDEX IF NOT EXISTS idx_ctrldot_sessions_started_at ON ctrldot_sessions(started_at);

-- Events table (append-only, maps to kernel operations)
-- We'll use kernel's operations table for events, but add a view/helper table for Ctrl Dot event metadata
CREATE TABLE IF NOT EXISTS ctrldot_events (
  event_id TEXT PRIMARY KEY,
  op_seq BIGINT NOT NULL REFERENCES operations(seq),
  event_type TEXT NOT NULL,
  agent_id TEXT REFERENCES ctrldot_agents(agent_id),
  session_id TEXT REFERENCES ctrldot_sessions(session_id),
  severity TEXT NOT NULL DEFAULT 'info',
  payload_json JSONB,
  action_hash TEXT,
  cost_gbp DOUBLE PRECISION,
  cost_tokens BIGINT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_ctrldot_events_created_at ON ctrldot_events(created_at);
CREATE INDEX IF NOT EXISTS idx_ctrldot_events_agent_created ON ctrldot_events(agent_id, created_at);
CREATE INDEX IF NOT EXISTS idx_ctrldot_events_action_hash ON ctrldot_events(action_hash, created_at);
CREATE INDEX IF NOT EXISTS idx_ctrldot_events_type ON ctrldot_events(event_type);

-- Limits state table (budget tracking, token buckets, etc.)
CREATE TABLE IF NOT EXISTS ctrldot_limits_state (
  agent_id TEXT NOT NULL REFERENCES ctrldot_agents(agent_id),
  window_start TIMESTAMPTZ NOT NULL,
  window_type TEXT NOT NULL DEFAULT 'daily', -- daily, hourly, etc.
  budget_spent_gbp DOUBLE PRECISION NOT NULL DEFAULT 0,
  budget_spent_tokens BIGINT NOT NULL DEFAULT 0,
  action_count INTEGER NOT NULL DEFAULT 0,
  PRIMARY KEY (agent_id, window_start, window_type)
);
CREATE INDEX IF NOT EXISTS idx_ctrldot_limits_state_agent_window ON ctrldot_limits_state(agent_id, window_start);

-- Halted agents table
CREATE TABLE IF NOT EXISTS ctrldot_halted_agents (
  agent_id TEXT PRIMARY KEY REFERENCES ctrldot_agents(agent_id),
  reason TEXT NOT NULL,
  halted_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_ctrldot_halted_agents_halted_at ON ctrldot_halted_agents(halted_at);

COMMIT;

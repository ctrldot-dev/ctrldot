-- Panic mode state (single row, id=1)
CREATE TABLE IF NOT EXISTS ctrldot_panic_state (
  id INTEGER PRIMARY KEY CHECK (id = 1),
  enabled INTEGER NOT NULL DEFAULT 0,
  enabled_at TEXT,
  expires_at TEXT,
  ttl_seconds INTEGER NOT NULL DEFAULT 0,
  reason TEXT
);
INSERT OR IGNORE INTO ctrldot_panic_state (id, enabled, ttl_seconds) VALUES (1, 0, 0);

-- Ctrl Dot panic mode state (single row)
BEGIN;

CREATE TABLE IF NOT EXISTS ctrldot_panic_state (
  id INTEGER PRIMARY KEY CHECK (id = 1),
  enabled BOOLEAN NOT NULL DEFAULT false,
  enabled_at TIMESTAMPTZ,
  expires_at TIMESTAMPTZ,
  ttl_seconds INTEGER NOT NULL DEFAULT 0,
  reason TEXT
);

INSERT INTO ctrldot_panic_state (id, enabled, ttl_seconds)
VALUES (1, false, 0)
ON CONFLICT (id) DO NOTHING;

COMMIT;

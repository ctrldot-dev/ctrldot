-- Futurematic Kernel v0.1 - Initial schema
-- Postgres 14+ recommended

BEGIN;

-- Namespaces
CREATE TABLE IF NOT EXISTS namespaces (
  namespace_id TEXT PRIMARY KEY,
  name         TEXT NOT NULL,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Plans (single-use in v0.1)
CREATE TABLE IF NOT EXISTS plans (
  plan_id       TEXT PRIMARY KEY,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  actor_id      TEXT NOT NULL,
  namespace_id  TEXT,
  asof_seq      BIGINT,
  policy_hash   TEXT NOT NULL,
  plan_hash     TEXT NOT NULL,
  class         INT NOT NULL,
  intents_json  JSONB NOT NULL,
  expanded_json JSONB NOT NULL,
  policy_report_json JSONB NOT NULL,
  applied_op_id TEXT,
  applied_seq   BIGINT
);
CREATE INDEX IF NOT EXISTS idx_plans_actor ON plans(actor_id);
CREATE INDEX IF NOT EXISTS idx_plans_applied ON plans(applied_seq);

-- Operations log (append-only)
CREATE TABLE IF NOT EXISTS operations (
  seq          BIGSERIAL PRIMARY KEY,
  op_id        TEXT NOT NULL UNIQUE,
  occurred_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  actor_id     TEXT NOT NULL,
  capabilities JSONB NOT NULL,
  plan_id      TEXT NOT NULL,
  plan_hash    TEXT NOT NULL,
  class        INT NOT NULL,
  changes_json JSONB NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_operations_occurred_at ON operations(occurred_at);
CREATE INDEX IF NOT EXISTS idx_operations_plan_id ON operations(plan_id);

-- Nodes (projection)
CREATE TABLE IF NOT EXISTS nodes (
  node_id     TEXT PRIMARY KEY,
  title       TEXT NOT NULL,
  meta_json   JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_seq BIGINT NOT NULL,
  retired_seq BIGINT,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_nodes_visible ON nodes(created_seq, retired_seq);

-- Role assignments (projection)
CREATE TABLE IF NOT EXISTS role_assignments (
  role_assignment_id TEXT PRIMARY KEY,
  node_id            TEXT NOT NULL REFERENCES nodes(node_id),
  namespace_id       TEXT NOT NULL REFERENCES namespaces(namespace_id),
  role               TEXT NOT NULL,
  meta_json          JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_seq        BIGINT NOT NULL,
  retired_seq        BIGINT,
  created_at         TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_roles_node_ns ON role_assignments(namespace_id, node_id, role);
CREATE INDEX IF NOT EXISTS idx_roles_visible ON role_assignments(created_seq, retired_seq);

-- Links (projection)
CREATE TABLE IF NOT EXISTS links (
  link_id      TEXT PRIMARY KEY,
  from_node_id TEXT NOT NULL REFERENCES nodes(node_id),
  to_node_id   TEXT NOT NULL REFERENCES nodes(node_id),
  type         TEXT NOT NULL,
  namespace_id TEXT REFERENCES namespaces(namespace_id),
  meta_json    JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_seq  BIGINT NOT NULL,
  retired_seq  BIGINT,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_links_from ON links(namespace_id, type, from_node_id);
CREATE INDEX IF NOT EXISTS idx_links_to   ON links(namespace_id, type, to_node_id);
CREATE INDEX IF NOT EXISTS idx_links_visible ON links(created_seq, retired_seq);

-- Materials (projection)
CREATE TABLE IF NOT EXISTS materials (
  material_id TEXT PRIMARY KEY,
  node_id     TEXT NOT NULL REFERENCES nodes(node_id),
  content_ref TEXT NOT NULL,
  media_type  TEXT NOT NULL,
  byte_size   BIGINT NOT NULL,
  hash        TEXT,
  meta_json   JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_seq BIGINT NOT NULL,
  retired_seq BIGINT,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_materials_node ON materials(node_id);
CREATE INDEX IF NOT EXISTS idx_materials_visible ON materials(created_seq, retired_seq);

-- Policy sets (active policy per namespace)
CREATE TABLE IF NOT EXISTS policy_sets (
  policy_set_id TEXT PRIMARY KEY,
  namespace_id  TEXT NOT NULL REFERENCES namespaces(namespace_id),
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  policy_yaml   TEXT NOT NULL,
  policy_hash   TEXT NOT NULL,
  is_active     BOOLEAN NOT NULL DEFAULT FALSE,
  created_seq   BIGINT NOT NULL,
  retired_seq   BIGINT
);
CREATE INDEX IF NOT EXISTS idx_policy_active ON policy_sets(namespace_id, is_active) WHERE is_active = TRUE;

COMMIT;

-- Proposal sets and resolutions for Decision Ledger (Focus → Options → Resolution → Recorded)

BEGIN;

-- Proposal sets (AI-generated options; stored when Dot returns options)
CREATE TABLE IF NOT EXISTS proposal_sets (
  proposal_set_id   TEXT PRIMARY KEY,
  question          TEXT NOT NULL,
  focused_node_id   TEXT NOT NULL,
  namespace_id      TEXT NOT NULL REFERENCES namespaces(namespace_id),
  options_json      JSONB NOT NULL,
  recommended_option_id TEXT,
  confidence        DOUBLE PRECISION,
  raw_model_output  TEXT,
  created_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_proposal_sets_node ON proposal_sets(focused_node_id);
CREATE INDEX IF NOT EXISTS idx_proposal_sets_namespace ON proposal_sets(namespace_id);

-- Resolutions (human or delegated choice from a proposal set)
CREATE TABLE IF NOT EXISTS resolutions (
  resolution_id     TEXT PRIMARY KEY,
  proposal_set_id   TEXT NOT NULL REFERENCES proposal_sets(proposal_set_id),
  option_id        TEXT NOT NULL,
  resolver         TEXT NOT NULL,
  rationale        TEXT,
  constraints      TEXT,
  assumptions_snap TEXT,
  created_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_resolutions_proposal ON resolutions(proposal_set_id);

COMMIT;

-- Wipe only Fin ledger projection data so Fin namespaces can be re-seeded.
-- Keeps Product ledger, operations, plans, namespaces, and policy_sets.

BEGIN;

-- Nodes that have any role in a FinLedger namespace
CREATE TEMP TABLE fin_node_ids AS
  SELECT DISTINCT node_id
  FROM role_assignments
  WHERE namespace_id LIKE 'FinLedger%';

DELETE FROM materials WHERE node_id IN (SELECT node_id FROM fin_node_ids);
DELETE FROM links WHERE namespace_id LIKE 'FinLedger%';
DELETE FROM role_assignments WHERE namespace_id LIKE 'FinLedger%';
DELETE FROM nodes WHERE node_id IN (SELECT node_id FROM fin_node_ids);

COMMIT;

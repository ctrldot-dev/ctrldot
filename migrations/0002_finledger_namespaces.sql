-- FinLedger domain namespaces (for policy set prefix and seed graphs)
-- Do not modify Product Ledger namespaces.

BEGIN;

INSERT INTO namespaces (namespace_id, name) VALUES
  ('FinLedger', 'Financial Ledger'),
  ('FinLedger:/Kesteron/Treasury', 'Kesteron Treasury'),
  ('FinLedger:/Kesteron/StablecoinReserves', 'Kesteron Stablecoin Reserves')
ON CONFLICT (namespace_id) DO NOTHING;

COMMIT;

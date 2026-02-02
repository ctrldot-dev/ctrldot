-- Product Ledger namespaces: one namespace per demo ledger (FieldServe, AssetLink).
-- Aligns with Fin Ledger structure (one namespace per root). Do not modify Fin Ledger.

BEGIN;

INSERT INTO namespaces (namespace_id, name) VALUES
  ('ProductLedger', 'Product Ledger'),
  ('ProductLedger:/Kesteron/FieldServe', 'Kesteron FieldServe'),
  ('ProductLedger:/Kesteron/AssetLink', 'Kesteron AssetLink')
ON CONFLICT (namespace_id) DO NOTHING;

COMMIT;

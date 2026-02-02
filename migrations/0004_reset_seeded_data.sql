-- Reset seeded ledger data (operations, plans, projections). Keeps namespaces and policy_sets.
-- Run before re-seeding with the new 4 namespaces (FieldServe, AssetLink, Treasury, StablecoinReserves).

BEGIN;

-- Projections (depend on nodes): drop first or truncate nodes with CASCADE
TRUNCATE materials, links, role_assignments, nodes CASCADE;
TRUNCATE plans;
TRUNCATE operations RESTART IDENTITY;

COMMIT;

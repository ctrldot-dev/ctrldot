# Seeding the Product Ledger (split namespaces)

The Product Ledger is split into two namespaces (like Fin Ledger), so the left nav shows one row per ledger:

- **ProductLedger:/Kesteron/AssetLink**
- **ProductLedger:/Kesteron/FieldServe**
- **FinLedger:/Kesteron/Treasury**
- **FinLedger:/Kesteron/StablecoinReserves**

## Prerequisites

- Kernel running (e.g. `DB_URL=... PORT=8080 go run ./cmd/kernel`)
- Migrations applied (including `0003_productledger_namespaces.sql`)

## Steps

1. **Apply migration 0003** (if not already applied):

   ```bash
   psql "$DB_URL" -f migrations/0003_productledger_namespaces.sql
   ```

2. **Register the Product Ledger policy set** (for ProductLedger:*):

   ```bash
   go run ./cmd/bootstrap_productledger
   ```

3. **Seed the two Product Ledger namespaces** (and optionally Fin Ledger):

   ```bash
   go run ./cmd/seed_finledger \
     kesteron-fieldserve-product-seed.yaml \
     kesteron-assetlink-product-seed.yaml
   ```

   To also (re-)seed Fin Ledger:

   ```bash
   go run ./cmd/seed_finledger \
     kesteron-fieldserve-product-seed.yaml \
     kesteron-assetlink-product-seed.yaml \
     kesteron-treasury-finledger-seed.yaml \
     kesteron-stablecoinreserves-finledger-seed.yaml
   ```

## Acceptance

- Left nav shows 4 rows: AssetLink, FieldServe, Treasury, StablecoinReserves.
- Each row opens a single-root tree; refresh keeps the same structure.

## Note on old ProductLedger:/Kesteron

The old single namespace `ProductLedger:/Kesteron` is not removed by the migration. If you want exactly 4 rows, use a fresh DB or avoid seeding the old namespace; existing data in `ProductLedger:/Kesteron` will still appear in the nav if it has role assignments.

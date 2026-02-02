# AssetLink Product Profile

## Overview

AssetLink is Kesteron's single source of truth for asset state and semantics. It provides a stable, versioned data model that downstream systems depend on for operational decisions.

## Responsibility and Scope

AssetLink is responsible for:
- Maintaining canonical asset state across all customer environments
- Providing a stable, versioned schema for asset data
- Serving as the authoritative source for asset semantics
- Tracking asset lifecycle and state transitions

## Why Asset Semantics Matter

Asset semantics define:
- What data fields mean in operational context
- How asset states transition over time
- What dependencies exist between assets
- How downstream systems should interpret asset data

Changes to asset semantics can have cascading effects:
- FieldServe depends on AssetLink schema for service orchestration
- Customer portals display asset information based on AssetLink models
- Analytics systems aggregate data using AssetLink's semantic model

## Impact of Changes

When AssetLink schema changes:
- Downstream systems must be notified and updated
- Historical data interpretation may change
- Service logic may need adjustment
- Customer-facing features may be affected

## Known Consumers

AssetLink is consumed by:
- **FieldServe**: Uses asset state for service orchestration decisions
- **Customer Portals**: Displays asset information to end users
- **Analytics Platforms**: Aggregates asset data for reporting
- **Integration Services**: Exposes asset data to external systems

## Semantic Stability

AssetLink prioritizes semantic stability:
- Schema changes are versioned and documented
- Breaking changes require explicit coordination with consumers
- Historical data remains interpretable through versioning
- Changes impacting FieldServe are explicitly recorded in the Product Ledger

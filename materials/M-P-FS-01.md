# FieldServe Product Profile

## Overview

FieldServe is Kesteron's customer-critical service orchestration platform. It coordinates field operations, asset management, and service delivery across multiple customer environments.

## What FieldServe Does

FieldServe provides:
- Real-time orchestration of field service operations
- Integration with asset management systems (primarily AssetLink)
- Service delivery tracking and reporting
- Customer-facing service status APIs

## What FieldServe Does Not Do

FieldServe does not:
- Store asset state directly (relies on AssetLink)
- Manage customer billing or contracts
- Handle low-level device communication protocols

## Primary Users and Stakeholders

- **Field Operations Teams**: Primary users who coordinate service delivery
- **Customer Support**: Use FieldServe to track service status and respond to inquiries
- **Product Management**: Monitor service quality and plan improvements
- **Engineering Teams**: Maintain and enhance the platform

## Operational Context

FieldServe operates in a customer-critical environment where:
- Service disruptions directly impact customer operations
- Changes must be traceable and explainable
- System dependencies (especially on AssetLink) must be explicit
- Post-deployment explainability is essential for incident response

## Key Outcomes

FieldServe is accountable for:
1. **Service Reliability**: Minimizing unplanned service disruptions
2. **Operational Transparency**: Ensuring all changes are explainable after deployment
3. **Dependency Management**: Maintaining clear visibility into system dependencies

## Integration with Kesteron Platform

FieldServe is a core component of the Kesteron platform, working closely with:
- **AssetLink**: Primary source of asset state and semantics
- **Customer Portals**: Provides service status and tracking
- **Analytics Systems**: Feeds operational metrics and events

Changes to FieldServe's dependencies (especially AssetLink schema changes) must be explicitly recorded to maintain operational coherence.

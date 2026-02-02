#!/usr/bin/env bash
set -euo pipefail

# Product Ledger split: run once per namespace. Do not use the old single namespace.
#   NS="ProductLedger:/Kesteron/FieldServe" ./seed-kesteron-materials.sh
#   NS="ProductLedger:/Kesteron/AssetLink"  ./seed-kesteron-materials.sh
NS="${NS:-}"

if [[ -z "$NS" ]] || [[ "$NS" == "ProductLedger:/Kesteron" ]]; then
  echo "ERROR: Use a split product namespace. Example:" >&2
  echo "  NS=\"ProductLedger:/Kesteron/FieldServe\" $0" >&2
  echo "  NS=\"ProductLedger:/Kesteron/AssetLink\"  $0" >&2
  exit 1
fi

die() {
  echo "ERROR: $*" >&2
  exit 1
}

find_node_by_title() {
  local title="$1"
  local node_id
  # Get history and search for CreateNode with matching title
  node_id=$(dot history "$NS" --json 2>/dev/null | \
    grep -o "\"title\":\"[^\"]*$title[^\"]*\"" | \
    head -1 | \
    grep -B 20 "CreateNode" | \
    grep -o "node:[a-f0-9-]\{36\}" | \
    head -1) || true
  
  if [[ -z "$node_id" ]]; then
    # Try alternative: use dot show with expand to find nodes
    # This is a fallback - may not work if there are many nodes
    echo "" >&2
    return 1
  fi
  echo "$node_id"
}

create_material_from_file() {
  local node_id="${1:-}"
  local material_id="${2:-}"
  local title="${3:-}"
  local material_type="${4:-}"
  local content_file="${5:-}"
  
  [[ -n "$node_id" && -n "$material_id" && -n "$title" && -n "$material_type" && -n "$content_file" ]] || {
    die "create_material_from_file: missing required parameter"
  }

  echo "Creating material: $material_id ($title) for node $node_id" >&2

  # Read content from file
  local content
  content="$(cat "$content_file")"
  
  # Use absolute path or relative path that Web Dot can serve
  local full_content_ref="/materials/${material_id}.md"
  
  # Calculate byte size
  local byte_size
  byte_size=$(wc -c < "$content_file" | tr -d ' ')

  # Create material via plan/apply
  local actor_id
  actor_id="$(dot config get actor_id 2>/dev/null || echo "system:seed")"
  local server
  server="$(dot config get server 2>/dev/null || echo "http://localhost:8080")"
  
  local plan_json
  plan_json=$(cat <<EOF
{
  "actor_id": "$actor_id",
  "capabilities": ["read", "write:additive"],
  "namespace_id": "$NS",
  "intents": [
    {
      "kind": "CreateMaterial",
      "namespace_id": "$NS",
      "payload": {
        "material_id": "$material_id",
        "node_id": "$node_id",
        "content_ref": "$full_content_ref",
        "media_type": "text/markdown",
        "byte_size": $byte_size,
        "meta": {
          "title": "$title",
          "type": "$material_type",
          "category": "$(get_category_for_type "$material_type")"
        }
      }
    }
  ]
}
EOF
)

  # Create plan
  local plan_response
  plan_response=$(curl -s -X POST "$server/v1/plan" \
    -H "Content-Type: application/json" \
    -d "$plan_json") || die "Failed to create plan for material $material_id"

  # Check if material already exists (plan might succeed but material exists)
  if echo "$plan_response" | grep -q "already exists\|conflict\|duplicate" 2>/dev/null; then
    echo "  Material $material_id already exists, skipping" >&2
    return 0
  fi

  local plan_id
  plan_id=$(echo "$plan_response" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
  [[ -n "$plan_id" ]] || die "Could not extract plan ID for material $material_id"

  local plan_hash
  plan_hash=$(echo "$plan_response" | grep -o '"hash":"[^"]*"' | head -1 | cut -d'"' -f4)
  [[ -n "$plan_hash" ]] || die "Could not extract plan hash for material $material_id"

  # Apply plan
  local apply_json
  apply_json=$(cat <<EOF
{
  "actor_id": "$actor_id",
  "capabilities": ["read", "write:additive"],
  "plan_id": "$plan_id",
  "plan_hash": "$plan_hash"
}
EOF
)

  curl -s -X POST "$server/v1/apply" \
    -H "Content-Type: application/json" \
    -d "$apply_json" >/dev/null || die "Failed to apply plan for material $material_id"
}

get_category_for_type() {
  local type="$1"
  case "$type" in
    Profile) echo "Profiles" ;;
    Rationale) echo "Rationales" ;;
    JTBD) echo "Jobs to be Done" ;;
    Evidence) echo "Evidence" ;;
    *) echo "Notes" ;;
  esac
}

create_material() {
  local node_id="$1"
  local material_id="$2"
  local title="$3"
  local material_type="$4"
  # Content is passed via stdin (heredoc)

  # Create content file
  local content_dir="materials"
  mkdir -p "$content_dir"
  local content_file="$content_dir/${material_id}.md"
  cat > "$content_file"

  # Call the file-based function
  create_material_from_file "$node_id" "$material_id" "$title" "$material_type" "$content_file"
}

echo "Using namespace: $NS" >&2
dot use "$NS" || die "Failed to use namespace $NS"

echo "Finding existing nodes..." >&2

# Find nodes by searching history for their titles (using flexible matching)
find_node_by_pattern() {
  local pattern="$1"
  # Get JSON history, parse it to find CreateNode operations with matching title
  dot history "$NS" --json 2>/dev/null | \
    python3 -c "
import json
import sys
import re
try:
    data = json.load(sys.stdin)
    pattern = sys.argv[1]
    for op in data.get('result', []):
        for change in op.get('changes', []):
            if change.get('kind') == 'CreateNode':
                payload = change.get('payload', {})
                title = payload.get('title', '')
                node_id = payload.get('node_id', '')
                if re.search(pattern, title, re.IGNORECASE):
                    print(node_id)
                    sys.exit(0)
except:
    pass
" "$pattern" 2>/dev/null || echo ""
}

# Find nodes that exist in this namespace (split: FieldServe has P-FS/FS-*/SO-1; AssetLink has P-AL/SO-1).
# FS-D1 and E-1 are not in current seed YAMLs; we skip materials for them if not found.
PFS=$(find_node_by_pattern "P-FS FieldServe") || PFS=""
PAL=$(find_node_by_pattern "P-AL AssetLink") || PAL=""
FSG1=$(find_node_by_pattern "FS-G1.*unplanned.*disruption") || FSG1=""
FSG2=$(find_node_by_pattern "FS-G2.*explainable.*deployment") || FSG2=""
FSJ1=$(find_node_by_pattern "FS-J1.*operational impact.*before deployment") || FSJ1=""
FSJ2=$(find_node_by_pattern "FS-J2.*Reconstruct.*change.*made") || FSJ2=""
FSD1=$(find_node_by_pattern "FS-D1.*AssetLink.*schema.*changes") || FSD1=""
E1=$(find_node_by_pattern "E-1.*Incident.*review") || E1=""
SO1=$(find_node_by_pattern "SO-1.*reliability.*explainability") || SO1=""

# Require at least one node so we don't silently create nothing (wrong namespace or seed not run)
if [[ -z "$PFS" && -z "$PAL" && -z "$FSG1" && -z "$FSG2" && -z "$FSJ1" && -z "$FSJ2" && -z "$SO1" ]]; then
  die "No nodes found in namespace $NS. Run the product ledger seed for this namespace first (e.g. kesteron-fieldserve-product-seed.yaml for FieldServe, kesteron-assetlink-product-seed.yaml for AssetLink)."
fi

echo "Found nodes in $NS:" >&2
[[ -n "$PFS" ]] && echo "  P-FS: $PFS" >&2
[[ -n "$PAL" ]] && echo "  P-AL: $PAL" >&2
[[ -n "$FSG1" ]] && echo "  FS-G1: $FSG1" >&2
[[ -n "$FSG2" ]] && echo "  FS-G2: $FSG2" >&2
[[ -n "$FSJ1" ]] && echo "  FS-J1: $FSJ1" >&2
[[ -n "$FSJ2" ]] && echo "  FS-J2: $FSJ2" >&2
[[ -n "$FSD1" ]] && echo "  FS-D1: $FSD1" >&2
[[ -n "$E1" ]] && echo "  E-1: $E1" >&2
[[ -n "$SO1" ]] && echo "  SO-1: $SO1" >&2
echo "" >&2

# Materials: create only for nodes found in this namespace (skip with message if node missing)
echo "Creating materials..." >&2

# M-P-FS-01: FieldServe Product Profile (FieldServe namespace only)
if [[ -n "$PFS" ]]; then
create_material "$PFS" "M-P-FS-01" "FieldServe Product Profile" "Profile" <<'EOFMARKDOWN'
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
EOFMARKDOWN
else
  echo "  Skipping M-P-FS-01 (P-FS not in this namespace)" >&2
fi

# M-P-AL-01: AssetLink Product Profile (AssetLink namespace only)
if [[ -n "$PAL" ]]; then
create_material "$PAL" "M-P-AL-01" "AssetLink Product Profile" "Profile" <<'EOFMARKDOWN'
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
EOFMARKDOWN
else
  echo "  Skipping M-P-AL-01 (P-AL not in this namespace)" >&2
fi

# M-G-FS-01: Goal Rationale - Reduce Unplanned Service Disruption (FieldServe only)
if [[ -n "$FSG1" ]]; then
create_material "$FSG1" "M-G-FS-01" "Goal Rationale: Reduce Unplanned Service Disruption" "Rationale" <<'EOFMARKDOWN'
# Goal Rationale: Reduce Unplanned Service Disruption

## What "Unplanned Disruption" Means

Unplanned service disruption occurs when:
- FieldServe becomes unavailable or degraded without prior notice
- Service orchestration fails due to system errors
- Customer operations are impacted by unexpected system behavior
- Dependencies (especially AssetLink) fail or change unexpectedly

## Business and Customer Impact

### Customer Impact
- Service delivery delays or failures
- Loss of visibility into field operations
- Inability to track service status
- Customer trust and satisfaction degradation

### Business Impact
- Increased support burden
- Potential SLA violations
- Reputational damage
- Operational inefficiency

## Why This Goal is Prioritized Now

This goal is prioritized because:
1. **Customer-Critical Operations**: FieldServe directly supports customer operations
2. **Dependency Complexity**: Growing dependency on AssetLink requires better coordination
3. **Incident History**: Recent incidents highlighted gaps in change management
4. **Scalability Concerns**: As the platform grows, unplanned disruptions become more costly

## How Success is Assessed

Success indicators include:
- **Quantitative**: Reduction in unplanned downtime incidents
- **Qualitative**: Improved incident response times
- **Operational**: Better visibility into system dependencies
- **Customer**: Fewer customer-reported service issues

## Relationship to Other Goals

This goal supports:
- **SO-1**: Improve reliability and explainability of customer-critical services
- **FS-G2**: Explainability after deployment helps prevent future disruptions
EOFMARKDOWN
else
  echo "  Skipping M-G-FS-01 (FS-G1 not in this namespace)" >&2
fi

# M-G-FS-02: Goal Rationale - Explainability After Deployment (FieldServe only)
if [[ -n "$FSG2" ]]; then
create_material "$FSG2" "M-G-FS-02" "Goal Rationale: Explainability After Deployment" "Rationale" <<'EOFMARKDOWN'
# Goal Rationale: Ensure Operational Changes are Explainable After Deployment

## Why Post-Hoc Explainability Matters

After a deployment or incident, teams need to answer:
- Why was this change made?
- What was the decision process?
- What evidence supported the decision?
- What trade-offs were considered?
- How does this relate to other changes?

## Typical Questions After Incidents

When issues arise, teams ask:
1. **Causality**: "Did this change cause the problem?"
2. **Context**: "What was happening when this was deployed?"
3. **Rationale**: "Why was this approach chosen over alternatives?"
4. **Dependencies**: "What other systems were affected?"
5. **History**: "Has this pattern caused issues before?"

## Why Logs and Metrics Alone Are Insufficient

Traditional observability tools provide:
- **What happened**: Logs show events and errors
- **When it happened**: Timestamps and sequences
- **How it happened**: Stack traces and error messages

But they don't capture:
- **Why it happened**: The reasoning behind decisions
- **What was considered**: Alternative approaches evaluated
- **What was known**: Context and assumptions at decision time
- **What was intended**: Expected outcomes and trade-offs

## Why Explanations Must Be Durable and Human-Readable

### Durability
- Explanations must survive team changes, tool migrations, and time
- They should be part of the system of record, not ephemeral chat logs
- They need to be queryable and linkable to related decisions

### Human-Readability
- Explanations must be understandable by humans, not just machines
- They should tell a story that makes sense months or years later
- They should connect to business context and customer impact

## Connection to Product Ledger

The Product Ledger provides:
- **Durable Records**: Decisions are part of the immutable ledger
- **Structured Relationships**: Links connect decisions to goals, evidence, and affected systems
- **Temporal Context**: History shows when decisions were made and how they evolved
- **Human-Readable Format**: Materials (like this one) provide narrative context

This goal directly supports the Product Ledger thesis: that explicit, durable decision records improve operational coherence and reduce future incidents.
EOFMARKDOWN
else
  echo "  Skipping M-G-FS-02 (FS-G2 not in this namespace)" >&2
fi

# M-J-FS-01: JTBD - Understand Operational Impact Before Deployment (FieldServe only)
if [[ -n "$FSJ1" ]]; then
create_material "$FSJ1" "M-J-FS-01" "JTBD: Understand Operational Impact Before Deployment" "JTBD" <<'EOFMARKDOWN'
# Jobs to Be Done: Understand Operational Impact Before Deployment

## JTBD Framing

**As a** field operations engineer or product manager  
**I need to** understand the operational impact of system changes before deployment  
**So that** I can make informed decisions, prevent incidents, and coordinate with stakeholders

## Constraints

### Time Pressure
- Deployment windows are often tight
- Business requirements may demand rapid changes
- Customer needs may be urgent

### Incomplete Information
- Full impact analysis may be impractical
- Dependencies may not be fully known
- Edge cases may be discovered only in production

### Complexity
- Systems have many interdependencies
- Changes can have cascading effects
- Historical context may be needed to understand impact

## What Good Looks Like

Before deployment, teams have:
- **Clear Understanding**: Know which systems and customers will be affected
- **Dependency Map**: Understand upstream and downstream dependencies
- **Risk Assessment**: Identify potential failure modes and mitigation strategies
- **Stakeholder Alignment**: Coordinate with affected teams and customers
- **Rollback Plan**: Know how to revert if issues arise

## What Failure Looks Like

When this job is not done well:
- **Surprise Incidents**: Changes cause unexpected problems
- **Cascading Failures**: Issues propagate to dependent systems
- **Customer Impact**: Unplanned service disruptions
- **Blame Games**: Teams point fingers because context was lost
- **Reactive Firefighting**: Teams spend time fixing issues that could have been prevented

## How Product Ledger Helps

The Product Ledger supports this job by:
- **Explicit Dependencies**: Links show how systems depend on each other
- **Decision Records**: Past decisions provide context for current changes
- **Evidence Trail**: Historical evidence shows what worked and what didn't
- **Coherence View**: Teams can see how changes fit into the larger system
EOFMARKDOWN
else
  echo "  Skipping M-J-FS-01 (FS-J1 not in this namespace)" >&2
fi

# M-J-FS-02: JTBD - Reconstruct Why a Change Was Made (FieldServe only)
if [[ -n "$FSJ2" ]]; then
create_material "$FSJ2" "M-J-FS-02" "JTBD: Reconstruct Why a Change Was Made" "JTBD" <<'EOFMARKDOWN'
# Jobs to Be Done: Reconstruct Why a Change Was Made When Issues Arise

## JTBD Framing

**As a** incident responder, engineer, or product manager  
**I need to** reconstruct why a change was made when issues arise  
**So that** I can understand root causes, make informed fixes, and prevent similar issues

## Who Asks "Why Was This Changed?"

- **Incident Responders**: Need context to diagnose problems
- **On-Call Engineers**: Must understand system state during incidents
- **Product Managers**: Need to assess impact and communicate to stakeholders
- **Auditors**: Require traceability for compliance
- **New Team Members**: Need to understand system evolution

## When This Question Typically Arises

This question comes up:
- **During Incidents**: "Did this recent change cause the problem?"
- **During Post-Mortems**: "What was the reasoning behind this decision?"
- **During Audits**: "Can we show why this change was necessary?"
- **During Onboarding**: "How did the system get to this state?"
- **During Planning**: "What did we learn from past changes?"

## Common Failure Mode Today

### Slack Archaeology
Teams spend hours searching through:
- Slack channels for decision discussions
- Email threads for context
- Meeting notes for rationale
- Code comments for hints

### Guesswork
When context is lost:
- Teams make assumptions about why changes were made
- Decisions are re-debated without historical context
- Similar mistakes are repeated
- Blame is assigned without understanding

### Information Silos
Context is scattered across:
- Different communication tools
- Personal notes and memories
- Code repositories
- Documentation that may be outdated

## What Information Should Be Available Months Later

Months after a change, teams should be able to find:
- **Decision Rationale**: Why this approach was chosen
- **Alternatives Considered**: What other options were evaluated
- **Evidence**: What data or incidents supported the decision
- **Trade-offs**: What was accepted or sacrificed
- **Dependencies**: What other systems or decisions this relates to
- **Stakeholders**: Who was involved and what they said

## How Product Ledger Addresses This

The Product Ledger provides:
- **Durable Records**: Decisions are part of the immutable ledger
- **Structured Links**: Connect decisions to goals, evidence, and affected systems
- **Materials**: Human-readable explanations that survive time
- **Temporal Context**: History shows when and how decisions evolved
- **Queryability**: Can find related decisions and evidence quickly

This job is central to the Product Ledger's value proposition: making the "why" of system changes as durable and accessible as the "what" and "when."
EOFMARKDOWN
else
  echo "  Skipping M-J-FS-02 (FS-J2 not in this namespace)" >&2
fi

# M-D-FS-01: Decision Rationale - Recording AssetLink Schema Changes (FieldServe only; node FS-D1 not in current seed)
if [[ -n "$FSD1" ]]; then
create_material "$FSD1" "M-D-FS-01" "Decision Rationale: Recording AssetLink Schema Changes" "Rationale" <<'EOFMARKDOWN'
# Decision Rationale: Require Explicit Recording of AssetLink Schema Changes Impacting FieldServe

## Context Leading to the Decision

### The Incident
A recent production incident revealed a critical gap: when AssetLink changed its asset schema, FieldServe's service orchestration logic broke because it relied on fields that were renamed. The incident took hours to diagnose because:

1. The schema change was not explicitly communicated to FieldServe team
2. The relationship between AssetLink and FieldServe was not clearly documented
3. The rationale for the schema change was lost in Slack threads
4. FieldServe team had to reverse-engineer what changed and why

### The Pattern
This incident highlighted a broader pattern:
- AssetLink and FieldServe are tightly coupled but the coupling is implicit
- Schema changes in AssetLink can break FieldServe without warning
- There's no systematic way to track and coordinate these dependencies
- Historical context about why changes were made is often lost

## Options Considered

### Option 1: Do Nothing
**Pros:**
- No process overhead
- Teams continue current workflow

**Cons:**
- Incidents will continue to occur
- No improvement in coordination
- Context loss continues

### Option 2: Require Explicit Recording in Product Ledger
**Pros:**
- Creates durable record of changes and rationale
- Makes dependencies explicit and queryable
- Supports post-incident analysis
- Enables proactive coordination

**Cons:**
- Requires process change
- Initial setup overhead
- Teams must learn new tooling

### Option 3: Automated Dependency Detection
**Pros:**
- Could catch some issues automatically
- Less manual work

**Cons:**
- Cannot capture rationale or context
- May produce false positives
- Doesn't address human coordination needs

## Why This Decision Was Chosen

We chose Option 2 (Explicit Recording) because:
1. **Durability**: Product Ledger provides permanent, queryable records
2. **Context Preservation**: Materials can capture rationale and trade-offs
3. **Explicit Dependencies**: Links make relationships visible
4. **Human Coordination**: Supports the social process of change management
5. **Incident Prevention**: Makes it easier to coordinate changes proactively

## Trade-offs Accepted

- **Process Overhead**: Teams must record changes explicitly
- **Learning Curve**: Teams need to learn Product Ledger concepts
- **Not Fully Automated**: Still requires human judgment and coordination
- **Initial Investment**: Time to set up materials and links

## What Future Problems This Decision Prevents

This decision prevents:
- **Surprise Breakages**: FieldServe team will know about AssetLink changes
- **Lost Context**: Rationale for changes will be preserved
- **Coordination Failures**: Dependencies will be explicit
- **Repeated Mistakes**: Historical decisions will be queryable
- **Blame Games**: Clear record of who decided what and why

## Implementation

This decision is implemented by:
- Creating Decision node (FS-D1) linked to relevant Goals
- Linking Decision to Evidence (E-1) that supports it
- Creating Materials (like this one) that explain the rationale
- Establishing process for recording future AssetLink schema changes
- Linking AssetLink and FieldServe goals to show dependency

## Success Criteria

This decision will be successful if:
- Future AssetLink schema changes are recorded in Product Ledger
- FieldServe team is notified of changes before deployment
- Incidents related to schema changes decrease
- Post-incident analysis can quickly find relevant decisions and context
EOFMARKDOWN
else
  echo "  Skipping M-D-FS-01 (FS-D1 not in this namespace)" >&2
fi

# M-E-01: Incident Review Summary (Evidence node E-1 not in current seed)
if [[ -n "$E1" ]]; then
create_material "$E1" "M-E-01" "Incident Review Summary (AI-assisted, Human-approved)" "Evidence" <<'EOFMARKDOWN'
# Incident Review Summary

**Note**: This summary was AI-assisted but reviewed and approved by human incident responders.

## Incident Overview

**Date**: 2025-01-15  
**Duration**: 4 hours  
**Severity**: P1 (Customer Impact)  
**Affected Systems**: FieldServe, AssetLink  
**Root Cause**: AssetLink schema change broke FieldServe service orchestration logic

## What Went Wrong

### The Change
AssetLink deployed a schema change that renamed the `asset_status` field to `current_state` in their asset data model. This change was:
- Deployed during a scheduled maintenance window
- Documented in AssetLink's internal changelog
- Not explicitly communicated to FieldServe team

### The Failure
FieldServe's service orchestration logic relied on the `asset_status` field to make routing decisions. When the field was renamed:
- FieldServe could not read asset state
- Service orchestration decisions failed
- Customer service requests were not routed correctly
- Some services were delayed, others were routed incorrectly

### Detection
The issue was detected when:
- Customer support received reports of incorrect service routing
- Field operations teams noticed missing service assignments
- Monitoring showed increased error rates in FieldServe API

## What Information Was Missing at the Time

During incident response, teams discovered:
1. **No Explicit Dependency Record**: The relationship between AssetLink schema and FieldServe logic was not documented
2. **Lost Rationale**: The reason for renaming `asset_status` to `current_state` was in a Slack thread that was hard to find
3. **No Coordination Process**: There was no systematic way to notify FieldServe of AssetLink changes
4. **Historical Context Lost**: Previous similar incidents were not easily discoverable

## How This Incident Led to FS-D1

This incident directly led to Decision FS-D1: "Require explicit recording of AssetLink schema changes impacting FieldServe."

The incident review concluded:
- The technical fix (updating FieldServe to use `current_state`) was straightforward
- The process gap (lack of coordination and documentation) was the real problem
- A systematic approach to recording dependencies and changes was needed
- The Product Ledger could provide the durable, queryable record needed

## AI Assistance and Human Review

This summary was:
- **AI-Assisted**: Initial draft generated from incident logs, Slack threads, and monitoring data
- **Human-Reviewed**: Incident responders verified facts, added context, and corrected interpretations
- **Human-Approved**: Final version approved by incident commander and product managers

The AI assistance helped:
- Aggregate information from multiple sources quickly
- Identify patterns in logs and errors
- Draft initial narrative structure

Human review ensured:
- Accuracy of technical details
- Appropriate business context
- Correct attribution of decisions
- Alignment with team understanding

## Lessons Learned

1. **Explicit Dependencies Matter**: Implicit dependencies between systems are a source of risk
2. **Context Preservation**: Rationale for changes must be preserved in durable, queryable form
3. **Coordination Process**: Systematic processes for cross-team coordination are essential
4. **Historical Context**: Past incidents should be easily discoverable to prevent repeats

## Actions Taken

1. **Immediate**: Updated FieldServe to use new `current_state` field
2. **Short-term**: Established FS-D1 decision and recording process
3. **Long-term**: Building Product Ledger materials and links to support ongoing coordination
EOFMARKDOWN
else
  echo "  Skipping M-E-01 (E-1 not in this namespace)" >&2
fi

# M-ENT-01: Kesteron Enterprise Intent Overview (create only in FieldServe run; material_id is global)
if [[ -n "$SO1" ]] && [[ -n "$PFS" ]]; then
create_material "$SO1" "M-ENT-01" "Kesteron Enterprise Intent Overview" "Profile" <<'EOFMARKDOWN'
# Kesteron Enterprise Intent Overview

## Kesteron's High-Level Objectives

Kesteron operates customer-critical services that require:
- **Reliability**: Services must be available and performant
- **Explainability**: Changes and decisions must be traceable
- **Accountability**: Teams must be able to show why decisions were made
- **Coherence**: System changes must align with business objectives

## Why Traceability and Coherence Matter at Board Level

### Regulatory and Compliance
- External stakeholders require audit trails
- Compliance frameworks demand decision traceability
- Risk management requires understanding of system dependencies

### Business Continuity
- Customer trust depends on reliable, explainable services
- Incident response requires quick access to decision context
- Strategic planning needs visibility into system evolution

### Organizational Learning
- Past decisions inform future strategy
- Patterns in incidents reveal systemic issues
- Knowledge must survive team changes and tool migrations

## How Product, Intent, and Ledger Reinforce Each Other

### Product Layer
- **Products** (like FieldServe, AssetLink) deliver customer value
- **Goals** define what products should achieve
- **Jobs** describe user needs products address
- **Work Items** are concrete steps toward goals

### Intent Layer
- **Strategic Objectives** (SO) define enterprise direction
- **Assurance Obligations** (AO) ensure compliance and risk management
- **Transformation Themes** (TT) guide organizational change

### Ledger Layer
- **Operations** record what happened and when
- **Decisions** capture why changes were made
- **Evidence** provides support for decisions
- **Materials** preserve human-readable context

Together, these layers create:
- **Alignment**: Products support enterprise intent
- **Traceability**: Decisions link to goals and evidence
- **Coherence**: Changes align with strategic direction
- **Durability**: Context survives time and organizational change

## How AI is Used Responsibly Within This System

### AI Assistance, Human Oversight
- AI helps aggregate information and draft summaries
- Humans review, verify, and approve all AI-generated content
- Decision authority remains with humans
- AI is a tool, not a decision-maker

### Explicit Boundaries
- AI assistance is explicitly noted in materials (e.g., "AI-assisted, human-approved")
- Human review and approval are required for all critical content
- AI cannot create or modify ledger records without human intent

### Transparency
- The system makes it clear when AI was involved
- Human reviewers are identified
- Decision processes are documented
- Audit trails show human oversight

## The Product Ledger Thesis

The Product Ledger demonstrates that:
- **Explicit is Better than Implicit**: Dependencies and decisions should be recorded
- **Durable is Better than Ephemeral**: Context should survive tool and team changes
- **Structured is Better than Unstructured**: Links and relationships enable queryability
- **Human-Readable is Essential**: Materials provide narrative context machines cannot

This system enables Kesteron to:
- Maintain operational coherence as systems evolve
- Respond to incidents with full context
- Make strategic decisions with historical perspective
- Demonstrate accountability to stakeholders
EOFMARKDOWN
else
  echo "  Skipping M-ENT-01 (created in FieldServe run only; SO-1 or P-FS not in this namespace)" >&2
fi

echo "" >&2
echo "✅ Materials seed complete." >&2
echo "Namespace: $NS" >&2
echo "Materials created in: materials/" >&2

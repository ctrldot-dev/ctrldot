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

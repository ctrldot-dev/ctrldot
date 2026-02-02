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

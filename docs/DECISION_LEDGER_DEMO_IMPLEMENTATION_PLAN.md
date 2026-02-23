# Decision Ledger Demo — Implementation Plan

**Goal**: Demonstrate accountable decision-making with AI assistance: **AI proposes options → Human resolves → System records → Downstream work follows.**

**Definition of Done**:
- Dot can be focused on any node
- Options Talent returns a valid ProposalSet using an open/local model
- Kernel supports `resolve` and enforces "no apply without resolution" by default
- Ledger records ProposalSet + Resolution with links and attribution
- UI clearly shows the lifecycle: Focus → Options → Resolution → Recorded

---

## Prerequisites & Dependencies

| Component | Status | Notes |
|-----------|--------|-------|
| **Kernel** | Exists | `cmd/kernel`, `internal/kernel`, `internal/api` — Plan/Apply workflow |
| **web-dot** | Exists | React app with LedgerTreeNav, IntentTab, LedgerTab, MaterialsTab |
| **Centralised AI** | In use | [centralised-ai-tool](https://github.com/future-gareth/centralised-ai-tool) — `POST /ai/generate` (e.g. https://garethapi.com/central-ai) |
| **Product Ledger** | Exists | FieldServe, AssetLink namespaces, seeded nodes |

---

## Phase 1 — Focus Dot on Product Area

**Job 1.1** — Let a user focus Dot on a specific product area.

### WI 1.1.1 — Add "Focus Dot on this" action to Product Tree nodes
- **Where**: `web-dot/src/components/LedgerTreeNav.tsx`
- **What**: Add a button or context-menu item on each tree node (e.g. icon next to node title)
- **Action**: On click → dispatch `DotContextChanged` (custom event or callback)
- **Payload**: `{ nodeId, namespaceId, nodeTitle, nodePath? }`

### WI 1.1.2 — Render focused context in Dot panel
- **Where**: New `DotPanel` component or extend `IntentTab` / `App.tsx`
- **What**: Show a " Dot Context" section: node title, path, context chip
- **Optional**: Links to node description/materials if available

### WI 1.1.3 — Pass focused context to inference calls
- **Where**: When calling Centralised AI `/ai/generate` (Phase 2)
- **What**: Include `focused_context: { node_id, namespace_id, title, materials_summary? }` in the request body

---

## Phase 2 — Generate Options (ProposalSet)

**Job 1.2** — Generate multiple options from an open/local model via Centralised AI.

### WI 1.2.1 — Define ProposalSet schema
- **Where**: `internal/domain/proposal.go` (Kernel) + `web-dot/src/types/domain.ts` (UI)
- **Schema**:
  ```json
  {
    "proposal_set_id": "uuid",
    "question": "string",
    "focused_node_id": "string",
    "namespace_id": "string",
    "options": [
      {
        "id": "uuid",
        "summary": "string",
        "tradeoffs": "string",
        "risks": "string",
        "reversibility": "string",
        "assumptions": "string",
        "signals": []
      }
    ],
    "recommended_option_id": "uuid?",
    "confidence": "number?",
    "created_at": "timestamp"
  }
  ```
- **Validation**: `options.length >= 3` (default policy)

### WI 1.2.2 — Use Centralised AI for options generation
- **Where**: [centralised-ai-tool](https://github.com/future-gareth/centralised-ai-tool) (e.g. https://garethapi.com/central-ai)
- **What**: `POST /ai/generate` with a system prompt that instructs the model to return JSON with `options[]` (at least 3), each with id, summary, tradeoffs, risks, reversibility, assumptions.

### WI 1.2.3 — Add Dot UI rendering for ProposalSet
- **Where**: New component `web-dot/src/components/ProposalSetCard.tsx` or extend Dot panel
- **What**: Display each option as a card; label state: **"Proposed — not yet resolved."**
- **Layout**: Cards in a row or list; highlight `recommended_option_id` if present

### WI 1.2.4 — Integrate Dot UI with Centralised AI
- **Where**: `web-dot/server/routes/gui.ts`, `web-dot/server/config.ts`
- **What**:
  - Config: `CENTRAL_AI_URL` (default: `https://garethapi.com/central-ai`)
  - Invoke `POST /ai/generate` with system_prompt + prompt (question + focused context)
  - Parse `response` as JSON; validate `options.length >= 3`; return ProposalSet for client to store

---

## Phase 3 — Resolution Step in Kernel

**Job 1.3** — Add explicit Resolution to the Kernel.

### WI 1.3.1 — Add kernel command: `resolve`
- **Where**: `internal/kernel/service.go`, `internal/api/handlers.go`
- **Inputs**: `proposal_set_id`, `option_id`, `resolver` (identity), optional `rationale`, optional `constraints`
- **Output**: `resolution_id`

**Implementation**:
- New handler: `POST /v1/resolve`
- New domain type: `Resolution` in `internal/domain/resolution.go`
- New store: `resolutions` table (or `proposal_resolutions`); persist resolution record

### WI 1.3.2 — Add ledger event type: `ResolutionRecorded`
- **Where**: `internal/domain/operation.go` or new event types
- **What**: Resolution is a first-class ledger event (or stored in operations log with new `event_type`)
- **Fields**: proposal reference, chosen option, resolver identity, timestamp, rationale, assumptions snapshot

**Storage**:
- Migration: `0006_proposal_resolutions.sql` — table for proposal_sets and resolutions
- Or: use operations log with `changes_json` containing ProposalSet/Resolution payload

### WI 1.3.3 — Enforce "resolution required" rule in Kernel
- **Where**: `internal/kernel/service.go` (Apply logic)
- **What**: Before apply, check if the plan is "decision-bearing"
- **Rule**: If plan class is decision-bearing and no resolution exists → refuse with "Resolution required."
- **Policy**: Allow override via policy (e.g. `low-risk` class allows single option / no resolution)

### WI 1.3.4 — Link `apply` to `resolution_id`
- **Where**: `ApplyRequest` in `internal/kernel`, `internal/api/handlers.go`
- **What**: Add optional `resolution_id` to apply request
- **Enforcement**: For decision-bearing plans, `resolution_id` is required

---

## Phase 4 — Record on Decision Ledger & Show Live

**Job 1.4** — Record the decision and show it in the UI.

### WI 1.4.1 — Add ledger feed UI entry for ProposalSet + Resolution
- **Where**: `web-dot/src/components/LedgerTab.tsx`, `web-dot/server/routes/gui.ts`
- **What**:
  - Kernel/API: `GET /v1/history` or new `GET /v1/decision_ledger` — include proposal_sets and resolutions
  - UI: After options generation → show `ProposalSetCreated` linked to focused node
  - UI: After user chooses → show `ResolutionRecorded` linked to focused node

### WI 1.4.2 — Ensure ledger entries link to details
- **Where**: `LedgerEntryDetails.tsx`, ledger row click handler
- **What**: Click ProposalSet → show all options; click Resolution → show chosen option + rationale

---

## Phase 5 — UI State Machine (One-Minute Demo)

**Goal 2** — Make the demo understandable in one minute.

### WI 2.1.1 — Dot panel state machine
- **Where**: `web-dot/src/components/DotPanel.tsx` (or similar)
- **States**: `NoContext` → `ContextFocused` → `OptionsProposed` → `Resolved` → (optional) `Applied`
- **Transitions**:
  - Focus on node → `ContextFocused`
  - Options returned from Talent → `OptionsProposed`
  - User selects option + rationale → `Resolved`
  - User applies plan → `Applied`

### WI 2.1.2 — Disable Apply until resolved
- **Where**: Apply button / plan-apply UI
- **What**: Disable Apply until `resolution_id` exists; show message: **"Resolution required."**

### WI 2.1.3 — Make Resolver explicit
- **Where**: Resolution display, ledger entry
- **What**: When resolved, show "Resolved by Gareth" (or user identity). If delegated to AI: "Resolved by AI (delegated)" and record delegation.

---

## Phase 6 — Kernel Contract (Model-Agnostic)

**Goal 3** — Keep it model-agnostic.

### WI 3.1.1 — Kernel validation: minimum option count
- **Where**: When storing/accepting ProposalSet
- **What**: Reject if `options.length < 3` (default policy). Allow override by policy.

### WI 3.1.2 — Store raw model output as artefact
- **Where**: `proposal_sets` table or operations log
- **What**: Keep `raw_model_output` (or `artefact_ref`) for provenance

### WI 3.1.3 — Add policy hooks
- **Where**: `internal/policy/` or PolicySet YAML
- **What**: Decision classes:
  - `low-risk`: single option allowed, no resolution required
  - `medium-risk` / `high-risk`: multiple options + human resolution required

---

## Suggested Implementation Order

| Phase | Sequence | Dependencies |
|-------|----------|--------------|
| **1.1 Focus** | 1 | None |
| **1.2 ProposalSet schema** | 2 | None |
| **1.2 Centralised AI options** | 3 | Centralised AI service |
| **1.2 Dot UI for ProposalSet** | 4 | 1.2.1 |
| **1.2 Centralised AI integration** | 5 | 1.2.2, 1.1 |
| **1.3 Kernel resolve** | 6 | 1.2.1 |
| **1.3 ResolutionRecorded** | 7 | 1.3.1 |
| **1.3 Enforce resolution** | 8 | 1.3.1, 1.3.2 |
| **1.3 Link apply to resolution** | 9 | 1.3.1 |
| **1.4 Ledger feed** | 10 | 1.3.2 |
| **2.1 State machine** | 11 | 1.1, 1.2.3, 1.3 |
| **2.1 Disable Apply** | 12 | 1.3.4 |
| **2.1 Resolver explicit** | 13 | 1.3.2 |
| **3.1 Validation & policy** | 14 | 1.3 |

---

## Demo Script (Audience Flow)

1. User clicks **Focus Dot on this** on a Product Tree node.
2. User asks Dot a question about that area.
3. Dot returns **3 options** (cards). UI says **Proposed — not yet resolved**.
4. User clicks **Choose Option B**, adds a short rationale.
5. The Decision Ledger instantly shows **ResolutionRecorded** linked to the node.
6. Only now can the system **Apply** or **Create Work Items**.

---

## File Checklist

| File | Action |
|------|--------|
| `internal/domain/proposal.go` | Create |
| `internal/domain/resolution.go` | Create |
| `migrations/0006_proposal_resolutions.sql` | Create |
| `internal/api/handlers.go` | Add `POST /v1/resolve`, extend Apply |
| `internal/kernel/service.go` | Add Resolve, extend Apply checks |
| `internal/store/proposals.go` | Create |
| `web-dot/src/components/DotPanel.tsx` | Create (or extend IntentTab) |
| `web-dot/src/components/ProposalSetCard.tsx` | Create |
| `web-dot/src/components/LedgerTreeNav.tsx` | Add Focus action |
| `web-dot/server/routes/gui.ts` | Add Centralised AI options/infer, decision ledger |
| `web-dot/server/config.ts` | Add `CENTRAL_AI_URL` |

---

## Out of Scope (for MVP)

- Centralised AI service (default: https://garethapi.com/central-ai; override with CENTRAL_AI_URL)
- Full policy engine for decision classes (can start with hardcoded "3 options required")
- Delegated AI resolution (can add later)
- Cloud sync or multi-user

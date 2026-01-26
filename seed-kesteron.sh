#!/usr/bin/env bash
set -euo pipefail

NS="${NS:-ProductLedger:/Kesteron}"

extract_node_id() {
  tr ' ' '\n' | grep -E '^node:[A-Za-z0-9-]+' | head -n 1
}

die() {
  echo "ERROR: $*" >&2
  exit 1
}

create_node() {
  local title="$1"
  local nid
  nid="$(dot new node "$title" --yes | extract_node_id || true)"
  [[ -n "${nid}" ]] || die "Could not extract node id from dot output for: ${title}"
  echo "$nid"
}

assign_role() {
  local nid="$1"
  local role="$2"
  dot role assign "$nid" "$role" --yes >/dev/null || die "role assign failed for $nid ($role)"
}

create_node_with_role() {
  local title="$1"
  local role="$2"

  echo "Creating node: \"$title\" (role: $role)" >&2
  local nid
  nid="$(create_node "$title")"
  assign_role "$nid" "$role"

  # IMPORTANT: only emit node id on stdout (so callers capture a clean ID)
  echo "$nid"
}

create_link() {
  local from="$1"
  local ltype="$2"
  local to="$3"

  echo "Link: $from -[$ltype]-> $to" >&2
  dot link "$from" "$ltype" "$to" --yes >/dev/null || die "link failed: $from -[$ltype]-> $to"
}

echo "Using namespace: $NS" >&2
dot use "$NS" || die "Failed to use namespace $NS"

echo "Running smoke test..." >&2
SMOKE="$(dot new node "Seed Smoke Test $(date +%s)" --yes | extract_node_id || true)"
[[ -n "${SMOKE}" ]] || die "Smoke test failed: couldn't create node (is server running? is namespace present?)"
echo "Smoke test OK: $SMOKE" >&2

# Enterprise Intent
SO1="$(create_node_with_role "SO-1 Improve reliability and explainability of customer-critical services" "EnterpriseIntent.StrategicObjective")"
AO1="$(create_node_with_role "AO-1 Maintain traceable decision records for externally visible changes" "EnterpriseIntent.AssuranceObligation")"
TT1="$(create_node_with_role "TT-1 Increase automation without reducing accountability" "EnterpriseIntent.TransformationTheme")"

# Products
PFS="$(create_node_with_role "P-FS FieldServe" "Product")"
PAL="$(create_node_with_role "P-AL AssetLink" "Product")"

# FieldServe tree
FSG1="$(create_node_with_role "FS-G1 Reduce unplanned service disruption" "Goal")"
FSG2="$(create_node_with_role "FS-G2 Ensure operational changes are explainable after deployment" "Goal")"
FSJ1="$(create_node_with_role "FS-J1 Understand operational impact of system changes before deployment" "Job")"
FSJ2="$(create_node_with_role "FS-J2 Reconstruct why a change was made when issues arise" "Job")"
FSW1="$(create_node_with_role "FS-W1 Record deployment decisions and rationale in the Product Ledger" "WorkItem")"
FSW2="$(create_node_with_role "FS-W2 Link operational changes to affected downstream systems" "WorkItem")"

create_link "$PFS" "CONTAINS" "$FSG1"
create_link "$PFS" "CONTAINS" "$FSG2"
create_link "$FSG1" "CONTAINS" "$FSJ1"
create_link "$FSG2" "CONTAINS" "$FSJ2"
create_link "$FSJ1" "CONTAINS" "$FSW1"
create_link "$FSJ2" "CONTAINS" "$FSW2"

create_link "$FSG1" "SUPPORTS" "$SO1"
create_link "$FSG2" "SUPPORTS" "$SO1"
create_link "$FSG2" "SATISFIES" "$AO1"

# AssetLink tree
ALG1="$(create_node_with_role "AL-G1 Provide a single, trusted source of asset state" "Goal")"
ALG2="$(create_node_with_role "AL-G2 Make data-model changes explicit and reviewable" "Goal")"
ALJ1="$(create_node_with_role "AL-J1 Allow consuming systems to understand asset data semantics" "Job")"
ALJ2="$(create_node_with_role "AL-J2 Explain historical data changes to internal and external stakeholders" "Job")"
ALW1="$(create_node_with_role "AL-W1 Introduce versioned asset schema records" "WorkItem")"

create_link "$PAL" "CONTAINS" "$ALG1"
create_link "$PAL" "CONTAINS" "$ALG2"
create_link "$ALG1" "CONTAINS" "$ALJ1"
create_link "$ALG2" "CONTAINS" "$ALJ2"
create_link "$ALJ2" "CONTAINS" "$ALW1"

create_link "$ALG1" "SUPPORTS" "$SO1"
create_link "$ALG2" "SATISFIES" "$AO1"
create_link "$ALG2" "ADVANCES" "$TT1"

# Cross-product dependency
D1="$(create_node_with_role "D-1 FieldServe depends on AssetLink asset status semantics" "Dependency")"

create_link "$FSG1" "DEPENDS_ON" "$D1"
create_link "$ALG1" "DEPENDS_ON" "$D1"
create_link "$D1" "AFFECTS" "$PFS"
create_link "$D1" "AFFECTS" "$PAL"

# Decision + Evidence
FSD1="$(create_node_with_role "FS-D1 Require explicit recording of AssetLink schema changes impacting FieldServe" "Decision")"
E1="$(create_node_with_role "E-1 Incident review summary (AI-assisted, human-approved)" "Evidence")"

create_link "$FSG2" "DECIDED_BY" "$FSD1"
create_link "$ALG2" "DECIDED_BY" "$FSD1"
create_link "$FSD1" "EVIDENCED_BY" "$E1"
create_link "$FSD1" "SATISFIES" "$AO1"

echo "" >&2
echo "✅ Kesteron seed complete." >&2
echo "Namespace: $NS" >&2

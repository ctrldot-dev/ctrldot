# FMPL v0.1 — Minimal Policy Language

Policies are namespace-scoped. Evaluated during Plan and rechecked during Apply.

## PolicySet format (YAML)
```yaml
policy_set: ProductTree:MachinePay
version: 0.1
applies_to:
  namespace: "ProductTree:/MachinePay"
rules:
  - id: no_cycles_parent_of
    when:
      op: ["CreateLink","Move"]
      link_type: "PARENT_OF"
    require:
      - predicate: "acyclic"
        args: { link_type: "PARENT_OF" }
    effect:
      deny: "Hierarchy cannot contain cycles."

  - id: workitem_only_under_job
    when:
      op: ["CreateLink","Move"]
      link_type: "PARENT_OF"
    require:
      - predicate: "role_edge_allowed"
        args:
          parent_role: ["Job"]
          child_role: ["WorkItem"]
    effect:
      deny: "Work Items must be children of Jobs."

  - id: workitem_single_parent
    when:
      op: ["CreateLink","Move"]
      link_type: "PARENT_OF"
    require:
      - predicate: "child_has_only_one_parent"
        args:
          child_role: "WorkItem"
          link_type: "PARENT_OF"
    effect:
      deny: "Work Items must have exactly one parent."
```

## Predicates required in v0.1
- acyclic(link_type)
- role_edge_allowed(parent_role[], child_role[])
- child_has_only_one_parent(child_role, link_type)
- has_capability(cap)  (optional to wire in v0.1; include stub)

## Effects
- deny: blocks apply
- warn: does not block apply
- info: does not block apply

## Notes
- Policies must be implementable via indexed lookups; avoid unbounded graph walks.
- `acyclic` may perform a bounded traversal for PARENT_OF within a namespace.

# Acceptance Tests (v0.1)

These define “done”. Implement integration tests against a real Postgres (docker-compose or testcontainers).

## A. Plan/Apply + Audit
1. Plan hashing is deterministic: repeating the same Plan request against the same as-of state yields the same `plan.hash`.
2. Apply appends exactly one Operation with the next monotonic `seq`.
3. Re-applying the same plan MUST fail with CONFLICT (plans are single-use in v0.1).

## B. As-of semantics
4. Create a node at seq N. Query/expand asof_seq=N-1 MUST NOT show the node; asof_seq=N MUST show it.
5. Create a link at seq A then retire it at seq B. Expand asof_seq=B-1 MUST show the link; asof_seq=B MUST NOT.

## C. Policy enforcement
6. Attempt to create a PARENT_OF cycle MUST be denied with `POLICY_DENIED` and the rule id `no_cycles_parent_of`.
7. Attempt to link a WorkItem under a Goal via PARENT_OF MUST be denied with rule id `workitem_only_under_job`.
8. Attempt to give a WorkItem a second parent MUST be denied with rule id `workitem_single_parent`.

## D. Diff
9. Diff between two seqs for a namespace MUST return the net Changes (created and retired links/nodes) between those seqs.

## E. Expand
10. Expand with depth=1 MUST return: requested nodes + their roles (if namespace provided) + adjacent links (in namespace if provided).

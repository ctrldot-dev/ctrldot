package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/futurematic/kernel/internal/domain"
)

// GetActivePolicySet retrieves the active policy set for a namespace.
// Exact namespace_id match is tried first; if none, the longest prefix match is used.
func (t *PostgresTx) GetActivePolicySet(ctx context.Context, namespaceID string) (*domain.PolicySet, error) {
	var policySet domain.PolicySet
	var createdAt time.Time
	var retiredSeq sql.NullInt64

	err := t.tx.QueryRowContext(ctx,
		`SELECT policy_set_id, namespace_id, policy_yaml, policy_hash, created_at, created_seq, retired_seq
		 FROM policy_sets
		 WHERE is_active = TRUE AND retired_seq IS NULL
		   AND (namespace_id = $1 OR $1 LIKE namespace_id || ':%' OR $1 LIKE namespace_id || '/%')
		 ORDER BY length(namespace_id) DESC
		 LIMIT 1`,
		namespaceID,
	).Scan(&policySet.ID, &policySet.NamespaceID, &policySet.PolicyYAML, &policySet.PolicyHash, &createdAt, &policySet.CreatedSeq, &retiredSeq)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No active policy set is valid
		}
		return nil, fmt.Errorf("failed to get active policy set: %w", err)
	}

	policySet.CreatedAt = createdAt

	if retiredSeq.Valid {
		policySet.RetiredSeq = &retiredSeq.Int64
	}

	return &policySet, nil
}

// StorePolicySet stores a policy set
func (t *PostgresTx) StorePolicySet(ctx context.Context, policySet domain.PolicySet, seq int64) error {
	// When storing a new policy set, we need to deactivate old ones
	// First, deactivate all existing active policy sets for this namespace
	_, err := t.tx.ExecContext(ctx,
		`UPDATE policy_sets SET is_active = FALSE WHERE namespace_id = $1 AND is_active = TRUE`,
		policySet.NamespaceID,
	)
	if err != nil {
		return fmt.Errorf("failed to deactivate old policy sets: %w", err)
	}

	// Then insert the new one as active
	_, err = t.tx.ExecContext(ctx,
		`INSERT INTO policy_sets (policy_set_id, namespace_id, policy_yaml, policy_hash, created_at, created_seq, is_active)
		 VALUES ($1, $2, $3, $4, $5, $6, TRUE)
		 ON CONFLICT (policy_set_id) DO UPDATE SET
		   namespace_id = EXCLUDED.namespace_id,
		   policy_yaml = EXCLUDED.policy_yaml,
		   policy_hash = EXCLUDED.policy_hash,
		   created_at = EXCLUDED.created_at,
		   created_seq = EXCLUDED.created_seq,
		   is_active = TRUE,
		   retired_seq = NULL`,
		policySet.ID, policySet.NamespaceID, policySet.PolicyYAML, policySet.PolicyHash, time.Now(), seq,
	)
	if err != nil {
		return fmt.Errorf("failed to insert policy set: %w", err)
	}
	return nil
}

package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/futurematic/kernel/internal/domain"
)

// AppendOperation appends an operation to the operations log
func (t *PostgresTx) AppendOperation(ctx context.Context, op domain.Operation) error {
	capabilitiesJSON, err := json.Marshal(op.Capabilities)
	if err != nil {
		return fmt.Errorf("failed to marshal capabilities: %w", err)
	}

	changesJSON, err := json.Marshal(op.Changes)
	if err != nil {
		return fmt.Errorf("failed to marshal changes: %w", err)
	}

	_, err = t.tx.ExecContext(ctx,
		`INSERT INTO operations (seq, op_id, occurred_at, actor_id, capabilities, plan_id, plan_hash, class, changes_json)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		op.Seq, op.ID, op.OccurredAt, op.ActorID, capabilitiesJSON, op.PlanID, op.PlanHash, op.Class, changesJSON,
	)
	if err != nil {
		return fmt.Errorf("failed to insert operation: %w", err)
	}
	return nil
}

// GetOperation retrieves an operation by sequence number
func (t *PostgresTx) GetOperation(ctx context.Context, seq int64) (*domain.Operation, error) {
	var op domain.Operation
	var occurredAt time.Time
	var capabilitiesJSON []byte
	var changesJSON []byte

	err := t.tx.QueryRowContext(ctx,
		`SELECT seq, op_id, occurred_at, actor_id, capabilities, plan_id, plan_hash, class, changes_json
		 FROM operations WHERE seq = $1`,
		seq,
	).Scan(&op.Seq, &op.ID, &occurredAt, &op.ActorID, &capabilitiesJSON, &op.PlanID, &op.PlanHash, &op.Class, &changesJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("operation not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get operation: %w", err)
	}

	op.OccurredAt = occurredAt

	if err := json.Unmarshal(capabilitiesJSON, &op.Capabilities); err != nil {
		return nil, fmt.Errorf("failed to unmarshal capabilities: %w", err)
	}

	if err := json.Unmarshal(changesJSON, &op.Changes); err != nil {
		return nil, fmt.Errorf("failed to unmarshal changes: %w", err)
	}

	return &op, nil
}

// GetOperationsForTarget retrieves operations for a target (node ID or namespace ID)
func (t *PostgresTx) GetOperationsForTarget(ctx context.Context, target string, limit int) ([]domain.Operation, error) {
	if limit <= 0 {
		limit = 100
	}

	// Query operations that affect the target
	// This is a simplified version - in a real implementation, we'd need to parse changes_json
	// to find operations that reference the target
	rows, err := t.tx.QueryContext(ctx,
		`SELECT seq, op_id, occurred_at, actor_id, capabilities, plan_id, plan_hash, class, changes_json
		 FROM operations
		 WHERE changes_json::text LIKE '%' || $1 || '%'
		 ORDER BY seq DESC
		 LIMIT $2`,
		target, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query operations: %w", err)
	}
	defer rows.Close()

	var operations []domain.Operation
	for rows.Next() {
		var op domain.Operation
		var occurredAt time.Time
		var capabilitiesJSON []byte
		var changesJSON []byte

		if err := rows.Scan(&op.Seq, &op.ID, &occurredAt, &op.ActorID, &capabilitiesJSON, &op.PlanID, &op.PlanHash, &op.Class, &changesJSON); err != nil {
			return nil, fmt.Errorf("failed to scan operation: %w", err)
		}

		op.OccurredAt = occurredAt

		if err := json.Unmarshal(capabilitiesJSON, &op.Capabilities); err != nil {
			return nil, fmt.Errorf("failed to unmarshal capabilities: %w", err)
		}

		if err := json.Unmarshal(changesJSON, &op.Changes); err != nil {
			return nil, fmt.Errorf("failed to unmarshal changes: %w", err)
		}

		operations = append(operations, op)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating operations: %w", err)
	}

	return operations, nil
}

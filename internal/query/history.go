package query

import (
	"context"
	"fmt"

	"github.com/futurematic/kernel/internal/domain"
)

// History implements the Engine interface
func (e *engine) History(ctx context.Context, req HistoryRequest) ([]domain.Operation, error) {
	// Open a read transaction
	tx, err := e.store.OpenTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to open transaction: %w", err)
	}
	defer tx.Rollback()

	limit := req.Limit
	if limit <= 0 {
		limit = 100
	}

	operations, err := tx.GetOperationsForTarget(ctx, req.Target, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get operations: %w", err)
	}

	return operations, nil
}

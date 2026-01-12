package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/futurematic/kernel/internal/domain"
	_ "github.com/lib/pq"
)

// PostgresStore implements Store using Postgres
type PostgresStore struct {
	db *sql.DB
}

// NewPostgresStore creates a new Postgres store
func NewPostgresStore(dbURL string) (*PostgresStore, error) {
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &PostgresStore{db: db}, nil
}

// Close closes the database connection
func (s *PostgresStore) Close() error {
	return s.db.Close()
}

// OpenTx opens a new transaction
func (s *PostgresStore) OpenTx(ctx context.Context) (Tx, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	return &PostgresTx{tx: tx}, nil
}

// GetNextSeq returns the next sequence number
func (s *PostgresStore) GetNextSeq(ctx context.Context) (int64, error) {
	var seq int64
	err := s.db.QueryRowContext(ctx, "SELECT COALESCE(MAX(seq), 0) + 1 FROM operations").Scan(&seq)
	if err != nil {
		return 0, fmt.Errorf("failed to get next seq: %w", err)
	}
	return seq, nil
}

// ResolveAsOf resolves an AsOf to a sequence number
func (s *PostgresStore) ResolveAsOf(ctx context.Context, asof domain.AsOf) (int64, error) {
	if asof.Seq != nil {
		return *asof.Seq, nil
	}
	if asof.Time != nil {
		var seq int64
		err := s.db.QueryRowContext(ctx,
			"SELECT COALESCE(MAX(seq), 0) FROM operations WHERE occurred_at <= $1",
			*asof.Time,
		).Scan(&seq)
		if err != nil {
			return 0, fmt.Errorf("failed to resolve asof time: %w", err)
		}
		return seq, nil
	}
	// If neither is set, return latest seq
	var seq int64
	err := s.db.QueryRowContext(ctx, "SELECT COALESCE(MAX(seq), 0) FROM operations").Scan(&seq)
	if err != nil {
		return 0, fmt.Errorf("failed to get latest seq: %w", err)
	}
	return seq, nil
}

// GetActivePolicySet retrieves the active policy set for a namespace
func (s *PostgresStore) GetActivePolicySet(ctx context.Context, namespaceID string) (*domain.PolicySet, error) {
	var policySet domain.PolicySet
	var createdAt time.Time
	var retiredSeq sql.NullInt64

	err := s.db.QueryRowContext(ctx,
		`SELECT policy_set_id, namespace_id, policy_yaml, policy_hash, created_at, created_seq, retired_seq
		 FROM policy_sets
		 WHERE namespace_id = $1 AND is_active = TRUE AND retired_seq IS NULL
		 ORDER BY created_seq DESC
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

// CreateNamespace creates a namespace
func (s *PostgresStore) CreateNamespace(ctx context.Context, namespace domain.Namespace) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO namespaces (namespace_id, name)
		 VALUES ($1, $2)
		 ON CONFLICT (namespace_id) DO NOTHING`,
		namespace.ID, namespace.Name,
	)
	if err != nil {
		return fmt.Errorf("failed to create namespace: %w", err)
	}
	return nil
}

// PostgresTx implements Tx using Postgres
type PostgresTx struct {
	tx *sql.Tx
}

// Commit commits the transaction
func (t *PostgresTx) Commit() error {
	return t.tx.Commit()
}

// Rollback rolls back the transaction
func (t *PostgresTx) Rollback() error {
	return t.tx.Rollback()
}

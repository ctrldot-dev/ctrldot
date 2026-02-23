package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/futurematic/kernel/internal/domain"
)

// StoreProposalSet stores a proposal set (call within a transaction if needed for consistency)
func (t *PostgresTx) StoreProposalSet(ctx context.Context, ps *domain.ProposalSet) error {
	optionsJSON, err := json.Marshal(ps.Options)
	if err != nil {
		return fmt.Errorf("marshal options: %w", err)
	}
	_, err = t.tx.ExecContext(ctx,
		`INSERT INTO proposal_sets (proposal_set_id, question, focused_node_id, namespace_id, options_json, recommended_option_id, confidence, raw_model_output, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		 ON CONFLICT (proposal_set_id) DO NOTHING`,
		ps.ProposalSetID, ps.Question, ps.FocusedNodeID, ps.NamespaceID, optionsJSON, ps.RecommendedOptionID, ps.Confidence, ps.RawModelOutput, ps.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert proposal set: %w", err)
	}
	return nil
}

// StoreResolution stores a resolution (call within a transaction if needed)
func (t *PostgresTx) StoreResolution(ctx context.Context, r *domain.Resolution) error {
	_, err := t.tx.ExecContext(ctx,
		`INSERT INTO resolutions (resolution_id, proposal_set_id, option_id, resolver, rationale, constraints, assumptions_snap, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		r.ResolutionID, r.ProposalSetID, r.OptionID, r.Resolver, r.Rationale, r.Constraints, r.AssumptionsSnap, r.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert resolution: %w", err)
	}
	return nil
}

// GetProposalSet retrieves a proposal set by ID (Store-level, no tx)
func (s *PostgresStore) GetProposalSet(ctx context.Context, proposalSetID string) (*domain.ProposalSet, error) {
	var ps domain.ProposalSet
	var optionsJSON []byte
	var recOptID sql.NullString
	var confidence sql.NullFloat64
	err := s.db.QueryRowContext(ctx,
		`SELECT proposal_set_id, question, focused_node_id, namespace_id, options_json, recommended_option_id, confidence, raw_model_output, created_at
		 FROM proposal_sets WHERE proposal_set_id = $1`,
		proposalSetID,
	).Scan(&ps.ProposalSetID, &ps.Question, &ps.FocusedNodeID, &ps.NamespaceID, &optionsJSON, &recOptID, &confidence, &ps.RawModelOutput, &ps.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrProposalSetNotFound
		}
		return nil, fmt.Errorf("get proposal set: %w", err)
	}
	if err := json.Unmarshal(optionsJSON, &ps.Options); err != nil {
		return nil, fmt.Errorf("unmarshal options: %w", err)
	}
	if recOptID.Valid {
		ps.RecommendedOptionID = &recOptID.String
	}
	if confidence.Valid {
		ps.Confidence = &confidence.Float64
	}
	return &ps, nil
}

// GetResolution retrieves a resolution by ID (Store-level)
func (s *PostgresStore) GetResolution(ctx context.Context, resolutionID string) (*domain.Resolution, error) {
	var r domain.Resolution
	err := s.db.QueryRowContext(ctx,
		`SELECT resolution_id, proposal_set_id, option_id, resolver, rationale, constraints, assumptions_snap, created_at
		 FROM resolutions WHERE resolution_id = $1`,
		resolutionID,
	).Scan(&r.ResolutionID, &r.ProposalSetID, &r.OptionID, &r.Resolver, &r.Rationale, &r.Constraints, &r.AssumptionsSnap, &r.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrResolutionNotFound
		}
		return nil, fmt.Errorf("get resolution: %w", err)
	}
	return &r, nil
}

// ListProposalSetsForNode returns recent proposal sets for a node (Store-level)
func (s *PostgresStore) ListProposalSetsForNode(ctx context.Context, nodeID, namespaceID string, limit int) ([]domain.ProposalSet, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := s.db.QueryContext(ctx,
		`SELECT proposal_set_id, question, focused_node_id, namespace_id, options_json, recommended_option_id, confidence, raw_model_output, created_at
		 FROM proposal_sets WHERE focused_node_id = $1 AND namespace_id = $2 ORDER BY created_at DESC LIMIT $3`,
		nodeID, namespaceID, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("list proposal sets: %w", err)
	}
	defer rows.Close()
	var out []domain.ProposalSet
	for rows.Next() {
		var ps domain.ProposalSet
		var optionsJSON []byte
		var recOptID sql.NullString
		var confidence sql.NullFloat64
		if err := rows.Scan(&ps.ProposalSetID, &ps.Question, &ps.FocusedNodeID, &ps.NamespaceID, &optionsJSON, &recOptID, &confidence, &ps.RawModelOutput, &ps.CreatedAt); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(optionsJSON, &ps.Options); err != nil {
			return nil, err
		}
		if recOptID.Valid {
			ps.RecommendedOptionID = &recOptID.String
		}
		if confidence.Valid {
			ps.Confidence = &confidence.Float64
		}
		out = append(out, ps)
	}
	return out, rows.Err()
}

// GetResolutionForProposalSet returns the resolution for a proposal set if any (Store-level)
func (s *PostgresStore) GetResolutionForProposalSet(ctx context.Context, proposalSetID string) (*domain.Resolution, error) {
	var r domain.Resolution
	err := s.db.QueryRowContext(ctx,
		`SELECT resolution_id, proposal_set_id, option_id, resolver, rationale, constraints, assumptions_snap, created_at
		 FROM resolutions WHERE proposal_set_id = $1`,
		proposalSetID,
	).Scan(&r.ResolutionID, &r.ProposalSetID, &r.OptionID, &r.Resolver, &r.Rationale, &r.Constraints, &r.AssumptionsSnap, &r.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get resolution for proposal set: %w", err)
	}
	return &r, nil
}

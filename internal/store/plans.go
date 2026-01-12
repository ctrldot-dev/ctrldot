package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/futurematic/kernel/internal/domain"
)

// StorePlan stores a plan in the database
func (t *PostgresTx) StorePlan(ctx context.Context, plan domain.Plan, policyHash string) error {
	intentsJSON, err := json.Marshal(plan.Intents)
	if err != nil {
		return fmt.Errorf("failed to marshal intents: %w", err)
	}

	expandedJSON, err := json.Marshal(plan.Expanded)
	if err != nil {
		return fmt.Errorf("failed to marshal expanded: %w", err)
	}

	policyReportJSON, err := json.Marshal(plan.PolicyReport)
	if err != nil {
		return fmt.Errorf("failed to marshal policy report: %w", err)
	}

	_, err = t.tx.ExecContext(ctx,
		`INSERT INTO plans (plan_id, created_at, actor_id, namespace_id, asof_seq, policy_hash, plan_hash, class, intents_json, expanded_json, policy_report_json)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		 ON CONFLICT (plan_id) DO NOTHING`,
		plan.ID, plan.CreatedAt, plan.ActorID, plan.NamespaceID, plan.AsOfSeq, policyHash, plan.Hash, plan.Class, intentsJSON, expandedJSON, policyReportJSON,
	)
	if err != nil {
		return fmt.Errorf("failed to insert plan: %w", err)
	}
	return nil
}

// GetPlan retrieves a plan by ID
func (t *PostgresTx) GetPlan(ctx context.Context, planID string) (*domain.Plan, error) {
	var plan domain.Plan
	var createdAt time.Time
	var namespaceID sql.NullString
	var asofSeq sql.NullInt64
	var intentsJSON, expandedJSON, policyReportJSON []byte
	var appliedOpID sql.NullString
	var appliedSeq sql.NullInt64

	err := t.tx.QueryRowContext(ctx,
		`SELECT plan_id, created_at, actor_id, namespace_id, asof_seq, plan_hash, class, intents_json, expanded_json, policy_report_json, applied_op_id, applied_seq
		 FROM plans WHERE plan_id = $1`,
		planID,
	).Scan(&plan.ID, &createdAt, &plan.ActorID, &namespaceID, &asofSeq, &plan.Hash, &plan.Class, &intentsJSON, &expandedJSON, &policyReportJSON, &appliedOpID, &appliedSeq)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("plan not found")
		}
		return nil, fmt.Errorf("failed to get plan: %w", err)
	}

	plan.CreatedAt = createdAt

	if namespaceID.Valid {
		plan.NamespaceID = &namespaceID.String
	}

	if asofSeq.Valid {
		plan.AsOfSeq = &asofSeq.Int64
	}

	if err := json.Unmarshal(intentsJSON, &plan.Intents); err != nil {
		return nil, fmt.Errorf("failed to unmarshal intents: %w", err)
	}

	if err := json.Unmarshal(expandedJSON, &plan.Expanded); err != nil {
		return nil, fmt.Errorf("failed to unmarshal expanded: %w", err)
	}

	if err := json.Unmarshal(policyReportJSON, &plan.PolicyReport); err != nil {
		return nil, fmt.Errorf("failed to unmarshal policy report: %w", err)
	}

	return &plan, nil
}

// IsPlanApplied checks if a plan has already been applied
func (t *PostgresTx) IsPlanApplied(ctx context.Context, planID string) (bool, error) {
	var appliedSeq sql.NullInt64
	err := t.tx.QueryRowContext(ctx,
		"SELECT applied_seq FROM plans WHERE plan_id = $1", planID,
	).Scan(&appliedSeq)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil // Plan doesn't exist, so not applied
		}
		return false, fmt.Errorf("failed to check if plan is applied: %w", err)
	}
	return appliedSeq.Valid, nil
}

// MarkPlanApplied marks a plan as applied
func (t *PostgresTx) MarkPlanApplied(ctx context.Context, planID string, opID string, seq int64) error {
	_, err := t.tx.ExecContext(ctx,
		`UPDATE plans SET applied_op_id = $1, applied_seq = $2 WHERE plan_id = $3`,
		opID, seq, planID,
	)
	if err != nil {
		return fmt.Errorf("failed to mark plan as applied: %w", err)
	}
	return nil
}

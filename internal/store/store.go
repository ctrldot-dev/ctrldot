package store

import (
	"context"

	"github.com/futurematic/kernel/internal/domain"
)

// Store provides database operations
type Store interface {
	// OpenTx opens a new transaction
	OpenTx(ctx context.Context) (Tx, error)

	// GetNextSeq returns the next sequence number (outside of transaction)
	GetNextSeq(ctx context.Context) (int64, error)

	// ResolveAsOf resolves an AsOf to a sequence number
	// If AsOf has a seq, returns it directly
	// If AsOf has a time, queries operations table for latest seq at/before that time
	ResolveAsOf(ctx context.Context, asof domain.AsOf) (int64, error)

	// GetActivePolicySet retrieves the active policy set for a namespace (outside transaction)
	GetActivePolicySet(ctx context.Context, namespaceID string) (*domain.PolicySet, error)

	// Namespaces
	CreateNamespace(ctx context.Context, namespace domain.Namespace) error
}

// Tx represents a database transaction
type Tx interface {
	// Commit commits the transaction
	Commit() error

	// Rollback rolls back the transaction
	Rollback() error

	// Operations
	AppendOperation(ctx context.Context, op domain.Operation) error
	GetOperation(ctx context.Context, seq int64) (*domain.Operation, error)
	GetOperationsForTarget(ctx context.Context, target string, limit int) ([]domain.Operation, error)

	// Nodes
	CreateNode(ctx context.Context, node domain.Node, seq int64) error
	GetNode(ctx context.Context, nodeID string, asofSeq int64) (*domain.Node, error)
	RetireNode(ctx context.Context, nodeID string, seq int64) error

	// Links
	CreateLink(ctx context.Context, link domain.Link, seq int64) error
	GetLink(ctx context.Context, linkID string, asofSeq int64) (*domain.Link, error)
	GetLinksFrom(ctx context.Context, fromNodeID string, namespaceID *string, linkType *string, asofSeq int64) ([]domain.Link, error)
	GetLinksTo(ctx context.Context, toNodeID string, namespaceID *string, linkType *string, asofSeq int64) ([]domain.Link, error)
	RetireLink(ctx context.Context, linkID string, seq int64) error

	// Materials
	CreateMaterial(ctx context.Context, material domain.Material, seq int64) error
	GetMaterial(ctx context.Context, materialID string, asofSeq int64) (*domain.Material, error)
	GetMaterialsForNode(ctx context.Context, nodeID string, asofSeq int64) ([]domain.Material, error)
	RetireMaterial(ctx context.Context, materialID string, seq int64) error

	// Role Assignments
	CreateRoleAssignment(ctx context.Context, role domain.RoleAssignment, seq int64) error
	GetRoleAssignments(ctx context.Context, nodeID string, namespaceID string, asofSeq int64) ([]domain.RoleAssignment, error)
	RetireRoleAssignment(ctx context.Context, roleAssignmentID string, seq int64) error

	// Plans
	StorePlan(ctx context.Context, plan domain.Plan, policyHash string) error
	GetPlan(ctx context.Context, planID string) (*domain.Plan, error)
	IsPlanApplied(ctx context.Context, planID string) (bool, error)
	MarkPlanApplied(ctx context.Context, planID string, opID string, seq int64) error

	// Policy Sets
	GetActivePolicySet(ctx context.Context, namespaceID string) (*domain.PolicySet, error)
	StorePolicySet(ctx context.Context, policySet domain.PolicySet, seq int64) error
}

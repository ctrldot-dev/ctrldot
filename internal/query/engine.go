package query

import (
	"context"

	"github.com/futurematic/kernel/internal/domain"
	"github.com/futurematic/kernel/internal/store"
)

// Engine provides query operations
type Engine interface {
	// Expand expands nodes with their roles, links, and materials
	Expand(ctx context.Context, req ExpandRequest) (*ExpandResult, error)

	// History retrieves operations for a target
	History(ctx context.Context, req HistoryRequest) ([]domain.Operation, error)

	// Diff computes the difference between two sequence numbers
	Diff(ctx context.Context, req DiffRequest) (*DiffResult, error)

	// ListNamespaces returns namespace IDs that have at least one node (role assignment) at latest seq
	ListNamespaces(ctx context.Context) ([]string, error)

	// GetNamespaceRoot returns the root node for a namespace (no incoming CONTAINS), preferring role *.Domain
	GetNamespaceRoot(ctx context.Context, namespaceID string) (nodeID, title, role string, err error)
}

// ExpandRequest contains parameters for expand
type ExpandRequest struct {
	NodeIDs     []string
	NamespaceID *string
	Depth       int
	AsOfSeq     int64
}

// ExpandResult contains the expanded data
type ExpandResult struct {
	Nodes     []domain.Node           `json:"nodes"`
	Roles     []domain.RoleAssignment  `json:"role_assignments"`
	Links     []domain.Link           `json:"links"`
	Materials []domain.Material       `json:"materials"`
}

// HistoryRequest contains parameters for history
type HistoryRequest struct {
	Target string
	Limit  int
}

// DiffRequest contains parameters for diff
type DiffRequest struct {
	ASeq   int64
	BSeq   int64
	Target string // node ID or namespace ID
}

// DiffResult contains the computed diff
type DiffResult struct {
	Changes []domain.Change
}

// NewEngine creates a new query engine
func NewEngine(store store.Store) Engine {
	return &engine{
		store: store,
	}
}

type engine struct {
	store store.Store
}

// ListNamespaces implements Engine
func (e *engine) ListNamespaces(ctx context.Context) ([]string, error) {
	tx, err := e.store.OpenTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	asofSeq, err := e.store.ResolveAsOf(ctx, domain.AsOf{})
	if err != nil {
		return nil, err
	}
	return tx.ListNamespaceIDsWithNodes(ctx, asofSeq)
}

// GetNamespaceRoot implements Engine
func (e *engine) GetNamespaceRoot(ctx context.Context, namespaceID string) (nodeID, title, role string, err error) {
	tx, err := e.store.OpenTx(ctx)
	if err != nil {
		return "", "", "", err
	}
	defer tx.Rollback()
	asofSeq, err := e.store.ResolveAsOf(ctx, domain.AsOf{})
	if err != nil {
		return "", "", "", err
	}
	return tx.GetNamespaceRoot(ctx, namespaceID, asofSeq)
}

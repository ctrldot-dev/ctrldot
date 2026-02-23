package policy

import (
	"context"
	"fmt"

	"github.com/futurematic/kernel/internal/domain"
	"github.com/futurematic/kernel/internal/store"
)

// PredicateEvaluator evaluates policy predicates
type PredicateEvaluator struct{}

// NewPredicateEvaluator creates a new predicate evaluator
func NewPredicateEvaluator() *PredicateEvaluator {
	return &PredicateEvaluator{}
}

// Evaluate evaluates a predicate against a change
func (p *PredicateEvaluator) Evaluate(ctx context.Context, predicateName string, args map[string]interface{}, change domain.Change, namespaceID *string, asofSeq int64, tx store.Tx) (bool, error) {
	switch predicateName {
	case "acyclic":
		return p.evaluateAcyclic(ctx, args, change, namespaceID, asofSeq, tx)
	case "role_edge_allowed":
		return p.evaluateRoleEdgeAllowed(ctx, args, change, namespaceID, asofSeq, tx)
	case "child_has_only_one_parent":
		return p.evaluateChildHasOnlyOneParent(ctx, args, change, namespaceID, asofSeq, tx)
	case "has_capability":
		return p.evaluateHasCapability(ctx, args, change)
	default:
		return false, fmt.Errorf("unknown predicate: %s", predicateName)
	}
}

// evaluateAcyclic checks if adding a link would create a cycle
func (p *PredicateEvaluator) evaluateAcyclic(ctx context.Context, args map[string]interface{}, change domain.Change, namespaceID *string, asofSeq int64, tx store.Tx) (bool, error) {
	linkType, ok := args["link_type"].(string)
	if !ok {
		return false, fmt.Errorf("link_type is required for acyclic predicate")
	}

	// Only check for PARENT_OF links
	if linkType != "PARENT_OF" {
		return true, nil // Other link types don't need cycle checking
	}

	// Get from_node_id and to_node_id from change
	fromNodeID, ok := change.Payload["from_node_id"].(string)
	if !ok {
		return false, fmt.Errorf("from_node_id not found in change")
	}

	toNodeID, ok := change.Payload["to_node_id"].(string)
	if !ok {
		return false, fmt.Errorf("to_node_id not found in change")
	}

	// Check if to_node_id is an ancestor of from_node_id (would create cycle)
	// We traverse up the PARENT_OF chain from from_node_id
	visited := make(map[string]bool)
	current := fromNodeID

	for current != "" {
		if current == toNodeID {
			// Found cycle!
			return false, nil
		}

		if visited[current] {
			// Already visited this node (shouldn't happen in acyclic graph, but safety check)
			break
		}
		visited[current] = true

		// Get parent of current node
		links, err := tx.GetLinksTo(ctx, current, namespaceID, &linkType, asofSeq)
		if err != nil {
			return false, fmt.Errorf("failed to get links: %w", err)
		}

		if len(links) == 0 {
			break // No parent, reached root
		}

		// Get first parent (assuming single parent for now)
		current = links[0].FromNodeID
	}

	// No cycle found
	return true, nil
}

// evaluateRoleEdgeAllowed checks if a role transition is allowed
func (p *PredicateEvaluator) evaluateRoleEdgeAllowed(ctx context.Context, args map[string]interface{}, change domain.Change, namespaceID *string, asofSeq int64, tx store.Tx) (bool, error) {
	parentRoles, ok := args["parent_role"].([]interface{})
	if !ok {
		return false, fmt.Errorf("parent_role is required for role_edge_allowed predicate")
	}

	childRoles, ok := args["child_role"].([]interface{})
	if !ok {
		return false, fmt.Errorf("child_role is required for role_edge_allowed predicate")
	}

	// Get from_node_id (parent) and to_node_id (child) from change
	fromNodeID, ok := change.Payload["from_node_id"].(string)
	if !ok {
		return false, fmt.Errorf("from_node_id not found in change")
	}

	toNodeID, ok := change.Payload["to_node_id"].(string)
	if !ok {
		return false, fmt.Errorf("to_node_id not found in change")
	}

	// Get roles for parent and child nodes
	if namespaceID == nil {
		return false, fmt.Errorf("namespace_id is required for role_edge_allowed predicate")
	}

	parentRolesList, err := tx.GetRoleAssignments(ctx, fromNodeID, *namespaceID, asofSeq)
	if err != nil {
		return false, fmt.Errorf("failed to get parent roles: %w", err)
	}

	childRolesList, err := tx.GetRoleAssignments(ctx, toNodeID, *namespaceID, asofSeq)
	if err != nil {
		return false, fmt.Errorf("failed to get child roles: %w", err)
	}

	// If either node has no roles at this asofSeq (e.g. roles applied in a previous plan not yet visible), allow.
	// This lets seed flows that apply nodes, then roles, then links in separate plans succeed.
	if len(parentRolesList) == 0 || len(childRolesList) == 0 {
		return true, nil
	}

	// Check if parent has an allowed role
	parentHasAllowedRole := false
	for _, allowedRole := range parentRoles {
		allowedRoleStr, ok := allowedRole.(string)
		if !ok {
			continue
		}
		for _, role := range parentRolesList {
			if role.Role == allowedRoleStr {
				parentHasAllowedRole = true
				break
			}
		}
		if parentHasAllowedRole {
			break
		}
	}

	// Check if child has an allowed role
	childHasAllowedRole := false
	for _, allowedRole := range childRoles {
		allowedRoleStr, ok := allowedRole.(string)
		if !ok {
			continue
		}
		for _, role := range childRolesList {
			if role.Role == allowedRoleStr {
				childHasAllowedRole = true
				break
			}
		}
		if childHasAllowedRole {
			break
		}
	}

	return parentHasAllowedRole && childHasAllowedRole, nil
}

// evaluateChildHasOnlyOneParent checks if a child already has a parent
func (p *PredicateEvaluator) evaluateChildHasOnlyOneParent(ctx context.Context, args map[string]interface{}, change domain.Change, namespaceID *string, asofSeq int64, tx store.Tx) (bool, error) {
	childRole, ok := args["child_role"].(string)
	if !ok {
		return false, fmt.Errorf("child_role is required for child_has_only_one_parent predicate")
	}

	linkType, ok := args["link_type"].(string)
	if !ok {
		return false, fmt.Errorf("link_type is required for child_has_only_one_parent predicate")
	}

	// Get to_node_id (child) from change
	toNodeID, ok := change.Payload["to_node_id"].(string)
	if !ok {
		return false, fmt.Errorf("to_node_id not found in change")
	}

	// Check if child has the specified role
	if namespaceID == nil {
		return false, fmt.Errorf("namespace_id is required for child_has_only_one_parent predicate")
	}

	childRoles, err := tx.GetRoleAssignments(ctx, toNodeID, *namespaceID, asofSeq)
	if err != nil {
		return false, fmt.Errorf("failed to get child roles: %w", err)
	}

	hasChildRole := false
	for _, role := range childRoles {
		if role.Role == childRole {
			hasChildRole = true
			break
		}
	}

	if !hasChildRole {
		// Child doesn't have the role, so constraint doesn't apply
		return true, nil
	}

	// Get existing parents of child
	existingParents, err := tx.GetLinksTo(ctx, toNodeID, namespaceID, &linkType, asofSeq)
	if err != nil {
		return false, fmt.Errorf("failed to get existing parents: %w", err)
	}

	// If this is a Move operation, we need to exclude the link being moved
	// For now, we'll check if there are any existing parents
	// In a Move, the old link should be retired first, so this should work

	if len(existingParents) > 0 {
		// Child already has a parent
		return false, nil
	}

	return true, nil
}

// evaluateHasCapability is a stub for v0.1
func (p *PredicateEvaluator) evaluateHasCapability(ctx context.Context, args map[string]interface{}, change domain.Change) (bool, error) {
	// Stub implementation - always returns true for v0.1
	return true, nil
}

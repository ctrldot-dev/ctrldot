package query

import (
	"context"
	"fmt"
	"strings"

	"github.com/futurematic/kernel/internal/domain"
)

// Expand implements the Engine interface
func (e *engine) Expand(ctx context.Context, req ExpandRequest) (*ExpandResult, error) {
	// Open a read transaction
	tx, err := e.store.OpenTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to open transaction: %w", err)
	}
	defer tx.Rollback()

	// Resolve asofSeq if 0 (means latest)
	asofSeq := req.AsOfSeq
	if asofSeq == 0 {
		resolvedSeq, err := e.store.ResolveAsOf(ctx, domain.AsOf{})
		if err != nil {
			return nil, fmt.Errorf("failed to resolve latest sequence: %w", err)
		}
		asofSeq = resolvedSeq
	}

	// Normalize node IDs (handle both node:uuid and uuid formats)
	normalizedNodeIDs := make([]string, len(req.NodeIDs))
	for i, nodeID := range req.NodeIDs {
		normalizedNodeIDs[i] = normalizeNodeID(nodeID)
	}

	result := &ExpandResult{
		Nodes:     []domain.Node{},
		Roles:     []domain.RoleAssignment{},
		Links:     []domain.Link{},
		Materials: []domain.Material{},
	}

	// Track visited nodes to avoid duplicates
	visitedNodes := make(map[string]bool)
	nodesToProcess := make([]string, len(normalizedNodeIDs))
	copy(nodesToProcess, normalizedNodeIDs)

	// Process nodes with depth support
	for depth := 0; depth <= req.Depth && len(nodesToProcess) > 0; depth++ {
		currentLevel := nodesToProcess
		nodesToProcess = []string{}

		for _, nodeID := range currentLevel {
			if visitedNodes[nodeID] {
				continue
			}
			visitedNodes[nodeID] = true

			// Get node (always fetch requested nodes, regardless of roles/links)
			node, err := tx.GetNode(ctx, nodeID, asofSeq)
			if err != nil {
				// Node might not exist at this asof - skip
				continue
			}
			result.Nodes = append(result.Nodes, *node)

			// Get roles if namespace is provided
			if req.NamespaceID != nil {
				roles, err := tx.GetRoleAssignments(ctx, nodeID, *req.NamespaceID, asofSeq)
				if err == nil {
					result.Roles = append(result.Roles, roles...)
				}
			}

			// Get materials
			materials, err := tx.GetMaterialsForNode(ctx, nodeID, asofSeq)
			if err == nil {
				result.Materials = append(result.Materials, materials...)
			}

			// Get links (if depth allows)
			if depth < req.Depth {
				// Get outgoing links
				linksFrom, err := tx.GetLinksFrom(ctx, nodeID, req.NamespaceID, nil, asofSeq)
				if err == nil {
					for _, link := range linksFrom {
						result.Links = append(result.Links, link)
						// Add to_node_id to next level
						if !visitedNodes[link.ToNodeID] {
							nodesToProcess = append(nodesToProcess, link.ToNodeID)
						}
					}
				}

				// Get incoming links
				linksTo, err := tx.GetLinksTo(ctx, nodeID, req.NamespaceID, nil, asofSeq)
				if err == nil {
					for _, link := range linksTo {
						result.Links = append(result.Links, link)
						// Add from_node_id to next level
						if !visitedNodes[link.FromNodeID] {
							nodesToProcess = append(nodesToProcess, link.FromNodeID)
						}
					}
				}
			} else {
				// At max depth, still get links but don't expand nodes
				linksFrom, err := tx.GetLinksFrom(ctx, nodeID, req.NamespaceID, nil, asofSeq)
				if err == nil {
					result.Links = append(result.Links, linksFrom...)
				}

				linksTo, err := tx.GetLinksTo(ctx, nodeID, req.NamespaceID, nil, asofSeq)
				if err == nil {
					result.Links = append(result.Links, linksTo...)
				}
			}
		}
	}

	return result, nil
}

// normalizeNodeID normalizes a node ID to the canonical format
// Accepts both "node:uuid" and "uuid" formats
// Returns the canonical format (with "node:" prefix)
func normalizeNodeID(id string) string {
	// Remove leading/trailing whitespace
	id = strings.TrimSpace(id)
	
	// If it already has "node:" prefix, return as-is
	if strings.HasPrefix(id, "node:") {
		return id
	}
	
	// Otherwise, add "node:" prefix
	return "node:" + id
}

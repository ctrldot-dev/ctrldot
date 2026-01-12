package query

import (
	"context"
	"fmt"

	"github.com/futurematic/kernel/internal/domain"
	"github.com/futurematic/kernel/internal/store"
)

// Diff implements the Engine interface
func (e *engine) Diff(ctx context.Context, req DiffRequest) (*DiffResult, error) {
	// Open a read transaction
	tx, err := e.store.OpenTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to open transaction: %w", err)
	}
	defer tx.Rollback()

	// Get all facts at a_seq
	factsA, err := e.getFactsAtSeq(ctx, tx, req.ASeq, req.Target)
	if err != nil {
		return nil, fmt.Errorf("failed to get facts at seq %d: %w", req.ASeq, err)
	}

	// Get all facts at b_seq
	factsB, err := e.getFactsAtSeq(ctx, tx, req.BSeq, req.Target)
	if err != nil {
		return nil, fmt.Errorf("failed to get facts at seq %d: %w", req.BSeq, err)
	}

	// Compute diff
	changes := e.computeDiff(factsA, factsB)

	return &DiffResult{
		Changes: changes,
	}, nil
}

// FactSet represents a set of facts at a given sequence
type FactSet struct {
	Nodes     map[string]*domain.Node
	Links     map[string]*domain.Link
	Materials map[string]*domain.Material
	Roles     map[string]*domain.RoleAssignment
}

// getFactsAtSeq retrieves all facts visible at a given sequence
func (e *engine) getFactsAtSeq(ctx context.Context, tx store.Tx, seq int64, target string) (*FactSet, error) {
	facts := &FactSet{
		Nodes:     make(map[string]*domain.Node),
		Links:     make(map[string]*domain.Link),
		Materials: make(map[string]*domain.Material),
		Roles:     make(map[string]*domain.RoleAssignment),
	}

	// If target is a node ID, get that node and its related facts
	if target != "" {
		node, err := tx.GetNode(ctx, target, seq)
		if err == nil {
			facts.Nodes[node.ID] = node

			// Get links
			linksFrom, _ := tx.GetLinksFrom(ctx, target, nil, nil, seq)
			for i := range linksFrom {
				facts.Links[linksFrom[i].ID] = &linksFrom[i]
			}

			linksTo, _ := tx.GetLinksTo(ctx, target, nil, nil, seq)
			for i := range linksTo {
				facts.Links[linksTo[i].ID] = &linksTo[i]
			}

			// Get materials
			materials, _ := tx.GetMaterialsForNode(ctx, target, seq)
			for i := range materials {
				facts.Materials[materials[i].ID] = &materials[i]
			}

			// Get roles (if namespace is in target)
			// For now, we'll skip roles as we don't have a way to get all roles for a namespace
		}
	}

	return facts, nil
}

// computeDiff computes the net changes between two fact sets
func (e *engine) computeDiff(factsA, factsB *FactSet) []domain.Change {
	var changes []domain.Change

	// Find nodes that were created or retired
	for id, nodeB := range factsB.Nodes {
		if _, exists := factsA.Nodes[id]; !exists {
			// Node was created
			changes = append(changes, domain.Change{
				Kind: domain.ChangeCreateNode,
				Payload: map[string]interface{}{
					"node_id": nodeB.ID,
					"title":   nodeB.Title,
					"meta":    nodeB.Meta,
				},
			})
		}
	}

	for id := range factsA.Nodes {
		if _, exists := factsB.Nodes[id]; !exists {
			// Node was retired
			changes = append(changes, domain.Change{
				Kind:    domain.ChangeRetireNode,
				Payload: map[string]interface{}{"node_id": id},
			})
		}
	}

	// Find links that were created or retired
	for id, linkB := range factsB.Links {
		if _, exists := factsA.Links[id]; !exists {
			// Link was created
			changes = append(changes, domain.Change{
				Kind:        domain.ChangeCreateLink,
				NamespaceID: linkB.NamespaceID,
				Payload: map[string]interface{}{
					"link_id":      linkB.ID,
					"from_node_id": linkB.FromNodeID,
					"to_node_id":   linkB.ToNodeID,
					"type":         linkB.Type,
					"meta":         linkB.Meta,
				},
			})
		}
	}

	for id := range factsA.Links {
		if _, exists := factsB.Links[id]; !exists {
			// Link was retired
			changes = append(changes, domain.Change{
				Kind:    domain.ChangeRetireLink,
				Payload: map[string]interface{}{"link_id": id},
			})
		}
	}

	// Find materials that were created or retired
	for id, materialB := range factsB.Materials {
		if _, exists := factsA.Materials[id]; !exists {
			// Material was created
			changes = append(changes, domain.Change{
				Kind:        domain.ChangeCreateMaterial,
				NamespaceID: nil,
				Payload: map[string]interface{}{
					"material_id": materialB.ID,
					"node_id":     materialB.NodeID,
					"content_ref": materialB.ContentRef,
					"media_type":  materialB.MediaType,
					"byte_size":   materialB.ByteSize,
					"hash":        materialB.Hash,
					"meta":        materialB.Meta,
				},
			})
		}
	}

	for id := range factsA.Materials {
		if _, exists := factsB.Materials[id]; !exists {
			// Material was retired
			changes = append(changes, domain.Change{
				Kind:    domain.ChangeRetireMaterial,
				Payload: map[string]interface{}{"material_id": id},
			})
		}
	}

	return changes
}

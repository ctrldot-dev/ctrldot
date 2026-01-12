package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/futurematic/kernel/internal/domain"
)

// Helper function to check if a row is visible at a given seq
func isVisibleAtSeq(createdSeq, retiredSeq sql.NullInt64, asofSeq int64) bool {
	if createdSeq.Int64 > asofSeq {
		return false
	}
	if retiredSeq.Valid && retiredSeq.Int64 <= asofSeq {
		return false
	}
	return true
}

// ===== Nodes =====

// CreateNode creates a node in the projection table
func (t *PostgresTx) CreateNode(ctx context.Context, node domain.Node, seq int64) error {
	metaJSON, err := json.Marshal(node.Meta)
	if err != nil {
		return fmt.Errorf("failed to marshal meta: %w", err)
	}

	_, err = t.tx.ExecContext(ctx,
		`INSERT INTO nodes (node_id, title, meta_json, created_seq)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (node_id) DO UPDATE SET
		   title = EXCLUDED.title,
		   meta_json = EXCLUDED.meta_json,
		   created_seq = EXCLUDED.created_seq,
		   retired_seq = NULL`,
		node.ID, node.Title, metaJSON, seq,
	)
	if err != nil {
		return fmt.Errorf("failed to insert node: %w", err)
	}
	return nil
}

// GetNode retrieves a node by ID at a given sequence
func (t *PostgresTx) GetNode(ctx context.Context, nodeID string, asofSeq int64) (*domain.Node, error) {
	var node domain.Node
	var metaJSON []byte
	var createdSeq, retiredSeq sql.NullInt64

	err := t.tx.QueryRowContext(ctx,
		`SELECT node_id, title, meta_json, created_seq, retired_seq
		 FROM nodes WHERE node_id = $1`,
		nodeID,
	).Scan(&node.ID, &node.Title, &metaJSON, &createdSeq, &retiredSeq)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("node not found")
		}
		return nil, fmt.Errorf("failed to get node: %w", err)
	}

	// Check visibility at asofSeq
	if !isVisibleAtSeq(createdSeq, retiredSeq, asofSeq) {
		return nil, fmt.Errorf("node not visible at seq %d", asofSeq)
	}

	if err := json.Unmarshal(metaJSON, &node.Meta); err != nil {
		node.Meta = make(map[string]interface{})
	}

	return &node, nil
}

// RetireNode retires a node by setting retired_seq
func (t *PostgresTx) RetireNode(ctx context.Context, nodeID string, seq int64) error {
	_, err := t.tx.ExecContext(ctx,
		`UPDATE nodes SET retired_seq = $1 WHERE node_id = $2 AND retired_seq IS NULL`,
		seq, nodeID,
	)
	if err != nil {
		return fmt.Errorf("failed to retire node: %w", err)
	}
	return nil
}

// ===== Links =====

// CreateLink creates a link in the projection table
func (t *PostgresTx) CreateLink(ctx context.Context, link domain.Link, seq int64) error {
	metaJSON, err := json.Marshal(link.Meta)
	if err != nil {
		return fmt.Errorf("failed to marshal meta: %w", err)
	}

	_, err = t.tx.ExecContext(ctx,
		`INSERT INTO links (link_id, from_node_id, to_node_id, type, namespace_id, meta_json, created_seq)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 ON CONFLICT (link_id) DO UPDATE SET
		   from_node_id = EXCLUDED.from_node_id,
		   to_node_id = EXCLUDED.to_node_id,
		   type = EXCLUDED.type,
		   namespace_id = EXCLUDED.namespace_id,
		   meta_json = EXCLUDED.meta_json,
		   created_seq = EXCLUDED.created_seq,
		   retired_seq = NULL`,
		link.ID, link.FromNodeID, link.ToNodeID, link.Type, link.NamespaceID, metaJSON, seq,
	)
	if err != nil {
		return fmt.Errorf("failed to insert link: %w", err)
	}
	return nil
}

// GetLink retrieves a link by ID at a given sequence
func (t *PostgresTx) GetLink(ctx context.Context, linkID string, asofSeq int64) (*domain.Link, error) {
	var link domain.Link
	var metaJSON []byte
	var namespaceID sql.NullString
	var createdSeq, retiredSeq sql.NullInt64

	err := t.tx.QueryRowContext(ctx,
		`SELECT link_id, from_node_id, to_node_id, type, namespace_id, meta_json, created_seq, retired_seq
		 FROM links WHERE link_id = $1`,
		linkID,
	).Scan(&link.ID, &link.FromNodeID, &link.ToNodeID, &link.Type, &namespaceID, &metaJSON, &createdSeq, &retiredSeq)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("link not found")
		}
		return nil, fmt.Errorf("failed to get link: %w", err)
	}

	if namespaceID.Valid {
		link.NamespaceID = &namespaceID.String
	}

	if !isVisibleAtSeq(createdSeq, retiredSeq, asofSeq) {
		return nil, fmt.Errorf("link not visible at seq %d", asofSeq)
	}

	if err := json.Unmarshal(metaJSON, &link.Meta); err != nil {
		link.Meta = make(map[string]interface{})
	}

	return &link, nil
}

// GetLinksFrom retrieves links from a node at a given sequence
func (t *PostgresTx) GetLinksFrom(ctx context.Context, fromNodeID string, namespaceID *string, linkType *string, asofSeq int64) ([]domain.Link, error) {
	query := `SELECT link_id, from_node_id, to_node_id, type, namespace_id, meta_json, created_seq, retired_seq
			  FROM links
			  WHERE from_node_id = $1
			    AND created_seq <= $2
			    AND (retired_seq IS NULL OR retired_seq > $2)`
	args := []interface{}{fromNodeID, asofSeq}
	argIdx := 3

	if namespaceID != nil {
		query += fmt.Sprintf(" AND namespace_id = $%d", argIdx)
		args = append(args, *namespaceID)
		argIdx++
	}

	if linkType != nil {
		query += fmt.Sprintf(" AND type = $%d", argIdx)
		args = append(args, *linkType)
		argIdx++
	}

	rows, err := t.tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query links: %w", err)
	}
	defer rows.Close()

	var links []domain.Link
	for rows.Next() {
		var link domain.Link
		var metaJSON []byte
		var nsID sql.NullString
		var createdSeq, retiredSeq sql.NullInt64

		if err := rows.Scan(&link.ID, &link.FromNodeID, &link.ToNodeID, &link.Type, &nsID, &metaJSON, &createdSeq, &retiredSeq); err != nil {
			return nil, fmt.Errorf("failed to scan link: %w", err)
		}

		if nsID.Valid {
			link.NamespaceID = &nsID.String
		}

		if err := json.Unmarshal(metaJSON, &link.Meta); err != nil {
			link.Meta = make(map[string]interface{})
		}

		links = append(links, link)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating links: %w", err)
	}

	return links, nil
}

// GetLinksTo retrieves links to a node at a given sequence
func (t *PostgresTx) GetLinksTo(ctx context.Context, toNodeID string, namespaceID *string, linkType *string, asofSeq int64) ([]domain.Link, error) {
	query := `SELECT link_id, from_node_id, to_node_id, type, namespace_id, meta_json, created_seq, retired_seq
			  FROM links
			  WHERE to_node_id = $1
			    AND created_seq <= $2
			    AND (retired_seq IS NULL OR retired_seq > $2)`
	args := []interface{}{toNodeID, asofSeq}
	argIdx := 3

	if namespaceID != nil {
		query += fmt.Sprintf(" AND namespace_id = $%d", argIdx)
		args = append(args, *namespaceID)
		argIdx++
	}

	if linkType != nil {
		query += fmt.Sprintf(" AND type = $%d", argIdx)
		args = append(args, *linkType)
		argIdx++
	}

	rows, err := t.tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query links: %w", err)
	}
	defer rows.Close()

	var links []domain.Link
	for rows.Next() {
		var link domain.Link
		var metaJSON []byte
		var nsID sql.NullString
		var createdSeq, retiredSeq sql.NullInt64

		if err := rows.Scan(&link.ID, &link.FromNodeID, &link.ToNodeID, &link.Type, &nsID, &metaJSON, &createdSeq, &retiredSeq); err != nil {
			return nil, fmt.Errorf("failed to scan link: %w", err)
		}

		if nsID.Valid {
			link.NamespaceID = &nsID.String
		}

		if err := json.Unmarshal(metaJSON, &link.Meta); err != nil {
			link.Meta = make(map[string]interface{})
		}

		links = append(links, link)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating links: %w", err)
	}

	return links, nil
}

// RetireLink retires a link by setting retired_seq
func (t *PostgresTx) RetireLink(ctx context.Context, linkID string, seq int64) error {
	_, err := t.tx.ExecContext(ctx,
		`UPDATE links SET retired_seq = $1 WHERE link_id = $2 AND retired_seq IS NULL`,
		seq, linkID,
	)
	if err != nil {
		return fmt.Errorf("failed to retire link: %w", err)
	}
	return nil
}

// ===== Materials =====

// CreateMaterial creates a material in the projection table
func (t *PostgresTx) CreateMaterial(ctx context.Context, material domain.Material, seq int64) error {
	metaJSON, err := json.Marshal(material.Meta)
	if err != nil {
		return fmt.Errorf("failed to marshal meta: %w", err)
	}

	_, err = t.tx.ExecContext(ctx,
		`INSERT INTO materials (material_id, node_id, content_ref, media_type, byte_size, hash, meta_json, created_seq)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		 ON CONFLICT (material_id) DO UPDATE SET
		   node_id = EXCLUDED.node_id,
		   content_ref = EXCLUDED.content_ref,
		   media_type = EXCLUDED.media_type,
		   byte_size = EXCLUDED.byte_size,
		   hash = EXCLUDED.hash,
		   meta_json = EXCLUDED.meta_json,
		   created_seq = EXCLUDED.created_seq,
		   retired_seq = NULL`,
		material.ID, material.NodeID, material.ContentRef, material.MediaType, material.ByteSize, material.Hash, metaJSON, seq,
	)
	if err != nil {
		return fmt.Errorf("failed to insert material: %w", err)
	}
	return nil
}

// GetMaterial retrieves a material by ID at a given sequence
func (t *PostgresTx) GetMaterial(ctx context.Context, materialID string, asofSeq int64) (*domain.Material, error) {
	var material domain.Material
	var metaJSON []byte
	var hash sql.NullString
	var createdSeq, retiredSeq sql.NullInt64

	err := t.tx.QueryRowContext(ctx,
		`SELECT material_id, node_id, content_ref, media_type, byte_size, hash, meta_json, created_seq, retired_seq
		 FROM materials WHERE material_id = $1`,
		materialID,
	).Scan(&material.ID, &material.NodeID, &material.ContentRef, &material.MediaType, &material.ByteSize, &hash, &metaJSON, &createdSeq, &retiredSeq)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("material not found")
		}
		return nil, fmt.Errorf("failed to get material: %w", err)
	}

	if hash.Valid {
		material.Hash = &hash.String
	}

	if !isVisibleAtSeq(createdSeq, retiredSeq, asofSeq) {
		return nil, fmt.Errorf("material not visible at seq %d", asofSeq)
	}

	if err := json.Unmarshal(metaJSON, &material.Meta); err != nil {
		material.Meta = make(map[string]interface{})
	}

	return &material, nil
}

// GetMaterialsForNode retrieves all materials for a node at a given sequence
func (t *PostgresTx) GetMaterialsForNode(ctx context.Context, nodeID string, asofSeq int64) ([]domain.Material, error) {
	rows, err := t.tx.QueryContext(ctx,
		`SELECT material_id, node_id, content_ref, media_type, byte_size, hash, meta_json, created_seq, retired_seq
		 FROM materials
		 WHERE node_id = $1
		   AND created_seq <= $2
		   AND (retired_seq IS NULL OR retired_seq > $2)`,
		nodeID, asofSeq,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query materials: %w", err)
	}
	defer rows.Close()

	var materials []domain.Material
	for rows.Next() {
		var material domain.Material
		var metaJSON []byte
		var hash sql.NullString
		var createdSeq, retiredSeq sql.NullInt64

		if err := rows.Scan(&material.ID, &material.NodeID, &material.ContentRef, &material.MediaType, &material.ByteSize, &hash, &metaJSON, &createdSeq, &retiredSeq); err != nil {
			return nil, fmt.Errorf("failed to scan material: %w", err)
		}

		if hash.Valid {
			material.Hash = &hash.String
		}

		if err := json.Unmarshal(metaJSON, &material.Meta); err != nil {
			material.Meta = make(map[string]interface{})
		}

		materials = append(materials, material)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating materials: %w", err)
	}

	return materials, nil
}

// RetireMaterial retires a material by setting retired_seq
func (t *PostgresTx) RetireMaterial(ctx context.Context, materialID string, seq int64) error {
	_, err := t.tx.ExecContext(ctx,
		`UPDATE materials SET retired_seq = $1 WHERE material_id = $2 AND retired_seq IS NULL`,
		seq, materialID,
	)
	if err != nil {
		return fmt.Errorf("failed to retire material: %w", err)
	}
	return nil
}

// ===== Role Assignments =====

// CreateRoleAssignment creates a role assignment in the projection table
func (t *PostgresTx) CreateRoleAssignment(ctx context.Context, role domain.RoleAssignment, seq int64) error {
	metaJSON, err := json.Marshal(role.Meta)
	if err != nil {
		return fmt.Errorf("failed to marshal meta: %w", err)
	}

	_, err = t.tx.ExecContext(ctx,
		`INSERT INTO role_assignments (role_assignment_id, node_id, namespace_id, role, meta_json, created_seq)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 ON CONFLICT (role_assignment_id) DO UPDATE SET
		   node_id = EXCLUDED.node_id,
		   namespace_id = EXCLUDED.namespace_id,
		   role = EXCLUDED.role,
		   meta_json = EXCLUDED.meta_json,
		   created_seq = EXCLUDED.created_seq,
		   retired_seq = NULL`,
		role.ID, role.NodeID, role.NamespaceID, role.Role, metaJSON, seq,
	)
	if err != nil {
		return fmt.Errorf("failed to insert role assignment: %w", err)
	}
	return nil
}

// GetRoleAssignments retrieves role assignments for a node in a namespace at a given sequence
func (t *PostgresTx) GetRoleAssignments(ctx context.Context, nodeID string, namespaceID string, asofSeq int64) ([]domain.RoleAssignment, error) {
	rows, err := t.tx.QueryContext(ctx,
		`SELECT role_assignment_id, node_id, namespace_id, role, meta_json, created_seq, retired_seq
		 FROM role_assignments
		 WHERE node_id = $1 AND namespace_id = $2
		   AND created_seq <= $3
		   AND (retired_seq IS NULL OR retired_seq > $3)`,
		nodeID, namespaceID, asofSeq,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query role assignments: %w", err)
	}
	defer rows.Close()

	var roles []domain.RoleAssignment
	for rows.Next() {
		var role domain.RoleAssignment
		var metaJSON []byte
		var createdSeq, retiredSeq sql.NullInt64

		if err := rows.Scan(&role.ID, &role.NodeID, &role.NamespaceID, &role.Role, &metaJSON, &createdSeq, &retiredSeq); err != nil {
			return nil, fmt.Errorf("failed to scan role assignment: %w", err)
		}

		if err := json.Unmarshal(metaJSON, &role.Meta); err != nil {
			role.Meta = make(map[string]interface{})
		}

		roles = append(roles, role)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating role assignments: %w", err)
	}

	return roles, nil
}

// RetireRoleAssignment retires a role assignment by setting retired_seq
func (t *PostgresTx) RetireRoleAssignment(ctx context.Context, roleAssignmentID string, seq int64) error {
	_, err := t.tx.ExecContext(ctx,
		`UPDATE role_assignments SET retired_seq = $1 WHERE role_assignment_id = $2 AND retired_seq IS NULL`,
		seq, roleAssignmentID,
	)
	if err != nil {
		return fmt.Errorf("failed to retire role assignment: %w", err)
	}
	return nil
}

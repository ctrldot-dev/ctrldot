package domain

// RoleAssignment represents contextual typing of a node within a namespace
type RoleAssignment struct {
	ID          string                 `json:"role_assignment_id"`
	NodeID      string                 `json:"node_id"`
	NamespaceID string                 `json:"namespace_id"`
	Role        string                 `json:"role"`
	Meta        map[string]interface{} `json:"meta,omitempty"`
}

// Validate checks if the role assignment is valid
func (r RoleAssignment) Validate() error {
	if r.ID == "" {
		return ErrInvalidRoleAssignmentID
	}
	if r.NodeID == "" {
		return ErrInvalidNodeID
	}
	if r.NamespaceID == "" {
		return ErrInvalidNamespaceID
	}
	if r.Role == "" {
		return ErrInvalidRole
	}
	return nil
}

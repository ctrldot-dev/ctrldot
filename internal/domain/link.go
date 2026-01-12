package domain

// Link represents a typed relationship between nodes
type Link struct {
	ID          string                 `json:"link_id"`
	FromNodeID  string                 `json:"from_node_id"`
	ToNodeID    string                 `json:"to_node_id"`
	Type        string                 `json:"type"`
	NamespaceID *string                `json:"namespace_id,omitempty"`
	Meta        map[string]interface{} `json:"meta,omitempty"`
}

// Validate checks if the link is valid
func (l Link) Validate() error {
	if l.ID == "" {
		return ErrInvalidLinkID
	}
	if l.FromNodeID == "" {
		return ErrInvalidFromNodeID
	}
	if l.ToNodeID == "" {
		return ErrInvalidToNodeID
	}
	if l.Type == "" {
		return ErrInvalidLinkType
	}
	if l.FromNodeID == l.ToNodeID {
		return ErrSelfLink
	}
	return nil
}

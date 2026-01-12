package domain

// Node represents a primitive identity with metadata
type Node struct {
	ID    string                 `json:"node_id"`
	Title string                 `json:"title"`
	Meta  map[string]interface{} `json:"meta,omitempty"`
}

// Validate checks if the node is valid
func (n Node) Validate() error {
	if n.ID == "" {
		return ErrInvalidNodeID
	}
	if n.Title == "" {
		return ErrInvalidNodeTitle
	}
	return nil
}

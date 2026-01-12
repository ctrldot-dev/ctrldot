package domain

// Namespace represents a context (e.g., ProductTree:/MachinePay)
type Namespace struct {
	ID   string `json:"namespace_id"`
	Name string `json:"name"`
}

// Validate checks if the namespace is valid
func (n Namespace) Validate() error {
	if n.ID == "" {
		return ErrInvalidNamespaceID
	}
	if n.Name == "" {
		return ErrInvalidNamespaceName
	}
	return nil
}

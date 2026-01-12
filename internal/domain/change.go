package domain

// Change represents an atomic change to the system
// Changes are expanded from Intents and applied atomically
type Change struct {
	Kind        string                 `json:"kind"`
	NamespaceID *string                `json:"namespace_id,omitempty"`
	Payload     map[string]interface{} `json:"payload"`
}

// Change kinds (atomic operations)
const (
	ChangeCreateNode          = "CreateNode"
	ChangeCreateLink          = "CreateLink"
	ChangeCreateMaterial      = "CreateMaterial"
	ChangeCreateRoleAssignment = "CreateRoleAssignment"
	ChangeRetireNode          = "RetireNode"
	ChangeRetireLink          = "RetireLink"
	ChangeRetireMaterial      = "RetireMaterial"
	ChangeRetireRoleAssignment = "RetireRoleAssignment"
)

// Validate checks if the change is valid
func (c Change) Validate() error {
	if c.Kind == "" {
		return ErrInvalidChangeKind
	}
	validKinds := map[string]bool{
		ChangeCreateNode:           true,
		ChangeCreateLink:           true,
		ChangeCreateMaterial:       true,
		ChangeCreateRoleAssignment: true,
		ChangeRetireNode:           true,
		ChangeRetireLink:           true,
		ChangeRetireMaterial:       true,
		ChangeRetireRoleAssignment: true,
	}
	if !validKinds[c.Kind] {
		return ErrInvalidChangeKind
	}
	return nil
}

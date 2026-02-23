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
	// Kernel domain changes
	ChangeCreateNode          = "CreateNode"
	ChangeCreateLink          = "CreateLink"
	ChangeCreateMaterial      = "CreateMaterial"
	ChangeCreateRoleAssignment = "CreateRoleAssignment"
	ChangeRetireNode          = "RetireNode"
	ChangeRetireLink          = "RetireLink"
	ChangeRetireMaterial      = "RetireMaterial"
	ChangeRetireRoleAssignment = "RetireRoleAssignment"
	
	// Ctrl Dot domain changes
	ChangeCreateAgent         = "CreateAgent"
	ChangeCreateSession       = "CreateSession"
	ChangeEndSession          = "EndSession"
	ChangeAppendEvent         = "AppendEvent"
	ChangeUpdateLimitsState   = "UpdateLimitsState"
	ChangeHaltAgent           = "HaltAgent"
	ChangeResumeAgent         = "ResumeAgent"
)

// Validate checks if the change is valid
func (c Change) Validate() error {
	if c.Kind == "" {
		return ErrInvalidChangeKind
	}
	validKinds := map[string]bool{
		// Kernel domain
		ChangeCreateNode:           true,
		ChangeCreateLink:           true,
		ChangeCreateMaterial:       true,
		ChangeCreateRoleAssignment: true,
		ChangeRetireNode:           true,
		ChangeRetireLink:           true,
		ChangeRetireMaterial:       true,
		ChangeRetireRoleAssignment: true,
		// Ctrl Dot domain
		ChangeCreateAgent:       true,
		ChangeCreateSession:     true,
		ChangeEndSession:        true,
		ChangeAppendEvent:       true,
		ChangeUpdateLimitsState: true,
		ChangeHaltAgent:         true,
		ChangeResumeAgent:       true,
	}
	if !validKinds[c.Kind] {
		return ErrInvalidChangeKind
	}
	return nil
}

package domain

// Intent represents a proposed change to the system
// Intents are expanded into atomic Changes by the planner
type Intent struct {
	Kind        string                 `json:"kind"`
	NamespaceID *string                `json:"namespace_id,omitempty"`
	Payload     map[string]interface{} `json:"payload"`
}

// Intent kinds
const (
	// Kernel domain intents
	IntentCreateNode    = "CreateNode"
	IntentCreateLink    = "CreateLink"
	IntentCreateMaterial = "CreateMaterial"
	IntentAssignRole    = "AssignRole"
	IntentRetireNode    = "RetireNode"
	IntentRetireLink    = "RetireLink"
	IntentRetireMaterial = "RetireMaterial"
	IntentMove          = "Move"
	
	// Ctrl Dot domain intents
	IntentRegisterAgent = "RegisterAgent"
	IntentProposeAction = "ProposeAction"
	IntentCreateSession = "CreateSession"
	IntentEndSession    = "EndSession"
	IntentAppendEvent   = "AppendEvent"
	IntentUpdateLimitsState = "UpdateLimitsState"
	IntentHaltAgent     = "HaltAgent"
	IntentResumeAgent   = "ResumeAgent"
)

// Validate checks if the intent is valid
func (i Intent) Validate() error {
	if i.Kind == "" {
		return ErrInvalidIntentKind
	}
	validKinds := map[string]bool{
		// Kernel domain
		IntentCreateNode:     true,
		IntentCreateLink:     true,
		IntentCreateMaterial: true,
		IntentAssignRole:     true,
		IntentRetireNode:     true,
		IntentRetireLink:     true,
		IntentRetireMaterial: true,
		IntentMove:           true,
		// Ctrl Dot domain
		IntentRegisterAgent:  true,
		IntentProposeAction:  true,
		IntentCreateSession:  true,
		IntentEndSession:     true,
		IntentAppendEvent:    true,
		IntentUpdateLimitsState: true,
		IntentHaltAgent:      true,
		IntentResumeAgent:    true,
	}
	if !validKinds[i.Kind] {
		return ErrInvalidIntentKind
	}
	return nil
}

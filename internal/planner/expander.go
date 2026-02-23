package planner

import (
	"context"
	"fmt"

	"github.com/futurematic/kernel/internal/domain"
	"github.com/futurematic/kernel/internal/store"
	"github.com/google/uuid"
)

// Expander expands intents into atomic changes
type Expander struct{}

// NewExpander creates a new expander
func NewExpander() *Expander {
	return &Expander{}
}

// Expand expands intents into atomic changes
func (e *Expander) Expand(ctx context.Context, intents []domain.Intent, namespaceID *string, asofSeq int64, store store.Store) ([]domain.Change, error) {
	var changes []domain.Change

	for _, intent := range intents {
		expanded, err := e.expandIntent(ctx, intent, namespaceID, asofSeq, store)
		if err != nil {
			return nil, fmt.Errorf("failed to expand intent %s: %w", intent.Kind, err)
		}
		changes = append(changes, expanded...)
	}

	return changes, nil
}

// expandIntent expands a single intent into one or more changes
func (e *Expander) expandIntent(ctx context.Context, intent domain.Intent, namespaceID *string, asofSeq int64, store store.Store) ([]domain.Change, error) {
	switch intent.Kind {
	case domain.IntentCreateNode:
		return e.expandCreateNode(intent, namespaceID)
	case domain.IntentCreateLink:
		return e.expandCreateLink(intent, namespaceID, asofSeq, store)
	case domain.IntentCreateMaterial:
		return e.expandCreateMaterial(intent, namespaceID, asofSeq, store)
	case domain.IntentAssignRole:
		return e.expandAssignRole(intent, namespaceID)
	case domain.IntentRetireNode:
		return e.expandRetireNode(intent)
	case domain.IntentRetireLink:
		return e.expandRetireLink(intent)
	case domain.IntentRetireMaterial:
		return e.expandRetireMaterial(intent)
	case domain.IntentMove:
		return e.expandMove(intent, namespaceID, asofSeq, store)
	// Ctrl Dot domain intents
	case domain.IntentAppendEvent:
		return e.expandAppendEvent(intent)
	case domain.IntentRegisterAgent:
		return e.expandRegisterAgent(intent)
	case domain.IntentCreateSession:
		return e.expandCreateSession(intent)
	case domain.IntentEndSession:
		return e.expandEndSession(intent)
	case domain.IntentUpdateLimitsState:
		return e.expandUpdateLimitsState(intent)
	case domain.IntentHaltAgent:
		return e.expandHaltAgent(intent)
	case domain.IntentResumeAgent:
		return e.expandResumeAgent(intent)
	default:
		return nil, fmt.Errorf("unknown intent kind: %s", intent.Kind)
	}
}

func (e *Expander) expandCreateNode(intent domain.Intent, namespaceID *string) ([]domain.Change, error) {
	// Extract node ID from payload or generate one
	nodeID, ok := intent.Payload["node_id"].(string)
	if !ok || nodeID == "" {
		nodeID = "node:" + uuid.New().String()
	}

	title, ok := intent.Payload["title"].(string)
	if !ok || title == "" {
		return nil, fmt.Errorf("title is required for CreateNode")
	}

	meta, _ := intent.Payload["meta"].(map[string]interface{})
	if meta == nil {
		meta = make(map[string]interface{})
	}

	change := domain.Change{
		Kind:        domain.ChangeCreateNode,
		NamespaceID: namespaceID,
		Payload: map[string]interface{}{
			"node_id": nodeID,
			"title":   title,
			"meta":    meta,
		},
	}

	return []domain.Change{change}, nil
}

func (e *Expander) expandCreateLink(intent domain.Intent, namespaceID *string, asofSeq int64, store store.Store) ([]domain.Change, error) {
	// Extract link fields
	fromNodeID, ok := intent.Payload["from_node_id"].(string)
	if !ok || fromNodeID == "" {
		return nil, fmt.Errorf("from_node_id is required for CreateLink")
	}

	toNodeID, ok := intent.Payload["to_node_id"].(string)
	if !ok || toNodeID == "" {
		return nil, fmt.Errorf("to_node_id is required for CreateLink")
	}

	linkType, ok := intent.Payload["type"].(string)
	if !ok || linkType == "" {
		return nil, fmt.Errorf("type is required for CreateLink")
	}

	// Use namespace from intent if provided, otherwise use default
	nsID := namespaceID
	if intent.NamespaceID != nil {
		nsID = intent.NamespaceID
	}

	// Check if nodes exist (simplified - in real implementation, we'd check via store)
	// For now, we'll assume nodes exist or will be created

	linkID, ok := intent.Payload["link_id"].(string)
	if !ok || linkID == "" {
		linkID = "link:" + uuid.New().String()
	}

	meta, _ := intent.Payload["meta"].(map[string]interface{})
	if meta == nil {
		meta = make(map[string]interface{})
	}

	change := domain.Change{
		Kind:        domain.ChangeCreateLink,
		NamespaceID: nsID,
		Payload: map[string]interface{}{
			"link_id":      linkID,
			"from_node_id": fromNodeID,
			"to_node_id":   toNodeID,
			"type":         linkType,
			"meta":         meta,
		},
	}

	return []domain.Change{change}, nil
}

func (e *Expander) expandCreateMaterial(intent domain.Intent, namespaceID *string, asofSeq int64, store store.Store) ([]domain.Change, error) {
	nodeID, ok := intent.Payload["node_id"].(string)
	if !ok || nodeID == "" {
		return nil, fmt.Errorf("node_id is required for CreateMaterial")
	}

	contentRef, ok := intent.Payload["content_ref"].(string)
	if !ok || contentRef == "" {
		return nil, fmt.Errorf("content_ref is required for CreateMaterial")
	}

	mediaType, ok := intent.Payload["media_type"].(string)
	if !ok || mediaType == "" {
		return nil, fmt.Errorf("media_type is required for CreateMaterial")
	}

	byteSize, ok := intent.Payload["byte_size"].(float64) // JSON numbers are float64
	if !ok {
		return nil, fmt.Errorf("byte_size is required for CreateMaterial")
	}

	materialID, ok := intent.Payload["material_id"].(string)
	if !ok || materialID == "" {
		materialID = "material:" + uuid.New().String()
	}

	hash, _ := intent.Payload["hash"].(string)
	meta, _ := intent.Payload["meta"].(map[string]interface{})
	if meta == nil {
		meta = make(map[string]interface{})
	}

	payload := map[string]interface{}{
		"material_id": materialID,
		"node_id":     nodeID,
		"content_ref": contentRef,
		"media_type":  mediaType,
		"byte_size":   int64(byteSize),
		"meta":        meta,
	}
	if hash != "" {
		payload["hash"] = hash
	}

	change := domain.Change{
		Kind:        domain.ChangeCreateMaterial,
		NamespaceID: namespaceID,
		Payload:     payload,
	}

	return []domain.Change{change}, nil
}

func (e *Expander) expandAssignRole(intent domain.Intent, namespaceID *string) ([]domain.Change, error) {
	nodeID, ok := intent.Payload["node_id"].(string)
	if !ok || nodeID == "" {
		return nil, fmt.Errorf("node_id is required for AssignRole")
	}

	role, ok := intent.Payload["role"].(string)
	if !ok || role == "" {
		return nil, fmt.Errorf("role is required for AssignRole")
	}

	nsID := namespaceID
	if intent.NamespaceID != nil {
		nsID = intent.NamespaceID
	}
	if nsID == nil {
		return nil, fmt.Errorf("namespace_id is required for AssignRole")
	}

	roleAssignmentID, ok := intent.Payload["role_assignment_id"].(string)
	if !ok || roleAssignmentID == "" {
		roleAssignmentID = "role:" + uuid.New().String()
	}

	meta, _ := intent.Payload["meta"].(map[string]interface{})
	if meta == nil {
		meta = make(map[string]interface{})
	}

	change := domain.Change{
		Kind:        domain.ChangeCreateRoleAssignment,
		NamespaceID: nsID,
		Payload: map[string]interface{}{
			"role_assignment_id": roleAssignmentID,
			"node_id":            nodeID,
			"namespace_id":       *nsID,
			"role":               role,
			"meta":               meta,
		},
	}

	return []domain.Change{change}, nil
}

func (e *Expander) expandRetireNode(intent domain.Intent) ([]domain.Change, error) {
	nodeID, ok := intent.Payload["node_id"].(string)
	if !ok || nodeID == "" {
		return nil, fmt.Errorf("node_id is required for RetireNode")
	}

	change := domain.Change{
		Kind:    domain.ChangeRetireNode,
		Payload: map[string]interface{}{"node_id": nodeID},
	}

	return []domain.Change{change}, nil
}

func (e *Expander) expandRetireLink(intent domain.Intent) ([]domain.Change, error) {
	linkID, ok := intent.Payload["link_id"].(string)
	if !ok || linkID == "" {
		return nil, fmt.Errorf("link_id is required for RetireLink")
	}

	change := domain.Change{
		Kind:    domain.ChangeRetireLink,
		Payload: map[string]interface{}{"link_id": linkID},
	}

	return []domain.Change{change}, nil
}

func (e *Expander) expandRetireMaterial(intent domain.Intent) ([]domain.Change, error) {
	materialID, ok := intent.Payload["material_id"].(string)
	if !ok || materialID == "" {
		return nil, fmt.Errorf("material_id is required for RetireMaterial")
	}

	change := domain.Change{
		Kind:    domain.ChangeRetireMaterial,
		Payload: map[string]interface{}{"material_id": materialID},
	}

	return []domain.Change{change}, nil
}

func (e *Expander) expandMove(intent domain.Intent, namespaceID *string, asofSeq int64, store store.Store) ([]domain.Change, error) {
	// Move is: RetireLink (old parent) + CreateLink (new parent)
	linkID, ok := intent.Payload["link_id"].(string)
	if !ok || linkID == "" {
		return nil, fmt.Errorf("link_id is required for Move")
	}

	toNodeID, ok := intent.Payload["to_node_id"].(string)
	if !ok || toNodeID == "" {
		return nil, fmt.Errorf("to_node_id is required for Move")
	}

	// Get the existing link to find from_node_id and type
	// For now, we'll require these in the payload
	fromNodeID, ok := intent.Payload["from_node_id"].(string)
	if !ok || fromNodeID == "" {
		return nil, fmt.Errorf("from_node_id is required for Move")
	}

	linkType, ok := intent.Payload["type"].(string)
	if !ok || linkType == "" {
		return nil, fmt.Errorf("type is required for Move")
	}

	nsID := namespaceID
	if intent.NamespaceID != nil {
		nsID = intent.NamespaceID
	}

	// Retire old link
	retireChange := domain.Change{
		Kind:    domain.ChangeRetireLink,
		Payload: map[string]interface{}{"link_id": linkID},
	}

	// Create new link (with same ID or new ID?)
	// For Move, we typically keep the same link ID but update the to_node_id
	// However, since we're retiring and creating, we might need a new link ID
	newLinkID := "link:" + uuid.New().String()
	meta, _ := intent.Payload["meta"].(map[string]interface{})
	if meta == nil {
		meta = make(map[string]interface{})
	}

	createChange := domain.Change{
		Kind:        domain.ChangeCreateLink,
		NamespaceID: nsID,
		Payload: map[string]interface{}{
			"link_id":      newLinkID,
			"from_node_id": fromNodeID,
			"to_node_id":   toNodeID,
			"type":         linkType,
			"meta":         meta,
		},
	}

	return []domain.Change{retireChange, createChange}, nil
}

// Ctrl Dot intent expansions

func (e *Expander) expandAppendEvent(intent domain.Intent) ([]domain.Change, error) {
	// AppendEvent intent directly maps to ChangeAppendEvent
	change := domain.Change{
		Kind:    domain.ChangeAppendEvent,
		Payload: intent.Payload,
	}
	return []domain.Change{change}, nil
}

func (e *Expander) expandRegisterAgent(intent domain.Intent) ([]domain.Change, error) {
	change := domain.Change{
		Kind:    domain.ChangeCreateAgent,
		Payload: intent.Payload,
	}
	return []domain.Change{change}, nil
}

func (e *Expander) expandCreateSession(intent domain.Intent) ([]domain.Change, error) {
	change := domain.Change{
		Kind:    domain.ChangeCreateSession,
		Payload: intent.Payload,
	}
	return []domain.Change{change}, nil
}

func (e *Expander) expandEndSession(intent domain.Intent) ([]domain.Change, error) {
	change := domain.Change{
		Kind:    domain.ChangeEndSession,
		Payload: intent.Payload,
	}
	return []domain.Change{change}, nil
}

func (e *Expander) expandUpdateLimitsState(intent domain.Intent) ([]domain.Change, error) {
	change := domain.Change{
		Kind:    domain.ChangeUpdateLimitsState,
		Payload: intent.Payload,
	}
	return []domain.Change{change}, nil
}

func (e *Expander) expandHaltAgent(intent domain.Intent) ([]domain.Change, error) {
	change := domain.Change{
		Kind:    domain.ChangeHaltAgent,
		Payload: intent.Payload,
	}
	return []domain.Change{change}, nil
}

func (e *Expander) expandResumeAgent(intent domain.Intent) ([]domain.Change, error) {
	change := domain.Change{
		Kind:    domain.ChangeResumeAgent,
		Payload: intent.Payload,
	}
	return []domain.Change{change}, nil
}

package output

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/futurematic/kernel/internal/domain"
)

// Formatter is an interface for output formatting
type Formatter interface {
	PrintPlan(plan interface{}) error
	PrintOperation(op interface{}) error
	PrintExpand(result interface{}) error
	PrintHistory(ops []domain.Operation) error
	PrintDiff(result interface{}) error
	PrintStatus(status interface{}) error
	PrintConfig(cfg interface{}) error
}

// NewFormatter creates a new formatter based on format type
func NewFormatter(format string, w io.Writer) Formatter {
	if format == "json" {
		return &JSONFormatter{w: w}
	}
	return &TextFormatter{w: w}
}

// Helper function to format time
func formatTime(t time.Time) string {
	return t.Format(time.RFC3339)
}

// Helper function to format change
func formatChange(change domain.Change) string {
	var parts []string
	parts = append(parts, change.Kind)

	switch change.Kind {
	case domain.ChangeCreateNode:
		if nodeID, ok := change.Payload["node_id"].(string); ok {
			parts = append(parts, nodeID)
		}
		if title, ok := change.Payload["title"].(string); ok {
			parts = append(parts, fmt.Sprintf("title=%q", title))
		}
	case domain.ChangeCreateLink:
		if from, ok := change.Payload["from_node_id"].(string); ok {
			parts = append(parts, fmt.Sprintf("from=%s", from))
		}
		if to, ok := change.Payload["to_node_id"].(string); ok {
			parts = append(parts, fmt.Sprintf("to=%s", to))
		}
		if linkType, ok := change.Payload["type"].(string); ok {
			parts = append(parts, fmt.Sprintf("type=%s", linkType))
		}
	case domain.ChangeCreateMaterial:
		if materialID, ok := change.Payload["material_id"].(string); ok {
			parts = append(parts, materialID)
		}
	case domain.ChangeCreateRoleAssignment:
		if nodeID, ok := change.Payload["node_id"].(string); ok {
			parts = append(parts, nodeID)
		}
		if role, ok := change.Payload["role"].(string); ok {
			parts = append(parts, fmt.Sprintf("role=%s", role))
		}
	case domain.ChangeRetireNode, domain.ChangeRetireLink, domain.ChangeRetireMaterial:
		if id, ok := change.Payload["node_id"].(string); ok {
			parts = append(parts, id)
		} else if id, ok := change.Payload["link_id"].(string); ok {
			parts = append(parts, id)
		} else if id, ok := change.Payload["material_id"].(string); ok {
			parts = append(parts, id)
		}
	}

	return strings.Join(parts, " ")
}

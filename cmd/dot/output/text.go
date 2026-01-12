package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/futurematic/kernel/cmd/dot/client"
	"github.com/futurematic/kernel/cmd/dot/config"
	"github.com/futurematic/kernel/internal/domain"
)

// TextFormatter formats output as human-readable text
type TextFormatter struct {
	w io.Writer
}

func (f *TextFormatter) PrintPlan(plan interface{}) error {
	// Handle different plan types (could be from client or direct domain)
	var p struct {
		ID           string                 `json:"id"`
		Hash         string                 `json:"hash"`
		Class        int                    `json:"class"`
		PolicyReport domain.PolicyReport    `json:"policy_report"`
		Expanded     []domain.Change        `json:"expanded"`
	}

	// Try type assertion or JSON unmarshal
	switch v := plan.(type) {
	case *client.PlanResponse:
		p.ID = v.ID
		p.Hash = v.Hash
		p.Class = v.Class
		p.PolicyReport = v.PolicyReport
		p.Expanded = v.Expanded
	default:
		// Try JSON marshal/unmarshal as fallback
		data, err := json.Marshal(plan)
		if err != nil {
			return fmt.Errorf("invalid plan type: %T", plan)
		}
		if err := json.Unmarshal(data, &p); err != nil {
			return fmt.Errorf("failed to parse plan: %w", err)
		}
	}

	fmt.Fprintf(f.w, "PLAN %s hash=%s class=%d\n", p.ID, p.Hash, p.Class)

	// Print policy report
	if len(p.PolicyReport.Denies) > 0 {
		fmt.Fprintf(f.w, "DENIES:\n")
		for _, deny := range p.PolicyReport.Denies {
			fmt.Fprintf(f.w, "  - %s: %s\n", deny.RuleID, deny.Message)
		}
	}
	if len(p.PolicyReport.Warns) > 0 {
		fmt.Fprintf(f.w, "WARNS:\n")
		for _, warn := range p.PolicyReport.Warns {
			fmt.Fprintf(f.w, "  - %s: %s\n", warn.RuleID, warn.Message)
		}
	}
	if len(p.PolicyReport.Infos) > 0 {
		fmt.Fprintf(f.w, "INFOS:\n")
		for _, info := range p.PolicyReport.Infos {
			fmt.Fprintf(f.w, "  - %s: %s\n", info.RuleID, info.Message)
		}
	}

	// Print changes
	if len(p.Expanded) > 0 {
		fmt.Fprintf(f.w, "CHANGES:\n")
		for _, change := range p.Expanded {
			fmt.Fprintf(f.w, "  + %s\n", formatChange(change))
		}
	}

	return nil
}

func (f *TextFormatter) PrintOperation(op interface{}) error {
	var o struct {
		ID         string          `json:"id"`
		Seq        int64           `json:"seq"`
		OccurredAt time.Time       `json:"occurred_at"`
		Changes    []domain.Change `json:"changes"`
	}

	// Try type assertion or JSON unmarshal
	switch v := op.(type) {
	case *client.ApplyResponse:
		o.ID = v.ID
		o.Seq = v.Seq
		o.OccurredAt = v.OccurredAt
		o.Changes = v.Changes
	default:
		// Try JSON marshal/unmarshal as fallback
		data, err := json.Marshal(op)
		if err != nil {
			return fmt.Errorf("invalid operation type: %T", op)
		}
		if err := json.Unmarshal(data, &o); err != nil {
			return fmt.Errorf("failed to parse operation: %w", err)
		}
	}

	fmt.Fprintf(f.w, "APPLIED %s seq=%d occurred_at=%s\n", o.ID, o.Seq, o.OccurredAt.Format(time.RFC3339))
	if len(o.Changes) > 0 {
		fmt.Fprintf(f.w, "CHANGES:\n")
		for _, change := range o.Changes {
			fmt.Fprintf(f.w, "  + %s\n", formatChange(change))
		}
	}

	return nil
}

func (f *TextFormatter) PrintExpand(result interface{}) error {
	var r struct {
		Nodes           []domain.Node           `json:"nodes"`
		Links           []domain.Link           `json:"links"`
		Materials       []domain.Material       `json:"materials"`
		RoleAssignments []domain.RoleAssignment  `json:"role_assignments"`
	}

	// Try type assertion or JSON unmarshal
	switch v := result.(type) {
	case *client.ExpandResponse:
		r.Nodes = v.Nodes
		r.Links = v.Links
		r.Materials = v.Materials
		r.RoleAssignments = v.RoleAssignments
	default:
		// Try JSON marshal/unmarshal as fallback
		data, err := json.Marshal(result)
		if err != nil {
			return fmt.Errorf("invalid expand result type: %T", result)
		}
		if err := json.Unmarshal(data, &r); err != nil {
			return fmt.Errorf("failed to parse expand result: %w", err)
		}
	}

	// Print nodes
	for _, node := range r.Nodes {
		fmt.Fprintf(f.w, "NODE %s title=%q\n", node.ID, node.Title)
	}

	// Print roles
	if len(r.RoleAssignments) > 0 {
		fmt.Fprintf(f.w, "ROLES:\n")
		for _, role := range r.RoleAssignments {
			fmt.Fprintf(f.w, "  %s: %s\n", role.NodeID, role.Role)
		}
	}

	// Print links grouped by type
	linksByType := make(map[string][]domain.Link)
	for _, link := range r.Links {
		linksByType[link.Type] = append(linksByType[link.Type], link)
	}

	for linkType, links := range linksByType {
		fmt.Fprintf(f.w, "LINKS %s:\n", linkType)
		for _, link := range links {
			fmt.Fprintf(f.w, "  %s -> %s\n", link.FromNodeID, link.ToNodeID)
		}
	}

	// Print materials summary
	if len(r.Materials) > 0 {
		fmt.Fprintf(f.w, "MATERIALS: %d\n", len(r.Materials))
	}

	return nil
}

func (f *TextFormatter) PrintHistory(ops []domain.Operation) error {
	for _, op := range ops {
		fmt.Fprintf(f.w, "seq=%d op=%s occurred_at=%s actor=%s class=%d\n",
			op.Seq, op.ID, op.OccurredAt.Format(time.RFC3339), op.ActorID, op.Class)
		if len(op.Changes) > 0 {
			for _, change := range op.Changes {
				fmt.Fprintf(f.w, "  %s\n", formatChange(change))
			}
		}
	}
	return nil
}

func (f *TextFormatter) PrintDiff(result interface{}) error {
	var r struct {
		Changes []domain.Change `json:"changes"`
	}

	// Try type assertion or JSON unmarshal
	switch v := result.(type) {
	case *client.DiffResponse:
		r.Changes = v.Changes
	default:
		// Try JSON marshal/unmarshal as fallback
		data, err := json.Marshal(result)
		if err != nil {
			return fmt.Errorf("invalid diff result type: %T", result)
		}
		if err := json.Unmarshal(data, &r); err != nil {
			return fmt.Errorf("failed to parse diff result: %w", err)
		}
	}

	fmt.Fprintf(f.w, "DIFF:\n")
	for _, change := range r.Changes {
		fmt.Fprintf(f.w, "  %s\n", formatChange(change))
	}

	return nil
}

func (f *TextFormatter) PrintStatus(status interface{}) error {
	var s struct {
		OK          bool   `json:"ok"`
		Server      string `json:"server"`
		ActorID     string `json:"actor_id"`
		NamespaceID string `json:"namespace_id"`
	}

	// Try type assertion or JSON unmarshal
	switch v := status.(type) {
	case *struct {
		OK          bool
		Server      string
		ActorID     string
		NamespaceID string
	}:
		s.OK = v.OK
		s.Server = v.Server
		s.ActorID = v.ActorID
		s.NamespaceID = v.NamespaceID
	default:
		// Try JSON marshal/unmarshal as fallback
		data, err := json.Marshal(status)
		if err != nil {
			return fmt.Errorf("invalid status type: %T", status)
		}
		if err := json.Unmarshal(data, &s); err != nil {
			return fmt.Errorf("failed to parse status: %w", err)
		}
	}

	fmt.Fprintf(f.w, "server=%s actor=%s namespace=%s ok=%v\n",
		s.Server, s.ActorID, s.NamespaceID, s.OK)
	return nil
}

func (f *TextFormatter) PrintConfig(cfg interface{}) error {
	var c struct {
		Server      string   `json:"server"`
		ActorID     string   `json:"actor_id"`
		NamespaceID string   `json:"namespace_id"`
		Capabilities []string `json:"capabilities"`
	}

	// Try type assertion or JSON unmarshal
	switch v := cfg.(type) {
	case *config.Config:
		c.Server = v.Server
		c.ActorID = v.ActorID
		c.NamespaceID = v.NamespaceID
		c.Capabilities = v.Capabilities
	default:
		// Try JSON marshal/unmarshal as fallback
		data, err := json.Marshal(cfg)
		if err != nil {
			return fmt.Errorf("invalid config type: %T", cfg)
		}
		if err := json.Unmarshal(data, &c); err != nil {
			return fmt.Errorf("failed to parse config: %w", err)
		}
	}

	fmt.Fprintf(f.w, "server=%s\n", c.Server)
	fmt.Fprintf(f.w, "actor_id=%s\n", c.ActorID)
	fmt.Fprintf(f.w, "namespace_id=%s\n", c.NamespaceID)
	fmt.Fprintf(f.w, "capabilities=%s\n", strings.Join(c.Capabilities, ","))
	return nil
}

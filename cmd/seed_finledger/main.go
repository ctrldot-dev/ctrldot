// seed_finledger ingests FinLedger seed YAML graphs into the kernel via plan/apply.
// Usage: go run ./cmd/seed_finledger [kernel_url] < kesteron-treasury-finledger-seed.yaml
//
//	Or: go run ./cmd/seed_finledger http://localhost:8080 kesteron-treasury-finledger-seed.yaml kesteron-stablecoinreserves-finledger-seed.yaml
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/futurematic/kernel/cmd/dot/client"
	"github.com/futurematic/kernel/internal/domain"
	"gopkg.in/yaml.v3"
)

type seedNode struct {
	Title string `yaml:"title"`
	Role  string `yaml:"role"`
}

type seedLink struct {
	From string `yaml:"from"`
	To   string `yaml:"to"`
	Type string `yaml:"type"`
}

type seedGraph struct {
	Namespace string     `yaml:"namespace"`
	Nodes     []seedNode `yaml:"nodes"`
	Links     []seedLink `yaml:"links"`
}

func main() {
	kernelURL := os.Getenv("KERNEL_URL")
	if kernelURL == "" {
		kernelURL = "http://localhost:8080"
	}
	var files []string
	for _, a := range os.Args[1:] {
		if strings.HasPrefix(a, "http://") || strings.HasPrefix(a, "https://") {
			kernelURL = a
		} else {
			files = append(files, a)
		}
	}
	if len(files) == 0 {
		files = []string{"kesteron-treasury-finledger-seed.yaml", "kesteron-stablecoinreserves-finledger-seed.yaml"}
	}

	c := client.NewClient(kernelURL)
	if _, err := c.Healthz(); err != nil {
		fmt.Fprintf(os.Stderr, "kernel not reachable at %s: %v\n", kernelURL, err)
		os.Exit(1)
	}

	for _, path := range files {
		rootID, err := ingest(c, path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ingest %s: %v\n", path, err)
			os.Exit(1)
		}
		fmt.Printf("Ingested %s (root for tree: %s)\n", path, rootID)
	}
}

func ingest(c *client.Client, path string) (rootNodeID string, err error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	var g seedGraph
	if err := yaml.Unmarshal(data, &g); err != nil {
		return "", fmt.Errorf("parse YAML: %w", err)
	}
	if g.Namespace == "" || len(g.Nodes) == 0 {
		return "", fmt.Errorf("namespace and nodes required")
	}
	ns := &g.Namespace
	actorID := "system:seed"
	caps := []string{"read", "write:additive"}

	// 1) Create all nodes (one plan)
	var nodeIntents []domain.Intent
	for _, n := range g.Nodes {
		nodeIntents = append(nodeIntents, domain.Intent{
			Kind:        domain.IntentCreateNode,
			NamespaceID: ns,
			Payload:     map[string]interface{}{"title": n.Title},
		})
	}
	planResp, err := c.Plan(client.PlanRequest{
		ActorID:      actorID,
		Capabilities: caps,
		NamespaceID:  ns,
		AsOf:         domain.AsOf{},
		Intents:      nodeIntents,
	})
	if err != nil {
		return "", fmt.Errorf("plan nodes: %w", err)
	}
	applyResp, err := c.Apply(client.ApplyRequest{
		ActorID:      actorID,
		Capabilities: caps,
		PlanID:       planResp.ID,
		PlanHash:     planResp.Hash,
	})
	if err != nil {
		return "", fmt.Errorf("apply nodes: %w", err)
	}
	titleToID := make(map[string]string)
	for _, ch := range applyResp.Changes {
		if ch.Kind != domain.ChangeCreateNode {
			continue
		}
		title, _ := ch.Payload["title"].(string)
		nodeID, _ := ch.Payload["node_id"].(string)
		if title != "" && nodeID != "" {
			titleToID[title] = nodeID
		}
	}

	// 2) Assign roles (one plan)
	var roleIntents []domain.Intent
	for _, n := range g.Nodes {
		nodeID, ok := titleToID[n.Title]
		if !ok {
			return "", fmt.Errorf("node not found for role: %s", n.Title)
		}
		roleIntents = append(roleIntents, domain.Intent{
			Kind:        domain.IntentAssignRole,
			NamespaceID: ns,
			Payload: map[string]interface{}{
				"node_id":      nodeID,
				"namespace_id": g.Namespace,
				"role":         n.Role,
			},
		})
	}
	if len(roleIntents) > 0 {
		planResp, err = c.Plan(client.PlanRequest{
			ActorID:      actorID,
			Capabilities: caps,
			NamespaceID:  ns,
			AsOf:         domain.AsOf{},
			Intents:      roleIntents,
		})
		if err != nil {
			return "", fmt.Errorf("plan roles: %w", err)
		}
		if _, err = c.Apply(client.ApplyRequest{
			ActorID:      actorID,
			Capabilities: caps,
			PlanID:       planResp.ID,
			PlanHash:     planResp.Hash,
		}); err != nil {
			return "", fmt.Errorf("apply roles: %w", err)
		}
	}

	// 3) Create links (one plan)
	var linkIntents []domain.Intent
	for _, l := range g.Links {
		fromID, ok := titleToID[l.From]
		if !ok {
			return "", fmt.Errorf("from node not found: %s", l.From)
		}
		toID, ok := titleToID[l.To]
		if !ok {
			return "", fmt.Errorf("to node not found: %s", l.To)
		}
		linkIntents = append(linkIntents, domain.Intent{
			Kind:        domain.IntentCreateLink,
			NamespaceID: ns,
			Payload: map[string]interface{}{
				"from_node_id": fromID,
				"to_node_id":   toID,
				"type":         l.Type,
			},
		})
	}
	// Root for tree view is first node in seed (e.g. "Kesteron Treasury Root")
	if len(g.Nodes) > 0 {
		rootNodeID = titleToID[g.Nodes[0].Title]
	}
	if len(linkIntents) == 0 {
		return rootNodeID, nil
	}
	planResp, err = c.Plan(client.PlanRequest{
		ActorID:      actorID,
		Capabilities: caps,
		NamespaceID:  ns,
		AsOf:         domain.AsOf{},
		Intents:      linkIntents,
	})
	if err != nil {
		return "", fmt.Errorf("plan links: %w", err)
	}
	if _, err = c.Apply(client.ApplyRequest{
		ActorID:      actorID,
		Capabilities: caps,
		PlanID:       planResp.ID,
		PlanHash:     planResp.Hash,
	}); err != nil {
		return "", fmt.Errorf("apply links: %w", err)
	}
	return rootNodeID, nil
}

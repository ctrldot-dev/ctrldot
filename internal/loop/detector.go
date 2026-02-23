package loop

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"

	"github.com/futurematic/kernel/internal/config"
	"github.com/futurematic/kernel/internal/domain"
	"github.com/futurematic/kernel/internal/runtime"
)

// Detector detects action loops
type Detector struct {
	store  runtime.RuntimeStore
	config *config.Config
}

// NewDetector creates a new loop detector
func NewDetector(store runtime.RuntimeStore, cfg *config.Config) *Detector {
	return &Detector{
		store:  store,
		config: cfg,
	}
}

// Detect detects if an action is part of a loop (uses engine config).
func (d *Detector) Detect(ctx context.Context, proposal domain.ActionProposal) bool {
	return d.DetectWithConfig(ctx, proposal, d.config)
}

// DetectWithConfig detects if an action is part of a loop using the given config.
// When cfg.Loop is set (e.g. panic overlay), uses that window and repeat count; otherwise default 10min/25 and 60s/10.
func (d *Detector) DetectWithConfig(ctx context.Context, proposal domain.ActionProposal, cfg *config.Config) bool {
	if cfg == nil {
		cfg = d.config
	}
	// Compute action hash if not provided
	actionHash := proposal.Context.Hash
	if actionHash == "" {
		actionHash = d.computeActionHash(proposal)
	}

	// When Loop overlay is set (panic mode), use that window and stop count
	if cfg.Loop != nil && cfg.Loop.WindowSeconds > 0 && cfg.Loop.StopRepeats > 0 {
		sinceTS := time.Now().Add(-time.Duration(cfg.Loop.WindowSeconds) * time.Second).Unix() * 1000
		events, err := d.store.ListEvents(ctx, runtime.EventFilter{AgentID: &proposal.AgentID, SinceTS: &sinceTS, Limit: 100})
		if err != nil {
			return false
		}
		count := 0
		for _, event := range events {
			if event.ActionHash == actionHash {
				count++
			}
		}
		return count >= cfg.Loop.StopRepeats
	}

	// Default: check events in last 10 minutes
	sinceTS := time.Now().Add(-10 * time.Minute).Unix() * 1000
	events, err := d.store.ListEvents(ctx, runtime.EventFilter{AgentID: &proposal.AgentID, SinceTS: &sinceTS, Limit: 100})
	if err != nil {
		return false // Can't check, allow
	}

	count := 0
	for _, event := range events {
		if event.ActionHash == actionHash {
			count++
		}
	}

	maxIterations := cfg.Agents.Default.MaxIterationsPerAction
	if maxIterations <= 0 {
		maxIterations = 25
	}
	if count >= maxIterations {
		return true
	}

	// Hard-coded safety net: 10 within 60s
	sinceTS60s := time.Now().Add(-60 * time.Second).Unix() * 1000
	events60s, err := d.store.ListEvents(ctx, runtime.EventFilter{AgentID: &proposal.AgentID, SinceTS: &sinceTS60s, Limit: 100})
	if err == nil {
		count60s := 0
		for _, event := range events60s {
			if event.ActionHash == actionHash {
				count60s++
			}
		}
		if count60s >= 10 {
			return true
		}
	}

	return false
}

func (d *Detector) computeActionHash(proposal domain.ActionProposal) string {
	// Hash of agent_id + action.type + target + stable inputs
	data := map[string]interface{}{
		"agent_id":   proposal.AgentID,
		"action_type": proposal.Action.Type,
		"target":     proposal.Action.Target,
		"inputs":     proposal.Action.Inputs,
	}

	jsonData, _ := json.Marshal(data)
	hash := sha256.Sum256(jsonData)
	return hex.EncodeToString(hash[:])
}

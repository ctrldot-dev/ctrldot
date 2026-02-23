package config

import (
	"time"

	"github.com/futurematic/kernel/internal/domain"
)

// USD to GBP approximate; used only for panic budget clamp when config uses GBP.
const usdToGBP = 0.79

// Effective returns a config with panic overlay applied when panic is enabled.
// Caller can pass the result to engines for this request. Does not modify base.
func Effective(base *Config, panicState *domain.PanicState) *Config {
	if base == nil {
		return nil
	}
	if panicState == nil || !panicState.Enabled {
		return base
	}
	// Clone base and apply panic overlay
	out := *base
	out.Agents = cloneAgentsConfig(base.Agents)
	out.Rules = cloneRulesConfig(base.Rules)

	// Budget clamp: min(agent default, panic max) in GBP
	panicBudgetGBP := base.Panic.MaxDailyBudgetUSD * usdToGBP
	if out.Agents.Default.DailyBudgetGBP <= 0 || out.Agents.Default.DailyBudgetGBP > panicBudgetGBP {
		out.Agents.Default.DailyBudgetGBP = panicBudgetGBP
	}
	// Panic thresholds
	if base.Panic.Thresholds.WarnPct > 0 {
		out.Agents.Default.WarnPct = []float64{base.Panic.Thresholds.WarnPct}
	}
	if base.Panic.Thresholds.ThrottlePct > 0 {
		out.Agents.Default.ThrottlePct = base.Panic.Thresholds.ThrottlePct
	}
	if base.Panic.Thresholds.StopPct > 0 {
		out.Agents.Default.HardStopPct = base.Panic.Thresholds.StopPct
	}

	// Resolution: force require for all non–safe-read actions
	if base.Panic.Resolution.ForceRequireResolution {
		out.Rules.RequireResolution = []string{"git.push", "filesystem.delete", "filesystem.write", "tool.call", "exec", "network.", "http.", "web."}
	}
	// Filesystem: restrict to panic workspace roots when set
	if len(base.Panic.Filesystem.WorkspaceRoots) > 0 {
		out.Rules.Filesystem.AllowRoots = base.Panic.Filesystem.WorkspaceRoots
	} else if base.Panic.Filesystem.Mode == "read_only" {
		out.Rules.Filesystem.AllowRoots = nil
	}
	// Network: default deny + allowlist
	if base.Panic.Network.DefaultDeny {
		out.Rules.Network.DenyAll = true
		if len(base.Panic.Network.AllowDomains) > 0 {
			out.Rules.Network.AllowDomains = base.Panic.Network.AllowDomains
		}
	}

	// Loop: tighter window and repeat count when panic on
	windowSec := base.Panic.Loop.WindowSeconds
	if windowSec <= 0 {
		windowSec = 60
	}
	stopRepeats := base.Panic.Loop.StopRepeats
	if stopRepeats <= 0 {
		stopRepeats = 5
	}
	out.Loop = &LoopOverlay{WindowSeconds: windowSec, StopRepeats: stopRepeats}
	out.Agents.Default.MaxIterationsPerAction = stopRepeats

	return &out
}

func cloneAgentsConfig(a AgentsConfig) AgentsConfig {
	out := AgentsConfig{Default: a.Default}
	if len(a.Default.WarnPct) > 0 {
		out.Default.WarnPct = make([]float64, len(a.Default.WarnPct))
		copy(out.Default.WarnPct, a.Default.WarnPct)
	}
	return out
}

func cloneRulesConfig(r RulesConfig) RulesConfig {
	out := RulesConfig{
		RequireResolution: make([]string, len(r.RequireResolution)),
		Filesystem:        r.Filesystem,
		Network:           r.Network,
	}
	copy(out.RequireResolution, r.RequireResolution)
	if len(r.Filesystem.AllowRoots) > 0 {
		out.Filesystem.AllowRoots = make([]string, len(r.Filesystem.AllowRoots))
		copy(out.Filesystem.AllowRoots, r.Filesystem.AllowRoots)
	}
	if len(r.Network.AllowDomains) > 0 {
		out.Network.AllowDomains = make([]string, len(r.Network.AllowDomains))
		copy(out.Network.AllowDomains, r.Network.AllowDomains)
	}
	return out
}

// PanicExpired returns true if panic is enabled but past expires_at.
func PanicExpired(panicState *domain.PanicState) bool {
	if panicState == nil || !panicState.Enabled || panicState.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*panicState.ExpiresAt)
}

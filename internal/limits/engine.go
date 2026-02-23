package limits

import (
	"context"
	"fmt"
	"time"

	"github.com/futurematic/kernel/internal/config"
	"github.com/futurematic/kernel/internal/domain"
	"github.com/futurematic/kernel/internal/runtime"
)

// Engine evaluates budget and limits
type Engine struct {
	store  runtime.RuntimeStore
	config *config.Config
}

// NewEngine creates a new limits engine
func NewEngine(store runtime.RuntimeStore, cfg *config.Config) *Engine {
	return &Engine{
		store:  store,
		config: cfg,
	}
}

// Evaluate evaluates limits and returns decision, warnings, and throttle info (uses engine config).
func (e *Engine) Evaluate(ctx context.Context, proposal domain.ActionProposal, agent *domain.Agent) (domain.Decision, []domain.Warning, *domain.ThrottleInfo) {
	return e.EvaluateWithConfig(ctx, proposal, agent, e.config)
}

// EvaluateWithConfig evaluates limits using the given config (e.g. effective config when panic is on).
func (e *Engine) EvaluateWithConfig(ctx context.Context, proposal domain.ActionProposal, agent *domain.Agent, cfg *config.Config) (domain.Decision, []domain.Warning, *domain.ThrottleInfo) {
	if cfg == nil {
		cfg = e.config
	}
	// Get current window (daily)
	now := time.Now()
	windowStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	windowStartTS := windowStart.Unix() * 1000

	// Get limits state (store returns nil, nil when no row exists)
	state, err := e.store.GetLimitsState(ctx, proposal.AgentID, windowStartTS, "daily")
	if err != nil || state == nil {
		state = &domain.LimitsState{
			AgentID:          proposal.AgentID,
			WindowStart:      windowStart,
			WindowType:       "daily",
			BudgetSpentGBP:   0,
			BudgetSpentTokens: 0,
			ActionCount:     0,
		}
	}

	var defaults config.AgentDefaults
	if cfg != nil {
		defaults = cfg.Agents.Default
	} else {
		defaults = config.AgentDefaults{
			DailyBudgetGBP:        10.0,
			WarnPct:                []float64{0.70, 0.90},
			ThrottlePct:            0.95,
			HardStopPct:            1.00,
			MaxIterationsPerAction: 25,
		}
	}
	budgetLimit := defaults.DailyBudgetGBP
	if budgetLimit <= 0 {
		budgetLimit = 10.0
	}

	// Calculate new totals
	newBudgetSpent := state.BudgetSpentGBP + proposal.Cost.EstimatedGBP
	budgetPct := newBudgetSpent / budgetLimit

	var warnings []domain.Warning
	var throttle *domain.ThrottleInfo

	// Check thresholds
	if budgetPct >= defaults.HardStopPct {
		return domain.DecisionStop, warnings, nil
	}

	if budgetPct >= defaults.ThrottlePct {
		throttle = &domain.ThrottleInfo{
			MaxParallelTasks: e.config.DegradeModes.Cheap.MaxParallelTasks,
			ModelPolicy:      e.config.DegradeModes.Cheap.ModelPolicy,
			ToolRestrictions: e.config.DegradeModes.Cheap.DenyTools,
		}
		if cfg != nil {
			throttle.MaxParallelTasks = cfg.DegradeModes.Cheap.MaxParallelTasks
			throttle.ModelPolicy = cfg.DegradeModes.Cheap.ModelPolicy
			throttle.ToolRestrictions = cfg.DegradeModes.Cheap.DenyTools
		}
		return domain.DecisionThrottle, warnings, throttle
	}

	// Check warn thresholds
	for _, warnPct := range defaults.WarnPct {
		if budgetPct >= warnPct && budgetPct < warnPct+0.01 { // Only warn once per threshold
			warnings = append(warnings, domain.Warning{
				Code:    fmt.Sprintf("BUDGET_%d", int(warnPct*100)),
				Message: fmt.Sprintf("Agent at %.0f%% of daily budget (£%.2f/£%.2f).", budgetPct*100, newBudgetSpent, budgetLimit),
			})
		}
	}

	if len(warnings) > 0 {
		return domain.DecisionWarn, warnings, nil
	}

	return domain.DecisionAllow, warnings, nil
}

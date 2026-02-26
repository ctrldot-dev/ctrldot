package ctrldot

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/futurematic/kernel/internal/config"
	"github.com/futurematic/kernel/internal/ctrldot/recommendations"
	"github.com/futurematic/kernel/internal/domain"
	"github.com/futurematic/kernel/internal/ledger/autobundle"
	"github.com/futurematic/kernel/internal/ledger/sink"
	"github.com/futurematic/kernel/internal/loop"
	"github.com/futurematic/kernel/internal/limits"
	"github.com/futurematic/kernel/internal/resolution"
	"github.com/futurematic/kernel/internal/rules"
	"github.com/futurematic/kernel/internal/runtime"
	"github.com/google/uuid"
)

// Service provides Ctrl Dot functionality
type Service interface {
	// RegisterAgent registers a new agent
	RegisterAgent(ctx context.Context, agentID string, displayName string, defaultMode string) (*domain.Agent, error)

	// ProposeAction evaluates an action proposal and returns a decision
	ProposeAction(ctx context.Context, proposal domain.ActionProposal) (*domain.DecisionResponse, error)

	// StartSession starts a new session for an agent
	StartSession(ctx context.Context, agentID string, metadata map[string]interface{}) (*domain.Session, error)

	// EndSession ends a session
	EndSession(ctx context.Context, sessionID string) error

	// GetEvents retrieves events
	GetEvents(ctx context.Context, agentID *string, sinceTS *int64, limit int) ([]domain.Event, error)

	// ListAgents lists all agents
	ListAgents(ctx context.Context) ([]domain.Agent, error)

	// GetAgent retrieves an agent
	GetAgent(ctx context.Context, agentID string) (*domain.Agent, error)

	// HaltAgent halts an agent
	HaltAgent(ctx context.Context, agentID string, reason string) error

	// ResumeAgent resumes a halted agent
	ResumeAgent(ctx context.Context, agentID string) error

	// GetPanicState returns the current panic mode state
	GetPanicState(ctx context.Context) (*domain.PanicState, error)
	// SetPanicState updates panic mode state (persisted in RuntimeStore)
	SetPanicState(ctx context.Context, state domain.PanicState) error

	// GetAutobundleStatus returns current autobundle config (enabled, output_dir, triggers, etc.)
	GetAutobundleStatus(ctx context.Context) (*config.AutobundleConfig, error)

	// GetCapabilities returns agent-discovery capabilities (no secrets; paths expanded).
	GetCapabilities(ctx context.Context) (*domain.CapabilitiesResponse, error)

	// GetAgentLimits returns current budget/limits state for an agent (daily window).
	GetAgentLimits(ctx context.Context, agentID string) (*domain.AgentLimitsResponse, error)

	// GetLimitsConfig returns default limits from config (read-only view).
	GetLimitsConfig(ctx context.Context) (*domain.LimitsConfigResponse, error)
}

// service implements Service
type service struct {
	runtimeStore   runtime.RuntimeStore
	limitsEngine   *limits.Engine
	rulesEngine    *rules.Engine
	loopDetector   *loop.Detector
	resolutionMgr  *resolution.Manager
	ledgerSink     sink.LedgerSink
	autobundleMgr  *autobundle.Manager
	config         *config.Config
}

// NewService creates a new Ctrl Dot service using RuntimeStore for all runtime state.
// autobundleMgr may be nil to disable auto-bundles.
func NewService(
	runtimeStore runtime.RuntimeStore,
	limitsEngine *limits.Engine,
	rulesEngine *rules.Engine,
	loopDetector *loop.Detector,
	resolutionMgr *resolution.Manager,
	ledgerSink sink.LedgerSink,
	autobundleMgr *autobundle.Manager,
	cfg *config.Config,
) Service {
	return &service{
		runtimeStore:  runtimeStore,
		limitsEngine: limitsEngine,
		rulesEngine:  rulesEngine,
		loopDetector: loopDetector,
		resolutionMgr: resolutionMgr,
		ledgerSink:   ledgerSink,
		autobundleMgr: autobundleMgr,
		config:       cfg,
	}
}

// RegisterAgent registers a new agent
func (s *service) RegisterAgent(ctx context.Context, agentID string, displayName string, defaultMode string) (*domain.Agent, error) {
	if defaultMode == "" {
		defaultMode = domain.AgentModeNormal
	}

	agent := domain.Agent{
		AgentID:     agentID,
		DisplayName: displayName,
		CreatedAt:   time.Now(),
		DefaultMode: defaultMode,
	}

	if err := s.runtimeStore.CreateAgent(ctx, agent); err != nil {
		return nil, fmt.Errorf("failed to create agent: %w", err)
	}

	eventID := "evt:" + uuid.New().String()
	event := domain.Event{
		EventID:     eventID,
		TS:          time.Now(),
		Type:        domain.EventTypeAgentRegistered,
		AgentID:     agentID,
		Severity:    domain.EventSeverityInfo,
		PayloadJSON: map[string]interface{}{
			"agent_id":     agentID,
			"display_name": displayName,
			"default_mode": defaultMode,
		},
	}
	if err := s.runtimeStore.AppendEvent(ctx, &event); err != nil {
		return nil, fmt.Errorf("failed to append event: %w", err)
	}

	return &agent, nil
}

// ProposeAction evaluates an action proposal and returns a decision
func (s *service) ProposeAction(ctx context.Context, proposal domain.ActionProposal) (*domain.DecisionResponse, error) {
	agent, err := s.runtimeStore.GetAgent(ctx, proposal.AgentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get agent: %w", err)
	}
	if agent == nil {
		return &domain.DecisionResponse{
			Decision: domain.DecisionDeny,
			Reason:   "Agent not registered",
		}, nil
	}

	halted, err := s.runtimeStore.IsAgentHalted(ctx, proposal.AgentID)
	if err != nil {
		return nil, fmt.Errorf("failed to check halted status: %w", err)
	}
	if halted {
		return &domain.DecisionResponse{
			Decision: domain.DecisionStop,
			Reason:   "Agent is halted",
		}, nil
	}

	// Load panic state and apply TTL auto-disable
	panicState, err := s.runtimeStore.GetPanicState(ctx)
	if err != nil {
		panicState = nil
	}
	if config.PanicExpired(panicState) {
		disabled := *panicState
		disabled.Enabled = false
		_ = s.runtimeStore.SetPanicState(ctx, disabled)
		panicState = &disabled
	}
	effectiveConfig := config.Effective(s.config, panicState)

	ruleDecision, ruleReason := s.rulesEngine.EvaluateWithConfig(ctx, proposal, effectiveConfig)
	loopStop := s.loopDetector.DetectWithConfig(ctx, proposal, effectiveConfig)
	limitDecision, warnings, throttle := s.limitsEngine.EvaluateWithConfig(ctx, proposal, agent, effectiveConfig)

	finalDecision := ruleDecision
	responseReason := ruleReason
	if ruleDecision == domain.DecisionDeny {
		finalDecision = domain.DecisionDeny
		responseReason = ruleReason
	} else if loopStop {
		finalDecision = domain.DecisionStop
		responseReason = "Loop detected: repeated action"
	} else if limitDecision == domain.DecisionStop || limitDecision == domain.DecisionDeny {
		finalDecision = limitDecision
		if limitDecision == domain.DecisionStop {
			responseReason = "Budget limit reached"
		}
	} else if limitDecision == domain.DecisionThrottle && finalDecision == domain.DecisionAllow {
		finalDecision = domain.DecisionThrottle
	} else if limitDecision == domain.DecisionWarn && finalDecision == domain.DecisionAllow {
		finalDecision = domain.DecisionWarn
	}

	eventID := "evt:" + uuid.New().String()
	decisionEvent := domain.Event{
		EventID:     eventID,
		TS:          time.Now(),
		Type:        domain.EventTypeDecisionIssued,
		AgentID:     proposal.AgentID,
		SessionID:   proposal.SessionID,
		Severity:    domain.EventSeverityInfo,
		PayloadJSON: map[string]interface{}{
			"decision":    string(finalDecision),
			"action_type": proposal.Action.Type,
			"action_hash": proposal.Context.Hash,
		},
		ActionHash: proposal.Context.Hash,
		CostGBP:    &proposal.Cost.EstimatedGBP,
		CostTokens: &proposal.Cost.EstimatedTokens,
	}
	if err := s.runtimeStore.AppendEvent(ctx, &decisionEvent); err != nil {
		// Log but don't fail the response
		_ = err
	}

	// Persist updated limits state when we allow execution
	if finalDecision == domain.DecisionAllow || finalDecision == domain.DecisionWarn || finalDecision == domain.DecisionThrottle {
		now := time.Now()
		windowStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		windowStartTS := windowStart.Unix() * 1000
		state, _ := s.runtimeStore.GetLimitsState(ctx, proposal.AgentID, windowStartTS, "daily")
		if state == nil {
			state = &domain.LimitsState{
				AgentID:            proposal.AgentID,
				WindowStart:        windowStart,
				WindowType:         "daily",
				BudgetSpentGBP:     0,
				BudgetSpentTokens:  0,
				ActionCount:        0,
			}
		}
		state.BudgetSpentGBP += proposal.Cost.EstimatedGBP
		state.BudgetSpentTokens += proposal.Cost.EstimatedTokens
		state.ActionCount++
		_ = s.runtimeStore.UpdateLimitsState(ctx, *state)
	}

	reasonCodes := reasonCodesFromOutcome(finalDecision, responseReason)
	response := &domain.DecisionResponse{
		Decision:      finalDecision,
		Warnings:      warnings,
		Throttle:      throttle,
		Reason:        responseReason,
		LedgerEventID: eventID,
	}
	for _, code := range reasonCodes {
		response.Reasons = append(response.Reasons, domain.Reason{Code: code, Message: responseReason})
	}
	if finalDecision == domain.DecisionDeny || finalDecision == domain.DecisionStop || finalDecision == domain.DecisionThrottle {
		if rec := recommendations.Recommend(ctx, recommendations.RecommendOptions{
			Decision:         finalDecision,
			ReasonText:       responseReason,
			ReasonCodes:      reasonCodes,
			ActionType:       proposal.Action.Type,
			PanicEnabled:     panicState != nil && panicState.Enabled,
			ResolutionAbsent: strings.Contains(responseReason, "resolution") || strings.Contains(responseReason, "Resolution"),
			AgentID:          proposal.AgentID,
			SessionID:        proposal.SessionID,
		}); rec != nil {
			response.Recommendation = rec
		}
	}

	if finalDecision == domain.DecisionAllow || finalDecision == domain.DecisionWarn || finalDecision == domain.DecisionThrottle {
		token, err := s.resolutionMgr.GenerateToken(ctx, proposal.AgentID, proposal.Action.Type, 10*time.Minute)
		if err == nil {
			response.ExecutionToken = token
		}
	}

	// Emit decision record to ledger sink (noop, bundle, or kernel_http)
	budgetLimit := 10.0
	if effectiveConfig != nil {
		budgetLimit = effectiveConfig.Agents.Default.DailyBudgetGBP
	}
	if budgetLimit <= 0 {
		budgetLimit = 10.0
	}
	record := buildDecisionRecord(proposal, response, &decisionEvent, budgetLimit)
	_ = s.ledgerSink.EmitDecision(ctx, record)
	_ = s.ledgerSink.EmitEvent(ctx, &decisionEvent)

	// Auto-bundle on DENY/STOP (debounced per session+trigger)
	if s.autobundleMgr != nil && (finalDecision == domain.DecisionDeny || finalDecision == domain.DecisionStop) {
		effectivePanic := panicState != nil && panicState.Enabled
		trigger := autobundle.TriggerDecisionDeny
		if finalDecision == domain.DecisionStop {
			trigger = autobundle.TriggerDecisionStop
			if strings.Contains(responseReason, "Loop") {
				trigger = autobundle.TriggerLoopStop
			} else if strings.Contains(strings.ToLower(responseReason), "budget") {
				trigger = autobundle.TriggerBudgetStop
			}
		}
		nextSteps := []string{}
		if response.Recommendation != nil && len(response.Recommendation.NextSteps) > 0 {
			nextSteps = response.Recommendation.NextSteps
		}
		codes := make([]string, 0, len(response.Reasons))
		for _, r := range response.Reasons {
			codes = append(codes, r.Code)
		}
		if path, err := s.autobundleMgr.MaybeBundleOnDecision(ctx, record, trigger, effectivePanic, nextSteps, codes); err == nil && path != "" {
			response.AutobundlePath = path
			response.AutobundleTrigger = trigger
		}
	}

	return response, nil
}

// reasonCodesFromOutcome returns stable reason codes for the given decision and reason text.
func reasonCodesFromOutcome(decision domain.Decision, reason string) []string {
	var codes []string
	if decision == domain.DecisionStop {
		if strings.Contains(reason, "Loop") {
			codes = append(codes, recommendations.CodeLoopStopThreshold)
		} else if strings.Contains(strings.ToLower(reason), "budget") {
			codes = append(codes, recommendations.CodeBudgetStopThreshold)
		}
	}
	if decision == domain.DecisionDeny {
		if strings.Contains(reason, "resolution") || strings.Contains(reason, "Requires resolution") {
			codes = append(codes, recommendations.CodeResolutionRequired)
		} else if strings.Contains(strings.ToLower(reason), "filesystem") {
			codes = append(codes, recommendations.CodeFilesystemDenied)
		} else if strings.Contains(strings.ToLower(reason), "network") {
			codes = append(codes, recommendations.CodeNetworkDomainDenied)
		}
	}
	if len(codes) == 0 && (decision == domain.DecisionDeny || decision == domain.DecisionStop) {
		codes = append(codes, "DENY_OR_STOP")
	}
	return codes
}

// StartSession starts a new session for an agent
func (s *service) StartSession(ctx context.Context, agentID string, metadata map[string]interface{}) (*domain.Session, error) {
	sessionID := "sess:" + uuid.New().String()
	session := domain.Session{
		SessionID: sessionID,
		AgentID:   agentID,
		StartedAt: time.Now(),
		Metadata:  metadata,
	}

	if err := s.runtimeStore.CreateSession(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return &session, nil
}

// EndSession ends a session
func (s *service) EndSession(ctx context.Context, sessionID string) error {
	return s.runtimeStore.EndSession(ctx, sessionID)
}

// GetEvents retrieves events
func (s *service) GetEvents(ctx context.Context, agentID *string, sinceTS *int64, limit int) ([]domain.Event, error) {
	return s.runtimeStore.ListEvents(ctx, runtime.EventFilter{AgentID: agentID, SinceTS: sinceTS, Limit: limit})
}

// ListAgents lists all agents
func (s *service) ListAgents(ctx context.Context) ([]domain.Agent, error) {
	return s.runtimeStore.ListAgents(ctx)
}

// GetAgent retrieves an agent
func (s *service) GetAgent(ctx context.Context, agentID string) (*domain.Agent, error) {
	return s.runtimeStore.GetAgent(ctx, agentID)
}

// HaltAgent halts an agent
func (s *service) HaltAgent(ctx context.Context, agentID string, reason string) error {
	return s.runtimeStore.HaltAgent(ctx, agentID, reason)
}

// ResumeAgent resumes a halted agent
func (s *service) ResumeAgent(ctx context.Context, agentID string) error {
	return s.runtimeStore.ResumeAgent(ctx, agentID)
}

// GetPanicState returns the current panic mode state from the runtime store.
// If panic is enabled but TTL has expired, it is auto-disabled and persisted before returning.
func (s *service) GetPanicState(ctx context.Context) (*domain.PanicState, error) {
	state, err := s.runtimeStore.GetPanicState(ctx)
	if err != nil || state == nil {
		return state, err
	}
	if config.PanicExpired(state) {
		disabled := *state
		disabled.Enabled = false
		_ = s.runtimeStore.SetPanicState(ctx, disabled)
		return &disabled, nil
	}
	return state, nil
}

// SetPanicState updates panic mode state in the runtime store.
func (s *service) SetPanicState(ctx context.Context, state domain.PanicState) error {
	return s.runtimeStore.SetPanicState(ctx, state)
}

// GetAutobundleStatus returns current autobundle config from service config.
func (s *service) GetAutobundleStatus(ctx context.Context) (*config.AutobundleConfig, error) {
	if s.config == nil {
		return &config.AutobundleConfig{}, nil
	}
	return &s.config.Autobundle, nil
}

// GetCapabilities returns capabilities for agent discovery (GET /v1/capabilities).
func (s *service) GetCapabilities(ctx context.Context) (*domain.CapabilitiesResponse, error) {
	out := &domain.CapabilitiesResponse{
		CtrlDot: domain.CtrlDotCapabilities{
			Version: "0.1.0",
			Build:   domain.BuildInfo{GitSHA: "dev", BuiltAt: ""},
			API:     domain.APIInfo{BaseURL: "http://127.0.0.1:7777", Version: "v1"},
			RuntimeStore: domain.RuntimeStoreInfo{Kind: "sqlite"},
			LedgerSink:   domain.LedgerSinkInfo{Kind: "none"},
			Features: domain.FeaturesInfo{
				ResolutionTokens: true,
				LoopDetector:     true,
				BudgetLimits:     true,
				RulesEngine:      true,
				AutoBundles:      true,
				BundleVerify:     true,
			},
		},
	}
	if s.config == nil {
		return out, nil
	}
	cfg := s.config

	// Base URL from config
	host := cfg.Server.Host
	if host == "" {
		host = "127.0.0.1"
	}
	port := cfg.Server.Port
	if port <= 0 {
		port = 7777
	}
	out.CtrlDot.API.BaseURL = fmt.Sprintf("http://%s:%d", host, port)
	out.CtrlDot.RuntimeStore.Kind = cfg.RuntimeStore.Kind
	if out.CtrlDot.RuntimeStore.Kind == "" {
		out.CtrlDot.RuntimeStore.Kind = "sqlite"
	}
	if cfg.RuntimeStore.SQLitePath != "" {
		out.CtrlDot.RuntimeStore.SQLitePath = expandPathForCapabilities(cfg.RuntimeStore.SQLitePath)
	}
	out.CtrlDot.LedgerSink.Kind = cfg.LedgerSink.Kind
	if out.CtrlDot.LedgerSink.Kind == "" {
		out.CtrlDot.LedgerSink.Kind = "none"
	}
	if cfg.LedgerSink.Bundle.OutputDir != "" {
		out.CtrlDot.LedgerSink.BundleDir = expandPathForCapabilities(cfg.LedgerSink.Bundle.OutputDir)
	}

	// Panic state from store
	panicState, _ := s.runtimeStore.GetPanicState(ctx)
	if panicState != nil && panicState.Enabled {
		out.CtrlDot.Panic.Enabled = true
		out.CtrlDot.Panic.ExpiresAt = panicState.ExpiresAt
		out.CtrlDot.Panic.Effective = &domain.PanicEffectiveInfo{
			MaxDailyBudgetUSD: cfg.Panic.MaxDailyBudgetUSD,
			NetworkDefaultDeny: cfg.Panic.Network.DefaultDeny,
			FilesystemMode:     cfg.Panic.Filesystem.Mode,
			Loop: domain.LoopInfo{
				WindowSeconds: cfg.Panic.Loop.WindowSeconds,
				StopRepeats:   cfg.Panic.Loop.StopRepeats,
			},
		}
		if out.CtrlDot.Panic.Effective.Loop.WindowSeconds <= 0 {
			out.CtrlDot.Panic.Effective.Loop.WindowSeconds = 60
		}
		if out.CtrlDot.Panic.Effective.Loop.StopRepeats <= 0 {
			out.CtrlDot.Panic.Effective.Loop.StopRepeats = 5
		}
	}
	out.CtrlDot.Features.AutoBundles = cfg.Autobundle.Enabled
	return out, nil
}

func expandPathForCapabilities(p string) string {
	if strings.HasPrefix(p, "~/") {
		home, _ := os.UserHomeDir()
		if home != "" {
			return filepath.Join(home, p[2:])
		}
	}
	return p
}

// GetAgentLimits returns current budget/limits state for an agent (daily window).
func (s *service) GetAgentLimits(ctx context.Context, agentID string) (*domain.AgentLimitsResponse, error) {
	now := time.Now()
	windowStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	windowStartTS := windowStart.Unix() * 1000

	state, err := s.runtimeStore.GetLimitsState(ctx, agentID, windowStartTS, "daily")
	if err != nil {
		return nil, err
	}
	spent := 0.0
	actionCount := 0
	if state != nil {
		spent = state.BudgetSpentGBP
		actionCount = state.ActionCount
	}

	defaults := config.AgentDefaults{
		DailyBudgetGBP:      10.0,
		WarnPct:             []float64{0.70, 0.90},
		ThrottlePct:         0.95,
		HardStopPct:         1.00,
		MaxIterationsPerAction: 25,
	}
	if s.config != nil {
		defaults = s.config.Agents.Default
	}
	limit := defaults.DailyBudgetGBP
	if limit <= 0 {
		limit = 10.0
	}
	pct := 0.0
	if limit > 0 {
		pct = spent / limit
	}

	return &domain.AgentLimitsResponse{
		AgentID:     agentID,
		WindowStart: windowStart,
		WindowType:  "daily",
		SpentGBP:    spent,
		LimitGBP:    limit,
		Percentage:  pct,
		WarnPct:     defaults.WarnPct,
		ThrottlePct: defaults.ThrottlePct,
		HardStopPct: defaults.HardStopPct,
		ActionCount: actionCount,
	}, nil
}

// GetLimitsConfig returns default limits from config (read-only).
func (s *service) GetLimitsConfig(ctx context.Context) (*domain.LimitsConfigResponse, error) {
	defaults := config.AgentDefaults{
		DailyBudgetGBP:      10.0,
		WarnPct:             []float64{0.70, 0.90},
		ThrottlePct:         0.95,
		HardStopPct:         1.00,
		MaxIterationsPerAction: 25,
	}
	if s.config != nil {
		defaults = s.config.Agents.Default
	}
	return &domain.LimitsConfigResponse{
		DailyBudgetGBP: defaults.DailyBudgetGBP,
		WarnPct:        defaults.WarnPct,
		ThrottlePct:    defaults.ThrottlePct,
		HardStopPct:    defaults.HardStopPct,
	}, nil
}

func buildDecisionRecord(proposal domain.ActionProposal, response *domain.DecisionResponse, ev *domain.Event, budgetLimit float64) *sink.DecisionRecord {
	spent := 0.0
	if ev.CostGBP != nil {
		spent = *ev.CostGBP
	}
	return &sink.DecisionRecord{
		ID:                     ev.EventID,
		AgentID:                 proposal.AgentID,
		SessionID:              proposal.SessionID,
		Timestamp:              ev.TS,
		ActionType:             proposal.Action.Type,
		ActionTarget:           sink.RedactMap(proposal.Action.Target),
		ActionInputs:           sink.RedactMap(proposal.Action.Inputs),
		Decision:               response.Decision,
		Reason:                 response.Reason,
		Warnings:               response.Warnings,
		Throttle:               response.Throttle,
		BudgetSpent:            spent,
		BudgetLimit:            budgetLimit,
		ActionHash:             proposal.Context.Hash,
		ExecutionTokenPresent:  response.ExecutionToken != "",
	}
}

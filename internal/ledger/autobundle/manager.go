package autobundle

import (
	"context"
	"path/filepath"
	"sync"
	"time"

	"github.com/futurematic/kernel/internal/config"
	"github.com/futurematic/kernel/internal/domain"
	"github.com/futurematic/kernel/internal/ledger/sink"
	"github.com/futurematic/kernel/internal/ledger/sink/bundle"
	"github.com/futurematic/kernel/internal/runtime"
)

// Trigger constants for manifest.Trigger.
const (
	TriggerDecisionDeny   = "decision_deny"
	TriggerDecisionStop   = "decision_stop"
	TriggerLoopStop      = "loop_stop"
	TriggerBudgetStop    = "budget_stop"
	TriggerShutdown      = "shutdown"
	TriggerPanicOn       = "panic_on"
	TriggerPanicOff      = "panic_off"
	TriggerManualTest    = "manual_test"
)

// Manager creates signed bundles on DENY/STOP/shutdown/panic toggle, with debounce.
type Manager struct {
	cfg     *config.Config
	store   runtime.RuntimeStore
	mu      sync.Mutex
	lastAt  map[string]time.Time // key: sessionID or sessionID+"."+trigger for debounce
	daemonVersion string
}

// NewManager creates an autobundle manager. cfg must not be nil; store may be nil (no event tail).
func NewManager(cfg *config.Config, store runtime.RuntimeStore, daemonVersion string) *Manager {
	if daemonVersion == "" {
		daemonVersion = "0.1.0"
	}
	return &Manager{
		cfg:           cfg,
		store:         store,
		lastAt:        make(map[string]time.Time),
		daemonVersion: daemonVersion,
	}
}

// MaybeBundleOnDecision writes a bundle when trigger is enabled and not debounced.
// trigger should be one of TriggerDecisionDeny, TriggerDecisionStop, TriggerLoopStop, TriggerBudgetStop.
// nextSteps and reasonCodes are optional (for README.md).
func (m *Manager) MaybeBundleOnDecision(ctx context.Context, record *sink.DecisionRecord, trigger string, effectivePanic bool, nextSteps []string, reasonCodes []string) (path string, err error) {
	if m.cfg == nil || !m.cfg.Autobundle.Enabled {
		return "", nil
	}
	if !m.triggerEnabled(trigger) {
		return "", nil
	}
	sessionID := record.SessionID
	if sessionID == "" {
		sessionID = "_no_session"
	}
	debounceKey := sessionID + "." + trigger
	if m.debounced(debounceKey) {
		return "", nil
	}

	outputDir := m.outputDir()
	if outputDir == "" {
		return "", nil
	}

	decisions := []*sink.DecisionRecord{record}
	var events []*domain.Event
	if m.store != nil && m.cfg.Autobundle.Include.EventsTail > 0 {
		since := time.Now().Add(-1 * time.Hour).Unix() * 1000
		agentID := record.AgentID
		list, _ := m.store.ListEvents(ctx, runtime.EventFilter{AgentID: &agentID, SinceTS: &since, Limit: m.cfg.Autobundle.Include.EventsTail})
		for i := range list {
			events = append(events, &list[i])
		}
	}
	var cfgSnapshot *config.Config
	if m.cfg.Autobundle.Include.ConfigSnapshot {
		cfgSnapshot = m.cfg
	}

	path, err = bundle.WriteOne(bundle.WriteOneOptions{
		OutputDir:             outputDir,
		SignEnabled:           m.cfg.LedgerSink.Bundle.Sign.Enabled,
		PrivateKeyPath:       m.cfg.LedgerSink.Bundle.Sign.KeyPath,
		PublicKeyPath:        m.cfg.LedgerSink.Bundle.Sign.PublicKeyPath,
		RuntimeStoreKind:     m.runtimeKind(),
		DaemonVersion:        m.daemonVersion,
		SessionID:            record.SessionID,
		AgentID:              record.AgentID,
		Decisions:            decisions,
		Events:               events,
		ConfigSnapshot:       cfgSnapshot,
		Trigger:              trigger,
		DecisionID:           record.ID,
		EffectivePanicEnabled: effectivePanic,
		ReasonCodes:          reasonCodes,
		NextSteps:            nextSteps,
	})
	if err != nil {
		return "", err
	}
	m.mu.Lock()
	m.lastAt[debounceKey] = time.Now()
	m.mu.Unlock()
	return path, nil
}

// MaybeBundleOnShutdown writes a shutdown bundle if triggers.on_shutdown is enabled.
// Debounce does not apply to shutdown.
func (m *Manager) MaybeBundleOnShutdown(ctx context.Context) (path string, err error) {
	if m.cfg == nil || !m.cfg.Autobundle.Enabled || !m.cfg.Autobundle.Triggers.OnShutdown {
		return "", nil
	}
	outputDir := m.outputDir()
	if outputDir == "" {
		return "", nil
	}

	var events []*domain.Event
	if m.store != nil && m.cfg.Autobundle.Include.EventsTail > 0 {
		since := time.Now().Add(-1 * time.Hour).Unix() * 1000
		list, _ := m.store.ListEvents(ctx, runtime.EventFilter{SinceTS: &since, Limit: m.cfg.Autobundle.Include.EventsTail})
		for i := range list {
			events = append(events, &list[i])
		}
	}
	var cfgSnapshot *config.Config
	if m.cfg.Autobundle.Include.ConfigSnapshot {
		cfgSnapshot = m.cfg
	}

	path, err = bundle.WriteOne(bundle.WriteOneOptions{
		OutputDir:             outputDir,
		SignEnabled:           m.cfg.LedgerSink.Bundle.Sign.Enabled,
		PrivateKeyPath:       m.cfg.LedgerSink.Bundle.Sign.KeyPath,
		PublicKeyPath:        m.cfg.LedgerSink.Bundle.Sign.PublicKeyPath,
		RuntimeStoreKind:     m.runtimeKind(),
		DaemonVersion:        m.daemonVersion,
		SessionID:            "",
		AgentID:              "",
		Decisions:            nil,
		Events:               events,
		ConfigSnapshot:       cfgSnapshot,
		Trigger:              TriggerShutdown,
		EffectivePanicEnabled: false,
	})
	if err != nil {
		return "", err
	}
	return path, nil
}

// MaybeBundleOnPanicToggle writes a bundle when panic is toggled if triggers.on_panic_toggle is enabled.
func (m *Manager) MaybeBundleOnPanicToggle(ctx context.Context, panicOn bool) (path string, err error) {
	if m.cfg == nil || !m.cfg.Autobundle.Enabled || !m.cfg.Autobundle.Triggers.OnPanicToggle {
		return "", nil
	}
	outputDir := m.outputDir()
	if outputDir == "" {
		return "", nil
	}

	trigger := TriggerPanicOff
	if panicOn {
		trigger = TriggerPanicOn
	}
	path, err = bundle.WriteOne(bundle.WriteOneOptions{
		OutputDir:             outputDir,
		SignEnabled:           m.cfg.LedgerSink.Bundle.Sign.Enabled,
		PrivateKeyPath:       m.cfg.LedgerSink.Bundle.Sign.KeyPath,
		PublicKeyPath:        m.cfg.LedgerSink.Bundle.Sign.PublicKeyPath,
		RuntimeStoreKind:     m.runtimeKind(),
		DaemonVersion:        m.daemonVersion,
		SessionID:            "",
		AgentID:              "",
		Decisions:            nil,
		Events:               nil,
		ConfigSnapshot:       nil,
		Trigger:              trigger,
		EffectivePanicEnabled: panicOn,
	})
	if err != nil {
		return "", err
	}
	return path, nil
}

// MaybeBundleTest forces a bundle with trigger manual_test (for CLI test). No debounce.
func (m *Manager) MaybeBundleTest(ctx context.Context) (path string, err error) {
	if m.cfg == nil || !m.cfg.Autobundle.Enabled {
		return "", nil
	}
	outputDir := m.outputDir()
	if outputDir == "" {
		return "", nil
	}

	var events []*domain.Event
	if m.store != nil && m.cfg.Autobundle.Include.EventsTail > 0 {
		since := time.Now().Add(-1 * time.Hour).Unix() * 1000
		list, _ := m.store.ListEvents(ctx, runtime.EventFilter{SinceTS: &since, Limit: m.cfg.Autobundle.Include.EventsTail})
		for i := range list {
			events = append(events, &list[i])
		}
	}
	var cfgSnapshot *config.Config
	if m.cfg.Autobundle.Include.ConfigSnapshot {
		cfgSnapshot = m.cfg
	}

	return bundle.WriteOne(bundle.WriteOneOptions{
		OutputDir:      outputDir,
		SignEnabled:    m.cfg.LedgerSink.Bundle.Sign.Enabled,
		PrivateKeyPath: m.cfg.LedgerSink.Bundle.Sign.KeyPath,
		PublicKeyPath:  m.cfg.LedgerSink.Bundle.Sign.PublicKeyPath,
		RuntimeStoreKind: m.runtimeKind(),
		DaemonVersion:  m.daemonVersion,
		SessionID:      "",
		AgentID:        "",
		Decisions:      nil,
		Events:         events,
		ConfigSnapshot: cfgSnapshot,
		Trigger:        TriggerManualTest,
	})
}

func (m *Manager) triggerEnabled(trigger string) bool {
	switch trigger {
	case TriggerDecisionDeny:
		return m.cfg.Autobundle.Triggers.OnDeny
	case TriggerDecisionStop:
		return m.cfg.Autobundle.Triggers.OnStop
	case TriggerLoopStop:
		return m.cfg.Autobundle.Triggers.OnLoopStop
	case TriggerBudgetStop:
		return m.cfg.Autobundle.Triggers.OnBudgetStop
	default:
		return false
	}
}

func (m *Manager) debounced(key string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	last, ok := m.lastAt[key]
	if !ok {
		return false
	}
	debounceSec := m.cfg.Autobundle.DebounceSeconds
	if debounceSec <= 0 {
		debounceSec = 10
	}
	return time.Since(last) < time.Duration(debounceSec)*time.Second
}

func (m *Manager) outputDir() string {
	dir := m.cfg.Autobundle.OutputDir
	if dir == "" {
		dir = m.cfg.LedgerSink.Bundle.OutputDir
	}
	if dir == "" {
		home, _ := filepath.Abs(".")
		return filepath.Join(home, ".ctrldot", "bundles")
	}
	return dir
}

func (m *Manager) runtimeKind() string {
	k := m.cfg.RuntimeStore.Kind
	if k == "" {
		return "sqlite"
	}
	return k
}

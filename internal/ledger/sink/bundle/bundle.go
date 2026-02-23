package bundle

import (
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/futurematic/kernel/internal/config"
	"github.com/futurematic/kernel/internal/domain"
	"github.com/futurematic/kernel/internal/ledger/sink"
	"gopkg.in/yaml.v3"
)

const bundleVersion = "1"

// Sink buffers decisions and events per session and writes a signed bundle on Close().
type Sink struct {
	outputDir       string
	signEnabled     bool
	privateKeyPath  string
	publicKeyPath   string
	runtimeStoreKind string
	daemonVersion   string
	configSnapshot  *config.Config // redacted when writing
	mu              sync.Mutex
	sessions        map[string]*sessionBundle
}

type sessionBundle struct {
	SessionID string
	AgentID   string
	Created   time.Time
	Decisions []*sink.DecisionRecord
	Events    []*domain.Event
}

// BundleManifest is written as manifest.json.
// Optional trigger fields are set by auto-bundles; VerifyBundle ignores unknown fields.
type BundleManifest struct {
	BundleVersion          string            `json:"bundle_version"`
	CreatedAt              time.Time         `json:"created_at"`
	DaemonVersion          string           `json:"daemon_version"`
	RuntimeStoreKind       string           `json:"runtime_store_kind"`
	LedgerSinkKind         string           `json:"ledger_sink_kind"`
	SessionID              string           `json:"session_id"`
	AgentID                string           `json:"agent_id"`
	Hashes                 map[string]string `json:"hashes"`
	Redactions             []string         `json:"redactions,omitempty"`
	// Auto-bundle metadata (optional)
	Trigger                string    `json:"trigger,omitempty"`                  // e.g. decision_deny, decision_stop, shutdown, panic_on
	TriggeredAt            time.Time `json:"triggered_at,omitempty"`             // when trigger fired
	DecisionID              string    `json:"decision_id,omitempty"`               // ID of decision that triggered (if any)
	EffectivePanicEnabled   bool      `json:"effective_panic_enabled,omitempty"`   // panic was on when bundle created
}

// NewSink creates a bundle sink. Output dir and key paths are expanded (~).
func NewSink(cfg *config.Config, runtimeStoreKind, daemonVersion string) (*Sink, error) {
	outputDir := expandPath(cfg.LedgerSink.Bundle.OutputDir)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("bundle output dir: %w", err)
	}
	s := &Sink{
		outputDir:        outputDir,
		signEnabled:      cfg.LedgerSink.Bundle.Sign.Enabled,
		privateKeyPath:   expandPath(cfg.LedgerSink.Bundle.Sign.KeyPath),
		publicKeyPath:    expandPath(cfg.LedgerSink.Bundle.Sign.PublicKeyPath),
		runtimeStoreKind: runtimeStoreKind,
		daemonVersion:    daemonVersion,
		configSnapshot:   cfg,
		sessions:         make(map[string]*sessionBundle),
	}
	if s.signEnabled {
		if err := s.ensureKeypair(); err != nil {
			return nil, err
		}
	}
	return s, nil
}

func expandPath(p string) string {
	if len(p) >= 2 && p[:2] == "~/" {
		home, _ := os.UserHomeDir()
		if home != "" {
			return filepath.Join(home, p[2:])
		}
	}
	return p
}

func (s *Sink) ensureKeypair() error {
	dir := filepath.Dir(s.privateKeyPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("keys dir: %w", err)
	}
	if _, err := os.Stat(s.privateKeyPath); err == nil {
		return nil
	}
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		return fmt.Errorf("generate key: %w", err)
	}
	if err := os.WriteFile(s.privateKeyPath, priv, 0600); err != nil {
		return fmt.Errorf("write private key: %w", err)
	}
	if err := os.WriteFile(s.publicKeyPath, pub, 0644); err != nil {
		return fmt.Errorf("write public key: %w", err)
	}
	return nil
}

func (s *Sink) getOrCreateSession(sessionID, agentID string) *sessionBundle {
	if b, ok := s.sessions[sessionID]; ok {
		return b
	}
	b := &sessionBundle{
		SessionID: sessionID,
		AgentID:   agentID,
		Created:   time.Now(),
		Decisions: nil,
		Events:    nil,
	}
	s.sessions[sessionID] = b
	return b
}

// EmitDecision buffers the decision for the session.
func (s *Sink) EmitDecision(ctx context.Context, d *sink.DecisionRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	sessionID := d.SessionID
	if sessionID == "" {
		sessionID = "_no_session"
	}
	b := s.getOrCreateSession(sessionID, d.AgentID)
	b.Decisions = append(b.Decisions, d)
	return nil
}

// EmitEvent buffers the event for the session.
func (s *Sink) EmitEvent(ctx context.Context, e *domain.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	sessionID := e.SessionID
	if sessionID == "" {
		sessionID = "_no_session"
	}
	b := s.getOrCreateSession(sessionID, e.AgentID)
	b.Events = append(b.Events, e)
	return nil
}

// Close writes each session's bundle to disk (manifest, files, signature) then clears state.
func (s *Sink) Close() error {
	s.mu.Lock()
	sessions := make(map[string]*sessionBundle, len(s.sessions))
	for k, v := range s.sessions {
		sessions[k] = v
	}
	s.sessions = make(map[string]*sessionBundle)
	s.mu.Unlock()

	for _, b := range sessions {
		if err := s.writeBundle(b); err != nil {
			return err
		}
	}
	return nil
}

// SessionCount returns the number of buffered sessions (for tests).
func (s *Sink) SessionCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.sessions)
}

func (s *Sink) writeBundle(b *sessionBundle) error {
	// One bundle per session: bundle_<created>_<sessionID>
	createdStr := b.Created.UTC().Format("2006-01-02T150405Z")
	safeSess := b.SessionID
	if len(safeSess) > 36 {
		safeSess = safeSess[:36]
	}
	dirName := fmt.Sprintf("bundle_%s_%s", createdStr, safeSess)
	dir := filepath.Join(s.outputDir, dirName)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("bundle dir: %w", err)
	}

	// decision_records.jsonl
	decPath := filepath.Join(dir, "decision_records.jsonl")
	decFile, err := os.Create(decPath)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(decFile)
	for _, d := range b.Decisions {
		if err := enc.Encode(d); err != nil {
			decFile.Close()
			return err
		}
	}
	if err := decFile.Close(); err != nil {
		return err
	}

	// events.jsonl
	evPath := filepath.Join(dir, "events.jsonl")
	evFile, err := os.Create(evPath)
	if err != nil {
		return err
	}
	evEnc := json.NewEncoder(evFile)
	for _, e := range b.Events {
		if err := evEnc.Encode(e); err != nil {
			evFile.Close()
			return err
		}
	}
	if err := evFile.Close(); err != nil {
		return err
	}

	// config_snapshot.yaml (redacted)
	cfgPath := filepath.Join(dir, "config_snapshot.yaml")
	cfgRedacted := redactConfig(s.configSnapshot)
	cfgYaml, err := yaml.Marshal(cfgRedacted)
	if err != nil {
		return err
	}
	if err := os.WriteFile(cfgPath, cfgYaml, 0644); err != nil {
		return err
	}

	// Hashes of the three files (manifest will be written after we sign it)
	hashes := make(map[string]string)
	for name, path := range map[string]string{
		"decision_records.jsonl": decPath,
		"events.jsonl":           evPath,
		"config_snapshot.yaml":   cfgPath,
	} {
		h, err := sha256File(path)
		if err != nil {
			return err
		}
		hashes[name] = h
	}

	manifest := BundleManifest{
		BundleVersion:    bundleVersion,
		CreatedAt:        b.Created,
		DaemonVersion:    s.daemonVersion,
		RuntimeStoreKind: s.runtimeStoreKind,
		LedgerSinkKind:   "bundle",
		SessionID:        b.SessionID,
		AgentID:          b.AgentID,
		Hashes:           hashes,
		Redactions:       RedactKeys,
	}
	manifestBytes, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}
	manifestPath := filepath.Join(dir, "manifest.json")
	if err := os.WriteFile(manifestPath, manifestBytes, 0644); err != nil {
		return err
	}

	var sig []byte
	if s.signEnabled {
		priv, err := os.ReadFile(s.privateKeyPath)
		if err != nil {
			return fmt.Errorf("read private key: %w", err)
		}
		if len(priv) != ed25519.PrivateKeySize {
			return fmt.Errorf("invalid private key size")
		}
		sig = ed25519.Sign(priv, manifestBytes)
		if err := os.WriteFile(filepath.Join(dir, "signature.ed25519"), sig, 0644); err != nil {
			return err
		}
		pub, _ := os.ReadFile(s.publicKeyPath)
		if len(pub) == ed25519.PublicKeySize {
			_ = os.WriteFile(filepath.Join(dir, "public_key.ed25519"), pub, 0644)
		}
	}
	return nil
}

func sha256File(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:]), nil
}

// redactConfig returns a copy of config with sensitive values redacted (for config_snapshot.yaml).
func redactConfig(cfg *config.Config) interface{} {
	if cfg == nil {
		return nil
	}
	// Minimal struct for snapshot: only safe top-level fields
	return map[string]interface{}{
		"runtime_store": map[string]interface{}{
			"kind":        cfg.RuntimeStore.Kind,
			"sqlite_path": cfg.RuntimeStore.SQLitePath,
		},
		"ledger_sink": map[string]interface{}{
			"kind": cfg.LedgerSink.Kind,
		},
		"server": map[string]interface{}{
			"host": cfg.Server.Host,
			"port": cfg.Server.Port,
		},
		"events": map[string]interface{}{
			"retention_days": cfg.Events.RetentionDays,
			"max_rows":       cfg.Events.MaxRows,
		},
	}
}

// VerifyBundle verifies signature and manifest hashes for a bundle directory.
func VerifyBundle(dir string) error {
	manifestPath := filepath.Join(dir, "manifest.json")
	manifestData, err := os.ReadFile(manifestPath)
	if err != nil {
		return fmt.Errorf("read manifest: %w", err)
	}
	var m BundleManifest
	if err := json.Unmarshal(manifestData, &m); err != nil {
		return fmt.Errorf("parse manifest: %w", err)
	}
	for name, wantHash := range m.Hashes {
		path := filepath.Join(dir, name)
		got, err := sha256File(path)
		if err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}
		if got != wantHash {
			return fmt.Errorf("%s: hash mismatch (got %s, want %s)", name, got, wantHash)
		}
	}
	sigPath := filepath.Join(dir, "signature.ed25519")
	pubPath := filepath.Join(dir, "public_key.ed25519")
	pub, err := os.ReadFile(pubPath)
	if err != nil {
		return fmt.Errorf("read public key: %w", err)
	}
	if len(pub) != ed25519.PublicKeySize {
		return fmt.Errorf("invalid public key size")
	}
	sig, err := os.ReadFile(sigPath)
	if err != nil {
		return fmt.Errorf("read signature: %w", err)
	}
	if !ed25519.Verify(pub, manifestData, sig) {
		return fmt.Errorf("signature verification failed")
	}
	return nil
}

// ListBundles returns bundle directory names in outputDir, sorted by name (newest last).
func ListBundles(outputDir string) ([]string, error) {
	outputDir = expandPath(outputDir)
	entries, err := os.ReadDir(outputDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() && len(e.Name()) > 7 && e.Name()[:7] == "bundle_" {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	return names, nil
}

// WriteOneOptions supplies inputs for writing a single bundle (e.g. auto-bundle).
type WriteOneOptions struct {
	OutputDir             string
	SignEnabled           bool
	PrivateKeyPath        string
	PublicKeyPath         string
	RuntimeStoreKind      string
	DaemonVersion         string
	SessionID             string
	AgentID               string
	Decisions             []*sink.DecisionRecord
	Events                []*domain.Event
	ConfigSnapshot        *config.Config
	Trigger               string   // e.g. decision_deny, decision_stop, shutdown, panic_on
	DecisionID            string
	EffectivePanicEnabled bool
	ReasonCodes           []string // for README.md
	NextSteps             []string // for README.md (runnable commands)
}

// WriteOne writes a single bundle directory with the given decisions, events, and optional trigger metadata.
// Returns the bundle directory path. VerifyBundle works on the result.
func WriteOne(opts WriteOneOptions) (string, error) {
	opts.OutputDir = expandPath(opts.OutputDir)
	if err := os.MkdirAll(opts.OutputDir, 0755); err != nil {
		return "", fmt.Errorf("bundle output dir: %w", err)
	}
	now := time.Now()
	createdStr := now.UTC().Format("2006-01-02T150405Z")
	safeSess := opts.SessionID
	if safeSess == "" {
		safeSess = "_autobundle"
	}
	if len(safeSess) > 36 {
		safeSess = safeSess[:36]
	}
	dirName := fmt.Sprintf("bundle_%s_%s", createdStr, safeSess)
	dir := filepath.Join(opts.OutputDir, dirName)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("bundle dir: %w", err)
	}

	decPath := filepath.Join(dir, "decision_records.jsonl")
	decFile, err := os.Create(decPath)
	if err != nil {
		return "", err
	}
	enc := json.NewEncoder(decFile)
	for _, d := range opts.Decisions {
		if err := enc.Encode(d); err != nil {
			decFile.Close()
			return "", err
		}
	}
	if err := decFile.Close(); err != nil {
		return "", err
	}

	evPath := filepath.Join(dir, "events.jsonl")
	evFile, err := os.Create(evPath)
	if err != nil {
		return "", err
	}
	evEnc := json.NewEncoder(evFile)
	for _, e := range opts.Events {
		_ = evEnc.Encode(e)
	}
	if err := evFile.Close(); err != nil {
		return "", err
	}

	hashes := make(map[string]string)
	for name, path := range map[string]string{"decision_records.jsonl": decPath, "events.jsonl": evPath} {
		h, err := sha256File(path)
		if err != nil {
			return "", err
		}
		hashes[name] = h
	}
	if opts.ConfigSnapshot != nil {
		cfgPath := filepath.Join(dir, "config_snapshot.yaml")
		cfgYaml, err := yaml.Marshal(redactConfig(opts.ConfigSnapshot))
		if err != nil {
			return "", err
		}
		if err := os.WriteFile(cfgPath, cfgYaml, 0644); err != nil {
			return "", err
		}
		h, err := sha256File(cfgPath)
		if err != nil {
			return "", err
		}
		hashes["config_snapshot.yaml"] = h
	}

	manifest := BundleManifest{
		BundleVersion:          bundleVersion,
		CreatedAt:              now,
		DaemonVersion:          opts.DaemonVersion,
		RuntimeStoreKind:       opts.RuntimeStoreKind,
		LedgerSinkKind:         "bundle",
		SessionID:              opts.SessionID,
		AgentID:                opts.AgentID,
		Hashes:                 hashes,
		Redactions:             RedactKeys,
		Trigger:                opts.Trigger,
		TriggeredAt:            now,
		DecisionID:             opts.DecisionID,
		EffectivePanicEnabled:  opts.EffectivePanicEnabled,
	}
	manifestBytes, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return "", err
	}
	manifestPath := filepath.Join(dir, "manifest.json")
	if err := os.WriteFile(manifestPath, manifestBytes, 0644); err != nil {
		return "", err
	}
	if opts.SignEnabled {
		priv, err := os.ReadFile(expandPath(opts.PrivateKeyPath))
		if err != nil {
			return "", fmt.Errorf("read private key: %w", err)
		}
		if len(priv) != ed25519.PrivateKeySize {
			return "", fmt.Errorf("invalid private key size")
		}
		sig := ed25519.Sign(priv, manifestBytes)
		if err := os.WriteFile(filepath.Join(dir, "signature.ed25519"), sig, 0644); err != nil {
			return "", err
		}
		pub, _ := os.ReadFile(expandPath(opts.PublicKeyPath))
		if len(pub) == ed25519.PublicKeySize {
			_ = os.WriteFile(filepath.Join(dir, "public_key.ed25519"), pub, 0644)
		}
	}

	// README.md (not in manifest hashes; for humans and agents)
	outcome := "DENY/STOP"
	if len(opts.Decisions) > 0 {
		outcome = string(opts.Decisions[0].Decision)
	}
	reasonSummary := ""
	if len(opts.Decisions) > 0 {
		reasonSummary = opts.Decisions[0].Reason
	}
	_ = WriteREADME(dir, ReadMEOptions{
		Trigger:       opts.Trigger,
		Timestamp:     now,
		AgentID:       opts.AgentID,
		SessionID:     opts.SessionID,
		Outcome:       outcome,
		ReasonCodes:   opts.ReasonCodes,
		ReasonSummary: reasonSummary,
		PanicEnabled:  opts.EffectivePanicEnabled,
		NextSteps:     opts.NextSteps,
		BundleDirName: dirName,
	})

	return dir, nil
}

var _ sink.LedgerSink = (*Sink)(nil)

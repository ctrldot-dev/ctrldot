package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"gopkg.in/yaml.v3"
)

// Config represents the Ctrl Dot configuration
type Config struct {
	Server          ServerConfig        `yaml:"server"`
	Ledger          LedgerConfig        `yaml:"ledger"`
	RuntimeStore    RuntimeStoreConfig  `yaml:"runtime_store"`
	LedgerSink      LedgerSinkConfig    `yaml:"ledger_sink"`
	Events          EventsConfig        `yaml:"events"`
	Agents          AgentsConfig        `yaml:"agents"`
	Rules           RulesConfig         `yaml:"rules"`
	DegradeModes    DegradeModesConfig  `yaml:"degrade_modes"`
	Panic           PanicConfig         `yaml:"panic"`
	Autobundle      AutobundleConfig    `yaml:"autobundle"`
	DisplayCurrency string              `yaml:"display_currency"` // "gbp", "usd", "eur" — for UI display only; values stored in GBP
	// Loop is set by Effective() when panic is on; loop detector uses it for window/repeats.
	Loop *LoopOverlay `yaml:"-"`
}

// AutobundleConfig configures automatic bundle creation on DENY/STOP/shutdown.
type AutobundleConfig struct {
	Enabled          bool                  `yaml:"enabled"`
	OutputDir        string                `yaml:"output_dir"`
	DebounceSeconds  int                   `yaml:"debounce_seconds"`
	Triggers         AutobundleTriggers    `yaml:"triggers"`
	Include          AutobundleInclude     `yaml:"include"`
}

// AutobundleTriggers specifies when to create an auto-bundle.
type AutobundleTriggers struct {
	OnDeny        bool `yaml:"on_deny"`
	OnStop        bool `yaml:"on_stop"`
	OnBudgetStop  bool `yaml:"on_budget_stop"`
	OnLoopStop    bool `yaml:"on_loop_stop"`
	OnShutdown    bool `yaml:"on_shutdown"`
	OnPanicToggle bool `yaml:"on_panic_toggle"`
}

// AutobundleInclude specifies what to include in auto-bundles.
type AutobundleInclude struct {
	EventsTail     int  `yaml:"events_tail"`
	DecisionsTail  int  `yaml:"decisions_tail"`
	ConfigSnapshot bool `yaml:"config_snapshot"`
}

// LoopOverlay overrides loop detection window and repeat count (e.g. when panic is on).
type LoopOverlay struct {
	WindowSeconds int
	StopRepeats   int
}

// PanicConfig configures panic mode (strict overlay when enabled).
type PanicConfig struct {
	Enabled           bool                `yaml:"enabled"`
	TTLSeconds        int                 `yaml:"ttl_seconds"`
	MaxDailyBudgetUSD float64             `yaml:"max_daily_budget_usd"`
	Thresholds        PanicThresholds     `yaml:"thresholds"`
	Resolution        PanicResolution     `yaml:"resolution"`
	Filesystem        PanicFilesystem     `yaml:"filesystem"`
	Network           PanicNetwork        `yaml:"network"`
	Loop              PanicLoop           `yaml:"loop"`
	Exec              PanicExec           `yaml:"exec"`
}

// PanicThresholds overrides warn/throttle/stop percentages when panic is on.
type PanicThresholds struct {
	WarnPct     float64 `yaml:"warn_pct"`
	ThrottlePct float64 `yaml:"throttle_pct"`
	StopPct     float64 `yaml:"stop_pct"`
}

// PanicResolution forces require_resolution when panic is on.
type PanicResolution struct {
	ForceRequireResolution bool `yaml:"force_require_resolution"`
}

// PanicFilesystem restricts filesystem to workspace or read-only when panic is on.
type PanicFilesystem struct {
	Mode           string   `yaml:"mode"` // "workspace_only" | "read_only"
	WorkspaceRoots []string `yaml:"workspace_roots"`
}

// PanicNetwork default-deny with allowlist when panic is on.
type PanicNetwork struct {
	DefaultDeny  bool     `yaml:"default_deny"`
	AllowDomains []string `yaml:"allow_domains"`
}

// PanicLoop tighter repeat thresholds when panic is on.
type PanicLoop struct {
	ThrottleRepeats int `yaml:"throttle_repeats"`
	StopRepeats     int `yaml:"stop_repeats"`
	WindowSeconds   int `yaml:"window_seconds"`
}

// PanicExec requires resolution for exec when panic is on.
type PanicExec struct {
	RequireResolution bool     `yaml:"require_resolution"`
	AllowCommands     []string `yaml:"allow_commands"`
}

// RuntimeStoreConfig configures where Ctrl Dot runtime state is stored (sqlite or postgres).
type RuntimeStoreConfig struct {
	Kind       string `yaml:"kind"`        // "sqlite" | "postgres"
	SQLitePath string `yaml:"sqlite_path"` // path to SQLite DB file
	DBURL     string `yaml:"db_url"`       // for postgres
}

// LedgerSinkConfig configures where decision records are emitted.
type LedgerSinkConfig struct {
	Kind       string                 `yaml:"kind"`   // "none" | "kernel_http" | "bundle"
	KernelHTTP LedgerKernelHTTPConfig `yaml:"kernel_http"` // used when kind == "kernel_http"
	Bundle     LedgerBundleConfig     `yaml:"bundle"`     // used when kind == "bundle"
}

// LedgerKernelHTTPConfig configures the Kernel HTTP sink.
type LedgerKernelHTTPConfig struct {
	BaseURL   string `yaml:"base_url"`   // e.g. http://127.0.0.1:8080
	APIKey    string `yaml:"api_key"`    // optional
	TimeoutMs int    `yaml:"timeout_ms"` // e.g. 2000
	Required  bool   `yaml:"required"`   // if true, failures affect decision path; default false (best-effort)
}

// LedgerBundleConfig configures the signed bundle sink.
type LedgerBundleConfig struct {
	OutputDir string              `yaml:"output_dir"` // e.g. ~/.ctrldot/bundles
	Sign      LedgerBundleSign    `yaml:"sign"`
}

// LedgerBundleSign configures Ed25519 signing for bundles.
type LedgerBundleSign struct {
	Enabled       bool   `yaml:"enabled"`
	KeyPath       string `yaml:"key_path"`        // private key, e.g. ~/.ctrldot/keys/ctrldot_ed25519
	PublicKeyPath string `yaml:"public_key_path"` // optional, e.g. ~/.ctrldot/keys/ctrldot_ed25519.pub
}

// EventsConfig configures event retention for the runtime event log.
type EventsConfig struct {
	RetentionDays int `yaml:"retention_days"`
	MaxRows       int `yaml:"max_rows"`
}

// ServerConfig contains server settings
type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

// LedgerConfig contains ledger/database settings
type LedgerConfig struct {
	DBURL string `yaml:"db_url"`
}

// AgentsConfig contains agent default settings
type AgentsConfig struct {
	Default AgentDefaults `yaml:"default"`
}

// AgentDefaults contains default agent budget and limits
type AgentDefaults struct {
	DailyBudgetGBP      float64   `yaml:"daily_budget_gbp"`
	WarnPct             []float64 `yaml:"warn_pct"`
	ThrottlePct         float64   `yaml:"throttle_pct"`
	HardStopPct         float64   `yaml:"hard_stop_pct"`
	MaxIterationsPerAction int    `yaml:"max_iterations_per_action"`
}

// RulesConfig contains domain rules
type RulesConfig struct {
	RequireResolution []string        `yaml:"require_resolution"`
	Filesystem        FilesystemRules  `yaml:"filesystem"`
	Network           NetworkRules     `yaml:"network"`
}

// FilesystemRules contains filesystem access rules
type FilesystemRules struct {
	AllowRoots []string `yaml:"allow_roots"`
}

// NetworkRules contains network access rules
type NetworkRules struct {
	DenyAll      bool     `yaml:"deny_all"`
	AllowDomains []string `yaml:"allow_domains"`
}

// DegradeModesConfig contains throttle/degrade mode settings
type DegradeModesConfig struct {
	Cheap DegradeMode `yaml:"cheap"`
}

// DegradeMode defines a degraded operation mode
type DegradeMode struct {
	ModelPolicy    string   `yaml:"model_policy"`
	MaxParallelTasks int    `yaml:"max_parallel_tasks"`
	DenyTools      []string `yaml:"deny_tools"`
}

// EnsureDefaultConfigFile creates the config directory and writes a default config file if it doesn't exist.
func EnsureDefaultConfigFile(configPath string) error {
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}
	if _, err := os.Stat(configPath); err == nil {
		return nil // already exists
	}
	data, err := yaml.Marshal(DefaultConfig())
	if err != nil {
		return fmt.Errorf("marshal default config: %w", err)
	}
	header := "# Ctrl Dot config — edit and restart the daemon to apply changes.\n"
	return os.WriteFile(configPath, append([]byte(header), data...), 0600)
}

// Load loads configuration from file or environment
func Load(configPath string) (*Config, error) {
	cfg := DefaultConfig()

	// Expand ~ to home directory
	if len(configPath) >= 2 && configPath[:2] == "~/" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		configPath = filepath.Join(home, configPath[2:])
	}

	// If config file doesn't exist, create it with defaults so the user has a file to edit
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		if err := EnsureDefaultConfigFile(configPath); err != nil {
			return nil, fmt.Errorf("failed to create default config: %w", err)
		}
	}

	// Try to load from file if it exists
	if _, err := os.Stat(configPath); err == nil {
		data, err := os.ReadFile(configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}

		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("failed to parse config file: %w", err)
		}
	}

	// Override with environment variables
	if v := os.Getenv("DB_URL"); v != "" {
		cfg.Ledger.DBURL = v
		cfg.RuntimeStore.DBURL = v
	}
	if v := os.Getenv("CTRLDOT_RUNTIME_STORE"); v != "" {
		cfg.RuntimeStore.Kind = v
	}
	if v := os.Getenv("CTRLDOT_SQLITE_PATH"); v != "" {
		cfg.RuntimeStore.SQLitePath = v
	}
	if v := os.Getenv("CTRLDOT_LEDGER_SINK"); v != "" {
		cfg.LedgerSink.Kind = v
	}
	if v := os.Getenv("CTRLDOT_KERNEL_URL"); v != "" {
		cfg.LedgerSink.KernelHTTP.BaseURL = v
	}
	if v := os.Getenv("CTRLDOT_BUNDLE_DIR"); v != "" {
		cfg.LedgerSink.Bundle.OutputDir = v
	}
	if v := os.Getenv("CTRLDOT_PANIC"); v != "" {
		cfg.Panic.Enabled = v == "1" || v == "true" || v == "on"
	}
	if v := os.Getenv("CTRLDOT_PANIC_TTL"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			cfg.Panic.TTLSeconds = n
		}
	}
	if v := os.Getenv("CTRLDOT_PANIC_BUDGET_USD"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil && f >= 0 {
			cfg.Panic.MaxDailyBudgetUSD = f
		}
	}
	if v := os.Getenv("CTRLDOT_AUTOBUNDLE"); v != "" {
		cfg.Autobundle.Enabled = v == "1" || v == "true" || v == "on"
	}
	if v := os.Getenv("CTRLDOT_AUTOBUNDLE_DIR"); v != "" {
		cfg.Autobundle.OutputDir = v
	}

	return cfg, nil
}

// Write writes configuration to a file. Path is expanded (e.g. ~ to home).
func Write(configPath string, cfg *Config) error {
	if len(configPath) >= 2 && configPath[:2] == "~/" {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("home dir: %w", err)
		}
		configPath = filepath.Join(home, configPath[2:])
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	header := "# Ctrl Dot config — edit and restart the daemon to apply changes.\n"
	return os.WriteFile(configPath, append([]byte(header), data...), 0600)
}

// DefaultConfig returns default configuration
func DefaultConfig() *Config {
	home, _ := os.UserHomeDir()
	sqlitePath := "~/.ctrldot/ctrldot.sqlite"
	if home != "" {
		sqlitePath = filepath.Join(home, ".ctrldot", "ctrldot.sqlite")
	}
	return &Config{
		Server: ServerConfig{
			Host: "127.0.0.1",
			Port: 7777,
		},
		Ledger: LedgerConfig{
			DBURL: "postgres://kernel:kernel@localhost:5432/kernel?sslmode=disable",
		},
		RuntimeStore: RuntimeStoreConfig{
			Kind:       "sqlite",
			SQLitePath: sqlitePath,
			DBURL:      "postgres://kernel:kernel@localhost:5432/kernel?sslmode=disable",
		},
		LedgerSink: LedgerSinkConfig{
			Kind: "none",
			KernelHTTP: LedgerKernelHTTPConfig{
				BaseURL:   "http://127.0.0.1:8080",
				TimeoutMs: 2000,
				Required:  false,
			},
			Bundle: LedgerBundleConfig{
				OutputDir: filepath.Join(home, ".ctrldot", "bundles"),
				Sign: LedgerBundleSign{
					Enabled:       true,
					KeyPath:       filepath.Join(home, ".ctrldot", "keys", "ctrldot_ed25519"),
					PublicKeyPath: filepath.Join(home, ".ctrldot", "keys", "ctrldot_ed25519.pub"),
				},
			},
		},
		Events: EventsConfig{
			RetentionDays: 7,
			MaxRows:       50000,
		},
		Agents: AgentsConfig{
			Default: AgentDefaults{
				DailyBudgetGBP:      10.0,
				WarnPct:             []float64{0.70, 0.90},
				ThrottlePct:         0.95,
				HardStopPct:          1.00,
				MaxIterationsPerAction: 25,
			},
		},
		Rules: RulesConfig{
			RequireResolution: []string{"git.push", "filesystem.delete"},
			Filesystem: FilesystemRules{
				AllowRoots: []string{"~/dev"},
			},
			Network: NetworkRules{
				DenyAll:      true,
				AllowDomains: []string{"api.openai.com", "api.anthropic.com"},
			},
		},
		DegradeModes: DegradeModesConfig{
			Cheap: DegradeMode{
				ModelPolicy:     "cheap",
				MaxParallelTasks: 2,
				DenyTools:       []string{"web"},
			},
		},
		Panic: PanicConfig{
			Enabled:           false,
			TTLSeconds:        0,
			MaxDailyBudgetUSD: 5.0,
			Thresholds: PanicThresholds{
				WarnPct:     0.40,
				ThrottlePct: 0.60,
				StopPct:     0.90,
			},
			Resolution: PanicResolution{ForceRequireResolution: true},
			Filesystem: PanicFilesystem{Mode: "workspace_only", WorkspaceRoots: []string{}},
			Network: PanicNetwork{
				DefaultDeny: true,
				AllowDomains: []string{"pypi.org", "files.pythonhosted.org", "registry.npmjs.org", "github.com", "raw.githubusercontent.com"},
			},
			Loop:  PanicLoop{ThrottleRepeats: 3, StopRepeats: 5, WindowSeconds: 60},
			Exec:  PanicExec{RequireResolution: true, AllowCommands: []string{}},
		},
		Autobundle: AutobundleConfig{
			Enabled:         true,
			OutputDir:       "", // default: same as LedgerSink.Bundle.OutputDir
			DebounceSeconds: 10,
			Triggers: AutobundleTriggers{
				OnDeny:        true,
				OnStop:        true,
				OnBudgetStop:  true,
				OnLoopStop:    true,
				OnShutdown:    true,
				OnPanicToggle: false,
			},
			Include: AutobundleInclude{
				EventsTail:     500,
				DecisionsTail:  200,
				ConfigSnapshot: true,
			},
		},
	}
}

package domain

import "time"

// CapabilitiesResponse is the response for GET /v1/capabilities (agent discovery).
// No secrets; paths expanded to absolute.
type CapabilitiesResponse struct {
	CtrlDot CtrlDotCapabilities `json:"ctrldot"`
}

// CtrlDotCapabilities describes the Ctrl Dot daemon and its effective configuration.
type CtrlDotCapabilities struct {
	Version       string                 `json:"version"`
	Build         BuildInfo              `json:"build"`
	API           APIInfo                `json:"api"`
	RuntimeStore  RuntimeStoreInfo       `json:"runtime_store"`
	LedgerSink    LedgerSinkInfo         `json:"ledger_sink"`
	Panic         PanicCapabilities      `json:"panic"`
	Features      FeaturesInfo           `json:"features"`
}

// BuildInfo holds build metadata (placeholders if not set at build time).
type BuildInfo struct {
	GitSHA  string `json:"git_sha"`
	BuiltAt string `json:"built_at"`
}

// APIInfo describes the API base URL and version.
type APIInfo struct {
	BaseURL string `json:"base_url"`
	Version string `json:"version"`
}

// RuntimeStoreInfo describes where runtime state is stored.
type RuntimeStoreInfo struct {
	Kind       string `json:"kind"`
	SQLitePath string `json:"sqlite_path,omitempty"`
}

// LedgerSinkInfo describes where decision records are emitted.
type LedgerSinkInfo struct {
	Kind      string `json:"kind"`
	BundleDir string `json:"bundle_dir,omitempty"`
}

// PanicCapabilities describes current panic state and effective overlay when enabled.
type PanicCapabilities struct {
	Enabled   bool                 `json:"enabled"`
	ExpiresAt *time.Time           `json:"expires_at,omitempty"`
	Effective *PanicEffectiveInfo  `json:"effective,omitempty"`
}

// PanicEffectiveInfo is present when panic is enabled (effective overlay).
type PanicEffectiveInfo struct {
	MaxDailyBudgetUSD  float64       `json:"max_daily_budget_usd"`
	NetworkDefaultDeny  bool         `json:"network_default_deny"`
	FilesystemMode      string       `json:"filesystem_mode"`
	Loop                LoopInfo     `json:"loop"`
}

// LoopInfo describes loop detection parameters.
type LoopInfo struct {
	WindowSeconds int `json:"window_seconds"`
	StopRepeats   int `json:"stop_repeats"`
}

// FeaturesInfo describes which features are enabled.
type FeaturesInfo struct {
	ResolutionTokens bool `json:"resolution_tokens"`
	LoopDetector     bool `json:"loop_detector"`
	BudgetLimits     bool `json:"budget_limits"`
	RulesEngine      bool `json:"rules_engine"`
	AutoBundles      bool `json:"auto_bundles"`
	BundleVerify     bool `json:"bundle_verify"`
}

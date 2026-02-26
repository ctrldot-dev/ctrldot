package bios

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/futurematic/kernel/internal/config"
)

// Message types for async API results (Bubble Tea)
type (
	healthResultMsg struct {
		OK      bool
		Version string
		Err     error
	}
	panicStateMsg struct {
		Enabled    bool
		Reason     string
		TTLSeconds int
		Err        error
	}
	agentsListMsg struct {
		Agents []agentEntry
		Err    error
	}
	agentEntry struct {
		AgentID string `json:"agent_id"`
	}
	agentLimitsMsg struct {
		AgentID    string
		SpentGBP   float64
		LimitGBP   float64
		Percentage float64
		Err        error
	}
	rulesLoadedMsg struct {
		RequireResolution []string
		AllowRoots        []string
		NetworkDenyAll    bool
		AllowDomains      []string
		Err               error
	}
	saveResultMsg struct {
		Target string // "rules" or "limits"
		Err    error
	}
)

// FetchHealth runs GET /v1/health and sends healthResultMsg.
func FetchHealth(serverURL string) tea.Cmd {
	return func() tea.Msg {
		resp, err := http.Get(serverURL + "/v1/health")
		if err != nil {
			return healthResultMsg{Err: err}
		}
		defer resp.Body.Close()
		var out struct {
			OK      bool   `json:"ok"`
			Version string `json:"version"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
			return healthResultMsg{Err: err}
		}
		return healthResultMsg{OK: out.OK, Version: out.Version}
	}
}

// FetchPanicState runs GET /v1/panic and sends panicStateMsg.
func FetchPanicState(serverURL string) tea.Cmd {
	return func() tea.Msg {
		resp, err := http.Get(serverURL + "/v1/panic")
		if err != nil {
			return panicStateMsg{Err: err}
		}
		defer resp.Body.Close()
		var out struct {
			Enabled    bool   `json:"enabled"`
			Reason     string `json:"reason"`
			TTLSeconds int    `json:"ttl_seconds"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
			return panicStateMsg{Err: err}
		}
		return panicStateMsg{Enabled: out.Enabled, Reason: out.Reason, TTLSeconds: out.TTLSeconds}
	}
}

// FetchAgents runs GET /v1/agents and sends agentsListMsg.
// The daemon returns a top-level JSON array of agent objects.
func FetchAgents(serverURL string) tea.Cmd {
	return func() tea.Msg {
		resp, err := http.Get(serverURL + "/v1/agents")
		if err != nil {
			return agentsListMsg{Err: err}
		}
		defer resp.Body.Close()
		var agents []agentEntry
		if err := json.NewDecoder(resp.Body).Decode(&agents); err != nil {
			return agentsListMsg{Err: err}
		}
		if agents == nil {
			agents = []agentEntry{}
		}
		return agentsListMsg{Agents: agents}
	}
}

// FetchAgentLimits runs GET /v1/agents/{id}/limits and sends agentLimitsMsg.
// Accepts both spent_gbp and budget_spent_gbp; on non-200 or decode error still sends
// a msg so the UI can show "—" instead of staying on "loading".
func FetchAgentLimits(serverURL, agentID string) tea.Cmd {
	return func() tea.Msg {
		u := serverURL + "/v1/agents/" + url.PathEscape(agentID) + "/limits"
		resp, err := http.Get(u)
		if err != nil {
			return agentLimitsMsg{AgentID: agentID, Err: err}
		}
		defer resp.Body.Close()
		// Accept 200 with flexible JSON; on 404/500 or decode error still report so UI stops showing "loading"
		var data map[string]interface{}
		decErr := json.NewDecoder(resp.Body).Decode(&data)
		if decErr != nil || resp.StatusCode != http.StatusOK {
			// Still send success with zeros so we show "—" not "loading"
			return agentLimitsMsg{AgentID: agentID, SpentGBP: 0, LimitGBP: 0, Percentage: 0}
		}
		spent, _ := numberFromMap(data, "spent_gbp", "budget_spent_gbp")
		limit, _ := numberFromMap(data, "limit_gbp")
		pct, _ := numberFromMap(data, "percentage")
		// percentage might be 0-1 from API
		if pct > 0 && pct <= 1 {
			pct = pct * 100
		}
		return agentLimitsMsg{AgentID: agentID, SpentGBP: spent, LimitGBP: limit, Percentage: pct}
	}
}

func numberFromMap(m map[string]interface{}, keys ...string) (float64, bool) {
	for _, k := range keys {
		if v, ok := m[k]; ok && v != nil {
			switch n := v.(type) {
			case float64:
				return n, true
			case int:
				return float64(n), true
			case int64:
				return float64(n), true
			}
		}
	}
	return 0, false
}

// SetPanicOn runs POST /v1/panic/on and then sends updated panicStateMsg.
func SetPanicOn(serverURL string) tea.Cmd {
	return func() tea.Msg {
		body := map[string]interface{}{"ttl_seconds": 0, "reason": ""}
		bodyBytes, _ := json.Marshal(body)
		req, err := http.NewRequest(http.MethodPost, serverURL+"/v1/panic/on", bytes.NewReader(bodyBytes))
		if err != nil {
			return panicStateMsg{Err: err}
		}
		req.Header.Set("Content-Type", "application/json")
		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			return panicStateMsg{Err: err}
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return panicStateMsg{Err: fmt.Errorf("panic on: %s", resp.Status)}
		}
		return panicStateMsg{Enabled: true}
	}
}

// SetPanicOff runs POST /v1/panic/off and then sends updated panicStateMsg.
func SetPanicOff(serverURL string) tea.Cmd {
	return func() tea.Msg {
		req, _ := http.NewRequest(http.MethodPost, serverURL+"/v1/panic/off", nil)
		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			return panicStateMsg{Err: err}
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return panicStateMsg{Err: fmt.Errorf("panic off: %s", resp.Status)}
		}
		return panicStateMsg{Enabled: false}
	}
}

// ConfigPath returns the Ctrl Dot config file path (CTRLDOT_CONFIG or ~/.ctrldot/config.yaml).
func ConfigPath() string {
	if p := os.Getenv("CTRLDOT_CONFIG"); p != "" {
		return p
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".ctrldot", "config.yaml")
}

// LoadRules runs config.Load in a Cmd and sends rulesLoadedMsg.
func LoadRules(configPath string) tea.Cmd {
	return func() tea.Msg {
		cfg, err := config.Load(configPath)
		if err != nil {
			return rulesLoadedMsg{Err: err}
		}
		return rulesLoadedMsg{
			RequireResolution: cfg.Rules.RequireResolution,
			AllowRoots:        cfg.Rules.Filesystem.AllowRoots,
			NetworkDenyAll:    cfg.Rules.Network.DenyAll,
			AllowDomains:      cfg.Rules.Network.AllowDomains,
		}
	}
}

// limitsConfigLoadedMsg is sent when default limits config has been loaded.
type limitsConfigLoadedMsg struct {
	DailyBudgetGBP  float64
	WarnPct         []float64
	ThrottlePct     float64
	HardStopPct     float64
	MaxIter         int
	DisplayCurrency string // "gbp", "usd", "eur"
	Err             error
}

// LoadLimitsConfig loads agents.default from config and sends limitsConfigLoadedMsg.
func LoadLimitsConfig(configPath string) tea.Cmd {
	return func() tea.Msg {
		cfg, err := config.Load(configPath)
		if err != nil {
			return limitsConfigLoadedMsg{Err: err}
		}
		d := cfg.Agents.Default
		cur := strings.ToLower(cfg.DisplayCurrency)
		if cur != "usd" && cur != "eur" {
			cur = "gbp"
		}
		return limitsConfigLoadedMsg{
			DailyBudgetGBP:  d.DailyBudgetGBP,
			WarnPct:         d.WarnPct,
			ThrottlePct:     d.ThrottlePct,
			HardStopPct:     d.HardStopPct,
			MaxIter:         d.MaxIterationsPerAction,
			DisplayCurrency: cur,
		}
	}
}

// SaveRules updates the rules section in config and writes the file. Sends saveResultMsg.
func SaveRules(configPath string, require []string, allowRoots []string, denyAll bool, domains []string) tea.Cmd {
	return func() tea.Msg {
		cfg, err := config.Load(configPath)
		if err != nil {
			return saveResultMsg{Target: "rules", Err: err}
		}
		cfg.Rules.RequireResolution = require
		cfg.Rules.Filesystem.AllowRoots = allowRoots
		cfg.Rules.Network.DenyAll = denyAll
		cfg.Rules.Network.AllowDomains = domains
		if err := config.Write(configPath, cfg); err != nil {
			return saveResultMsg{Target: "rules", Err: err}
		}
		return saveResultMsg{Target: "rules"}
	}
}

// SaveLimitsConfig updates agents.default and display_currency in config and writes the file. Sends saveResultMsg.
func SaveLimitsConfig(configPath string, dailyGBP float64, warnPct []float64, throttlePct, hardStopPct float64, maxIter int, displayCurrency string) tea.Cmd {
	return func() tea.Msg {
		cfg, err := config.Load(configPath)
		if err != nil {
			return saveResultMsg{Target: "limits", Err: err}
		}
		cfg.Agents.Default.DailyBudgetGBP = dailyGBP
		cfg.Agents.Default.WarnPct = warnPct
		cfg.Agents.Default.ThrottlePct = throttlePct
		cfg.Agents.Default.HardStopPct = hardStopPct
		cfg.Agents.Default.MaxIterationsPerAction = maxIter
		cur := strings.ToLower(displayCurrency)
		if cur != "usd" && cur != "eur" {
			cur = "gbp"
		}
		cfg.DisplayCurrency = cur
		if err := config.Write(configPath, cfg); err != nil {
			return saveResultMsg{Target: "limits", Err: err}
		}
		return saveResultMsg{Target: "limits"}
	}
}

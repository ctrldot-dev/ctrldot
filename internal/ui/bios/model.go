package bios

import (
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
	tea "github.com/charmbracelet/bubbletea"
)

const (
	navWidth = 20
)

// LimitsData holds one agent's budget/limits for display.
type LimitsData struct {
	SpentGBP   float64
	LimitGBP   float64
	Percentage float64
}

// Model is the root Bubble Tea model for the BIOS TUI.
type Model struct {
	width  int
	height int
	nav    list.Model
	ready  bool

	ServerURL string
	configPath string

	// Panel data (filled by API / config)
	healthOK      bool
	healthVersion string
	healthErr     error

	agents       []agentEntry
	agentLimits  map[string]LimitsData // agentID -> limits
	agentsErr    error

	// Default limits (config agents.default) — loaded when Limits panel is shown
	limitsDailyBudget  float64
	limitsWarnPct      []float64
	limitsThrottlePct  float64
	limitsHardStopPct  float64
	limitsMaxIter      int
	limitsErr          error
	limitsEditing      bool
	limitsFieldIndex   int // 0-5: daily, warn, throttle, hardstop, maxiter, display_currency
	limitsInput        textinput.Model
	limitsInputActive  bool
	limitsSaveMsg      string
	limitsDisplayCurrency string // "gbp", "usd", "eur"

	rulesRequire   []string
	rulesAllowRoots []string
	rulesDenyAll   bool
	rulesDomains   []string
	rulesErr       error

	// Rules edit mode: e to enter, up/down select field, Enter to edit, Esc to exit, s to save
	rulesEditing    bool
	rulesFieldIndex int           // 0=require, 1=roots, 2=denyAll, 3=domains
	rulesInput      textinput.Model
	rulesInputActive bool         // true when typing in the input
	rulesSaveMsg    string        // "Saved" or error after s
	panicEnabled bool
	panicReason  string
	panicTTL     int
	panicErr     error

	prevNavIndex int
}

// NewModel creates a new BIOS model with default nav items. serverURL is the Ctrl Dot daemon URL (e.g. http://127.0.0.1:7777).
func NewModel(serverURL string) Model {
	items := []list.Item{
		NavItem{TitleVal: "Status", DescVal: ""},
		NavItem{TitleVal: "Limits", DescVal: ""},
		NavItem{TitleVal: "Rules", DescVal: ""},
		NavItem{TitleVal: "Panic", DescVal: ""},
	}
	l := list.New(items, NewNavDelegate(), navWidth, 14)
	l.SetShowHelp(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowPagination(false)
	l.Styles = StyledListStyles()
	l.Title = "Ctrl Dot"

	return Model{
		nav:          l,
		ServerURL:    serverURL,
		configPath:   ConfigPath(),
		prevNavIndex:     -1,
		agentLimits:      make(map[string]LimitsData),
		rulesInput:       newRulesTextInput(),
		limitsInput:      newRulesTextInput(),
		limitsDisplayCurrency: "gbp",
	}
}

func newRulesTextInput() textinput.Model {
	ti := textinput.New()
	ti.Placeholder = "value"
	ti.Width = 40
	ti.PromptStyle = Muted
	ti.TextStyle = NavItemStyle
	ti.CursorStyle = NavSelectedStyle
	return ti
}

// rulesFieldLabel and rulesFieldValue return display strings for the four rules fields.
func (m *Model) rulesFieldValue(i int) string {
	switch i {
	case 0:
		return strings.Join(m.rulesRequire, ", ")
	case 1:
		return strings.Join(m.rulesAllowRoots, ", ")
	case 2:
		if m.rulesDenyAll {
			return "yes"
		}
		return "no"
	case 3:
		return strings.Join(m.rulesDomains, ", ")
	}
	return ""
}

func rulesFieldLabel(i int) string {
	switch i {
	case 0:
		return "Require resolution"
	case 1:
		return "Filesystem allow roots"
	case 2:
		return "Network deny all"
	case 3:
		return "Allow domains"
	}
	return ""
}

// applyRulesInputValue parses m.rulesInput.Value() and sets the current rules field.
func (m *Model) applyRulesInputValue() {
	v := strings.TrimSpace(m.rulesInput.Value())
	switch m.rulesFieldIndex {
	case 0:
		m.rulesRequire = splitComma(v)
	case 1:
		m.rulesAllowRoots = splitComma(v)
	case 2:
		m.rulesDenyAll = strings.ToLower(v) == "yes" || strings.ToLower(v) == "true" || v == "1"
	case 3:
		m.rulesDomains = splitComma(v)
	}
}

func splitComma(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}

// Limits config fields (agents.default): 0=daily_gbp, 1=warn_pct, 2=throttle_pct, 3=hard_stop_pct, 4=max_iter, 5=display_currency
func limitsFieldLabel(i int) string {
	switch i {
	case 0:
		return "Daily budget"
	case 1:
		return "Warn at (%)"
	case 2:
		return "Throttle at (%)"
	case 3:
		return "Hard stop at (%)"
	case 4:
		return "Max iterations per action"
	case 5:
		return "Display currency"
	}
	return ""
}

// currencySymbol returns the symbol for the display currency.
func currencySymbol(cur string) string {
	switch strings.ToLower(cur) {
	case "usd":
		return "$"
	case "eur":
		return "€"
	default:
		return "£"
	}
}

// gbpToDisplay converts a GBP amount to the display currency (approximate rates).
func gbpToDisplay(gbp float64, cur string) float64 {
	switch strings.ToLower(cur) {
	case "usd":
		return gbp * 1.27
	case "eur":
		return gbp * 1.17
	default:
		return gbp
	}
}

func (m *Model) limitsFieldValue(i int) string {
	switch i {
	case 0:
		return strconv.FormatFloat(gbpToDisplay(m.limitsDailyBudget, m.limitsDisplayCurrency), 'f', 2, 64)
	case 1:
		return joinFloatPct(m.limitsWarnPct)
	case 2:
		return strconv.FormatFloat(m.limitsThrottlePct*100, 'f', 1, 64)
	case 3:
		return strconv.FormatFloat(m.limitsHardStopPct*100, 'f', 1, 64)
	case 4:
		return strconv.Itoa(m.limitsMaxIter)
	case 5:
		return strings.ToUpper(m.limitsDisplayCurrency)
	}
	return ""
}

func joinFloatPct(p []float64) string {
	if len(p) == 0 {
		return ""
	}
	var b strings.Builder
	for i, v := range p {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(strconv.FormatFloat(v*100, 'f', 1, 64))
	}
	return b.String()
}

func (m *Model) applyLimitsInputValue() {
	v := strings.TrimSpace(m.limitsInput.Value())
	switch m.limitsFieldIndex {
	case 0:
		if f, err := strconv.ParseFloat(v, 64); err == nil && f >= 0 {
			// User enters in display currency; convert back to GBP for storage
			cur := m.limitsDisplayCurrency
			switch strings.ToLower(cur) {
			case "usd":
				m.limitsDailyBudget = f / 1.27
			case "eur":
				m.limitsDailyBudget = f / 1.17
			default:
				m.limitsDailyBudget = f
			}
		}
	case 1:
		// comma-sep percentages, e.g. "70, 90"
		parts := splitComma(v)
		pct := make([]float64, 0, len(parts))
		for _, p := range parts {
			if f, err := strconv.ParseFloat(strings.TrimSpace(p), 64); err == nil && f >= 0 {
				pct = append(pct, f/100)
			}
		}
		if len(pct) > 0 {
			m.limitsWarnPct = pct
		}
	case 2:
		if f, err := strconv.ParseFloat(v, 64); err == nil && f >= 0 {
			m.limitsThrottlePct = f / 100
		}
	case 3:
		if f, err := strconv.ParseFloat(v, 64); err == nil && f >= 0 {
			m.limitsHardStopPct = f / 100
		}
	case 4:
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			m.limitsMaxIter = n
		}
	case 5:
		cur := strings.ToLower(v)
		if cur == "usd" || cur == "eur" || cur == "gbp" {
			m.limitsDisplayCurrency = cur
		}
	}
}

// Init runs once at startup; fetches status and panic so they show quickly.
func (m Model) Init() tea.Cmd {
	return tea.Batch(FetchHealth(m.ServerURL), FetchPanicState(m.ServerURL))
}

// fetchForPanel returns a Cmd to load data for the currently selected panel.
func (m *Model) fetchForPanel() tea.Cmd {
	idx := m.nav.Index()
	switch idx {
	case 0:
		return FetchHealth(m.ServerURL)
	case 1:
		return tea.Batch(FetchAgents(m.ServerURL), LoadLimitsConfig(m.configPath))
	case 2:
		return LoadRules(m.configPath)
	case 3:
		return FetchPanicState(m.ServerURL)
	}
	return nil
}

// Update handles messages.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		m.nav.SetSize(navWidth, msg.Height-4)
		return m, nil

	case healthResultMsg:
		m.healthOK = msg.OK
		m.healthVersion = msg.Version
		m.healthErr = msg.Err
		return m, nil

	case panicStateMsg:
		m.panicEnabled = msg.Enabled
		m.panicReason = msg.Reason
		m.panicTTL = msg.TTLSeconds
		m.panicErr = msg.Err
		return m, nil

	case agentsListMsg:
		m.agents = msg.Agents
		m.agentsErr = msg.Err
		m.agentLimits = make(map[string]LimitsData)
		if msg.Err == nil {
			var cmds []tea.Cmd
			for _, a := range msg.Agents {
				cmds = append(cmds, FetchAgentLimits(m.ServerURL, a.AgentID))
			}
			if len(cmds) > 0 {
				return m, tea.Batch(cmds...)
			}
		}
		return m, nil

	case agentLimitsMsg:
		if msg.Err == nil {
			m.agentLimits[msg.AgentID] = LimitsData{
				SpentGBP:   msg.SpentGBP,
				LimitGBP:   msg.LimitGBP,
				Percentage: msg.Percentage,
			}
		}
		return m, nil

	case rulesLoadedMsg:
		m.rulesRequire = msg.RequireResolution
		m.rulesAllowRoots = msg.AllowRoots
		m.rulesDenyAll = msg.NetworkDenyAll
		m.rulesDomains = msg.AllowDomains
		m.rulesErr = msg.Err
		return m, nil

	case limitsConfigLoadedMsg:
		m.limitsDailyBudget = msg.DailyBudgetGBP
		m.limitsWarnPct = msg.WarnPct
		m.limitsThrottlePct = msg.ThrottlePct
		m.limitsHardStopPct = msg.HardStopPct
		m.limitsMaxIter = msg.MaxIter
		m.limitsDisplayCurrency = msg.DisplayCurrency
		m.limitsErr = msg.Err
		return m, nil

	case saveResultMsg:
		if msg.Target == "rules" {
			if msg.Err != nil {
				m.rulesSaveMsg = "Error: " + msg.Err.Error()
			} else {
				m.rulesSaveMsg = "Saved."
			}
		} else if msg.Target == "limits" {
			if msg.Err != nil {
				m.limitsSaveMsg = "Error: " + msg.Err.Error()
			} else {
				m.limitsSaveMsg = "Saved."
			}
		}
		return m, nil

	case tea.KeyMsg:
		s := msg.String()
		if s == "q" || s == "ctrl+c" {
			return m, tea.Quit
		}
		// Limits panel: e=edit, s=save (default limits config)
		if m.nav.Index() == 1 {
			if m.limitsInputActive {
				if s == "enter" || s == "esc" {
					if s == "enter" {
						m.applyLimitsInputValue()
					}
					m.limitsInput.Blur()
					m.limitsInputActive = false
					return m, nil
				}
				var cmd tea.Cmd
				m.limitsInput, cmd = m.limitsInput.Update(msg)
				return m, cmd
			}
			if m.limitsEditing {
				switch s {
				case "esc":
					m.limitsEditing = false
					return m, nil
				case "s":
					m.limitsSaveMsg = "Saving..."
					return m, SaveLimitsConfig(m.configPath, m.limitsDailyBudget, m.limitsWarnPct, m.limitsThrottlePct, m.limitsHardStopPct, m.limitsMaxIter, m.limitsDisplayCurrency)
				case "up", "k":
					if m.limitsFieldIndex > 0 {
						m.limitsFieldIndex--
					}
					return m, nil
				case "down", "j":
					if m.limitsFieldIndex < 5 {
						m.limitsFieldIndex++
					}
					return m, nil
				case "enter":
					m.limitsInput.SetValue(m.limitsFieldValue(m.limitsFieldIndex))
					m.limitsInputActive = true
					return m, m.limitsInput.Focus()
				}
			} else {
				if strings.ToLower(s) == "e" || s == "enter" {
					m.limitsEditing = true
					m.limitsFieldIndex = 0
					m.limitsSaveMsg = ""
					return m, nil
				}
				if strings.ToLower(s) == "s" {
					m.limitsSaveMsg = "Saving..."
					return m, SaveLimitsConfig(m.configPath, m.limitsDailyBudget, m.limitsWarnPct, m.limitsThrottlePct, m.limitsHardStopPct, m.limitsMaxIter, m.limitsDisplayCurrency)
				}
			}
		}
		// Rules panel: e=edit, s=save, Esc=exit edit, up/down/Enter when editing
		if m.nav.Index() == 2 {
			if m.rulesInputActive {
				if s == "enter" || s == "esc" {
					// Apply or cancel
					if s == "enter" {
						m.applyRulesInputValue()
					}
					m.rulesInput.Blur()
					m.rulesInputActive = false
					return m, nil
				}
				var cmd tea.Cmd
				m.rulesInput, cmd = m.rulesInput.Update(msg)
				return m, cmd
			}
			if m.rulesEditing {
				switch s {
				case "esc":
					m.rulesEditing = false
					return m, nil
				case "s":
					m.rulesSaveMsg = "Saving..."
					return m, SaveRules(m.configPath, m.rulesRequire, m.rulesAllowRoots, m.rulesDenyAll, m.rulesDomains)
				case "up", "k":
					if m.rulesFieldIndex > 0 {
						m.rulesFieldIndex--
					}
					return m, nil
				case "down", "j":
					if m.rulesFieldIndex < 3 {
						m.rulesFieldIndex++
					}
					return m, nil
				case "enter":
					m.rulesInput.SetValue(m.rulesFieldValue(m.rulesFieldIndex))
					m.rulesInputActive = true
					return m, m.rulesInput.Focus()
				}
			} else {
				// View mode: e or E or Enter = edit, s or S = save
				if strings.ToLower(s) == "e" || s == "enter" {
					m.rulesEditing = true
					m.rulesFieldIndex = 0
					m.rulesSaveMsg = ""
					return m, nil
				}
				if strings.ToLower(s) == "s" {
					m.rulesSaveMsg = "Saving..."
					return m, SaveRules(m.configPath, m.rulesRequire, m.rulesAllowRoots, m.rulesDenyAll, m.rulesDomains)
				}
			}
		}
		// Panic panel: y/e = on, n/d = off
		if m.nav.Index() == 3 {
			if s == "y" || s == "e" {
				return m, SetPanicOn(m.ServerURL)
			}
			if s == "n" || s == "d" {
				return m, SetPanicOff(m.ServerURL)
			}
		}
		// r = refresh current panel
		if s == "r" {
			return m, m.fetchForPanel()
		}
	}

	var cmd tea.Cmd
	m.nav, cmd = m.nav.Update(msg)
	// When nav selection changes, fetch data for the new panel
	idx := m.nav.Index()
	if idx != m.prevNavIndex {
		m.prevNavIndex = idx
		fetchCmd := m.fetchForPanel()
		if fetchCmd != nil {
			return m, tea.Batch(cmd, fetchCmd)
		}
	}
	return m, cmd
}

// View renders the UI.
func (m Model) View() string {
	if !m.ready || m.width <= 0 || m.height <= 0 {
		return ""
	}
	ui := lipgloss.JoinVertical(lipgloss.Left,
		m.renderStatusBar(),
		m.renderBody(),
		m.renderFooter(),
	)
	return RenderFrame(m.width, m.height, ui)
}

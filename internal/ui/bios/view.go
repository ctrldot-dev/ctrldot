package bios

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m *Model) renderStatusBar() string {
	left := StatusBar.Render(" Ctrl Dot BIOS ")
	right := Muted.Render(" q: quit  r: refresh ")
	return lipgloss.JoinHorizontal(lipgloss.Top,
		left,
		lipgloss.NewStyle().Width(m.width - lipgloss.Width(left) - lipgloss.Width(right)).Render(""),
		right,
	)
}

func (m *Model) renderBody() string {
	navView := App.Width(navWidth).Height(m.height - 4).Render(m.nav.View())
	panel := m.renderPanel()
	panelWidth := m.width - navWidth - 2
	if panelWidth < 10 {
		panelWidth = 10
	}
	panelView := Panel.Width(panelWidth).Height(m.height - 4).Padding(1, 2).Render(panel)
	return lipgloss.JoinHorizontal(lipgloss.Top, navView, "  ", panelView)
}

func (m *Model) renderPanel() string {
	header := Muted.Render("Ctrl◻ Dot")
	idx := m.nav.Index()
	var body string
	switch idx {
	case 0:
		body = m.renderStatusPanel()
	case 1:
		body = m.renderLimitsPanel()
	case 2:
		body = m.renderRulesPanel()
	case 3:
		body = m.renderPanicPanel()
	default:
		body = Muted.Render("—")
	}
	return header + "\n\n" + body
}

func (m *Model) renderStatusPanel() string {
	title := Title.Render("Status")
	if m.healthErr != nil {
		return title + "\n\n" + Danger.Render("Daemon: not running") + "\n" +
			Muted.Render(m.healthErr.Error()) + "\n\nServer: "+m.ServerURL
	}
	if !m.healthOK {
		return title + "\n\n" + Muted.Render("Daemon: unknown state") + "\n\nServer: " + m.ServerURL
	}
	version := m.healthVersion
	if version == "" {
		version = "—"
	}
	return title + "\n\n" +
		NavItemStyle.Render("Daemon: running") + "\n" +
		RenderField("Version", version, false) + "\n" +
		RenderField("Server", m.ServerURL, false)
}

func (m *Model) renderLimitsPanel() string {
	title := Title.Render("Limits")
	if m.limitsInputActive {
		label := limitsFieldLabel(m.limitsFieldIndex)
		return title + "\n\n" + Muted.Render("Default limits") + "\n" + Muted.Render(label) + "\n" + m.limitsInput.View() + "\n\n" + Muted.Render("Enter: apply  Esc: cancel")
	}
	if m.limitsEditing {
		var lines []string
		for i := 0; i < 6; i++ {
			label := limitsFieldLabel(i)
			val := m.limitsFieldValue(i)
			if val == "" {
				val = "—"
			}
			if i == 0 {
				val = currencySymbol(m.limitsDisplayCurrency) + val
			}
			if i == m.limitsFieldIndex {
				lines = append(lines, NavSelectedStyle.Render("▶ "+label+": "+val))
			} else {
				lines = append(lines, NavItemStyle.Render("  "+label+": "+val))
			}
		}
		hint := Muted.Render("↑/↓ select  Enter: edit  Esc: exit  s: save")
		msg := ""
		if m.limitsSaveMsg != "" {
			msg = "\n" + Muted.Render(m.limitsSaveMsg)
		}
		return title + "\n\n" + Muted.Render("Default (config)") + "\n" + strings.Join(lines, "\n") + "\n\n" + hint + msg
	}
	// View mode: default limits + per-agent
	if m.limitsErr != nil {
		return title + "\n\n" + Danger.Render("Could not load limits config") + "\n" + Muted.Render(m.limitsErr.Error())
	}
	var out []string
	sym := currencySymbol(m.limitsDisplayCurrency)
	out = append(out, Muted.Render("Default (config)"))
	out = append(out, RenderField("Daily budget", sym+fmt.Sprintf("%.2f", gbpToDisplay(m.limitsDailyBudget, m.limitsDisplayCurrency)), false))
	out = append(out, RenderField("Warn at (%)", m.limitsFieldValue(1), false))
	out = append(out, RenderField("Throttle at (%)", m.limitsFieldValue(2), false))
	out = append(out, RenderField("Hard stop at (%)", m.limitsFieldValue(3), false))
	out = append(out, RenderField("Max iter/action", m.limitsFieldValue(4), false))
	out = append(out, RenderField("Display currency", strings.ToUpper(m.limitsDisplayCurrency), false))
	if m.limitsSaveMsg != "" {
		out = append(out, "", Muted.Render(m.limitsSaveMsg))
	}
	out = append(out, "", Muted.Render("e or Enter: edit  s: save  r: refresh"))
	out = append(out, "", Muted.Render("Per-agent (from daemon):"))
	if m.agentsErr != nil {
		out = append(out, Danger.Render("Could not load agents")+" "+Muted.Render(m.agentsErr.Error()))
	} else if len(m.agents) == 0 {
		out = append(out, Muted.Render("No agents registered."))
	} else {
		for _, a := range m.agents {
			id := a.AgentID
			lim, ok := m.agentLimits[id]
			if !ok {
				out = append(out, NavItemStyle.Render(id)+" "+Muted.Render("(loading…)"))
				continue
			}
			if lim.LimitGBP == 0 {
				out = append(out, NavItemStyle.Render(id+"  —"))
			} else {
				pct := lim.Percentage
				if pct <= 1 {
					pct = pct * 100
				}
				spent := gbpToDisplay(lim.SpentGBP, m.limitsDisplayCurrency)
				limit := gbpToDisplay(lim.LimitGBP, m.limitsDisplayCurrency)
				out = append(out, NavItemStyle.Render(fmt.Sprintf("%s  %s%.2f / %s%.2f  (%.1f%%)", id, sym, spent, sym, limit, pct)))
			}
		}
	}
	return title + "\n\n" + strings.Join(out, "\n")
}

func (m *Model) renderRulesPanel() string {
	title := Title.Render("Rules")
	if m.rulesErr != nil {
		return title + "\n\n" + Danger.Render("Could not load config") + "\n" + Muted.Render(m.rulesErr.Error()) + "\n\nConfig: " + m.configPath
	}
	if m.rulesInputActive {
		// Editing one field with text input
		label := rulesFieldLabel(m.rulesFieldIndex)
		return title + "\n\n" + Muted.Render(label) + "\n" + m.rulesInput.View() + "\n\n" + Muted.Render("Enter: apply  Esc: cancel")
	}
	if m.rulesEditing {
		// Edit mode: show 4 fields with selection
		var lines []string
		for i := 0; i < 4; i++ {
			label := rulesFieldLabel(i)
			val := m.rulesFieldValue(i)
			if val == "" {
				val = "—"
			}
			if i == m.rulesFieldIndex {
				lines = append(lines, NavSelectedStyle.Render("▶ "+label+": "+val))
			} else {
				lines = append(lines, NavItemStyle.Render("  "+label+": "+val))
			}
		}
		hint := Muted.Render("↑/↓ select  Enter: edit  Esc: exit  s: save")
		msg := ""
		if m.rulesSaveMsg != "" {
			msg = "\n" + Muted.Render(m.rulesSaveMsg)
		}
		return title + "\n\n" + strings.Join(lines, "\n") + "\n\n" + hint + msg
	}
	// View mode
	require := strings.Join(m.rulesRequire, ", ")
	if require == "" {
		require = "—"
	}
	roots := strings.Join(m.rulesAllowRoots, ", ")
	if roots == "" {
		roots = "—"
	}
	domains := strings.Join(m.rulesDomains, ", ")
	if domains == "" {
		domains = "—"
	}
	denyAll := "no"
	if m.rulesDenyAll {
		denyAll = "yes"
	}
	out := title + "\n\n" +
		RenderField("Require resolution", require, false) + "\n" +
		RenderField("Filesystem allow roots", roots, false) + "\n" +
		RenderField("Network deny all", denyAll, false) + "\n" +
		RenderField("Allow domains", domains, false)
	if m.rulesSaveMsg != "" {
		out += "\n\n" + Muted.Render(m.rulesSaveMsg)
	}
	out += "\n\n" + Muted.Render("e or Enter: edit  s: save  r: refresh")
	return out
}

func (m *Model) renderPanicPanel() string {
	title := Title.Render("Panic")
	if m.panicErr != nil {
		return title + "\n\n" + Danger.Render("Could not reach daemon") + "\n" + Muted.Render(m.panicErr.Error()) + "\n\n" +
			Muted.Render("y: enable  n: disable  r: refresh")
	}
	state := NavItemStyle.Render("OFF")
	if m.panicEnabled {
		state = NavSelectedStyle.Render("ON")
	}
	out := title + "\n\nState: " + state + "\n"
	if m.panicEnabled {
		if m.panicReason != "" {
			out += RenderField("Reason", m.panicReason, false) + "\n"
		}
		if m.panicTTL > 0 {
			out += RenderField("TTL", fmt.Sprintf("%d s", m.panicTTL), false) + "\n"
		}
	}
	out += "\n" + Muted.Render("y or e: enable   n or d: disable   r: refresh")
	return out
}

func (m *Model) renderFooter() string {
	left := Muted.Render(" ↑/↓ nav  enter select  r refresh ")
	right := Muted.Render(fmt.Sprintf(" %d×%d ", m.width, m.height))
	return lipgloss.JoinHorizontal(lipgloss.Top,
		left,
		lipgloss.NewStyle().Width(m.width - lipgloss.Width(left) - lipgloss.Width(right)).Render(""),
		right,
	)
}

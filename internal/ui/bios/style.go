package bios

import (
	"github.com/charmbracelet/lipgloss"
)

// Core palette — brand indigo background #3E2C8F, soft white, minimal accent.
var (
	ColorBG      = lipgloss.Color("#3E2C8F") // brand indigo background
	ColorText    = lipgloss.Color("#F4F4F6")  // soft white
	ColorMuted   = lipgloss.Color("#C9C7E6")  // muted lavender/grey for secondary text
	ColorAccent  = lipgloss.Color("#5A43B6")  // brighter indigo for selection highlight
	ColorBorder  = lipgloss.Color("#6B5BC5")  // border/lines
	ColorWarning = lipgloss.Color("#F2D36B")  // soft yellow warning (minimal)
	ColorDanger  = lipgloss.Color("#FF6B6B")  // soft red (minimal)
)

// Base styles reused by all views.
var (
	App = lipgloss.NewStyle().
		Background(ColorBG).
		Foreground(ColorText)

	Title = lipgloss.NewStyle().
		Foreground(ColorText).
		Bold(true)

	Muted = lipgloss.NewStyle().
		Foreground(ColorMuted)

	Border = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorBorder)

	NavItemStyle = lipgloss.NewStyle().
		Foreground(ColorText)

	NavSelectedStyle = lipgloss.NewStyle().
		Background(ColorAccent).
		Foreground(ColorText).
		Bold(true)

	StatusBar = lipgloss.NewStyle().
		Background(ColorBG).
		Foreground(ColorText)

	Panel = lipgloss.NewStyle().
		Background(ColorBG).
		Foreground(ColorText)

	Warning = lipgloss.NewStyle().Foreground(ColorWarning)
	Danger  = lipgloss.NewStyle().Foreground(ColorDanger)
)

// RenderFrame fills the terminal viewport with the brand background so no black shows.
func RenderFrame(w, h int, content string) string {
	return App.Width(w).Height(h).Render(content)
}

// RenderField renders a label: value line for the right panel; selected uses NavSelected.
func RenderField(label, value string, selected bool) string {
	l := Muted.Render(label)
	v := NavItemStyle.Render(value)
	if selected {
		return NavSelectedStyle.Padding(0, 1).Render(l + ": " + v)
	}
	return NavItemStyle.Padding(0, 1).Render(l + ": " + v)
}

// RenderModal renders a modal box with border and brand background.
func RenderModal(title, body string) string {
	box := Border.
		Background(ColorBG).
		Foreground(ColorText).
		Padding(1, 2).
		Render(Title.Render(title) + "\n\n" + body)
	return box
}

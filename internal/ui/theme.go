package ui

import "charm.land/lipgloss/v2"

// theme holds the styles used to render the UI. Colors default to the terminal
// palette so they adapt to dark and light backgrounds.
type theme struct {
	header lipgloss.Style
	banner lipgloss.Style
	status lipgloss.Style
	hint   lipgloss.Style
	help   lipgloss.Style
	pet    lipgloss.Style
}

func newTheme() theme {
	return theme{
		header: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("14")),
		banner: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("214")),
		status: lipgloss.NewStyle().Foreground(lipgloss.Color("12")),
		hint:   lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
		help:   lipgloss.NewStyle().Foreground(lipgloss.Color("250")),
		pet:    lipgloss.NewStyle().Foreground(lipgloss.Color("15")),
	}
}

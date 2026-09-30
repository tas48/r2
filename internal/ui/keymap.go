package ui

import tea "charm.land/bubbletea/v2"

func (m Model) handleKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "w":
		m.drink()
	case "b":
		m.takeBreak()
	case "?":
		m.help = !m.help
	case "space":
		// Pet interaction lands with the pet; nothing to do yet.
	}
	return m, nil
}

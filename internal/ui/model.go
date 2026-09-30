package ui

import (
	"fmt"
	"time"

	tea "charm.land/bubbletea/v2"
)

// tickMsg carries the simulation clock forward. A single self-rescheduling
// tick drives both animation and the reminder scheduler without a busy loop.
type tickMsg time.Time

const tickInterval = time.Second

// Model is the root Bubble Tea model.
type Model struct {
	width  int
	height int
	now    time.Time
}

// New builds the initial model.
func New() Model {
	return Model{}
}

// Init schedules the first tick.
func (m Model) Init() tea.Cmd {
	return tick()
}

// Update handles terminal events.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyPressMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	case tickMsg:
		m.now = time.Time(msg)
		return m, tick()
	}
	return m, nil
}

// View renders the whole UI in the alternate screen buffer.
func (m Model) View() tea.View {
	v := tea.NewView(m.content())
	v.AltScreen = true
	v.WindowTitle = "r2"
	return v
}

func (m Model) content() string {
	if m.width == 0 {
		return "r2\n\nwaiting for terminal size...\n\npress q to quit"
	}
	return fmt.Sprintf(
		"r2\n\ntick: %s\nsize: %dx%d\n\npress q to quit",
		m.now.Format("15:04:05"), m.width, m.height,
	)
}

func tick() tea.Cmd {
	return tea.Tick(tickInterval, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

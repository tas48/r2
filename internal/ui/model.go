package ui

import (
	"fmt"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/tas48/r2/internal/agent"
)

// tickMsg carries the simulation clock forward. A single self-rescheduling
// tick drives both animation and the reminder scheduler without a busy loop.
type tickMsg time.Time

const tickInterval = time.Second

// Model is the root Bubble Tea model. It owns the agent and drives it on ticks.
type Model struct {
	agent   *agent.Agent
	persist func(agent.State) error

	width  int
	height int
	now    time.Time

	banner string
	help   bool

	theme theme
}

// New builds the initial model. persist may be nil to disable persistence.
func New(ag *agent.Agent, persist func(agent.State) error) Model {
	return Model{agent: ag, persist: persist, theme: newTheme()}
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
		return m.handleKey(msg)
	case tickMsg:
		m.now = time.Time(msg)
		m.applyEvents(m.agent.Step(m.now))
		return m, tick()
	}
	return m, nil
}

func (m *Model) drink() {
	m.applyEvents(m.agent.Drink(m.now))
	m.save()
}

func (m *Model) takeBreak() {
	m.applyEvents(m.agent.LogBreak(m.now))
	m.save()
}

func (m *Model) applyEvents(events []agent.Event) {
	for _, ev := range events {
		switch e := ev.(type) {
		case agent.ReminderFired:
			m.banner = e.Text
		case agent.WaterLogged:
			m.banner = fmt.Sprintf("Logged %d ml (%d/%d ml)", m.agent.Config().BottleML, e.TotalML, e.GoalML)
		case agent.BreakLogged:
			m.banner = "Nice. Break logged."
		case agent.DayRolled:
			m.banner = "New day — counters reset."
		}
	}
}

func (m *Model) save() {
	if m.persist == nil {
		return
	}
	if err := m.persist(m.agent.Snapshot()); err != nil {
		m.banner = "Could not save progress."
	}
}

// View renders the whole UI in the alternate screen buffer.
func (m Model) View() tea.View {
	v := tea.NewView(m.render())
	v.AltScreen = true
	v.WindowTitle = "r2"
	return v
}

func tick() tea.Cmd {
	return tea.Tick(tickInterval, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

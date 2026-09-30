package ui

import (
	"fmt"
	"strings"
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
}

// New builds the initial model. persist may be nil to disable persistence.
func New(ag *agent.Agent, persist func(agent.State) error) Model {
	return Model{agent: ag, persist: persist}
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

func (m Model) handleKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "w":
		m.drink()
	case "b":
		m.takeBreak()
	case "space":
		// Pet interaction lands with the pet; nothing to do yet.
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

func (m Model) render() string {
	if m.width == 0 {
		return "r2\n\nwaiting for terminal size...\n\npress q to quit"
	}
	var b strings.Builder
	fmt.Fprintf(&b, "r2  %s\n\n", m.now.Format("15:04:05"))
	if m.banner != "" {
		fmt.Fprintf(&b, "%s\n\n", m.banner)
	}
	water := m.agent.Water()
	fmt.Fprintf(&b, "water  %d/%d ml\n", water.ConsumedML, water.GoalML)
	fmt.Fprintf(&b, "breaks %d\n", m.agent.Breaks().BreaksToday)
	b.WriteString("\nw drink · b break · q quit")
	return b.String()
}

func tick() tea.Cmd {
	return tea.Tick(tickInterval, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

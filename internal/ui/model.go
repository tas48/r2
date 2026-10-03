package ui

import (
	"fmt"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/tas48/r2/internal/agent"
	"github.com/tas48/r2/internal/pet"
)

// tickMsg carries the simulation clock forward. A single self-rescheduling
// tick drives both animation and the reminder scheduler without a busy loop.
type tickMsg time.Time

const tickInterval = time.Second

// Model is the root Bubble Tea model. It owns the agent and the pet, and
// drives both on ticks.
type Model struct {
	agent   *agent.Agent
	pet     *pet.Pet
	persist func(agent.State) error

	width  int
	height int
	now    time.Time

	banner string
	help   bool

	theme theme
}

// New builds the initial model. persist may be nil to disable persistence.
func New(ag *agent.Agent, p *pet.Pet, persist func(agent.State) error) Model {
	return Model{agent: ag, pet: p, persist: persist, theme: newTheme()}
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
		m.advance(time.Time(msg))
		return m, tick()
	}
	return m, nil
}

// advance moves time forward, steps the agent and the pet, and reacts to the
// events the agent produced.
func (m *Model) advance(now time.Time) {
	dt := now.Sub(m.now)
	if m.now.IsZero() || dt < 0 {
		dt = 0
	}
	m.now = now
	m.applyEvents(m.agent.Step(now))
	m.pet.Step(dt, now, m.stageMaxX())
}

func (m *Model) drink() {
	m.applyEvents(m.agent.Drink(m.now))
	m.save()
}

func (m *Model) takeBreak() {
	m.applyEvents(m.agent.LogBreak(m.now))
	m.save()
}

func (m *Model) interact() {
	m.pet.React(pet.Event{Kind: pet.EventInteract}, m.now)
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
		if petEvent, ok := reactTo(ev); ok {
			m.pet.React(petEvent, m.now)
		}
	}
}

// reactTo maps an agent event to a pet reaction.
func reactTo(ev agent.Event) (pet.Event, bool) {
	switch ev.(type) {
	case agent.WaterLogged:
		return pet.Event{Kind: pet.EventWater}, true
	case agent.BreakLogged:
		return pet.Event{Kind: pet.EventBreak}, true
	case agent.ReminderFired:
		return pet.Event{Kind: pet.EventReminder}, true
	case agent.DayRolled:
		return pet.Event{Kind: pet.EventDayRolled}, true
	default:
		return pet.Event{}, false
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

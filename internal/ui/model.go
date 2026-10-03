package ui

import (
	"fmt"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/tas48/r2/internal/agent"
	"github.com/tas48/r2/internal/weather"
)

// tickMsg carries the simulation clock forward. A single self-rescheduling
// tick drives the reminder scheduler without a busy loop.
type tickMsg time.Time

const tickInterval = time.Second

// Model is the root Bubble Tea model. It owns the agent, drives it on ticks and
// renders a responsive dashboard.
type Model struct {
	agent   *agent.Agent
	persist func(agent.State) error
	weather weather.Provider
	city    string

	width  int
	height int
	now    time.Time

	banner string

	forecast   weather.Forecast
	hasWeather bool
	weatherErr error

	theme theme
}

// New builds the initial model. persist may be nil to disable persistence;
// provider may be nil to disable weather.
func New(ag *agent.Agent, provider weather.Provider, city string, persist func(agent.State) error) Model {
	return Model{agent: ag, weather: provider, city: city, persist: persist, theme: newTheme()}
}

// Init schedules the first tick and, when configured, the first forecast fetch.
func (m Model) Init() tea.Cmd {
	if m.weather == nil {
		return tick()
	}
	return tea.Batch(tick(), m.fetch())
}

// Update handles terminal events.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyPressMsg:
		return m.handleKey(msg)
	case weatherMsg:
		m.hasWeather = true
		m.forecast = msg.forecast
		m.weatherErr = msg.err
		m.applyEvents(m.agent.ApplyWeather(msg.forecast.RainProb, msg.forecast.TempC, msg.forecast.RainLikely, msg.err, m.now))
	case tickMsg:
		m.now = time.Time(msg)
		return m, m.applyEvents(m.agent.Step(m.now))
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

// applyEvents renders the events and returns any follow-up command, such as an
// asynchronous weather fetch.
func (m *Model) applyEvents(events []agent.Event) tea.Cmd {
	var cmd tea.Cmd
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
		case agent.WeatherRequested:
			if m.weather != nil {
				cmd = m.fetch()
			}
		case agent.WeatherUpdated:
			if e.Err != nil {
				m.weatherErr = e.Err
			} else if e.RainLikely {
				m.banner = fmt.Sprintf("Rain likely (%d%%) — close the window", e.RainProb)
			}
		}
	}
	return cmd
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

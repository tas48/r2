package app

import (
	"fmt"
	"net/http"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/tas48/r2/internal/agent"
	"github.com/tas48/r2/internal/config"
	"github.com/tas48/r2/internal/store"
	"github.com/tas48/r2/internal/ui"
	"github.com/tas48/r2/internal/weather"
)

// Run wires config, persistence, the agent, weather and the TUI, then runs the
// program.
func Run() error {
	now := time.Now()

	cfgPath, err := config.Path()
	if err != nil {
		return err
	}
	cfg, err := config.Ensure(cfgPath)
	if err != nil {
		return err
	}

	statePath, err := store.Path()
	if err != nil {
		return err
	}
	st, err := store.New(statePath).Load(now)
	if err != nil {
		return err
	}

	ag := agent.New(toAgentConfig(cfg), now)
	ag.Seed(toAgentState(st), now)

	provider := weather.NewOpenMeteo(&http.Client{Timeout: 10 * time.Second}, cfg.RainThreshold, cfg.RainHorizonH)

	if _, err := tea.NewProgram(ui.New(ag, provider, cfg.City, persistFunc(statePath))).Run(); err != nil {
		return fmt.Errorf("running tui: %w", err)
	}
	return nil
}

func persistFunc(path string) func(agent.State) error {
	st := store.New(path)
	return func(s agent.State) error {
		return st.Save(toStoreState(s))
	}
}

func toAgentConfig(c config.Config) agent.Config {
	return agent.Config{
		WaterGoalML:  c.WaterGoalML,
		BottleML:     c.BottleML,
		WaterEvery:   time.Duration(c.WaterEvery),
		BreakEvery:   time.Duration(c.BreakEvery),
		WeatherEvery: time.Duration(c.WeatherEvery),
	}
}

func toAgentState(s store.State) agent.State {
	return agent.State{
		Day:         s.Day,
		WaterML:     s.WaterML,
		LastDrink:   s.LastDrink,
		LastBreak:   s.LastBreak,
		BreaksToday: s.BreaksToday,
	}
}

func toStoreState(s agent.State) store.State {
	return store.State{
		Day:         s.Day,
		WaterML:     s.WaterML,
		LastDrink:   s.LastDrink,
		LastBreak:   s.LastBreak,
		BreaksToday: s.BreaksToday,
	}
}

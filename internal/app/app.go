package app

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/tas48/r2/internal/agent"
	"github.com/tas48/r2/internal/config"
	"github.com/tas48/r2/internal/geo"
	"github.com/tas48/r2/internal/store"
	"github.com/tas48/r2/internal/ui"
	"github.com/tas48/r2/internal/weather"
)

// Run wires config, persistence, the agent, location, weather and the TUI,
// then runs the program.
func Run() error {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))
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

	client := &http.Client{Timeout: 10 * time.Second}
	place := resolvePlace(logger, client, cfg.City)
	provider := weather.NewOpenMeteo(client, cfg.RainThreshold, cfg.RainHorizonH)

	if _, err := tea.NewProgram(ui.New(ag, provider, place, persistFunc(statePath))).Run(); err != nil {
		return fmt.Errorf("running tui: %w", err)
	}
	return nil
}

// Place is where the dashboard should fetch weather from. When coordinates are
// known the provider uses them directly (more precise than the city name).
type Place = ui.Place

// resolvePlace tries precise location, then reverse-geocodes it to a readable
// name, and finally falls back to the configured city.
func resolvePlace(logger *slog.Logger, client *http.Client, fallbackCity string) ui.Place {
	ctx, cancel := contextWithTimeout(8 * time.Second)
	defer cancel()

	location, err := geo.DefaultResolver(logger).Locate(ctx)
	if err != nil {
		logger.Warn("using configured city; location lookup failed", "city", fallbackCity, "error", err)
		return ui.Place{Name: fallbackCity}
	}

	name := firstNonEmpty(location.City, fallbackCity)
	reverse := weather.NewReverseGeocoder(client)
	if refined, err := reverse.Name(ctx, location.Latitude, location.Longitude); err == nil && refined != "" {
		name = refined
	}
	return ui.Place{
		Name:      name,
		Latitude:  location.Latitude,
		Longitude: location.Longitude,
		HasCoords: true,
	}
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
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

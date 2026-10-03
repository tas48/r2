package ui

import (
	"context"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/tas48/r2/internal/weather"
)

// weatherMsg carries the result of an asynchronous forecast fetch.
type weatherMsg struct {
	forecast weather.Forecast
	err      error
}

// fetch runs the network call off the update loop.
func (m Model) fetch() tea.Cmd {
	provider := m.weather
	city := m.city
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		forecast, err := provider.Forecast(ctx, city)
		return weatherMsg{forecast: forecast, err: err}
	}
}

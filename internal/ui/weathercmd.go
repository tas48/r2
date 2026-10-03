package ui

import (
	"context"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/tas48/r2/internal/weather"
)

// Place is where weather comes from: a display name plus optional coordinates
// resolved from IP or the Windows Location API.
type Place struct {
	Name      string
	Latitude  float64
	Longitude float64
	HasCoords bool
}

// weatherMsg carries the result of an asynchronous forecast fetch.
type weatherMsg struct {
	forecast weather.Forecast
	err      error
}

// fetch runs the network call off the update loop, preferring coordinates.
func (m Model) fetch() tea.Cmd {
	provider := m.weather
	place := m.place
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
		defer cancel()
		var (
			forecast weather.Forecast
			err      error
		)
		if place.HasCoords {
			forecast, err = provider.ForecastAt(ctx, place.Name, place.Latitude, place.Longitude)
		} else {
			forecast, err = provider.Forecast(ctx, place.Name)
		}
		return weatherMsg{forecast: forecast, err: err}
	}
}

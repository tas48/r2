package ui

import (
	"fmt"
	"strings"
	"testing"
	"time"

	golden "github.com/charmbracelet/x/exp/golden"

	"github.com/tas48/r2/internal/agent"
	"github.com/tas48/r2/internal/weather"
)

func sampleForecast() weather.Forecast {
	now := base
	hours := make([]weather.HourSample, 0, 6)
	for i := 0; i < 6; i++ {
		hours = append(hours, weather.HourSample{
			Time:     now.Add(time.Duration(i) * time.Hour),
			TempC:    20 + float64(i),
			RainProb: 10 * i,
			WindKph:  12,
		})
	}
	return weather.Forecast{
		Location:   "Rio Paranaiba",
		RainProb:   70,
		TempC:      24,
		FeelsLikeC: 26,
		WindKph:    18,
		Humidity:   63,
		UVIndex:    7,
		IsDay:      true,
		NextHours:  hours,
		Daily: []weather.DaySample{
			{Date: now, TempMinC: 18, TempMaxC: 29, RainProb: 70, WindKph: 18},
			{Date: now.Add(24 * time.Hour), TempMinC: 17, TempMaxC: 27, RainProb: 30, WindKph: 14},
			{Date: now.Add(48 * time.Hour), TempMinC: 16, TempMaxC: 26, RainProb: 10, WindKph: 10},
		},
	}
}

func sizedModel(width, height int) Model {
	ag := agent.New(agent.DefaultConfig(), base)
	ag.Drink(base)
	m := New(ag, nil, Place{Name: "Rio Paranaiba"}, nil)
	m.width = width
	m.height = height
	m.now = base
	m.hasWeather = true
	m.forecast = sampleForecast()
	return m
}

func TestRenderSizeMatrix(t *testing.T) {
	sizes := [][2]int{{10, 3}, {20, 8}, {30, 10}, {40, 15}, {100, 8}, {30, 30}, {80, 24}, {120, 40}}
	var b strings.Builder
	for _, s := range sizes {
		fmt.Fprintf(&b, "== %dx%d ==\n%s\n\n", s[0], s[1], plain(sizedModel(s[0], s[1]).render()))
	}
	golden.RequireEqual(t, b.String())
}

func TestRenderWeatherAlert(t *testing.T) {
	m := sizedModel(60, 20)
	m.forecast.Alerts = append(m.forecast.Alerts, weather.Alert{Kind: weather.AlertHeat, Text: "Heat alert: feels like 39°C"})
	golden.RequireEqual(t, plain(m.render()))
}

func TestRenderTinyWeather(t *testing.T) {
	m := sizedModel(20, 4)
	golden.RequireEqual(t, plain(m.render()))
}

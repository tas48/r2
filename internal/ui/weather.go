package ui

import (
	"fmt"
	"math"
	"strings"

	"github.com/tas48/r2/internal/weather"
)

// weatherCondition classifies a forecast into an icon and a label.
type weatherCondition struct {
	icon  string
	label string
}

func conditionFor(f weather.Forecast) weatherCondition {
	switch {
	case f.WindKph >= 40:
		return weatherCondition{"≋", "windy"}
	case f.RainLikely || f.RainProb >= 50:
		return weatherCondition{"☔", "rain"}
	case f.RainProb >= 25:
		return weatherCondition{"☁", "clouds"}
	case f.TempC <= 5:
		return weatherCondition{"❄", "cold"}
	default:
		return weatherCondition{"☀", "clear"}
	}
}

func (m Model) weatherLine() string {
	if !m.hasWeather {
		return "weather  loading…"
	}
	if m.weatherErr != nil {
		return "weather  unavailable"
	}
	f := m.forecast
	c := conditionFor(f)
	temperature := fmt.Sprintf("%.0f°C", f.TempC)
	if m.width >= 30 {
		return fmt.Sprintf("weather %s %s %s rain %d%% wind %.0fkm/h", c.icon, c.label, temperature, f.RainProb, f.WindKph)
	}
	return fmt.Sprintf("%s %s rain %d%%", c.icon, temperature, f.RainProb)
}

// temperatureBar renders a compact bar for a -5..35 C range.
func temperatureBar(tempC float64, width int) string {
	if width <= 0 {
		return ""
	}
	ratio := (tempC + 5) / 40
	if ratio < 0 {
		ratio = 0
	}
	if ratio > 1 {
		ratio = 1
	}
	filled := int(math.Round(ratio * float64(width)))
	return strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
}

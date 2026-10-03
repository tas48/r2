package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/tas48/r2/internal/weather"
)

// weatherIcon returns an icon for a given condition.
func weatherIcon(rainProb int, windKph, tempC float64, isDay bool) string {
	switch {
	case windKph >= 40:
		return "≋"
	case rainProb >= 50:
		return "☔"
	case rainProb >= 25:
		return "☁"
	case tempC <= 3:
		return "❄"
	case !isDay:
		return "☾"
	default:
		return "☀"
	}
}

func (m Model) weatherLines(width int) []string {
	if !m.hasWeather {
		return []string{"weather  loading…"}
	}
	if m.weatherErr != nil {
		return []string{"weather  unavailable"}
	}
	f := m.forecast
	icon := weatherIcon(f.RainProb, f.WindKph, f.TempC, f.IsDay)
	place := f.Location
	if place == "" {
		place = m.place.Name
	}

	if width < 20 {
		return []string{truncate(fmt.Sprintf("%s %.0f° rain %d%%", icon, f.TempC, f.RainProb), width)}
	}

	lines := []string{
		truncate(fmt.Sprintf("%s %s  %.0f°C  rain %d%%", icon, place, f.TempC, f.RainProb), width),
		truncate(fmt.Sprintf("  feels %.0f° · hum %d%% · wind %.0f km/h · uv %.0f",
			f.FeelsLikeC, f.Humidity, f.WindKph, f.UVIndex), width),
	}

	if next := nextStrip(f, 5); next != "" && width >= 30 {
		lines = append(lines, truncate(next, width))
	}
	if days := dayStrip(f, 3); days != "" && width >= 30 {
		lines = append(lines, truncate(days, width))
	}
	return lines
}

// nextStrip renders the next few hours as "HH° icon" pairs.
func nextStrip(f weather.Forecast, count int) string {
	if len(f.NextHours) == 0 {
		return ""
	}
	var parts []string
	for i, h := range f.NextHours {
		if i >= count {
			break
		}
		icon := weatherIcon(h.RainProb, h.WindKph, h.TempC, true)
		parts = append(parts, fmt.Sprintf("%s%.0f°%s", h.Time.Format("15"), h.TempC, icon))
	}
	return strings.Join(parts, " ")
}

// dayStrip renders upcoming days as "weekday icon min/max".
func dayStrip(f weather.Forecast, count int) string {
	if len(f.Daily) == 0 {
		return ""
	}
	var parts []string
	for i, d := range f.Daily {
		if i >= count {
			break
		}
		label := d.Date.Format("Mon")
		if i == 0 {
			label = "today"
		}
		icon := weatherIcon(d.RainProb, d.WindKph, d.TempMaxC, true)
		parts = append(parts, fmt.Sprintf("%s%.0f/%.0f%s", label, d.TempMinC, d.TempMaxC, icon))
	}
	return strings.Join(parts, "  ")
}

// alertLine returns the highest-priority alert as a single line, or "".
func (m Model) alertLine() string {
	if !m.hasWeather || m.weatherErr != nil || len(m.forecast.Alerts) == 0 {
		return ""
	}
	return "⚠ " + m.forecast.Alerts[0].Text
}

// temperatureBar renders a compact bar for a -5..35 C range.
func temperatureBar(tempC float64, width int) string {
	if width <= 0 {
		return ""
	}
	ratio := (tempC + 5) / 40
	ratio = clamp01(ratio)
	filled := int(ratio*float64(width) + 0.5)
	return strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

var _ = time.Now

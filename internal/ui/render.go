package ui

import (
	"fmt"
	"strings"
)

// render lays out the active dashboard elements for the current size.
func (m Model) render() string {
	l := compute(m.width, m.height, m.banner != "")
	if l.Width == 0 {
		return ""
	}
	rows := make([]string, 0, l.Height)
	if l.Header {
		rows = append(rows, m.theme.header.Render(truncate(m.headerLine(), l.Width)))
	}
	if l.Banner {
		rows = append(rows, m.theme.banner.Render(truncate(m.banner, l.Width)), "")
	}
	if l.Weather {
		rows = append(rows, m.theme.weather.Render(truncate(m.weatherLine(), l.Width)))
	}
	if l.Status {
		rows = append(rows, m.theme.status.Render(truncate(m.statusLine(), l.Width)))
	}
	if l.Hint {
		rows = append(rows, m.theme.hint.Render(truncate(m.hintLine(), l.Width)))
	}
	return strings.Join(rows, "\n")
}

func (m Model) headerLine() string {
	return fmt.Sprintf("r2  %s", m.now.Format("15:04"))
}

func (m Model) statusLine() string {
	water := m.agent.Water()
	breaks := m.agent.Breaks().BreaksToday
	if m.width >= 30 {
		return fmt.Sprintf("water %d/%dml %s breaks %d", water.ConsumedML, water.GoalML, waterBar(water.ConsumedML, water.GoalML), breaks)
	}
	return fmt.Sprintf("%d/%dml", water.ConsumedML, water.GoalML)
}

func (m Model) hintLine() string {
	switch {
	case m.width >= 30:
		return "w drink · b break · ? help · q quit"
	case m.width >= 20:
		return "? help · q quit"
	default:
		return "? help"
	}
}

// waterBar renders an 8-cell progress bar for the daily water goal.
func waterBar(consumed, goal int) string {
	const width = 8
	if goal <= 0 {
		return strings.Repeat("░", width)
	}
	filled := consumed * width / goal
	if filled > width {
		filled = width
	}
	return strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
}

func (m Model) helpLines() []string {
	items := []string{
		"w      drink a bottle of water",
		"b      log a break",
		"?      toggle this help",
		"q      quit",
	}
	if m.weather != nil {
		items = append([]string{fmt.Sprintf("weather for %s", m.city), ""}, items...)
	}
	lines := make([]string, len(items))
	for i, item := range items {
		lines[i] = truncate(item, m.width)
	}
	return lines
}

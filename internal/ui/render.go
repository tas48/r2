package ui

import (
	"fmt"
	"strings"
)

// render lays out the active elements for the current terminal size.
func (m Model) render() string {
	l := compute(m.width, m.height, m.banner != "")
	if l.Width == 0 {
		return ""
	}
	rows := make([]string, 0, l.Height)
	if l.Header {
		rows = append(rows, m.theme.header.Render(m.headerLine()))
	}
	if l.Banner {
		rows = append(rows, m.theme.banner.Render(truncate(m.banner, l.Width)), "")
	}
	rows = append(rows, m.stageLines(l.Stage.H)...)
	if l.Status {
		rows = append(rows, m.theme.status.Render(m.statusLine()))
	}
	if l.Hint {
		rows = append(rows, m.theme.hint.Render(m.hintLine()))
	}
	return strings.Join(rows, "\n")
}

func (m Model) headerLine() string {
	return fmt.Sprintf("r2  %s", m.now.Format("15:04"))
}

func (m Model) statusLine() string {
	water := m.agent.Water()
	breaks := m.agent.Breaks().BreaksToday
	if m.width >= 24 {
		return fmt.Sprintf("water %d/%dml · breaks %d", water.ConsumedML, water.GoalML, breaks)
	}
	return fmt.Sprintf("%d/%dml", water.ConsumedML, water.GoalML)
}

func (m Model) hintLine() string {
	switch {
	case m.width >= 38:
		return "w drink · b break · ? help · q quit"
	case m.width >= 20:
		return "? help · q quit"
	default:
		return "? help"
	}
}

// stageLines returns the pet area. Until the pet lands it is blank, or shows
// the help overlay when toggled.
func (m Model) stageLines(rows int) []string {
	lines := make([]string, rows)
	if !m.help {
		return lines
	}
	for i, line := range m.helpLines() {
		if i >= rows {
			break
		}
		if line == "" {
			continue
		}
		lines[i] = m.theme.help.Render(line)
	}
	return lines
}

func (m Model) helpLines() []string {
	items := []string{
		"help",
		"",
		"w      drink a bottle of water",
		"b      log a break",
		"space  interact with the pet",
		"?      toggle this help",
		"q      quit",
	}
	lines := make([]string, len(items))
	for i, item := range items {
		lines[i] = truncate(item, m.width)
	}
	return lines
}

// truncate shortens s to width runes, appending an ellipsis when clipped.
func truncate(s string, width int) string {
	if width <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) <= width {
		return s
	}
	if width == 1 {
		return "…"
	}
	return string(runes[:width-1]) + "…"
}

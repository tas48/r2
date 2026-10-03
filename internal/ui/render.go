package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/tas48/r2/internal/pet"
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
		return "w drink · b break · space pet · ? help · q quit"
	case m.width >= 20:
		return "? help · q quit"
	default:
		return "? help"
	}
}

// stageLines returns the pet area: the pet sprite centered vertically, or the
// help overlay when toggled.
func (m Model) stageLines(rows int) []string {
	lines := make([]string, rows)
	if rows <= 0 {
		return lines
	}
	if m.help {
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
	sprite := pet.SpriteFor(m.pet.State(), spriteSizeFor(rows))
	frame := sprite.Frames[frameIndex(sprite, m.now.Sub(m.pet.Since()))]
	x := m.pet.X()
	if x < 0 {
		x = (m.width - sprite.Width) / 2
	}
	top := (rows - sprite.Height) / 2
	if top < 0 {
		top = 0
	}
	for i, line := range frame {
		row := top + i
		if row >= rows {
			break
		}
		lines[row] = m.theme.pet.Render(placeAt(line, x))
	}
	return lines
}

// stageMaxX is the highest column the pet may walk to for the current layout.
func (m Model) stageMaxX() int {
	if m.pet == nil {
		return 0
	}
	l := compute(m.width, m.height, m.banner != "")
	if l.Stage.H <= 0 {
		return 0
	}
	width := pet.SpriteFor(m.pet.State(), spriteSizeFor(l.Stage.H)).Width
	if maxX := l.Stage.W - width; maxX > 0 {
		return maxX
	}
	return 0
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

func spriteSizeFor(stageRows int) pet.Size {
	switch {
	case stageRows >= 3:
		return pet.SizeFull
	case stageRows == 2:
		return pet.SizeCompact
	default:
		return pet.SizeMini
	}
}

func frameIndex(sprite pet.Sprite, elapsed time.Duration) int {
	if len(sprite.Frames) == 0 || sprite.FrameDur <= 0 || elapsed <= 0 {
		return 0
	}
	return int(elapsed/sprite.FrameDur) % len(sprite.Frames)
}

// placeAt left-pads a line to column x, trimming trailing spaces to keep the
// output free of trailing whitespace.
func placeAt(line string, x int) string {
	line = strings.TrimRight(line, " ")
	if x <= 0 {
		return line
	}
	return strings.Repeat(" ", x) + line
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

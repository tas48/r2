package ui

import (
	"fmt"
	"strings"
	"testing"

	golden "github.com/charmbracelet/x/exp/golden"

	"github.com/tas48/r2/internal/agent"
)

func sizedModel(width, height int) Model {
	ag := agent.New(agent.DefaultConfig(), base)
	ag.Drink(base)
	m := New(ag, nil)
	m.width = width
	m.height = height
	m.now = base
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

func TestRenderBanner(t *testing.T) {
	m := sizedModel(40, 15)
	m.banner = "Time to drink water (500/2000 ml today)"
	golden.RequireEqual(t, plain(m.render()))
}

func TestRenderTinyBanner(t *testing.T) {
	m := sizedModel(20, 4)
	m.banner = "Time to drink water"
	golden.RequireEqual(t, plain(m.render()))
}

func TestRenderHelp(t *testing.T) {
	m := sizedModel(60, 20)
	m.help = true
	golden.RequireEqual(t, plain(m.render()))
}

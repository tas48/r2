package ui

import (
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/tas48/r2/internal/agent"
)

var base = time.Date(2026, time.September, 30, 9, 0, 0, 0, time.Local)

func key(code rune, text string) tea.KeyPressMsg {
	return tea.KeyPressMsg(tea.Key{Code: code, Text: text})
}

func newModel() Model {
	return New(agent.New(agent.DefaultConfig(), base), nil, "", nil)
}

func TestQuitKey(t *testing.T) {
	m := newModel()
	_, cmd := m.Update(key('q', "q"))
	if cmd == nil {
		t.Fatal("expected a command from the quit key")
	}
}

func TestDrinkKeyLogsWaterAndPersists(t *testing.T) {
	ag := agent.New(agent.DefaultConfig(), base)
	var saved *agent.State
	m := New(ag, nil, "", func(s agent.State) error { saved = &s; return nil })
	m.now = base

	updated, _ := m.Update(key('w', "w"))
	if got := ag.Water().ConsumedML; got != 500 {
		t.Fatalf("expected 500 ml logged, got %d", got)
	}
	if saved == nil || saved.WaterML != 500 {
		t.Fatalf("expected persisted 500 ml, got %#v", saved)
	}
	if updated.(Model).banner == "" {
		t.Fatal("expected a banner after logging water")
	}
}

func TestBreakKeyLogsBreak(t *testing.T) {
	ag := agent.New(agent.DefaultConfig(), base)
	m := New(ag, nil, "", nil)
	m.now = base

	m.Update(key('b', "b"))
	if ag.Breaks().BreaksToday != 1 {
		t.Fatalf("expected one break, got %d", ag.Breaks().BreaksToday)
	}
}

func TestTickFiresReminderBanner(t *testing.T) {
	ag := agent.New(agent.DefaultConfig(), base)
	m := New(ag, nil, "", nil)

	updated, _ := m.Update(tickMsg(base.Add(30 * time.Minute)))
	if updated.(Model).banner == "" {
		t.Fatal("expected a reminder banner after the water interval")
	}
}

func TestTickBeforeIntervalKeepsBannerEmpty(t *testing.T) {
	ag := agent.New(agent.DefaultConfig(), base)
	m := New(ag, nil, "", nil)

	updated, _ := m.Update(tickMsg(base.Add(10 * time.Minute)))
	if updated.(Model).banner != "" {
		t.Fatalf("expected no banner, got %q", updated.(Model).banner)
	}
}

package agent

import (
	"testing"
	"time"
)

func TestSeedRestoresCounters(t *testing.T) {
	a := newTestAgent()
	st := State{
		Day:         base.Format(dayLayout),
		WaterML:     1000,
		LastDrink:   base.Add(time.Minute),
		LastBreak:   base.Add(2 * time.Minute),
		BreaksToday: 2,
	}
	a.Seed(st, base)

	if got := a.Snapshot(); got.WaterML != st.WaterML || got.BreaksToday != st.BreaksToday {
		t.Fatalf("seed did not restore counters: %#v", got)
	}
	if got := a.Water().ConsumedML; got != 1000 {
		t.Fatalf("expected 1000 ml after seed, got %d", got)
	}
}

func TestSeedRestartsIntervals(t *testing.T) {
	a := newTestAgent()
	a.Seed(State{Day: base.Format(dayLayout)}, base)
	// The water interval restarts from the seed time, not from construction.
	if events := a.Step(base.Add(29 * time.Minute)); len(events) != 0 {
		t.Fatalf("expected no events before interval, got %v", events)
	}
	if got := countReminders(a.Step(base.Add(30*time.Minute)), KindWater); got != 1 {
		t.Fatalf("expected one water reminder, got %d", got)
	}
}

func TestSeedStaleDayRollsOverOnNextStep(t *testing.T) {
	a := newTestAgent()
	a.Seed(State{Day: "2026-09-29", WaterML: 1500, BreaksToday: 3}, base)

	events := a.Step(base)
	rolled := false
	for _, ev := range events {
		if _, ok := ev.(DayRolled); ok {
			rolled = true
		}
	}
	if !rolled {
		t.Fatalf("expected DayRolled for a stale seed, got %v", events)
	}
	if got := a.Water().ConsumedML; got != 0 {
		t.Fatalf("expected water reset after rollover, got %d", got)
	}
}

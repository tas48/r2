package agent

import (
	"testing"
	"time"
)

var base = time.Date(2026, time.September, 30, 9, 0, 0, 0, time.Local)

func newTestAgent() *Agent {
	return New(DefaultConfig(), base)
}

func TestStepDoesNotFireBeforeInterval(t *testing.T) {
	a := newTestAgent()
	if events := a.Step(base.Add(29 * time.Minute)); len(events) != 0 {
		t.Fatalf("expected no events, got %v", events)
	}
}

func TestStepFiresWaterAtInterval(t *testing.T) {
	a := newTestAgent()
	events := a.Step(base.Add(30 * time.Minute))
	if len(events) != 1 {
		t.Fatalf("expected one event, got %d", len(events))
	}
	rem, ok := events[0].(ReminderFired)
	if !ok || rem.Kind != KindWater {
		t.Fatalf("expected water reminder, got %#v", events[0])
	}
}

func TestStepDoesNotBurstMissedReminders(t *testing.T) {
	a := newTestAgent()
	// Jump past three water intervals in a single step: only one water reminder.
	if got := countReminders(a.Step(base.Add(100*time.Minute)), KindWater); got != 1 {
		t.Fatalf("expected a single water reminder, got %d", got)
	}
	// Schedule stays realigned: the next one fires 20 minutes later.
	if got := countReminders(a.Step(base.Add(110*time.Minute)), KindWater); got != 0 {
		t.Fatalf("expected no water reminder mid-interval, got %d", got)
	}
	if got := countReminders(a.Step(base.Add(120*time.Minute)), KindWater); got != 1 {
		t.Fatalf("expected water reminder at the realigned interval, got %d", got)
	}
}

func countReminders(events []Event, kind Kind) int {
	n := 0
	for _, ev := range events {
		if rem, ok := ev.(ReminderFired); ok && rem.Kind == kind {
			n++
		}
	}
	return n
}

func TestDrinkLogsAndResetsWaterSchedule(t *testing.T) {
	a := newTestAgent()
	events := a.Drink(base.Add(10 * time.Minute))
	if len(events) != 1 {
		t.Fatalf("expected water logged event, got %d", len(events))
	}
	logged, ok := events[0].(WaterLogged)
	if !ok || logged.TotalML != 500 || logged.GoalML != 2000 {
		t.Fatalf("unexpected water log: %#v", events[0])
	}
	if events := a.Step(base.Add(35 * time.Minute)); len(events) != 0 {
		t.Fatalf("expected no reminder before postponed interval, got %v", events)
	}
	if events := a.Step(base.Add(40 * time.Minute)); len(events) != 1 {
		t.Fatalf("expected reminder after postponed interval, got %v", events)
	}
}

func TestDrinkAccumulates(t *testing.T) {
	a := newTestAgent()
	a.Drink(base)
	a.Drink(base.Add(time.Minute))
	a.Drink(base.Add(2 * time.Minute))
	if got := a.Water().ConsumedML; got != 1500 {
		t.Fatalf("expected 1500 ml, got %d", got)
	}
}

func TestLogBreak(t *testing.T) {
	a := newTestAgent()
	events := a.LogBreak(base.Add(50 * time.Minute))
	if len(events) != 1 {
		t.Fatalf("expected break logged event, got %d", len(events))
	}
	if _, ok := events[0].(BreakLogged); !ok {
		t.Fatalf("expected BreakLogged, got %#v", events[0])
	}
	if got := a.Breaks().BreaksToday; got != 1 {
		t.Fatalf("expected 1 break today, got %d", got)
	}
}

func TestDayRolloverResetsCounters(t *testing.T) {
	a := newTestAgent()
	a.Drink(base.Add(5 * time.Minute))

	events := a.Step(base.Add(24 * time.Hour))
	rolled := false
	for _, ev := range events {
		if _, ok := ev.(DayRolled); ok {
			rolled = true
		}
	}
	if !rolled {
		t.Fatalf("expected DayRolled, got %v", events)
	}
	if got := a.Water().ConsumedML; got != 0 {
		t.Fatalf("expected water reset to 0, got %d", got)
	}
}

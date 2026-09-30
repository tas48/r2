package store

import (
	"path/filepath"
	"testing"
	"time"
)

var today = time.Date(2026, time.September, 30, 12, 0, 0, 0, time.Local)

func TestLoadMissingReturnsFreshState(t *testing.T) {
	s := New(filepath.Join(t.TempDir(), "state.json"))
	got, err := s.Load(today)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Day != "2026-09-30" || got.WaterML != 0 || got.BreaksToday != 0 {
		t.Fatalf("expected fresh state for today, got %#v", got)
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	s := New(filepath.Join(t.TempDir(), "nested", "state.json"))
	want := State{Day: "2026-09-30", WaterML: 1000, LastDrink: today, LastBreak: today, BreaksToday: 2}
	if err := s.Save(want); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	got, err := s.Load(today)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if got.Day != want.Day || got.WaterML != want.WaterML || got.BreaksToday != want.BreaksToday {
		t.Fatalf("round trip mismatch: got %#v want %#v", got, want)
	}
	if !got.LastDrink.Equal(want.LastDrink) || !got.LastBreak.Equal(want.LastBreak) {
		t.Fatalf("timestamps mismatch: got %#v want %#v", got, want)
	}
}

func TestLoadResetsStaleDay(t *testing.T) {
	s := New(filepath.Join(t.TempDir(), "state.json"))
	stale := State{Day: "2026-09-29", WaterML: 1500, BreaksToday: 3}
	if err := s.Save(stale); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	got, err := s.Load(today)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if got.Day != "2026-09-30" || got.WaterML != 0 || got.BreaksToday != 0 {
		t.Fatalf("expected counters reset, got %#v", got)
	}
}

func TestSaveReplacesExisting(t *testing.T) {
	s := New(filepath.Join(t.TempDir(), "state.json"))
	if err := s.Save(State{Day: "2026-09-30", WaterML: 500}); err != nil {
		t.Fatalf("first save failed: %v", err)
	}
	if err := s.Save(State{Day: "2026-09-30", WaterML: 1500}); err != nil {
		t.Fatalf("second save failed: %v", err)
	}

	got, err := s.Load(today)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if got.WaterML != 1500 {
		t.Fatalf("expected replaced value 1500, got %d", got.WaterML)
	}
}

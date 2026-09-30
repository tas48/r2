package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadMissingFileReturnsDefaults(t *testing.T) {
	got, err := Load(filepath.Join(t.TempDir(), "absent.json"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != Default() {
		t.Fatalf("expected defaults, got %#v", got)
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "config.json")
	want := Config{
		City:        "Recife",
		WaterGoalML: 2500,
		BottleML:    750,
		WaterEvery:  Duration(20 * time.Minute),
		BreakEvery:  Duration(50 * time.Minute),
	}
	if err := Save(path, want); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	got, err := Load(path)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if got != want {
		t.Fatalf("round trip mismatch: got %#v want %#v", got, want)
	}
}

func TestLoadPartialConfigKeepsDefaults(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte(`{"water_goal_ml": 3000}`), 0o644); err != nil {
		t.Fatalf("writing config: %v", err)
	}

	got, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.WaterGoalML != 3000 {
		t.Fatalf("expected water goal 3000, got %d", got.WaterGoalML)
	}
	if got.BottleML != Default().BottleML || got.WaterEvery != Default().WaterEvery {
		t.Fatalf("expected other fields to keep defaults, got %#v", got)
	}
}

func TestLoadRejectsInvalidDuration(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte(`{"water_every": "soon"}`), 0o644); err != nil {
		t.Fatalf("writing config: %v", err)
	}
	if _, err := Load(path); err == nil {
		t.Fatal("expected an error for an invalid duration")
	}
}

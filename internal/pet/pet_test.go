package pet

import (
	"math/rand"
	"testing"
	"time"
)

var base = time.Date(2026, time.September, 30, 9, 0, 0, 0, time.Local)

func newPet() *Pet {
	return New(base, rand.New(rand.NewSource(1)))
}

func TestNewStartsIdleUnpositioned(t *testing.T) {
	p := newPet()
	if p.State() != Idle {
		t.Fatalf("expected idle, got %s", p.State())
	}
	if p.X() != -1 {
		t.Fatalf("expected unpositioned pet, got x=%d", p.X())
	}
}

func TestStepTransitionsAfterStay(t *testing.T) {
	p := newPet()
	if got := p.State(); got != Idle {
		t.Fatalf("expected idle, got %s", got)
	}
	// Well past the longest idle stay.
	p.Step(9*time.Second, base.Add(9*time.Second), 20)
	if p.State() == Idle {
		t.Fatalf("expected a transition out of idle")
	}
}

func TestReactMapsEvents(t *testing.T) {
	tests := []struct {
		event EventKind
		want  State
	}{
		{EventWater, Happy},
		{EventBreak, Happy},
		{EventInteract, Happy},
		{EventReminder, Curious},
		{EventDayRolled, Wake},
	}
	for _, tt := range tests {
		p := newPet()
		p.React(Event{Kind: tt.event}, base)
		if p.State() != tt.want {
			t.Fatalf("event %d: got %s, want %s", tt.event, p.State(), tt.want)
		}
	}
}

func TestWalkClampsToArena(t *testing.T) {
	p := newPet()
	p.state = Walk
	p.until = base.Add(time.Hour)
	p.x = -1
	p.dir = 1

	// A huge delta would overshoot the right edge.
	p.Step(10*time.Second, base.Add(10*time.Second), 4)
	if p.X() != 4 || p.Direction() != -1 {
		t.Fatalf("expected clamp to right edge and turn, got x=%d dir=%d", p.X(), p.Direction())
	}
}

func TestWalkCentersWhenUnpositioned(t *testing.T) {
	p := newPet()
	p.state = Walk
	p.until = base.Add(time.Hour)
	p.dir = 1
	p.x = -1

	p.Step(time.Millisecond, base.Add(time.Millisecond), 10)
	if p.X() != 5 {
		t.Fatalf("expected centered start near 5, got %d", p.X())
	}
}

func TestSpriteForFallsBackToSmallerVariant(t *testing.T) {
	mini := SpriteFor(Sleep, SizeMini)
	if mini.Height != 1 {
		t.Fatalf("expected a one-line sprite, got height %d", mini.Height)
	}
	full := SpriteFor(Sleep, SizeFull)
	if full.Height != 3 {
		t.Fatalf("expected a three-line sprite, got height %d", full.Height)
	}
	if len(full.Frames) < 2 {
		t.Fatalf("expected an animated sprite, got %d frames", len(full.Frames))
	}
}

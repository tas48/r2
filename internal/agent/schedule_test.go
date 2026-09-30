package agent

import (
	"testing"
	"time"
)

func TestSchedulerDueAdvancesWithoutDrift(t *testing.T) {
	var s scheduler
	s.add(KindWater, 30*time.Minute, base)

	if got := s.due(base.Add(10 * time.Minute)); len(got) != 0 {
		t.Fatalf("expected nothing due, got %v", got)
	}
	if got := s.due(base.Add(30 * time.Minute)); len(got) != 1 {
		t.Fatalf("expected water due, got %v", got)
	}
	if got := s.due(base.Add(100 * time.Minute)); len(got) != 1 {
		t.Fatalf("expected water due once, got %v", got)
	}
	if got := s.due(base.Add(120 * time.Minute)); len(got) != 1 {
		t.Fatalf("expected water due at the realigned interval, got %v", got)
	}
}

func TestSchedulerPostpone(t *testing.T) {
	var s scheduler
	s.add(KindWater, 30*time.Minute, base)
	s.postpone(KindWater, base.Add(10*time.Minute))

	if got := s.due(base.Add(35 * time.Minute)); len(got) != 0 {
		t.Fatalf("expected postponed task, got %v", got)
	}
	if got := s.due(base.Add(40 * time.Minute)); len(got) != 1 {
		t.Fatalf("expected task after postpone, got %v", got)
	}
}

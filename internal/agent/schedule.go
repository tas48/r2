package agent

import "time"

type scheduled struct {
	kind  Kind
	every time.Duration
	next  time.Time
}

// scheduler tracks recurring tasks and advances them by whole intervals so a
// delayed tick neither accumulates drift nor fires a burst of missed reminders.
type scheduler struct {
	tasks []scheduled
}

func newScheduler(now time.Time, cfg Config) scheduler {
	var s scheduler
	s.add(KindWater, cfg.WaterEvery, now)
	s.add(KindBreak, cfg.BreakEvery, now)
	return s
}

func (s *scheduler) add(kind Kind, every time.Duration, now time.Time) {
	s.tasks = append(s.tasks, scheduled{kind: kind, every: every, next: now.Add(every)})
}

// due returns the kinds whose interval has elapsed and realigns their next
// fire time to the next multiple of the interval.
func (s *scheduler) due(now time.Time) []Kind {
	var kinds []Kind
	for i := range s.tasks {
		task := &s.tasks[i]
		if task.every <= 0 || now.Before(task.next) {
			continue
		}
		for !now.Before(task.next) {
			task.next = task.next.Add(task.every)
		}
		kinds = append(kinds, task.kind)
	}
	return kinds
}

// postpone restarts the interval for a task after the user satisfies it, so it
// does not fire immediately after being handled.
func (s *scheduler) postpone(kind Kind, now time.Time) {
	for i := range s.tasks {
		if s.tasks[i].kind == kind {
			s.tasks[i].next = now.Add(s.tasks[i].every)
		}
	}
}

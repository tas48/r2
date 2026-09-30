package agent

import "time"

// WaterTracker holds the daily water progress.
type WaterTracker struct {
	ConsumedML int
	GoalML     int
	BottleML   int
	LastDrink  time.Time
}

// BreakTracker holds the daily break progress.
type BreakTracker struct {
	LastBreak   time.Time
	BreaksToday int
}

func (w *WaterTracker) drink(now time.Time) WaterLogged {
	w.ConsumedML += w.BottleML
	w.LastDrink = now
	return WaterLogged{TotalML: w.ConsumedML, GoalML: w.GoalML, At: now}
}

func (b *BreakTracker) logBreak(now time.Time) BreakLogged {
	since := now.Sub(b.LastBreak)
	b.LastBreak = now
	b.BreaksToday++
	return BreakLogged{SinceLastBreak: since, At: now}
}

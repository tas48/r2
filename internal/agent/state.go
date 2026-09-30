package agent

import "time"

// State is a snapshot of the agent's daily counters, used for persistence.
type State struct {
	Day         string
	WaterML     int
	LastDrink   time.Time
	LastBreak   time.Time
	BreaksToday int
}

// Snapshot returns the counters to persist.
func (a *Agent) Snapshot() State {
	return State{
		Day:         a.day,
		WaterML:     a.water.ConsumedML,
		LastDrink:   a.water.LastDrink,
		LastBreak:   a.breaks.LastBreak,
		BreaksToday: a.breaks.BreaksToday,
	}
}

// Seed restores counters from a snapshot and restarts the intervals from now.
// A stale day is kept as-is so the next Step reports DayRolled and resets.
func (a *Agent) Seed(st State, now time.Time) {
	a.water.ConsumedML = st.WaterML
	a.water.LastDrink = st.LastDrink
	a.breaks.LastBreak = st.LastBreak
	a.breaks.BreaksToday = st.BreaksToday
	if st.Day != "" {
		a.day = st.Day
	}
	a.sched = newScheduler(now, a.cfg)
}

package agent

import "time"

const dayLayout = "2006-01-02"

// Config holds the tunable parameters of the agent. Defaults are hardcoded for
// now; configuration loading will overwrite them in a later step.
type Config struct {
	WaterGoalML  int
	BottleML     int
	WaterEvery   time.Duration
	BreakEvery   time.Duration
	WeatherEvery time.Duration
}

func DefaultConfig() Config {
	return Config{
		WaterGoalML:  2000,
		BottleML:     500,
		WaterEvery:   30 * time.Minute,
		BreakEvery:   45 * time.Minute,
		WeatherEvery: 15 * time.Minute,
	}
}

// Agent is the rule-based brain. It is pure: it never performs IO and receives
// time explicitly. The UI advances it with Step and interprets the events.
type Agent struct {
	cfg    Config
	day    string
	water  WaterTracker
	breaks BreakTracker
	sched  scheduler

	rainProb   int
	rainLikely bool
	weatherErr error
}

func New(cfg Config, now time.Time) *Agent {
	return &Agent{
		cfg:    cfg,
		day:    now.Format(dayLayout),
		water:  WaterTracker{GoalML: cfg.WaterGoalML, BottleML: cfg.BottleML},
		breaks: BreakTracker{LastBreak: now},
		sched:  newScheduler(now, cfg),
	}
}

// Step advances the agent to now and returns the events produced.
func (a *Agent) Step(now time.Time) []Event {
	var events []Event
	if ev, ok := a.rollDay(now); ok {
		events = append(events, ev)
	}
	for _, kind := range a.sched.due(now) {
		events = append(events, a.fire(kind, now))
	}
	return events
}

// Drink registers one bottle of water and restarts the water interval.
func (a *Agent) Drink(now time.Time) []Event {
	a.sched.postpone(KindWater, now)
	return []Event{a.water.drink(now)}
}

// LogBreak registers a break and restarts the break interval.
func (a *Agent) LogBreak(now time.Time) []Event {
	a.sched.postpone(KindBreak, now)
	return []Event{a.breaks.logBreak(now)}
}

// ApplyWeather records a forecast result and returns the resulting event.
func (a *Agent) ApplyWeather(rainProb int, tempC float64, rainLikely bool, err error, now time.Time) []Event {
	a.rainProb, a.rainLikely, a.weatherErr = rainProb, rainLikely, err
	return []Event{WeatherUpdated{RainProb: rainProb, TempC: tempC, RainLikely: rainLikely, Err: err, At: now}}
}
func (a *Agent) Water() WaterTracker  { return a.water }
func (a *Agent) Breaks() BreakTracker { return a.breaks }
func (a *Agent) Config() Config       { return a.cfg }

func (a *Agent) fire(kind Kind, now time.Time) Event {
	switch kind {
	case KindWater:
		return ReminderFired{Kind: KindWater, Text: waterReminder(a.water), At: now}
	case KindBreak:
		return ReminderFired{Kind: KindBreak, Text: breakReminder(), At: now}
	case KindWeather:
		return WeatherRequested{At: now}
	default:
		return ReminderFired{Kind: kind, At: now}
	}
}

func (a *Agent) rollDay(now time.Time) (Event, bool) {
	today := now.Format(dayLayout)
	if a.day == today {
		return nil, false
	}
	a.day = today
	a.water.ConsumedML = 0
	a.breaks.BreaksToday = 0
	return DayRolled{Date: today, At: now}, true
}
